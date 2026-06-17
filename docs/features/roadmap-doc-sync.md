# roadmap-doc-sync

Make `.workflow/roadmap.json` the single source of truth and **generate** the
human-readable `ROADMAP.md` from it, with a drift-check gate that fails
`centinela validate` when the committed `ROADMAP.md` no longer matches what the
generator would produce.

## Problem

Centinela keeps two roadmaps: the machine-readable `.workflow/roadmap.json`
(consumed by `roadmap validate`, analysis/quality gates, the `dependsOn`
scheduler, defer/promote) and the human-readable `ROADMAP.md` (the narrative
the maintainer actually reads and edits). Today they are maintained **by hand,
independently**, and they drift: this very roadmap has drifted from
`roadmap.json` twice already. The person hurt is the operator/maintainer who
hand-syncs the two files — they either notice the drift late (after merging an
inconsistent roadmap) or never, eroding trust in the roadmap as a planning
source. Centinela's founding principle is "treat every failure as an
engineering problem to fix permanently, not a prompt to retry"; a hand-synced
pair of files is exactly the manual toil this product exists to mechanize.

The fix: enrich `roadmap.json` so it carries the human-facing prose currently
living only in `ROADMAP.md`, generate `ROADMAP.md` deterministically from that
single source, and add a mechanical drift gate so the two can never silently
diverge again.

## User Stories

- As a maintainer, I edit a feature's narrative once (in `roadmap.json`, via
  the file or future tooling), run `centinela roadmap generate`, and get a
  `ROADMAP.md` that is byte-identical to what every machine consumer sees — so
  I never hand-edit two files.
- As a maintainer, when I (or an agent) hand-edit `ROADMAP.md` directly and
  forget to update `roadmap.json`, `centinela validate` fails with the first
  differing line and tells me to run `centinela roadmap generate` — so drift is
  caught at the gate, not in review.
- As a maintainer adopting the gate on a repo that isn't yet clean, I set
  `severity = "warn"` so drift is surfaced without blocking merge, then ratchet
  to `fail` once `ROADMAP.md` is regenerated.
- As a CI operator, the generated `ROADMAP.md` is identical on macOS, Linux,
  and CI runners — no map-ordering nondeterminism — so the gate never
  flip-flops between platforms.

## Acceptance Criteria

1. **Schema carries prose.** `roadmap.json` accepts a top-level `intro`
   (string), a per-phase `note` (string, possibly multi-paragraph), and
   per-feature `description` (string) and `fixes` (string). All are optional;
   `Load()`/`Save()` round-trip them and unknown-key preservation
   (rawio/rawmutate) is unaffected.
2. **`centinela roadmap generate` writes `ROADMAP.md`** rendered
   deterministically from `roadmap.json`, exiting 0 and reporting the path
   written.
3. **Generated output is byte-stable and map-free.** Running `generate` twice
   produces identical bytes; the renderer iterates only ordered slices
   (`Phases`, `Features`, `DependsOn`) — never a Go map — and the file ends with
   exactly one trailing newline.
4. **No live status in the generated file.** The generated `ROADMAP.md` contains
   no per-feature ✓/✅/in-progress marker; live status remains exclusively in
   `centinela roadmap`. The file is the static plan only.
5. **Drift gate passes when in sync.** With `[gates.roadmap_drift] enabled =
   true` and an in-sync `ROADMAP.md`, `centinela validate`'s `roadmap-drift`
   result is `Pass`.
6. **Drift gate fails (or warns) on mismatch.** Hand-edit `ROADMAP.md`; under
   `severity = "fail"` the gate returns `Fail` with a detail naming the first
   differing line number and the remediation `run centinela roadmap generate`;
   under `severity = "warn"` it returns `Warn` (non-blocking). After
   `centinela roadmap generate` it returns `Pass`.
7. **Missing `ROADMAP.md` is a clear failure.** With the gate enabled and no
   `ROADMAP.md` on disk, the gate returns `Fail`/`Warn` (per severity) with a
   message saying the file is missing and to run `generate` — never a panic or
   a bare I/O error.
8. **Unknown severity is rejected at config load.** A `severity` other than
   `fail`/`warn` fails `config.Load()` with a clear error (mirrors
   spec-traceability), and is a no-op when the gate is disabled.
9. **One-time migration committed.** `roadmap.json` is populated with the full
   existing `ROADMAP.md` prose and `ROADMAP.md` is regenerated so the committed
   file is byte-identical to generator output; `centinela validate` passes on
   the committed tree.
