Feature: Coverage Hardening
  As a Centinela maintainer
  I want total Go statement coverage to be >= 97%
  So that parallel PR merges cannot silently tip main below the 95% gate

  Background:
    Given the Go test suite compiles and all existing tests pass
    And scripts/check-coverage.sh is present with an unmodified threshold of 95.0%

  Scenario: Total coverage meets the hardened target
    Given the full Go test suite including all new colocated *_test.go files
    When scripts/check-coverage.sh runs with the default threshold
    Then the reported total statement coverage is at least 97.0%
    And the script exits with status 0

  Scenario: Coverage gate still passes at the configured floor
    Given the coverage gate threshold remains configured at 95.0%
    And no changes have been made to scripts/check-coverage.sh
    When scripts/check-coverage.sh runs after the coverage-hardening tests are merged
    Then the gate passes because actual coverage (>= 97.0%) exceeds the floor (95.0%)
    And the gate threshold value in scripts/check-coverage.sh is still 95.0

  Scenario: New tests are colocated and within size limits
    Given the set of test files added by coverage-hardening
    When each file is inspected for package declaration and line count
    Then every new *_test.go file declares the same package as the file under test
    And no new *_test.go file exceeds 100 lines
    And no new source file (production or test) exceeds 130 lines

  Scenario: No production behaviour changed
    Given the full test suite runs including all new tests
    When go test ./... completes
    Then all tests pass with exit status 0
    And no existing acceptance or behavioural spec regresses
    And any testability seams added are pure extractions with no observable behaviour change

  Scenario: Hard-to-unit-test paths are explicitly deferred, not faked
    Given the functions runMcpServe, mcpConnectSelf, runVulnTool, and WriteBytesAtomic I/O error branches
    When the coverage-hardening feature is complete
    Then none of those functions are covered by hollow or assertion-free tests
    And each is recorded in the roadmap as a deferred backlog item with slug:
      | slug                                        |
      | unit-test-mcp-server-in-memory-transport    |
      | fault-inject-atomic-write-error-paths       |
      | unit-test-vuln-tool-external-seam           |
    And centinela roadmap list --status deferred shows all three slugs
