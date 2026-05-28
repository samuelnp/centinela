# Senior-Engineer Report: roadmap-parallel-readiness

## Corrective pass (final state)

A prior pass left the suite RED and deviated from the plan. This pass fixed both:

- **FIX 1 — pure render.** `internal/ui/render_session.go` `RenderSessionRehydration`
  now has the plan signature `(r *roadmap.Roadmap, ready []string, hasIncomplete bool)`
  and uses the passed `ready`/`hasIncomplete` directly. The internal `roadmap.ReadySet(r)`
  + `r.Summary()` calls were removed, so the UI layer no longer derives state
  (no disk/workflow access). `cmd/centinela/hook_session.go` already computed
  `ready` + `hasIncomplete`; it now passes them through and the dead `next`
  variable/hack was deleted.

- **FIX 2 — tests faithfully updated to the new contracts** (no assertions weakened):
  - `internal/ui/render_session_test.go`, `tests/unit/session_context_render_unit_test.go`:
    new pure signature; planned no-dep features now annotate `(ready)`; assert the
    plural "Ready to start now:" block, one `docs/features/<f>.md` pointer per ready
    feature; complete state (nil ready, !hasIncomplete) keeps "Roadmap complete".
  - `cmd/centinela/hook_session_test.go`, `tests/acceptance/session_context_rehydration_test.go`:
    hook now emits the plural ready frontier; the across-phases test was renamed
    `...ListsReadyFrontier` and asserts the frontier header + ready feature + pointer.
  - `internal/roadmap/analysis_test.go`: `ValidateAnalysis` tests now assert the new
    contract (role + coverage + analysis-references-unknown-feature only). The
    cycle/unknown/self-dep coverage that moved to `ValidateDependencies` was relocated
    into the new `internal/roadmap/dependencies_test.go` (2-node cycle, self-dep
    1-node cycle, unknown slug, valid + nil graphs). Net: no loss of real coverage.
  - `internal/planadvisor/roadmap_context_test.go`, `tests/unit/...`, `tests/integration/...`:
    `dependencyNames` reads deps from `roadmap.json`; fixtures now declare
    `dependsOn:["dep"]` on the roadmap.json feature, keeping the same assertions.

Result: `go test ./...` → 699 passed, 0 failed (21 packages). `go vet ./...` clean.
`internal/roadmap/analysis_test.go` was split (the relocated dep test went to
`dependencies_test.go`) to stay ≤100 lines.

## What was implemented per plan step

### Step 1 — Schema: `Feature.DependsOn`
Added `DependsOn []string \`json:"dependsOn,omitempty"\`` to `internal/roadmap/roadmap.go` `Feature` struct. `omitempty` keeps existing roadmap.json byte-identical when no deps are set. Added `ValidateDependencies(&r)` call in `Load()` after unmarshal.

### Step 2 — Option B validation move
- Created `internal/roadmap/dependencies.go` with `ValidateDependencies(r *Roadmap) error`. Builds the feature name set via `roadmapFeatureSet`, checks each dep is known, then calls `hasCycle(deps)` (reusing `analysis_cycle.go` unchanged).
- Rewrote `internal/roadmap/analysis.go` to drop `DependsOn` from `AnalysisFeature` and remove the dep/cycle checks from `ValidateAnalysis`. Only role, markdown presence, and feature-coverage checks remain.

### Step 3 — Read-path switches
- `internal/planadvisor/roadmap_context.go`: `dependencyNames` now loads `roadmap.Load()` and walks phases to find `f.DependsOn` for the named feature. Dropped `encoding/json` and `os` imports, dropped `roadmap.Analysis` read-path.
- `internal/docgen/load.go`: `loadRoadmapNodes` now decodes `roadmap.json` phases with `DependsOn` inline struct and populates `RoadmapNode{Name, DependsOn}`. Removed the analysis-json fallback path.

### Step 4 — Readiness derivation
Created `internal/roadmap/readiness.go` with:
- `FeatureReadiness{Name, State, BlockedBy}` struct (never persisted)
- `DeriveReadiness(r *Roadmap) []FeatureReadiness` — single pass, classifies each feature
- `ReadySet(r *Roadmap) []string` — names where `State == "ready"` in declared order
- `UnmetDependencies(r *Roadmap, feature string) []string` — for the start guard
- `classifyFeature` and `collectUnmet` private helpers

