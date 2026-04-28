Feature: OpenCode greeting workflow guidance
  As an OpenCode user starting a Centinela project
  I want setup and workflow requirements shown before casual greeting replies
  So OpenCode behaves like Claude Code at project startup

  Scenario: Generated OpenCode instructions prioritize Centinela over greetings
    Given a project initialized with OpenCode support
    When the first user prompt is a greeting
    And Centinela setup or workflow guidance is required
    Then the generated OpenCode instructions require the agent to explain the Centinela requirement first
    And the agent should not answer only with casual conversation

  Scenario: OpenCode plugin still injects runtime setup directives first
    Given a project initialized with OpenCode support
    When OpenCode appends prompt context
    Then setup and migration directives should appear before autostart and workflow context
