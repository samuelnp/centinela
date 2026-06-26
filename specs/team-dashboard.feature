Feature: centinela dashboard — multi-feature team status board
  As a developer or lead running several Centinela features at once
  I want `centinela dashboard` to print a single read-only board with three panels
  So that in-flight feature state, roadmap burn-down, and gate health are visible
  without polling each worktree by hand

  # centinela dashboard reads active workflow JSONs, roadmap.json, and the
  # telemetry event log read-only and computes three panels via the pure
  # internal/teamdashboard aggregator. The command touches no files. Empty or
  # missing sources each yield an honest empty-state panel, never an error.
  # --json emits the full stable Dashboard struct. Scenario titles map 1:1 to
  # Go acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with a valid centinela.toml
    And the workflow state directory is ".workflow"
    And the telemetry log is at ".workflow/telemetry/events.jsonl"

  # ---------------------------------------------------------------------------
  # Happy path — full board from on-disk state
  # ---------------------------------------------------------------------------

  Scenario: Dashboard prints three panels from current on-disk state and exits 0
    Given one active workflow "alpha" is at step "code" (step 2 of 5) started 3 days ago
    And the roadmap has two schedulable phases: "Q1" with 2 features (1 done) and "Q2" with 1 feature (0 done)
    And the telemetry log contains 2 gate-failure events for gate "coverage" and 1 for "import-graph"
    When the operator runs:
      centinela dashboard
    Then the command exits with code 0
    And the output contains an "In-flight features" panel
    And the output contains a "Roadmap burn-down" panel
    And the output contains a "Gate health" panel
    And no files are written or modified

  Scenario: Dashboard is read-only — no files are created or written
    Given one active workflow "beta" is at step "tests" (step 3 of 5)
    And the roadmap is present with one schedulable phase
    And the telemetry log contains at least one gate-failure event
    When the operator runs:
      centinela dashboard
    Then the command exits with code 0
    And the mtime of every file under ".workflow" is unchanged
    And no new files appear anywhere under the project directory

  # ---------------------------------------------------------------------------
  # In-flight panel — row content
  # ---------------------------------------------------------------------------

  Scenario: In-flight row shows feature name, step, X/5 progress, age, profile, archetype, worktree, and owner
    Given one active workflow "alpha" with:
      | field              | value                      |
      | step               | code                       |
      | stepIndex          | 2                          |
      | stepTotal          | 5                          |
      | startedAt          | 3 days ago                 |
      | enforcementProfile | strict                     |
      | archetype          | hexagonal                  |
      | worktreePath       | .worktrees/alpha           |
      | owner              | Alice Smith                |
    When the operator runs:
      centinela dashboard
    Then the in-flight panel row for "alpha" contains "code"
    And the in-flight panel row for "alpha" contains "2/5"
    And the in-flight panel row for "alpha" contains "3d"
    And the in-flight panel row for "alpha" contains "strict"
    And the in-flight panel row for "alpha" contains "hexagonal"
    And the in-flight panel row for "alpha" contains ".worktrees/alpha"
    And the in-flight panel row for "alpha" contains "Alice Smith"

  Scenario: Step index reflects done-count position in the ordered step list
    Given one active workflow "alpha" is at step "validate" with stepIndex 4 and stepTotal 5
    When the operator runs:
      centinela dashboard
    Then the in-flight panel row for "alpha" contains "4/5"

  Scenario: Age is computed as floor days from StartedAt to now
    Given one active workflow "alpha" started exactly 7 days and 23 hours ago
    When the operator runs:
      centinela dashboard
    Then the in-flight panel row for "alpha" shows age "7d"

  Scenario: Zero StartedAt shows age 0d and does not crash
    Given one active workflow "alpha" has a zero StartedAt value
    When the operator runs:
      centinela dashboard
    Then the in-flight panel row for "alpha" shows age "0d"
    And the command exits with code 0

  Scenario: Blank profile renders as "default" in the output
    Given one active workflow "alpha" has an empty EnforcementProfile field
    When the operator runs:
      centinela dashboard
    Then the in-flight panel row for "alpha" displays "default" for the profile

  Scenario: Blank archetype renders as "canonical" in the output
    Given one active workflow "alpha" has an empty Archetype field
    When the operator runs:
      centinela dashboard
    Then the in-flight panel row for "alpha" displays "canonical" for the archetype

  Scenario: Blank worktree path renders as a dash in the output
    Given one active workflow "alpha" has an empty WorktreePath field
    When the operator runs:
      centinela dashboard
    Then the in-flight panel row for "alpha" displays "—" for the worktree

  Scenario: Multiple active workflows each appear as a row in file-mtime descending order
    Given two active workflows "alpha" (modified 1 hour ago) and "beta" (modified 2 hours ago)
    When the operator runs:
      centinela dashboard
    Then the in-flight panel lists "alpha" before "beta"

  # ---------------------------------------------------------------------------
  # Owner column — git-derived, best-effort fallback
  # ---------------------------------------------------------------------------

  Scenario: Owner column shows git-derived committer name when commits exist on the branch
    Given one active workflow "gamma" on branch "gamma"
    And the latest commit on branch "gamma" was authored by "Bob Jones"
    When the operator runs:
      centinela dashboard
    Then the in-flight panel row for "gamma" contains "Bob Jones"

  Scenario: Feature with no commits on branch shows "unknown" owner, not an error
    Given one active workflow "gamma" on branch "gamma"
    And branch "gamma" has no commits
    When the operator runs:
      centinela dashboard
    Then the in-flight panel row for "gamma" contains "unknown"
    And the command exits with code 0

  Scenario: Git unavailable falls back to "unknown" owner and does not propagate an error
    Given one active workflow "gamma" on branch "gamma"
    And the git executable is unavailable
    When the operator runs:
      centinela dashboard
    Then the in-flight panel row for "gamma" contains "unknown"
    And the command exits with code 0

  Scenario: Owner derivation error does not abort the rest of the dashboard
    Given two active workflows "alpha" and "beta"
    And the git owner call for "alpha" returns an error
    And the git owner call for "beta" returns "Carol"
    When the operator runs:
      centinela dashboard
    Then the in-flight panel row for "alpha" contains "unknown"
    And the in-flight panel row for "beta" contains "Carol"
    And the command exits with code 0

  # ---------------------------------------------------------------------------
  # Roadmap burn-down panel
  # ---------------------------------------------------------------------------

  Scenario: Roadmap panel shows per-phase done/total counts and an overall N/M done line
    Given a roadmap with:
      | phase | totalFeatures | doneFeatures |
      | Q1    | 3             | 2            |
      | Q2    | 4             | 1            |
    When the operator runs:
      centinela dashboard
    Then the roadmap panel shows "Q1" with "2/3"
    And the roadmap panel shows "Q2" with "1/4"
    And the roadmap panel contains an overall "3/7 done" line

  Scenario: Roadmap burn-down matches roadmap.Summary() schedulable phases only
    Given a roadmap containing phases "Backlog", "Baseline", "Q1", and "Q2"
    And "Backlog" and "Baseline" are non-schedulable phases
    And "Q1" has 2 features total with 1 done and "Q2" has 3 features total with 2 done
    When the operator runs:
      centinela dashboard
    Then the roadmap panel does not mention "Backlog"
    And the roadmap panel does not mention "Baseline"
    And the roadmap panel contains an overall "3/5 done" line

  Scenario: Empty roadmap with zero schedulable features shows 0/0 done and exits 0
    Given a roadmap that exists but contains no schedulable features
    When the operator runs:
      centinela dashboard
    Then the roadmap panel contains "0/0 done"
    And the command exits with code 0

  # ---------------------------------------------------------------------------
  # Gate health panel
  # ---------------------------------------------------------------------------

  Scenario: Gate health panel ranks gate-failure events by count descending
    Given a telemetry log containing:
      | type         | gate         |
      | gate-failure | coverage     |
      | gate-failure | coverage     |
      | gate-failure | coverage     |
      | gate-failure | import-graph |
      | gate-failure | import-graph |
    When the operator runs:
      centinela dashboard
    Then the gate health panel lists "coverage" with count 3
    And the gate health panel lists "import-graph" with count 2
    And "coverage" appears before "import-graph" in the gate health panel

  Scenario: Gate-failure event with empty Gate field buckets under "<none>"
    Given a telemetry log containing a gate-failure event with an empty Gate field
    When the operator runs:
      centinela dashboard
    Then the gate health panel contains an entry rendered as "<none>"
    And the command exits with code 0

  Scenario: Gate health ranking matches insights.Gates output for the same events
    Given a telemetry log with a known mix of gate-failure events
    And the insights command produces a Gates section for that same log
    When the operator runs:
      centinela dashboard
    Then the gate health panel lists gates in the same order as the insights Gates section
    And the fail counts match for every gate

  Scenario: Non-gate-failure event types are excluded from gate health counts
    Given a telemetry log containing:
      | type          | gate      |
      | block         | coverage  |
      | step-advanced | coverage  |
      | gate-failure  | coverage  |
    When the operator runs:
      centinela dashboard
    Then the gate health panel lists "coverage" with count 1

  # ---------------------------------------------------------------------------
  # Empty / degraded states — honest empty panels, never errors
  # ---------------------------------------------------------------------------

  Scenario: No active workflows shows honest empty in-flight panel and exits 0
    Given no active workflow files exist under ".workflow"
    And the roadmap and telemetry log are present
    When the operator runs:
      centinela dashboard
    Then the command exits with code 0
    And the in-flight panel contains "no active features"
    And the roadmap panel and gate health panel are still rendered

  Scenario: No telemetry shows honest empty gate-health panel and exits 0
    Given one active workflow exists
    And the roadmap is present
    And no telemetry log file exists
    When the operator runs:
      centinela dashboard
    Then the command exits with code 0
    And the gate health panel contains "no gate failures recorded"
    And the in-flight panel and roadmap panel are still rendered

  Scenario: Empty telemetry log shows honest empty gate-health panel and exits 0
    Given the telemetry log exists but contains no events
    When the operator runs:
      centinela dashboard
    Then the command exits with code 0
    And the gate health panel contains "no gate failures recorded"

  Scenario: Telemetry present but no gate-failure events shows honest empty gate-health panel
    Given a telemetry log containing only block and step-advanced events
    When the operator runs:
      centinela dashboard
    Then the gate health panel contains "no gate failures recorded"
    And the command exits with code 0

  Scenario: Absent or unreadable roadmap shows honest empty roadmap panel and exits 0
    Given one active workflow exists
    And no roadmap.json file exists
    When the operator runs:
      centinela dashboard
    Then the command exits with code 0
    And the roadmap panel contains "no roadmap"
    And the in-flight panel and gate health panel are still rendered

  Scenario: All three sources missing yields three honest empty panels and exits 0
    Given no active workflow files exist
    And no roadmap.json file exists
    And no telemetry log file exists
    When the operator runs:
      centinela dashboard
    Then the command exits with code 0
    And the in-flight panel contains "no active features"
    And the roadmap panel contains "no roadmap"
    And the gate health panel contains "no gate failures recorded"

  # ---------------------------------------------------------------------------
  # --json output
  # ---------------------------------------------------------------------------

  Scenario: --json emits full Dashboard struct as indented JSON and exits 0
    Given one active workflow "alpha" at step "code" with owner "Alice"
    And the roadmap has one schedulable phase "Q1" with 2 features (1 done)
    And the telemetry log contains 2 gate-failure events for gate "coverage"
    When the operator runs:
      centinela dashboard --json
    Then the command exits with code 0
    And the output is valid JSON
    And the JSON contains top-level fields "Features", "Roadmap", "Gates"
    And the JSON "Features" array contains one entry with fields "Feature", "Step", "StepIndex", "StepTotal", "AgeDays", "Profile", "Archetype", "Worktree", "Owner"
    And the JSON "Roadmap" object contains fields "Present", "Planned", "InProgress", "Done", "Total", "Phases"
    And the JSON "Gates" array contains entries with fields "Gate" and "Fails"
    And the output contains no ANSI escape sequences

  Scenario: --json with nil roadmap emits Roadmap.Present false and exits 0
    Given no roadmap.json file exists
    When the operator runs:
      centinela dashboard --json
    Then the command exits with code 0
    And the output is valid JSON
    And the JSON field "Roadmap.Present" is false

  Scenario: --json with no active workflows emits empty Features array and exits 0
    Given no active workflow files exist
    When the operator runs:
      centinela dashboard --json
    Then the command exits with code 0
    And the output is valid JSON
    And the JSON field "Features" is an empty array

  Scenario: --json with no gate-failure events emits empty Gates array and exits 0
    Given no telemetry log file exists
    When the operator runs:
      centinela dashboard --json
    Then the command exits with code 0
    And the output is valid JSON
    And the JSON field "Gates" is an empty array

  Scenario: --json output has stable field names across two runs on the same state
    Given a fixed on-disk state with one workflow, a roadmap, and a telemetry log
    When the operator runs centinela dashboard --json twice in succession
    Then both outputs are byte-identical

  Scenario: --json field names are the stable contract — no renamed or extra top-level keys
    Given a fixed on-disk state
    When the operator runs:
      centinela dashboard --json
    Then the JSON object has exactly the top-level fields "Features", "Roadmap", "Gates"
    And no additional or renamed fields appear

  # ---------------------------------------------------------------------------
  # Aggregator purity — no I/O, no git in internal/teamdashboard
  # ---------------------------------------------------------------------------

  Scenario: internal/teamdashboard.Compute performs no file I/O or git calls
    Given Compute is called with a fully-populated in-memory Inputs struct
    When Compute runs to completion
    Then no file descriptors are opened inside internal/teamdashboard
    And no os/exec or git calls are made inside internal/teamdashboard
    And the returned Dashboard matches the expected output deterministically

  Scenario: Compute is deterministic — same Inputs struct produces byte-identical Dashboard
    Given a fixed Inputs struct with two workflows, a roadmap, and gate-failure events
    When Compute is called twice in succession
    Then both returned Dashboard values are identical

  # ---------------------------------------------------------------------------
  # Non-TTY / piped output
  # ---------------------------------------------------------------------------

  Scenario: Piped output contains no ANSI escape sequences
    Given one active workflow, a roadmap, and a telemetry log are present
    When the operator pipes the output:
      centinela dashboard | cat
    Then the output contains no ANSI escape sequences
    And the output is plain text parseable by grep

  Scenario: --json piped output also contains no ANSI escape sequences
    Given a fixed on-disk state
    When the operator pipes:
      centinela dashboard --json | cat
    Then the output contains no ANSI escape sequences
    And the output is valid JSON
