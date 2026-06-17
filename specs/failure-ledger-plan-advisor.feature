Feature: failure-ledger plan advisor — feed recurring gate failures forward into planning
  As an engineer starting a feature, a governance owner, or a maintainer of a clean repo
  I want the plan advisor to read the governance-telemetry ledger during the plan step
  and surface the gates that have recently bitten this repo as a context line and a
  pre-warning question
  So that the failure modes the loop already recorded are prevented up front in the plan,
  not rediscovered at the validate gate

  # The plan advisor (internal/planadvisor, Directive(feature, cfg)) runs automatically
  # during the plan step. This feature makes it consult the telemetry ledger at
  # .workflow/telemetry/events.jsonl read-only, aggregate gate-failure events with the
  # SAME counting logic as `centinela insights` (count desc, then gate name asc, empty
  # Gate → "<none>"), and surface the top-N recurring gates as:
  #   (a) a "Recurring gate failures" line in the advisor's "Relevant context:" block, and
  #   (b) a pre-warning planning question naming the worst gate(s), subject to the existing
  #       plan_question_limit cap and lens tagging.
  # The clean-repo guarantee is absolute: missing/empty ledger, no gate-failure events, or
  # [telemetry] enabled = false → advisor output is BYTE-IDENTICAL to today's. The advisor
  # never writes the ledger and only acts during the plan step (headless / other steps
  # already exit unchanged). Scenario titles map 1:1 to Go acceptance tests
  # (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with a valid centinela.toml
    And the active workflow step is "plan"
    And the telemetry ledger is at ".workflow/telemetry/events.jsonl"
    And "[telemetry] enabled" is true unless a scenario states otherwise

  # ---------------------------------------------------------------------------
  # Clean-repo guarantee — byte-identical to today when there is nothing to warn about
  # ---------------------------------------------------------------------------

  Scenario: Missing ledger file leaves advisor output byte-identical to today
    Given no telemetry ledger file exists
    When the plan advisor builds its directive for a feature
    Then the directive contains no "Recurring gate failures" line
    And the directive adds no gate-failure pre-warning question
    And the directive is byte-identical to the advisor output produced with the ledger feature disabled
    And no error message or stack trace is emitted

  Scenario: Empty ledger file leaves advisor output byte-identical to today
    Given the telemetry ledger exists but contains no events
    When the plan advisor builds its directive for a feature
    Then the directive contains no "Recurring gate failures" line
    And the directive adds no gate-failure pre-warning question

  Scenario: Ledger with only block and step-advanced events produces no recurring-failure output
    Given a telemetry ledger containing:
      | type          | reason      | fileType | feature |
      | block         | out-of-step | plan     | alpha   |
      | block         | need-init   | source   | alpha   |
      | step-advanced |             |          | alpha   |
    When the plan advisor builds its directive for a feature
    Then the directive contains no "Recurring gate failures" line
    And the directive adds no gate-failure pre-warning question

  Scenario: Telemetry disabled in config suppresses all ledger-derived failure context
    Given "[telemetry] enabled" is false
    And a telemetry ledger containing several gate-failure events for gate "g1-file-size"
    When the plan advisor builds its directive for a feature
    Then the advisor does not read the ledger
    And the directive contains no "Recurring gate failures" line
    And the directive adds no gate-failure pre-warning question
    And the directive is byte-identical to the advisor output produced on a missing ledger

  # ---------------------------------------------------------------------------
  # Recurring gate failures — context summary line
  # ---------------------------------------------------------------------------

  Scenario: Recurring gate failures appear in the context summary ranked by count descending
    Given a telemetry ledger containing:
      | type         | gate         |
      | gate-failure | g1-file-size |
      | gate-failure | g1-file-size |
      | gate-failure | g1-file-size |
      | gate-failure | coverage     |
      | gate-failure | coverage     |
      | gate-failure | import-graph |
    When the plan advisor builds its directive for a feature
    Then the "Relevant context:" block includes a "Recurring gate failures" line
    And that line names "g1-file-size" with count 3
    And that line names "coverage" with count 2
    And that line names "import-graph" with count 1
    And "g1-file-size" appears before "coverage" which appears before "import-graph" on that line

  Scenario: Recurring gate failures counts match centinela insights for the same ledger
    Given a telemetry ledger containing a known mix of gate-failure events
    When the plan advisor builds its directive for a feature
    And the operator runs "centinela insights" on the same ledger
    Then the advisor's "Recurring gate failures" counts agree with the insights Gates section for every gate
    And neither uses a separately-implemented counter

  Scenario: Ties in failure count break by gate name ascending for reproducible output
    Given a telemetry ledger containing:
      | type         | gate   |
      | gate-failure | z-gate |
      | gate-failure | z-gate |
      | gate-failure | a-gate |
      | gate-failure | a-gate |
      | gate-failure | m-gate |
      | gate-failure | m-gate |
    When the plan advisor builds its directive for a feature
    Then the "Recurring gate failures" line orders the gates "a-gate", "m-gate", "z-gate"

  Scenario: A gate-failure event with an empty Gate field buckets under "<none>" without crashing
    Given a telemetry ledger containing a gate-failure event with an empty Gate field
    When the plan advisor builds its directive for a feature
    Then the "Recurring gate failures" line renders that bucket as "<none>"
    And the advisor does not crash or panic

  Scenario: Only the top-N gates are listed when more distinct gates failed
    Given a telemetry ledger containing 8 distinct gate-failure buckets each with a different count
    And the advisor failure top-N is configured to 3
    When the plan advisor builds its directive for a feature
    Then the "Recurring gate failures" line lists exactly 3 gates
    And the 3 listed gates are the 3 with the highest counts in deterministic order

  # ---------------------------------------------------------------------------
  # Pre-warning question — convert recurring failures into a planning question
  # ---------------------------------------------------------------------------

  Scenario: A gate recurring at or above threshold produces a pre-warning question naming that gate
    Given the recurrence threshold for a pre-warning question is 3
    And a telemetry ledger in which gate "g1-file-size" has failed 5 times
    When the plan advisor builds its directive for a feature
    Then the directive includes a pre-warning question naming "g1-file-size"
    And that question carries a lens tag like the other advisor questions

  Scenario: A gate below the recurrence threshold produces no pre-warning question
    Given the recurrence threshold for a pre-warning question is 3
    And a telemetry ledger in which every gate has failed at most 2 times
    When the plan advisor builds its directive for a feature
    Then the directive adds no gate-failure pre-warning question

  Scenario: The pre-warning question respects the plan_question_limit cap
    Given "plan_question_limit" is set to 3
    And the advisor already has 3 questions to ask before the gate-failure pre-warning
    And a telemetry ledger in which gate "coverage" has recurred above threshold
    When the plan advisor builds its directive for a feature
    Then the directive contains at most 3 questions in total
    And the question count never exceeds the configured plan_question_limit

  # ---------------------------------------------------------------------------
  # Scope of action — plan step only, read-only, deterministic
  # ---------------------------------------------------------------------------

  Scenario: The advisor only surfaces recurring failures during the plan step
    Given a telemetry ledger in which gate "g1-file-size" has recurred above threshold
    When the workflow step is "code" rather than "plan"
    Then the advisor does not run and surfaces no "Recurring gate failures" line
    And no gate-failure pre-warning question is produced

  Scenario: Headless mode leaves advisor behaviour unchanged
    Given the advisor mode resolves to off for headless invocation
    And a telemetry ledger in which gate "coverage" has recurred above threshold
    When the plan advisor builds its directive for a feature
    Then the advisor exits silently with no directive
    And the ledger contributes no output

  Scenario: The advisor never writes to the ledger
    Given a telemetry ledger with a known byte content
    When the plan advisor builds its directive for a feature
    Then the ledger file is unchanged byte-for-byte after the advisor runs

  Scenario: Two runs on the same ledger produce byte-identical advisor output
    Given a telemetry ledger containing a fixed set of gate-failure events
    When the plan advisor builds its directive for the same feature twice in succession
    Then both directives are byte-identical
