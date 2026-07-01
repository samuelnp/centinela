Feature: centinela roadmap --json contract — machine-readable roadmap views
  As a Magallanes backend rendering a "Plan project" page
  I want centinela roadmap [--json], roadmap ready --json, and roadmap show --json
  So that I can shell out to Centinela for a stable, machine-readable roadmap
  view without re-implementing status/readiness derivation

  # RoadmapView is a projection built from the persisted Roadmap by BuildView():
  # ordered phases[] -> ordered features[] carrying name, phase, status
  # (planned|in-progress|done), readiness (ready|blocked), dependsOn[],
  # blockedBy[], plus a top-level counts object. status derives from
  # FeatureStatus(name); readiness/blockedBy derive from
  # DeriveReadiness/classifyFeature. counts is scoped exactly like Summary():
  # Backlog/Baseline phases are non-schedulable and excluded from counts and
  # from ready. `roadmap show --json` is a different, simpler contract: it
  # dumps the persisted typed Roadmap verbatim (including Backlog/Baseline),
  # with no derived fields at all. Scenario titles map 1:1 to Go acceptance
  # tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with a valid .workflow/roadmap.json

  # ---------------------------------------------------------------------------
  # Design lock: readiness for done/in-progress rows (Outstanding question 1)
  # ---------------------------------------------------------------------------

  Scenario: readiness is empty for a done feature; status alone carries the signal
    Given the roadmap has a feature "auth-service" in phase "Q1" with status "done"
    When the operator runs:
      centinela roadmap --json
    Then the "auth-service" entry has "status": "done"
    And the "auth-service" entry has no "readiness" field
    And the "auth-service" entry has no "blockedBy" field

  Scenario: readiness is empty for an in-progress feature; status alone carries the signal
    Given the roadmap has a feature "billing-api" in phase "Q1" with status "in-progress"
    When the operator runs:
      centinela roadmap --json
    Then the "billing-api" entry has "status": "in-progress"
    And the "billing-api" entry has no "readiness" field
    And the "billing-api" entry has no "blockedBy" field

  Scenario: readiness is "ready" for a planned feature whose dependencies are all done
    Given the roadmap has a feature "checkout-ui" in phase "Q1" with status "planned"
    And "checkout-ui" depends on "auth-service" which has status "done"
    When the operator runs:
      centinela roadmap --json
    Then the "checkout-ui" entry has "status": "planned"
    And the "checkout-ui" entry has "readiness": "ready"
    And the "checkout-ui" entry has no "blockedBy" field

  Scenario: readiness is "blocked" for a planned feature with an unmet dependency
    Given the roadmap has a feature "reporting" in phase "Q1" with status "planned"
    And "reporting" depends on "billing-api" which has status "in-progress"
    When the operator runs:
      centinela roadmap --json
    Then the "reporting" entry has "status": "planned"
    And the "reporting" entry has "readiness": "blocked"
    And the "reporting" entry has "blockedBy": ["billing-api"]

  Scenario Outline: status/readiness convention example row
    Given a feature with status "<status>" and unmet dependencies "<unmet>"
    When BuildView derives its FeatureView
    Then the emitted status is "<status>"
    And the emitted readiness is "<readiness>"

    Examples:
      | status      | unmet          | readiness |
      | done        | (n/a)          | (omitted) |
      | in-progress | (n/a)          | (omitted) |
      | planned     | none           | ready     |
      | planned     | one-or-more    | blocked   |

  # ---------------------------------------------------------------------------
  # Happy path: roadmap --json full RoadmapView
  # ---------------------------------------------------------------------------

  Scenario: roadmap --json emits ordered phases and features with counts
    Given a roadmap with phase "Q1" containing features "auth-service" (done),
      "billing-api" (in-progress), "checkout-ui" (planned, depends on "auth-service"),
      and "reporting" (planned, depends on "billing-api")
    When the operator runs:
      centinela roadmap --json
    Then the command exits with code 0
    And the output is valid JSON
    And the JSON has top-level fields "phases" and "counts"
    And "phases" contains one entry named "Q1" with a "features" array of length 4
    And the features appear in declared roadmap order: "auth-service", "billing-api", "checkout-ui", "reporting"
    And each feature entry has fields "name", "phase", "status", "dependsOn"
    And "counts" is {"planned": 2, "inProgress": 1, "done": 1}

  Scenario: Phase with zero features renders as an empty features array
    Given a roadmap with a schedulable phase "Q2" that has no features
    When the operator runs:
      centinela roadmap --json
    Then the "Q2" phase entry has "features": []

  Scenario: Non-schedulable phases are excluded from roadmap --json
    Given a roadmap containing phases "Backlog", "Baseline", and "Q1"
    And "Backlog" and "Baseline" are non-schedulable phases
    When the operator runs:
      centinela roadmap --json
    Then the "phases" array does not contain an entry named "Backlog"
    And the "phases" array does not contain an entry named "Baseline"
    And "counts" only reflects features under "Q1"

  # ---------------------------------------------------------------------------
  # roadmap ready --json
  # ---------------------------------------------------------------------------

  Scenario: roadmap ready --json emits the ready feature names in declared order
    Given a roadmap with phase "Q1" containing features "auth-service" (done),
      "checkout-ui" (planned, depends on "auth-service"), and "reporting" (planned,
      depends on "billing-api" which is "in-progress")
    When the operator runs:
      centinela roadmap ready --json
    Then the command exits with code 0
    And the output is valid JSON array
    And the array equals ["checkout-ui"]

  Scenario: ready --json set is identical to the readiness:ready set in roadmap --json
    Given a roadmap with a mix of done, in-progress, ready, and blocked planned features
    When the operator runs:
      centinela roadmap ready --json
    And the operator runs:
      centinela roadmap --json
    Then the names in the ready --json array exactly match the names of every
      feature in roadmap --json whose "readiness" field is "ready"
    And no name appears in one set but not the other

  Scenario: ready --json when nothing is ready emits an empty array, not null
    Given a roadmap where every planned feature has at least one unmet dependency
    When the operator runs:
      centinela roadmap ready --json
    Then the command exits with code 0
    And the output is exactly "[]"

  # ---------------------------------------------------------------------------
  # roadmap show / list --json — persisted Roadmap verbatim
  # ---------------------------------------------------------------------------

  Scenario: roadmap show --json emits the persisted Roadmap verbatim, including non-schedulable phases
    Given a roadmap containing phases "Backlog", "Baseline", and "Q1"
    When the operator runs:
      centinela roadmap show --json
    Then the command exits with code 0
    And the output is valid JSON
    And the "phases" array contains entries named "Backlog", "Baseline", and "Q1"
    And no feature entry carries a "status" or "readiness" field
    And the output matches the on-disk .workflow/roadmap.json content field-for-field

  Scenario: roadmap list --json is an alias for roadmap show --json
    Given a roadmap with at least one phase and one feature
    When the operator runs:
      centinela roadmap list --json
    And the operator runs:
      centinela roadmap show --json
    Then both commands produce byte-identical output

  Scenario: roadmap show (no flag) prints the same text as roadmap (no flag)
    Given a roadmap with at least one phase and one feature
    When the operator runs:
      centinela roadmap show
    And the operator runs:
      centinela roadmap
    Then both outputs are byte-identical to ui.RenderRoadmap's rendering

  # ---------------------------------------------------------------------------
  # Determinism / byte-stability
  # ---------------------------------------------------------------------------

  Scenario: roadmap --json is byte-identical across two consecutive runs
    Given a fixed on-disk roadmap with multiple phases and features
    When the operator runs centinela roadmap --json twice in succession
    Then both outputs are byte-identical

  Scenario: roadmap ready --json is byte-identical across two consecutive runs
    Given a fixed on-disk roadmap
    When the operator runs centinela roadmap ready --json twice in succession
    Then both outputs are byte-identical

  Scenario: roadmap show --json is byte-identical across two consecutive runs
    Given a fixed on-disk roadmap
    When the operator runs centinela roadmap show --json twice in succession
    Then both outputs are byte-identical

  # ---------------------------------------------------------------------------
  # Empty roadmap
  # ---------------------------------------------------------------------------

  Scenario: Empty roadmap --json emits empty phases and all-zero counts
    Given a roadmap file that exists and contains no phases
    When the operator runs:
      centinela roadmap --json
    Then the command exits with code 0
    And the output is exactly {"phases":[],"counts":{"planned":0,"inProgress":0,"done":0}}

  Scenario: Empty roadmap ready --json emits an empty array
    Given a roadmap file that exists and contains no phases
    When the operator runs:
      centinela roadmap ready --json
    Then the command exits with code 0
    And the output is exactly "[]"

  Scenario: Empty roadmap show --json emits the persisted empty structure verbatim
    Given a roadmap file that exists and contains no phases
    When the operator runs:
      centinela roadmap show --json
    Then the command exits with code 0
    And the output is valid JSON matching the persisted empty Roadmap struct

  # ---------------------------------------------------------------------------
  # Missing / malformed source — no partial JSON, non-zero exit
  # ---------------------------------------------------------------------------

  Scenario: Missing roadmap.json fails roadmap --json with a stderr error and no stdout JSON
    Given no .workflow/roadmap.json file exists
    When the operator runs:
      centinela roadmap --json
    Then the command exits with a non-zero code
    And stderr contains an error message
    And stdout contains no JSON output

  Scenario: Missing roadmap.json fails roadmap ready --json with a stderr error and no stdout JSON
    Given no .workflow/roadmap.json file exists
    When the operator runs:
      centinela roadmap ready --json
    Then the command exits with a non-zero code
    And stderr contains an error message
    And stdout contains no JSON output

  Scenario: Missing roadmap.json fails roadmap show --json with a stderr error and no stdout JSON
    Given no .workflow/roadmap.json file exists
    When the operator runs:
      centinela roadmap show --json
    Then the command exits with a non-zero code
    And stderr contains an error message
    And stdout contains no JSON output

  Scenario: Missing roadmap.json also fails the text-mode commands the same way
    Given no .workflow/roadmap.json file exists
    When the operator runs:
      centinela roadmap
    Then the command exits with a non-zero code
    And stderr contains an error message

  Scenario: Malformed roadmap.json (invalid JSON) is rejected by Load, not partially rendered
    Given .workflow/roadmap.json contains invalid JSON syntax
    When the operator runs:
      centinela roadmap --json
    Then the command exits with a non-zero code
    And stderr contains an error message
    And stdout contains no JSON output

  Scenario: Dependency-cycle roadmap.json is rejected by Load, not partially rendered
    Given .workflow/roadmap.json declares a dependency cycle between two features
    When the operator runs:
      centinela roadmap --json
    Then the command exits with a non-zero code
    And stderr contains an error message
    And stdout contains no JSON output

  # ---------------------------------------------------------------------------
  # Text-output regression guard — unchanged when --json is absent
  # ---------------------------------------------------------------------------

  Scenario: roadmap (no flag) text output is byte-for-byte unchanged from today
    Given a fixed on-disk roadmap with multiple phases and features
    When the operator runs:
      centinela roadmap
    Then the output is byte-identical to the pre-existing ui.RenderRoadmap output
    And the output contains no JSON structure

  Scenario: roadmap ready (no flag) text output is byte-for-byte unchanged from today
    Given a fixed on-disk roadmap with a mix of ready and blocked features
    When the operator runs:
      centinela roadmap ready
    Then the output is byte-identical to the pre-existing ui.RenderReadyList output
    And the output contains no JSON structure
