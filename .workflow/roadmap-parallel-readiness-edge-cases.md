# Edge Cases — roadmap-parallel-readiness (tests step)

Hard paths in the dependency graph / readiness derivation / start guard, and the
test that guards each. "covered at 100%" refers to per-function statement
coverage in `coverage.out` after `go test ./...`.

## Validation — negative paths (rejected at `roadmap.Load`)

- **Dependency cycle (2-node A↔B).** `ValidateDependencies` → `hasCycle` must
  reject. Guarded by `internal/roadmap` `dependencies_test.go` (existing) and
  end-to-end by `tests/acceptance/roadmap_parallel_readiness_guard_test.go`
  `TestAcceptance_LoadRejections/two-node` — `centinela roadmap ready` exits
  non-zero and the error mentions "cycle".
- **Self-dependency (A→A) treated as a 1-node cycle.** Acceptance
  `TestAcceptance_LoadRejections/self` asserts non-zero exit + "cycle".
- **Longer cycle (A→C→B→A).** Acceptance `.../three-node` asserts the same. This
  proves cycle detection is not limited to direct back-edges.
- **Unknown dependency slug.** `ValidateDependencies` rejects a `dependsOn` naming
  a feature absent from the roadmap. Acceptance `.../unknown` asserts the error
  names the offending slug ("ghost").

## Readiness derivation — the subtle classification rules

- **Dependency in-progress (NOT done) keeps the dependent blocked.** The most
  error-prone rule: only `status == "done"` satisfies a dependency; "in-progress"
  must not. Guarded at unit level (`internal/roadmap/readiness_test.go`
  `TestDeriveReadiness_States` seeds dep `w` at step "code" → dependent `blk`
  stays blocked and lists `w`), at integration level
  (`TestIntegration_EmptyFrontierBlockedNotComplete` makes `base` in-progress),
  and via the start guard (`TestAcceptance_StartGuard` in-progress case refuses).
- **Multiple unmet deps are all listed in BlockedBy.** `collectUnmet` must collect
  every non-done dep, not short-circuit on the first. `TestDeriveReadiness_States`
  asserts both `w` and `p` appear in `blk.BlockedBy`;
  `TestCheckDependencyGuard_BlockedAndAllowed` asserts the guard error names both
  `dep-a` and `dep-b`.
- **Diamond partial completion (A→B,C ; D→B,C).** D must be ready ONLY when both
  B and C are done; with B done but C planned, D stays blocked and BlockedBy ==
  [c] (not [b]). Guarded by `tests/unit/...unit_test.go` `TestDeriveReadiness_Diamond`
  (both the ready and the blocked-by-c branch).
- **Cross-phase dependency.** A Phase 1 feature depending on a Phase 0 feature is
  classified across phases. `tests/integration/...integration_test.go`
  `TestIntegration_RoadmapRenderAnnotatesAndListsFrontier` uses a roadmap whose
  frontier and dependents span the same phase; the cmd-level
  `TestRunHookSession_ListsReadyFrontier` (existing) crosses Phase 0→Phase 1.

## Empty-frontier disambiguation (looks-complete vs is-complete)

- **Empty frontier with work remaining must NOT look complete.** When nothing is
  ready but planned/in-progress work remains, the rehydration must avoid the
  "Roadmap complete" message and surface the blocking reason instead.
  `TestIntegration_EmptyFrontierBlockedNotComplete` asserts the complete message
  is absent and the blocking reason is present. `renderReadyBlock(nil, true)` is
  unit-pinned in `internal/ui/render_readiness_test.go`.
- **All-done is genuinely complete.** `TestIntegration_AllDoneShowsComplete` and
  acceptance `TestAcceptance_RoadmapReadyEmptyStates` (all-done case) assert the
  complete state with no ready frontier and a clean exit 0.
- **Empty ready set still prints a non-empty empty-state line** (no feature names,
  no 🔓). `TestRenderReadyList_Branches` and the acceptance empty-state case.

## Backward compatibility

- **No-deps roadmap behaves exactly as before.** A roadmap where no feature
  declares `dependsOn` loads cleanly and every planned feature is "ready".
  Guarded implicitly by the no-dep features in `TestRunRoadmapReady_*` and by
  the `start` no-dep proceed case in `TestAcceptance_StartGuard`.
- **Empty `dependsOn: []` == absent field.** `collectUnmet(nil/[])` returns no
  unmet deps → feature classified "ready"; covered by `classifyFeature`'s default
  branch reaching 100% in `TestDeriveReadiness_States`.
