Feature: Plan validation checks by filename

  Scenario: Plan file exists with correct name but minimal content
    Given docs/plans/project-bootstrap.md exists
    And the file does not contain the string "project-bootstrap"
    When the plan step is validated for feature "project-bootstrap"
    Then validation passes

  Scenario: Plan file is missing entirely
    Given docs/plans/missing-feature.md does not exist
    When the plan step is validated for feature "missing-feature"
    Then validation fails with a clear error
