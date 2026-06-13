Feature: Governance telemetry event log
  As a Centinela maintainer and the five downstream insight features
  I want every governance signal (block, gate-failure, verify-rejection, refused/successful advance) appended to a durable JSONL log
  So that governance friction becomes queryable instead of scrolling past in a transcript

  # The log is an append-only JSONL file at .workflow/telemetry/events.jsonl, one
  # JSON object per line. Emission is best-effort and non-fatal (mirrors
  # memory.Capture): it never changes an exit code, blocks a write, or fails an
  # advance. Every line is self-describing via schema="centinela.telemetry/v1".
  # Scenario titles below map 1:1 to Go tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with telemetry enabled
    And the telemetry log directory is ".workflow/telemetry"
    And the telemetry log file is ".workflow/telemetry/events.jsonl"

  # --- Block events (hook_prewrite.go) ---

  Scenario: Out-of-step write appends a block event with full context
    Given an active workflow for feature "alpha" on step "code"
    When a prewrite of a plan-type file is blocked as out-of-step
    Then a "block" event is appended
    And the event reason is "out-of-step"
    And the event feature is "alpha"
    And the event step is "code"
    And the event fileType is the classified file type
    And the event targetPath is the refused path

  Scenario: Write with no active workflow appends a need-init block event
    Given no active workflow
    When a prewrite is blocked because the project needs init
    Then a "block" event is appended
    And the event reason is "need-init"
    And the event has no feature
    And the event has no step
    And the event fileType is the classified file type
    And the event targetPath is the refused path

  # --- Gate-failure events (validate.go) ---

  Scenario: A failing gate during validate appends a gate-failure event
    Given a validate run that produces one gate with status Fail
    When validate processes the gate results
    Then a "gate-failure" event is appended
    And the event gate is the failing gate name
    And the event message is the failing gate message
    And the event has no feature

  Scenario: Each failing gate appends its own gate-failure event
    Given a validate run that produces two gates with status Fail
    When validate processes the gate results
    Then two "gate-failure" events are appended
    And each event carries its own gate name and message

  # --- Verify-rejection event (complete.go runClaimVerification) ---

  Scenario: A failed claim verification appends a verify-rejection event with the failing checks
    Given an active workflow for feature "alpha" on step "tests"
    And claim verification reports two failing checks
    When the advance is hard-blocked by verification
    Then a "verify-rejection" event is appended
    And the event feature is "alpha"
    And the event step is "tests"
    And the event checks contain each failing check's claim, role, status, and detail

  # --- Complete-rejected events (complete.go runComplete abort branches) ---

  Scenario: An advance aborted by validate gates appends complete-rejected with reason gates
    Given an active workflow for feature "alpha" on step "validate"
    When the advance is aborted because validate gates failed
    Then a "complete-rejected" event is appended
    And the event reason is "gates"
    And the event feature is "alpha"
    And the event step is "validate"

  Scenario: An advance aborted by verification appends complete-rejected with reason verify
    Given an active workflow for feature "alpha" on step "tests"
    When the advance is aborted because claim verification failed
    Then a "complete-rejected" event is appended
    And the event reason is "verify"
    And the event feature is "alpha"
    And the event step is "tests"

  # --- Step-advanced event (complete.go after saveWorkflow) ---

  Scenario: A successful advance appends a step-advanced event carrying the just-completed step
    Given an active workflow for feature "alpha" on step "plan"
    When the advance from "plan" succeeds
    Then a "step-advanced" event is appended
    And the event feature is "alpha"
    And the event step is "plan"

  # --- Config: opt-out semantics ---

  Scenario: Telemetry disabled is a no-op and writes no file
    Given a Centinela-governed project with telemetry disabled
    When any governance event would be recorded
    Then no telemetry log file is written
    And no events are recorded

  Scenario: Absent telemetry config defaults to enabled and records events
    Given a Centinela-governed project with no telemetry config section
    When a governance event is recorded
    Then the telemetry log file is written
    And the event is present in the log

  # --- Schema and timestamp contract (every event) ---

  Scenario: Every recorded event carries the schema id and an RFC3339 timestamp
    Given telemetry is enabled
    When any event is recorded
    Then the event schema is "centinela.telemetry/v1"
    And the event timestamp parses as RFC3339 in UTC

  # --- Append-only accumulation and ordering ---

  Scenario: Multiple events accumulate append-only in call order
    Given telemetry is enabled
    When a "step-advanced" event is recorded then a "gate-failure" event is recorded
    Then the log contains two events in that order
    And the first recorded event is not overwritten

  Scenario: Two sequential records both land intact under append-only writes
    Given telemetry is enabled
    When two events are recorded sequentially
    Then both events are readable from the log
    And every line in the log parses as a complete JSON object

  # --- Non-blocking / best-effort contract ---

  Scenario: An I/O error while recording does not fail the host command
    Given telemetry is enabled but the log location cannot be written
    When an event is recorded
    Then recording returns no error to the caller
    And the host command's control flow and exit code are unchanged
    And a warning is emitted to stderr

  # --- Lenient reader contract ---

  Scenario: Read skips a corrupt line and returns the valid events
    Given a telemetry log containing one valid event line then a garbage line then another valid event line
    When the log is read
    Then the two valid events are returned
    And the garbage line is skipped without error

  Scenario: Read of a missing telemetry log returns no events and no error
    Given no telemetry log file exists
    When the log is read
    Then no events are returned
    And no error is raised

  # --- Derived rework metric ---

  Scenario: Rework is derivable from two complete-rejected events before a step-advanced
    Given an active workflow for feature "alpha" on step "validate"
    When two "complete-rejected" events for "alpha"/"validate" are recorded then a "step-advanced" event for "alpha"/"validate" is recorded
    When the log is read
    Then a reader counting complete-rejected before the step-advanced for "alpha"/"validate" computes a rework count of 2
