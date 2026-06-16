Feature: custom gate SDK — team-defined command-backed gates as first-class governance
  As a team with a project-specific mechanical rule, a governance owner, or a CI gate author
  I want to declare `[[gates.custom]]` command-backed gates that run inside `centinela validate`
  and produce the same structured Result as built-in gates — with severity, telemetry, and
  baseline participation
  So that project rules are first-class governance, not opaque pass/fail shell lines, and I never
  have to fork Centinela to enforce my own policy

  # A `[[gates.custom]]` entry declares: name (unique, not colliding with a built-in),
  # command (shell-executed, trusted checked-in config), severity (fail|warn), optional
  # output (blob|lines), optional enabled (default true), optional timeout_seconds.
  # During `centinela validate` each ENABLED custom gate runs its command:
  #   exit 0          → the gate PASSES.
  #   non-zero exit   → severity=fail → the gate FAILS and BLOCKS validate (exit 1).
  #                   → severity=warn → the gate is reported but does NOT block (exit 0).
  # Command stdout/stderr is surfaced in the gate's Details. output="lines" turns each
  # stdout line of a failing command into a separate Details violation entry (so the audit
  # ratchet fingerprints them individually). Custom-gate failures emit `gate-failure`
  # telemetry events and are baseline-able by `centinela audit baseline` exactly like
  # built-in gates. Config is validated up-front (empty command, empty/duplicate name,
  # built-in collision, invalid severity, invalid output mode → load/validate fails with a
  # clear indexed error). A hung command is killed by a per-gate timeout and fails the gate
  # rather than hanging validate. Custom gates are additive: built-ins keep their behaviour
  # and remain the reference implementation of the Result contract.
  # Scenario titles map 1:1 to Go acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with a valid centinela.toml
    And custom gates are declared as `[[gates.custom]]` array-of-tables entries
    And each entry has at least a `name` and a `command`
    And `severity` defaults to "fail" and `enabled` defaults to true unless a scenario states otherwise
    And a custom gate's command is executed through a shell with a per-gate timeout

  # ---------------------------------------------------------------------------
  # Passing — exit 0
  # ---------------------------------------------------------------------------

  Scenario: A passing custom gate appears in the validate gate report by its name
    Given a custom gate named "no-todo" with command "true" and severity "fail"
    When the operator runs:
      centinela validate
    Then the command exits with code 0
    And the gate report contains a gate named "no-todo"
    And the "no-todo" gate is reported as passing
    And the output does not contain an error message or stack trace

  # ---------------------------------------------------------------------------
  # Failing with severity=fail — blocks
  # ---------------------------------------------------------------------------

  Scenario: A failing severity-fail custom gate blocks validate and surfaces command output in its details
    Given a custom gate named "no-console-log" with command "sh -c 'echo found in src/app.js; exit 1'" and severity "fail"
    When the operator runs:
      centinela validate
    Then the command exits with a non-zero code
    And the "no-console-log" gate is reported as failing
    And the "no-console-log" gate's details contain "found in src/app.js"

  Scenario: A failing severity-fail custom gate that prints nothing falls back to a generic failure detail
    Given a custom gate named "silent-fail" with command "false" and severity "fail"
    When the operator runs:
      centinela validate
    Then the command exits with a non-zero code
    And the "silent-fail" gate is reported as failing
    And the "silent-fail" gate's details contain a non-empty generic failure message

  # ---------------------------------------------------------------------------
  # Failing with severity=warn — does NOT block
  # ---------------------------------------------------------------------------

  Scenario: A failing severity-warn custom gate is reported but does not block validate
    Given a custom gate named "style-nit" with command "sh -c 'echo nit; exit 1'" and severity "warn"
    When the operator runs:
      centinela validate
    Then the command exits with code 0
    And the "style-nit" gate is reported as a warning rather than a blocking failure
    And the "style-nit" gate's details contain "nit"

  # ---------------------------------------------------------------------------
  # Independence — one failure does not stop the others
  # ---------------------------------------------------------------------------

  Scenario: Multiple custom gates run independently and a failing one does not prevent the others from reporting
    Given a custom gate named "gate-a" with command "true" and severity "fail"
    And a custom gate named "gate-b" with command "sh -c 'echo b-broke; exit 1'" and severity "fail"
    And a custom gate named "gate-c" with command "true" and severity "fail"
    When the operator runs:
      centinela validate
    Then the command exits with a non-zero code
    And the gate report contains a gate named "gate-a" reported as passing
    And the gate report contains a gate named "gate-b" reported as failing
    And the gate report contains a gate named "gate-c" reported as passing

  # ---------------------------------------------------------------------------
  # Disabled / empty — no behaviour change
  # ---------------------------------------------------------------------------

  Scenario: A custom gate with enabled=false does not run and leaves validate output unchanged
    Given a custom gate named "skipped" with command "sh -c 'echo should-not-run; exit 1'" and enabled false
    When the operator runs:
      centinela validate
    Then the command exits with code 0
    And the gate report does not contain a gate named "skipped"
    And the output does not contain "should-not-run"

  Scenario: No custom gate entries leaves validate output byte-identical to a run with no custom gates configured
    Given the centinela.toml declares no `[[gates.custom]]` entries
    When the operator runs centinela validate with and without an empty custom-gates section
    Then both validate outputs are byte-identical
    And no custom gate appears in either gate report

  # ---------------------------------------------------------------------------
  # Config validation — reject bad declarations up front
  # ---------------------------------------------------------------------------

  Scenario: A custom gate with an empty command is rejected with a clear config error
    Given a custom gate named "no-cmd" with an empty command
    When the operator runs:
      centinela validate
    Then the command exits with a non-zero code
    And the output reports a config error identifying the gate at index 0
    And the output does not contain a runtime panic or stack trace

  Scenario: A custom gate with an empty name is rejected with a clear config error
    Given a custom gate with an empty name and command "true"
    When the operator runs:
      centinela validate
    Then the command exits with a non-zero code
    And the output reports a config error for a missing gate name

  Scenario: Two custom gates with duplicate names are rejected with a clear config error
    Given a custom gate named "dup" with command "true"
    And a second custom gate named "dup" with command "true"
    When the operator runs:
      centinela validate
    Then the command exits with a non-zero code
    And the output reports a config error naming the duplicate "dup"

  Scenario: A custom gate whose name collides with a built-in gate name is rejected
    Given a custom gate named "import_graph" with command "true"
    When the operator runs:
      centinela validate
    Then the command exits with a non-zero code
    And the output reports a config error that "import_graph" collides with a built-in gate

  Scenario: A custom gate with an invalid severity is rejected with a clear config error
    Given a custom gate named "bad-sev" with command "true" and severity "critical"
    When the operator runs:
      centinela validate
    Then the command exits with a non-zero code
    And the output reports a config error that severity must be "fail" or "warn"

  Scenario: A custom gate with an invalid output mode is rejected with a clear config error
    Given a custom gate named "bad-output" with command "true" and output "json"
    When the operator runs:
      centinela validate
    Then the command exits with a non-zero code
    And the output reports a config error that output must be "blob" or "lines"

  # ---------------------------------------------------------------------------
  # Timeout — a hung command fails the gate, never hangs validate
  # ---------------------------------------------------------------------------

  Scenario: A custom gate command that exceeds its timeout fails the gate with a timeout message
    Given a custom gate named "hang" with command "sleep 60" and timeout_seconds 1
    When the operator runs:
      centinela validate
    Then the command exits with a non-zero code within a few seconds
    And the "hang" gate is reported as failing
    And the "hang" gate's details contain a timeout message
    And validate does not hang waiting for the command

  # ---------------------------------------------------------------------------
  # Command not found / not executable — clear failure, no crash
  # ---------------------------------------------------------------------------

  Scenario: A custom gate whose command is not found fails the gate with a clear message and does not crash
    Given a custom gate named "missing-bin" with command "this-binary-does-not-exist --check" and severity "fail"
    When the operator runs:
      centinela validate
    Then the command exits with a non-zero code
    And the "missing-bin" gate is reported as failing
    And the "missing-bin" gate's details contain a clear command-execution-failed message
    And the output does not contain a runtime panic or stack trace

  # ---------------------------------------------------------------------------
  # output = "lines" — per-line violations
  # ---------------------------------------------------------------------------

  Scenario: A failing custom gate with output=lines turns each stdout line into a separate violation detail
    Given a custom gate named "per-line" with command "printf 'a.go:1\nb.go:2\nc.go:3\n'; exit 1" and output "lines"
    When the operator runs:
      centinela validate
    Then the command exits with a non-zero code
    And the "per-line" gate is reported as failing
    And the "per-line" gate has exactly 3 separate violation details
    And the details include "a.go:1" and "b.go:2" and "c.go:3" as distinct entries

  # ---------------------------------------------------------------------------
  # Baseline / ratchet participation
  # ---------------------------------------------------------------------------

  Scenario: A failing custom gate is baseline-able and then tolerated by audit while a new violation blocks
    Given a custom gate named "per-line" with command "printf 'a.go:1\nb.go:2\n'; exit 1" and output "lines"
    When the operator runs:
      centinela audit baseline
    Then the command exits with code 0
    And the baseline records a fingerprint for each custom-gate violation line
    When the custom gate's command later emits an additional line "c.go:3" and the operator runs:
      centinela audit
    Then the command exits with a non-zero code
    And the lines "a.go:1" and "b.go:2" are reported as baselined and tolerated
    And the line "c.go:3" is reported under the "new" partition

  # ---------------------------------------------------------------------------
  # Telemetry
  # ---------------------------------------------------------------------------

  Scenario: A failing custom gate is recorded as a gate-failure telemetry event
    Given a custom gate named "telemetry-fail" with command "false" and severity "fail"
    When the operator runs:
      centinela validate
    Then the command exits with a non-zero code
    And a "gate-failure" telemetry event is appended to the telemetry log
    And the recorded gate-failure event names the gate "telemetry-fail"

  # ---------------------------------------------------------------------------
  # Determinism
  # ---------------------------------------------------------------------------

  Scenario: Two validate runs with the same deterministic custom command produce the same gate report
    Given a custom gate named "stable" with command "sh -c 'echo x:1; echo y:2; exit 1'" and output "lines"
    When the operator runs centinela validate twice in succession
    Then both runs report the "stable" gate as failing
    And the "stable" gate's violation details are identical across the two runs
    And both runs exit with the same non-zero code
