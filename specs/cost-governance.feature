Feature: Cost governance
  As a Centinela user running agents under a budget
  I want per-feature/per-step token spend surfaced against budgets
  So that a runaway step is visible without ever blocking my workflow

  Background:
    Given a repository governed by Centinela
    And an active feature with a [cost] budget configured

  Scenario: Token spend is captured from the harness transcript
    Given a host-harness transcript with assistant messages carrying token usage
    When the cost capture hook runs for the active feature and step
    Then a cost-sample telemetry event is recorded with the summed input and output tokens
    And the event is attributed to the active feature, step, and model

  Scenario: Repeated capture does not double-count
    Given a cost sample was already recorded up to a transcript cursor
    When the capture hook runs again after more messages
    Then only the new tokens since the cursor are added

  Scenario: Over-budget surfaces a soft warning, never blocks
    Given the active step's recorded spend exceeds its step_token_budget
    When I run "centinela validate"
    Then a non-failing over-budget warning is shown
    And the command still exits 0

  Scenario: Cost report shows spend versus budget
    Given recorded cost samples for the feature
    When I run "centinela cost"
    Then each feature/step row shows used tokens, budget, and remaining
    And rows that exceed budget are flagged

  Scenario: Disabled or unconfigured cost is a silent no-op
    Given [cost] is disabled or all budgets are zero
    When the capture hook runs and I run "centinela validate"
    Then no cost samples are recorded and no cost output appears

  Scenario: Missing or malformed transcript degrades gracefully
    Given no transcript_path is provided or the transcript is unreadable
    When the cost capture hook runs
    Then no cost sample is recorded
    And no error is raised
