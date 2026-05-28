Feature: Claim verification — independently verify evidence against ground truth
  As a developer delegating steps to subagents
  I want centinela to independently re-derive ground truth for each evidence claim
  So a fabricated or optimistic report cannot advance the workflow past centinela complete

  # ── Happy path ─────────────────────────────────────────────────────────────

  Scenario: Honest evidence verifies green and completes unchanged
    Given a feature "honest-feature" at the tests step
    And qa-senior evidence with status "done" that asserts tests pass, coverage moved, non-stub outputs, and mapped edge cases
    And the configured validate.commands exit 0
    And all outputs files contain substantive content with real assertions
    And each edgeCases entry has a matching test name in the feature's test files
    And measured per-package coverage is within tolerance of the claimed figure
    When the user runs "centinela verify honest-feature"
    Then the verify report should show PASS for all four claim checks
    And the exit code should be 0
    When the user runs "centinela complete honest-feature"
    Then the step should advance without blocking

  # ── Negative path: tests-pass claim (HARD FAIL) ─────────────────────────────

  Scenario: Fabricated tests-pass claim is blocked
    Given a feature "fabricated-tests" at the tests step
    And qa-senior evidence asserting "tests pass" (status done)
    But the configured validate.commands exit with a non-zero code
    When the user runs "centinela verify fabricated-tests"
    Then the verify report should show FAIL for the tests-pass check
    And the output should name the failing command
    And the exit code should be non-zero
    When the user runs "centinela complete fabricated-tests"
    Then completion should be hard-blocked with the failing claim named
    And the step should NOT advance

  # ── Negative path: coverage claim (HARD FAIL) ───────────────────────────────

  Scenario: Inflated coverage claim is blocked
    Given a feature "inflated-coverage" at the tests step
    And qa-senior evidence claiming 92% coverage
    But re-derived per-package coverage (no -coverpkg) measures 78%
    And the difference exceeds the configured coverage_tolerance (default 0.1%)
    When the user runs "centinela verify inflated-coverage"
    Then the verify report should show FAIL for the coverage check
    And the output should show both the claimed figure (92%) and the measured figure (78%)
    And the exit code should be non-zero
    When the user runs "centinela complete inflated-coverage"
    Then completion should be hard-blocked with the coverage discrepancy named

  Scenario: Coverage claim within tolerance passes
    Given a feature "near-coverage" at the tests step
    And qa-senior evidence claiming 85.0% coverage
    And re-derived per-package coverage measures 84.95%
    And the difference (0.05%) is within the configured coverage_tolerance (default 0.1%)
    When the user runs "centinela verify near-coverage"
    Then the verify report should show PASS for the coverage check

  # ── Negative path: empty-stub outputs (HARD FAIL) ───────────────────────────

  Scenario: Empty-stub output file is blocked
    Given a feature "stub-outputs" at the tests step
    And qa-senior evidence listing a test output file as "tests/unit/foo_test.go"
    But "tests/unit/foo_test.go" contains only an empty "func TestFoo(t *testing.T) {}" body with no assertions
    When the user runs "centinela verify stub-outputs"
    Then the verify report should show FAIL for the outputs-stub check
    And the output should name "tests/unit/foo_test.go" as the offending file
    And the exit code should be non-zero
    When the user runs "centinela complete stub-outputs"
    Then completion should be hard-blocked with the stub file named

  Scenario: Legitimately tiny non-test file is not flagged as stub
    Given a feature "tiny-interface" at the tests step
    And an outputs file "internal/verify/runner.go" containing only a Go interface definition (under 40 lines, no test assertions required)
    When the user runs "centinela verify tiny-interface"
    Then the verify report should show PASS for the outputs-stub check for that file

  # ── Negative path: edge-case mapping (WARN, not hard fail) ─────────────────

  Scenario: Edge case with no corresponding test emits a warning but does not hard-block
    Given a feature "unmapped-edge" at the tests step
    And qa-senior evidence with edgeCases entry "timeout while suite runs"
    But no test name or assertion containing a recognisable match for "timeout" exists in the feature's test files
    When the user runs "centinela verify unmapped-edge"
    Then the verify report should show WARN for the edge-cases check
    And the output should name the unmatched edge case entry
    And the exit code should be non-zero (warning counts as non-zero for verify)
    When the user runs "centinela complete unmapped-edge"
    Then completion should NOT be hard-blocked by the edge-cases warning alone
    And the step should advance with the warning surfaced in the complete output

  Scenario: All edge cases matched by test names passes
    Given a feature "matched-edges" at the tests step
    And qa-senior evidence with edgeCases entries "missing evidence files" and "misconfigured test command"
    And the feature's test files contain test functions named "TestMissingEvidenceFiles" and "TestMisconfiguredTestCommand"
    When the user runs "centinela verify matched-edges"
    Then the verify report should show PASS for the edge-cases check

  # ── Edge-case: no evidence files (skip, not fail) ───────────────────────────

  Scenario: No evidence files for a step reports skip and does not block
    Given a feature "fresh-feature" that has been started but has no evidence JSON written yet
    When the user runs "centinela verify fresh-feature"
    Then the verify report should show SKIP for all claim checks with message "no claims to verify"
    And the exit code should be 0
    When the user runs "centinela complete fresh-feature"
    Then verification does not block completion (no claims present)

  # ── Edge-case: claim type absent in evidence (skip that check) ───────────────

  Scenario: Evidence omitting coverage field skips the coverage check
    Given a feature "no-coverage-claim" at the tests step
    And qa-senior evidence that does not include any coverage claim field
    When the user runs "centinela verify no-coverage-claim"
    Then the verify report should show SKIP for the coverage check
    And all other present claims should still be checked normally

  Scenario: Evidence omitting edgeCases field skips the edge-cases check
    Given a feature "no-edge-claim" at the tests step
    And qa-senior evidence that has an empty edgeCases list
    When the user runs "centinela verify no-edge-claim"
    Then the verify report should show SKIP for the edge-cases check

  # ── Edge-case: misconfigured test command (config error, not claim fail) ─────

  Scenario: Missing validate.commands surfaces as a configuration error not a claim failure
    Given a feature "bad-config" at the tests step
    And qa-senior evidence asserting tests pass
    But "validate.commands" is empty or absent in centinela.toml
    When the user runs "centinela verify bad-config"
    Then the verify report should show a CONFIG ERROR (not FAIL) for the tests-pass check
    And the error message should instruct the user to configure validate.commands
    And the exit code should be non-zero

  Scenario: Test command that references a missing binary surfaces as a configuration error
    Given a feature "missing-binary" at the tests step
    And qa-senior evidence asserting tests pass
    And validate.commands references a binary that is not on PATH
    When the user runs "centinela verify missing-binary"
    Then the verify report should show a CONFIG ERROR for the tests-pass check
    And the error message should name the missing binary
    And the output should be distinct from a claim-failure message

  # ── Edge-case: worktree resolution ──────────────────────────────────────────

  Scenario: Verify runs against the feature worktree not the root checkout
    Given the project has workflow.use_worktrees set to true
    And a feature "wt-feature" has a worktree at ".worktrees/wt-feature"
    And the user's current directory is inside ".worktrees/wt-feature"
    And ".worktrees/wt-feature/.workflow/wt-feature-qa-senior.json" contains the evidence
    When the user runs "centinela verify wt-feature"
    Then evidence and test commands should be resolved relative to ".worktrees/wt-feature/"
    And the root checkout's ".workflow/" should not be consulted

  Scenario: Verify from the root checkout resolves the correct worktree
    Given the project has workflow.use_worktrees set to true
    And a feature "root-run-feature" has a worktree at ".worktrees/root-run-feature"
    And the user's current directory is the repo root
    When the user runs "centinela verify root-run-feature"
    Then verify should resolve the worktree path for "root-run-feature" and operate inside it

  # ── Edge-case: timeout ───────────────────────────────────────────────────────

  Scenario: Suite that exceeds the verify timeout fails with a timeout error not a claim failure
    Given a feature "slow-suite" at the tests step
    And qa-senior evidence asserting tests pass
    And verify_timeout is set to 30s in centinela.toml
    But the test suite takes longer than 30 seconds to run
    When the user runs "centinela verify slow-suite"
    Then the verify report should show TIMEOUT for the tests-pass check
    And the output should name the command that timed out and the configured timeout value
    And the exit code should be non-zero
    When the user runs "centinela complete slow-suite"
    Then completion should be hard-blocked because the tests-pass claim could not be confirmed

  # ── CLI surface ───────────────────────────────────────────────────────────────

  Scenario: Verify report displays per-claim PASS FAIL SKIP WARN lines
    Given a feature with mixed claim results
    When the user runs "centinela verify <feature>"
    Then the output should contain one line per claim check prefixed with PASS, FAIL, SKIP, WARN, or TIMEOUT
    And a summary line should appear at the end (e.g. "2 passed, 1 failed, 1 skipped")

  Scenario: Complete gate surfaces verify failures inline with the structural gate
    Given "centinela complete <feature>" is run and verify finds a failing claim
    Then the complete output should include the full verify report before the gate error line
    And the error message should distinguish a claim-failure from a structural-evidence failure
