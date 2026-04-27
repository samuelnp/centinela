Feature: Refine UX specialist evidence
  As a maintainer
  I want UX specialist evidence to be mobile-first and cover required review concerns
  So user-facing interfaces are consistently reviewed for real usability and polish

  Scenario: UX evidence fails without mobile-first confirmation
    Given a user-facing feature in the code step
    When ux-ui-specialist evidence omits mobileFirst or sets it to false
    Then strict orchestration validation should fail with mobile-first details

  Scenario: UX evidence fails without required UX edge-case tags
    Given a user-facing feature in the code step
    When ux-ui-specialist evidence omits required UX review tags
    Then strict orchestration validation should fail with missing UX tag details

  Scenario: UX evidence passes with mobile-first and required tags
    Given a user-facing feature in the code step
    When ux-ui-specialist evidence sets mobileFirst to true, lists required tags, and points to real UI files
    Then strict orchestration validation should pass
