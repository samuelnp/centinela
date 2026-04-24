Feature: Explicit roadmap artifact guidance
  Scenario: Start command reports missing roadmap json explicitly
    Given a greenfield project with PROJECT.md and ROADMAP.md
    When .workflow/roadmap.json is missing and the user runs centinela start
    Then the error should mention .workflow/roadmap.json
    And the error should tell the user how to validate roadmap artifacts

  Scenario: Setup hook distinguishes missing roadmap json from missing roadmap markdown
    Given PROJECT.md exists and ROADMAP.md exists
    When .workflow/roadmap.json is missing
    Then the setup hook should instruct the agent to write .workflow/roadmap.json
    And it should show the exact required JSON format

  Scenario: Artifact template docs cover setup and per-feature files
    Given a user needs to recover from missing Centinela artifacts
    When they read the scaffolded artifact template documentation
    Then they should see setup artifact templates
    And they should see per-feature workflow artifact templates
