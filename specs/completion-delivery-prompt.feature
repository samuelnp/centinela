Feature: Completion delivery prompt
  As an operator (and the orchestrating agent) finishing a Centinela feature
  I want completion to surface only the valid delivery options and a command that acts on an explicit pick
  So completed work is delivered the team's way — asked first, never pushed or merged by a guess

  Scenario: Completion with an origin remote and worktree mode offers both PR and local merge
    Given a feature "alpha" has reached the "done" step in worktree mode at ".worktrees/alpha"
    And the repository has an "origin" remote
    When the user runs "centinela complete alpha"
    Then the output should contain a "CENTINELA DIRECTIVE:" naming the feature "alpha" and instructing the agent to ask the user how to deliver
    And the directive should list the option "centinela deliver alpha --via pr"
    And the directive should list the option "centinela deliver alpha --via merge"
    And the directive should state not to push or merge without the user's explicit choice
    And no branch should be pushed and no merge should be performed
    And the command should exit zero

  Scenario: Completion with no origin remote offers only the local-merge option
    Given a feature "beta" has reached the "done" step in worktree mode at ".worktrees/beta"
    And the repository has no "origin" remote
    When the user runs "centinela complete beta"
    Then the output should contain a "CENTINELA DIRECTIVE:" naming the feature "beta"
    And the directive should list the option "centinela deliver beta --via merge"
    And the directive should not list any "--via pr" option
    And no branch should be pushed and no merge should be performed
    And the command should exit zero

  Scenario: Completion in single-checkout mode with an origin remote offers only the PR option
    Given a feature "gamma" has reached the "done" step in single-checkout mode with no worktree path
    And the repository has an "origin" remote
    When the user runs "centinela complete gamma"
    Then the output should contain a "CENTINELA DIRECTIVE:" naming the feature "gamma"
    And the directive should list the option "centinela deliver gamma --via pr"
    And the directive should not list any "--via merge" option
    And no branch should be pushed and no merge should be performed
    And the command should exit zero

  Scenario: Completion with neither an origin remote nor worktree mode reports no delivery target
    Given a feature "delta" has reached the "done" step in single-checkout mode with no worktree path
    And the repository has no "origin" remote
    When the user runs "centinela complete delta"
    Then the output should contain a "CENTINELA DIRECTIVE:" stating no delivery target was detected
    And the directive should list neither a "--via pr" nor a "--via merge" option
    And no branch should be pushed and no merge should be performed
    And the command should exit zero

  Scenario: The completion directive never delivers by itself
    Given a feature "epsilon" has reached the "done" step in worktree mode at ".worktrees/epsilon"
    And the repository has an "origin" remote
    When the user runs "centinela complete epsilon"
    Then the directive should be emitted as text only
    And no "git push" should have run and no PR should have been created
    And the worktree at ".worktrees/epsilon" should be kept
    And the branch for "epsilon" should not have been merged into main

  Scenario: deliver without --via refuses to act and exits non-zero
    Given a feature "zeta" exists in worktree mode at ".worktrees/zeta"
    When the user runs "centinela deliver zeta"
    Then the command should report that "--via pr|merge" must be chosen
    And no branch should be pushed and no merge should be performed
    And the command should exit non-zero

  Scenario: deliver --via pr with no origin remote refuses to act and exits non-zero
    Given a feature "eta" exists in worktree mode at ".worktrees/eta"
    And the repository has no "origin" remote
    When the user runs "centinela deliver eta --via pr"
    Then the command should report that there is no origin remote and PR delivery is unavailable
    And no branch should be pushed
    And no PR should have been created
    And the command should exit non-zero

  Scenario: deliver --via merge delegates to the existing merge flow on a clean merge
    Given a feature "theta" exists in worktree mode at ".worktrees/theta"
    And merging "theta" into main produces no text conflict
    And "centinela validate" passes on the merged tree
    When the user runs "centinela deliver theta --via merge"
    Then the merge should be finalized through the existing "centinela merge" flow
    And the worktree at ".worktrees/theta" should be removed
    And no ".workflow/theta-merge-pending.json" marker should be written
    And the command should exit zero

  Scenario: deliver --via merge on a conflicted merge reuses the merge-steward dispatch
    Given a feature "iota" exists in worktree mode at ".worktrees/iota"
    And merging "iota" into main produces a git text conflict in at least one file
    When the user runs "centinela deliver iota --via merge"
    Then a ".workflow/iota-merge-pending.json" marker should be written
    And the output should contain a CENTINELA DIRECTIVE naming the merge-steward prompt and the feature "iota"
    And the worktree at ".worktrees/iota" should be kept
    And the command should exit non-zero

  Scenario: deliver --via pr with origin and gh available pushes and reports the opened PR
    Given a feature "kappa" exists in worktree mode at ".worktrees/kappa"
    And the repository has an "origin" remote
    And the "gh" CLI is available and authenticated
    When the user runs "centinela deliver kappa --via pr"
    Then the branch for "kappa" should be pushed to "origin"
    And a pull request should be opened via "gh"
    And the output should report the opened pull request URL
    And the command should exit zero

  Scenario: deliver --via pr when gh is absent still pushes, prints manual instructions, and exits non-zero
    Given a feature "lambda" exists in worktree mode at ".worktrees/lambda"
    And the repository has an "origin" remote
    And the "gh" CLI is absent or unauthenticated
    When the user runs "centinela deliver lambda --via pr"
    Then the branch for "lambda" should be pushed to "origin"
    And the output should contain manual instructions to open a pull request
    And the output should not claim that a pull request was opened
    And the command should exit non-zero
