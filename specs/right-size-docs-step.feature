Feature: Right-size the docs step
  Make the docs step surface-aware so internal features emit only a one-line
  changelog instead of the full knowledge-base guide, portal regeneration, and
  documentation-specialist ceremony, while user-facing features keep the full
  docs flow. Mirrors how the code step requires the ux-ui-specialist only for
  user-facing features.

  Scenario: A user-facing feature still requires the documentation-specialist role
    Given a feature whose brief declares the surface as user-facing
    When the required roles for the docs step are resolved
    Then the documentation-specialist role is required

  Scenario: An internal feature does not require the documentation-specialist role
    Given a feature whose brief does not declare a user-facing surface
    When the required roles for the docs step are resolved
    Then the documentation-specialist role is not required

  Scenario: A user-facing docs step still requires the knowledge-base guide
    Given a user-facing feature on the docs step
    And the knowledge-base markdown and page and the portal exist
    When the docs artifacts are validated
    Then validation passes

  Scenario: A user-facing docs step fails without the knowledge-base guide
    Given a user-facing feature on the docs step with no knowledge-base markdown
    When the docs artifacts are validated
    Then validation fails naming the missing knowledge-base guide

  Scenario: An internal docs step passes with only a one-line changelog
    Given an internal feature on the docs step with a one-line changelog entry and no knowledge-base guide
    When the docs artifacts are validated
    Then validation passes

  Scenario: An internal docs step fails without a changelog entry
    Given an internal feature on the docs step with no changelog entry
    When the docs artifacts are validated
    Then validation fails naming the missing changelog entry

  Scenario: An internal docs step fails when the changelog entry is blank
    Given an internal feature on the docs step whose changelog entry is empty
    When the docs artifacts are validated
    Then validation fails naming the changelog entry as empty

  Scenario: A clean merge regenerates the documentation portal
    Given a feature is merged successfully
    When the merge completes
    Then the documentation portal is regenerated

  Scenario: A portal regeneration failure does not fail a clean merge
    Given a feature is merged successfully
    And the portal regeneration fails
    When the merge completes
    Then the merge still succeeds and a notice is reported

  Scenario: The default surface is internal when none is declared
    Given a feature whose brief declares no surface
    When the surface is resolved
    Then the feature is treated as internal
