Feature: centinela roadmap edit/move/reorder — in-place mutate and relocate
  As an operator (and the Magallanes Plan page consuming roadmap --json)
  I want centinela roadmap edit/update, roadmap move, and roadmap reorder
  So that I can rename/retarget a feature, fix its dependencies, relocate it to
  the correct phase, and reposition it among siblings — without ever breaking
  `roadmap validate` and without hand-editing roadmap.json

  # This feature reuses the raw-feature helpers roadmap-crud-add-remove
  # (shipped) introduced: findFeature/featurePhase, removeFeatureAt/
  # replaceFeatureAt/insertFeatureAt, toRoadmap, ValidateDependencies,
  # validateSlug, validateNoCollision, and the Backlog/Baseline guards.
  # Every mutation is validate-then-mutate-then-write-once: a REJECTED edit,
  # move, or reorder leaves roadmap.json byte-identical (raw rawDoc mutation
  # happens in memory only; the single writeRawRoadmap call is the last step).
  # Scenario titles map 1:1 to Go acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with a valid .workflow/roadmap.json
    And phase "Phase 1: Foundations" is a schedulable, non-Backlog, non-Baseline phase
    And phase "Phase 2: Growth" is a schedulable, non-Backlog, non-Baseline phase

  # ---------------------------------------------------------------------------
  # edit — field-only changes, unspecified fields intact
  # ---------------------------------------------------------------------------

  Scenario: edit changes only the flags provided, leaving unspecified fields intact
    Given "auth-service" exists in phase "Phase 1: Foundations" with
      "description": "Original description" and "archetype": "canonical"
      and no "dependsOn"
    When the operator runs:
      centinela roadmap edit auth-service --description "Updated description"
    Then the command exits with code 0
    And the "auth-service" entry has "description": "Updated description"
    And the "auth-service" entry still has "archetype": "canonical"
    And the "auth-service" entry still has no "dependsOn" (field unchanged, not cleared)
    And every untouched phase in roadmap.json is byte-identical to before the edit

  Scenario: edit --depends-on distinguishes "clear deps" from "unchanged" via the Changed sentinel
    Given "checkout-ui" exists in phase "Phase 1: Foundations" with
      "dependsOn": ["auth-service"]
    When the operator runs:
      centinela roadmap edit checkout-ui --description "New copy"
    Then the command exits with code 0
    And the "checkout-ui" entry still has "dependsOn": ["auth-service"]
      # --depends-on was not passed at all: unchanged, not cleared.

  Scenario: edit --depends-on "" (sentinel present) clears dependencies
    Given "checkout-ui" exists in phase "Phase 1: Foundations" with
      "dependsOn": ["auth-service"]
    When the operator runs:
      centinela roadmap edit checkout-ui --depends-on ""
    Then the command exits with code 0
    And the "checkout-ui" entry has no "dependsOn" field (or an empty array)
      # --depends-on was passed empty: explicitly cleared.

  Scenario: edit --depends-on replaces the dependency list with the provided set
    Given "checkout-ui" exists in phase "Phase 1: Foundations" with
      "dependsOn": ["auth-service"]
    And "billing-api" exists in phase "Phase 1: Foundations"
    When the operator runs:
      centinela roadmap edit checkout-ui --depends-on billing-api
    Then the command exits with code 0
    And the "checkout-ui" entry has "dependsOn": ["billing-api"]

  # ---------------------------------------------------------------------------
  # edit --name — rename + dependent rewrite
  # ---------------------------------------------------------------------------

  Scenario: edit --name renames the feature and rewrites dependents' dependsOn across ALL phases
    Given "auth-service" exists in phase "Phase 1: Foundations"
    And "checkout-ui" (in phase "Phase 1: Foundations") declares "dependsOn": ["auth-service"]
    And "billing-api" (in phase "Phase 2: Growth") declares "dependsOn": ["auth-service"]
    When the operator runs:
      centinela roadmap edit auth-service --name auth-service-v2
    Then the command exits with code 0
    And .workflow/roadmap.json contains no feature named "auth-service"
    And .workflow/roadmap.json contains a feature "auth-service-v2" in phase "Phase 1: Foundations"
    And the "checkout-ui" entry now has "dependsOn": ["auth-service-v2"]
    And the "billing-api" entry now has "dependsOn": ["auth-service-v2"]
    And running "centinela roadmap validate" exits with code 0 (PASS)

  Scenario: edit --name refuses an invalid slug
    Given "auth-service" exists in phase "Phase 1: Foundations"
    When the operator runs:
      centinela roadmap edit auth-service --name "Not_Kebab!"
    Then the command exits with a non-zero code
    And stderr contains "invalid feature slug"
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: edit --name refuses a collision with an existing feature, naming the owning phase
    Given "auth-service" exists in phase "Phase 1: Foundations"
    And "billing-api" exists in phase "Phase 2: Growth"
    When the operator runs:
      centinela roadmap edit auth-service --name billing-api
    Then the command exits with a non-zero code
    And stderr contains "\"billing-api\" already exists in phase \"Phase 2: Growth\""
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: edit --name to the SAME name is a no-op — no dependents rewritten, file byte-identical
    Given "auth-service" exists in phase "Phase 1: Foundations"
    And "checkout-ui" declares "dependsOn": ["auth-service"]
    When the operator runs:
      centinela roadmap edit auth-service --name auth-service
    Then the command exits with code 0
    And .workflow/roadmap.json is byte-identical to before the command ran

  # ---------------------------------------------------------------------------
  # edit — cycle / unknown-dep rejection, byte-identical
  # ---------------------------------------------------------------------------

  Scenario Outline: edit refuses a dependsOn that is unknown or introduces a cycle, leaving roadmap.json untouched
    Given a fixed on-disk roadmap.json (captured as "before")
    And "checkout-ui" exists in phase "Phase 1: Foundations"
    When the operator runs:
      centinela roadmap edit checkout-ui --depends-on "<deps>"
    Then the command exits with a non-zero code
    And stderr contains "<error-substring>"
    And .workflow/roadmap.json is byte-identical to "before"

    Examples:
      | deps                         | error-substring                    |
      | ghost-feature                | depends on unknown feature         |
      | checkout-ui                  | roadmap dependency cycle detected  |

  Scenario: edit --depends-on introduces a multi-hop cycle across two features is refused
    Given "auth-service" (phase "Phase 1: Foundations") declares "dependsOn": ["checkout-ui"]
    And "checkout-ui" exists in phase "Phase 1: Foundations" with no "dependsOn"
    When the operator runs:
      centinela roadmap edit checkout-ui --depends-on auth-service
    Then the command exits with a non-zero code
    And stderr contains "roadmap dependency cycle detected"
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: edit/update a slug that does not exist errors "not found"
    Given no feature named "ghost-feature" exists anywhere in the roadmap
    When the operator runs:
      centinela roadmap edit ghost-feature --description "x"
    Then the command exits with a non-zero code
    And stderr contains "not found"
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: update is an alias for edit
    Given "auth-service" exists in phase "Phase 1: Foundations"
    When the operator runs:
      centinela roadmap update auth-service --description "Updated via alias"
    Then the command exits with code 0
    And the "auth-service" entry has "description": "Updated via alias"

  # ---------------------------------------------------------------------------
  # move — general phase→phase relocation
  # ---------------------------------------------------------------------------

  Scenario: move relocates a feature to the target phase, appending by default
    Given "checkout-ui" exists in phase "Phase 1: Foundations"
    And phase "Phase 2: Growth" has existing features "billing-api", "reporting"
    When the operator runs:
      centinela roadmap move checkout-ui --to-phase "Phase 2: Growth"
    Then the command exits with code 0
    And "checkout-ui" is no longer in phase "Phase 1: Foundations"
    And "checkout-ui" is now the last feature in phase "Phase 2: Growth"
    And every untouched phase in roadmap.json is byte-identical to before the move

  Scenario Outline: move --before/--after anchors the feature at the first, last, or middle position
    Given "checkout-ui" exists in phase "Phase 1: Foundations"
    And phase "Phase 2: Growth" has features in order: "billing-api", "reporting", "invoicing"
    When the operator runs:
      centinela roadmap move checkout-ui --to-phase "Phase 2: Growth" <anchor-flag> "<anchor>"
    Then the command exits with code 0
    And phase "Phase 2: Growth" now has features in order: "<expected-order>"

    Examples:
      | anchor-flag | anchor      | expected-order                                     |
      | --before    | billing-api | checkout-ui, billing-api, reporting, invoicing     |
      | --after     | invoicing   | billing-api, reporting, invoicing, checkout-ui     |
      | --after     | billing-api | billing-api, checkout-ui, reporting, invoicing     |
      | --before    | invoicing   | billing-api, reporting, checkout-ui, invoicing     |

  Scenario: move preserves the feature's draft status and quality entries
    Given "new-widget" is a draft feature in phase "Phase 1: Foundations"
    When the operator runs:
      centinela roadmap move new-widget --to-phase "Phase 2: Growth"
    Then the command exits with code 0
    And "new-widget" is now in phase "Phase 2: Growth"
    And the "new-widget" entry still has "draft": true

  Scenario: move preserves quality entries for an already-promoted feature
    Given "auth-service" exists in phase "Phase 1: Foundations" with a
      .workflow/roadmap-quality.json entry for "auth-service"
    When the operator runs:
      centinela roadmap move auth-service --to-phase "Phase 2: Growth"
    Then the command exits with code 0
    And "auth-service" is now in phase "Phase 2: Growth"
    And .workflow/roadmap-quality.json still contains an "auth-service" entry, unchanged

  Scenario: move is allowed for a feature that another feature depends on (dependency is by name, not phase)
    Given "auth-service" exists in phase "Phase 1: Foundations"
    And "checkout-ui" declares "dependsOn": ["auth-service"]
    When the operator runs:
      centinela roadmap move auth-service --to-phase "Phase 2: Growth"
    Then the command exits with code 0
    And the "checkout-ui" entry still has "dependsOn": ["auth-service"]
    And running "centinela roadmap validate" exits with code 0 (PASS)

  Scenario: move --before/--after an anchor that IS the feature itself is a no-op, byte-identical
    Given "checkout-ui" exists in phase "Phase 1: Foundations" as the only feature
    When the operator runs:
      centinela roadmap move checkout-ui --to-phase "Phase 1: Foundations" --after checkout-ui
    Then the command exits with code 0
    And .workflow/roadmap.json is byte-identical to before the command ran

  # ---------------------------------------------------------------------------
  # move — guards, byte-identical on rejection
  # ---------------------------------------------------------------------------

  Scenario Outline: move refuses Backlog/Baseline as source or target, and unknown phase/anchor
    Given a fixed on-disk roadmap.json (captured as "before")
    And "checkout-ui" exists in phase "Phase 1: Foundations"
    When the operator runs:
      centinela roadmap move "<slug>" --to-phase "<target>" <extra-flags>
    Then the command exits with a non-zero code
    And stderr contains "<error-substring>"
    And .workflow/roadmap.json is byte-identical to "before"

    Examples:
      | slug            | target                | extra-flags              | error-substring     |
      | checkout-ui     | Backlog                |                          | unknown phase       |
      | checkout-ui     | Baseline                |                          | unknown phase       |
      | checkout-ui     | Phase 9: Nonexistent    |                          | unknown phase       |
      | checkout-ui     | Phase 2: Growth         | --before ghost-anchor    | unknown feature     |

  Scenario: move of a feature currently in the Backlog phase is refused, directing to promote
    Given "legacy-finding" exists in phase "Backlog"
    When the operator runs:
      centinela roadmap move legacy-finding --to-phase "Phase 2: Growth"
    Then the command exits with a non-zero code
    And stderr contains "promote"
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: move of a feature currently in the Baseline phase is refused
    Given "shipped-baseline-item" exists in phase "Baseline"
    When the operator runs:
      centinela roadmap move shipped-baseline-item --to-phase "Phase 2: Growth"
    Then the command exits with a non-zero code
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: move a slug that does not exist errors "not found"
    Given no feature named "ghost-feature" exists anywhere in the roadmap
    When the operator runs:
      centinela roadmap move ghost-feature --to-phase "Phase 2: Growth"
    Then the command exits with a non-zero code
    And stderr contains "not found"
    And .workflow/roadmap.json is byte-identical to before the command ran

  # ---------------------------------------------------------------------------
  # reorder — within/across phase reposition by anchor
  # ---------------------------------------------------------------------------

  Scenario: reorder repositions a feature within its own phase
    Given phase "Phase 1: Foundations" has features in order: "auth-service", "checkout-ui", "billing-api"
    When the operator runs:
      centinela roadmap reorder billing-api --before auth-service
    Then the command exits with code 0
    And phase "Phase 1: Foundations" now has features in order: "billing-api", "auth-service", "checkout-ui"
    And every untouched phase in roadmap.json is byte-identical to before the reorder

  Scenario: reorder repositions a feature relative to an anchor in a different phase, moving it across
    Given "checkout-ui" exists in phase "Phase 1: Foundations"
    And phase "Phase 2: Growth" has features in order: "billing-api", "reporting"
    When the operator runs:
      centinela roadmap reorder checkout-ui --after billing-api
    Then the command exits with code 0
    And "checkout-ui" is no longer in phase "Phase 1: Foundations"
    And phase "Phase 2: Growth" now has features in order: "billing-api", "checkout-ui", "reporting"

  Scenario: a no-op reorder (already adjacent in the requested position) leaves the file byte-identical
    Given phase "Phase 1: Foundations" has features in order: "auth-service", "checkout-ui", "billing-api"
    When the operator runs:
      centinela roadmap reorder checkout-ui --after auth-service
    Then the command exits with code 0
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: reorder into a Backlog/Baseline phase (via an anchor there) is refused
    Given "legacy-finding" exists in phase "Backlog"
    And "checkout-ui" exists in phase "Phase 1: Foundations"
    When the operator runs:
      centinela roadmap reorder checkout-ui --after legacy-finding
    Then the command exits with a non-zero code
    And stderr contains "unknown phase"
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: reorder a slug that does not exist errors "not found"
    Given no feature named "ghost-feature" exists anywhere in the roadmap
    When the operator runs:
      centinela roadmap reorder ghost-feature --before auth-service
    Then the command exits with a non-zero code
    And stderr contains "not found"
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: reorder against an unknown anchor errors clearly and leaves the file untouched
    Given "checkout-ui" exists in phase "Phase 1: Foundations"
    When the operator runs:
      centinela roadmap reorder checkout-ui --before ghost-anchor
    Then the command exits with a non-zero code
    And stderr contains "unknown feature"
    And .workflow/roadmap.json is byte-identical to before the command ran

  # ---------------------------------------------------------------------------
  # cross-cutting — empty/missing roadmap.json, atomic single write
  # ---------------------------------------------------------------------------

  Scenario Outline: edit/move/reorder against an empty roadmap errors cleanly with no silent mutation
    Given .workflow/roadmap.json is exactly {"phases":[]}
    When the operator runs:
      centinela roadmap <command> ghost-feature <extra-flags>
    Then the command exits with a non-zero code
    And stderr contains "not found"
    And .workflow/roadmap.json remains exactly {"phases":[]}

    Examples:
      | command | extra-flags               |
      | edit    | --description "x"         |
      | move    | --to-phase "Phase 1"      |
      | reorder | --before anchor           |

  Scenario: edit/move/reorder against a missing or malformed roadmap.json surfaces an error and leaves the file untouched
    Given .workflow/roadmap.json does not exist
    When the operator runs:
      centinela roadmap edit auth-service --description "x"
    Then the command exits with a non-zero code
    And stderr contains an error message
    And .workflow/roadmap.json is still absent

  Scenario: every mutation performs exactly one atomic write — a rejected edit/move/reorder writes nothing
    Given a fixed on-disk roadmap.json (captured as "before") with its file mtime recorded
    When the operator runs an edit, move, or reorder that gets refused
      (invalid slug, cycle, unknown phase, or unknown anchor)
    Then the command exits with a non-zero code
    And .workflow/roadmap.json's mtime and contents are unchanged from "before"
      # Validation happens in memory against the decoded raw document before
      # the single writeRawRoadmap call; a refusal never reaches that call.
