Feature: Typed evidence CLI
  As an LLM agent executing a Centinela workflow step
  I want a typed CLI for authoring and validating .workflow evidence
  So I never hand-write JSON via python/jq/heredoc

  Scenario: Init drops a schema-valid skeleton
    Given an active workflow "alpha" on step "plan"
    When I run "centinela evidence init alpha big-thinker"
    Then ".workflow/alpha-big-thinker.json" exists with all required fields
    And the file is pretty-printed with stable key order
    And "_meta.cli_version" matches the running binary

  Scenario: Set writes a scalar field atomically
    Given an initialized "alpha-big-thinker.json"
    When I run "centinela evidence set alpha big-thinker status done"
    Then the JSON reflects status="done"
    And no temp file remains in ".workflow/"

  Scenario: Append extends a list field without duplicating entries
    Given an initialized "alpha-big-thinker.json"
    When I run "centinela evidence append alpha big-thinker outputs docs/features/alpha.md" twice
    Then the "outputs" list contains "docs/features/alpha.md" exactly once

  Scenario: Read returns a single field for predecessor inspection
    Given a completed "alpha-feature-specialist.json" with outputs
    When I run "centinela evidence read alpha feature-specialist --field outputs"
    Then stdout is the JSON-encoded outputs list
    And exit code is 0

  Scenario: Validate exits non-zero with a fix hint on missing field
    Given "alpha-big-thinker.json" missing the "edgeCases" field
    When I run "centinela evidence validate alpha"
    Then exit code is non-zero
    And stderr contains "centinela evidence append alpha big-thinker edgeCases"

  Scenario: Atomic write survives a crash mid-append
    Given an initialized "alpha-big-thinker.json"
    When the writer is killed after temp-write but before rename
    Then the original JSON is unchanged
    And "centinela evidence repair alpha" removes the orphaned temp file

  Scenario: Concurrent writes serialize via advisory lock
    Given an initialized "alpha-big-thinker.json" with an empty "outputs" list
    When two agents simultaneously run "centinela evidence append alpha big-thinker outputs foo.md" and "centinela evidence append alpha big-thinker outputs bar.md"
    Then both commands exit 0
    And the resulting "outputs" contains both "foo.md" and "bar.md" exactly once each
    And the JSON on disk is valid (no interleaved or truncated content)

  Scenario: Init with unknown feature slug exits non-zero
    Given no active workflow named "ghost"
    When I run "centinela evidence init ghost big-thinker"
    Then exit code is non-zero
    And stderr names the unknown feature and lists the active features

  Scenario: Read against a not-yet-initialized role exits non-zero with a hint
    Given ".workflow/alpha-qa-senior.json" does not exist
    When I run "centinela evidence read alpha qa-senior --field outputs"
    Then exit code is non-zero
    And stderr suggests "centinela evidence init alpha qa-senior"

  Scenario: Artifact new with unknown kind exits non-zero
    Given an active workflow "alpha"
    When I run "centinela artifact new alpha bogus-kind"
    Then exit code is non-zero
    And stderr lists the allowed kinds: edge-cases, gatekeeper, production-readiness, documentation-specialist

  Scenario: Schema version skew preserves unknown fields
    Given an older binary wrote an "extra.legacy_field" into the JSON
    When a newer binary runs "centinela evidence validate alpha"
    Then validation passes
    And the unknown field is preserved on next round-trip

  Scenario: Free-form attachments use the extra slot
    Given an initialized "alpha-big-thinker.json"
    When I run "centinela evidence set alpha big-thinker extra.note 'reviewed by sam'"
    Then validation still passes
    And the value is stored under "extra.note"

  Scenario: Artifact templates drop pre-filled stubs
    Given an active workflow "alpha"
    When I run "centinela artifact new alpha edge-cases"
    Then ".workflow/alpha-edge-cases.md" exists with the templated sections
    And re-running the command without --force fails with exit non-zero

  Scenario: Postwrite hook reformats hand-written evidence
    Given the agent writes ".workflow/alpha-big-thinker.json" via the Write tool with minified JSON
    When the PostToolUse hook fires
    Then the file is rewritten with pretty-print and stable key order
    And other features' ".workflow/" files are untouched

  Scenario: Postwrite formatter is scoped to the active feature
    Given two active workflows "alpha" and "beta"
    And the worktree CWD belongs to "alpha"
    When ".workflow/beta-big-thinker.json" is modified by an unrelated tool
    Then the postwrite hook does NOT reformat it

  Scenario: Agent prompts forbid hand-written JSON
    Given the acceptance test scans every "docs/architecture/*-prompt.md"
    Then no prompt contains "python3 -c"
    And no prompt contains "<<EOF" with a ".workflow" path
    And every prompt references "centinela evidence" as the authoring path

  Scenario: Scaffold mirror parity covers prompts
    Given a prompt is edited in "docs/architecture/"
    When the scaffold parity test runs
    Then the matching file in "internal/scaffold/assets/docs/architecture/" must equal it
    Otherwise the test fails with the diff
