Feature: Local harness support
  A local-model operator can point Centinela at a local backend and get correct
  governance by default. Declaring an [orchestration.local] block (1) wires a
  managed OpenCode provider at the local endpoint and (2) defaults the declared
  local driver model to the limited capability → strict profile with explicit
  provenance. Every new tier is the LOWEST precedence and is gated on a non-empty
  local block: an explicit --profile or global enforcement_profile still wins, and
  a config with no local block resolves byte-for-byte identically to pre-feature
  (managed opencode output and capability/profile precedence both unchanged). All
  validation is shape-only — endpoint/model/api_key_env are opaque strings whose
  availability is the runner's job, never verified at config load.

  Background:
    Given the allowed local providers are "ollama" and "openai-compatible"
    And both provider kinds emit an OpenCode provider block using npm "@ai-sdk/openai-compatible"
    And the local driver-model tier sits BELOW [orchestration] driver_model and below explicit capability mappings

  # ---------------------------------------------------------------------------
  # AC#1 — Ollama provider wired into managed opencode.json, other harnesses untouched
  # ---------------------------------------------------------------------------

  Scenario: An Ollama local block wires a managed OpenCode provider at the endpoint
    Given the config declares [orchestration.local] with provider "ollama", endpoint "http://localhost:11434/v1" and model "qwen2.5-coder"
    When the OpenCode adapter plans setup
    Then opencode.json gains a managed provider block keyed "ollama"
    And the provider block's options.baseURL is "http://localhost:11434/v1"
    And the provider block's models contains "qwen2.5-coder"
    And ".claude/settings.json" is not created or modified
    And ".aider.conf.yml" is not created or modified

  # ---------------------------------------------------------------------------
  # AC#2 — Generic openai-compatible provider block
  # ---------------------------------------------------------------------------

  Scenario: A generic openai-compatible local block writes an openai-compatible provider
    Given the config declares [orchestration.local] with provider "openai-compatible", endpoint "http://localhost:8000/v1", model "llama-3.1-8b" and api_key_env "LOCAL_API_KEY"
    When the OpenCode adapter plans setup
    Then opencode.json gains a managed provider block keyed "openai-compatible"
    And the provider block uses npm "@ai-sdk/openai-compatible"
    And the provider block's options.baseURL is "http://localhost:8000/v1"
    And the provider block's options.apiKey references the env var "LOCAL_API_KEY"
    And the provider block's models contains "llama-3.1-8b"

  # ---------------------------------------------------------------------------
  # AC#3 — Declared local model defaults to limited → strict with provenance
  # ---------------------------------------------------------------------------

  Scenario: A declared local model with no capability class defaults to limited then strict
    Given the config declares [orchestration.local] with provider "ollama", endpoint "http://localhost:11434/v1" and model "qwen2.5-coder"
    And the model "qwen2.5-coder" has no built-in and no explicit capability class
    And no global enforcement_profile is configured
    And no profile flag was passed
    When the effective profile is resolved
    Then the effective profile is strict
    And the resolved capability class is limited

  Scenario: Status attributes the strict profile to the local default
    Given a feature started with the local model "qwen2.5-coder" declared and no explicit profile
    When the status view is rendered
    Then the status shows "Profile  strict  (local default: qwen2.5-coder → limited → strict)"

  # ---------------------------------------------------------------------------
  # AC#4 — Explicit --profile / global enforcement_profile still wins (back-compat invariant)
  # ---------------------------------------------------------------------------

  Scenario: An explicit global enforcement_profile beats the local default
    Given the config declares [orchestration.local] with provider "ollama", endpoint "http://localhost:11434/v1" and model "qwen2.5-coder"
    And the global enforcement_profile is guided
    And no profile flag was passed
    When the effective profile is resolved
    Then the effective profile is guided

  Scenario: A per-feature --profile flag beats the local default
    Given the config declares [orchestration.local] with provider "ollama", endpoint "http://localhost:11434/v1" and model "qwen2.5-coder"
    And the feature was started with profile outcome
    When the effective profile is resolved
    Then the effective profile is outcome

  # ---------------------------------------------------------------------------
  # AC#5 — No local block = zero-config, byte-identical to pre-feature
  # ---------------------------------------------------------------------------

  Scenario: A config with no local block resolves byte-identically to pre-feature
    Given a config with no [orchestration.local] block
    And no global enforcement_profile is configured
    And no profile flag was passed
    When the effective profile is resolved
    Then the effective profile is strict
    And no managed provider block is emitted into opencode.json
    And the managed opencode.json output is byte-for-byte identical to the pre-feature golden snapshot

  # ---------------------------------------------------------------------------
  # AC#6 — Malformed local config fails loudly naming the offending key
  # ---------------------------------------------------------------------------

  Scenario Outline: A malformed [orchestration.local] block fails config load naming the key
    Given the config declares [orchestration.local] with provider "<provider>", endpoint "<endpoint>" and model "<model>"
    When the config is loaded
    Then loading fails with an error naming "<key>"

    Examples:
      | provider          | endpoint                     | model         | key      |
      | groq              | http://localhost:11434/v1    | qwen2.5-coder | provider |
      | ollama            |                              | qwen2.5-coder | endpoint |
      | ollama            | http://localhost:11434/v1    |               | model    |

  Scenario: The unknown-provider error lists the allowed providers
    Given the config declares [orchestration.local] with provider "groq", endpoint "http://localhost:11434/v1" and model "qwen2.5-coder"
    When the config is loaded
    Then loading fails with an error listing "ollama" and "openai-compatible"

  # ---------------------------------------------------------------------------
  # AC#7 — Re-running init/migrate is idempotent (managed markers respected)
  # ---------------------------------------------------------------------------

  Scenario: Re-running init with the same local config rewrites nothing
    Given a project already initialised with an "ollama" local block wired into opencode.json
    When I run "centinela init --agent opencode" again with the same local config
    Then the managed provider block in opencode.json is unchanged
    And the apply reports no change for opencode.json
    And the exit code is 0

  Scenario: The provider block is only rewritten on a real change
    Given a project already initialised with an "ollama" local block at endpoint "http://localhost:11434/v1"
    When the local endpoint changes to "http://localhost:11500/v1" and migrate runs
    Then the managed provider block's options.baseURL becomes "http://localhost:11500/v1"
    And the apply reports a change for opencode.json

  # ---------------------------------------------------------------------------
  # AC#8 — Hermetic end-to-end governed run under strict (THE ACCEPTANCE BAR)
  # ---------------------------------------------------------------------------

  @acceptance @e2e @hermetic
  Scenario: A governed run under strict with the local provider wired passes gates and claim verification
    Given a hermetic project with an "ollama" local block wired to a local stub backend
    And the driver model resolves to the declared local model at the limited → strict default
    And no real network call is made to any local server
    When a feature is taken end-to-end through plan, code, tests, validate and docs under the strict profile
    Then all gate checks pass
    And claim verification passes
    And the wired opencode.json provider points at the configured stub baseURL

  # ---------------------------------------------------------------------------
  # Edge cases
  # ---------------------------------------------------------------------------

  Scenario: Provider value is normalized by trim and lowercase before validation
    Given the config declares [orchestration.local] with provider "  Ollama  ", endpoint "http://localhost:11434/v1" and model "qwen2.5-coder"
    When the config is loaded
    Then loading succeeds
    And the resolved local provider is "ollama"

  Scenario: Endpoint and model opaque strings are trimmed but never existence-checked
    Given the config declares [orchestration.local] with provider "ollama", endpoint "  http://localhost:11434/v1  " and model "  qwen2.5-coder  "
    When the config is loaded
    Then loading succeeds
    And the resolved local endpoint is "http://localhost:11434/v1"
    And the resolved local model is "qwen2.5-coder"
    And no reachability check is performed against the endpoint

  Scenario: An api_key_env naming a missing environment variable still loads
    Given the config declares [orchestration.local] with provider "openai-compatible", endpoint "http://localhost:8000/v1", model "llama-3.1-8b" and api_key_env "DEFINITELY_UNSET_VAR"
    And the environment variable "DEFINITELY_UNSET_VAR" is unset
    When the config is loaded
    Then loading succeeds
    And the provider block's options.apiKey references the env var "DEFINITELY_UNSET_VAR"

  Scenario: A user's hand-written provider in opencode.json is not clobbered
    Given opencode.json already contains a hand-written provider keyed "my-custom-provider"
    And the config declares [orchestration.local] with provider "ollama", endpoint "http://localhost:11434/v1" and model "qwen2.5-coder"
    When the OpenCode adapter plans and applies setup
    Then the hand-written provider "my-custom-provider" is preserved unchanged
    And a managed provider block keyed "ollama" is added alongside it

  Scenario: A pre-existing provider under the managed key is not overwritten
    Given opencode.json already contains a provider keyed "ollama" not written by Centinela
    And the config declares [orchestration.local] with provider "ollama", endpoint "http://localhost:11434/v1" and model "qwen2.5-coder"
    When the OpenCode adapter plans and applies setup
    Then the existing "ollama" provider is not overwritten
    And the apply reports no change for opencode.json

  Scenario: An explicitly mapped local model id wins over the local default
    Given the config declares [orchestration.local] with provider "ollama", endpoint "http://localhost:11434/v1" and model "qwen2.5-coder"
    And the orchestration capabilities config maps "qwen2.5-coder" to class "capable"
    And no global enforcement_profile is configured
    And no profile flag was passed
    When the effective profile is resolved
    Then the resolved capability class is capable
    And the effective profile is guided

  Scenario: No local block engages no capability tier and emits no provider
    Given a config with no [orchestration.local] block
    When the OpenCode adapter plans setup
    Then no managed provider block is emitted into opencode.json
    And the local capability default tier is not engaged
