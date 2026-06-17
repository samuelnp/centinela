# Plan — roadmap-doc-sync

Implements the feature brief at `docs/features/roadmap-doc-sync.md`. Approach
(user-chosen): **enrich JSON, full-fidelity generation**. `roadmap.json` becomes
the single source of truth; `ROADMAP.md` is generated deterministically from it;
a `roadmap_drift` gate fails `centinela validate` on divergence.

The **dominant design pressure is the ≤100-line hard rule (G1)**. A full-fidelity
markdown generator cannot live in one file, so the generator and its tests are
pre-split by responsibility (intro / phase / feature). Every new source file
below is named and budgeted ≤100 lines. No G1 exception is sought.

---

## Canonical markdown format (PINNED — generate the file to match THIS)

The generator emits, in order, iterating ONLY ordered slices:

1. **Title:** `# Roadmap` then one blank line.
2. **Intro blockquote** (from `Roadmap.Intro`): each line of `Intro` prefixed
   `> ` (with a trailing space → `> text`); an empty line *inside* the intro is
   emitted as a bare `>` (no trailing space). Then one blank line.
3. **Per phase** (iterate `Roadmap.Phases` in slice order):
   - `## ` + `Phase.Name` (the name carries any authored status glyph verbatim,
     e.g. `## ✅ Phase 0: Bootstrap`). Then one blank line.
   - If `Phase.Note != ""`: render it as a blockquote (same per-line `> ` rule,
     blank lines inside → bare `>`). Then one blank line.
   - **Backlog phase** (`IsBacklogPhaseName(name)` true): for each feature emit a
     deferred-finding bullet instead of the normal bullet (see below).
   - **Normal phase:** for each feature (slice order) emit the feature block.
   - One blank line after the last feature of the phase (i.e. blank line between
     phases). The final phase's trailing blank collapses into the single EOF
     newline.
4. **Feature bullet (normal):**
   - Base: `- **<Name>**`.
   - If `Description != ""`: ` — <Description>` appended on the same logical
     bullet. (Long descriptions are emitted as a SINGLE line — no word-wrapping;
     v1 picks the one-line canonical format. Out of scope: wrap fidelity.)
   - If `Fixes != ""`: a following line, indented two spaces:
     `  *Fixes: <Fixes>*`.
   - If `DependsOn` non-empty: append ` (depends on a, b, c)` to the description
     line (declared slice order, comma-space joined). When there is no
     description, the clause attaches directly: `- **<Name>** (depends on a)`.
5. **Feature bullet (Backlog / deferred finding):**
   `- **<Name>** — <Summary> *(deferred <DeferredAt> · <Source.Feature>/<Source.Role>)*`
   Omit the parenthetical fields that are empty; never emit empty `()`.
6. **EOF:** exactly one trailing `\n`. No trailing whitespace on any line.

Determinism contract: no Go map is iterated anywhere in rendering. `DependsOn`,
`Features`, `Phases` are slices and rendered in declared order. The migration
regenerates `ROADMAP.md` to this exact format — we do NOT bend the generator to
reproduce the current hand-written file character-for-character; we pick this
format and rewrite the committed file once.

---

## Step 1 — plan

Artifacts (this step): feature brief `docs/features/roadmap-doc-sync.md`, this
plan, and the Gherkin spec `specs/roadmap-doc-sync.feature` (authored by the
feature-specialist role). Scenarios trace to the Acceptance Criteria: generate
writes byte-stable output; drift gate fails on hand-edit; passes after
regenerate; missing ROADMAP.md fails clearly; unknown severity rejected;
Backlog renders; determinism (generate twice → identical).

## Step 2 — code

**Schema** — `internal/roadmap/roadmap.go` (edit, stays ≤100 lines):
add `Description`, `Fixes` to `Feature`; `Note` to `Phase`; `Intro` to
`Roadmap`, all `json:",omitempty"`. (Current file is 97 lines; the additions are
struct-field lines only — if it crosses 100, move the three struct definitions
into a new `internal/roadmap/types.go` and leave `Load`/`Save`/`Summary` in
`roadmap.go`.) No change to `Load`/`Save` logic — `MarshalIndent` and
rawio/rawmutate already preserve typed and unknown keys.

**Generator** — pure functions in `internal/roadmap` (supporting domain; NOT in
`internal/ui`):
- `internal/roadmap/mdgen.go` (new, ≤100): `RenderMarkdown(*Roadmap) ([]byte,
  error)` — owns the top-level loop: title, `renderIntro`, phase loop delegating
  to `renderPhase`, blank-line joining, and the single-trailing-newline
  guarantee. Plus `renderIntro`/`renderBlockquote(string) []string` helper
  (per-line `> ` with bare `>` for blank inner lines) if it fits; otherwise the
  blockquote helper moves to `mdgen_phase.go`.
