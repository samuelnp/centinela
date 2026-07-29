Feature: Unified plan specialist — one planner role replaces big-thinker + feature-specialist
  As an orchestrating agent running the plan step
  I want ONE planner subagent, at the reasoning tier, producing one evidence
    pair with both the strategy lens and the spec lens
  So that the plan step costs roughly half the subagent tokens and produces
    half the evidence files, while an in-flight legacy workflow keeps working
    exactly as it does today and a fresh workflow cannot dodge the new
    contract by hand-authoring old-format files

  # Design decisions this spec encodes (docs/plans/unified-plan-specialist.md):
  # D1 the role slug is "planner"; RoleBigThinker/RoleFeatureSpecial remain
  # declared constants, retired from RequiredRoles but still accepted by the
  # evidence CLI and config for legacy workflows. D2 back-compat is state-dated
  # via PlanContract pinned at `centinela start` ("planner-v1"); empty means the
  # workflow predates the migration and requires the COMPLETE legacy pair —
  # this supersedes the brief's original "accepts EITHER set" wording (Feature
  # Brief AC4), which the big-thinker report records as a transition hole a
  # pinned-contract rule closes. D3 one contract-aware resolver,
  # workflow.RequiredEvidenceRoles, feeds the hook directive, the complete
  # gate, and claim verification alike — no surface may diverge (the PR #83
  # CRITICAL class). D4 planner resolves to the reasoning tier. D5/D6 the
  # merged prompt doc keeps both lenses' full content in strategy-then-spec
  # order and targets <=130 lines. D7 legacy-role evidence stubs are refused
  # contract-awarely: allowed for an unpinned (legacy) workflow, refused
  # naming "planner" for a pinned or nonexistent workflow. D8 legacy model
  # override keys (big-thinker/feature-specialist) alias to planner and never
  # brick config load; deprecation is surfaced once, at doctor/start, never on
  # every hook invocation. D9 plan-advisor lens tags become "strategy"/"spec".

  # ── Fresh workflow: one planner role, one evidence pair ──────────────────

  Scenario: A fresh workflow's plan directive names exactly one planner role
    Given a feature "new-widget" at the plan step
    And its workflow was pinned with PlanContract "planner-v1" at start
    When the user requests hook context for "new-widget"
    Then the directive should name exactly one role, "planner"
    And the directive should name its model tier as reasoning
    And the directive should list only .workflow/new-widget-planner.md and
      .workflow/new-widget-planner.json as required plan evidence
    And the directive should not mention "big-thinker" or "feature-specialist"

  Scenario: Planner evidence with both artifacts passes plan complete
    Given a feature "new-widget" at the plan step
    And its workflow was pinned with PlanContract "planner-v1" at start
    And .workflow/new-widget-planner.md and .workflow/new-widget-planner.json
      exist with status "done", inputs snapshotting every docs/features/*.md
      plus docs/plans/new-widget.md, a non-empty edgeCases list, and
      handoffTo "senior-engineer"
    When the user runs "centinela complete new-widget"
    Then the plan step should advance without blocking
    And the code step should become active

  Scenario: A fresh workflow refuses a hand-authored legacy-role evidence pair
    Given a feature "new-widget" at the plan step
    And its workflow was pinned with PlanContract "planner-v1" at start
    And .workflow/new-widget-big-thinker.md and
      .workflow/new-widget-feature-specialist.md were hand-authored to
      satisfy the old two-role set, with no planner evidence written
    When the user runs "centinela complete new-widget"
    Then completion should be hard-blocked
    And the block message should name "planner" as the required role
    And the message should not accept the legacy pair as satisfying a
      "planner-v1" workflow
      # a fresh feature cannot dodge the new format by forging old-named files

  Scenario: evidence init for a retired legacy role errors on a pinned workflow
    Given a feature "new-widget" at the plan step
    And its workflow was pinned with PlanContract "planner-v1" at start
    When the operator runs "centinela evidence init new-widget big-thinker"
    Then the command should fail
    And the error should say the role is retired and point at "planner"

  Scenario: evidence init for a retired legacy role errors with no workflow state at all
    Given no workflow state exists for feature "ghost-feature"
    When the operator runs "centinela evidence init ghost-feature feature-specialist"
    Then the command should fail
    And the error should say the role is retired and point at "planner"

  # ── Legacy in-flight workflow: state-dated back-compat (D2) ──────────────

  Scenario: A legacy in-flight workflow still requires and accepts the complete two-role set
    Given a feature "already-in-flight" at the plan step
    And its workflow's PlanContract is empty because it started before this
      feature shipped
    And .workflow/already-in-flight-big-thinker.md and
      .workflow/already-in-flight-feature-specialist.md both exist with
      status "done", the required input snapshot, and handoffTo values that
      chain big-thinker -> feature-specialist -> senior-engineer
    When the user runs "centinela complete already-in-flight"
    Then the plan step should advance without blocking
      # empty PlanContract => today's two-role gate, verbatim

  Scenario: A partial legacy set still fails, naming this workflow's required set
    Given a feature "already-in-flight" at the plan step
    And its workflow's PlanContract is empty because it started before this
      feature shipped
    And only .workflow/already-in-flight-big-thinker.md exists
    When the user runs "centinela complete already-in-flight"
    Then completion should be hard-blocked
    And the block message should name the legacy pair (big-thinker AND
      feature-specialist) as the set this workflow requires
    And the message should carry a one-line contract annotation explaining
      this workflow predates "planner-v1"
    And the message should not offer "planner" as an alternative this
      workflow could satisfy

  Scenario: evidence init for a legacy role succeeds on an unpinned workflow
    Given a feature "already-in-flight" at the plan step
    And its workflow's PlanContract is empty because it started before this
      feature shipped
    When the operator runs "centinela evidence init already-in-flight big-thinker"
    Then the command should succeed
    And .workflow/already-in-flight-big-thinker.json should be written

  # ── Directive and gate must agree for every contract state (D3) ──────────

  Scenario Outline: The hook directive and the complete gate name the same required set
    Given a feature "<feature>" at the plan step
    And its workflow's PlanContract is "<contract>"
    When the user requests hook context for "<feature>"
    And the user separately runs "centinela complete <feature>" with no
      evidence written
    Then the directive's named role set should equal "<roles>"
    And the complete-gate's required role set should equal "<roles>"

    Examples:
      | feature           | contract     | roles                              |
      | pinned-workflow    | planner-v1   | planner                            |
      | unpinned-workflow   |              | big-thinker, feature-specialist   |

  # ── Guided/outcome profile still names the single planner ────────────────

  Scenario: A guided profile with no subagent evidence still names one planner
    Given a feature "guided-feature" at the plan step configured for the
      guided profile (no subagent evidence required)
    And its workflow was pinned with PlanContract "planner-v1" at start
    When the user requests hook context for "guided-feature"
    Then the directive should name the single role "planner"
    And no evidence requirement should be printed in the directive

  # ── Config: legacy model-override keys alias to planner, deprecation surfaced once (D8) ─

  Scenario: A legacy per-role model override key still resolves, aliased to planner
    Given a centinela.toml with "[orchestration.models]" key "big-thinker"
      set to a specific model and no explicit "planner" entry
    When the orchestration model tier is resolved for role "planner"
    Then the resolved model should equal the "big-thinker" override value
    And loading the config should not error

  Scenario: The legacy-key deprecation notice appears at doctor and start, not on every hook call
    Given a centinela.toml with legacy "[orchestration.models]" keys
      "big-thinker" and "feature-specialist"
    When the user runs "centinela doctor"
    Then a deprecation check should suggest migrating to "planner"
    When the user runs "centinela start another-feature"
    Then the same deprecation notice should print once
    When the user requests hook context during the plan step
    Then the hook directive output should NOT repeat the deprecation notice

  # ── Plan-advisor: two lenses, one agent (D9) ──────────────────────────────

  Scenario: Plan-advisor phrases its questions as two lenses for one agent
    Given the plan-advisor is invoked for a feature at the plan step
    When the advisor renders its header and question lens tags
    Then the header should read "One planner agent, two lenses: strategy
      first, then spec."
    And each rendered question should be tagged "[strategy]" or "[spec]"
    And no question should be tagged "[big-thinker]" or "[feature-specialist]"
    And the advisor should not instruct delegating to two separate agents

  # ── Planner prompt doc: both lenses, ordered, within budget (D5/D6) ───────

  Scenario: The planner prompt doc contains both lenses in strategy-then-spec order
    Given the file docs/architecture/planner-prompt.md
    When the file is read
    Then it should contain a "Lens 1: strategy" heading
    And it should contain a "Lens 2: spec" heading
    And the strategy heading's line number should be less than the spec
      heading's line number
    And it should contain a single "## Purpose" heading
    And it should contain a single "## Prompt Template" heading
    And it should contain a single "## Required Artifact" heading
    And it should contain exactly one CLI-authoring-rules block (not two)

  Scenario: The planner prompt doc respects the line budget
    Given the file docs/architecture/planner-prompt.md
    When the file is measured
    Then it should be at most 130 lines
    Or, if the budget map records an explicit "planner-prompt.md" exception,
      the total line count of all plan-step prompt docs should be strictly
      less than 215

  Scenario: The legacy prompt docs no longer exist
    Given the unified-plan-specialist feature is implemented
    When docs/architecture is inspected
    Then docs/architecture/big-thinker-prompt.md should not exist
    And docs/architecture/feature-specialist-prompt.md should not exist
    And docs/architecture/planner-prompt.md should exist

  # ── Scaffold parity for changed prompt docs ───────────────────────────────

  Scenario: The planner prompt is mirrored byte-identically in the scaffold tree
    Given docs/architecture/planner-prompt.md exists
    When the scaffold mirror is checked
    Then internal/scaffold/assets/docs/architecture/planner-prompt.md should exist
    And the two files should be byte-identical
    And internal/scaffold/assets/docs/architecture/big-thinker-prompt.md
      should not exist
    And internal/scaffold/assets/docs/architecture/feature-specialist-prompt.md
      should not exist

  # ── Statusline shows planner ──────────────────────────────────────────────

  Scenario: The statusline names planner for plan-step delegation
    Given a feature "new-widget" at the plan step with in-progress evidence
    When the user runs "centinela status new-widget" or views the statusline
    Then the displayed role for the plan step should be "planner"
    And it should not display "big-thinker" or "feature-specialist" for a
      "planner-v1" workflow

  Scenario: The statusline still names the legacy pair for an unpinned workflow
    Given a feature "already-in-flight" at the plan step with an empty
      PlanContract
    When the user runs "centinela status already-in-flight" or views the
      statusline
    Then the displayed roles for the plan step should include "big-thinker"
      and "feature-specialist"
