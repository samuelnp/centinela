Feature: Host harness adapters
  As a Centinela maintainer
  I want each host harness to be expressed as a registered HarnessAdapter
  So that adding new harnesses requires only one new file, not edits to the sync planner, apply switch, or CLI commands

  # ── AC1: Registry lookup and typed error for unknown agents ──────────────────

  Scenario: Registry resolves a known agent name to its adapter
    Given the adapter registry is initialised
    When I look up the agent "claude"
    Then I receive the claude HarnessAdapter without error
    When I look up the agent "opencode"
    Then I receive the opencode HarnessAdapter without error
    When I look up the agent "aider"
    Then I receive the aider HarnessAdapter without error

  Scenario: Registry returns a typed error for an unknown agent
    Given the adapter registry is initialised
    When I look up the agent "vscode"
    Then I receive a typed ErrUnknownAgent error
    And the error message lists the registered harness names "claude", "opencode", "aider"
    And no panic occurs

  # ── AC2: BuildSyncPlan driven by the registry ─────────────────────────────────

  Scenario: BuildSyncPlan for "claude" produces only Claude managed items
    Given the adapter registry is initialised
    When I call BuildSyncPlan with agent "claude"
    Then the plan contains a SyncItem for ".claude/settings.json"
    And the plan contains no items for "opencode.json" or "AGENTS.md" or ".aider.conf.yml"

  Scenario: BuildSyncPlan for "opencode" produces only OpenCode managed items
    Given the adapter registry is initialised
    When I call BuildSyncPlan with agent "opencode"
    Then the plan contains a SyncItem for "opencode.json"
    And the plan contains a SyncItem for ".opencode/plugins/centinela.js"
    And the plan contains a SyncItem for "AGENTS.md"
    And the plan contains no items for ".claude/settings.json" or ".aider.conf.yml"

  Scenario: BuildSyncPlan for "aider" produces only Aider managed items
    Given the adapter registry is initialised
    When I call BuildSyncPlan with agent "aider"
    Then the plan contains a SyncItem for "AGENTS.md"
    And the plan contains a SyncItem for ".aider.conf.yml"
    And the plan contains no items for ".claude/settings.json" or "opencode.json"

  Scenario: BuildSyncPlan for "both" composes Claude and OpenCode items
    Given the adapter registry is initialised
    When I call BuildSyncPlan with agent "both"
    Then the plan contains all Claude items and all OpenCode items
    And the plan contains no items for ".aider.conf.yml"
    And the result is identical to calling BuildSyncPlan for "claude" plus BuildSyncPlan for "opencode"

  Scenario: BuildSyncPlan contains no per-harness if-ladder
    Given the BuildSyncPlan implementation
    Then it iterates the registry to compose the plan
    And there is no hardcoded "useClaude" or "useOpenCode" predicate inside BuildSyncPlan

  # ── AC3: Capabilities declared by each adapter ───────────────────────────────

  Scenario: Claude adapter declares all three capabilities
    Given the adapter registry is initialised
    When I call Capabilities() on the claude adapter
    Then the result contains "blocks-writes"
    And the result contains "prompt-context"
    And the result contains "rules-file"

  Scenario: OpenCode adapter declares all three capabilities
    Given the adapter registry is initialised
    When I call Capabilities() on the opencode adapter
    Then the result contains "blocks-writes"
    And the result contains "prompt-context"
    And the result contains "rules-file"

  Scenario: Aider adapter declares prompt-context and rules-file but not blocks-writes
    Given the adapter registry is initialised
    When I call Capabilities() on the aider adapter
    Then the result contains "prompt-context"
    And the result contains "rules-file"
    And the result does NOT contain "blocks-writes"

  # ── AC4: Golden-file byte parity for Claude and OpenCode ─────────────────────

  Scenario: Claude managed output is byte-for-byte identical after refactor
    Given a fixture project with a pre-refactor snapshot of ".claude/settings.json"
    When BuildSyncPlan is called with agent "claude" and the plan is applied
    Then the emitted ".claude/settings.json" bytes match the golden snapshot exactly

  Scenario: OpenCode managed output is byte-for-byte identical after refactor
    Given a fixture project with a pre-refactor snapshot of "opencode.json", ".opencode/plugins/centinela.js", and "AGENTS.md"
    When BuildSyncPlan is called with agent "opencode" and the plan is applied
    Then each emitted file's bytes match the golden snapshot exactly

  # ── AC5: Aider init and migrate are idempotent and scoped ────────────────────

  Scenario: centinela init --agent aider writes Aider managed files
    Given a project with no Aider configuration
    When I run "centinela init --agent aider"
    Then "AGENTS.md" is created containing a Centinela managed region
    And ".aider.conf.yml" is created containing a managed region with a "read:" entry pointing to "AGENTS.md"
    And ".claude/settings.json" is not created or modified
    And "opencode.json" is not created or modified

  Scenario: centinela init --agent aider is idempotent on re-run
    Given a project already initialised with "centinela init --agent aider"
    When I run "centinela init --agent aider" again
    Then "AGENTS.md" is unchanged
    And ".aider.conf.yml" is unchanged
    And the exit code is 0

  Scenario: centinela migrate --agent aider is idempotent and scoped
    Given a project already managed by centinela for claude
    When I run "centinela migrate --agent aider"
    Then "AGENTS.md" is created or updated with a Centinela managed region
    And ".aider.conf.yml" is created or updated with a managed region
    And ".claude/settings.json" is not modified

  Scenario: Pre-existing unmanaged .aider.conf.yml is not clobbered
    Given a project with a hand-written ".aider.conf.yml" containing no centinela managed marker
    When I run "centinela init --agent aider"
    Then ".aider.conf.yml" is not overwritten
    And a manual-review warning is surfaced to the user

  # ── AC6: --agent validation lists registered harnesses ───────────────────────

  Scenario: --agent with a known value is accepted
    Given the centinela CLI
    When I run "centinela init --agent aider"
    Then the command does not fail with a validation error

  Scenario: --agent with an unknown value lists registered harnesses
    Given the centinela CLI
    When I run "centinela init --agent unknown-tool"
    Then the command exits with a non-zero status
    And the error output contains "claude"
    And the error output contains "opencode"
    And the error output contains "aider"

  Scenario: isValidAgent is resolved by the registry, not a hardcoded list
    Given the adapter registry
    When I call the registry's validation function with "aider"
    Then it returns true without consulting a hardcoded string list

  # ── AC7: Capability-parity invariant ─────────────────────────────────────────

  Scenario: Every registered adapter declares a non-empty capability set
    Given the adapter registry is initialised
    When I iterate all registered adapters
    Then each adapter's Capabilities() result is non-empty

  Scenario: Any adapter claiming blocks-writes wires a prewrite hook
    Given the adapter registry is initialised
    When I iterate all registered adapters
    Then for every adapter whose Capabilities() includes "blocks-writes"
    The plan items produced by PlanItems() include a SyncItem of kind SyncKindPrewriteHook

  Scenario: Aider does not wire a prewrite hook
    Given the adapter registry is initialised
    When I call PlanItems() on the aider adapter
    Then no SyncItem in the result has kind SyncKindPrewriteHook

  # ── Edge cases ────────────────────────────────────────────────────────────────

  Scenario: AGENTS.md shared surface - OpenCode and Aider share the file without double-write
    Given a project initialised for "opencode"
    When I run "centinela init --agent aider"
    Then "AGENTS.md" is written exactly once
    And the Centinela managed region appears exactly once in "AGENTS.md"

  Scenario: Partial existing install - adding Aider leaves Claude files untouched
    Given a project fully initialised for "claude"
    When I run "centinela init --agent aider"
    Then ".claude/settings.json" retains its original content unchanged
    And "AGENTS.md" is created or updated
    And ".aider.conf.yml" is created

  Scenario: Hook-less harness cannot claim blocks-writes
    Given an adapter implementation that declares "blocks-writes" in its Capabilities
    When the capability-parity test runs
    Then the test asserts that a SyncKindPrewriteHook item exists in PlanItems
    And an adapter with no prewrite hook item fails this assertion

  Scenario: both selector composes adapters without a special-case branch
    Given the registry contains a "both" composition entry for claude and opencode
    When BuildSyncPlan is called with agent "both"
    Then the planner iterates the registry for "both" and unions the two adapters' PlanItems results
    And there is no hardcoded "both" special case outside the registry definition
