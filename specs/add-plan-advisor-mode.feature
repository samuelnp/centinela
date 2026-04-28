Feature: Add plan advisor mode
  As a maintainer
  I want planning prompts to behave like adaptive advisors
  So the model expands the user's reasoning without repeating generic questions

  Scenario: Advisor mode activates only during the plan step
    Given an active workflow in the plan step
    When the plan-advisor hook runs
    Then it should emit advisor guidance for big-thinker and feature-specialist lenses

  Scenario: Advisor mode stays quiet outside the plan step
    Given an active workflow outside the plan step
    When the plan-advisor hook runs
    Then it should emit no advisor guidance

  Scenario: Advisor mode asks only missing questions
    Given a plan-step workflow with a partially complete feature brief and spec
    When the plan-advisor hook runs
    Then it should ask at most 4 high-value questions about missing planning coverage
    And it should avoid repeating topics already documented

  Scenario: User-facing planning asks UX and mobile-first questions only when missing
    Given a user-facing plan-step workflow without UX and mobile-first detail
    When the plan-advisor hook runs
    Then it should ask targeted UX and mobile-first planning questions
