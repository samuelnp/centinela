Feature: Edge-case subagent requirement in tests phase
  As a maintainer
  I want hard-path analysis mandatory during tests
  So production risks are discovered before release

  Scenario: Tests phase blocks without edge-case report
    Given a feature is in tests step
    And unit/integration/acceptance tests exist
    When I run "centinela complete <feature>"
    Then completion should fail with guidance to create ".workflow/<feature>-edge-cases.md"

  Scenario: Tests phase passes with edge-case report
    Given a feature is in tests step
    And required tests exist
    And edge-case report exists at ".workflow/<feature>-edge-cases.md"
    When I run "centinela complete <feature>"
    Then tests step should complete successfully

  Scenario: Context hook reminds about missing report
    Given an active workflow in tests step
    And edge-case report is missing
    When context hook runs
    Then output should include a reminder to run edge-case analysis and write the report
