Feature: centinela doctor — project health diagnostics and safe repair
  As a developer or operator working in a Centinela-governed project
  I want `centinela doctor` to diagnose project-health problems in a single pass
  and `centinela doctor --fix` to apply only safe, idempotent repairs automatically
  So that broken hook wiring, roadmap drift, abandoned worktrees, stale workflow state,
  orphaned evidence, config drift, and binary version skew are visible up front
  rather than causing mysterious mid-workflow failures

  # Doctor runs every enabled check in a deterministic fixed order and prints one
  # line per check: status glyph (✓/⚠/✗), check name, message, optional details.
  # Exit 0 when no check is ERROR (OK and WARN both pass). Exit 1 when any ERROR.
  # --fix applies only Repair.Safe==true repairs, then re-diagnoses and re-renders.
  # Destructive actions (worktree remove, .workflow deletion) are NEVER applied
  # by --fix; they are reported with the exact command the user must run.
  # Scenario titles map 1:1 to Go acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with a valid centinela.toml
    And the project is inside a git repository

  # ---------------------------------------------------------------------------
  # Happy path — all checks OK
  # ---------------------------------------------------------------------------

  Scenario: Healthy project reports all checks OK and exits 0
    Given all hook entries are correctly wired in .claude/settings.json
    And ROADMAP.md is in sync with roadmap.json and no phase name contains a live-status glyph
    And no worktrees correspond to merged or fully-complete branches
    And no .workflow state exists without a corresponding branch or worktree
    And no orphaned *.json.tmp files exist under .workflow/
    And centinela.toml has verify_timeout at or above the floor and all gates dirs exist
    And the installed centinela binary version matches the Makefile VERSION
    When the operator runs:
      centinela doctor
    Then the command exits with code 0
    And the output contains one line per check each prefixed with "✓"
    And the output contains a summary line matching "N ok, 0 warn, 0 error"
    And no file on disk is modified

  # ---------------------------------------------------------------------------
  # Check 1 — hook wiring
  # ---------------------------------------------------------------------------

  Scenario: Missing hook entries are flagged as ERROR and --fix re-wires them
    Given .claude/settings.json exists but is missing the centinela hook entries
    When the operator runs:
      centinela doctor
    Then the command exits with code 1
    And the output contains a line with "✗" and the check name "hooks"
    And the output describes which hook entries are missing
    When the operator runs:
      centinela doctor --fix
    Then the hook entries are present in .claude/settings.json
    And the post-fix report line for "hooks" is prefixed with "✓"
    And the command exits with code 0

  Scenario: Hook re-wire under --fix is idempotent
    Given centinela doctor --fix has already re-wired the missing hook entries
    When the operator runs:
      centinela doctor --fix
    Then .claude/settings.json is byte-identical to its state before this second run
    And the report line for "hooks" is prefixed with "✓"
    And the command exits with code 0

  Scenario: No .claude directory causes hooks check to degrade to WARN not crash
    Given no .claude/ directory exists in the project
    When the operator runs:
      centinela doctor
    Then the command exits with code 0
    And the output contains a line with "⚠" and the check name "hooks"
    And the output mentions "centinela setup"
    And no panic or unhandled error is produced

  # ---------------------------------------------------------------------------
  # Check 2 — roadmap drift and phase-name glyph
  # ---------------------------------------------------------------------------

  Scenario: ROADMAP.md drift from roadmap.json is flagged
    Given roadmap.json has been edited but ROADMAP.md has not been regenerated
    When the operator runs:
      centinela doctor
    Then the command exits with code 1
    And the output contains a line with "✗" and the check name "roadmap"
    And the output indicates ROADMAP.md is out of sync with roadmap.json

  Scenario: Roadmap drift is repaired by --fix via regeneration
    Given roadmap.json has been edited but ROADMAP.md has not been regenerated
    When the operator runs:
      centinela doctor --fix
    Then ROADMAP.md is regenerated to match roadmap.json
    And the post-fix report line for "roadmap" is prefixed with "✓"
    And the command exits with code 0

  Scenario: Phase name containing a live-status glyph is flagged as ERROR
    Given roadmap.json contains a phase whose name starts with a Unicode status glyph such as "✅ Phase 0: Bootstrap"
    When the operator runs:
      centinela doctor
    Then the command exits with code 1
    And the output contains a line with "✗" and the check name "roadmap"
    And the output names the offending phase and explains the glyph breaks prefix detection

  Scenario: Phase-name glyph is stripped by --fix and re-diagnosis passes
    Given roadmap.json contains a phase named "✅ Phase 0: Bootstrap"
    When the operator runs:
      centinela doctor --fix
    Then the leading glyph is removed from the phase name in roadmap.json
    And ROADMAP.md is regenerated from the repaired roadmap.json
    And the post-fix report line for "roadmap" is prefixed with "✓"
    And the command exits with code 0

  Scenario: Roadmap glyph strip under --fix is idempotent
    Given centinela doctor --fix has already stripped the phase-name glyph
    When the operator runs:
      centinela doctor --fix
    Then roadmap.json is byte-identical to its state before this second run
    And the report line for "roadmap" is prefixed with "✓"
    And the command exits with code 0

  # ---------------------------------------------------------------------------
  # Check 3 — abandoned worktrees
  # ---------------------------------------------------------------------------

  Scenario: Abandoned worktree for a merged branch is reported with the removal command
    Given a worktree exists under .worktrees/ whose branch has been merged into main
    When the operator runs:
      centinela doctor
    Then the command exits with code 1
    And the output contains a line with "✗" and the check name "worktrees"
    And the output contains the exact git worktree remove command for the abandoned worktree
    And the worktree directory still exists on disk

  Scenario: --fix does NOT remove an abandoned worktree
    Given a worktree exists under .worktrees/ whose branch has been merged into main
    When the operator runs:
      centinela doctor --fix
    Then the worktree directory still exists on disk after --fix completes
    And the report line for "worktrees" still shows "✗" or "⚠" with the removal command
    And the output does NOT indicate the worktree was deleted

  Scenario: No worktrees present causes worktrees check to report OK
    Given no worktree directories exist under .worktrees/
    When the operator runs:
      centinela doctor
    Then the report line for "worktrees" is prefixed with "✓"

  # ---------------------------------------------------------------------------
  # Check 4 — stale .workflow state
  # ---------------------------------------------------------------------------

  Scenario: Orphaned .workflow state with no corresponding branch is reported
    Given a .workflow/*.json file exists for a feature whose branch no longer exists and has no active worktree
    When the operator runs:
      centinela doctor
    Then the command exits with code 1
    And the output contains a line with "✗" and the check name "workflow-state"
    And the output names the orphaned workflow file and provides the manual deletion command

  Scenario: --fix does NOT delete orphaned .workflow state
    Given a .workflow/*.json file exists for a feature with no live branch or worktree
    When the operator runs:
      centinela doctor --fix
    Then the .workflow/*.json file still exists on disk after --fix completes
    And the report line for "workflow-state" still surfaces the finding with a manual command

  # ---------------------------------------------------------------------------
  # Check 5 — orphaned evidence *.json.tmp
  # ---------------------------------------------------------------------------

  Scenario: Orphaned *.json.tmp files are flagged and removed by --fix
    Given one or more *.json.tmp files exist under .workflow/ from a crashed atomic write
    When the operator runs:
      centinela doctor
    Then the command exits with code 1
    And the output contains a line with "✗" and the check name "evidence"
    And the output lists the orphaned .json.tmp paths
    When the operator runs:
      centinela doctor --fix
    Then all *.json.tmp files under .workflow/ are deleted
    And the post-fix report line for "evidence" is prefixed with "✓"
    And the command exits with code 0

  Scenario: Orphaned evidence removal under --fix is idempotent
    Given centinela doctor --fix has already removed all *.json.tmp files
    When the operator runs:
      centinela doctor --fix
    Then no error is produced for the evidence check
    And the report line for "evidence" is prefixed with "✓"

  # ---------------------------------------------------------------------------
  # Check 6 — config drift
  # ---------------------------------------------------------------------------

  Scenario: verify_timeout below the suite floor is flagged as WARN
    Given centinela.toml has verify_timeout set below the minimum floor of 180 seconds
    When the operator runs:
      centinela doctor
    Then the command exits with code 0
    And the output contains a line with "⚠" and the check name "config"
    And the output mentions "verify_timeout" and the recommended minimum value

  Scenario: Gate referencing a non-existent directory is flagged as WARN
    Given centinela.toml references a gate directory that does not exist on disk
    When the operator runs:
      centinela doctor
    Then the command exits with code 0
    And the output contains a line with "⚠" and the check name "config"
    And the output names the missing directory

  Scenario: Unknown TOML keys in centinela.toml are flagged as WARN
    Given centinela.toml contains a key not recognized by the centinela config schema
    When the operator runs:
      centinela doctor
    Then the command exits with code 0
    And the output contains a line with "⚠" and the check name "config"
    And the output names each unrecognized key

  Scenario: Config check is report-only and --fix does not modify centinela.toml
    Given centinela.toml has verify_timeout below the floor
    When the operator runs:
      centinela doctor --fix
    Then centinela.toml is byte-identical to its state before --fix ran
    And the report line for "config" still shows "⚠" with the advisory message

  # ---------------------------------------------------------------------------
  # Check 7 — binary version skew
  # ---------------------------------------------------------------------------

  Scenario: Installed binary version behind Makefile VERSION is flagged as WARN
    Given the installed centinela binary reports version "0.15.0"
    And the Makefile in the repo root defines VERSION as a newer value
    When the operator runs:
      centinela doctor
    Then the command exits with code 0
    And the output contains a line with "⚠" and the check name "version"
    And the output reports both the installed version and the Makefile version
    And the output recommends running "make install"

  Scenario: Binary version check is report-only and --fix does not reinstall
    Given the installed binary is behind the Makefile VERSION
    When the operator runs:
      centinela doctor --fix
    Then the installed centinela binary is unchanged
    And the report line for "version" still shows "⚠" with the "make install" recommendation

  Scenario: centinela binary not found on PATH causes version check to degrade to WARN
    Given the centinela binary is not found on PATH
    When the operator runs:
      centinela doctor
    Then the command exits with code 0
    And the output contains a line with "⚠" and the check name "version"
    And no panic or unhandled error is produced

  # ---------------------------------------------------------------------------
  # Exit-code and output contract
  # ---------------------------------------------------------------------------

  Scenario: Any ERROR check causes exit code 1
    Given at least one check produces an ERROR diagnosis
    When the operator runs:
      centinela doctor
    Then the command exits with code 1

  Scenario: Only WARN checks present causes exit code 0
    Given all checks produce OK or WARN diagnoses and none produce ERROR
    When the operator runs:
      centinela doctor
    Then the command exits with code 0

  Scenario: Summary line always present and reflects actual check counts
    Given a project with two checks at OK, one at WARN, and one at ERROR
    When the operator runs:
      centinela doctor
    Then the output ends with a summary line reading "2 ok, 1 warn, 1 error"

  Scenario: Output is deterministic — check order is fixed regardless of finding severity
    Given two consecutive runs of centinela doctor on the same project state
    When the operator runs centinela doctor twice in succession
    Then both outputs list checks in the same fixed order
    And both outputs are byte-identical

  Scenario: Non-TTY output is plain and parseable with no spinner or ANSI codes
    Given standard output is redirected to a file (non-TTY)
    When the operator runs:
      centinela doctor
    Then every output line starts with one of "✓", "⚠", "✗", or is the summary line
    And no ANSI escape sequences are present in the output

  # ---------------------------------------------------------------------------
  # --fix behavioral contract
  # ---------------------------------------------------------------------------

  Scenario: --fix attempts all safe repairs even when one fails
    Given a project with three fixable issues and the second repair will fail at runtime
    When the operator runs:
      centinela doctor --fix
    Then the first repair is applied successfully
    And the third repair is applied successfully despite the second failing
    And the failed repair's check appears as "✗" in the post-fix report
    And the command exits with code 1

  Scenario: --fix partial success renders a clear per-check post-fix report
    Given centinela doctor --fix has been run and one repair succeeded while another failed
    When the output is examined
    Then each check line reflects the post-fix state of that specific check
    And the summary counts match the post-fix diagnosis results

  Scenario: --fix never performs destructive actions — worktree and .workflow intact after fix
    Given a project with an abandoned worktree and orphaned .workflow state alongside fixable evidence tmp files
    When the operator runs:
      centinela doctor --fix
    Then the abandoned worktree directory still exists on disk
    And the orphaned .workflow state file still exists on disk
    And the orphaned *.json.tmp files have been removed
    And the report shows the worktree and workflow-state findings with manual commands

  # ---------------------------------------------------------------------------
  # Multiple simultaneous problems
  # ---------------------------------------------------------------------------

  Scenario: Multiple problems in one run are all reported in a single pass
    Given a project where hooks are missing, ROADMAP.md is drifted, and a *.json.tmp file exists
    When the operator runs:
      centinela doctor
    Then the output contains finding lines for "hooks", "roadmap", and "evidence"
    And all three findings appear in the single command invocation output
    And the summary line reflects all three errors

  Scenario: --fix with multiple fixable problems repairs all of them in one invocation
    Given a project where hooks are missing and a *.json.tmp file exists
    When the operator runs:
      centinela doctor --fix
    Then hook entries are present in .claude/settings.json
    And no *.json.tmp files remain under .workflow/
    And the post-fix report shows "✓" for both "hooks" and "evidence"

  # ---------------------------------------------------------------------------
  # Robustness and environment edge cases
  # ---------------------------------------------------------------------------

  Scenario: Doctor runs from inside a worktree and resolves the repo root correctly
    Given the shell CWD is inside .worktrees/some-feature
    When the operator runs:
      centinela doctor
    Then all checks operate against the canonical repo root, not the worktree subdirectory
    And the hook check reads .claude/settings.json from the repo root
    And the roadmap check reads roadmap.json from the repo root
    And no check reports a missing-file error due to path resolution

  Scenario: Doctor runs from repo root and all checks locate their targets
    Given the shell CWD is the repo root
    When the operator runs:
      centinela doctor
    Then no check fails due to path resolution

  Scenario: Not inside a git repo causes git-dependent checks to degrade gracefully
    Given the project directory is not inside any git repository
    When the operator runs:
      centinela doctor
    Then the command does not panic or produce an unhandled error
    And the worktrees check output contains "⚠" with a message about no git context
    And the version check output contains "⚠" or "✓" and does not crash
    And the remaining checks that do not depend on git still produce diagnoses

  Scenario: Doctor does not require an active centinela workflow to run
    Given no centinela workflow is currently active for any feature
    When the operator runs:
      centinela doctor
    Then the command runs to completion without error due to absent workflow
    And all checks produce their normal OK/WARN/ERROR diagnosis

  Scenario: Doctor completes without crashing when a check's dependency is missing
    Given centinela.toml cannot be parsed due to a syntax error
    When the operator runs:
      centinela doctor
    Then the config check degrades to ERROR with a clear message naming the parse failure
    And all other checks that do not depend on the config still produce diagnoses
    And the command does not panic

