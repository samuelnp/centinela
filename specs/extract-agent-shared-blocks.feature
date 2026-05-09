Feature: Extract shared blocks from agent prompts
  As a Centinela maintainer paying LLM API costs
  I want repeated invocation boilerplate and stack-example matrices
  pulled out of individual prompt files into shared references
  So that each agent invocation uses fewer context tokens.

  Scenario: Shared invocation reference exists
    Given the extract-agent-shared-blocks feature is implemented
    When the documentation tree is inspected
    Then docs/architecture/agent-invocation.md should exist
    And it should describe the Agent-tool invocation pattern
    And it should describe the .workflow/<feature>-<role>.{md,json} artifact convention

  Scenario: Each affected prompt references the shared invocation
    Given a prompt file under docs/architecture/ that previously
    contained the "How to Invoke" boilerplate
    When the file is read
    Then it should reference agent-invocation.md

  Scenario: Gatekeeper duplicate decision table is removed
    Given the gatekeeper prompt
    When the file is read
    Then the dedicated "Decision Rules" table should be absent
    And the SAFE/WARNING/BLOCKING decisions should remain in the
    Output Format Recommendation block

  Scenario: Stack-specific examples live in a shared reference
    Given the extract-agent-shared-blocks feature is implemented
    When the documentation tree is inspected
    Then docs/architecture/stack-checks-reference.md should exist
    And production-readiness-prompt.md.template should reference it
    And production-readiness-prompt.md.template should not contain
    the four-language inline example matrix

  Scenario: Scaffold mirror parity is preserved
    Given any documentation file under docs/architecture/ that was
    added or edited by this feature
    When the scaffold mirror is checked
    Then internal/scaffold/assets/docs/architecture/<same-filename>
    should exist and be byte-identical
