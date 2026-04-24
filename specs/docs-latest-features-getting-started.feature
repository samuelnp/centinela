Feature: Latest features documentation and workflow onboarding
  Scenario: README highlights current user-facing capabilities
    Given a user is evaluating what Centinela supports today
    When they read the README
    Then they should see the latest features grouped by workflow theme
    And they should see current commands for roadmap, migration, and docs flows
    And they should not see stale four-step workflow references

  Scenario: Getting started teaches the enforced workflow
    Given a user wants to learn the Centinela process
    When they follow the getting started documentation
    Then they should see how to bootstrap a project
    And they should see how to start and advance a feature
    And they should see how validation and documentation generation fit into the workflow

  Scenario: Generated HTML presents the same current story
    Given a user opens the generated project documentation HTML
    When they review the presentation sections
    Then they should see the latest features summary
    And they should see a getting started workflow section
    And the page should remain generated from Centinela artifacts
