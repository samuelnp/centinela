# Edge Cases: dedupe-validate-suite-runs

### Edge-Case Report: dedupe-validate-suite-runs
**Date:** 2026-07-29

#### Risk Matrix
- **Case:** Stale/doctored coverage.out masks a failing suite
- **Impact:** High
- **Likelihood:** Low
- **Why:** Reuse only reads a profile after command 1 of the SAME validate run rewrote it; a command-1 failure fails validate regardless. Proven impossible-by-ordering: `tests/unit/coverage_reuse_branch_test.go` shows reuse requires an explicitly set `COVERAGE_PROFILE` + existing file, and `tests/acceptance/dedupe_validate_suite_runs_script_test.go` pins the bare-invocation `-z` arm (stale on-disk file never activates reuse).

- **Case:** Missing named profile silently skips coverage (fail-open)
- **Impact:** High
- **Likelihood:** Low
- **Why:** `TestCoverageScript_MissingProfileFallsBackToRunningSuite` proves `COVERAGE_PROFILE=missing.out` triggers a real suite run that fails on the module's failing test — fail-safe, never fail-open.

- **Case:** COVERAGE_VALUE precedence over profile/suite changes
- **Impact:** Medium
- **Likelihood:** Low
- **Why:** `TestCoverageValueFastPathBypassesSuite` runs the script in an empty dir (no go.mod, no profile) — exit 0 proves the fast path bypasses both, unchanged.

- **Case:** Tree mutates during the ~200s gate suite run; reused outcome vouches for the wrong tree
- **Impact:** High
- **Likelihood:** Medium
- **Why:** `TestTreeChangeDuringSuiteRunBlocksCompletion` pins the post-run `workflow.VerificationFresh` re-check strictly between `executeValidation()` and `completedValidationOutcome()`.

- **Case:** verify/verdict/MCP surfaces inherit the reuse and stop re-running
- **Impact:** Medium
- **Likelihood:** Low
- **Why:** `TestStandaloneVerifySurfacesStillRerunSuite` asserts `verifyDepsFor` never mentions `PriorTestRun`; only the complete gate passes a non-nil prior.

- **Case:** Prompt-parity drift between docs/architecture and scaffold mirror
- **Impact:** Medium
- **Likelihood:** Medium
- **Why:** Byte-compare in `dedupe_validate_suite_runs_prompts_test.go` for both prompts, plus the pre-existing `TestScaffoldArchitectureMirrorParity` (neither prompt is allowlisted).

#### Missing or Weak Scenarios
- None outstanding — all 12 spec scenarios carry `// Scenario:` tagged acceptance assertions; reuse/fallback are additionally proven behaviorally at unit tier.

#### Proposed/Added Tests
- Unit: tests/unit/coverage_reuse_branch_test.go (+ coverage_reuse_helper_test.go) — reuse skips the suite (failing test proves skip); missing profile fail-safe re-run.
- Integration: tests/integration/coverage_validate_config_test.go (profile-pairing invariant); tests/integration/ci_validate_workflow_integration_test.go (no bare suite step).
- Acceptance: tests/acceptance/dedupe_validate_suite_runs_{config,prompts,script,complete}_test.go — 12/12 scenarios; colocated cmd/centinela/complete_validation_outcome_test.go for the synthesized PASS record.

#### Residual Risks
- Cross-process reuse of the gatekeeper's run remains deferred (`cross-process-suite-result-reuse`, recorded at plan step) — any on-disk record is writable by the agent it would excuse.
- Behavioral CI assertion is static YAML parsing; an exotic workflow edit that runs the suite under a differently named step would need the gatekeeper's human/agent review. Mitigation: negative assertions on both `name: Run test suite` and `run: go test ./...`.

#### Deferred Findings
- none (the pre-agreed `cross-process-suite-result-reuse` deferral was recorded at the plan step).
