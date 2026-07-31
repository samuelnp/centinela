Feature: dynamic-model-routing
  Static model routing stays the byte-identical default; with
  [orchestration] routing_mode = "dynamic" the orchestrator may route a
  role's tier per feature via `centinela route`, floors and refusal rules
  guard against governance weakening, and every decision is audited.

  Background:
    Given a centinela project with an initialized workflow for feature "f"
    And centinela.toml sets [orchestration] routing_mode = "dynamic"

  # --- happy paths ---

  Scenario: Downgrade with a reason is recorded before the step starts
    Given the workflow "f" is on the "plan" step
    And no evidence artifact exists for role "senior-engineer"
    When I run "centinela route set f senior-engineer balanced --reason 'config-only change'"
    Then the command exits 0
    And .workflow/f.json records modelRoutes.senior-engineer with tier "balanced", the reason, and an RFC3339 decidedAt
    And .workflow/telemetry/events.jsonl gains a "route-decision" event with role "senior-engineer", tier "balanced", prevTier "reasoning", and the reason

  Scenario: The hook directive reflects a routed tier and static config for un-routed roles
    Given role "senior-engineer" is routed to tier "balanced" for "f"
    And the workflow "f" is on the "code" step
    When "centinela hook orchestration" runs
    Then the delegate line annotates "senior-engineer" with "model: sonnet (claude)"
    And every un-routed role in the directive resolves through [orchestration.models] and built-in defaults unchanged

  Scenario: Upgrade is allowed anytime without a reason
    Given the workflow "f" is on the "code" step
    And the evidence artifact .workflow/f-senior-engineer.md already exists
    When I run "centinela route set f senior-engineer reasoning"
    Then the command exits 0
    And modelRoutes.senior-engineer records tier "reasoning" with an empty reason

  Scenario: route show renders the effective table with static fallbacks
    Given role "qa-senior" is routed to tier "fast" with reason "trivial rename" for "f"
    When I run "centinela route show f"
    Then the output lists each scheduled role with its effective tier
    And "qa-senior" shows tier "fast", source "routed", the reason, and its decidedAt
    And "gatekeeper" shows source "static" and floor "reasoning"

  Scenario: The routing directive is emitted while roles are unset and silent once routed
    Given the workflow "f" is on the "plan" step and role "planner" has no route
    When "centinela hook orchestration" runs
    Then the output contains a "routing (dynamic)" directive line naming "planner", its floor "balanced", and the "centinela route set" command
    When role "planner" is routed to tier "reasoning" and the hook runs again
    Then the output contains no "routing (dynamic)" line

  # --- refusals ---

  Scenario: A tier below the role's floor is refused naming the floor
    When I run "centinela route set f gatekeeper balanced --reason 'save cost'"
    Then the command exits non-zero
    And the error names the "gatekeeper" floor "reasoning"
    And .workflow/f.json records no route for "gatekeeper"

  Scenario: A downgrade below the static default without --reason is refused
    Given the static default tier for "qa-senior" is "balanced"
    When I run "centinela route set f qa-senior fast"
    Then the command exits non-zero
    And the error says a downgrade below the static default requires --reason
    And .workflow/f.json records no route for "qa-senior"

  Scenario: A downgrade is refused once the role's step is underway
    Given the workflow "f" is on the "code" step
    And the evidence artifact .workflow/f-senior-engineer.md already exists
    When I run "centinela route set f senior-engineer fast --reason 'late savings'"
    Then the command exits non-zero
    And the error says the "code" step is already underway for "senior-engineer"

  Scenario: A downgrade is refused for a completed step
    Given the workflow "f" has completed the "plan" step
    When I run "centinela route set f planner balanced --reason 'too late'"
    Then the command exits non-zero
    And the error says the "plan" step is already underway or completed

  Scenario: Unknown role and unknown tier are refused
    When I run "centinela route set f wizard fast --reason 'x'"
    Then the command exits non-zero and the error lists the allowed role slugs
    When I run "centinela route set f qa-senior turbo --reason 'x'"
    Then the command exits non-zero and the error lists "reasoning, balanced, fast"

  Scenario: A role not scheduled in this workflow's steps is refused
    Given feature "f" is not user-facing, so "ux-ui-specialist" is not required by any step
    When I run "centinela route set f ux-ui-specialist fast --reason 'x'"
    Then the command exits non-zero
    And the error says the role is not scheduled in this workflow

  # --- static mode ---

  Scenario: Static mode output is byte-identical and route commands are refused
    Given centinela.toml sets no routing_mode key
    When "centinela hook orchestration" runs and its output is captured
    And centinela.toml sets [orchestration] routing_mode = "static" and the hook runs again
    Then both outputs are byte-identical
    And neither output contains a "routing" directive line
    When I run "centinela route set f qa-senior fast --reason 'x'"
    Then the command exits non-zero and the error names [orchestration] routing_mode

  # --- compatibility ---

  Scenario: A pre-existing workflow JSON without modelRoutes loads cleanly with no routes
    Given .workflow/f.json was written by a binary that predates modelRoutes
    When I run "centinela route show f"
    Then every scheduled role shows source "static"
    And "centinela complete f" and the hooks behave exactly as before