10. **Backlog renders from deferred fields.** The Backlog phase is rendered
    using its deferred-finding fields (`summary`, `source`, `deferredAt`), not
    the `description`/`fixes` bullet shape, and is excluded from live-status
    semantics exactly as today.

## Edge Cases

- **Backlog rendering.** Backlog features have `summary`/`source`/`deferredAt`
  but no `description`/`fixes`; render them in a deferred-finding line format,
  distinct from normal feature bullets. Backlog must still appear in the file.
- **Feature with no description and no fixes.** Render the bare bullet
  `- **name**` with no em-dash clause and no `*Fixes:*` line. Never emit a
  dangling ` — ` or an empty `*Fixes: *`.
- **Feature with description but no fixes** (and vice-versa): emit the present
  field only; the `*Fixes:*` line appears iff `fixes` is non-empty.
- **Multi-paragraph phase note.** A `note` containing `\n\n` must render as a
  blockquote where blank separator lines are emitted as a bare `>` (matching the
  current `ROADMAP.md` blockquote-with-blank-line style), not as an empty line
  that breaks the blockquote.
- **Empty `dependsOn` vs present.** When `DependsOn` is empty, emit no deps
  annotation; when present, append a single canonical `(depends on a, b)` clause
  in declared slice order.
- **Missing `ROADMAP.md`.** Gate reports a clear "missing" failure (AC7);
  `generate` simply creates it.
- **Trailing newline.** Exactly one `\n` at EOF; the byte-compare and the
  migration must agree on this or the gate fails on a committed-clean tree.
- **Non-ASCII.** Prose contains em-dashes, `✅`/`✓` (in phase *headings* like
  `## ✅ Phase 0`), curly quotes, accents. The generator must pass UTF-8 through
  byte-for-byte; phase-heading status glyphs live in the `name` string verbatim
  (they are part of the authored phase name, not live status).
- **Gate severity warn vs fail.** Same mismatch yields `Warn` (non-blocking) or
  `Fail` (blocking) purely by config; default ships as `warn` for safe adoption.
- **Determinism / no map ordering.** Any iteration over a Go map in the renderer
  is a latent CI flake; the design forbids it. Covered by a determinism unit
  test (generate twice → byte-equal).
- **Live status intentionally absent.** A reviewer expecting ✓ marks in the
  generated file is wrong by design — the generated file is the plan, not the
  dashboard. An explicit unit test asserts no status glyph is emitted for
  features (as opposed to authored phase-name glyphs).
- **Phase with zero features.** A phase whose `features` array is empty must
  still render its heading (and optional note blockquote) without panicking or
  emitting stray blank lines. The file must remain valid with exactly one
  trailing newline.
- **CRLF vs LF.** The generator must produce LF-only line endings on all
  platforms (macOS, Linux, Windows CI runners). The drift gate must treat a
  CRLF-terminated on-disk file as drifted even when content is otherwise
  identical, and never normalise line endings silently during comparison.
- **Feature with fixes but no description.** The `*Fixes: …*` line must still
  appear even when `description` is empty, preceded only by the bare
  `- **name**` bullet (no dangling ` — `). This is the inverse of the
  description-only case and is distinct from the no-prose case.
- **Gate disabled with bad severity.** A `severity` value that would otherwise
  fail validation is a no-op when `enabled = false`; config load must succeed so
  operators can disable the gate first and clean up the config later.

## Data Model

New optional fields, added to the existing structs in `internal/roadmap`
(rawio/rawmutate preserve them through defer/promote because they round-trip as
typed JSON):

```go
type Feature struct {
    Name        string   `json:"name"`
    DependsOn   []string `json:"dependsOn,omitempty"`
    Archetype   string   `json:"archetype,omitempty"`
    Description string   `json:"description,omitempty"` // NEW: human bullet prose
    Fixes       string   `json:"fixes,omitempty"`       // NEW: "*Fixes: …*" clause
    Summary     string   `json:"summary,omitempty"`     // deferred-finding one-liner
    Source      *Source  `json:"source,omitempty"`
    DeferredAt  string   `json:"deferredAt,omitempty"`
}

type Phase struct {
    Name     string    `json:"name"`
    Note     string    `json:"note,omitempty"` // NEW: blockquote rationale prose
    Features []Feature `json:"features"`
}

type Roadmap struct {
    Intro  string  `json:"intro,omitempty"` // NEW: top-of-file blockquote
    Phases []Phase `json:"phases"`
}
```

