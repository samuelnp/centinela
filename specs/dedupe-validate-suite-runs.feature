Feature: dedupe-validate-suite-runs
  Remove redundant full-suite executions from the validate cycle while
  preserving every trust invariant: the gatekeeper still runs the gate
  itself, complete still runs the suite machine-side once, and coverage is
  always measured from a profile written by that same run.

  Scenario: validate runs the suite exactly once
    Given centinela.toml validate.commands pairs "go test ./... -coverprofile=coverage.out" with "COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh"
    And both entries name the same profile file
    When the validate.commands list is inspected
    Then exactly one entry executes the full test suite
    And the coverage entry reuses that entry's profile instead of re-running it

  Scenario: check-coverage.sh reuses an existing profile when COVERAGE_PROFILE is set
    Given a temp Go module whose recorded coverage profile was written by a passing run
    And a failing test has since been added to the module
    When check-coverage.sh runs with COVERAGE_PROFILE naming the existing profile and a threshold the profile meets
    Then the script exits 0 without re-running the suite
    And a re-run would have failed, proving the internal go test was skipped

  Scenario: check-coverage.sh falls back to running the suite when the profile is missing
    Given a temp Go module with passing tests
    And COVERAGE_PROFILE names a file that does not exist
    When check-coverage.sh runs with a threshold the module meets
    Then the script runs the suite itself and writes the named profile
    And the script exits 0

  Scenario: bare check-coverage.sh invocation stays self-contained
    Given COVERAGE_PROFILE is not set in the environment
    And a stale coverage.out file exists on disk
    When check-coverage.sh runs
    Then the script runs the full suite itself before measuring coverage
    And the default MIN_COVERAGE floor of 95.0 is applied

  Scenario: COVERAGE_VALUE fast path bypasses the suite unchanged
    Given COVERAGE_VALUE is set to a value at or above the threshold
    When check-coverage.sh runs
    Then it exits 0 without running the suite or reading a profile

  Scenario: complete's claim verification reuses the gate's own run
    Given a validate-step complete whose executeValidation just exited 0
    And the working tree is still fresh after the run
    When claim verification executes the tests-pass check
    Then it consumes the synthesized outcome of the gate's own run with exit code 0
    And the suite is not executed a second time in the same complete

  Scenario: a failing validate command blocks completion before any reuse
    Given a validate-step complete whose validate.commands include a failing command
    When centinela complete runs the validate gates
    Then executeValidation fails and completion is blocked
    And no prior test outcome is synthesized or reused

  Scenario: a tree change during the gate's suite run blocks completion
    Given a validate-step complete whose suite run succeeds
    And the working tree was mutated while the suite ran
    When the post-run freshness re-check executes
    Then VerificationFresh fails and completion is blocked before claim verification

  Scenario: standalone verify surfaces still re-run the suite for real
    Given the centinela verify, verdict, and MCP surfaces build deps via verifyDepsFor
    When any of them verifies the tests-pass claim
    Then no PriorTestRun is set and every validate command is re-executed

  Scenario: gatekeeper prompt copies stay identical and carry the conditional suite mandate
    Given docs/architecture/gatekeeper-prompt.md and its scaffold mirror
    When both files are compared byte for byte
    Then they are identical
    And the Mandatory Execution section makes the bare suite run conditional on validate.commands not already running the full suite

  Scenario: qa-senior prompt copies carry the one-full-run guidance
    Given docs/architecture/qa-senior-prompt.md and its scaffold mirror
    When both files are compared byte for byte
    Then they are identical
    And the prompt directs one full profiled run plus coverage check before closing the tests step

  Scenario: CI runs the suite once via centinela validate
    Given .github/workflows/validate.yml
    When the validate job's steps are inspected
    Then no bare "Run test suite" step exists
    And "go run ./cmd/centinela validate" remains the single suite execution
