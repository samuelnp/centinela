Feature: Workflow archetypes
  Named lightweight tracks (hotfix, refactor, spike) beside the canonical
  five-step workflow, so diagnosis, restructuring, and exploration aren't
  forced through a feature-shaped plan-to-docs pipeline. Archetypes reuse
  the canonical step names, so verification stays attached to the step.

  Scenario: The hotfix archetype resolves to a code-tests-validate order
    Given a feature started with the hotfix archetype
    When its step order is resolved
    Then the order is code, tests, validate
    And the plan and docs steps are absent

  Scenario: The refactor archetype resolves to a plan-code-tests-validate order
    Given a feature started with the refactor archetype
    When its step order is resolved
    Then the order is plan, code, tests, validate
    And the docs step is absent

  Scenario: The spike archetype resolves to a plan-code order with no validate step
    Given a feature started with the spike archetype
    When its step order is resolved
    Then the order is plan, code
    And the order contains no validate step

  Scenario: The default archetype is the canonical five-step order
    Given a feature started with no archetype specified
    When its step order is resolved
    Then the archetype is canonical
    And the order is plan, code, tests, validate, docs

  Scenario: A ship-gated archetype runs gates and claim verification
    Given a feature whose resolved order contains the validate step
    When that validate step is reached and completed
    Then the ship gate fires, running gates and claim verification before advancing

  Scenario: A spike never reaches the ship gate
    Given a spike feature whose order omits the validate step
    When the spike is worked through to its final step
    Then the validate gate is never triggered for it

  Scenario: An explicit archetype flag overrides the roadmap archetype
    Given a roadmap that assigns a feature the refactor archetype
    And the feature is started with an explicit hotfix archetype flag
    When the archetype is resolved
    Then the resolved archetype is hotfix

  Scenario: The active archetype is pinned in the workflow state
    Given a feature started with the hotfix archetype
    When the workflow state is reloaded
    Then the persisted archetype is hotfix

  Scenario: Archetype and enforcement profile are independent
    Given a feature started with the spike archetype and the strict profile
    When the workflow is created
    Then the step order is the spike order
    And the enforcement profile is strict

  Scenario: An unknown archetype value is rejected
    Given a feature start request with an unsupported archetype name
    When the archetype is validated
    Then validation fails with an error naming the archetype

  Scenario: The status output shows the active archetype
    Given a feature started with the spike archetype
    When the workflow status is rendered
    Then the output names the spike archetype
