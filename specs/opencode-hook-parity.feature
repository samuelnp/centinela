Feature: OpenCode hook parity
  As an OpenCode user
  I want Centinela to provide the same hook guidance as Claude integration
  So that workflow guardrails are consistent across agents

  Scenario: Plugin invokes prewrite and postwrite around file edits
    Given an OpenCode project initialized by centinela
    When the plugin handles a write, edit, or patch tool call
    Then it should call "centinela hook prewrite" before execution
    And it should call "centinela hook postwrite" after execution

  Scenario: Plugin invokes setup and context on prompt submit
    Given an OpenCode project initialized by centinela
    When the plugin handles prompt lifecycle events
    Then it should call "centinela hook setup"
    And it should call "centinela hook context"

  Scenario: Hook output is appended only when non-empty
    Given hook command output with blank lines
    When the plugin receives the output
    Then blank output should be ignored
    And non-empty output should be appended to prompt context
