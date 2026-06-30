Feature: Codex harness support
  As a Centinela maintainer
  I want OpenAI Codex registered as a full first-class HarnessAdapter
  So that "centinela init --agent codex" wires PreToolUse/PostToolUse/
  UserPromptSubmit hooks and AGENTS.md with the same governance as Claude Code

  # ── AC1: Registry ────────────────────────────────────────────────────────────

  Scenario: Codex is a valid --agent selector
    Given the adapter registry is initialised
    When I look up the agent "codex"
    Then I receive the codex HarnessAdapter without error
    And the adapter Name() returns "codex"

  Scenario: both composite is unchanged by codex addition
    Given the adapter registry is initialised
    When I call BuildSyncPlan with agent "both"
    Then the plan contains no items for ".codex/config.toml"
    And the plan is identical to composing "claude" plus "opencode" only

  # ── AC2: Capabilities ────────────────────────────────────────────────────────

  Scenario: Codex adapter declares all three capabilities
    Given the adapter registry is initialised
    When I call Capabilities() on the codex adapter
    Then the result contains "blocks-writes"
    And the result contains "prompt-context"
    And the result contains "rules-file"

  Scenario: Codex adapter satisfies prewrite-hook parity invariant
    Given the adapter registry is initialised
    When I call PlanItems() on the codex adapter
    Then the items include a SyncItem of kind SyncKindPrewriteHook
    And that item's Path is ".codex/config.toml"

  # ── AC3: Init writes managed files ───────────────────────────────────────────

  Scenario: centinela init --agent codex writes managed .codex/config.toml
    Given a project with no Codex configuration
    When I run "centinela init --agent codex"
    Then ".codex/config.toml" is created
    And ".codex/config.toml" begins with the centinela managed-version header
    And ".codex/config.toml" contains a PreToolUse hook wiring centinela hook prewrite
    And ".codex/config.toml" contains a PostToolUse hook wiring centinela hook postwrite
    And ".codex/config.toml" contains UserPromptSubmit hook entries
    And "AGENTS.md" is created containing a Centinela managed region
    And ".claude/settings.json" is not created or modified

  # ── AC4: Idempotency / no drift ──────────────────────────────────────────────

  Scenario: init then migrate setup reports no pending drift
    Given a project where "centinela init --agent codex" has been run
    When I call BuildSyncPlan with agent "codex"
    Then the plan HasChanges() returns false
    And no item is marked as create or update

  Scenario: centinela init --agent codex is idempotent on re-run
    Given a project already initialised with "centinela init --agent codex"
    When I run "centinela init --agent codex" again
    Then ".codex/config.toml" is unchanged
    And "AGENTS.md" is unchanged
    And the exit code is 0

  # ── AC5: Unmanaged file protection ───────────────────────────────────────────

  Scenario: Pre-existing unmanaged .codex/config.toml is not clobbered
    Given a project with a hand-written ".codex/config.toml" containing no managed header
    When I run "centinela init --agent codex"
    Then ".codex/config.toml" is not overwritten
    And a manual-review warning is surfaced to the user

  # ── AC6: Golden parity ───────────────────────────────────────────────────────

  Scenario: Codex managed output matches golden fixture byte-for-byte
    Given the golden fixture at testdata/golden/codex/.codex/config.toml
    When BuildSyncPlan is called with agent "codex" and the plan is applied
    Then the emitted ".codex/config.toml" bytes match the golden fixture exactly
    And the emitted "AGENTS.md" bytes match the golden fixture exactly
