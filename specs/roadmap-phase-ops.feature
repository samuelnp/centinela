Feature: centinela roadmap phase add/rename/remove — phase-level structural ops
  As an operator (and the Magallanes Plan page consuming roadmap --json)
  I want centinela roadmap phase add/rename/remove
  So that I can insert, rename, or delete a phase without hand-editing
  roadmap.json, without ever breaking `roadmap validate`

  # Final feature of the Roadmap Editing Suite. Reuses the raw-feature helpers
  # from roadmap-crud-add-remove/roadmap-edit-move: toRoadmap, ValidateDependencies,
  # ValidateAnalysis, ValidateQuality, isBacklogPhaseName/isBaselinePhaseName,
  # decodePhase/setPhase/knownPhaseList. The NEW complexity: insertPhaseAt/
  # removePhaseAt must reindex the raw `dirty` map (keyed by phase index) so a
  # phase inserted/removed in the middle does not corrupt a later, already-dirty
  # phase's rendered bytes. Every mutation is validate-then-mutate-then-write-once:
  # a REJECTED add/rename/remove leaves roadmap.json byte-identical.
  # Scenario titles map 1:1 to Go acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with a valid .workflow/roadmap.json
    And phase "Phase 1: Foundations" is a schedulable, non-Backlog, non-Baseline phase
    And phase "Phase 2: Growth" is a schedulable, non-Backlog, non-Baseline phase
    And phase "Backlog" is the reserved Backlog phase, positioned last
    And no phase named "Baseline" exists in this fixture unless stated otherwise

  # ---------------------------------------------------------------------------
  # phase add — insertion position, note, byte-identical untouched phases
  # ---------------------------------------------------------------------------

  Scenario: phase add with --after inserts immediately after the named phase
    Given the roadmap has phases in order: "Phase 1: Foundations", "Phase 2: Growth", "Backlog"
    When the operator runs:
      centinela roadmap phase add "Phase 1.5: Bridge" --after "Phase 1: Foundations"
    Then the command exits with code 0
    And the roadmap now has phases in order: "Phase 1: Foundations", "Phase 1.5: Bridge", "Phase 2: Growth", "Backlog"
    And the "Phase 1.5: Bridge" entry has "features": []
    And phase "Phase 1: Foundations" is byte-identical to before the command ran
    And phase "Phase 2: Growth" is byte-identical to before the command ran

  Scenario: phase add without --after lands before the Backlog phase
    Given the roadmap has phases in order: "Phase 1: Foundations", "Phase 2: Growth", "Backlog"
    When the operator runs:
      centinela roadmap phase add "Phase 3: Scale"
    Then the command exits with code 0
    And the roadmap now has phases in order: "Phase 1: Foundations", "Phase 2: Growth", "Phase 3: Scale", "Backlog"

  Scenario: phase add without --after and without a Backlog phase lands last
    Given the roadmap has phases in order: "Phase 1: Foundations", "Phase 2: Growth"
    And no phase named "Backlog" exists
    When the operator runs:
      centinela roadmap phase add "Phase 3: Scale"
    Then the command exits with code 0
    And the roadmap now has phases in order: "Phase 1: Foundations", "Phase 2: Growth", "Phase 3: Scale"

  Scenario: phase add --note sets the phase note
    When the operator runs:
      centinela roadmap phase add "Phase 3: Scale" --note "Post-GA hardening work"
    Then the command exits with code 0
    And the "Phase 3: Scale" entry has "note": "Post-GA hardening work"

  Scenario: phase add --after the Backlog phase inserts as a normal schedulable phase, not inside Backlog
    Given the roadmap has phases in order: "Phase 1: Foundations", "Backlog"
    When the operator runs:
      centinela roadmap phase add "Phase 2: Growth" --after "Backlog"
    Then the command exits with code 0
    And the roadmap now has phases in order: "Phase 1: Foundations", "Backlog", "Phase 2: Growth"
    And the "Phase 2: Growth" entry has "features": []
      # allowed as a target position, but the new phase is a normal schedulable
      # phase, not a member of Backlog

  Scenario: phase add on an empty roadmap succeeds as the first phase
    Given .workflow/roadmap.json is exactly {"phases":[]}
    When the operator runs:
      centinela roadmap phase add "Phase 1: Foundations"
    Then the command exits with code 0
    And the roadmap now has phases in order: "Phase 1: Foundations"

  # ---------------------------------------------------------------------------
  # phase add — rejections, byte-identical
  # ---------------------------------------------------------------------------

  Scenario Outline: phase add refuses a duplicate name, reserved name, empty name, or unknown --after anchor
    Given a fixed on-disk roadmap.json (captured as "before")
    And the roadmap has phases in order: "Phase 1: Foundations", "Phase 2: Growth", "Backlog"
    When the operator runs:
      centinela roadmap phase add "<name>" <extra-flags>
    Then the command exits with a non-zero code
    And stderr contains "<error-substring>"
    And .workflow/roadmap.json is byte-identical to "before"

    Examples:
      | name                  | extra-flags                       | error-substring        |
      | Phase 1: Foundations  |                                    | already exists         |
      | Backlog               |                                    | reserved phase name    |
      | Baseline               |                                    | reserved phase name    |
      |                        |                                    | phase name is required |
      | Phase 3: Scale         | --after "Phase 9: Nonexistent"    | unknown phase          |

  # ---------------------------------------------------------------------------
  # phase rename — in place, untouched phases/features, rejections
  # ---------------------------------------------------------------------------

  Scenario: phase rename renames in place, leaving its features and other phases untouched
    Given phase "Phase 1: Foundations" contains features "auth-service", "checkout-ui"
    When the operator runs:
      centinela roadmap phase rename "Phase 1: Foundations" "Phase 1: Core"
    Then the command exits with code 0
    And the phase formerly named "Phase 1: Foundations" is now named "Phase 1: Core"
    And "Phase 1: Core" still contains features "auth-service", "checkout-ui" unchanged
    And phase "Phase 2: Growth" is byte-identical to before the command ran

  Scenario: phase rename to the SAME name is a no-op, byte-identical
    Given phase "Phase 1: Foundations" exists
    When the operator runs:
      centinela roadmap phase rename "Phase 1: Foundations" "Phase 1: Foundations"
    Then the command exits with code 0
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario Outline: phase rename refuses an unknown old name, a collision, an empty new name, or either side reserved
    Given a fixed on-disk roadmap.json (captured as "before")
    And the roadmap has phases in order: "Phase 1: Foundations", "Phase 2: Growth", "Backlog"
    When the operator runs:
      centinela roadmap phase rename "<old>" "<new>"
    Then the command exits with a non-zero code
    And stderr contains "<error-substring>"
    And .workflow/roadmap.json is byte-identical to "before"

    Examples:
      | old                   | new                   | error-substring         |
      | Phase 9: Nonexistent  | Phase 3: Scale        | not found               |
      | Phase 1: Foundations  | Phase 2: Growth       | already exists          |
      | Phase 1: Foundations  |                        | phase name is required |
      | Backlog                | Phase 3: Scale        | reserved phase name    |
      | Phase 1: Foundations  | Baseline               | reserved phase name    |

  # ---------------------------------------------------------------------------
  # phase remove — empty-phase remove, non-empty refusal, --force, reserved
  # ---------------------------------------------------------------------------

  Scenario: phase remove deletes an empty phase
    Given phase "Phase 3: Scale" exists with "features": []
    And the roadmap has phases in order: "Phase 1: Foundations", "Phase 2: Growth", "Phase 3: Scale", "Backlog"
    When the operator runs:
      centinela roadmap phase remove "Phase 3: Scale"
    Then the command exits with code 0
    And the roadmap now has phases in order: "Phase 1: Foundations", "Phase 2: Growth", "Backlog"
    And phase "Phase 1: Foundations" is byte-identical to before the command ran

  Scenario: phase remove of a non-empty phase without --force is refused, naming the feature count
    Given phase "Phase 2: Growth" contains 2 features: "billing-api", "reporting"
    When the operator runs:
      centinela roadmap phase remove "Phase 2: Growth"
    Then the command exits with a non-zero code
    And stderr contains "2 features"
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: phase remove --force removes the phase, its features, and their analysis/quality entries, then validate PASSes
    Given phase "Phase 2: Growth" contains scored (non-draft) features "billing-api", "reporting"
    And .workflow/roadmap-analysis.json has entries for "billing-api" and "reporting"
    And .workflow/roadmap-quality.json has entries for "billing-api" and "reporting"
    And no surviving feature depends on "billing-api" or "reporting"
    When the operator runs:
      centinela roadmap phase remove "Phase 2: Growth" --force
    Then the command exits with code 0
    And the roadmap no longer contains phase "Phase 2: Growth"
    And .workflow/roadmap.json contains no feature named "billing-api" or "reporting"
    And .workflow/roadmap-analysis.json contains no entry for "billing-api" or "reporting"
    And .workflow/roadmap-quality.json contains no entry for "billing-api" or "reporting"
    And running "centinela roadmap validate" exits with code 0 (PASS)

  Scenario: phase remove --force is REFUSED byte-identical when a surviving feature depends on a removed one
    Given phase "Phase 2: Growth" contains feature "billing-api"
    And "checkout-ui" (in phase "Phase 1: Foundations") declares "dependsOn": ["billing-api"]
    When the operator runs:
      centinela roadmap phase remove "Phase 2: Growth" --force
    Then the command exits with a non-zero code
    And stderr contains "depends on"
    And .workflow/roadmap.json is byte-identical to before the command ran
    And .workflow/roadmap-analysis.json is byte-identical to before the command ran
    And .workflow/roadmap-quality.json is byte-identical to before the command ran

  Scenario: phase remove of an unknown phase errors "not found"
    When the operator runs:
      centinela roadmap phase remove "Phase 9: Nonexistent"
    Then the command exits with a non-zero code
    And stderr contains "not found"
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario Outline: phase remove refuses the reserved Backlog/Baseline phase, with or without --force
    Given a fixed on-disk roadmap.json (captured as "before")
    When the operator runs:
      centinela roadmap phase remove "<name>" <extra-flags>
    Then the command exits with a non-zero code
    And stderr contains "reserved phase name"
    And .workflow/roadmap.json is byte-identical to "before"

    Examples:
      | name     | extra-flags |
      | Backlog  |              |
      | Backlog  | --force      |
      | Baseline | --force      |

  Scenario: phase remove of the only phase leaves an empty roadmap
    Given the roadmap has exactly one phase, "Phase 1: Foundations", with "features": []
    When the operator runs:
      centinela roadmap phase remove "Phase 1: Foundations"
    Then the command exits with code 0
    And .workflow/roadmap.json is exactly {"phases":[]}

  # ---------------------------------------------------------------------------
  # dirty-map reindex invariant — the core new complexity
  # ---------------------------------------------------------------------------

  Scenario: inserting an earlier phase while mutating a later feature reindexes the dirty map so both renders are correct
    Given the roadmap has phases in order: "Phase 1: Foundations", "Phase 2: Growth", "Phase 3: Scale"
    And phase "Phase 3: Scale" contains feature "reporting"
    When a single run inserts "Phase 0: Bootstrap" before "Phase 1: Foundations"
      AND edits feature "reporting" in "Phase 3: Scale" (now shifted to index 3) in the SAME process
    Then the command exits with code 0
    And the roadmap now has phases in order: "Phase 0: Bootstrap", "Phase 1: Foundations", "Phase 2: Growth", "Phase 3: Scale"
    And the "reporting" entry reflects the edit correctly, not corrupted or applied to the wrong phase
    And phase "Phase 2: Growth" is byte-identical to before the run

  Scenario: removing a middle phase while mutating a later feature reindexes the dirty map so both renders are correct
    Given the roadmap has phases in order: "Phase 1: Foundations", "Phase 2: Growth", "Phase 3: Scale", "Phase 4: Polish"
    And phase "Phase 2: Growth" is empty
    And phase "Phase 4: Polish" contains feature "docs-site"
    When a single run removes "Phase 2: Growth"
      AND edits feature "docs-site" in "Phase 4: Polish" (now shifted to index 2) in the SAME process
    Then the command exits with code 0
    And the roadmap now has phases in order: "Phase 1: Foundations", "Phase 3: Scale", "Phase 4: Polish"
    And the "docs-site" entry reflects the edit correctly, not corrupted or applied to the wrong phase
    And phase "Phase 3: Scale" is byte-identical to before the run

  # ---------------------------------------------------------------------------
  # cross-cutting — empty/missing/malformed roadmap.json
  # ---------------------------------------------------------------------------

  Scenario Outline: phase rename/remove against an empty roadmap errors "not found"
    Given .workflow/roadmap.json is exactly {"phases":[]}
    When the operator runs:
      centinela roadmap phase <command> "Phase 1: Foundations" <extra-flags>
    Then the command exits with a non-zero code
    And stderr contains "not found"
    And .workflow/roadmap.json remains exactly {"phases":[]}

    Examples:
      | command | extra-flags       |
      | rename  | "Phase 1: Core"   |
      | remove  |                   |

  Scenario: phase add/rename/remove against a missing roadmap.json surfaces an error and leaves the file absent
    Given .workflow/roadmap.json does not exist
    When the operator runs:
      centinela roadmap phase add "Phase 1: Foundations"
    Then the command exits with a non-zero code
    And stderr contains an error message
    And .workflow/roadmap.json is still absent

  Scenario: phase add/rename/remove against a malformed roadmap.json surfaces an error and leaves the file untouched
    Given .workflow/roadmap.json contains invalid JSON: "{ not valid json"
    When the operator runs:
      centinela roadmap phase add "Phase 1: Foundations"
    Then the command exits with a non-zero code
    And stderr contains an error message
    And .workflow/roadmap.json is byte-identical to before the command ran

  Scenario: every phase add/rename/remove performs exactly one atomic write — a rejected op writes nothing
    Given a fixed on-disk roadmap.json (captured as "before") with its file mtime recorded
    When the operator runs a phase add, rename, or remove that gets refused
      (duplicate/reserved/empty name, unknown anchor/phase, or a failed --force revalidation)
    Then the command exits with a non-zero code
    And .workflow/roadmap.json's mtime and contents are unchanged from "before"
      # Validation happens in memory against the decoded raw document before
      # the single writeRawRoadmap call; a refusal never reaches that call.
