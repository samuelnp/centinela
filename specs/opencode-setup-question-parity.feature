Feature: OpenCode setup question parity
  As a developer initializing a Centinela project with OpenCode
  I want the setup prompt to ask the same questions Claude asks
  So project configuration has consistent detail across agents

  Scenario: Missing PROJECT.md shows the exact setup checklist
    Given a Centinela project initialized for OpenCode
    When PROJECT.md is missing
    Then the setup directive should ask for project name
    And the setup directive should ask for elevator pitch
    And the setup directive should ask for tech stack
    And the setup directive should ask for architecture archetype
    And the setup directive should ask for locales
    And the setup directive should ask for folder layout

  Scenario: Setup checklist remains before feature workflow
    Given PROJECT.md is missing
    When the setup directive is rendered
    Then it should not mention centinela start <feature>
    And it should still hand off to roadmap setup after PROJECT.md is written
