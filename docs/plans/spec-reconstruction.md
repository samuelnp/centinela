# Plan: spec-reconstruction

> Consume the shipped `analyze.Inventory` (`.workflow/analysis.json`) to
> deterministically reconstruct a behavioral spec corpus: one
> `specs/<slug>.feature` Gherkin skeleton + one `docs/features/<slug>.md` brief
> stub per significant module/surface, with honest `# TODO: confirm` gaps. No
> LLM, no in-process inference — the same "deterministic skeleton + swappable
> seam" philosophy `synthesize` established. Reconciles big-thinker +
> feature-specialist.

## Layer decision (resolved)

`internal/reconstruct/` is an **aggregator** package (NOT domain): it imports the
domain package `internal/analyze` read-only, exactly like `internal/synthesize`
(and `internal/doctor`, `internal/insights`, `internal/calibration`,
`internal/audit`). A `reconstruct → analyze` edge is domain-from-aggregator
(allowed); `analyze` never imports `reconstruct`, so no cycle. It must not import
`cmd/` or `internal/ui`; it is imported only by `cmd/` (its `Reconstruction` type
by `internal/ui` for rendering). **centinela.toml**: add
`internal/reconstruct/**` to the existing `aggregator` layer `paths` and extend
the aggregator comment block. **PROJECT.md**: register `internal/reconstruct` in
the G2 prose (mirroring the `synthesize` sentence), and add it to the Folder
Structure block. `internal/ui/render_reconstruct.go` keeps `reconstruct` free of
any `ui` import. **Scaffold mirror**: the `internal/scaffold/assets/centinela.toml`
is a *generic* new-project template with **no** Centinela-specific aggregator
paths (verified: zero `internal/synthesize`/`aggregator` matches), so **no mirror
edit is required** — same as the `synthesize` slice.

## Selection model (deterministic)

`Select(inv) []Target` builds the sorted, de-duplicated target set from the
Inventory via a swappable rule table (mirroring `synthesize/rules.go` +
`signals.go`):

- **Signals** (`signals.go`, lowercased/flattened once): `inv.Packages`,
  `inv.Manifests` (kinds, frameworks, deps), `inv.Graph` (Go edges, may be
  empty), `inv.PrimaryLanguage`.
- **Promote rules** (`rules.go`, data table): a package becomes a target when it
  *owns behavior* — top-level/non-leaf packages, packages a graph edge points
  *into* (consumed surfaces), and manifest-declared command/endpoint surfaces.
  Each rule carries an inferred `Role` hint (`command | endpoint | module`).
- **Exclude rules** (data table): test-only (`_test`, `tests/`, `spec/`),
  generated (`.pb.go`, `node_modules`, `vendor/`, `dist/`, `gen/`), config-leaf
  (`config`, `internal/config`-style leaves with no graph in-edges), and
  vendored. Exclusion wins over promotion.
- **Bounding & ordering**: targets sorted by slug ascending; count capped at a
  constant (`maxTargets`, e.g. 50) so the review set stays reviewable. Slugs are
  derived from the package path and **collision-disambiguated deterministically**
  (append a path-derived suffix) so two packages never map to one file.
- `Selector` need not be a separate interface — selection is an internal step of
  the `Reconstructor`. The single public seam is `Reconstructor`.

## Reconstruction model (deterministic skeleton + swappable seam)

`Reconstructor` is the public interface, default `NewReconstructor()` →
`ruleReconstructor` (the drop-in seam for a future LLM backend, exactly like
`synthesize.NewInferer()`):

```go
type Reconstructor interface {
    Reconstruct(inv analyze.Inventory) Reconstruction
}
```

`Reconstruct` runs `Select`, then for each target assembles **pure strings** (no
I/O):

- **`specs/<slug>.feature`** — a `Feature:` line + a derived one-line narrative
  from the target's role, then **one or more `Scenario:` skeletons** whose
  Given/When/Then are explicit `# TODO: confirm` markers (never a fabricated
  concrete assertion). Role-aware: `command` → an invocation scenario;
  `endpoint` → a request/response scenario; `module` → a single behavior
  scenario. A target with no inferable behavior still yields a `Feature:` + one
  `# TODO: confirm` scenario stub (never empty, never assertion-bearing). Shape
  is verified to parse with the real `spec_traceability` parser (a `Feature:`
  line + ≥1 line matching `^\s+Scenario:`).
