# Plan: brownfield-roadmap-generation

> Consume the frozen `analyze.Inventory` + the in-process `reconstruct.Reconstruction`
> to emit a **draft roadmap** that records already-built capability as a
> **Baseline** phase (never re-planned) and net-new work / gaps as schedulable
> features — deterministically (no LLM), never clobbering a real `roadmap.json`.
> Reconciles big-thinker + feature-specialist.

## Layer decision (resolved)

`internal/brownmap` is an **aggregator** package (NOT domain): it imports the
domain package `internal/analyze` (read-only, via `analyze.Load`), the sibling
aggregator `internal/reconstruct` (read-only, for `NewReconstructor()` +
`Reconstruction`/`Target`), and the domain package `internal/roadmap` (read-only,
for the `Roadmap`/`Phase`/`Feature` types + the `BaselinePhaseName` convention).
It must not import `cmd/` or `internal/ui`; it is imported only by `cmd/` and its
result type by `internal/ui` for rendering. `analyze`, `reconstruct`, and
`roadmap` never import `brownmap`, so there is no cycle.

**Import-graph wrinkle (must land with the code):** the `aggregator` layer in
`centinela.toml` currently `allow = ["domain","leaf"]`. `brownmap → analyze` and
`brownmap → roadmap` are domain edges (already allowed). `brownmap → reconstruct`
is **aggregator→aggregator**, which is not yet permitted. Resolution: add
`"aggregator"` to the aggregator layer's `allow` so aggregators may compose. This
is the cleanest option and is documented in PROJECT.md G2. (Rejected alternative:
duplicate reconstruct's target-selection rule table inside `brownmap` to avoid the
edge — needless duplication.)

**centinela.toml:** `aggregator` layer `paths` += `internal/brownmap/**`; the same
layer's `allow` += `"aggregator"`. **PROJECT.md:** register `internal/brownmap` in
the G2 prose + folder structure + layer/gatekeeper tables, stating its three
read-only edges, the new `allow` entry, and the no-cycle argument. Mirror the toml
into `internal/scaffold/assets` only if the parity test covers that toml.
`internal/ui/render_brownfield.go` keeps `brownmap` free of any `ui` import.

## Baseline-vs-gap representation (the central decision — resolved)

**Use a dedicated `Baseline` phase, identified by a phase-name convention**,
exactly mirroring the existing `Backlog` and `Phase 0: Bootstrap` conventions in
`internal/roadmap`. Rationale:

- Status is **derived** from workflow state, never stored — so "already-built =
  done" cannot be set via a field. The existing `Backlog` phase already solves the
  isomorphic problem ("these entries are not schedulable work") purely with a
  phase-name predicate (`isBacklogPhaseName`) wired into `Summary`, the
  `NonBacklogFeatureSet` coverage set, `readiness`, and the render skip — **one
  place each**. Baseline reuses that exact mechanism.
- No change to `roadmap.json`'s schema or the frozen `types.go`. Baseline entries
  are real `Feature`s in a real `Phase`; they simply live under a name the
  exemption predicate recognizes.
- Reversible / honest: the user sees a clearly-labelled Baseline phase in
  `ROADMAP.md` and can edit it; nothing fabricates fake `.workflow/*.json` stubs.

**Rejected alternatives:** (a) a new `Feature.Baseline bool` field — touches the
frozen schema and every consumer; (b) generating completed workflow-state stubs —
fabricates state, pollutes `.workflow/`, irreversible; (c) a `Note`/marker string
convention on features — not enforceable in the exemption logic without parsing
prose.

Implementation: add `BaselinePhaseName` constant + `isBaselinePhaseName` /
`IsBaselinePhaseName` in `internal/roadmap` (new small file `baseline.go`,
mirroring `backlog.go`), and wire the predicate into the **same** four sites that
already skip Backlog (`Summary`, the non-schedulable coverage set, `readiness`,
the phase render). Each edit is additive (an `|| isBaselinePhaseName(...)`).

## Generation model (deterministic)

`internal/brownmap`:

```go
type Plan struct {              // typed, byte-stable result
    Roadmap      roadmap.Roadmap // Baseline phase + gap phase(s)
    BaselineCount int
    GapCount      int
    DraftPath     string         // where it was/will be written
}

type Brownfielder interface {   // swap seam for a future LLM backend
    Generate(inv analyze.Inventory, goals []string) Plan
}
func NewBrownfielder() Brownfielder // deterministic default
```

`Generate`:
1. `r := reconstruct.NewReconstructor().Reconstruct(inv)` — sorted `Targets`,
   per-target TODO signal (`r.TodoCount` overall; per-target via the skeleton — if
   per-target TODO counts aren't exposed, the feature-specialist adds a tiny
   accessor to reconstruct, NOT a logic dup).
2. Build the **Baseline** `Phase` (name = `BaselinePhaseName`): one
   `roadmap.Feature` per target — `Name` from target slug, `Description` from its
   role/reason, `Source = {feature: "brownfield-roadmap-generation", role:
   "big-thinker"}`-style provenance.
3. Build **gap** `Phase`(s): one `Feature` per TODO-bearing target (net-new work
   to confirm) + one per `--goal` string. `Fixes` carries the gap rationale.
4. Assemble `roadmap.Roadmap{Intro, Phases: [Baseline, Gaps...]}`. Pure
   struct/string assembly over already-sorted inputs → byte-stable.

## Source files (each ≤100 lines, aggregator layer unless noted)

1. `internal/brownmap/brownmap.go` — package doc + `Plan`/`Brownfielder` types +
   `NewBrownfielder`.
