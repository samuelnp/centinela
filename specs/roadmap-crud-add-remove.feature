Feature: centinela roadmap add/remove — direct authoring with a draft lifecycle
  As an operator (and the Magallanes Plan page consuming roadmap --json)
  I want centinela roadmap add, roadmap remove/rm, and a generalized
  roadmap promote that finalizes an in-place draft
  So that I can author a roadmap feature directly in a real phase, delete one,
  and finalize a drafted item — without ever breaking `roadmap validate`

  # A "draft" is `Feature.Draft bool` (json:"draft,omitempty"), persisted in
  # roadmap.json. It has exactly ONE coverage-set exemption hook
  # (NonBacklogFeatureSet: `if f.Draft { continue }`) and three more READERS
  # that must independently agree with the persisted field:
  #   1. NonBacklogFeatureSet  — draft exempt from the >=9 analysis/quality set
  #   2. classifyFeature/ReadySet — draft classifies State:"draft", excluded
  #   3. Summary()             — draft not counted as committed planned work
  #   4. BuildView (roadmap --json) — draft:true, readiness:"draft"
  # `centinela start <draft>` is refused, mirroring the Backlog refusal.
  # Every mutation is validate-then-mutate-then-write-once: a REJECTED add,
  # remove, or promote leaves roadmap.json byte-identical (raw rawDoc mutation
  # happens in memory only; the single writeRawRoadmap call is the last step).
  # Scenario titles map 1:1 to Go acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with a valid .workflow/roadmap.json
    And phase "Phase 1: Foundations" is a schedulable, non-Backlog, non-Baseline phase

  # ---------------------------------------------------------------------------
  # add — happy path
  # ---------------------------------------------------------------------------

  Scenario: add creates a draft in a chosen schedulable phase and validate stays PASS
    Given the roadmap has phase "Phase 1: Foundations" with no feature named "new-widget"
    When the operator runs:
      centinela roadmap add new-widget --phase "Phase 1: Foundations"
    Then the command exits with code 0
    And .workflow/roadmap.json contains a feature "new-widget" in phase "Phase 1: Foundations"
      with "draft": true
    And .workflow/roadmap.json is valid JSON and parses via Load without error
    And running "centinela roadmap validate" exits with code 0 (PASS)
    And every untouched phase in roadmap.json is byte-identical to before the add

  Scenario: add accepts optional description, depends-on, and archetype flags
    Given the roadmap has phase "Phase 1: Foundations" with a done feature "auth-service"
    When the operator runs:
      centinela roadmap add new-widget --phase "Phase 1: Foundations" \
        --description "Adds the widget" --depends-on auth-service --archetype canonical
    Then the command exits with code 0
    And the "new-widget" entry has "description": "Adds the widget"
    And the "new-widget" entry has "dependsOn": ["auth-service"]
    And the "new-widget" entry has "archetype": "canonical"
    And the "new-widget" entry has "draft": true

  # ---------------------------------------------------------------------------
  # add — rejections, each byte-identical on reject
  # ---------------------------------------------------------------------------

  Scenario Outline: add rejects invalid input and leaves roadmap.json byte-identical
    Given a fixed on-disk roadmap.json (captured as "before")
    When the operator runs:
      centinela roadmap add "<slug>" --phase "<phase>" <extra-flags>
    Then the command exits with a non-zero code
    And stderr contains "<error-substring>"
    And .workflow/roadmap.json is byte-identical to "before"

    Examples:
      | slug                | phase                     | extra-flags                  | error-substring                              |
      | Not_Kebab!           | Phase 1: Foundations      |                               | invalid feature slug                         |
      | auth-service         | Phase 1: Foundations      |                               | slug collision                               |
      | new-widget           | Phase 9: Nonexistent      |                               | unknown phase                                |
      | new-widget           | Backlog                   |                               | unknown phase                                |
      | new-widget           | Baseline                  |                               | unknown phase                                |
      | new-widget           | Phase 1: Foundations      | --depends-on ghost-feature    | depends on unknown feature                   |
      | new-widget           | Phase 1: Foundations      | --depends-on new-widget       | roadmap dependency cycle detected            |

    # Notes on the table:
    # - "auth-service" is assumed pre-existing in another phase for the
    #   duplicate-name row — the error must name the OWNING phase.
    # - Backlog/Baseline rows assert add refuses non-schedulable targets:
    #   drafts live only in schedulable phases; Backlog authoring stays `defer`.
    # - The self-dependency row is the smallest reproducible cycle: a feature
    #   cannot depend on itself, and ValidateDependencies must catch it before
    #   any write (the draft does not exist yet when the cycle is evaluated,
    #   so this models "a dependsOn that would create a cycle").

  Scenario: add duplicate name across a different phase reports the owning phase
    Given "billing-api" exists in phase "Phase 2: Growth"
    When the operator runs:
      centinela roadmap add billing-api --phase "Phase 1: Foundations"
    Then the command exits with a non-zero code
    And stderr contains "\"billing-api\" already exists in phase \"Phase 2: Growth\""
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: add against an empty roadmap errors "unknown phase" with no silent phase creation
    Given .workflow/roadmap.json is exactly {"phases":[]}
    When the operator runs:
      centinela roadmap add new-widget --phase "Phase 1: Foundations"
    Then the command exits with a non-zero code
    And stderr contains "unknown phase"
    And .workflow/roadmap.json remains exactly {"phases":[]}

  Scenario: add against a missing or malformed roadmap.json surfaces an error and leaves the file untouched
    Given .workflow/roadmap.json does not exist
    When the operator runs:
      centinela roadmap add new-widget --phase "Phase 1: Foundations"
    Then the command exits with a non-zero code
    And stderr contains an error message
    And .workflow/roadmap.json is still absent

  # ---------------------------------------------------------------------------
  # remove — happy path and not-found
  # ---------------------------------------------------------------------------

  Scenario: remove deletes a planned feature and leaves the file valid
    Given "old-widget" exists in phase "Phase 1: Foundations" with status "planned"
    And no other feature depends on "old-widget"
    When the operator runs:
      centinela roadmap remove old-widget
    Then the command exits with code 0
    And .workflow/roadmap.json contains no feature named "old-widget"
    And .workflow/roadmap.json is valid JSON and parses via Load without error
    And every untouched phase in roadmap.json is byte-identical to before the remove

  Scenario: rm is an alias for remove
    Given "old-widget" exists in phase "Phase 1: Foundations" with status "planned"
    And no other feature depends on "old-widget"
    When the operator runs:
      centinela roadmap rm old-widget
    Then the command exits with code 0
    And .workflow/roadmap.json contains no feature named "old-widget"

  Scenario: remove a feature that does not exist errors "not found"
    Given no feature named "ghost-feature" exists anywhere in the roadmap
    When the operator runs:
      centinela roadmap remove ghost-feature
    Then the command exits with a non-zero code
    And stderr contains "not found"
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: remove the last feature of a phase leaves the phase with an empty features array
    Given phase "Phase 3: Solo" contains exactly one planned feature "lonely-feature"
    And no other feature depends on "lonely-feature"
    When the operator runs:
      centinela roadmap remove lonely-feature
    Then the command exits with code 0
    And phase "Phase 3: Solo" still exists in roadmap.json with "features": []

  # ---------------------------------------------------------------------------
  # remove — guards
  # ---------------------------------------------------------------------------

  Scenario: remove is refused when another feature depends on it, naming the dependents
    Given "auth-service" exists in phase "Phase 1: Foundations" with status "planned"
    And "checkout-ui" (planned) declares "dependsOn": ["auth-service"]
    When the operator runs:
      centinela roadmap remove auth-service
    Then the command exits with a non-zero code
    And stderr contains "checkout-ui"
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: remove is refused when the only dependent is itself a draft
    Given "auth-service" exists in phase "Phase 1: Foundations" with status "planned"
    And "draft-consumer" is a draft feature that declares "dependsOn": ["auth-service"]
    When the operator runs:
      centinela roadmap remove auth-service
    Then the command exits with a non-zero code
    And stderr contains "draft-consumer"
    And .workflow/roadmap.json is byte-identical to before the command ran
    # A draft is still a real dependent: dependency guards do not special-case
    # drafts, only the analysis/quality coverage set does.

  Scenario Outline: remove is refused for an in-progress or done feature
    Given "<feature>" exists in phase "Phase 1: Foundations" with status "<status>"
    And no other feature depends on "<feature>"
    When the operator runs:
      centinela roadmap remove "<feature>"
    Then the command exits with a non-zero code
    And stderr contains "<status>"
    And .workflow/roadmap.json is byte-identical to before the command ran

    Examples:
      | feature       | status      |
      | billing-api   | in-progress |
      | auth-service  | done        |

  # ---------------------------------------------------------------------------
  # promote — generalized, branched by the slug's CURRENT location
  # ---------------------------------------------------------------------------

  Scenario: promote finalizes a draft in place — no phase move, draft cleared, artifacts written
    Given "new-widget" is a draft feature already in schedulable phase "Phase 1: Foundations"
    When the operator runs:
      centinela roadmap promote new-widget --scores 9,9,9,9,9,9
    Then the command exits with code 0
    And "new-widget" is still in phase "Phase 1: Foundations" (no phase move occurred)
    And the "new-widget" feature entry no longer has "draft": true
    And .workflow/roadmap-analysis.json contains a "new-widget" entry
    And .workflow/roadmap-quality.json contains a "new-widget" entry with overall score 9
    And running "centinela roadmap validate" exits with code 0 (PASS)

  Scenario: promote of a Backlog finding still moves it into --phase (unchanged behavior)
    Given "legacy-finding" is a Backlog finding (not a draft, phase "Backlog")
    When the operator runs:
      centinela roadmap promote legacy-finding --phase "Phase 1: Foundations" --scores 9,9,9,9,9,9
    Then the command exits with code 0
    And "legacy-finding" is no longer in the "Backlog" phase
    And "legacy-finding" is now in phase "Phase 1: Foundations"
    And .workflow/roadmap-analysis.json contains a "legacy-finding" entry
    And .workflow/roadmap-quality.json contains a "legacy-finding" entry

  Scenario: promote of a non-draft, non-Backlog slug is a clear error
    Given "auth-service" exists in phase "Phase 1: Foundations" with status "planned"
    And "auth-service" is NOT a draft and NOT in the Backlog phase
    When the operator runs:
      centinela roadmap promote auth-service --scores 9,9,9,9,9,9
    Then the command exits with a non-zero code
    And stderr contains an error naming "auth-service" as neither a draft nor a Backlog finding
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: promote of a draft with overall score below 9 is refused, draft flag left intact
    Given "new-widget" is a draft feature already in schedulable phase "Phase 1: Foundations"
    When the operator runs:
      centinela roadmap promote new-widget --scores 9,9,9,9,9,8
    Then the command exits with a non-zero code
    And stderr contains "overall score must be at least 9"
    And the "new-widget" feature entry still has "draft": true
    And .workflow/roadmap.json is byte-identical to before the command ran
    And .workflow/roadmap-analysis.json and .workflow/roadmap-quality.json are unchanged

  # ---------------------------------------------------------------------------
  # THE four-reader draft invariant — the single highest-value scenario
  # ---------------------------------------------------------------------------

  Scenario: a freshly-added draft simultaneously satisfies all four draft readers
    Given the roadmap has phase "Phase 1: Foundations" with no feature named "new-widget"
    When the operator runs:
      centinela roadmap add new-widget --phase "Phase 1: Foundations"
    Then running "centinela roadmap validate" exits with code 0 (PASS)
      # Reader 1: NonBacklogFeatureSet exempts "new-widget" from the >=9
      # analysis/quality coverage set, so validate does not demand scores for it.
    And "new-widget" does NOT appear in the output of "centinela roadmap ready --json"
      # Reader 2: classifyFeature returns State:"draft" (not "ready"), so
      # ReadySet excludes it.
    And "centinela roadmap --json" reports "counts.planned" excluding "new-widget"
      # Reader 3: Summary()/tally does not count an unscored draft as committed
      # planned work.
    And the "new-widget" entry in "centinela roadmap --json" has "draft": true
      and "readiness": "draft"
      # Reader 4: BuildView/buildFeatureView flows the draft state through with
      # a non-empty, non-"ready" readiness value.
    And the operator runs:
      centinela start new-widget
    Then the command exits with a non-zero code
      # start_guard refuses a draft, mirroring the Backlog refusal — no scores
      # yet means starting it would bypass the >=9 gate.
    And stderr contains "draft"

  # ---------------------------------------------------------------------------
  # rendering, determinism, regression
  # ---------------------------------------------------------------------------

  Scenario: ROADMAP.md renders a deterministic " *(draft)*" marker for a draft feature
    Given "new-widget" is a draft feature in phase "Phase 1: Foundations"
    When ROADMAP.md is regenerated from roadmap.json (roadmap-doc-sync)
    Then the "new-widget" bullet line ends with " *(draft)*"
    And regenerating ROADMAP.md twice in succession produces byte-identical output

  Scenario: roadmap --json is byte-identical across two consecutive runs after add/remove/promote
    Given a fixed on-disk roadmap containing at least one draft feature
    When the operator runs centinela roadmap --json twice in succession
    Then both outputs are byte-identical

  Scenario: existing non-draft roadmap --json output is unchanged by the draft extension
    Given a roadmap with only non-draft features across done, in-progress, ready, and blocked states
    When the operator runs:
      centinela roadmap --json
    Then no feature entry carries a "draft" field
    And no feature entry has "readiness": "draft"
    And the output is otherwise identical to the pre-existing roadmap-json-contract shape
      (fields "name", "phase", "status", "dependsOn", and conditionally "readiness"/"blockedBy")
