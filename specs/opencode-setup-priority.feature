Feature: OpenCode setup priority
  As an OpenCode user
  I want Centinela setup guidance to override casual greetings
  So project bootstrap starts automatically when required files are missing

  Scenario: Generated OpenCode instructions require setup before conversation
    Given an OpenCode project initialized by centinela
    When PROJECT.md is missing
    Then generated OpenCode instructions should tell the agent to start setup immediately
    And they should not require the user to explicitly ask for PROJECT.md configuration

  Scenario: Plugin injects setup guidance before other prompt context
    Given an OpenCode project initialized by centinela
    When the plugin handles prompt lifecycle events
    Then setup guidance should be appended before autostart, orchestration, and workflow context

  Scenario: Setup docs describe OpenCode bootstrap explicitly
    Given a new Centinela project
    When the user opens an OpenCode session first
    Then the docs should say setup prompts apply to the coding agent, not only Claude
