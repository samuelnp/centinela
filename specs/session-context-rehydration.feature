Feature: Session context rehydration
  As a developer using Claude Code with Centinela
  I want a clean active-workflows panel and a SessionStart roadmap bootstrap
  So that after /clear (or startup/compact/resume) the model rediscovers project state without noise

  # ---------------------------------------------------------------------------
  # Half A — Active-workflows panel (UserPromptSubmit, centinela hook context)
  # ---------------------------------------------------------------------------

  Scenario: An evidence JSON in .workflow/ is not rendered as an active workflow
    Given .workflow/ contains a real workflow-state file "alpha.json" whose feature is "alpha" and currentStep is "code"
    And .workflow/ contains an evidence JSON "alpha-qa-senior.json" that has no currentStep field
    When the user submits a prompt and centinela hook context runs
    Then the active-workflows panel should list the feature "alpha"
    And the active-workflows panel should not list any entry derived from "alpha-qa-senior.json"
    And the command should exit zero

  Scenario: A done workflow is excluded while a genuine non-done workflow is shown
    Given .workflow/ contains a real workflow-state file "beta.json" whose feature is "beta" and currentStep is "done"
    And .workflow/ contains a real workflow-state file "gamma.json" whose feature is "gamma" and currentStep is "tests"
    When the user submits a prompt and centinela hook context runs
    Then the active-workflows panel should list the feature "gamma"
    And the active-workflows panel should not list the feature "beta"

  Scenario: Ad-hoc roadmap JSON files are not treated as active workflows
    Given .workflow/ contains "roadmap.json" and "roadmap-quality.json" whose base names do not match any feature field
    And .workflow/ contains a real workflow-state file "delta.json" whose feature is "delta" and currentStep is "plan"
    When the user submits a prompt and centinela hook context runs
    Then the active-workflows panel should list the feature "delta"
    And the active-workflows panel should not list an entry named "roadmap" or "roadmap-quality"

  Scenario: Duplicate feature entries are deduplicated to a single panel row
    Given .workflow/ contains a real workflow-state file "epsilon.json" whose feature is "epsilon" and currentStep is "code"
    And .workflow/ contains multiple evidence JSONs for feature "epsilon" such as "epsilon-big-thinker.json" and "epsilon-qa-senior.json"
    When the user submits a prompt and centinela hook context runs
    Then the feature "epsilon" should appear exactly once in the active-workflows panel

  Scenario: More active workflows than the cap show only the most-recently-touched plus a "+N more" hint
    Given .workflow/ contains 7 distinct real non-done workflow-state files for 7 distinct features
    And the workflow files have distinct modification times
    When the user submits a prompt and centinela hook context runs
    Then the active-workflows panel should list at most 5 features
    And the listed features should be the 5 most-recently-touched in modification-time descending order
    And the panel should include a "+2 more" indicator

  Scenario: At-or-below the cap shows no "+N more" hint
    Given .workflow/ contains 3 distinct real non-done workflow-state files for 3 distinct features
    When the user submits a prompt and centinela hook context runs
    Then the active-workflows panel should list 3 features
    And the panel should not include any "+N more" indicator

  # ---------------------------------------------------------------------------
  # Half B — SessionStart rehydration hook (centinela hook session)
  # ---------------------------------------------------------------------------

  Scenario Outline: SessionStart injects the rehydration payload on each supported source
    Given PROJECT.md and a valid roadmap with declared phases are present
    And the first incomplete feature across all phases is "next-feature"
    When a SessionStart event from source "<source>" runs centinela hook session
    Then the output should contain "CENTINELA DIRECTIVE: session rehydration"
    And the output should contain the full roadmap with per-feature status
    And the output should name the next feature to plan as "next-feature"
    And the output should list the pointer path "PROJECT.md"
    And the output should list the pointer path "docs/features/next-feature.md"
    And the output should not inline the contents of PROJECT.md or docs/features/next-feature.md
    And the command should exit zero

    Examples:
      | source  |
      | startup |
      | clear   |
      | compact |
      | resume  |

  Scenario: Next feature is the first incomplete across all phases, not just Phase 0
    Given PROJECT.md and a valid roadmap are present
    And every Phase 0 feature has status "done"
    And the first incomplete Phase 1 feature is "phase-1-first"
    When a SessionStart event runs centinela hook session
    Then the output should name the next feature to plan as "phase-1-first"
    And the output should list the pointer path "docs/features/phase-1-first.md"

  Scenario: Every roadmap feature done yields a graceful roadmap-complete state with no next feature
    Given PROJECT.md and a valid roadmap are present
    And every feature in every phase has status "done"
    When a SessionStart event runs centinela hook session
    Then the output should indicate the roadmap is complete
    And no next-feature name should be emitted
    And no "docs/features/<next>.md" pointer should be emitted
    And the command should not crash and should exit zero

  Scenario: Missing roadmap is handled gracefully without crashing
    Given neither ROADMAP.md nor .workflow/roadmap.json is present
    When a SessionStart event runs centinela hook session
    Then the command should not crash and should exit zero
    And no "CENTINELA DIRECTIVE: session rehydration" payload should be emitted

  Scenario: Invalid roadmap json is handled gracefully without crashing
    Given .workflow/roadmap.json is present but malformed
    When a SessionStart event runs centinela hook session
    Then the command should not crash and should exit zero
    And no "CENTINELA DIRECTIVE: session rehydration" payload should be emitted