- `internal/roadmap/mdgen_phase.go` (new, ≤100): `renderPhase(Phase) []string`
  — heading, optional note blockquote, dispatch to Backlog vs normal feature
  rendering. Hosts `renderBlockquote` if `mdgen.go` is tight.
- `internal/roadmap/mdgen_feature.go` (new, ≤100): `renderFeature(Feature)
  []string` (normal bullet + fixes + deps clause) and `renderBacklogFeature(
  Feature) []string` (deferred-finding line). Uses `IsBacklogPhaseName` decision
  made by the caller; this file just formats both shapes.

Builder pattern: each `render*` returns `[]string` (lines, no trailing newline);
`mdgen.go` joins with `\n` and appends the single EOF `\n`. This keeps each file
small and makes golden-testing line-exact.

**Command** — `cmd/centinela/roadmap_generate.go` (new, thin, ≤100):
adds `centinela roadmap generate` as a subcommand of the existing `roadmapCmd`
(in `cmd/centinela/roadmap.go`). It calls `roadmap.Load()`, `roadmap.
RenderMarkdown(r)`, writes `ROADMAP.md` (0644), prints the path. No business
logic — all formatting is in `internal/roadmap`. Register via `roadmapCmd.
AddCommand(...)` in this file's `init()`.

**Gate** — config + check + wiring (model exactly on spec-traceability):
- `internal/config/roadmap_drift.go` (new, ≤100): `RoadmapDriftConfig{Enabled
  bool, Severity string}`; `NormalizeRoadmapDrift` (default severity `warn`);
  `validateRoadmapDrift` (severity must be `fail`|`warn`, no-op when disabled).
- `internal/config/config.go` (edit): add `RoadmapDrift RoadmapDriftConfig
  \`toml:"roadmap_drift"\`` to `GatesConfig`.
- `internal/config/defaults.go` (edit, 1 line): `cfg.Gates.RoadmapDrift =
  NormalizeRoadmapDrift(cfg.Gates.RoadmapDrift)`.
- `internal/config/file_size_exceptions.go` (edit, in `validateConfig`): call
  `validateRoadmapDrift(cfg.Gates.RoadmapDrift)`.
- `internal/gates/roadmap_drift.go` (new, ≤100): `checkRoadmapDrift(cfg
  *config.Config, _ *gitdiff.Set) Result` — `roadmap.Load()`; on load error →
  Fail (names roadmap.json). `RenderMarkdown(r)` → `want`. Read `ROADMAP.md`;
  if missing → severity-mapped Fail/Warn with "ROADMAP.md missing — run
  centinela roadmap generate". `bytes.Equal(want, got)` → Pass ("ROADMAP.md is
  in sync."); else severity-mapped Fail/Warn with a Detail naming the first
  differing line (reuse a `firstDifferingLine` helper, same logic as the
  scaffold-parity test) + "run centinela roadmap generate". The gate ignores
  the diff filter (whole-file artifact).
- `internal/gates/gates.go` (edit): in `RunWithFilter`, append
  `if cfg.Gates.RoadmapDrift.Enabled { results = append(results,
  checkRoadmapDrift(cfg, filter)) }`.
- `centinela.toml` (edit): add
  `[gates.roadmap_drift]` `enabled = true` `severity = "warn"` with a comment
  mirroring the spec-traceability adoption note (ratchet to `fail` once clean).

**Migration** — populate `.workflow/roadmap.json` (worktree copy) with the full
existing `ROADMAP.md` prose: `intro` (the opening blockquote incl.
capability-spectrum principle + status/ordering line), each phase `note` (the
`>` blockquotes verbatim), each feature `description` and `fixes`. Then build a
local binary (`go build -o /tmp/cmr-centinela ./cmd/centinela`) and run
`/tmp/cmr-centinela roadmap generate` from the worktree to regenerate
`ROADMAP.md`. Diff regenerated vs original; reconcile the *canonical format*
choice and the JSON prose until content is faithful, then commit the
byte-clean pair. (Backlog's existing single deferred entry already carries
`summary`/`source`/`deferredAt` — it renders via the deferred-finding shape.)

## Step 3 — tests

Coverage gate is **95% per-package**: only colocated `_test.go` files (each
≤100 lines) move it; files under `tests/` do not. So unit coverage for the
generator/gate/config MUST be colocated.

