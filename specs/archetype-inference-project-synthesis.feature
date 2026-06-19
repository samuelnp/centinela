Feature: archetype inference and PROJECT.md synthesis — deterministic, draft-first
  As a team adopting Centinela on a brownfield codebase that already ran centinela analyze
  I want centinela synthesize to infer the best-fit archetype from the Inventory and draft a PROJECT.md
  So that onboarding produces a trustworthy starting PROJECT.md with no LLM call and never clobbers an existing one

  Scenario: A Rails-shaped inventory infers rails-native and drafts PROJECT.md
    Given an inventory whose primary language is Ruby with a Gemfile depending on rails
    And packages include app/models, app/controllers, and app/views
    When the operator runs centinela synthesize
    Then the exit code is zero
    And a PROJECT.md file is created
    And it declares the archetype rails-native
    And its layer mapping maps Models, Controllers, and Views to concrete paths

  Scenario: A Go n-tier inventory infers n-tier with handler service repository mapping
    Given an inventory whose primary language is Go with a go.mod module path
    And packages include handler, service, and repository
    When the operator runs centinela synthesize
    Then the exit code is zero
    And the drafted PROJECT.md declares the archetype n-tier

  Scenario: A game inventory with systems and components infers ecs
    Given an inventory whose packages include systems, components, and entities
    When the operator runs centinela synthesize
    Then the exit code is zero
    And the drafted PROJECT.md declares the archetype ecs

  Scenario: An ambiguous inventory is flagged low-confidence with a rationale
    Given an inventory whose signals score two archetypes within the tie margin
    When the operator runs centinela synthesize
    Then the exit code is zero
    And the summary reports confidence low
    And the summary explains why the inference is ambiguous
    And the drafted PROJECT.md is marked as a draft to confirm or correct

  Scenario: Running synthesize without an inventory fails with guidance
    Given the project directory has no analysis inventory
    When the operator runs centinela synthesize
    Then the exit code is non-zero
    And the error message tells the operator to run centinela analyze first
    And no PROJECT.md or PROJECT.draft.md is written

  Scenario: An existing PROJECT.md is preserved and a draft is written instead
    Given the project directory already contains a PROJECT.md
    When the operator runs centinela synthesize
    Then the exit code is zero
    And the original PROJECT.md is left byte-for-byte unchanged
    And a PROJECT.draft.md file is created with the synthesized draft

  Scenario: Re-running synthesize on the same inventory is byte-identical
    Given a fixed inventory
    When the operator runs centinela synthesize twice writing to a fresh target
    Then both drafted files are byte-identical
