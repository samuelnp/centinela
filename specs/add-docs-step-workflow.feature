Feature: Enforce docs as final workflow step
  As a maintainer
  I want every feature to pass through a docs step after validate
  So human-facing documentation stays current automatically

  Scenario: Existing workflow advances to docs after validate
    Given a standard feature workflow is active
    When validate is completed successfully
    Then the next step should be docs

  Scenario: Bootstrap workflow also includes docs
    Given a bootstrap feature workflow is active
    Then its step order should include docs as final step

  Scenario: Docs step requires documentation-specialist evidence
    Given strict orchestration mode is enabled
    When docs step completion is attempted without required evidence
    Then completion should fail with actionable missing evidence guidance
