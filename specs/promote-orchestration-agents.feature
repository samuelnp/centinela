Feature: Promote orchestration roles to standalone prompt files
  As a Centinela maintainer
  I want every orchestration role to have a Markdown prompt file
  So that plan, code, tests, and validate specialists have structured
  guidance equivalent to the existing report agents.

  Scenario: Each orchestration role has a prompt file under docs/architecture/
    Given the orchestration roles defined in internal/orchestration/policy.go
    When the promote-orchestration-agents feature is implemented
    Then docs/architecture/big-thinker-prompt.md should exist
    And docs/architecture/feature-specialist-prompt.md should exist
    And docs/architecture/senior-engineer-prompt.md should exist
    And docs/architecture/qa-senior-prompt.md should exist
    And docs/architecture/ux-ui-specialist-prompt.md should exist
    And docs/architecture/validation-specialist-prompt.md should exist

  Scenario: Each new prompt declares Purpose, Prompt Template, and Required Artifact
    Given a new orchestration prompt file
    When the file is read
    Then it should contain a "## Purpose" heading
    And it should contain a "## Prompt Template" heading
    And it should contain a "## Required Artifact" heading

  Scenario: Each new prompt is mirrored in the scaffold tree
    Given a new orchestration prompt file under docs/architecture/
    When the scaffold mirror is checked
    Then internal/scaffold/assets/docs/architecture/<same-filename> should exist
    And the two files should be byte-identical

  Scenario: Per-file length budget is respected
    Given a new orchestration prompt file
    When the file is measured
    Then it should be at most 70 lines

  Scenario: Runtime configuration is unchanged
    Given the promote-orchestration-agents feature is implemented
    When the runtime configuration files are inspected
    Then internal/setup/opencode_agent_config.go should be unchanged
    And internal/orchestration/policy.go should be unchanged
    And cmd/centinela/hook_orchestration.go should be unchanged