- **`docs/features/<slug>.md`** — a brief stub mirroring the
  `docs/features/*.md` shape (Problem / User value / TODO sections), every
  unknown a `# TODO: confirm`. Both artifacts are emitted per selected target.

The returned `Reconstruction` (typed result, the data model from the brief)
carries `Targets []Target` (sorted), `Written []string`, `Skipped []string`,
plus the per-skeleton `TodoCount`. `Reconstruct` is byte-stable: sorted targets,
stable section order, no map-iteration order.

## Output location & clobber policy (resolved)

- **Default output dir**: `.workflow/reconstructed/specs/` +
  `.workflow/reconstructed/features/` — a **review dir**, never clobbering
  hand-authored `specs/`. Flags `--in` (inventory path, default
  `analyze.DefaultOutPath`) and `--out` (review root, default
  `.workflow/reconstructed`) mirror `synthesize`.
- **Skip-if-exists**: any target whose **real** `specs/<slug>.feature` already
  exists in the repo's canonical `specs/` is **skipped** (recorded in `Skipped`),
  so a partially-spec'd repo is augmented, never overwritten — even though the
  default write target is the review dir (belt-and-suspenders against `--out
  specs`).
- **Re-run determinism**: writing to the review dir produces byte-identical files
  on an unchanged Inventory; a `WriteCorpus(outDir, recon)` helper writes each
  file via a single `os.WriteFile` (no partial files), `MkdirAll` the dirs,
  returns the written/skipped paths. Mirrors `synthesize/write.go`.

## Source files (each ≤100 lines, aggregator layer)

Under `internal/reconstruct/` unless noted:

1. `reconstruct.go` — package doc (aggregator contract, like `archetype.go`'s
   header) + result types: `Target`, `Reconstruction`.
2. `signals.go` — lowercased/flattened `signals` view over `analyze.Inventory`
   (`hasPkg`/`hasDep`/`hasFramework`/`hasKind`/graph in-edges), mirroring
   `synthesize/signals.go`.
3. `rules.go` — the promote/exclude rule table over `signals` (data, not control
   flow) + the `Role` hint each rule assigns. Split `predicates.go` only if >100
   lines.
4. `select.go` — `Select(inv) []Target`: apply rules, exclude, slugify,
   collision-disambiguate, sort, bound to `maxTargets`.
5. `slug.go` — deterministic `slugify(pkg)` + collision disambiguation (only if
   `select.go` would exceed 100 lines; otherwise fold in).
6. `reconstructor.go` — `Reconstructor` interface + `ruleReconstructor` +
   `NewReconstructor()`; `Reconstruct(inv)` orchestrator (Select → assemble →
   tally TodoCount).
7. `feature.go` — `featureSkeleton(t Target) string`: role-aware `Feature:` +
   `Scenario:` + `# TODO: confirm` Gherkin assembly (pure strings). Counts TODOs.
8. `brief.go` — `briefStub(t Target) string`: the `docs/features/<slug>.md` brief
   stub assembly (pure strings).
9. `templates.go` — the role-keyed skeleton/scenario templates as data tables
   (mirroring `synthesize/sections_*.go` + `profiles.go`), so adjusting skeleton
   shape is a table edit. Split into `templates.go` if `feature.go`+`brief.go`
   would exceed 100 lines each.
10. `write.go` — `WriteCorpus(outRoot string, r Reconstruction) (written,
    skipped []string, err error)`: skip-if-exists against canonical `specs/`,
    MkdirAll, single-call writes, byte-stable. Mirrors `synthesize/write.go`;
    holds the `DefaultOutRoot = ".workflow/reconstructed"` const.
11. `cmd/centinela/reconstruct.go` — thin Cobra command mirroring
    `synthesize.go`: `--in`/`--out`/`--json`; `analyze.Load` → on
    `analyze.ErrNoInventory` wrap "run `centinela analyze` first" + non-zero exit;
    `NewReconstructor().Reconstruct` → `WriteCorpus` → render summary. No business
    logic (G7).
12. `internal/ui/render_reconstruct.go` — `RenderReconstructionSummary(r
    Reconstruction) string`: targets selected, files written, files skipped,
    total TODO markers. Presentation only.

No change to `internal/analyze` (the `Load`/`ErrNoInventory` seam already exists
from the `synthesize` slice — reused as-is).

