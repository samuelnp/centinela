# Plan — delivery-artifact-generation

## Problem & Goal

`completion-delivery-prompt` shipped `centinela deliver <feature> --via pr|merge`.
The PR path today calls `gh pr create --head <feature> --fill`, so the PR body is
just scraped commit subjects, and `CHANGELOG.md` is edited by hand or skipped.
Centinela already holds rich, freshest-at-delivery evidence — the feature brief,
the plan, per-role orchestration evidence, the gatekeeper report, and
claim-verification results — but throws it away at the exact moment it is most
useful.

**Goal:** compose, from that evidence, two delivery artifacts at `--via pr`:

1. A Markdown **PR body** (summary, what/why, acceptance reference, gate status,
   provenance footer), passed to `gh pr create --body-file`.
2. A single Keep-a-Changelog **changelog line** inserted idempotently under the
   correct `### Added`/`### Changed`/`### Fixed` subsection of the
   `## [Unreleased]` block in `CHANGELOG.md`.

Composition is read-only, degrades section-by-section when a source is missing,
never fabricates a gate result, and never blocks delivery. `cmd/` stays a thin
orchestrator (G7).

## Proposed Architecture

A NEW read-only aggregator package **`internal/delivery`** is the natural home,
mirroring `internal/insights` / `internal/reconstruct`: it imports
domain/leaf/aggregator read-only, is consumed only by `cmd/`, and its rendered
output types may be imported by `internal/ui`. It reads:

- `internal/evidence` (read-only): `evidence.Read(feature, role)` for the JSON
  artifacts; `evidence.ReadCompanion(feature, role)` for the `.md` companions
  (gatekeeper report, changelog stub).
- `internal/verify` (read-only): the `VerificationResult` / `Tally()` for the
  gate-status line. The verification *run* (which executes commands) stays in
  `cmd/`; `internal/delivery` only consumes an already-produced
  `verify.VerificationResult` value passed in by the caller — keeping
  `internal/delivery` pure and side-effect-free.
- `internal/config` (read-only): only if a path/locale-free constant is needed;
  prefer none.
- Direct `os.ReadFile` for the feature brief (`docs/features/<feature>.md`) and
  plan (`docs/plans/<feature>.md`), and for `CHANGELOG.md`.

### Package API and files (each < 100 lines)

Input is a plain struct the caller (cmd/) populates by reading from disk, so the
package itself can be tested with in-memory inputs and stays side-effect-light.

- `delivery.go` — package doc + the input/output structs:
  - `type Evidence struct { Feature, Brief, Plan, GatekeeperReport, ChangelogStub string; Verification *verify.VerificationResult }`
    (all strings are the already-read file bodies; empty string = missing source).
  - `type PRBody struct { Markdown string }` and
    `type ChangelogEntry struct { Category, Line string }` (rendered outputs;
    `ChangelogEntry` may be imported by `internal/ui` if a preview panel is
    wanted — none planned for the first slice).
- `prbody.go` — `func ComposePRBody(e Evidence) string`: assembles the ordered
  sections (below). Pure string building; no I/O.
- `sections.go` — small per-section helpers (`summarySection`, `whatWhySection`,
  `acceptanceSection`, `gateStatusSection`, `provenanceFooter`) each returning
  `string`, each gracefully returning `""` when its source datum is absent.
- `extract.go` — pure Markdown helpers that pull a named `##` section body out of
  the brief/plan text (e.g. `Problem`, `Who / Why`, `Acceptance Summary`). Used
  by `sections.go`. No regex-heavy parsing; simple line scanning by heading.
- `changelog.go` — `func ComposeChangelog(e Evidence) ChangelogEntry`: picks the
  category and the line (below). Pure.
- `changelog_insert.go` — `func InsertEntry(changelogMD string, entry ChangelogEntry) (string, bool)`:
  pure transform that returns the new `CHANGELOG.md` text and whether it changed
  (idempotency / dedupe logic below). Takes and returns the full file text; the
  actual file read/write stays in `cmd/`.

