Feature: Deterministic artifact scaffolds
  As a Centinela workflow agent (especially a limited-capability driver model)
  I want evidence init and artifact new to stamp every mechanically-derivable value
  So I spend my tokens on substance, never re-deriving shape the framework already knows

  Background:
    Given an active workflow "demo" on step "plan"
    And the repo has feature briefs under "docs/features/" including "docs/features/demo.md"

  # --- Slice 1: inputs pre-fill (the value blocker) ---

  Scenario: Init pre-fills plan-snapshot inputs for big-thinker
    When I run "centinela evidence init demo big-thinker"
    Then ".workflow/demo-big-thinker.json" inputs equal RequiredPlanInputs("demo")
    And inputs contain every "docs/features/*.md" path
    And inputs contain "docs/plans/demo.md"

  Scenario: Init pre-fill lets big-thinker pass plan-snapshot validation with zero appends
    Given I run "centinela evidence init demo big-thinker"
    And every other required scalar field is set
    When I run "centinela evidence validate demo"
    Then the big-thinker plan-snapshot rule passes without any manual "evidence append ... inputs"

  Scenario: Init pre-fills plan-snapshot inputs for feature-specialist
    When I run "centinela evidence init demo feature-specialist"
    Then ".workflow/demo-feature-specialist.json" inputs equal RequiredPlanInputs("demo")
    And inputs contain "docs/plans/demo.md"

  Scenario: Init leaves inputs empty for senior-engineer
    When I run "centinela evidence init demo senior-engineer"
    Then ".workflow/demo-senior-engineer.json" inputs is an empty list

  Scenario: Init leaves inputs empty for every non-plan role
    When I run "centinela evidence init demo <role>"
    Then ".workflow/demo-<role>.json" inputs is an empty list
    Examples:
      | role                     |
      | senior-engineer          |
      | ux-ui-specialist         |
      | qa-senior                |
      | validation-specialist    |
      | documentation-specialist |
      | gatekeeper               |

  Scenario: PlanInputs is the only source shared with the validator
    Given the pre-fill path calls "evidence.PlanInputs"
    When PlanInputs delegates for a plan role
    Then it returns "orchestration.RequiredPlanInputs(feature)" verbatim
    And the validator computes its required set from the same RequiredPlanInputs

  Scenario: PlanInputs returns nil for a non-plan role
    When I call PlanInputs("demo", "senior-engineer")
    Then it returns nil

  Scenario: Skeleton stays empty so repair and docs templates are not poisoned
    When I build a skeleton via "evidence.Skeleton" for big-thinker
    Then the skeleton inputs is empty
    And SchemaSkeleton (repair) inputs is empty
    And the docsSpecialistPair inputs is empty

  Scenario: Init leaves outputs empty for every role
    When I run "centinela evidence init demo <role>"
    Then ".workflow/demo-<role>.json" outputs is an empty list
    Examples:
      | role            |
      | big-thinker     |
      | feature-specialist |
      | senior-engineer |

  Scenario: Init leaves edgeCases empty for every role
    When I run "centinela evidence init demo big-thinker"
    Then ".workflow/demo-big-thinker.json" edgeCases is an empty list

  Scenario: Init pre-fill is idempotent under force re-run
    Given I run "centinela evidence init demo big-thinker"
    When I run "centinela evidence init demo big-thinker --force"
    Then the inputs list is identical, sorted, and de-duplicated

  Scenario: Init pre-fill includes a feature brief created after the first init
    Given I run "centinela evidence init demo big-thinker"
    And a new "docs/features/zzz-late.md" brief is added afterward
    When I run "centinela evidence init demo big-thinker --force"
    Then inputs contain "docs/features/zzz-late.md"

  # --- Slice 2: FILL marker + companion skeletons ---

  Scenario: FillSlot renders the canonical marker
    When I call FillSlot("the impl file path")
    Then the result equals "<FILL: the impl file path>"

  Scenario: Companion skeleton seeds role-appropriate FILL slots
    When I run "centinela evidence init demo <role>"
    Then ".workflow/demo-<role>.md" contains "<FILL:"
    And it contains the section header "<header>"
    Examples:
      | role                     | header                |
      | big-thinker              | Problem               |
      | feature-specialist       | Acceptance Criteria   |
      | senior-engineer          | Files Touched         |
      | qa-senior                | Test Inventory        |
      | validation-specialist    | Gates Run             |
      | ux-ui-specialist         | Flow Review           |
      | documentation-specialist | KB Pages              |

  Scenario: Unknown role falls back to the one-line companion placeholder
    When I request the companion template for an unrecognized role
    Then it returns the legacy one-line placeholder
    And it contains no "<FILL:" marker

  Scenario: No FILL marker ever lands in an evidence JSON list field
    When I run "centinela evidence init demo big-thinker"
    Then no inputs, outputs, or edgeCases entry contains "<FILL:"

  # --- Slice 3: artifact new body upgrade ---

  Scenario: Gatekeeper artifact pre-fills Analyzed Specs from existing specs
    Given specs files "specs/a.feature" and "specs/b.feature" exist
    When I run "centinela artifact new demo gatekeeper"
    Then the body "Analyzed Specs" section lists "specs/a.feature" and "specs/b.feature"
    And the list is deterministic and sorted

  Scenario: Gatekeeper artifact Analyzed Specs is an empty list when no specs exist
    Given no "specs/*.feature" files exist
    When I run "centinela artifact new demo gatekeeper"
    Then the "Analyzed Specs" section lists no real spec paths and shows a single "<FILL:" prompt row

  Scenario: Artifact bodies use FILL slots for substance sections
    When I run "centinela artifact new demo gatekeeper"
    Then italic-prose placeholders are replaced by "<FILL:" slots in substance sections

  Scenario: Artifact Status and Date lines stay parseable by validate
    Given I run "centinela artifact new demo gatekeeper"
    When I run "centinela validate"
    Then the literal "**Status:**" and "**Date:**" lines are unchanged and still parse

  # --- Back-compat ---

  Scenario: Pre-existing minimal evidence JSON still validates
    Given a hand-written minimal "demo-big-thinker.json" with manually-listed snapshot inputs
    When I run "centinela evidence validate demo"
    Then validation passes with no schema change required