### Step 5 — Surfaces: render markers + rehydration
- `internal/ui/styles.go`: Added `IconReady = "🔓"` and `IconBlocked = "🔒"`.
- Created `internal/ui/render_readiness.go` with `readinessMarker(fr FeatureReadiness)`, `roadmapIcon(status)` (compat helper for existing tests), and `RenderReadyList(ready []string)`.
- `internal/ui/render_roadmap.go`: Rewrote `RenderRoadmap` to use `DeriveReadiness` + `readinessMarker`; updated `RenderRoadmapNeeded` scaffold text to show `dependsOn` example. Removed old `roadmapIcon` (moved to `render_readiness.go`).
- `internal/ui/render_session.go`: `RenderSessionRehydration(r *roadmap.Roadmap, ready []string, hasIncomplete bool)` is now a PURE render — it uses the passed `ready`/`hasIncomplete` directly (no `ReadySet`/`Summary` calls in the UI layer). It shows the plural "Ready to start now:" block, or "Roadmap complete." when the frontier is empty and nothing is incomplete, or the "everything in-progress or blocked" message when work remains but nothing is ready.

### Step 6 — Capabilities: `roadmap ready` + start guard
- Created `cmd/centinela/roadmap_ready.go`: `roadmapReadyCmd` subcommand under `roadmapCmd`, calls `roadmap.ReadySet(r)`, prints via `ui.RenderReadyList`. Exit 0 always.
- `cmd/centinela/start_guard.go`: Added `checkDependencyGuard(r, feature)` call after bootstrap checks; extracted as a helper function. Calls `roadmap.UnmetDependencies(r, feature)` and returns a clear error naming blocked deps. Logic stays in `internal/roadmap`, cmd is a thin orchestrator.
- `cmd/centinela/hook_session.go`: Replaced `FirstIncomplete` with `ReadySet`; computes `hasIncomplete` from `r.Summary()`; passes `(r, ready, hasIncomplete)` to the now-pure `RenderSessionRehydration` (the old `next` variable/hack was removed).

### Step 7 — Scaffold instruction text
Updated `RenderRoadmapNeeded` in `render_roadmap.go` to show `dependsOn` in the example JSON and add an explanatory line about the optional field.

## Files created

- `internal/roadmap/dependencies.go`
- `internal/roadmap/readiness.go`
- `internal/ui/render_readiness.go`
- `cmd/centinela/roadmap_ready.go`

## Files modified

- `internal/roadmap/roadmap.go` — schema + Load validation
- `internal/roadmap/analysis.go` — Option B: drop DependsOn + cycle checks
- `internal/roadmap/roadmap_test.go` — fix positional struct literal after DependsOn field added
- `internal/ui/styles.go` — add IconReady + IconBlocked
- `internal/ui/render_roadmap.go` — markers + scaffold example
- `internal/ui/render_session.go` — plural rehydration, compat signature
- `internal/ui/render_session_test.go` — update to match new output strings
- `internal/ui/render_other_test.go` — replace `roadmapIcon("done")` call with `IconDone`
- `cmd/centinela/hook_session.go` — pass ready set
- `cmd/centinela/start_guard.go` — dependency guard
- `internal/planadvisor/roadmap_context.go` — read deps from roadmap.json
- `internal/docgen/load.go` — graph edges from roadmap.json

## Option B ripple handling

`AnalysisFeature.DependsOn` was removed. Ripple effects:
- `analysis.go` itself: removed dep/cycle logic (Option B intent)
- `planadvisor/roadmap_context.go`: was reading `a.Features[].DependsOn` from analysis — switched to `roadmap.Load()` + `f.DependsOn`
- `docgen/load.go`: was using analysis-json as primary dep source — now reads roadmap.json directly
- `roadmap_test.go`: positional `Feature{{"done"}}` literal broke after adding `DependsOn` field — changed to named fields
- `render_other_test.go` and `render_zero_test.go`: referenced unexported `roadmapIcon` removed from `render_roadmap.go` — restored in `render_readiness.go` as compat helper

The refactor-invalidated tests (`tests/unit/session_context_render_unit_test.go`,
acceptance, the analysis/planadvisor tests) were repaired in this corrective pass
to the new contracts — see the "Corrective pass" section above. qa-senior will
author the brand-new feature-scenario tests in the tests step.

## Build & test results (final)

```
go build ./...   → EXIT: 0 (clean)
go vet ./...     → Go vet: No issues found
go test ./...    → Go test: 699 passed in 21 packages (0 failed)
```
