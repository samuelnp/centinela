Feature: evidence schema skeleton agrees with the completion handoff gate
  As an agent running `centinela evidence schema <role>` mid-workflow
  I want the printed handoffTo to be either the value the chain gate will
    actually accept, or an obvious unfilled slot
  So I never discover a wrong handoffTo only when `centinela complete` fails

  Scenario: Derive with feature — canonical internal feature, worktree CWD
    Given a workflow "demo-internal" started with archetype "canonical"
    And its brief has no "surface: user-facing" line
    And the shell CWD is inside ".worktrees/demo-internal"
    When I run `centinela evidence schema gatekeeper`
    Then the printed "feature" is "demo-internal"
    And the printed "handoffTo" equals what `centinela evidence init demo-internal gatekeeper` would prefill
    And `centinela evidence set demo-internal gatekeeper handoffTo <that value>` followed by `centinela evidence validate demo-internal` reports no handoffTo issue

  Scenario: Derive with feature — hotfix archetype has no docs step
    Given a workflow "demo-hotfix" started with archetype "hotfix"
    And the shell CWD is inside ".worktrees/demo-hotfix"
    When I run `centinela evidence schema gatekeeper`
    Then the printed "handoffTo" is "complete"
    And it is NOT "documentation-specialist"

  Scenario: Derive with feature — user-facing feature, same-step ux-ui hop
    Given a workflow "demo-uxfacing" started with archetype "canonical"
    And its brief has a "surface: user-facing" line
    And the shell CWD is inside ".worktrees/demo-uxfacing"
    When I run `centinela evidence schema senior-engineer`
    Then the printed "handoffTo" is "ux-ui-specialist"

  Scenario: No-feature path — CWD resolves nothing
    Given the shell CWD is not inside any ".worktrees/<feature>" path
    And there are zero active workflows on disk
    When I run `centinela evidence schema gatekeeper`
    Then the printed "feature" is "<feature-slug>"
    And the printed "handoffTo" is "<successor-role>"
    And the output is still valid JSON
    And pasting this JSON verbatim as evidence for a real, resolvable feature
      and running `centinela evidence validate <that feature>` reports a
      handoffTo issue naming the true successor and a fix command

  Scenario: No-feature path — ambiguous CWD (parallel sessions) never guesses
    Given the shell CWD is not inside any ".worktrees/<feature>" path
    And there are two or more active workflows on disk
    When I run `centinela evidence schema gatekeeper`
    Then the printed "feature" is "<feature-slug>"
    And the printed "handoffTo" is "<successor-role>"
    And neither active workflow's feature name appears in the output

  Scenario: Unknown role is rejected before any derivation happens
    When I run `centinela evidence schema bogus`
    Then the command fails with an "unknown role" error
    And nothing is printed to stdout

  Scenario: Merge-steward is out-of-band — never guessed, never placeholder'd
    Given the shell CWD is not inside any ".worktrees/<feature>" path
    And there are zero active workflows on disk
    When I run `centinela evidence schema merge-steward`
    Then the printed "handoffTo" is "complete"
    And it is NOT "<successor-role>"

  Scenario: Merge-steward with a resolvable feature is unaffected by derivation
    Given a workflow "demo-internal" started with archetype "canonical"
    And the shell CWD is inside ".worktrees/demo-internal"
    When I run `centinela evidence schema merge-steward`
    Then the printed "handoffTo" is "complete"

  Scenario: Derivation is identical from a subdirectory of the worktree
    Given a workflow "demo-internal" started with archetype "canonical"
    And the shell CWD is a nested subdirectory inside ".worktrees/demo-internal"
    When I run `centinela evidence schema gatekeeper`
    Then the printed "handoffTo" is "complete", the same value the worktree
      root prints
    And it is NOT "documentation-specialist", because a correct feature slug
      printed beside a legacy handoff is more convincing than no slug at all

  Scenario: A resolved slug with no readable workflow state is not guessed at
    Given the shell CWD contains a ".worktrees/not-a-real-feature" segment
    And no workflow state exists for that slug
    When I run `centinela evidence schema gatekeeper`
    Then the printed "handoffTo" is "<successor-role>"
    And the printed "feature" is "<feature-slug>", because a slug the CLI
      cannot verify must not be echoed back as though it were confirmed

  Scenario: The placeholder is refused loudly for a contract-required role
    Given evidence for a role the workflow's contract requires
    And its "handoffTo" is still the unfilled "<successor-role>" placeholder
    When "centinela complete" runs
    Then completion is rejected naming the successor the contract derives
    And the error prints the runnable command that repairs it

  Scenario: The placeholder is NOT refused for a role the contract omits
    Given evidence for a role this workflow's contract does not require
    And its "handoffTo" is still the unfilled "<successor-role>" placeholder
    When "centinela evidence validate" runs
    Then it reports the evidence as ok, because the chain check only covers
      roles the step actually requires — the loud refusal is conditional, and
      the command help says so
