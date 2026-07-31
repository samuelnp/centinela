# dynamic-routing-clock-skew — qa-senior

## Test Inventory

| Tier | File | Scenarios |
|------|------|-----------|
| unit (colocated) | `internal/workflow/model_routes_clock_skew_test.go` | coarse-mtime skew tolerated; earlier-run stub still rejected; grace boundary both sides |
| unit (tier) | `tests/unit/dynamic_model_routing_underway_unit_test.go` | pre-existing `...JSONStubFromThisRunCounts` — was CI-red, now passes on both platforms (unchanged) |
| integration/acceptance | unchanged | hotfix touches one internal predicate; no CLI surface, config key, or spec scenario changed |

Colocated placement is deliberate: coverage is measured per package, and the
tier directories do not move `internal/workflow`'s number.

## Red → green evidence

The fix was reverted in place (`cutoff := wf.StartedAt`) and the suite re-run:
`ToleratesCoarseMtimeSkew` and `GraceWindowBoundary` both FAILED on macOS, where
the original tests could not fail. The fix was restored and both passed. That is
the whole point of forcing the mtime with `os.Chtimes` instead of writing a file
and hoping the host's clock granularity exposes the ordering.

## Coverage Gaps

None for the defect. The residual risks (a >2s backward clock step, exotic
filesystem mtime rounding) are documented in the edge-case report and are not
reachable by a test that would be meaningful rather than tautological.

## Acceptance Wiring

`centinela.toml` `validate.commands` runs `go test ./... -coverprofile=coverage.out`,
which covers `tests/acceptance`, `tests/integration`, `tests/unit` and every
colocated package test in one profiled run. Unchanged by this hotfix.

## Regression Guards

The three new tests exist specifically so this cannot recur: any future change
that compares an artifact mtime directly against `StartedAt` fails
`ToleratesCoarseMtimeSkew`, and any attempt to "fix" that by widening the window
without bound fails `StillRejectsAnEarlierRunsStub`.

## Deferred Findings

none

## Handoff

- Next role: validation-specialist (gatekeeper runs the validate step).
- Note for the verifier: this branch carries all of `dynamic-model-routing` plus
  one commit; the defect it repairs was CI-only and green on macOS, so a local
  suite run alone is NOT sufficient evidence — the deterministic tests are.
