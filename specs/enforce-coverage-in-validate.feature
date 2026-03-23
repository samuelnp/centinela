Feature: Coverage threshold in validate
  As a maintainer
  I want coverage threshold checks in the validate pipeline
  So test coverage regressions are blocked automatically

  Scenario: Validate includes coverage command
    Given project validation commands are configured
    When centinela validate runs
    Then it should execute the coverage gate command

  Scenario: Coverage below threshold fails validation
    Given a configured minimum threshold
    When measured total coverage is below that threshold
    Then the coverage gate exits with failure

  Scenario: Coverage above threshold passes validation
    Given a configured minimum threshold
    When measured total coverage is at or above that threshold
    Then the coverage gate exits successfully
