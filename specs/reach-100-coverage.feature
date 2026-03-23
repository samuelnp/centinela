Feature: Full statement coverage
  As a maintainer
  I want all executable statements covered by tests
  So regressions are caught and behavior is explicit

  Scenario: Remaining uncovered branches are identified
    Given a full-package coverage profile
    When uncovered functions are listed
    Then targeted tests are added for each remaining branch

  Scenario: Hard-to-test exit paths are exercised safely
    Given command paths that call os.Exit or terminal runtimes
    When tests run via subprocess or seam helpers
    Then execution branches are covered without changing behavior

  Scenario: Coverage target is achieved
    Given all tests pass
    When full-package coverage is measured
    Then total statement coverage equals 100.0%
