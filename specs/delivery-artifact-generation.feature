Feature: Delivery artifact generation
  As an operator (and the orchestrating agent) delivering a completed Centinela feature
  I want the PR body and changelog line composed from the evidence Centinela already holds
  So delivery output is consistent, traceable to the plan and gates, and never fabricated

  Background:
    Given a feature "alpha" exists in worktree mode at ".worktrees/alpha"
    And the repository has an "origin" remote
    And a "CHANGELOG.md" exists with a "## [Unreleased]" block containing "### Added", "### Changed", and "### Fixed" subsections

  Scenario: PR delivery composes the body from evidence instead of the --fill commit dump
    Given the "gh" CLI is available and authenticated
    And the feature brief, plan, gatekeeper report, and changelog stub for "alpha" are present
    When the user runs "centinela deliver alpha --via pr"
    Then the branch for "alpha" should be pushed to "origin"
    And a pull request should be opened via "gh" using "--body-file" and not "--fill"
    And the PR body should contain a summary section sourced from the feature brief or plan
    And the PR body should contain a "What changed / Why" section
    And the PR body should contain an acceptance reference pointing to "specs/alpha.feature"
    And the PR body should contain a gate status line sourced from the gatekeeper report verdict
    And the PR body should end with a Centinela provenance footer
    And the command should exit zero

  Scenario: Delivery inserts exactly one Keep-a-Changelog line under the correct subsection
    Given the changelog stub for "alpha" seeds a "feat:" shaped line
    When the user runs "centinela deliver alpha --via pr"
    Then exactly one new bullet line should be added under "### Added" inside the "## [Unreleased]" block
    And no other "### " subsection should gain a bullet
    And no section outside the "## [Unreleased]" block should be modified
    And the command should exit zero

  Scenario: Re-running delivery does not duplicate the changelog line
    Given delivery for "alpha" has already inserted its "[Unreleased]" changelog line
    When the user runs "centinela deliver alpha --via pr" again
    Then the "[Unreleased]" block should still contain exactly one copy of that line
    And "CHANGELOG.md" should be left unchanged by the changelog step
    And the command should exit zero

  Scenario: A feat-shaped change lands under Added
    Given the seed changelog line for "alpha" begins with "feat:"
    When delivery composes the changelog entry for "alpha"
    Then the entry category should be "Added"

  Scenario: A fix-shaped change lands under Fixed
    Given the seed changelog line for "alpha" begins with "fix:"
    When delivery composes the changelog entry for "alpha"
    Then the entry category should be "Fixed"

  Scenario: Any other change shape lands under Changed
    Given the seed changelog line for "alpha" begins with "refactor:"
    When delivery composes the changelog entry for "alpha"
    Then the entry category should be "Changed"

  Scenario: A missing evidence source omits its section rather than fabricating it
    Given the "gh" CLI is available and authenticated
    And the gatekeeper report for "alpha" is absent
    When the user runs "centinela deliver alpha --via pr"
    Then the PR body should not contain a gate status line
    And the PR body should still contain the provenance footer
    And the remaining sources should still produce their sections
    And the command should exit zero

  Scenario: The gate status line is never faked when no passing gate can be sourced
    Given the "gh" CLI is available and authenticated
    And neither a gatekeeper verdict nor a verification tally can be sourced for "alpha"
    When the user runs "centinela deliver alpha --via pr"
    Then the PR body should not assert that any gate passed
    And the gate status line should be omitted
    And the command should exit zero

  Scenario: Composition still succeeds when the brief and plan are both absent
    Given the "gh" CLI is available and authenticated
    And the feature brief and plan for "alpha" are both absent
    When the user runs "centinela deliver alpha --via pr"
    Then the summary, what/why, and acceptance sections should be omitted
    And the PR body should still contain the provenance footer
    And a pull request should still be opened via "gh"
    And the command should exit zero

  Scenario: PR delivery with no origin remote is refused
    Given the repository has no "origin" remote
    When the user runs "centinela deliver alpha --via pr"
    Then the command should report that there is no origin remote and PR delivery is unavailable
    And no branch should be pushed
    And no PR should have been created
    And the command should exit non-zero

  Scenario: PR delivery when gh is absent pushes, prints honest manual instructions, and exits non-zero
    Given the "gh" CLI is absent or unauthenticated
    When the user runs "centinela deliver alpha --via pr"
    Then the branch for "alpha" should be pushed to "origin"
    And the output should contain manual instructions to open a pull request
    And the output should not claim that a pull request was opened
    And the command should exit non-zero
