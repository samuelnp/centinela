Feature: Hook context panel diet — no trailing whitespace, same governance signal
  As the orchestrating agent reading injected hook context
  I want the ACTIVE WORKFLOWS and nudge panels rendered without trailing-space
    padding
  So that every UserPromptSubmit turn stops paying for an invisible rectangle
    that no longer has a border to square

  # Design decisions this spec encodes (docs/plans/hook-context-panel-diet.md):
  # D1 the fix lives in internal/ui, wrapping the return of the six
  # padding-bearing render functions hook_context.go calls — none of them
  # have a caller outside hook_context.go, so no hook/TTY mode flag is needed.
  # D2 no panel content is removed: the step ladder is the only per-turn
  # signal of which archetype-specific steps a workflow runs (canonical,
  # hotfix, refactor, and spike all have different step sets), and nudge
  # panels self-identify their feature because they render in a flat loop
  # across all active workflows, not nested under a header.
  # D3 the size guard measures a fixed, non-repo-dependent fixture, never the
  # live active-workflow count. D4 RenderStatus / RenderRoadmap / other
  # CLI-only renderers are untouched — centinela status, validate, and
  # blocked-write output keep their current bytes exactly.

  Background:
    Given a workflow "sample-feature" is active at the "plan" step

  # ── Padding removal ──────────────────────────────────────────────────────

  Scenario: Hook context output carries no trailing whitespace
    When the UserPromptSubmit hook renders the active-workflows panel
    Then no line of the rendered output should end with a space or a tab

  Scenario: A feature-brief-required nudge panel carries no trailing whitespace
    Given "sample-feature" has no docs/features/sample-feature.md yet
    When the UserPromptSubmit hook renders the feature-brief-required panel
    Then no line of the rendered output should end with a space or a tab

  Scenario: A review-required nudge panel carries no trailing whitespace
    Given the "plan" step's artifacts for "sample-feature" already validate
    When the UserPromptSubmit hook renders the review-required panel
    Then no line of the rendered output should end with a space or a tab

  # ── Governance signal survives ───────────────────────────────────────────

  Scenario: The active feature, its step, and its progress still appear
    When the UserPromptSubmit hook renders the active-workflows panel
    Then the output should contain the feature name "sample-feature"
    And the output should contain the current step "plan"
    And the output should contain the step progress "0/5"

  Scenario: The full step ladder still distinguishes archetypes with different gates
    Given a "spike" archetype workflow "quick-check" is active at the "code" step
    When the UserPromptSubmit hook renders the active-workflows panel
    Then the output should show "quick-check" has no "validate" step in its ladder
    And the output should show "sample-feature" includes a "validate" step in its ladder

  Scenario: A blocking review-required instruction still appears when due
    Given the "plan" step's artifacts for "sample-feature" already validate
    And step confirmation mode requires a pause before advancing
    When the UserPromptSubmit hook renders context
    Then the output should instruct the agent to stop and ask before advancing
    And the output should name "sample-feature" so a multi-workflow session
      can tell which feature the instruction is about

  # ── Human TTY rendering unchanged ────────────────────────────────────────

  Scenario: centinela status output is unchanged, trailing whitespace included
    When a human runs "centinela status sample-feature" in a terminal
    Then the rendered output bytes should be identical to the pre-diet render
    And the render should still come from RenderStatus, not from any
      hook-only render function touched by this feature

  Scenario: A blocked-write refusal panel is unaffected by this feature
    When the prewrite hook blocks a write and renders the refusal panel
    Then the rendered output bytes should be identical to the pre-diet render

  # ── Size guard ────────────────────────────────────────────────────────────

  Scenario: The size guard passes at the recorded budget
    Given the fixed measurement fixture "sample-feature" at the "plan" step
      with its feature-brief-required panel active
    When the active-workflows panel, the review-required panel, and the
      feature-brief-required panel are rendered and concatenated
    Then the total byte length should be at or under the recorded budget

  Scenario: The size guard fails when panel output grows past budget
    Given the fixed measurement fixture "sample-feature" at the "plan" step
      with its feature-brief-required panel active
    And the rendered output has been artificially padded past the recorded
      budget
    When the size guard test runs
    Then it should fail, naming the byte length and the budget it exceeded

  Scenario: The size guard does not depend on this repository's live workflow count
    Given two different machines with different numbers of real active
      workflows and different numbers of docs/features/*.md files
    When each machine runs the size guard test
    Then both runs should measure the same fixed fixture and produce the
      same pass/fail result
