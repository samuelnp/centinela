# dedupe-validate-suite-runs — senior-engineer

### Senior-Engineer Report: dedupe-validate-suite-runs
**Date:** 2026-07-29

## Files Touched

| Path | Reason |
|------|--------|
| centinela.toml | Lever A: validate.commands = profiled suite run + COVERAGE_PROFILE reuse pair; [validate] and [verify] comment blocks rewritten for the single-run + reuse design; verify_timeout = 360 kept |
| scripts/check-coverage.sh | Lever A: reuse branch — skip internal `go test` only when COVERAGE_PROFILE is explicitly set AND the file exists; fail-safe self-run otherwise; MIN_COVERAGE/COVERAGE_VALUE lines untouched |
| cmd/centinela/complete_verify.go | Lever B: runClaimVerification gains `prior *verify.RunOutcome`; sets deps.PriorTestRun on the deps from verifyDepsFor (verifyDepsFor untouched) |
| cmd/centinela/complete_validate_gates.go | Lever B: second VerificationFresh check after executeValidation (closes the tree-change race, telemetry "gates" on failure), then passes completedValidationOutcome() to claim verification |
| cmd/centinela/complete_validation_outcome.go | Lever B: new helper; doc comment states the soundness invariant (only reachable after executeValidation returned nil in the same process, tree re-verified fresh) |
| cmd/centinela/verify_gate_test.go | Lever B: both runClaimVerification callers pass nil (real re-run semantics preserved) |
| docs/architecture/gatekeeper-prompt.md (+ scaffold mirror) | Lever C: Mandatory Execution item 2 conditional — validate run IS the suite run when validate.commands executes it in full; budget note softened; mirrors byte-identical (cmp verified) |
| docs/architecture/qa-senior-prompt.md (+ scaffold mirror) | Lever F: working-method guidance — affected-package iteration, exactly ONE full profiled run + COVERAGE_PROFILE coverage check before closing the tests step; mirrors byte-identical (cmp verified) |
| .github/workflows/validate.yml | Lever D: bare "Run test suite" step removed; `go run ./cmd/centinela validate` is the single suite execution, with a comment saying so |

## Architecture Compliance

- Boundary checks passed: cmd/ imports internal/{config,verify,workflow,treestate,telemetry} only — all existing allowed edges; no new internal package edges introduced.
- G1 file size: complete_verify.go 31, complete_validate_gates.go 38, complete_validation_outcome.go 20, verify_gate_test.go 53, check-coverage.sh 34 lines — all ≤100.
- G7 outer-layer rule: no business logic added to any outer layer; the reuse decision lives in the cmd gate orchestration (its existing home) and the verify domain seam (PriorTestRun) was already shipped.

## Type-Safety Notes

- `prior *verify.RunOutcome` is a typed pointer with nil = "re-run for real"; no stringly-typed flags or interface{} escapes.
- `completedValidationOutcome()` returns the concrete `*verify.RunOutcome`; classifyTestRun consumes it through the existing typed path.
- Shell reuse branch uses plain POSIX `[ -z ]`/`[ ! -f ]` under `set -eu` — unset-variable expansion is guarded with `${COVERAGE_PROFILE:-}`.

## Trade-Offs

- Synthesized in-process outcome vs. persisting the gate run's record on disk: in-process chosen — any on-disk record is writable by the agent it would excuse (cross-process reuse deferred as `cross-process-suite-result-reuse`, pre-recorded by planner).
- Second VerificationFresh call vs. trusting the pre-run check: re-check chosen; the ~200s suite window is a real mutation race and the check is cheap.
- `verifyDepsFor` left untouched vs. adding a prior param there: untouched — verify/verdict/MCP must never inherit reuse; only the one production caller in runValidateGates passes non-nil.

## Deviations (hook-blocked writes)

- The prewrite hook blocks ALL `tests/` tier writes during the code step
  ("Can't write \"tests\" files during \"code\" step"). Two planned
  breaking-test updates could NOT be made and are handed to qa-senior:
  1. tests/integration/coverage_validate_config_test.go — exact match
     `c == "./scripts/check-coverage.sh"` must become contains + the
     profile-pairing assertion (one command contains
     `-coverprofile=coverage.out` AND the coverage entry contains
     `COVERAGE_PROFILE=coverage.out`). Currently FAILS at runtime
     (compiles fine) — the only red test on the tree.
  2. tests/integration/ci_validate_workflow_integration_test.go — drop the
     bare `go test ./...` expectation, add the no-bare-suite-step negative
     assertion. Currently still passes only because a yml comment contains
     the substring; the intended strengthened assertions are pending.
- Colocated cmd/ test edits were allowed, so verify_gate_test.go was updated
  and every commit compiles.

## Deferred Findings

- none (the pre-agreed `cross-process-suite-result-reuse` deferral was recorded at the plan step).

## Handoff

- Next role: qa-senior
- Outstanding TODOs: apply the two blocked tests/integration updates above;
  add the new plan-mandated tests (script reuse/fail-safe/bare acceptance
  trio in tests/acceptance/, completedValidationOutcome colocated test,
  prompt-wording and toml single-suite-entry assertions); finish with
  exactly ONE full profiled run + COVERAGE_PROFILE coverage check.
