Feature: Documentation consistency
  As a Centinela user
  I want docs and scaffold docs to use current commands and wording
  So onboarding instructions are accurate

  Scenario: Legacy workflow script references are removed
    Given documentation files in root and scaffold assets
    When command references are reviewed
    Then they should use "centinela" CLI commands instead of script paths

  Scenario: New project setup instructions use current flow
    Given the new project guide
    When a user follows setup steps
    Then command examples should match current Centinela CLI behavior

  Scenario: Agent support wording is accurate
    Given docs that describe integration behavior
    When they mention supported agents
    Then wording should reflect Claude and OpenCode support where applicable
