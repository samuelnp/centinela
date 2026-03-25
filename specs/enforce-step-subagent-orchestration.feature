Feature: Strict subagent orchestration by workflow step
  As a maintainer
  I want Centinela to enforce required subagent delegation artifacts
  So each step has auditable specialist outputs

  Scenario: Plan step requires two specialist evidence pairs
    Given a strict-enabled workflow in plan step
    When completion is attempted without all plan role evidence
    Then completion should fail with missing artifact details

  Scenario: Code step requires senior engineer evidence pair
    Given a strict-enabled workflow in code step
    When completion is attempted with invalid or missing code evidence
    Then completion should fail with validation details

  Scenario: Tests step requires QA evidence pair
    Given a strict-enabled workflow in tests step
    When completion is attempted without QA role evidence
    Then completion should fail

  Scenario: Legacy workflow is not retroactively blocked
    Given an existing workflow without orchestration strict metadata
    When completion is attempted
    Then orchestration evidence should not be required
