Feature: README reflects current Centinela usage
  Scenario: README summarizes the latest usage-oriented features
    Given a developer is evaluating Centinela from the README
    When they read the latest features and getting started sections
    Then they should see current setup, roadmap, orchestration, validation, migration, and docs capabilities
    And they should understand that standard features follow plan, code, tests, validate, and docs in order

  Scenario: README links to a focused usage guide when the tutorial is long
    Given the README would become too verbose with a full example
    When a landing page MVP tutorial is added separately
    Then the README should link to the HOWTO from the getting started or workflow sections
    And the HOWTO should be discoverable by new users

  Scenario: HOWTO teaches agent collaboration for a landing page MVP
    Given a developer wants an agent to build a small landing page MVP
    When they follow the HOWTO
    Then they should see example prompts and Centinela commands for each workflow step
    And they should see when to approve step advancement
    And they should see validation and documentation commands before considering the work complete