The package performs **no writes**. All file reads of brief/plan/CHANGELOG and
all writes happen in `cmd/`; `internal/delivery` is given text and returns text.

## PR-Body Composition

`ComposePRBody(e)` emits these sections in order, each independently omitted (not
faked) when its source is missing:

| Section | Source | Degradation when missing |
|---------|--------|--------------------------|
| Title/summary | `## Problem` body of brief (`e.Brief`); fallback to plan `Problem & Goal` | If neither present, omit; never invent. |
| What / Why | brief `## Who / Why` + plan summary | Omit the absent half; keep what exists. |
| Acceptance reference | brief `## Acceptance Summary` + a literal pointer to `specs/<feature>.feature` | If brief lacks the heading, emit only the spec pointer; if spec path is unknown, omit. |
| Gate status | `e.Verification.Tally()` (pass/fail/skip/warn) + the gatekeeper report's verdict line scanned from `e.GatekeeperReport` (e.g. `SAFE`/`WARNING`/`UNSAFE`) | If `Verification == nil`, omit the claim-tally line; if gatekeeper text empty, omit the verdict line. **Never** print a passing gate it cannot source. |
| Provenance footer | Always: a fixed line e.g. `Generated by Centinela from <feature> delivery evidence.` (no secret/dynamic data) | Always present (only constant text), so it is the one guaranteed section. |

Extraction is heading-scoped: find `## <Heading>`, take lines until the next
`## ` or EOF, trim. Missing heading → empty string → section omitted. This keeps
each helper tiny and tolerant of brief/plan format drift.

## CHANGELOG Composition

`ComposeChangelog(e)` produces one `ChangelogEntry{Category, Line}`:

- **Seed line.** If `e.ChangelogStub` (the `.workflow/<feature>-changelog.md`
  body) has a non-blank first line that is not still a FILL slot, use it
  verbatim as the `- …` line. Otherwise derive a line from the brief's `##
  Problem`/title (`- <feature>: <one-line summary>`).
- **Category selection.** Inspect the seed line's conventional-commit-style
  prefix: `feat:`/`feature` → `Added`; `fix:`/`bug` → `Fixed`; everything else
  (`refactor:`, `chore:`, `change`, etc.) → `Changed`. Default `Changed` when
  ambiguous. Category is one of exactly `Added`/`Changed`/`Fixed` to match the
  existing subsections.

`InsertEntry(changelogMD, entry)`:

1. Locate the `## [Unreleased]` block (from that line until the next `## ` or
   the `---` separator / EOF).
2. Within it, find the `### <Category>` subsection. If absent, create it in the
   canonical order `Added` → `Changed` → `Fixed`.
3. **Idempotency:** if the exact normalized line already exists anywhere in the
   `[Unreleased]` block, return the text unchanged and `false`. Otherwise append
   the line at the end of the subsection's bullet list and return the new text +
   `true`. Normalization = trim trailing whitespace; compare the bullet text.

This makes re-running `deliver` a no-op on the changelog (acceptance criterion)
and keeps insertion confined to `[Unreleased]` so released sections are never
touched.

## cmd/ Wiring (thin orchestrator)

`runDeliverPR` in `cmd/centinela/deliver_pr.go` is enriched. To honor G7 and the
100-line rule, split reading-and-composing into a small cmd-side helper file:

- `cmd/centinela/deliver_artifacts.go` (new, < 100 lines):
  - `buildPRBody(feature string) (string, error)`: reads brief/plan/gatekeeper
    companion/changelog stub from disk, runs `verify.Verify` (or reuses the
    prior result — see below), builds `delivery.Evidence`, calls
    `delivery.ComposePRBody`. Writes the body to a temp file and returns its path
    for `--body-file`.
  - `writeChangelog(feature string) error`: reads `CHANGELOG.md` +
    changelog stub, calls `delivery.ComposeChangelog` + `delivery.InsertEntry`,
    writes back only when changed.