2. `internal/brownmap/generate.go` — `Generate(inv, goals) Plan` orchestrator.
3. `internal/brownmap/baseline.go` — build the Baseline phase from `Targets`.
4. `internal/brownmap/gaps.go` — build the gap phase(s) from TODO targets + goals.
5. `internal/brownmap/write.go` — `WriteDraft(path, Plan)`: never overwrite the
   canonical `roadmap.RoadmapFile`; write the draft path; report path + whether a
   real roadmap was left untouched. (Atomic temp-file+rename like roadmap's
   `writeAtomic`.)
6. `internal/roadmap/baseline.go` — `BaselinePhaseName` const +
   `isBaselinePhaseName`/`IsBaselinePhaseName` (mirror `backlog.go`).
7. `cmd/centinela/roadmap_brownfield.go` — thin Cobra subcommand under
   `roadmapCmd`; flags `--in` (analysis.json), `--out` (draft path), `--json`,
   `--goal` (repeatable); load → generate → write → render; actionable
   `ErrNoInventory` error; no business logic (G7).
8. `internal/ui/render_brownfield.go` — `RenderBrownfieldSummary(Plan)`:
   baseline count, gap count, draft path, "no gaps — supply --goal" hint.

Wiring edits (additive): the four `isBacklogPhaseName` skip sites in
`internal/roadmap` (`roadmap.go` Summary, `backlog.go` NonBacklog set / or a new
shared "non-schedulable" predicate, `readiness.go`, `mdgen_phase.go`/`rawrender.go`
render skip) gain an `|| isBaselinePhaseName(...)`. Prefer a single shared
`isNonSchedulablePhase(name)` helper so future conventions add in one place.

## cmd wiring & ui rendering

- `cmd/centinela/roadmap_brownfield.go` mirrors `cmd/centinela/reconstruct.go`:
  `analyze.Load(--in)` → `ErrNoInventory` guard → `brownmap.NewBrownfielder().
  Generate(inv, goals)` → `brownmap.WriteDraft(--out, plan)` → if `--json`, encode
  `plan`; else `fmt.Fprintln(out, ui.RenderBrownfieldSummary(plan))`.
- Render is presentation-only; `brownmap` returns the typed `Plan`, `ui` formats.

## PROJECT.md G2 edit needed

Append to the G2 prose (line 26), after the `reconstruct` sentence:

> `internal/brownmap` also joins the **aggregator** layer: a read-only brownfield
> roadmap generator for `centinela roadmap brownfield` that may import
> `internal/analyze` (domain) + `internal/roadmap` (domain) + `internal/reconstruct`
> (aggregator) read-only and stdlib only; it must not import `cmd/` or
> `internal/ui` and is imported only by `cmd/` (its `Plan` type by `internal/ui`
> for rendering). Its edges `brownmap → analyze`/`roadmap`/`reconstruct` are all
> allowed (the aggregator layer's `allow` is extended to `["domain","leaf",
> "aggregator"]` so aggregators may compose); `analyze`/`roadmap`/`reconstruct`
> never import `brownmap`, so there is no cycle.

Plus the folder-structure tree row and the layer/gatekeeper tables, and the
`centinela.toml` `[[gates.import_graph.layers]] name="aggregator"` `paths` +
`allow` edits described above.

## Test plan (per-package 95% → colocated `_test.go`, each ≤100)

- **Unit (`internal/brownmap`):** generate_test (fixture Inventory → Baseline phase
  with N entries + gap phase from TODOs + `--goal`; deterministic byte-stable
  order; empty inventory → empty Baseline, 0 gaps), baseline_test, gaps_test,
  write_test (never clobbers `roadmap.RoadmapFile`; draft path written; atomic).
- **Unit (`internal/roadmap`):** baseline_test (`isBaselinePhaseName` matches /
  rejects; `Summary` excludes Baseline; coverage set excludes Baseline; render
  skip). Guard the existing Backlog/Bootstrap behavior is unchanged.
- **Command:** `cmd/centinela/roadmap_brownfield_test.go` +
  `roadmap_brownfield_errors_test.go` (missing inventory → actionable error).
- **Integration:** `tests/integration/brownfield_pipeline_test.go` — real
  analyze→Save→Load→Generate→WriteDraft on a Go fixture, asserting an existing
  `roadmap.json` is byte-unchanged.
- **Acceptance:** `tests/acceptance/brownfield_{helper,happy,edge}_test.go` —
  build binary once, run `analyze` then `roadmap brownfield` over fixture repos;
  carry `// Acceptance:`/`// Scenario:` traceability for all spec scenarios.

All fixtures are committed/in-test `Inventory` values — no LLM, no network.

## Spec

`specs/brownfield-roadmap-generation.feature` — scenarios: built repo → Baseline
phase + gap phase; Baseline excluded from status/coverage; existing roadmap.json
not clobbered; deterministic re-run byte-identical; missing analysis.json →
actionable error; `--goal` adds a net-new gap feature; no TODOs/no goals →
Baseline-only with hint.

## Risks

- **Aggregator→aggregator edge trips `import_graph`** → mitigated by the
  `allow += "aggregator"` toml edit + PROJECT.md G2, landing with the code.
- **Clobbering a curated `roadmap.json`** → draft-only writer + acceptance test.
- **Baseline leaks into status/coverage** → single shared non-schedulable
  predicate, tested against Backlog/Bootstrap regressions.
- **≤100-line rule** → ~8 source + ~10 test files, splits pre-planned.
- **Per-package 95% coverage** → pure functions over fixture `Inventory` values.
