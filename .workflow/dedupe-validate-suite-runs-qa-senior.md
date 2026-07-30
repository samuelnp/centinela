# dedupe-validate-suite-runs — qa-senior

### QA-Senior Report: dedupe-validate-suite-runs
**Date:** 2026-07-29

## Test Inventory

| Tier        | File | Scenarios |
|-------------|------|-----------|
| unit        | tests/unit/coverage_reuse_branch_test.go (new) | reuse skips the internal suite (failing test proves the skip); missing named profile → fail-safe self-run fails on the failing test |
| unit        | tests/unit/coverage_reuse_helper_test.go (new) | hermetic temp-Go-module scaffolding helpers (no network, GOFLAGS cleared) |
| unit (colocated) | cmd/centinela/complete_validation_outcome_test.go (new) | completedValidationOutcome(): ExitCode 0, nil StartErr, no TimedOut, non-empty Output naming the verified tree |
| integration | tests/integration/coverage_validate_config_test.go (rewritten) | validate.commands contains the coverage script AND the profile-pairing invariant (-coverprofile=coverage.out ↔ COVERAGE_PROFILE=coverage.out) |
| integration | tests/integration/ci_validate_workflow_integration_test.go (rewritten) | validate.yml keeps `go run ./cmd/centinela validate`, drops the bare `go test ./...` expectation, adds the no-bare-suite-step negative assertion |
| acceptance  | tests/acceptance/dedupe_validate_suite_runs_config_test.go (new) | "validate runs the suite exactly once"; "CI runs the suite once via centinela validate" |
| acceptance  | tests/acceptance/dedupe_validate_suite_runs_prompts_test.go (new) | "gatekeeper prompt copies stay identical and carry the conditional suite mandate"; "qa-senior prompt copies carry the one-full-run guidance" (byte-parity + wording) |
| acceptance  | tests/acceptance/dedupe_validate_suite_runs_script_test.go (new) | "bare check-coverage.sh invocation stays self-contained"; "COVERAGE_VALUE fast path bypasses the suite unchanged" (behavioral, empty dir); reuse/fallback branch-condition text pins |
| acceptance  | tests/acceptance/dedupe_validate_suite_runs_complete_test.go (new) | "complete's claim verification reuses the gate's own run"; "a failing validate command blocks completion before any reuse"; "a tree change during the gate's suite run blocks completion"; "standalone verify surfaces still re-run the suite for real" |

## Coverage Gaps

- None — all 12 scenarios in specs/dedupe-validate-suite-runs.feature are asserted by a `// Scenario:`-tagged acceptance test; the two script-behavior scenarios are additionally proven behaviorally at unit tier via a temp Go module.

## Acceptance Wiring

```toml
[validate]
commands = [
  "go test ./... -coverprofile=coverage.out",
  "COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh",
  "./scripts/check-fmt.sh"
]
```
The profiled `go test ./...` run includes tests/acceptance (under ./...), satisfying the acceptance-execution gate in the single suite run.

## Execution Evidence (one full profiled run, per the one-full-run guidance)

- Iteration: affected packages only — `go test ./tests/unit/... ./tests/integration/... ./cmd/centinela/... ./internal/verify/...` → 853 passed; targeted acceptance run → 13 passed (12 new + scaffold parity).
- Final: `go test ./... -coverprofile=coverage.out` → **3725 passed in 45 packages** (exit 0).
- `COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh` → **coverage gate passed: 97.1% >= 95.0%** (reused the run's own profile).

## Regression Guards

- Guard rails untouched and green: tests/unit/coverage_gate_script_test.go, coverage_hardening integration+acceptance (MIN_COVERAGE:-95.0 literal), enforce_coverage_in_validate_test.go, internal/verify/claim_tests_test.go (PriorTestRun suppression), cmd/centinela/verify_gate_test.go (nil = real re-run).
- New guards: profile-pairing invariant (integration), no-bare-suite-step negative (integration + acceptance), verifyDepsFor must never mention PriorTestRun (acceptance), post-run VerificationFresh ordering (acceptance).

## Deferred Findings

- none (the pre-agreed `cross-process-suite-result-reuse` deferral was recorded at the plan step).

## Handoff

- Next role: validation-specialist
- Edge-case report: .workflow/dedupe-validate-suite-runs-edge-cases.md (produced this step).
