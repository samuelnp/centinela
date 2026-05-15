Feature: Merge Steward auto-dispatch
  As a maintainer merging completed feature worktrees
  I want a needed Merge Steward to be dispatched automatically and gate finalization
  So conflicted merges resolve through evidence, never silently or by hand

  Scenario: Clean merge does not dispatch the Steward (regression guard)
    Given a completed feature "gamma" exists at ".worktrees/gamma"
    And merging "gamma" into main produces no text conflict
    And "centinela validate" passes on the merged tree
    When the user runs "centinela merge gamma"
    Then no ".workflow/gamma-merge-pending.json" marker should be written
    And no CENTINELA DIRECTIVE for the Merge Steward should be emitted
    And the worktree at ".worktrees/gamma" should be removed
    And the command should exit zero

  Scenario: Text conflict writes the pending marker and dispatches the Steward
    Given a completed feature "delta" exists at ".worktrees/delta"
    And merging "delta" into main produces a git text conflict in at least one file
    When the user runs "centinela merge delta"
    Then a ".workflow/delta-merge-pending.json" marker should be written recording reason "git-text-conflict", the conflicted paths, and the worktree path
    And the output should contain a CENTINELA DIRECTIVE naming the merge-steward prompt, the feature "delta", and the "centinela merge --continue delta" resume command
    And the worktree at ".worktrees/delta" should be kept
    And the command should exit non-zero

  Scenario: Post-merge validate failure dispatches the Steward like a text conflict
    Given a completed feature "epsilon" exists at ".worktrees/epsilon"
    And merging "epsilon" into main produces no text conflict
    But "centinela validate" fails on the merged tree
    When the user runs "centinela merge epsilon"
    Then a ".workflow/epsilon-merge-pending.json" marker should be written recording reason "post-merge-validate-failed" and the worktree path
    And the output should contain a CENTINELA DIRECTIVE naming the merge-steward prompt, the feature "epsilon", and the "centinela merge --continue epsilon" resume command
    And the worktree at ".worktrees/epsilon" should be kept
    And the command should exit non-zero

  Scenario: The hook re-emits the directive while the marker exists without valid evidence
    Given a ".workflow/zeta-merge-pending.json" marker exists
    And no valid ".workflow/zeta-merge-steward.json" evidence is present
    When the merge hook runs on a subsequent prompt
    Then it should re-emit the CENTINELA DIRECTIVE naming the merge-steward prompt and the feature "zeta"

  Scenario: The hook stops re-emitting once valid steward evidence is present
    Given a ".workflow/eta-merge-pending.json" marker exists
    And valid ".workflow/eta-merge-steward.json" evidence is present
    When the merge hook runs on a subsequent prompt
    Then no CENTINELA DIRECTIVE for the Merge Steward should be emitted

  Scenario: The hook is silent when no pending marker exists
    Given no ".workflow/theta-merge-pending.json" marker exists
    When the merge hook runs on a subsequent prompt
    Then no CENTINELA DIRECTIVE for the Merge Steward should be emitted

  Scenario: Continue with APPLY evidence finalizes the merge
    Given a ".workflow/iota-merge-pending.json" marker exists and the worktree at ".worktrees/iota" is kept
    And ".workflow/iota-merge-steward.json" evidence is valid with status APPLY and handoffTo "complete"
    And the main working tree is clean
    When the user runs "centinela merge --continue iota"
    Then the worktree at ".worktrees/iota" should be removed
    And the ".workflow/iota-merge-pending.json" marker should be cleared
    And the command should exit zero

  Scenario: Continue with ESCALATE evidence keeps the merge blocked
    Given a ".workflow/kappa-merge-pending.json" marker exists and the worktree at ".worktrees/kappa" is kept
    And ".workflow/kappa-merge-steward.json" evidence is valid with status ESCALATE and handoffTo "user"
    When the user runs "centinela merge --continue kappa"
    Then the merge should not be finalized
    And the worktree at ".worktrees/kappa" should be kept
    And the ".workflow/kappa-merge-pending.json" marker should be kept
    And the Steward escalation note and proposed diff should be printed to stderr
    And the command should exit non-zero

  Scenario: Continue with missing steward evidence refuses to finalize
    Given a ".workflow/lambda-merge-pending.json" marker exists and the worktree at ".worktrees/lambda" is kept
    And no ".workflow/lambda-merge-steward.json" evidence file is present
    When the user runs "centinela merge --continue lambda"
    Then the merge should not be finalized
    And the worktree at ".worktrees/lambda" should be kept
    And the output should contain an actionable error stating steward evidence is required
    And the command should exit non-zero

  Scenario: Continue with schema-invalid steward evidence refuses to finalize
    Given a ".workflow/mu-merge-pending.json" marker exists and the worktree at ".worktrees/mu" is kept
    And ".workflow/mu-merge-steward.json" exists but fails the orchestration evidence validator
    When the user runs "centinela merge --continue mu"
    Then the merge should not be finalized
    And the worktree at ".worktrees/mu" should be kept
    And the output should contain the orchestration evidence validation error
    And the command should exit non-zero

  Scenario: Continue with APPLY evidence but a dirty main tree refuses to finalize
    Given a ".workflow/nu-merge-pending.json" marker exists and the worktree at ".worktrees/nu" is kept
    And ".workflow/nu-merge-steward.json" evidence is valid with status APPLY and handoffTo "complete"
    But the main working tree has uncommitted changes
    When the user runs "centinela merge --continue nu"
    Then the merge should not be finalized
    And the worktree at ".worktrees/nu" should be kept
    And the ".workflow/nu-merge-pending.json" marker should be kept
    And the command should exit non-zero

  Scenario: Continue with no pending marker reports a clear error
    Given no ".workflow/xi-merge-pending.json" marker exists
    When the user runs "centinela merge --continue xi"
    Then the command should report there is no pending merge to continue for "xi"
    And no worktree or marker state should change
    And the command should exit non-zero

  Scenario: Re-running merge while a pending marker exists does not lose the marker
    Given a ".workflow/omicron-merge-pending.json" marker exists from a prior "git-text-conflict"
    And merging "omicron" into main now produces a post-merge validate failure
    When the user runs "centinela merge omicron"
    Then exactly one ".workflow/omicron-merge-pending.json" marker should exist
    And the marker should be rewritten with reason "post-merge-validate-failed", not appended
    And the worktree at ".worktrees/omicron" should be kept
    And the command should exit non-zero
