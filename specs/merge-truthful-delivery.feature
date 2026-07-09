Feature: centinela merge / deliver --via merge — truthful delivery
  As an operator (or orchestrating agent) delivering a completed feature
  I want merge delivery to operate on the primary working tree and to verify
  the ref advanced and the worktree is gone before claiming anything
  So that "Merged X into main and removed its worktree" is always true —
  success is asserted on the ref, not on the absence of git errors

  # Field evidence: 2048-rust retrospective — 4 consecutive false successes
  # and one near branch-loss. Root cause: repo="." from inside a worktree
  # makes `git merge --no-ff <branch>` self-no-op ("Already up to date"),
  # Remove/Exists stat the wrong CWD, and merge.go printed success
  # unconditionally. The fix resolves the primary tree via
  # `git worktree list --porcelain` (first entry) and verifies
  # rev-parse HEAD before/after plus merge-base --is-ancestor.
  # Scenario titles map 1:1 to Go acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed git repository whose primary tree is on branch "main"
    And a completed feature "high-score" with worktree ".worktrees/high-score"
      containing commits on branch "high-score" that main does not have

  # ---------------------------------------------------------------------------
  # the regression — merge from inside the feature worktree
  # ---------------------------------------------------------------------------

  Scenario: merge from inside the feature worktree advances main and removes the worktree
    Given the operator's CWD is ".worktrees/high-score"
    And the primary tree's "git rev-parse main" is captured as "before"
    When the operator runs:
      centinela deliver high-score --via merge
    Then the command exits with code 0
    And "git rev-parse main" in the primary tree differs from "before"
      # the retrospective's named success metric: assert on the REF
    And the branch "high-score" is an ancestor of main
    And the directory ".worktrees/high-score" no longer exists
    And stdout contains the success message naming "high-score"

  Scenario: merge run from inside a worktree prints a one-line primary-tree notice
    Given the operator's CWD is ".worktrees/high-score"
    When the operator runs:
      centinela merge high-score
    Then the command exits with code 0
    And stdout contains a one-line notice naming the primary working tree path
      # IsInsideWorktree(".") finally has a caller

  Scenario: merge run from the primary tree behaves as before, with verification applied
    Given the operator's CWD is the primary working tree
    When the operator runs:
      centinela merge high-score
    Then the command exits with code 0
    And "git rev-parse main" advanced and ".worktrees/high-score" is removed
    And stdout contains no worktree notice

  # ---------------------------------------------------------------------------
  # refusals — never guess, never claim
  # ---------------------------------------------------------------------------

  Scenario: merge refuses when the primary tree cannot be resolved
    Given the operator's CWD is a directory that is not inside any git repository
    When the operator runs:
      centinela merge high-score
    Then the command exits with a non-zero code
    And stderr contains "cannot resolve primary working tree"
    And no merge is attempted and no success message is printed

  Scenario: merge refuses when the primary working tree is bare
    Given "git worktree list --porcelain" reports the first entry as bare
    When the operator runs:
      centinela merge high-score
    Then the command exits with a non-zero code
    And stderr mentions the bare primary working tree
    And no success message is printed

  Scenario: merge refuses when the primary tree is in detached HEAD state
    Given the primary working tree has a detached HEAD
    When the operator runs:
      centinela merge high-score
    Then the command exits with a non-zero code
    And "git rev-parse main" in the primary tree is unchanged
    And no success message is printed

  Scenario: merge refuses when the primary tree has the feature branch checked out
    Given the primary working tree has the branch "high-score" checked out
    When the operator runs:
      centinela merge high-score
    Then the command exits with a non-zero code
    And the merge is not reported as already merged
    And no success message is printed

  Scenario: dirty primary tree is still refused
    Given the primary working tree has uncommitted changes
    And the operator's CWD is ".worktrees/high-score"
    When the operator runs:
      centinela merge high-score
    Then the command exits with a non-zero code
    And stderr says the main working tree is dirty
    And "git rev-parse main" in the primary tree is unchanged
    And the directory ".worktrees/high-score" still exists
      # the existing isDirty guard now checks the RIGHT tree

  # ---------------------------------------------------------------------------
  # truthfulness — the success line follows the ref
  # ---------------------------------------------------------------------------

  Scenario: the success message is never printed when the ref did not advance
    Given a merge attempt where git exits 0 but the primary tree's HEAD
      does not move and "high-score" is NOT an ancestor of main
    When the merge flow completes
    Then the command exits with a non-zero code
    And stderr contains "did not advance"
    And stdout does not contain "Merged \"high-score\" into main"
      # RefAdvanced=false AND AlreadyMerged=false → hard error, no success

  Scenario: re-running deliver for an already-merged branch reports honestly
    Given branch "high-score" is already an ancestor of main
    And the primary tree's "git rev-parse main" is captured as "before"
    When the operator runs:
      centinela deliver high-score --via merge
    Then the command exits with code 0
    And "git rev-parse main" in the primary tree equals "before"
    And stdout says "high-score" was already merged
      # truthful wording: NOT the fabricated "Merged … into main"
    And the directory ".worktrees/high-score" no longer exists
      # cleanup still runs on the already-merged path

  Scenario: already-merged re-run with the worktree already gone is an idempotent honest success
    Given branch "high-score" is already an ancestor of main
    And the directory ".worktrees/high-score" does not exist
    When the operator runs:
      centinela merge high-score
    Then the command exits with code 0
    And stdout says "high-score" was already merged

  Scenario: removal is only claimed when the worktree directory is actually gone
    Given a merge where main advances but ".worktrees/high-score" survives removal
    When the merge flow completes
    Then the command exits with a non-zero code
    And no message claims the worktree was removed

  # ---------------------------------------------------------------------------
  # unchanged paths — no success claimed, no verification needed
  # ---------------------------------------------------------------------------

  Scenario: a text conflict still dispatches the Merge Steward with no success claim
    Given branch "high-score" conflicts textually with main
    When the operator runs:
      centinela merge high-score
    Then the merge stops on the conflict outcome as today
    And the worktree ".worktrees/high-score" is kept
    And no success message is printed

  Scenario: a validate failure after a clean text merge still dispatches the steward
    Given the merged tree fails "centinela validate"
    When the operator runs:
      centinela merge high-score
    Then the merge stops on the validate-fail outcome as today
    And the worktree ".worktrees/high-score" is kept
    And no success message is printed
