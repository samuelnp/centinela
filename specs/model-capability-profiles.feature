Feature: Model capability profiles
  A driver model's declared capability selects the default enforcement profile,
  so a frontier model is not taxed with strict ceremony and a weak local model
  gets maximum rails — without anyone passing a flag. The capability-derived
  default is a NEW lowest-priority precedence tier: an explicit --profile or an
  explicit global enforcement_profile always wins, and a project that declares no
  driver model resolves byte-identically to how it does today (strict default).

  Background:
    Given the built-in capability map covers the three Anthropic tier models:
      | model id          | class    | default profile |
      | claude-opus-4-7   | frontier | outcome         |
      | claude-sonnet-4-6 | capable  | guided          |
      | claude-haiku-4-5  | limited  | strict          |
    And the default class-to-profile mapping is frontier→outcome, capable→guided, limited→strict

  # ---------------------------------------------------------------------------
  # Precedence tier 4 — back-compat default (the load-bearing guarantee)
  # ---------------------------------------------------------------------------

  Scenario: Zero config resolves to strict byte-identically
    Given a feature started with no profile flag
    And no global enforcement_profile is configured
    And no driver model is configured
    When the effective profile is resolved
    Then the effective profile is strict

  # ---------------------------------------------------------------------------
  # Precedence tier 3 — capability default from the pinned driver model
  # ---------------------------------------------------------------------------

  Scenario: Frontier built-in driver model defaults to outcome
    Given the driver model is "claude-opus-4-7"
    And no global enforcement_profile is configured
    And no profile flag was passed
    When the effective profile is resolved
    Then the effective profile is outcome

  Scenario: Capable local driver model declared in config defaults to guided
    Given the orchestration capabilities config maps "local/some-capable" to class "capable"
    And the driver model is "local/some-capable"
    And no global enforcement_profile is configured
    And no profile flag was passed
    When the effective profile is resolved
    Then the effective profile is guided

  Scenario: Limited local driver model declared in config defaults to strict
    Given the orchestration capabilities config maps "local/weak-model" to class "limited"
    And the driver model is "local/weak-model"
    And no global enforcement_profile is configured
    And no profile flag was passed
    When the effective profile is resolved
    Then the effective profile is strict

  Scenario: Unknown driver model with no capability falls back to strict
    Given the driver model is "some/unknown-local-model"
    And the model has no built-in and no declared capability
    And no global enforcement_profile is configured
    And no profile flag was passed
    When the effective profile is resolved
    Then the effective profile is strict

  Scenario: Capability profiles override remaps a class to a different profile
    Given the orchestration capability_profiles config maps "frontier" to "guided"
    And the driver model is "claude-opus-4-7"
    And no global enforcement_profile is configured
    And no profile flag was passed
    When the effective profile is resolved
    Then the effective profile is guided

  # ---------------------------------------------------------------------------
  # Precedence tier 2 — explicit global enforcement_profile beats capability
  # ---------------------------------------------------------------------------

  Scenario: Explicit global enforcement_profile beats the capability default
    Given the global enforcement_profile is guided
    And the driver model is "claude-opus-4-7"
    And no profile flag was passed
    When the effective profile is resolved
    Then the effective profile is guided

  # ---------------------------------------------------------------------------
  # Precedence tier 1 — explicit --profile beats everything below
  # ---------------------------------------------------------------------------

  Scenario: Per-feature profile flag beats the capability default
    Given the feature was started with profile outcome
    And the global enforcement_profile is strict
    And the driver model is "claude-haiku-4-5"
    When the effective profile is resolved
    Then the effective profile is outcome

  # ---------------------------------------------------------------------------
  # Driver-model resolution precedence: flag > env > config
  # ---------------------------------------------------------------------------

  Scenario: Driver model flag overrides env overrides config
    Given the config sets driver_model to "config-model"
    And the CENTINELA_MODEL env var is "env-model"
    And the start command is given --model "flag-model"
    When the driver model is resolved at start
    Then the pinned driver model is "flag-model"

  Scenario: Driver model env overrides config when no flag is given
    Given the config sets driver_model to "config-model"
    And the CENTINELA_MODEL env var is "env-model"
    And the start command is given no --model flag
    When the driver model is resolved at start
    Then the pinned driver model is "env-model"

  Scenario: Driver model falls back to config when no flag and no env
    Given the config sets driver_model to "config-model"
    And the CENTINELA_MODEL env var is unset
    And the start command is given no --model flag
    When the driver model is resolved at start
    Then the pinned driver model is "config-model"

  Scenario: Driver model is empty when nothing is configured
    Given no driver_model config, no CENTINELA_MODEL env var, and no --model flag
    When the driver model is resolved at start
    Then the pinned driver model is empty

  # ---------------------------------------------------------------------------
  # Opaque model ids: --model accepts any string, normalization is lenient
  # ---------------------------------------------------------------------------

  Scenario: An opaque model id with no capability is accepted and pins without error
    Given the start command is given --model "totally/made-up-model"
    And the model has no built-in and no declared capability
    When the driver model is resolved at start
    Then the pinned driver model is "totally/made-up-model"
    And no error is raised

  Scenario: Capability class values are normalized by trim and lowercase
    Given the orchestration capabilities config maps "local/m" to class "  Frontier  "
    When the config is loaded
    Then loading succeeds
    And "local/m" resolves to capability class frontier

  # ---------------------------------------------------------------------------
  # Config validation (parse-time at config.Load)
  # ---------------------------------------------------------------------------

  Scenario: An unknown capability class value fails config load
    Given the orchestration capabilities config maps a model to class "genius"
    When the config is loaded
    Then loading fails with an error naming "genius"

  Scenario: An empty model id key in capabilities fails config load
    Given the orchestration capabilities config maps the empty model id to class "frontier"
    When the config is loaded
    Then loading fails with an error about the empty model id

  Scenario: An unknown class key in capability_profiles fails config load
    Given the orchestration capability_profiles config maps class "genius" to "guided"
    When the config is loaded
    Then loading fails with an error naming "genius"

  Scenario: An unknown profile value in capability_profiles fails config load
    Given the orchestration capability_profiles config maps "frontier" to "turbo"
    When the config is loaded
    Then loading fails with an error naming "turbo"

  Scenario: Absent capability tables are valid and change nothing
    Given no orchestration capabilities or capability_profiles tables are configured
    When the config is loaded
    Then loading succeeds
    And the effective profile for a zero-config feature is strict

  # ---------------------------------------------------------------------------
  # Status provenance (read-only presentation)
  # ---------------------------------------------------------------------------

  Scenario: Status shows the profile came from a frontier driver model
    Given a feature started with driver model "claude-opus-4-7" and no explicit profile
    When the status view is rendered
    Then the status shows "Profile  outcome  (driver: claude-opus-4-7 → frontier)"

  Scenario: Status shows strict default for an unknown driver model
    Given a feature started with driver model "some/unknown-local-model" and no explicit profile
    When the status view is rendered
    Then the status shows "Profile  strict  (driver: some/unknown-local-model → no capability, default strict)"

  Scenario: Status shows the global provenance when an explicit global profile wins
    Given a feature started with the global enforcement_profile guided and driver model "claude-opus-4-7"
    When the status view is rendered
    Then the status shows "Profile  guided  (global)"

  Scenario: Status shows the per-feature flag provenance when --profile was passed
    Given a feature started with profile outcome
    When the status view is rendered
    Then the status shows "Profile  outcome  (--profile)"

  Scenario: Status shows the strict default provenance for a zero-config feature
    Given a feature started with no profile flag, no global profile, and no driver model
    When the status view is rendered
    Then the status shows "Profile  strict  (default)"
