Feature: Capability calibration — per-model governance friction analysis
  As an operator managing a Centinela-governed project
  I want `centinela calibrate` to read telemetry events, compute per-model friction,
  and recommend tighter, looser, or unchanged enforcement profiles backed by evidence
  So that governance assignments are revisited against real behavior rather than intuition

  # centinela calibrate reads .workflow/telemetry/events.jsonl read-only.
  # Friction Rate = Rework / Advances; Rework = gate-failure + verify-rejection +
  # complete-rejected; Advances = step-advanced. HasRate = Advances > 0 (guarded).
  # Thresholds: highFrictionRate = 1.0, lowFrictionRate = 0.25, minAdvances = 3.
  # Profile strictness: strict(2) > guided(1) > outcome(0). tighter/looser clamp at ends.
  # Classification: !ok → Unclassified/None; Advances < 3 → WellCalibrated/Keep;
  # Rate >= 1.0 & tightenable → Undergoverned/Tighten; Rate <= 0.25 & loosenable →
  # Overgoverned/Loosen; else → WellCalibrated/Keep. Maxed extremes → WellCalibrated/Keep.
  # Models sorted by id ascending; "unattributed" forced last.
  # Scenario titles map 1:1 to Go acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with telemetry enabled
    And the telemetry log is at ".workflow/telemetry/events.jsonl"
    And the built-in capability map includes:
      | model id          | class    | default profile |
      | claude-opus-4-7   | frontier | outcome         |
      | claude-sonnet-4-6 | capable  | guided          |
      | claude-haiku-4-5  | limited  | strict          |

  # ---------------------------------------------------------------------------
  # Part 1 — Telemetry model stamping
  # ---------------------------------------------------------------------------

  Scenario: Event recorded during a workflow with a pinned DriverModel carries that model in the JSONL
    Given an active workflow for feature "alpha" pinned to driver model "claude-sonnet-4-6"
    When a governance event is recorded (e.g. a step-advanced)
    Then the event appended to the JSONL has model "claude-sonnet-4-6"

  Scenario: Event recorded with no driver model configured has an empty model field
    Given no driver model is configured (no workflow DriverModel, no env, no config)
    When a governance event is recorded
    Then the event appended to the JSONL has an empty model field or omits the model key

  Scenario: Legacy event without a model field parses cleanly and buckets as unattributed
    Given a telemetry log containing a valid event line with no "model" key
    When the log is read by the calibration reader
    Then the event unmarshals without error
    And the event's Model field is the empty string
    And calibration buckets the event under "unattributed"

  # ---------------------------------------------------------------------------
  # Part 2 — Calibration analysis: classification and recommendation
  # ---------------------------------------------------------------------------

  Scenario: Model with high friction under a tightenable profile is classified Undergoverned and recommended tighter profile
    Given a telemetry log where model "claude-sonnet-4-6" has:
      | event type        | count |
      | step-advanced     | 3     |
      | gate-failure      | 2     |
      | verify-rejection  | 1     |
    And "claude-sonnet-4-6" has class "capable" with current profile "guided"
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the record for "claude-sonnet-4-6" shows verdict "Undergoverned"
    And the record for "claude-sonnet-4-6" shows recommendation "Tighten"
    And the record for "claude-sonnet-4-6" shows recommended profile "strict"
    And the record for "claude-sonnet-4-6" cites Advances=3, Rework=3, Rate=1.00

  Scenario: Model with low friction under a loosenable profile is classified Overgoverned and recommended looser profile
    Given a telemetry log where model "claude-haiku-4-5" has:
      | event type    | count |
      | step-advanced | 4     |
      | gate-failure  | 1     |
    And "claude-haiku-4-5" has class "limited" with current profile "strict"
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the record for "claude-haiku-4-5" shows verdict "Overgoverned"
    And the record for "claude-haiku-4-5" shows recommendation "Loosen"
    And the record for "claude-haiku-4-5" shows recommended profile "guided"
    And the record for "claude-haiku-4-5" cites Advances=4, Rework=1, Rate=0.25

  Scenario: Model already at the strictest profile but high friction is classified WellCalibrated with recommendation Keep
    Given a telemetry log where model "claude-haiku-4-5" has:
      | event type        | count |
      | step-advanced     | 3     |
      | gate-failure      | 3     |
    And "claude-haiku-4-5" has class "limited" with current profile "strict"
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the record for "claude-haiku-4-5" shows verdict "WellCalibrated"
    And the record for "claude-haiku-4-5" shows recommendation "Keep"
    And the record for "claude-haiku-4-5" shows recommended profile "strict"

  Scenario: Model already at the loosest profile but low friction is classified WellCalibrated with recommendation Keep
    Given a telemetry log where model "claude-opus-4-7" has:
      | event type    | count |
      | step-advanced | 5     |
      | gate-failure  | 1     |
    And "claude-opus-4-7" has class "frontier" with current profile "outcome"
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the record for "claude-opus-4-7" shows verdict "WellCalibrated"
    And the record for "claude-opus-4-7" shows recommendation "Keep"
    And the record for "claude-opus-4-7" shows recommended profile "outcome"

  Scenario: Model with friction between thresholds is classified WellCalibrated and recommended Keep
    Given a telemetry log where model "claude-sonnet-4-6" has:
      | event type    | count |
      | step-advanced | 4     |
      | gate-failure  | 2     |
    And "claude-sonnet-4-6" has class "capable" with current profile "guided"
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the record for "claude-sonnet-4-6" shows verdict "WellCalibrated"
    And the record for "claude-sonnet-4-6" shows recommendation "Keep"
    And the record for "claude-sonnet-4-6" cites Advances=4, Rework=2, Rate=0.50

  Scenario: Model with fewer than 3 advances is classified WellCalibrated due to insufficient evidence regardless of rate
    Given a telemetry log where model "claude-haiku-4-5" has:
      | event type    | count |
      | step-advanced | 2     |
      | gate-failure  | 5     |
    And "claude-haiku-4-5" has class "limited" with current profile "strict"
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the record for "claude-haiku-4-5" shows verdict "WellCalibrated"
    And the record for "claude-haiku-4-5" shows recommendation "Keep"

  Scenario: Model with zero step-advanced events is guarded against division-by-zero and classified WellCalibrated
    Given a telemetry log where model "claude-haiku-4-5" has:
      | event type   | count |
      | gate-failure | 4     |
    And "claude-haiku-4-5" has class "limited" with current profile "strict"
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the record for "claude-haiku-4-5" shows Advances=0
    And the record for "claude-haiku-4-5" shows HasRate=false
    And the record for "claude-haiku-4-5" shows verdict "WellCalibrated"
    And no division-by-zero panic or NaN is produced

  Scenario: Model id with no capability class is classified Unclassified with no recommendation
    Given a telemetry log where model "local/unknown-model" has:
      | event type    | count |
      | step-advanced | 5     |
      | gate-failure  | 5     |
    And "local/unknown-model" has no declared capability class
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the record for "local/unknown-model" shows verdict "Unclassified"
    And the record for "local/unknown-model" shows recommendation "None"
    And no error or panic is produced

  Scenario: Unattributed bucket from events with no model is classified Unclassified and rendered last
    Given a telemetry log containing events with no model field:
      | event type    | count |
      | step-advanced | 5     |
      | gate-failure  | 5     |
    And the log also contains events for model "claude-sonnet-4-6" with 3 step-advanced and 3 gate-failure events
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the record for "unattributed" shows verdict "Unclassified"
    And "unattributed" appears after "claude-sonnet-4-6" in the output

  # ---------------------------------------------------------------------------
  # Threshold boundary inclusivity (exact edge cases)
  # ---------------------------------------------------------------------------

  Scenario: Rate exactly equal to highFrictionRate (1.0) triggers Undergoverned classification
    Given a telemetry log where model "claude-sonnet-4-6" has:
      | event type    | count |
      | step-advanced | 3     |
      | gate-failure  | 3     |
    And "claude-sonnet-4-6" has class "capable" with current profile "guided"
    When the operator runs:
      centinela calibrate
    Then the record for "claude-sonnet-4-6" shows verdict "Undergoverned"
    And the record for "claude-sonnet-4-6" shows recommendation "Tighten"

  Scenario: Rate exactly equal to lowFrictionRate (0.25) triggers Overgoverned classification
    Given a telemetry log where model "claude-haiku-4-5" has:
      | event type    | count |
      | step-advanced | 4     |
      | gate-failure  | 1     |
    And "claude-haiku-4-5" has class "limited" with current profile "strict"
    When the operator runs:
      centinela calibrate
    Then the record for "claude-haiku-4-5" shows verdict "Overgoverned"
    And the record for "claude-haiku-4-5" shows recommendation "Loosen"

  Scenario: Model with advances and zero rework has Rate 0.0 and is classified Overgoverned if loosenable
    Given a telemetry log where model "claude-haiku-4-5" has:
      | event type    | count |
      | step-advanced | 5     |
    And "claude-haiku-4-5" has class "limited" with current profile "strict"
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the record for "claude-haiku-4-5" shows Advances=5, Rework=0, Rate=0.00
    And the record for "claude-haiku-4-5" shows HasRate=true
    And the record for "claude-haiku-4-5" shows verdict "Overgoverned"
    And the record for "claude-haiku-4-5" shows recommendation "Loosen"

  Scenario: Model with only rework events and zero advances is WellCalibrated not Undergoverned
    Given a telemetry log where model "claude-sonnet-4-6" has:
      | event type   | count |
      | gate-failure | 10    |
    And "claude-sonnet-4-6" has class "capable" with current profile "guided"
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the record for "claude-sonnet-4-6" shows Advances=0, HasRate=false
    And the record for "claude-sonnet-4-6" shows verdict "WellCalibrated"
    And the record for "claude-sonnet-4-6" shows recommendation "Keep"

  # ---------------------------------------------------------------------------
  # Behavior / robustness
  # ---------------------------------------------------------------------------

  Scenario: Missing telemetry log prints clean empty-state report and exits 0
    Given no telemetry log file exists
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the output contains "no telemetry yet"
    And the output does not contain an error message or stack trace

  Scenario: Empty telemetry log prints clean empty-state report and exits 0
    Given the telemetry log exists but contains no events
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the output contains "no telemetry yet"

  Scenario: Malformed JSONL lines are skipped and valid events are still aggregated
    Given a telemetry log containing:
      | line                                        |
      | a valid step-advanced event for "claude-sonnet-4-6" |
      | {not valid json                             |
      | another valid step-advanced event for "claude-sonnet-4-6" |
    And "claude-sonnet-4-6" has class "capable" with current profile "guided"
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the record for "claude-sonnet-4-6" shows Advances=2
    And the output does not contain a parse error

  Scenario: --json emits structured Report as indented JSON and exits 0
    Given a telemetry log containing events for model "claude-haiku-4-5" with 3 step-advanced and 3 gate-failure events
    And "claude-haiku-4-5" has class "limited" with current profile "strict"
    When the operator runs:
      centinela calibrate --json
    Then the command exits with code 0
    And the output is valid JSON
    And the JSON contains top-level fields "ModelCount", "SpanStart", "SpanEnd", "Models"
    And each model entry contains fields "Model", "Class", "CurrentProfile", "Friction", "Recommendation", "RecommendedProfile", "Verdict"
    And the output contains no ANSI escape sequences

  Scenario: --json on empty log emits a valid JSON Report with zero models and exits 0
    Given no telemetry log file exists
    When the operator runs:
      centinela calibrate --json
    Then the command exits with code 0
    And the output is valid JSON
    And the JSON field "ModelCount" is 0
    And the JSON field "Models" is an empty array

  Scenario: Two runs on the same log produce byte-identical output
    Given a telemetry log containing a fixed set of events for multiple models
    When the operator runs centinela calibrate twice in succession
    Then both outputs are byte-identical

  Scenario: Two --json runs on the same log produce byte-identical JSON output
    Given a telemetry log containing a fixed set of events for multiple models
    When the operator runs centinela calibrate --json twice in succession
    Then both outputs are byte-identical

  Scenario: Models are sorted by id ascending with unattributed forced last
    Given a telemetry log containing events for:
      | model                   | step-advanced | gate-failure |
      | claude-sonnet-4-6       | 3             | 1            |
      | claude-haiku-4-5        | 3             | 1            |
      | (no model / unattributed) | 3           | 1            |
    When the operator runs:
      centinela calibrate
    Then "claude-haiku-4-5" appears before "claude-sonnet-4-6" in the output
    And "unattributed" appears after all other model records

  Scenario: Non-TTY piped output contains no ANSI escape sequences
    Given a telemetry log containing at least one classifiable event
    When the operator runs centinela calibrate with stdout piped to a file
    Then the output file contains no ANSI escape sequences

  Scenario: Multiple models in one log each receive an independent classification in a single pass
    Given a telemetry log containing:
      | model             | step-advanced | gate-failure |
      | claude-opus-4-7   | 4             | 1            |
      | claude-sonnet-4-6 | 3             | 3            |
      | claude-haiku-4-5  | 2             | 1            |
    When the operator runs:
      centinela calibrate
    Then the command exits with code 0
    And the record for "claude-opus-4-7" shows verdict "WellCalibrated"
    And the record for "claude-sonnet-4-6" shows verdict "Undergoverned"
    And the record for "claude-haiku-4-5" shows verdict "WellCalibrated"

  Scenario: Model id tie-breaking sorts by model id ascending for fully stable ordering
    Given a telemetry log containing events for models "zeta-model" and "alpha-model" each with equal friction
    And both models are mapped to the same capability class and current profile
    When the operator runs:
      centinela calibrate
    Then "alpha-model" appears before "zeta-model" in the output
