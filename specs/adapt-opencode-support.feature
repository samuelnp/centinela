Feature: OpenCode support in Centinela
  As a team using OpenCode
  I want Centinela to enforce workflow rules in OpenCode sessions
  So that I keep the same guardrails as Claude users

  Scenario: Init creates OpenCode artifacts
    Given a project initialized with Centinela
    When I run "centinela init --agent opencode"
    Then the project should contain "opencode.json"
    And the project should contain ".opencode/plugins/centinela.js"

  Scenario: Init keeps Claude compatibility by default
    Given a project initialized with Centinela
    When I run "centinela init"
    Then Claude hook settings should be configured
    And OpenCode artifacts should be configured

  Scenario: OpenCode blocks out-of-step writes
    Given feature "example" is in "plan" step
    When OpenCode tries to write a code file under "internal/"
    Then Centinela should block the write
    And show an explanation with current feature and step

  Scenario: OpenCode allows roadmap files in any step
    Given feature "example" is in "code" step
    When OpenCode writes "docs/features/example.md"
    Then Centinela should allow the write

  Scenario: Existing OpenCode config is preserved
    Given a project with an existing "opencode.json"
    When I run "centinela init --agent opencode"
    Then existing unrelated keys should remain unchanged
    And required Centinela entries should be present
