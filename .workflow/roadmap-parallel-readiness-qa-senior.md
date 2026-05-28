# QA Senior — Tests Validation

- Feature: roadmap-parallel-readiness
- Step: tests
- Role: qa-senior
- Status: done

## Inputs
- specs/roadmap-parallel-readiness.feature
- docs/plans/roadmap-parallel-readiness.md
- docs/features/roadmap-parallel-readiness.md
- .workflow/roadmap-parallel-readiness-senior-engineer.md
- internal/roadmap/readiness.go
- internal/roadmap/dependencies.go
- internal/roadmap/roadmap.go
- internal/ui/render_readiness.go
- internal/ui/render_roadmap.go
- internal/ui/render_session.go
- cmd/centinela/roadmap_ready.go
- cmd/centinela/start_guard.go
- cmd/centinela/hook_session.go

## Test Inventory

| Tier | File | Scenarios covered |
|------|------|-------------------|
| unit | tests/unit/roadmap_parallel_readiness_unit_test.go | DeriveReadiness four states + in-progress-dep-blocks + multi-unmet BlockedBy; diamond ready vs blocked-by-c; ReadySet declared order; UnmetDependencies; RenderRoadmap markers per state (🔓/🔒/done/in-progress); RenderReadyList populated + empty (6 scenarios) |
| unit (colocated) | internal/roadmap/readiness_test.go | DeriveReadiness/classifyFeature/collectUnmet states + nil; ReadySet declared order; UnmetDependencies satisfied/none/missing/nil (3 funcs) — pins internal/roadmap package coverage |
| unit (colocated) | internal/ui/render_readiness_test.go | roadmapIcon all branches; readinessMarker icon+annotation per state; RenderReadyList branches; renderReadyBlock ready/complete/blocked branches (4 funcs) |
| integration | tests/integration/roadmap_parallel_readiness_integration_test.go | roadmap.Load + RenderRoadmap annotated markers + plural frontier end-to-end; empty-frontier-but-blocked not-complete; all-done complete (3 scenarios) |
| acceptance | tests/acceptance/roadmap_parallel_readiness_test.go (+_scenarios_test.go, +_guard_test.go) | binary `roadmap ready` (ready set / empty-state / all-done exit 0); `roadmap` render 🔓/🔒 + dep names + done has no marker; `start` guard refused for planned/in-progress dep + proceeds for done/no-deps; load-time rejection of unknown / self / 2-node / 3-node cycle (10+ scenarios) |
| unit (colocated cmd) | cmd/centinela/roadmap_ready_test.go | runRoadmapReady ready frontier / empty-state / missing-roadmap error — pins cmd coverage |
| unit (colocated cmd) | cmd/centinela/start_guard_dependency_test.go | checkDependencyGuard blocked (names feature + all unmet deps) / no-deps allowed / all-done allowed |

All 28 .feature scenarios are mapped to executable assertions across the tiers
above (schema/back-compat, validation negatives, readiness derivation, the
`roadmap ready` command, the start guard, render markers, plural rehydration —
the rehydration scenarios are also pinned by the existing
cmd/centinela/hook_session_test.go and tests/unit/session_context_render_unit_test.go).

## Coverage Gaps
None. Every function flagged under-covered is now at 100% statement coverage:
DeriveReadiness, ReadySet, UnmetDependencies, classifyFeature, collectUnmet
(internal/roadmap/readiness.go); runRoadmapReady (cmd/centinela/roadmap_ready.go);
checkDependencyGuard (cmd/centinela/start_guard.go); roadmapIcon, readinessMarker,
RenderReadyList (internal/ui/render_readiness.go); renderReadyBlock
(internal/ui/render_session.go). Total coverage gate: 95.1% >= 95.0%.

Note on measurement: `scripts/check-coverage.sh` runs `go test ./...` with a
single per-package `-coverprofile` (no `-coverpkg`), so package coverage is only
credited by tests in that same package. The internal/roadmap, internal/ui and
cmd/centinela functions are therefore pinned by colocated `_test.go` files; the
tests/{unit,integration,acceptance} tiers provide the mandated cross-layer and
end-to-end behavioral coverage. All colocated `_test.go` files are <=100 lines.

## Acceptance Wiring
`centinela.toml` `[validate].commands` runs `go test ./...`, which executes
`tests/acceptance/...`. The acceptance suite builds the real binary
(`go build -o <tmp>/cent ./cmd/centinela`) and drives it against temp
roadmap.json + workflow fixtures, so the gatekeeper/validate step re-runs the
acceptance scenarios on every `centinela validate`.

## Handoff
- handoffTo: validation-specialist
