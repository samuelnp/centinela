Feature: Roadmap parallel readiness — dependency graph, frontier, and start guard
  As an operator driving multiple Centinela/Claude instances in parallel worktrees
  I want the roadmap to carry first-class dependency declarations
  So I can see which features are safe to start concurrently and be blocked from grabbing unready work

  # ─── SCHEMA & BACKWARD COMPATIBILITY ────────────────────────────────────────

  Scenario: A roadmap.json with dependsOn fields loads successfully
    Given a valid roadmap.json where feature "feature-b" declares dependsOn ["feature-a"]
    When the roadmap is loaded
    Then the load succeeds without error
    And feature "feature-b" has DependsOn containing "feature-a"

  Scenario: A roadmap.json without any dependsOn field loads exactly as before
    Given a valid roadmap.json where no feature declares a dependsOn field
    When the roadmap is loaded
    Then the load succeeds without error
    And every planned feature is classified as "ready"

  Scenario: An empty dependsOn array is treated the same as an absent dependsOn field
    Given a valid roadmap.json where feature "feature-a" declares dependsOn []
    When the roadmap is loaded
    Then the load succeeds without error
    And feature "feature-a" is classified as "ready"

  # ─── VALIDATION — NEGATIVE PATHS ────────────────────────────────────────────

  Scenario: A dependency on an unknown feature slug is rejected at load time
    Given a roadmap.json where feature "feature-b" declares dependsOn ["feature-does-not-exist"]
    When the roadmap is loaded
    Then the load fails with an error
    And the error message names "feature-does-not-exist" as the unknown dependency

  Scenario: A dependency cycle between two features is rejected at load time
    Given a roadmap.json where feature "feature-a" declares dependsOn ["feature-b"]
    And feature "feature-b" declares dependsOn ["feature-a"]
    When the roadmap is loaded
    Then the load fails with an error
    And the error message indicates a dependency cycle was detected

  Scenario: A self-dependency is rejected at load time as a cycle
    Given a roadmap.json where feature "feature-a" declares dependsOn ["feature-a"]
    When the roadmap is loaded
    Then the load fails with an error
    And the error message indicates a dependency cycle was detected

  Scenario: A longer cycle (A→B→C→A) is rejected at load time
    Given a roadmap.json where feature "feature-a" declares dependsOn ["feature-c"]
    And feature "feature-b" declares dependsOn ["feature-a"]
    And feature "feature-c" declares dependsOn ["feature-b"]
    When the roadmap is loaded
    Then the load fails with an error
    And the error message indicates a dependency cycle was detected

  # ─── READINESS DERIVATION ────────────────────────────────────────────────────

  Scenario: A done feature is classified as done
    Given a roadmap.json where feature "feature-a" has workflow status "done"
    When readiness is derived
    Then feature "feature-a" has readiness state "done"
    And feature "feature-a" has no BlockedBy entries

  Scenario: An in-progress feature is classified as in-progress
    Given a roadmap.json where feature "feature-a" has workflow status "in-progress"
    When readiness is derived
    Then feature "feature-a" has readiness state "in-progress"
    And feature "feature-a" has no BlockedBy entries

  Scenario: A planned feature with all dependencies done is classified as ready
    Given a roadmap.json where feature "feature-a" has workflow status "done"
    And feature "feature-b" declares dependsOn ["feature-a"] and has workflow status "planned"
    When readiness is derived
    Then feature "feature-b" has readiness state "ready"
    And feature "feature-b" has no BlockedBy entries

  Scenario: A planned feature with an unmet dependency is classified as blocked
    Given a roadmap.json where feature "feature-a" has workflow status "planned"
    And feature "feature-b" declares dependsOn ["feature-a"] and has workflow status "planned"
    When readiness is derived
    Then feature "feature-b" has readiness state "blocked"
    And feature "feature-b" has BlockedBy containing "feature-a"

  Scenario: A dependency that is in-progress keeps the dependent blocked
    Given a roadmap.json where feature "feature-a" has workflow status "in-progress"
    And feature "feature-b" declares dependsOn ["feature-a"] and has workflow status "planned"
    When readiness is derived
    Then feature "feature-b" has readiness state "blocked"
    And feature "feature-b" has BlockedBy containing "feature-a"

  Scenario: Multiple unmet dependencies are all listed in BlockedBy
    Given a roadmap.json where feature "feature-a" has workflow status "planned"
    And feature "feature-b" has workflow status "planned"
    And feature "feature-c" declares dependsOn ["feature-a", "feature-b"] and has workflow status "planned"
    When readiness is derived
    Then feature "feature-c" has readiness state "blocked"
    And feature "feature-c" has BlockedBy containing "feature-a"
    And feature "feature-c" has BlockedBy containing "feature-b"

  Scenario: Diamond dependency — D is ready only when both B and C are done
    Given a roadmap.json with features "feature-a", "feature-b", "feature-c", "feature-d"
    And feature "feature-b" declares dependsOn ["feature-a"]
    And feature "feature-c" declares dependsOn ["feature-a"]
    And feature "feature-d" declares dependsOn ["feature-b", "feature-c"]
    And features "feature-a", "feature-b", "feature-c" have workflow status "done"
    And feature "feature-d" has workflow status "planned"
    When readiness is derived
    Then feature "feature-d" has readiness state "ready"

  Scenario: Diamond dependency — D remains blocked when B is done but C is not
    Given a roadmap.json with features "feature-a", "feature-b", "feature-c", "feature-d"
    And feature "feature-b" declares dependsOn ["feature-a"]
    And feature "feature-c" declares dependsOn ["feature-a"]
    And feature "feature-d" declares dependsOn ["feature-b", "feature-c"]
    And features "feature-a", "feature-b" have workflow status "done"
    And feature "feature-c" has workflow status "planned"
    And feature "feature-d" has workflow status "planned"
    When readiness is derived
    Then feature "feature-d" has readiness state "blocked"
    And feature "feature-d" has BlockedBy containing "feature-c"

  # ─── centinela roadmap ready COMMAND ────────────────────────────────────────

  Scenario: roadmap ready prints each ready feature on its own line
    Given a roadmap.json where feature "feature-a" has workflow status "done"
    And feature "feature-b" declares dependsOn ["feature-a"] and has workflow status "planned"
    And feature "feature-c" has no dependsOn and has workflow status "planned"
    When the user runs "centinela roadmap ready"
    Then the command exits with code 0
    And the output contains "feature-b" on its own line
    And the output contains "feature-c" on its own line

  Scenario: roadmap ready prints a clear empty-state line when no features are ready
    Given a roadmap.json where every planned feature has an unmet dependency
    When the user runs "centinela roadmap ready"
    Then the command exits with code 0
    And the output contains a non-empty empty-state message
    And the output does not list any feature names

  Scenario: roadmap ready exits 0 when all features are done
    Given a roadmap.json where every feature has workflow status "done"
    When the user runs "centinela roadmap ready"
    Then the command exits with code 0
    And the output contains a non-empty empty-state message

  # ─── START GUARD ─────────────────────────────────────────────────────────────

  Scenario: centinela start is refused when a dependency is not done
    Given a roadmap.json where feature "feature-a" has workflow status "planned"
    And feature "feature-b" declares dependsOn ["feature-a"] and has workflow status "planned"
    When the user runs "centinela start feature-b"
    Then the command exits with a non-zero code
    And the error output names "feature-a" as an unmet dependency
    And the error output mentions "feature-b"

  Scenario: centinela start is refused when a dependency is only in-progress
    Given a roadmap.json where feature "feature-a" has workflow status "in-progress"
    And feature "feature-b" declares dependsOn ["feature-a"] and has workflow status "planned"
    When the user runs "centinela start feature-b"
    Then the command exits with a non-zero code
    And the error output names "feature-a" as an unmet dependency

  Scenario: centinela start proceeds when all dependencies are done
    Given a roadmap.json where feature "feature-a" has workflow status "done"
    And feature "feature-b" declares dependsOn ["feature-a"] and has workflow status "planned"
    When the user runs "centinela start feature-b"
    Then the command does not emit a dependency error

  Scenario: centinela start proceeds when the feature has no dependencies
    Given a roadmap.json where feature "feature-a" has no dependsOn and has workflow status "planned"
    When the user runs "centinela start feature-a"
    Then the command does not emit a dependency error

  # ─── RENDER MARKERS ──────────────────────────────────────────────────────────

  Scenario: centinela roadmap renders the ready marker on ready features
    Given a roadmap.json where feature "feature-a" has workflow status "done"
    And feature "feature-b" declares dependsOn ["feature-a"] and has workflow status "planned"
    When the user runs "centinela roadmap"
    Then the output contains "🔓" adjacent to "feature-b"

  Scenario: centinela roadmap renders the blocked marker and dep names on blocked features
    Given a roadmap.json where feature "feature-a" has workflow status "planned"
    And feature "feature-b" declares dependsOn ["feature-a"] and has workflow status "planned"
    When the user runs "centinela roadmap"
    Then the output contains "🔒" adjacent to "feature-b"
    And the output names "feature-a" as the blocking dependency for "feature-b"

  Scenario: centinela roadmap does not render ready or blocked markers on done features
    Given a roadmap.json where feature "feature-a" has workflow status "done"
    When the user runs "centinela roadmap"
    Then the output does not show a 🔓 or 🔒 marker for "feature-a"

  # ─── PLURAL REHYDRATION ───────────────────────────────────────────────────────

  Scenario: SessionStart rehydration lists all ready features when multiple are ready
    Given a roadmap.json where features "feature-b" and "feature-c" are both ready
    When a new session starts and the session hook runs
    Then the rehydration output lists "feature-b"
    And the rehydration output lists "feature-c"

  Scenario: SessionStart rehydration explains the block when frontier is empty but work remains
    Given a roadmap.json where all planned features have at least one unmet dependency
    And at least one feature has workflow status "planned" or "in-progress"
    When a new session starts and the session hook runs
    Then the rehydration output does not show the roadmap-complete message
    And the rehydration output indicates that no features are currently ready to start
    And the rehydration output references the blocking reason

  Scenario: SessionStart rehydration shows the roadmap-complete message when all features are done
    Given a roadmap.json where every feature has workflow status "done"
    When a new session starts and the session hook runs
    Then the rehydration output contains the roadmap-complete message
    And the rehydration output does not list any ready features
