Feature: Enforcement profiles
  Named strictness presets scale how much process is enforced while gates
  and claim verification stay constant, so Centinela fits any model from a
  small local one to a frontier model without being switched off.

  Scenario: Outcome profile allows writing code during the plan step
    Given a feature started with the outcome profile
    And the workflow is on the plan step
    When a source code file write is evaluated by the prewrite hook
    Then the write is allowed

  Scenario: Strict and guided profiles still block out-of-step writes
    Given a feature started with the strict profile
    And the workflow is on the plan step
    When a source code file write is evaluated by the prewrite hook
    Then the write is blocked

  Scenario: A write with no active workflow is always blocked
    Given no active workflow
    When a plan or code file write is evaluated by the prewrite hook
    Then the write is blocked regardless of profile

  Scenario: Outcome profile suppresses the stop-and-ask review prompt
    Given a feature using the outcome profile
    When the review-prompt decision is made for any step
    Then no review prompt is rendered

  Scenario: An explicit confirmation mode overrides the profile default
    Given the configuration sets step_confirmation_mode to every_step
    And the enforcement profile is outcome
    When the review-prompt decision is made
    Then the explicit every_step setting wins and a prompt is rendered

  Scenario: Strict profile requires subagent orchestration evidence
    Given a feature started with the strict profile
    When the feature is created
    Then its orchestration mode requires subagent evidence

  Scenario: Guided profile does not require subagent orchestration evidence
    Given a feature started with the guided profile
    When the feature is created
    Then its orchestration mode does not require subagent evidence

  Scenario: Outcome profile does not require subagent orchestration evidence
    Given a feature started with the outcome profile
    When the feature is created
    Then its orchestration mode does not require subagent evidence

  Scenario: A per-feature profile overrides the global setting
    Given the global enforcement profile is guided
    And a feature was started with the outcome profile override
    When the effective profile for that feature is resolved
    Then the effective profile is outcome

  Scenario: An unconfigured project keeps today's behavior (default strict)
    Given a project that sets neither an enforcement profile nor confirmation mode
    When the effective enforcement knobs are resolved
    Then the effective profile is strict
    And step-gating is on
    And the confirmation mode is every_step
    And subagent orchestration evidence is required

  Scenario: Gates and claim verification run under every profile
    Given a feature on the validate step with a failing gate or claim
    When completion is attempted under the strict, guided, or outcome profile
    Then completion is blocked by the gate and claim verification in every case

  Scenario: An unknown profile value is rejected at config load
    Given a centinela.toml whose enforcement_profile is an unsupported value
    When the configuration is loaded
    Then loading fails with an error naming the enforcement_profile field
