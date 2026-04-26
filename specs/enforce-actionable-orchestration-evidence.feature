Feature: Enforce actionable orchestration evidence
  As a maintainer
  I want strict orchestration evidence to point to applied project outputs
  So specialist discoveries cannot pass as insight-only paperwork

  Scenario: Plan evidence fails with summary-only outputs
    Given a strict-enabled workflow in plan step
    When big-thinker or feature-specialist evidence outputs are free-text summaries instead of file paths
    Then completion should fail with actionable output validation details

  Scenario: Code evidence fails without real implementation outputs
    Given a strict-enabled workflow in code step
    When senior-engineer evidence outputs only orchestration files or summaries
    Then completion should fail with implementation output validation details

  Scenario: Tests evidence fails without real test outputs
    Given a strict-enabled workflow in tests step
    When qa-senior evidence omits concrete test files or the edge-case report
    Then completion should fail with test output validation details

  Scenario: Actionable evidence passes with real output files
    Given a strict-enabled workflow with required role evidence
    When each evidence output points to valid step artifacts on disk
    Then strict orchestration validation should pass