Config / docs changes:
- `centinela.toml`: `aggregator` layer `paths` += `internal/reconstruct/**`;
  extend the aggregator comment block with the `reconstruct → analyze` rationale.
- `PROJECT.md`: G2 prose sentence registering `internal/reconstruct` as an
  aggregator; Folder Structure entry.
- Scaffold mirror: **none** (generic template carries no aggregator paths).

## Test plan (per-package 95% → colocated `_test.go`, each ≤100)

- **Unit (`internal/reconstruct`):**
  - `select_test.go` — fixture Inventories → expected sorted Targets:
    Go n-tier (handler/service/repo → module/command targets), polyglot
    (empty Go graph, manifest-only), empty/doc-only → 0 targets, slug-collision
    fixture → disambiguated, huge package list → bounded to `maxTargets`,
    exclusion (test/generated/vendored/config-leaf) precedence.
  - `rules_test.go` / `signals_test.go` — predicate + role-hint coverage.
  - `feature_test.go` — every generated `.feature` parses with the real
    `spec_traceability` parser (`Feature:` + ≥1 `Scenario:`); `# TODO: confirm`
    present; no fabricated concrete Given/When/Then; one small golden fragment.
  - `brief_test.go` — brief stub shape + TODO markers.
  - `reconstructor_test.go` — `Reconstruct` end-to-end on a fixture:
    determinism (two runs byte-identical), TodoCount tally, role-aware scenarios.
  - `write_test.go` — skip-if-exists (canonical `specs/<slug>.feature` present →
    `Skipped`), review-dir write, re-run byte-identical, no-partial-file on error.
- **Command:** `cmd/centinela/reconstruct_test.go` (happy path, `--json`) +
  `reconstruct_errors_test.go` (no inventory → `ErrNoInventory` message,
  non-zero exit, no files written).
- **Integration:** `tests/integration/reconstruct_pipeline_test.go` — real
  analyze → `Save` → `Load` → `Reconstruct` → `WriteCorpus` on a committed Go
  fixture repo; asserts files land in the review dir and re-run is clean.
- **Acceptance:** `tests/acceptance/reconstruct_{helper,happy,edge}_test.go` —
  build the binary once, run `analyze` then `reconstruct` over fixture repos;
  carry `// Acceptance:` / `// Scenario:` traceability comments for **all** spec
  scenarios so the `spec_traceability` gate maps each one. `.feature` files in
  the review dir are also parsed by the real parser to prove Gherkin validity.
  `validate.commands` must include the acceptance run.

All fixtures are committed / in-test `Inventory` values — no LLM, no network.

## Spec

`specs/spec-reconstruction.feature` — scenarios covering the 9 acceptance
criteria / edge cases:
1. Valid inventory → ≥1 `.feature` + ≥1 brief stub written to the review dir.
2. Every generated `.feature` parses with the `spec_traceability` parser.
3. Unknowable behavior → explicit `# TODO: confirm`, never fabricated.
4. Deterministic re-run → byte-identical output.
5. Existing hand-authored `specs/<slug>.feature` → skipped, reported, never
   overwritten.
6. Missing/old inventory → `ErrNoInventory` "run `centinela analyze` first",
   non-zero exit, no files written.
7. Concise stdout summary: targets / files written / TODO count.
8. (Empty/doc-only inventory → 0 targets, exit 0, no empty `.feature`.)
9. (Polyglot inventory → still selects manifest/package targets.)

## Risks

- **Thin skeletons add little value** → rule table picks *meaningful* targets +
  role-aware scenario stubs; the LLM seam carries the behavioral lift.
- **Clobbering hand-authored specs** → review-dir default + skip-if-exists, with
  an explicit acceptance test.
- **Layering / import-graph regression** → mapped as aggregator in
  `centinela.toml` + PROJECT.md before code; sole edge `reconstruct → analyze`.
- **Gherkin shape drift** → an acceptance/unit test parses every generated file
  with the *real* `spec_traceability` parser.
- **Scope creep on routes/flows** → v1 ships package/manifest-derived targets;
  framework-specific HTTP route / call-flow extraction is **deferred** to the
  roadmap (`brownfield-route-flow-extraction`).
- **Per-package 95% coverage** → pure string-assembly functions + fixture
  Inventories; ~12 source + ~12 test files, splits pre-planned to stay ≤100.
</content>
</invoke>
