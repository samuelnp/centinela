Feature: Spec traceability gate
  Verify that every Gherkin scenario in the spec directory maps to an
  acceptance test in the executed suite, scoped diff-aware, so scenarios
  cannot silently go unimplemented.

  Scenario: A scenario with a matching acceptance test passes the gate
    Given a spec file whose scenario has a matching acceptance test comment
    When the spec-traceability gate runs over that spec
    Then the gate passes
    And the message reports the covered scenario count

  Scenario: A scenario with no acceptance test fails the gate
    Given a spec file with a scenario that no acceptance test covers
    When the spec-traceability gate runs over that spec
    Then the gate fails
    And the details name the uncovered scenario and its spec file

  Scenario: Matching normalizes trailing period, spacing, and letter case
    Given a spec scenario named "Start the watcher"
    And an acceptance comment reading "// Scenario:  start the WATCHER ."
    When the spec-traceability gate runs over that spec
    Then the gate passes
    And the scenario is reported as covered

  Scenario: An acceptance header with a trailing annotation still matches its spec
    Given an acceptance test whose header reads "// Acceptance: specs/spec-traceability-gate.feature (AC4, AC5)"
    And that test carries a matching "// Scenario:" comment for a scenario in that spec
    When the spec-traceability gate runs over that spec
    Then the trailing annotation after the filename is ignored
    And the scenario is reported as covered

  Scenario: A Scenario Outline counts as one covered scenario
    Given a spec file containing a Scenario Outline with an examples table
    When the gate evaluates coverage
    Then the outline is treated as a single scenario for matching

  Scenario: Warn severity reports gaps without failing
    Given the gate is configured with severity set to warn, Centinela's own dogfood default
    And a spec scenario has no acceptance test
    When the gate runs
    Then the gate result status is warn rather than fail
    And the uncovered scenario is still listed in the details

  Scenario: Diff-aware scope gates only changed spec files
    Given an unchanged spec file with an uncovered scenario
    And a changed spec file whose scenarios are all covered
    When the gate runs in diff-aware mode
    Then the unchanged spec is not gated
    And the gate passes

  Scenario: No spec files in scope skips the gate
    Given no spec files are in the gate's scope
    When the gate runs
    Then the gate is skipped with an explanatory message

  Scenario: An unknown severity value is rejected at config load
    Given a centinela.toml that sets the gate severity to an unsupported value
    When the configuration is loaded
    Then loading fails with an error naming the severity field

  Scenario: The gate is registered and enabled for Centinela in warn mode
    Given Centinela's own centinela.toml
    When the configured gates are read
    Then the spec-traceability gate is enabled
    And its severity is warn
