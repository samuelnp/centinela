Feature: Add UX/UI specialist orchestration
  As a maintainer
  I want user-facing features to require a UX/UI specialist review
  So final-user interfaces get actionable visual and interaction quality checks

  Scenario: Internal feature keeps current code-step roles
    Given a strict-enabled workflow for a non-user-facing feature
    When orchestration roles are resolved for the code step
    Then only senior-engineer evidence should be required

  Scenario: User-facing feature requires UX specialist in code step
    Given a strict-enabled workflow for a user-facing feature
    When orchestration roles are resolved for the code step
    Then senior-engineer and ux-ui-specialist evidence should be required

  Scenario: UX evidence fails without real UI outputs
    Given a user-facing feature in the code step
    When ux-ui-specialist evidence lists summaries or non-UI files
    Then strict orchestration validation should fail with UX output details

  Scenario: UX evidence passes with real UI outputs
    Given a user-facing feature in the code step
    When ux-ui-specialist evidence points to real UI files and edge cases
    Then strict orchestration validation should pass
