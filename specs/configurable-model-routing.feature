Feature: Configurable model routing
  As a developer running Centinela on opencode, codex, or Claude Code
  I want to remap tiers and override roles to concrete model IDs in centinela.toml
  So that I can route each subagent to the best model for my provider and budget

  # AC1 — tier remap: model_map entry is used for the matching runner
  Scenario: Tier remap resolves the correct model for the active runner
    Given a centinela.toml with '[orchestration.model_map.reasoning]' containing 'opencode = "moonshotai/kimi-k2"'
    And the active runner is "opencode"
    When the orchestration hook emits the plan-step directive
    Then the directive contains an annotation for 'big-thinker' with model 'moonshotai/kimi-k2'

  # AC2 — role override: direct runner→model table beats the role's tier
  Scenario: Role override beats the role's tier for the active runner
    Given a centinela.toml with '[orchestration.models]' containing 'senior-engineer = { opencode = "deepseek/deepseek-coder" }'
    And the active runner is "opencode"
    When the orchestration hook emits the plan-step directive
    Then the directive contains an annotation for 'senior-engineer' with model 'deepseek/deepseek-coder'
    And the annotation for 'senior-engineer' does not fall back to the reasoning tier model

  # AC3 — no model_map entry for runner: falls back to built-in default
  Scenario: Role with a tier override but no model_map entry for the runner uses the built-in default
    Given a centinela.toml with '[orchestration.model_map.balanced]' containing 'opencode = "deepseek/deepseek-chat"'
    And the role 'feature-specialist' resolves to the 'balanced' tier
    And the active runner is "claude"
    When the orchestration hook emits the plan-step directive
    Then the directive contains an annotation for 'feature-specialist' with model 'claude-sonnet-4-6'

  # AC4 — back-compat: plain tier string in [orchestration.models] still works
  Scenario: Plain tier string value in orchestration.models is accepted and behaves as before
    Given a centinela.toml with '[orchestration.models]' containing 'qa-senior = "balanced"'
    When Centinela loads the config
    Then config loading succeeds
    And the resolved tier for 'qa-senior' is 'balanced'

  # AC5 — malformed config fails loudly at load time
  Scenario: Unknown runner key in model_map is rejected at config load time
    Given a centinela.toml with '[orchestration.model_map.reasoning]' containing 'gemini = "gemini-pro"'
    When Centinela loads the config
    Then config loading fails
    And the error message names the offending key 'gemini'

  Scenario: Unknown role key in orchestration.models is rejected at config load time
    Given a centinela.toml with '[orchestration.models]' containing 'backend-wizard = { opencode = "some-model" }'
    When Centinela loads the config
    Then config loading fails
    And the error message names the offending key 'backend-wizard'

  Scenario: Unknown tier key in model_map is rejected at config load time
    Given a centinela.toml with '[orchestration.model_map.turbo]' section present
    When Centinela loads the config
    Then config loading fails
    And the error message names the offending key 'turbo'

  Scenario: Empty model string in model_map is rejected at config load time
    Given a centinela.toml with '[orchestration.model_map.reasoning]' containing 'opencode = ""'
    When Centinela loads the config
    Then config loading fails
    And the error message names the offending key that has an empty model string

  # AC6 — absent tables: all roles resolve to built-in defaults (zero-config safe)
  Scenario: Absent model_map and models tables — all roles resolve to built-in defaults
    Given a centinela.toml with no '[orchestration.model_map]' section
    And a centinela.toml with no '[orchestration.models]' section
    When the orchestration hook emits the plan-step directive
    Then the directive contains an annotation for 'big-thinker' with model 'claude-opus-4-7' for the 'claude' runner
    And the directive contains an annotation for 'feature-specialist' with model 'claude-sonnet-4-6' for the 'claude' runner
    And the directive contains an annotation for 'documentation-specialist' with model 'claude-haiku-4-5-20251001' for the 'claude' runner

  # AC7 — no mapping for active runner: emits tier name + warning, never another runner's ID
  Scenario: Active runner with no mapping emits tier name and warning instead of another runner's concrete ID
    Given a centinela.toml with '[orchestration.model_map.reasoning]' containing 'opencode = "moonshotai/kimi-k2"'
    And the active runner is "codex"
    When the orchestration hook emits the plan-step directive
    Then the directive annotation for a reasoning-tier role carries the tier name 'reasoning'
    And a warning is surfaced indicating the runner has no concrete model for the 'reasoning' tier
    And the directive does not contain 'moonshotai/kimi-k2' in the codex column

  # Edge case — casing and whitespace normalization on tier and runner keys
  Scenario: Tier key with uppercase and surrounding whitespace is normalized and accepted
    Given a centinela.toml with a model_map entry for tier key ' Reasoning '
    When Centinela loads the config
    Then config loading succeeds
    And the tier key is normalized to 'reasoning' before validation

  Scenario: Runner key with uppercase and surrounding whitespace is normalized and accepted
    Given a centinela.toml with '[orchestration.model_map.reasoning]' containing ' Opencode  = "some-model"'
    When Centinela loads the config
    Then config loading succeeds
    And the runner key is normalized to 'opencode' before validation

  # Edge case — mixed forms in [orchestration.models]: one role a tier string, another a runner table
  Scenario: Mixed role value forms in orchestration.models are both valid
    Given a centinela.toml with '[orchestration.models]' containing 'qa-senior = "balanced"' and 'senior-engineer = { opencode = "deepseek/deepseek-coder" }'
    When Centinela loads the config
    Then config loading succeeds
    And the resolved tier for 'qa-senior' is 'balanced'
    And the resolved override model for 'senior-engineer' under 'opencode' is 'deepseek/deepseek-coder'

  # Edge case — role-override beats its own tier (precedence rule 1 over rule 2)
  Scenario: Role-level concrete override wins over a model_map entry for the same runner
    Given a centinela.toml with '[orchestration.model_map.reasoning]' containing 'opencode = "moonshotai/kimi-k2"'
    And a centinela.toml with '[orchestration.models]' containing 'big-thinker = { opencode = "deepseek/deepseek-coder" }'
    And the active runner is "opencode"
    When the orchestration hook emits the plan-step directive
    Then the directive contains an annotation for 'big-thinker' with model 'deepseek/deepseek-coder'
    And the directive does not contain 'moonshotai/kimi-k2' as the model for 'big-thinker'

  # Edge case — codex runner before codex-support lands: rule-4 fallback, never a wrong-vendor ID
  Scenario: Codex runner before codex-support lands falls back to tier name with ok=false
    Given a centinela.toml with no explicit codex entries in model_map or models
    And the codex runner has no built-in default model IDs (pre-codex-support)
    When the resolver is called with runner "codex" for any role
    Then the resolver returns the tier name for each role
    And ok is false for each resolved role
    And the returned model ID is never a Claude or opencode concrete model ID

  # Edge case — empty tables behave like absent tables
  Scenario: Empty model_map and models tables behave identically to absent tables
    Given a centinela.toml with empty '[orchestration.model_map]' and '[orchestration.models]' sections
    When the orchestration hook emits the plan-step directive
    Then every step role resolves to its built-in default model
