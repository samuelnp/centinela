Feature: Headless governance — non-interactive umbrella and machine-readable verdict
  As an unattended runner (CI, the Capataz daemon, a Magallanes fleet)
  I want a single headless signal that silences human-aimed prompts
  And a deterministic JSON verdict packet with a pass/fail exit code
  So that governance is a first-class machine-consumable output, while a
  zero-config human session keeps byte-identical behavior.

  # Traceability: each scenario title is stable and maps 1:1 to a Go test via
  # a `// Scenario: <title>` comment. Headless resolution lives in
  # internal/config (config.IsHeadless); the two prompt hooks short-circuit on
  # it BEFORE the per-knob resolver, so headless wins over explicit knobs. The
  # verdict packet lives in internal/verdict (AssembleVerdict) and is surfaced
  # by the `centinela verdict <feature>` command.

  Background:
    Given a started feature "headless-governance" with an active workflow
    And the workflow is on the validate step
    And no CENTINELA_HEADLESS environment variable is set unless a scenario sets it
    And no [headless] config section is present unless a scenario sets one

  # ---------------------------------------------------------------------------
  # Headless umbrella — resolver precedence and hook suppression
  # ---------------------------------------------------------------------------

  Scenario: Headless via env suppresses the step-review prompt even under every_step
    Given the config sets step_confirmation_mode to every_step
    And the CENTINELA_HEADLESS environment variable is set to "1"
    When the review-prompt decision is made for the validate step
    Then no review-required prompt is rendered
    And headless wins over the explicit every_step knob

  Scenario: Headless via config suppresses the step-review prompt
    Given the config sets [headless] enabled to true
    And the config sets step_confirmation_mode to every_step
    When the review-prompt decision is made for the validate step
    Then no review-required prompt is rendered

  Scenario: Headless via config suppresses the plan-advisor directive
    Given the config sets [headless] enabled to true
    And the workflow is on the plan step
    When the plan-advisor hook runs
    Then it emits no advisor directive

  Scenario: Plan advisor stays quiet under headless even when it would otherwise speak
    Given the config sets [headless] enabled to true
    And a plan-step workflow whose feature brief would normally trigger advisor questions
    When the plan-advisor hook runs
    Then the hook short-circuits before loading workflows and emits nothing

  Scenario: CI auto-detect opt-in makes the run headless
    Given the config sets [headless] detect_ci to true
    And the CI environment variable is set to "true"
    When headless is resolved
    Then headless is active

  Scenario: CI present but detect_ci off is not headless (back-compat)
    Given the config sets [headless] detect_ci to false
    And the CI environment variable is set to "true"
    When headless is resolved
    Then headless is not active

  Scenario: Zero-config default is not headless
    Given no CENTINELA_HEADLESS environment variable is set
    And [headless] enabled is false
    And [headless] detect_ci is false
    When headless is resolved
    Then headless is not active

  Scenario: Env override beats config off
    Given the config sets [headless] enabled to false
    And the CENTINELA_HEADLESS environment variable is set to "1"
    When headless is resolved
    Then headless is active

  Scenario: Empty env value falls through to config and detect_ci
    Given the CENTINELA_HEADLESS environment variable is set to an empty string
    And the config sets [headless] enabled to false
    And the config sets [headless] detect_ci to false
    When headless is resolved
    Then headless is not active

  # ---------------------------------------------------------------------------
  # Back-compat — hooks unchanged when headless is off
  # ---------------------------------------------------------------------------

  Scenario: Back-compat review prompt under every_step still renders when headless off
    Given no headless signal of any kind is active
    And the config sets step_confirmation_mode to every_step
    When the review-prompt decision is made for the validate step
    Then a review-required prompt is rendered exactly as before

  Scenario: Back-compat plan advisor still emits directives when headless off
    Given no headless signal of any kind is active
    And the workflow is on the plan step with advisor-triggering content
    When the plan-advisor hook runs
    Then advisor directives are emitted exactly as before

  # ---------------------------------------------------------------------------
  # Verdict packet — assembly and determinism
  # ---------------------------------------------------------------------------

  Scenario: Verdict packet pass emits exit code zero and JSON to stdout
    Given all gates pass
    And verify reports no failures
    And a generatedAt timestamp is injected
    When the verdict is assembled and the command runs
    Then the packet summary verdict is "pass"
    And the packet summary exitCode is 0
    And the JSON packet is written to stdout
    And the command exits 0

  Scenario: Verdict packet fail still emits JSON to stdout with exit code one
    Given at least one gate reports Fail
    And a generatedAt timestamp is injected
    When the verdict is assembled and the command runs
    Then the packet summary verdict is "fail"
    And the packet summary exitCode is 1
    And the JSON packet is still written to stdout
    And the command exits 1

  Scenario: A verify failure alone produces a fail verdict
    Given all gates pass
    And verify reports HasFailures true
    When the verdict is assembled
    Then the packet summary verdict is "fail"
    And the packet summary exitCode is 1

  Scenario: Warnings are reported but do not fail the verdict
    Given all gates pass
    And verify reports a WARN check and no failures
    When the verdict is assembled
    Then the packet summary verdict is "pass"
    And the packet summary exitCode is 0
    And the warning is present in the packet counts

  Scenario: Verdict JSON is deterministic for fixed inputs and injected timestamp
    Given a fixed set of gate results, verify checks, and on-disk evidence
    And the same generatedAt timestamp is injected on two runs
    When the verdict is assembled and marshaled with indentation twice
    Then both runs produce byte-identical JSON matching the golden file
    And no maps are marshaled and evidence is sorted by role

  Scenario: Verdict run info snapshots workflow and config provenance
    Given a workflow on the validate step with a resolved profile and archetype
    And a driver model and resolved headless state
    When the verdict is assembled with an injected generatedAt
    Then the packet run carries feature, step, profile, archetype, driverModel, headless, and generatedAt

  Scenario: Verdict evidence index lists every on-disk role evidence for the feature
    Given on-disk evidence files .workflow/headless-governance-big-thinker.json and .workflow/headless-governance-feature-specialist.json
    When the verdict is assembled
    Then the evidence index has one entry per on-disk role file
    And each entry carries role, status, handoffTo, generatedAt, and path
    And the entries are sorted by role name

  Scenario: Gate statuses are lowercased and verify statuses stay uppercase
    Given a gate result with status Pass and a verify check with status PASS
    When the verdict is assembled
    Then the gate status in the packet is "pass"
    And the verify status in the packet is "PASS"

  Scenario: Packet schema field is the versioned identifier
    When the verdict is assembled
    Then the packet schema field is "centinela.verdict/v1"

  Scenario: Verdict always full-scans gates in v1
    Given a feature with both changed and unchanged source files
    When the verdict is assembled with a nil gate filter
    Then all gates run a full scan rather than a changed-only scan

  # ---------------------------------------------------------------------------
  # Verdict command — CLI surfaces and exit mechanism
  # ---------------------------------------------------------------------------

  Scenario: Verdict command separates JSON stdout from status stderr
    Given a passing verdict
    When the verdict command runs
    Then the JSON packet appears on stdout only
    And any human-readable status text appears on stderr only

  Scenario: Verdict command exits via a silenced sentinel error on fail
    Given a failing verdict
    When the verdict command runs with SilenceErrors and SilenceUsage set
    Then the JSON packet reaches stdout before the command returns
    And the command returns a sentinel error carrying exit code 1
    And no cobra usage text is printed

  Scenario: Verdict command surfaces the resolved headless state in the packet
    Given the CENTINELA_HEADLESS environment variable is set to "1"
    When the verdict command runs for the feature
    Then the packet run headless field is true
    And the JSON shape is otherwise unchanged by the headless flag

  Scenario: Verdict on a feature with no on-disk evidence yields an empty evidence index
    Given a feature with gates and verify results but no .workflow role JSON files
    When the verdict is assembled
    Then the evidence index is an empty list
    And the packet still emits valid JSON with a computed summary
