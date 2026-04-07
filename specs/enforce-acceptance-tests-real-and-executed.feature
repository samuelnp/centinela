Feature: Real and executed acceptance tests enforcement
  As a maintainer
  I want acceptance artifacts to be executable and actually run
  So workflow quality cannot be satisfied by placeholder files

  Scenario: Tests step fails for comment-only acceptance file
    Given feature is in tests step
    And tests/acceptance contains only comment text
    When I run "centinela complete <feature>"
    Then completion should fail requiring executable acceptance artifacts

  Scenario: Tests step fails for placeholder no-op acceptance file
    Given feature is in tests step
    And tests/acceptance contains placeholder no-op logic
    When I run "centinela complete <feature>"
    Then completion should fail requiring executable acceptance artifacts

  Scenario: Tests step fails when acceptance execution command is missing
    Given feature is in tests step
    And acceptance artifacts are executable
    And validate commands do not run acceptance tests
    When I run "centinela complete <feature>"
    Then completion should fail requiring acceptance execution configuration
