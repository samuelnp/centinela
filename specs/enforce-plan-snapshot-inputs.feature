Feature: Enforce plan evidence project snapshot inputs
  As a maintainer
  I want plan subagents to prove they read current feature docs
  So planning decisions always include full project feature context

  Scenario: Plan evidence fails without full snapshot coverage
    Given docs/features files exist
    When big-thinker or feature-specialist evidence omits any feature brief path
    Then strict orchestration validation should fail with missing file details

  Scenario: Plan evidence passes with full snapshot coverage
    Given docs/features files exist
    When plan evidence inputs include all docs/features/*.md paths
    Then strict orchestration validation should pass