- `deliver_pr.go` changes:
  - replace the `ghCreatePR(feature)` seam call to additionally accept a body
    path: `gh pr create --head <feature> --body-file <path>` (drop `--fill`).
    Keep `ghCreatePR` an overridable `var` seam for tests.
  - before opening the PR (after the successful push, before `gh`), call
    `writeChangelog(feature)` and, if it changed the file, commit it
    (`git add CHANGELOG.md && git commit -m "docs(changelog): …"`) and re-push,
    OR fold it into the pre-push commit. **Recommendation:** write + commit the
    changelog line BEFORE the push so the PR includes it and there is a single
    push. Reorder `runDeliverPR` to: dirty-check → compose+commit changelog →
    push → build PR body file → `gh --body-file`.

**Verification reuse.** Running `verify.Verify` shells out to the test command.
At delivery the validate gate already passed, so re-running is wasteful and slow.
Recommendation for the first slice: pass `verify.Deps{PriorTestRun: …}` is not
available across processes, so instead **degrade gracefully** — attempt a
read-only verification with a short timeout; if it errors or times out, omit the
gate-tally line (the brief explicitly allows omission). Keep the gate-status
line driven primarily by the gatekeeper report verdict (a static `.md`), which
needs no command execution. This avoids a slow/fragile double test run on the
critical delivery path.

**No new subcommand.** The brief frames this as enriching `deliver --via pr`, and
the simplest thin design folds both artifacts into that path. A separate
`centinela changelog` subcommand is deferred (see Deferred Findings) unless
testing ergonomics demand it; the integration tests can call
`writeChangelog`/`buildPRBody` directly.

## PROJECT.md G2 Edit

Append to the **Architecture Choice → G2 rule** paragraph, mirroring the
`internal/insights`/`internal/reconstruct` allowances:

> `internal/delivery` also joins the **aggregator** layer: a read-only delivery
> artifact composer for `centinela deliver --via pr` that may import
> `internal/evidence`, `internal/verify`, and `internal/config` (all read-only)
> and stdlib only; it must not import `cmd/` or `internal/ui` and is itself
> imported only by `cmd/` (its `PRBody`/`ChangelogEntry` types by `internal/ui`
> for rendering, if ever needed). Its edges `delivery → evidence`/`verify` are
> aggregator→domain (allowed via the aggregator layer's
> `allow: ["domain","leaf","aggregator"]`); `evidence`/`verify` never import
> `delivery`, so there is no cycle.

Also add `internal/delivery/` to the Folder Structure block and the Gatekeeper
Paths table. Confirm `[gates.import_graph]` in `centinela.toml` maps the new
package into the `aggregator` layer (add `internal/delivery` to that layer's
`paths`), or it will surface as an unmapped-package warning.

## Rollout Sequence (smallest correct slice first)

1. **Slice A — changelog only (pure + wired).** `internal/delivery`
   `changelog.go` + `changelog_insert.go` with full unit tests; cmd-side
   `writeChangelog` wired into `runDeliverPR` before push. Idempotent insertion
   is the highest-value, lowest-risk piece and needs no `gh`. PROJECT.md G2 +
   import_graph mapping land here so the gate is green from the start.
2. **Slice B — PR body composition.** `prbody.go` + `sections.go` + `extract.go`
   + `delivery.go` structs, unit-tested with in-memory briefs/plans. Wire
   `buildPRBody` + `gh --body-file` into `runDeliverPR`, dropping `--fill`.
3. **Slice C — gate-status integration + graceful degradation polish.** Wire the
   gatekeeper-verdict scan and the optional short-timeout verification tally;
   prove section-by-section omission with missing-source tests. Acceptance test
   drives the binary end to end with `gh`/`git` seams stubbed.

