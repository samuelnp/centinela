Feature: Parallel feature worktrees with merge steward
  As a maintainer driving multiple feature workflows in parallel
  I want each feature to run in its own git worktree and merge back via a Steward
  So concurrent agents do not collide and conflicts surface before they reach main

  Scenario: Start provisions a worktree when use_worktrees is enabled
    Given the project has workflow.use_worktrees set to true
    And no feature named "alpha" is in flight
    When the user runs "centinela start alpha"
    Then a worktree should exist at ".worktrees/alpha"
    And a branch named "alpha" should be checked out inside that worktree
    And the feature workflow state should be written under ".worktrees/alpha/.workflow/"

  Scenario: Start runs in the main checkout when use_worktrees is disabled
    Given the project has workflow.use_worktrees set to false
    When the user runs "centinela start beta"
    Then no directory should be created under ".worktrees/"
    And the feature workflow state should be written under the repo root ".workflow/"

  Scenario: Wizard syncs tool ignore lists for new projects
    Given the wizard initialises a new project with workflow.use_worktrees enabled
    When the wizard finishes
    Then ".gitignore", ".eslintignore", ".prettierignore", ".dockerignore", and ".rgignore" should each include an entry for ".worktrees/"
    And "tsconfig.json" exclude should include ".worktrees"
    And re-running the wizard should not duplicate any of those entries

  Scenario: Migrate syncs tool ignore lists for existing projects
    Given an existing project without ".worktrees/" entries in its ignore files
    When the user runs "centinela migrate" and opts into workflow.use_worktrees
    Then ".gitignore", ".eslintignore", ".prettierignore", ".dockerignore", ".rgignore", and "tsconfig.json" exclude should each include ".worktrees/"
    And the migrate command should be safe to re-run with no further changes

  Scenario: Clean merge when git applies cleanly and validate passes
    Given a completed feature "gamma" exists at ".worktrees/gamma"
    And the main working tree is clean
    When the user runs "centinela merge gamma"
    Then git should merge "gamma" into main with no text conflict
    And "centinela validate" should run against the merged main tree and pass
    And the Merge Steward should not be invoked
    And the worktree at ".worktrees/gamma" should be removed after the merge succeeds

  Scenario: Text conflict invokes the Merge Steward
    Given a completed feature "delta" exists at ".worktrees/delta"
    And merging "delta" into main produces a git text conflict in at least one file
    When the user runs "centinela merge delta"
    Then the Merge Steward agent should be invoked with the conflicted paths and the feature spec
    And a Merge Steward evidence file should be written to ".workflow/delta-merge-steward.json"

  Scenario: Semantic conflict after a clean text merge invokes the Steward
    Given a completed feature "epsilon" exists at ".worktrees/epsilon"
    And merging "epsilon" into main produces no text conflict
    But "centinela validate" fails on the merged tree
    When the user runs "centinela merge epsilon"
    Then the Merge Steward agent should be invoked with the failing validate output and the feature spec
    And a Merge Steward evidence file should be written to ".workflow/epsilon-merge-steward.json"

  Scenario: Spec conflict across in-flight worktrees is detected before merging
    Given two in-flight features "zeta" and "eta" each have a worktree
    And "specs/zeta.feature" and "specs/eta.feature" both assert different observable outcomes for the same Given context
    When the user runs "centinela merge zeta"
    Then the spec-conflict pre-check should fail before git merge runs
    And the output should name both feature files and the conflicting scenario
    And no commits should be added to main

  Scenario: Merge Steward escalates uncertain resolutions to the user
    Given the Merge Steward is invoked for feature "theta"
    And the Steward's proposed resolution is not high-confidence
    When the Steward completes its evaluation
    Then the merge command should exit without modifying main
    And the user should see the Steward's proposed diff and the reason the confidence is low
    And the Steward evidence at ".workflow/theta-merge-steward.json" should record the escalation and the proposed diff

  Scenario: Restarting a feature with an existing worktree resumes in place
    Given a worktree already exists at ".worktrees/iota" from a previous start
    When the user runs "centinela start iota"
    Then the existing worktree should be reused without recreating the branch
    And no error should be reported
    And the existing ".worktrees/iota/.workflow/" state should be preserved

  Scenario: Merge fails fast when the main working tree is dirty
    Given the main checkout has uncommitted changes
    When the user runs "centinela merge kappa"
    Then the command should exit with a non-zero status before invoking git merge
    And the error message should instruct the user to commit or stash main before merging
    And the worktree at ".worktrees/kappa" should be left untouched
