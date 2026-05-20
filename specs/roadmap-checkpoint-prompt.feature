Feature: Roadmap definition checkpoint prompt
  As a project initiator who just finished defining the roadmap
  I want centinela hook setup to emit one explicit handoff directive
  So I can choose to keep iterating on the roadmap or start the first Phase 0 feature

  Scenario: Happy path emits the checkpoint directive when no marker exists
    Given PROJECT.md, ROADMAP.md, .workflow/roadmap.json are present
    And .workflow/roadmap-analysis.md and .workflow/roadmap-analysis.json are present
    And .workflow/roadmap-quality.md and .workflow/roadmap-quality.json are present
    And docs/architecture/production-readiness-prompt.md is present
    And no .workflow/roadmap-checkpoint.json marker exists
    And bootstrap is incomplete with a first incomplete Phase 0 feature "phase-0-feature-a"
    And no .workflow/phase-0-feature-a.json workflow file exists
    When the user submits a prompt and centinela hook setup runs
    Then the output should contain "CENTINELA DIRECTIVE: roadmap checkpoint"
    And the rendered panel should name the feature "phase-0-feature-a"
    And the command should exit zero

  Scenario: Suppressed when the marker is fresh against all roadmap artifacts
    Given the full set of roadmap-defining artifacts is present
    And .workflow/roadmap-checkpoint.json exists with choice "iterate" and "at" equal to or after the latest mtime of every roadmap-defining artifact
    And bootstrap is incomplete with a first incomplete Phase 0 feature "phase-0-feature-a"
    And no .workflow/phase-0-feature-a.json workflow file exists
    When the user submits a prompt and centinela hook setup runs
    Then no "CENTINELA DIRECTIVE: roadmap checkpoint" line should be emitted
    And no roadmap checkpoint panel should be rendered

  Scenario: Stale marker re-fires when ROADMAP.md was modified after the marker
    Given the full set of roadmap-defining artifacts is present
    And .workflow/roadmap-checkpoint.json exists with "at" set to some RFC 3339 timestamp
    And ROADMAP.md has an mtime strictly later than the marker "at"
    And bootstrap is incomplete with a first incomplete Phase 0 feature "phase-0-feature-a"
    And no .workflow/phase-0-feature-a.json workflow file exists
    When the user submits a prompt and centinela hook setup runs
    Then the output should contain "CENTINELA DIRECTIVE: roadmap checkpoint"
    And the rendered panel should name the feature "phase-0-feature-a"

  Scenario: Stale marker re-fires when any roadmap supporting artifact was modified after the marker
    Given the full set of roadmap-defining artifacts is present
    And .workflow/roadmap-checkpoint.json exists with "at" set to some RFC 3339 timestamp
    And .workflow/roadmap-analysis.json has an mtime strictly later than the marker "at"
    And bootstrap is incomplete with a first incomplete Phase 0 feature "phase-0-feature-a"
    When the user submits a prompt and centinela hook setup runs
    Then the output should contain "CENTINELA DIRECTIVE: roadmap checkpoint"
    And the rendered panel should name the feature "phase-0-feature-a"

  Scenario: Suppressed when bootstrap is already complete
    Given the full set of roadmap-defining artifacts is present
    And every Phase 0 bootstrap feature has status "done" in .workflow/roadmap.json
    When the user submits a prompt and centinela hook setup runs
    Then no "CENTINELA DIRECTIVE: roadmap checkpoint" line should be emitted
    And no roadmap checkpoint panel should be rendered

  Scenario: Suppressed when no Phase 0 bootstrap features exist in the roadmap
    Given the full set of roadmap-defining artifacts is present
    And .workflow/roadmap.json defines no Phase 0 bootstrap features
    When the user submits a prompt and centinela hook setup runs
    Then no "CENTINELA DIRECTIVE: roadmap checkpoint" line should be emitted

  Scenario: Suppressed when the workflow file for the first Phase 0 feature already exists
    Given the full set of roadmap-defining artifacts is present
    And the first incomplete Phase 0 feature is "phase-0-feature-a"
    And .workflow/phase-0-feature-a.json exists
    When the user submits a prompt and centinela hook setup runs
    Then no "CENTINELA DIRECTIVE: roadmap checkpoint" line should be emitted
    And no roadmap checkpoint panel should be rendered

  Scenario: Order of precedence — missing roadmap-defining artifact lets the existing setup directives fire instead
    Given PROJECT.md is present but ROADMAP.md is missing
    When the user submits a prompt and centinela hook setup runs
    Then the output should contain "CENTINELA DIRECTIVE: roadmap required"
    And no "CENTINELA DIRECTIVE: roadmap checkpoint" line should be emitted

  Scenario: Order of precedence — invalid roadmap.json yields the roadmap-json directive, not the checkpoint
    Given PROJECT.md and ROADMAP.md are present
    But .workflow/roadmap.json is malformed
    When the user submits a prompt and centinela hook setup runs
    Then the output should contain "CENTINELA DIRECTIVE: roadmap json"
    And no "CENTINELA DIRECTIVE: roadmap checkpoint" line should be emitted

  Scenario: Multiple Phase 0 features, only the first is done — picks the second as target
    Given the full set of roadmap-defining artifacts is present
    And the roadmap defines Phase 0 features "phase-0-alpha" and "phase-0-beta" in that order
    And "phase-0-alpha" has status "done" and "phase-0-beta" is incomplete
    And no .workflow/roadmap-checkpoint.json marker exists
    And no .workflow/phase-0-beta.json workflow file exists
    When the user submits a prompt and centinela hook setup runs
    Then the rendered panel should name the feature "phase-0-beta"

  Scenario: Malformed marker JSON is treated as missing and re-emits without crashing
    Given the full set of roadmap-defining artifacts is present
    And .workflow/roadmap-checkpoint.json exists but is not valid JSON
    And bootstrap is incomplete with a first incomplete Phase 0 feature "phase-0-feature-a"
    When the user submits a prompt and centinela hook setup runs
    Then the command should not crash
    And the output should contain "CENTINELA DIRECTIVE: roadmap checkpoint"

  Scenario: Marker "at" field unparseable as RFC 3339 is treated as stale and re-emits
    Given the full set of roadmap-defining artifacts is present
    And .workflow/roadmap-checkpoint.json exists with "at" set to a non-RFC-3339 string
    And bootstrap is incomplete with a first incomplete Phase 0 feature "phase-0-feature-a"
    When the user submits a prompt and centinela hook setup runs
    Then the output should contain "CENTINELA DIRECTIVE: roadmap checkpoint"