`Description`/`Fixes` are the per-feature bullet narrative; `Note` is the phase
rationale blockquote (multi-paragraph allowed); `Intro` is the opening
blockquote (capability-spectrum principle + the status/ordering line). Backlog
features keep using `Summary`/`Source`/`DeferredAt` and ignore
`Description`/`Fixes`.

## Integration Points

- **validate gate pipeline.** New `roadmap_drift` gate registered in
  `internal/gates/gates.go::RunWithFilter` behind `cfg.Gates.RoadmapDrift.Enabled`
  (filter is ignored — the gate is whole-file, not diff-scoped). `validate.go`
  already prints every `gates.Result`, so no command change.
- **config.** New `RoadmapDriftConfig{Enabled, Severity}` field on
  `GatesConfig`; `applyDefaults` calls `NormalizeRoadmapDrift`; `validateConfig`
  (in `internal/config/file_size_exceptions.go`) calls `validateRoadmapDrift`.
- **`centinela.toml`.** New `[gates.roadmap_drift]` block, shipped
  `enabled = true`, `severity = "warn"` (safe adoption; ratchet to `fail`).
- **Generation logic** lives in `internal/roadmap` (supporting domain) as a pure
  `RenderMarkdown(*Roadmap) ([]byte, error)`. The `cmd/centinela` command is a
  thin orchestrator. The markdown emitter does NOT go in `internal/ui` (that is
  terminal presentation only).
- **One-time migration.** Populate `roadmap.json` with all current `ROADMAP.md`
  prose, regenerate `ROADMAP.md`, commit both so the tree is byte-clean.

## Risks

- **Faithful prose migration.** Hand-transcribing the existing `ROADMAP.md`
  narrative into JSON risks dropping or altering a sentence; the generated file
  would then differ from intent (though the gate would still pass once
  regenerated). Mitigation: migrate, regenerate, and diff the regenerated file
  against the original `ROADMAP.md`; iterate the canonical format + JSON until
  the *semantic* content matches and the chosen line format is consistent.
- **Generator ↔ file divergence is the whole point and the whole risk.** If the
  canonical format isn't pinned exactly (blank-line-in-blockquote handling,
  deps clause, trailing newline), the committed file will never byte-match.
  Mitigation: define ONE canonical format in the plan, regenerate to match it
  (don't reverse-engineer the generator to the hand-written file char-for-char).
- **≤100-line / G1 pressure (dominant).** A full-fidelity markdown generator
  (intro + phase + note + feature + fixes + deps + Backlog) will exceed 100
  lines in one file. Mitigation: split into `mdgen.go` (orchestrator + intro +
  trailing-newline), `mdgen_phase.go` (phase heading + note blockquote),
  `mdgen_feature.go` (feature bullet + fixes + deps + Backlog line). Same for
  tests. This is the dominant design pressure — every file is named and
  budgeted in the plan.
- **CI determinism.** Map iteration anywhere in the renderer flakes CI.
  Mitigation: iterate ordered slices only; a determinism test (generate twice →
  byte-equal) guards it.

## Decomposition

This feature is **cohesive — keep it as one feature**. The schema fields, the
generator, the command, the gate, and the migration are a single closed loop:
the schema is meaningless without the generator, the generator is unverified
without the gate, and the gate is red until the migration regenerates the file.
Splitting would ship dead code (enriched schema no one renders) or a red gate
(gate with no in-sync file). The natural *file-level* work units (all inside one
feature) are:

1. Schema fields on `Feature`/`Phase`/`Roadmap` (`internal/roadmap/roadmap.go`).
2. Deterministic generator: `internal/roadmap/mdgen.go` + `mdgen_phase.go` +
   `mdgen_feature.go` → `RenderMarkdown(*Roadmap) ([]byte, error)`.
3. Command: `cmd/centinela/roadmap_generate.go` (thin).
4. Gate: `internal/gates/roadmap_drift.go` + config
   `internal/config/roadmap_drift.go` + wiring (gates.go, defaults.go,
   file_size_exceptions.go, GatesConfig, centinela.toml).
5. Migration: enrich `roadmap.json`, regenerate `ROADMAP.md` byte-exact.
6. Tests across all three tiers (unit golden/determinism/Backlog, integration
   round-trip + drift, acceptance generate/validate-fail/validate-pass).
