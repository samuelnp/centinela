Feature: Make orchestration evidence contract explicit in agent prompts
  As a maintainer or agent operator
  I want every agent prompt to spell out the JSON evidence schema and rules
  So agents produce passing evidence JSON the first time

  Scenario: A canonical evidence contract document exists
    Given the docs/architecture directory
    Then a file named evidence-contract.md should exist
    And it should describe the full JSON schema fields
    And it should list per-role rules for big-thinker, feature-specialist, senior-engineer, qa-senior, ux-ui-specialist, validation-specialist, and documentation-specialist

  Scenario: Plan-step prompts require the feature-doc snapshot
    Given the big-thinker and feature-specialist prompts
    Then each prompt should state that inputs must include every docs/features/*.md
    And each prompt should embed a role-specific JSON skeleton with realistic placeholders

  Scenario: Senior-engineer prompt rejects evidence-only outputs
    Given the senior-engineer prompt
    Then it should state that outputs must include at least one real implementation file outside .workflow/, tests/, docs/, and specs/

  Scenario: QA-senior prompt requires test file and edge-cases evidence
    Given the qa-senior prompt
    Then it should state that outputs must include at least one tests/ file
    And it should state that outputs must include .workflow/<feature>-edge-cases.md

  Scenario: UX-UI specialist prompt enforces mobileFirst and the eight UX tags
    Given the ux-ui-specialist prompt
    Then it should require mobileFirst set to true
    And it should list all eight required UX edge-case tags

  Scenario: Scaffold mirrors stay in sync with the docs prompts
    Given the internal/scaffold/assets/docs/architecture directory
    Then every updated prompt should have an identical companion under that path