Each slice keeps the binary shippable and the gates green.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|------------|
| Regressing `completion-delivery-prompt`'s deliver flow (push/`gh` ordering, exit codes) | High | Medium | Keep the existing `gitDeliver`/`ghAvailable`/`ghCreatePR` seams; only extend `ghCreatePR` to take a body path. Reuse existing `deliver_pr_test.go`/`deliver_test.go`; add cases asserting push-still-happens and the gh-absent honest-failure path is unchanged. |
| `gh --body-file` portability (path quoting, temp-file lifetime, CRLF) | Medium | Low | Use `os.CreateTemp`, pass an absolute path, `defer os.Remove`. `--body-file` is standard in all supported `gh` versions; no shell interpolation (exec args, not a shell string). |
| CHANGELOG idempotency / duplicate or misplaced lines | Medium | Medium | `InsertEntry` is a pure, fully unit-tested transform: dedupe by normalized bullet text scoped to `[Unreleased]`; golden-file tests for first-insert, re-insert (no-op), missing-subsection-creation, and "released sections untouched". |
| Layer-boundary gate (`import_graph`) failure for the new package | Medium | Medium | Land the PROJECT.md G2 paragraph + `centinela.toml` aggregator `paths` mapping in Slice A; run `centinela validate` locally before completing each slice. Verify `evidence`/`verify` do not import back. |
| Brief/plan heading drift breaks section extraction | Low | Medium | Heading scan is tolerant (missing heading → omit section, never error). Unit tests cover briefs missing each heading. |
| Slow/fragile double test run if verification re-executes the suite at delivery | Medium | Medium | Drive gate status primarily from the static gatekeeper report; make the verification tally optional behind a short timeout and omit on error/timeout. |
| File-size (100-line) gate on new files | Low | Low | Pre-split: 7 small source files in `internal/delivery` + 1 cmd helper, each scoped to one concern. |

## Test Strategy Outline

- **Unit (`tests/unit/` + colocated `internal/delivery/*_test.go` for the 95%
  per-package gate):**
  - `extract.go`: heading found / missing / last-section / empty body.
  - `sections.go` + `prbody.go`: each section present, each section omitted when
    its datum is empty, provenance footer always present, full-body golden.
  - `changelog.go`: category selection per prefix (feat/fix/other), stub-vs-brief
    seeding, FILL-slot stub ignored.
  - `changelog_insert.go`: first insert, idempotent re-insert (returns `false`),
    subsection creation in canonical order, released sections untouched.
  Colocate small `_test.go` files in `internal/delivery` (each ≤ 100 lines per
  G1) so the per-package coverage gate is satisfied (a `tests/` tier file alone
  does not move it).
- **Integration (`tests/integration/`):** `InsertEntry` against a realistic
  `CHANGELOG.md` fixture (round-trip stability); `buildPRBody`/`writeChangelog`
  cmd helpers driven with the `git`/`gh` seams and a temp repo + temp
  `CHANGELOG.md` — assert the body file is created, `--body-file` is passed (not
  `--fill`), and the changelog gains exactly one line, idempotently on re-run.
- **Acceptance (`tests/acceptance/`):** drive the built binary `centinela deliver
  <feature> --via pr` in a temp git repo with a local bare `origin` (avoid real
  network — see memory: real network push hangs the suite) and a stubbed/faked
  `gh` on `PATH`; assert the PR-create invocation received a `--body-file`
  whose contents include the composed sections + provenance footer, and that
  `CHANGELOG.md` got one new `[Unreleased]` line. `validate.commands` must
  include the acceptance execution.

## Handoff

To **feature-specialist**: turn this into the concrete file-by-file plan and the
`specs/delivery-artifact-generation.feature` Gherkin. Hold the line on: pure
composer (no I/O in `internal/delivery`), gate-status omission over fabrication,
changelog idempotency scoped to `[Unreleased]`, the G2/import_graph mapping
landing in Slice A, and local-bare-`origin` for acceptance (no network push).
