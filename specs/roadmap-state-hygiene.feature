Feature: roadmap state hygiene — self-committing mutations, stamp-safe regeneration, Backlog aging
  As an agent or operator mutating shared roadmap state from one of several
  parallel sessions
  I want every roadmap mutation to leave ROADMAP.md in sync and the tree clean,
  conflicts to be resolved by a command, and the Backlog to have an age
  So that a deferral can never dirty a tree into an aborted merge, void a
  verification stamp, or be lost in a hand-resolved conflict

  # Source: retrospective.md WS1.4 + WS6 + failure #10, plus Phase 13 field
  # evidence (every delivery conflicted on .workflow/roadmap.json and ROADMAP.md;
  # commit f3886dd records a verifier finding LOST in such a merge).
  # "roadmap state" throughout means exactly: .workflow/roadmap.json and
  # ROADMAP.md, plus the analysis/quality artifacts a promote also rewrites.
  # Scenario titles map 1:1 to Go acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed git repository with a valid .workflow/roadmap.json
    And ROADMAP.md is in sync with .workflow/roadmap.json
    And workflow.disable_auto_commit is not set

  # ---------------------------------------------------------------------------
  # AC1/AC2 — regeneration folded into every mutation, pathspec-scoped commit
  # ---------------------------------------------------------------------------

  Scenario: a deferral regenerates ROADMAP.md and commits roadmap state by itself
    When the operator runs: centinela roadmap defer "flaky-thing" --summary "one line"
    Then the command exits with code 0
    And ROADMAP.md byte-matches the roadmap renderer's output for the new roadmap.json
    And the roadmap_drift gate reports no drift without running roadmap generate
    And exactly one new commit exists
    And that commit's changed paths are exactly ".workflow/roadmap.json" and "ROADMAP.md"
    And that commit's message is "chore(roadmap): defer flaky-thing"
    And the working tree reports no modified roadmap state

  Scenario Outline: every roadmap.json mutation regenerates and commits
    When the operator runs a "<mutation>" mutation that changes roadmap.json
    Then ROADMAP.md is regenerated in the same command
    And exactly one commit is created whose changed paths are roadmap-state paths only
    And the commit message starts with "chore(roadmap): <verb>"

    Examples:
      | mutation      | verb         |
      | add           | add          |
      | remove        | remove       |
      | edit          | edit         |
      | move          | move         |
      | reorder       | reorder      |
      | promote       | promote      |
      | phase add     | phase add    |
      | phase rename  | phase rename |
      | phase remove  | phase remove |

  Scenario: a promote also commits the analysis and quality artifacts it rewrote
    Given the Backlog contains the finding "worth-doing"
    When the operator promotes "worth-doing" into a schedulable phase with scores
    Then the commit's changed paths include ".workflow/roadmap-analysis.json" and ".workflow/roadmap-quality.json"
    And no path outside the declared roadmap-state pathspec is in the commit

  Scenario: the removal message no longer tells the operator to sync by hand
    When the operator runs: centinela roadmap remove "planned-thing"
    Then the output does not contain "Remember to sync"

  # ---------------------------------------------------------------------------
  # AC3/AC7 — other work is untouched; no empty commits
  # ---------------------------------------------------------------------------

  Scenario: unrelated staged and unstaged changes survive the mutation commit
    Given "internal/example/thing.go" is staged with changes
    And "README.md" is modified but not staged
    When the operator runs: centinela roadmap defer "another-thing" --summary "x"
    Then the new commit does not contain "internal/example/thing.go"
    And the new commit does not contain "README.md"
    And "internal/example/thing.go" is still staged with the same changes
    And "README.md" is still modified and unstaged

  Scenario: a mutation that changes nothing on disk creates no commit
    When the operator runs a mutation that leaves roadmap.json byte-identical
    Then no new commit is created
    And the command exits with code 0

  # ---------------------------------------------------------------------------
  # AC4 — disable_auto_commit
  # ---------------------------------------------------------------------------

  Scenario: disable_auto_commit skips the commit but never skips regeneration
    Given workflow.disable_auto_commit is true
    When the operator runs: centinela roadmap defer "policy-thing" --summary "x"
    Then the command exits with code 0
    And ROADMAP.md is regenerated and in sync with roadmap.json
    And no new commit is created
    And the output states that roadmap state was left uncommitted

  # ---------------------------------------------------------------------------
  # AC5 — the commit is best-effort and never fails the mutation
  # ---------------------------------------------------------------------------

  Scenario Outline: a hostile git environment warns but never fails the mutation
    Given the repository is in the state "<git state>"
    When the operator runs: centinela roadmap defer "resilient-thing" --summary "x"
    Then the command exits with code 0
    And roadmap.json contains the finding "resilient-thing"
    And ROADMAP.md is in sync with roadmap.json
    And the output warns that roadmap state was left uncommitted because of "<reason>"

    Examples:
      | git state                | reason              |
      | not a git repository     | no git repository   |
      | no commits yet (no HEAD) | no HEAD             |
      | a merge in progress      | merge in progress   |
      | a rebase in progress     | rebase in progress  |
      | git commit fails         | commit failed       |

  # ---------------------------------------------------------------------------
  # AC6 — worktree locality
  # ---------------------------------------------------------------------------

  Scenario: a mutation inside a feature worktree commits to that worktree's branch
    Given the current directory is ".worktrees/some-feature" on branch "some-feature"
    When the operator runs: centinela roadmap defer "worktree-thing" --summary "x"
    Then the new commit is on branch "some-feature"
    And the primary checkout's HEAD is unchanged
    And the primary checkout's working tree and index are unchanged

  # ---------------------------------------------------------------------------
  # AC8 — generate stays standalone and never commits
  # ---------------------------------------------------------------------------

  Scenario: roadmap generate regenerates without committing
    Given ROADMAP.md has drifted from roadmap.json
    When the operator runs: centinela roadmap generate
    Then ROADMAP.md is back in sync
    And no new commit is created
    And ROADMAP.md is left modified in the working tree

  Scenario: roadmap generate works while a merge is in progress
    Given a merge is in progress
    When the operator runs: centinela roadmap generate
    Then the command exits with code 0
    And no commit is attempted

  # ---------------------------------------------------------------------------
  # AC9 — freshness is roadmap-state-blind, and only that
  # ---------------------------------------------------------------------------

  Scenario: a deferral after the verification stamp does not stale the verification
    Given a gatekeeper report stamped against the current revision and tree
    When the operator runs: centinela roadmap defer "late-finding" --summary "x"
    And the roadmap-state commit advances HEAD
    Then the verification freshness check reports the verification as fresh

  Scenario: an uncommitted regenerated ROADMAP.md does not stale the verification
    Given workflow.disable_auto_commit is true
    And a gatekeeper report stamped against the current revision and tree
    When the operator runs a roadmap mutation that regenerates ROADMAP.md
    Then the verification freshness check reports the verification as fresh

  Scenario: a source change committed after the stamp still stales the verification
    Given a gatekeeper report stamped against the current revision and tree
    When a commit touching "internal/example/thing.go" and roadmap state lands on HEAD
    Then the verification freshness check reports the verification as stale
    And the message names the verified revision and the current HEAD

  Scenario: an unreadable revision range fails closed
    Given a gatekeeper report stamped against a revision git cannot resolve
    When the verification freshness check runs
    Then the verification is reported as stale

  # ---------------------------------------------------------------------------
  # AC10 — roadmap backlog / --stale
  # ---------------------------------------------------------------------------

  Scenario: roadmap backlog lists deferred findings oldest first with their age
    Given the Backlog contains findings deferred 40, 20 and 2 days ago
    When the operator runs: centinela roadmap backlog
    Then the command exits with code 0
    And the findings are listed oldest first
    And each row shows the age in days, the slug, the source feature and role, and the summary
    And the footer reports 3 findings and how many are older than 14 days

  Scenario: --stale shows only findings older than the default threshold
    Given the Backlog contains findings deferred 40, 20 and 2 days ago
    When the operator runs: centinela roadmap backlog --stale
    Then only the findings deferred 40 and 20 days ago are listed

  Scenario: --older-than overrides the default threshold
    Given the Backlog contains findings deferred 40, 20 and 2 days ago
    When the operator runs: centinela roadmap backlog --stale --older-than 30
    Then only the finding deferred 40 days ago is listed

  Scenario: a finding with no deferredAt is reported as unknown age and counted stale
    Given the Backlog contains a finding with no "deferredAt" field
    And the Backlog contains a finding with an unparseable "deferredAt" value
    When the operator runs: centinela roadmap backlog --stale
    Then both findings are listed first with age "unknown"
    And the command exits with code 0

  Scenario: --json emits the machine shape for every finding
    When the operator runs: centinela roadmap backlog --json
    Then stdout is valid JSON with "threshold_days", "total", "stale" and "findings"
    And each finding carries slug, summary, source, deferredAt, ageDays and stale

  Scenario: an empty Backlog is an explicit empty state, not an error
    Given the roadmap has no Backlog phase
    When the operator runs: centinela roadmap backlog --stale
    Then the command exits with code 0
    And the output states that there are no deferred findings

  # ---------------------------------------------------------------------------
  # AC11 — the completion nudge
  # ---------------------------------------------------------------------------

  Scenario: completing the last schedulable feature nudges about the Backlog
    Given every schedulable roadmap feature except "last-one" is done
    And the Backlog contains 12 findings, 9 of them older than 14 days
    When the operator runs: centinela complete last-one on its final step
    Then the output reports the roadmap is complete
    And the output states 12 deferred findings remain, 9 older than 14 days
    And the output names "centinela roadmap backlog --stale"
    And the output names "centinela roadmap promote"

  Scenario: no nudge while schedulable work remains
    Given at least one schedulable roadmap feature is not done
    And the Backlog is non-empty
    When a workflow completes its final step
    Then no Backlog nudge is printed

  Scenario: no nudge when the Backlog is empty
    Given every schedulable roadmap feature is done
    And the Backlog is empty
    When a workflow completes its final step
    Then no Backlog nudge is printed

  Scenario: an unreadable roadmap suppresses the nudge without failing complete
    Given .workflow/roadmap.json cannot be parsed
    When a workflow completes its final step
    Then the command exits with code 0
    And no Backlog nudge is printed

  # ---------------------------------------------------------------------------
  # AC12/AC13 — roadmap resolve
  # ---------------------------------------------------------------------------

  Scenario: resolve unions Backlog findings from both sides of a conflict
    Given .workflow/roadmap.json is conflicted
    And the base has 5 Backlog findings
    And our side added 2 findings and their side added 3 different findings
    And no schedulable phase differs between the two sides
    When the operator runs: centinela roadmap resolve
    Then the command exits with code 0
    And roadmap.json contains all 10 findings ordered by deferredAt then name
    And ROADMAP.md is regenerated from the resolved roadmap.json
    And both ".workflow/roadmap.json" and "ROADMAP.md" are staged
    And the summary names how many findings came from each side

  Scenario: resolve keeps one entry when both sides added the same slug
    Given both sides added a Backlog finding with the slug "same-thing"
    When the operator runs: centinela roadmap resolve
    Then roadmap.json contains exactly one finding with the slug "same-thing"
    And its deferredAt is the earlier of the two

  Scenario: resolve honours a one-sided deletion
    Given the base contains the finding "promoted-away"
    And our side removed it and their side left it untouched
    When the operator runs: centinela roadmap resolve
    Then roadmap.json does not contain "promoted-away"

  Scenario: resolve refuses when a schedulable phase diverged on both sides
    Given both sides changed the features of phase "Phase 13: Lighter Centinela"
    When the operator runs: centinela roadmap resolve
    Then the command exits with a non-zero code
    And the message names phase "Phase 13: Lighter Centinela"
    And .workflow/roadmap.json still contains its conflict markers
    And nothing is staged

  Scenario: resolve refuses when one side of the conflict is not valid JSON
    Given the conflicted roadmap.json cannot be parsed on their side
    When the operator runs: centinela roadmap resolve
    Then the command exits with a non-zero code
    And the message names the unparseable side
    And nothing is staged

  Scenario: resolve regenerates when only ROADMAP.md is conflicted
    Given .workflow/roadmap.json merged cleanly
    And ROADMAP.md is conflicted
    When the operator runs: centinela roadmap resolve
    Then ROADMAP.md is regenerated from the merged roadmap.json with no markers left
    And ROADMAP.md is staged
    And the command exits with code 0

  Scenario: resolve is a no-op outside a conflict
    Given no roadmap-state path is conflicted
    When the operator runs: centinela roadmap resolve
    Then the command exits with code 0
    And the output states there is nothing to resolve
    And no file is modified or staged

  # ---------------------------------------------------------------------------
  # AC14 — canonical rendering
  # ---------------------------------------------------------------------------

  Scenario: every phase renders one feature object per line on every write
    When any roadmap mutation writes roadmap.json
    Then every phase's "features" array has exactly one compact JSON object per line
    And phases untouched by the mutation are not otherwise reformatted

  Scenario: a second mutation produces a diff confined to the lines it changed
    Given a roadmap.json already written by this renderer
    When a single feature's description is edited
    Then the resulting git diff touches only the lines of that feature

  Scenario: rendering is idempotent
    Given a roadmap.json already written by this renderer
    When a no-op mutation rewrites it
    Then the file is byte-identical
    And unknown top-level and per-feature fields survive the round trip
