Feature: centinela insights — governance-telemetry analytics report
  As a repo maintainer, governance owner, roadmap planner, or tooling author
  I want `centinela insights` to read the governance-telemetry log and report
  the most-triggered blocks, most-failed gates, features with the most rework,
  and mean steps-to-green
  So that prioritization is driven by counts from the event log, not gut feel

  # centinela insights reads .workflow/telemetry/events.jsonl read-only and
  # computes four ranked metrics from the parsed event slice. Output is a
  # sectioned human report (default) or structured JSON (--json). Empty or
  # missing log → clean "no telemetry yet" report, exit 0.
  # --top N (default 5) controls how many entries appear per ranked section.
  # Scenario titles map 1:1 to Go acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with a valid centinela.toml
    And the telemetry log is at ".workflow/telemetry/events.jsonl"

  # ---------------------------------------------------------------------------
  # Empty / missing log — must never be an error
  # ---------------------------------------------------------------------------

  Scenario: Missing telemetry log prints clean empty-state report and exits 0
    Given no telemetry log file exists
    When the operator runs:
      centinela insights
    Then the command exits with code 0
    And the output contains "no telemetry yet"
    And the output does not contain an error message or stack trace

  Scenario: Empty telemetry log prints clean empty-state report and exits 0
    Given the telemetry log exists but contains no events
    When the operator runs:
      centinela insights
    Then the command exits with code 0
    And the output contains "no telemetry yet"

  Scenario: Whitespace-only telemetry log is treated as empty and exits 0
    Given the telemetry log exists and contains only blank lines
    When the operator runs:
      centinela insights
    Then the command exits with code 0
    And the output contains "no telemetry yet"

  # ---------------------------------------------------------------------------
  # Most-triggered blocks
  # ---------------------------------------------------------------------------

  Scenario: Blocks section ranks block events by count descending
    Given a telemetry log containing:
      | type  | reason       | fileType | step |
      | block | out-of-step  | plan     | code |
      | block | out-of-step  | plan     | code |
      | block | out-of-step  | plan     | code |
      | block | need-init    | source   |      |
      | block | need-init    | source   |      |
    When the operator runs:
      centinela insights
    Then the command exits with code 0
    And the Blocks section lists "out-of-step · plan" with count 3
    And the Blocks section lists "need-init · source" with count 2
    And "out-of-step · plan" appears before "need-init · source" in the Blocks section

  Scenario: Blocks section respects --top N flag
    Given a telemetry log containing 10 distinct block buckets each with 1 event
    When the operator runs:
      centinela insights --top 3
    Then the command exits with code 0
    And the Blocks section contains exactly 3 entries

  Scenario: Default --top is 5 for the blocks section
    Given a telemetry log containing 8 distinct block buckets each with 1 event
    When the operator runs:
      centinela insights
    Then the Blocks section contains exactly 5 entries

  Scenario: Blocks section ties break by key ascending for stable ordering
    Given a telemetry log containing:
      | type  | reason   | fileType |
      | block | alpha    | plan     |
      | block | beta     | plan     |
    When the operator runs:
      centinela insights
    Then "alpha · plan" appears before "beta · plan" in the Blocks section

  Scenario: Block event with empty fileType buckets under reason and empty-fileType key
    Given a telemetry log containing a block event with reason "out-of-step" and no fileType field
    When the operator runs:
      centinela insights
    Then the Blocks section contains an entry whose key includes "out-of-step"
    And the entry does not crash or panic

  Scenario: Log with only block events shows non-empty Blocks section and gracefully empty Gates and Rework sections
    Given a telemetry log containing only block events
    When the operator runs:
      centinela insights
    Then the command exits with code 0
    And the Blocks section is non-empty
    And the Gates section contains no entries
    And the Rework section contains no entries

  Scenario: --top N larger than available block buckets returns all buckets without padding
    Given a telemetry log containing 2 distinct block buckets
    When the operator runs:
      centinela insights --top 10
    Then the Blocks section contains exactly 2 entries

  # ---------------------------------------------------------------------------
  # Most-failed gates
  # ---------------------------------------------------------------------------

  Scenario: Gates section ranks gate-failure events by count descending
    Given a telemetry log containing:
      | type         | gate            |
      | gate-failure | coverage        |
      | gate-failure | coverage        |
      | gate-failure | import-graph    |
    When the operator runs:
      centinela insights
    Then the command exits with code 0
    And the Gates section lists "coverage" with count 2
    And the Gates section lists "import-graph" with count 1
    And "coverage" appears before "import-graph" in the Gates section

  Scenario: Gates section respects --top N flag
    Given a telemetry log containing 7 distinct gate-failure buckets each with 1 event
    When the operator runs:
      centinela insights --top 2
    Then the Gates section contains exactly 2 entries

  Scenario: Gate-failure event with empty Gate field buckets under key rendered as none
    Given a telemetry log containing a gate-failure event with an empty Gate field
    When the operator runs:
      centinela insights
    Then the Gates section contains an entry rendered as "<none>"
    And the command exits with code 0

  Scenario: Log with only gate-failure events shows non-empty Gates section and gracefully empty other sections
    Given a telemetry log containing only gate-failure events
    When the operator runs:
      centinela insights
    Then the command exits with code 0
    And the Gates section is non-empty
    And the Blocks section contains no entries
    And the Rework section contains no entries

  Scenario: Gates section ties break by gate name ascending for stable ordering
    Given a telemetry log containing:
      | type         | gate      |
      | gate-failure | security  |
      | gate-failure | coverage  |
    When the operator runs:
      centinela insights
    Then "coverage" appears before "security" in the Gates section

  # ---------------------------------------------------------------------------
  # Features with most rework
  # ---------------------------------------------------------------------------

  Scenario: Rework section ranks features by gate-failure plus verify-rejection plus complete-rejected count
    Given a telemetry log containing:
      | type              | feature     |
      | gate-failure      | alpha       |
      | gate-failure      | alpha       |
      | verify-rejection  | alpha       |
      | complete-rejected | beta        |
      | gate-failure      | beta        |
    When the operator runs:
      centinela insights
    Then the command exits with code 0
    And the Rework section lists "alpha" with count 3
    And the Rework section lists "beta" with count 2
    And "alpha" appears before "beta" in the Rework section

  Scenario: Rework section excludes events with no feature field
    Given a telemetry log containing:
      | type         | feature |
      | gate-failure |         |
      | gate-failure | alpha   |
    When the operator runs:
      centinela insights
    Then the Rework section lists "alpha" with count 1
    And the Rework section does not list an entry with an empty feature name

  Scenario: Rework section respects --top N flag
    Given a telemetry log containing gate-failure events for 6 distinct features each with 1 event
    When the operator runs:
      centinela insights --top 3
    Then the Rework section contains exactly 3 entries

  Scenario: Rework section ties break by feature name ascending for stable ordering
    Given a telemetry log containing:
      | type         | feature |
      | gate-failure | zeta    |
      | gate-failure | alpha   |
    When the operator runs:
      centinela insights
    Then "alpha" appears before "zeta" in the Rework section

  Scenario: step-advanced events are not counted in rework score
    Given a telemetry log containing:
      | type          | feature |
      | step-advanced | alpha   |
      | step-advanced | alpha   |
      | gate-failure  | alpha   |
    When the operator runs:
      centinela insights
    Then the Rework section lists "alpha" with count 1

  # ---------------------------------------------------------------------------
  # Mean steps-to-green
  # ---------------------------------------------------------------------------

  Scenario: Mean steps-to-green is computed correctly from known counts
    Given a telemetry log containing:
      | type              | count |
      | step-advanced     | 4     |
      | complete-rejected | 2     |
    When the operator runs:
      centinela insights
    Then the command exits with code 0
    And the Steps-to-Green metric shows "1.50"

  Scenario: Zero step-advanced events renders steps-to-green as n/a without panic
    Given a telemetry log containing only complete-rejected events
    When the operator runs:
      centinela insights
    Then the command exits with code 0
    And the Steps-to-Green metric shows "n/a"

  Scenario: Single step-advanced with no rejections yields mean of 1.00
    Given a telemetry log containing exactly one step-advanced event and no complete-rejected events
    When the operator runs:
      centinela insights
    Then the Steps-to-Green metric shows "1.00"

  Scenario: Single step-advanced with one rejection yields mean of 2.00
    Given a telemetry log containing exactly one step-advanced event and one complete-rejected event
    When the operator runs:
      centinela insights
    Then the Steps-to-Green metric shows "2.00"

  Scenario: Log with only step-advanced events and no complete-rejected reports mean of 1.00
    Given a telemetry log containing 5 step-advanced events and no complete-rejected events
    When the operator runs:
      centinela insights
    Then the Steps-to-Green metric shows "1.00"

  # ---------------------------------------------------------------------------
  # --json output
  # ---------------------------------------------------------------------------

  Scenario: --json emits structured Report as indented JSON and exits 0
    Given a telemetry log containing at least one event of each type
    When the operator runs:
      centinela insights --json
    Then the command exits with code 0
    And the output is valid JSON
    And the JSON contains fields "EventCount", "SpanStart", "SpanEnd", "Blocks", "Gates", "Rework", "StepsToGreen"
    And the output contains no ANSI escape sequences

  Scenario: --json on empty log emits a valid JSON Report with zero counts and exits 0
    Given no telemetry log file exists
    When the operator runs:
      centinela insights --json
    Then the command exits with code 0
    And the output is valid JSON
    And the JSON field "EventCount" is 0

  Scenario: --json output shape is stable across two runs on the same log
    Given a telemetry log containing a fixed set of events
    When the operator runs centinela insights --json twice in succession
    Then both outputs are byte-identical

  Scenario: --json output has stable field names usable by tooling
    Given a telemetry log containing a known mix of events
    When the operator runs:
      centinela insights --json
    Then the JSON object has exactly the top-level fields in the Report contract
    And no additional or renamed fields appear

  # ---------------------------------------------------------------------------
  # Determinism — same input, same output
  # ---------------------------------------------------------------------------

  Scenario: Two runs on the same log produce byte-identical human output
    Given a telemetry log containing a fixed set of events
    When the operator runs centinela insights twice in succession
    Then both outputs are byte-identical

  Scenario: Ties between buckets with equal count are always broken by key ascending
    Given a telemetry log where three gate-failure buckets each have count 2 with keys "z-gate", "a-gate", and "m-gate"
    When the operator runs:
      centinela insights
    Then the Gates section order is "a-gate", "m-gate", "z-gate"

  # ---------------------------------------------------------------------------
  # Malformed / resilient input
  # ---------------------------------------------------------------------------

  Scenario: Malformed JSONL lines are skipped and valid events are still aggregated
    Given a telemetry log containing one valid block event, then a garbage line, then one valid gate-failure event
    When the operator runs:
      centinela insights
    Then the command exits with code 0
    And the Blocks section counts the one valid block event
    And the Gates section counts the one valid gate-failure event

  Scenario: A log with a single event of each type is reported without crash
    Given a telemetry log containing exactly one event of each type: block, gate-failure, verify-rejection, complete-rejected, step-advanced
    When the operator runs:
      centinela insights
    Then the command exits with code 0
    And the Blocks section lists one entry
    And the Gates section lists one entry
    And the Rework section lists entries for features that appear in gate-failure, verify-rejection, or complete-rejected events
    And the Steps-to-Green metric shows "2.00"

  # ---------------------------------------------------------------------------
  # Non-TTY / piped output
  # ---------------------------------------------------------------------------

  Scenario: Piped output contains no ANSI escape sequences
    Given a telemetry log containing at least one event
    When the operator pipes the output:
      centinela insights | cat
    Then the output contains no ANSI escape sequences
    And the output is plain text parseable by grep and awk

  # ---------------------------------------------------------------------------
  # SpanStart / SpanEnd coverage metadata
  # ---------------------------------------------------------------------------

  Scenario: Report includes span of earliest and latest event timestamps
    Given a telemetry log containing events with timestamps "2026-01-01T00:00:00Z" and "2026-06-01T12:00:00Z"
    When the operator runs:
      centinela insights
    Then the human report includes the span range "2026-01-01" through "2026-06-01"

  Scenario: Report includes total event count considered
    Given a telemetry log containing exactly 7 valid events
    When the operator runs:
      centinela insights
    Then the human report includes the total event count 7
