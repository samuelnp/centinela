Feature: OpenCode force setup flow
  As an OpenCode user in a new Centinela project
  I want missing project setup to block feature prompts
  So the agent configures PROJECT.md and the roadmap before feature work

  Scenario: Missing PROJECT.md blocks feature workflow suggestions
    Given an OpenCode project initialized by Centinela
    When PROJECT.md is missing
    And the user sends a greeting
    Then OpenCode instructions should require project setup immediately
    And OpenCode instructions should forbid suggesting centinela start <feature>
    And OpenCode instructions should forbid asking what feature to work on

  Scenario: Missing roadmap blocks feature prompts after PROJECT.md
    Given PROJECT.md exists
    And the roadmap is missing
    When the user sends a greeting
    Then OpenCode instructions should require roadmap bootstrap before feature work