- **Unit (colocated, move coverage):**
  - `internal/roadmap/mdgen_test.go` — `RenderMarkdown` golden (full roadmap →
    expected bytes), determinism (call twice → `bytes.Equal`), single-trailing-
    newline, intro multi-line + blank-inner-line blockquote.
  - `internal/roadmap/mdgen_feature_test.go` — feature with both fields; with
    description only; with fixes only; with neither (bare bullet, no dangling
    em-dash); with `DependsOn` (clause order); Backlog deferred-finding line.
  - `internal/roadmap/mdgen_phase_test.go` — phase heading with status glyph in
    name; multi-paragraph note blockquote; phase with no note.
  - `internal/config/roadmap_drift_test.go` — Normalize default `warn`;
    validate accepts `fail`/`warn`, rejects garbage, no-op when disabled.
  - `internal/gates/roadmap_drift_test.go` — in-sync → Pass; drifted → Fail
    (severity fail) / Warn (severity warn) with first-differing-line detail;
    missing ROADMAP.md → Fail/Warn; roadmap.json load error → Fail. (Split into
    a second `_test.go` if it exceeds 100 lines.)
- **Integration:** `tests/integration/roadmap_doc_sync_test.go` — generate→read
  →gate round-trip in a temp dir: write enriched roadmap.json, `RenderMarkdown`,
  write ROADMAP.md, gate Pass; mutate ROADMAP.md, gate Fail; regenerate, gate
  Pass.
- **Acceptance (executable, `tests/acceptance/`):**
  `tests/acceptance/roadmap_doc_sync_test.go` — build/run the CLI in a temp
  repo: `centinela roadmap generate` writes ROADMAP.md byte-equal to a second
  generate (idempotent); hand-edit ROADMAP.md then `centinela validate` exits
  non-zero with the drift message (severity fail); `centinela roadmap generate`
  then `centinela validate` passes. Tag each scenario with `// Acceptance:
  specs/roadmap-doc-sync.feature` + `// Scenario: <name>` so the
  spec-traceability gate maps them. Add the acceptance run to
  `validate.commands` if not already covered by `go test ./tests/acceptance/...`
  (it is). Author `.workflow/roadmap-doc-sync-edge-cases.md` enumerating the
  brief's edge cases.

## Step 4 — validate

Run `centinela validate` (gates incl. the new `roadmap_drift` on the committed
byte-clean tree → Pass; spec-traceability maps the new scenarios; G1 confirms
every new file ≤100). Full `go test ./...` + `go test ./tests/acceptance/...` +
coverage + fmt. Produce the gatekeeper report at
`.workflow/roadmap-doc-sync-gatekeeper.md` (SAFE/WARNING) verifying: all new
source + `_test.go` files ≤100 lines, no cross-layer imports (gate imports
config + roadmap + gitdiff only; generator is pure roadmap-domain; command is
thin), `centinela validate` passes, no business logic in cmd or ui.

## Step 5 — docs

This is an internal/operability feature (no end-user UI surface), so docs are
right-sized: documentation-specialist evidence pair
(`.workflow/roadmap-doc-sync-documentation-specialist.{md,json}`), a changelog
artifact (`.workflow/roadmap-doc-sync-changelog.md`, created early via
`evidence artifact new`), and `docs/project-docs/index.html` regeneration.
Document the new `centinela roadmap generate` command and the
`[gates.roadmap_drift]` config knob.

---

## File-level G1 budget (the dominant constraint)

| File | New/Edit | Role | Budget |
|------|----------|------|--------|
| `internal/roadmap/roadmap.go` | edit | +5 struct fields (split to `types.go` if >100) | ≤100 |
| `internal/roadmap/mdgen.go` | new | `RenderMarkdown` + intro + join + EOF newline | ≤100 |
| `internal/roadmap/mdgen_phase.go` | new | phase heading + note blockquote + dispatch | ≤100 |
| `internal/roadmap/mdgen_feature.go` | new | normal bullet + fixes + deps + Backlog line | ≤100 |
| `cmd/centinela/roadmap_generate.go` | new | thin `roadmap generate` command | ≤100 |
| `internal/config/roadmap_drift.go` | new | config struct + Normalize + validate | ≤100 |
| `internal/gates/roadmap_drift.go` | new | `checkRoadmapDrift` + first-diff detail | ≤100 |
| `internal/config/config.go` | edit | +1 GatesConfig field | ≤100 |
| `internal/config/defaults.go` | edit | +1 Normalize call | ≤100 |
| `internal/config/file_size_exceptions.go` | edit | +1 validate call | ≤100 |
| `internal/gates/gates.go` | edit | +1 enabled branch | ≤100 |
| `centinela.toml` | edit | `[gates.roadmap_drift]` block | n/a |
| unit `_test.go` ×5 | new | colocated coverage (≤100 each) | ≤100 |
| `tests/integration/...` `tests/acceptance/...` | new | round-trip + e2e | ≤100 |

If any generator file still threatens 100 lines after first cut, the
blockquote helper and the deps-clause builder are the spill targets to extract
into a tiny `mdgen_util.go`.
