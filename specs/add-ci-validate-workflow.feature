Feature: CI validate workflow
  As a maintainer
  I want validation gates to run in CI
  So regressions are blocked before merge

  Scenario: CI workflow runs centinela validate
    Given repository CI config
    When push or pull_request events occur
    Then the workflow should execute go tests and centinela validate

  Scenario: Coverage gate is enforced in CI
    Given centinela.toml includes the coverage script command
    When CI runs centinela validate
    Then coverage threshold failures should fail the job
