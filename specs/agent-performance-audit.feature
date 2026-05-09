Feature: Agent performance audit
  As a Centinela maintainer
  I want an audit of configured OpenCode agents and workflow coverage
  So I can reduce token waste while keeping specialist quality

  Scenario: Audit covers every configured OpenCode specialist
    Given Centinela configures native OpenCode subagents
    When the audit is written
    Then it should mention big-thinker
    And it should mention feature-specialist
    And it should mention senior-engineer
    And it should mention qa-senior
    And it should mention documentation-specialist
    And it should mention ux-ui-specialist

  Scenario: Audit covers all workflow steps
    Given the Centinela workflow has plan, code, tests, validate, and docs steps
    When the audit is written
    Then it should identify coverage for each step
    And it should call out validate-step native agent coverage if missing

  Scenario: Missing validate agent is generated
    Given Centinela configures native OpenCode subagents
    When Centinela injects opencode.json
    Then opencode.json should include validation-specialist as a subagent
    And the validate step should require validation-specialist evidence

  Scenario: Audit recommends performance improvements
    Given generated agent prompts consume context
    When the audit is written
    Then it should recommend prompt trimming opportunities
    And it should separate safe reductions from behavior-changing changes
