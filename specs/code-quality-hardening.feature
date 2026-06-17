Feature: Code quality hardening
  Fix the evidence key-order drift, gate gofmt formatting, and make
  config and workflow-state errors transparent instead of silent.

  Scenario: Hook formatter preserves the canonical evidence key order
    Given a role evidence JSON document containing a coverage field
    When the evidence is marshalled canonically and reformatted by the postwrite hook formatter
    Then both outputs are byte-identical
    And the coverage key appears between mobileFirst and handoffTo

  Scenario: Unformatted Go source fails the format check
    Given a Go source file that is not gofmt-formatted
    When the format check script runs over that file's tree
    Then it exits non-zero
    And it prints the offending file path

  Scenario: Formatted tree passes the format check
    Given a source tree where every Go file is gofmt-formatted
    When the format check script runs
    Then it exits zero
    And it prints nothing

  Scenario: Validate suite gates formatting
    Given the project centinela.toml
    When the validate command list is read
    Then it includes the format check script ./scripts/check-fmt.sh

  Scenario: Starting a feature with a corrupted config fails loudly
    Given a centinela.toml that cannot be parsed
    When the user runs centinela start for a new feature
    Then the command exits with an error naming centinela.toml
    And no workflow state file is created

  Scenario: Prompt hook degrades with a warning on corrupted config
    Given a centinela.toml that cannot be parsed
    When the prompt context hook runs
    Then the hook exits zero so the host session continues
    And the injected context contains a config warning naming the failure

  Scenario: Loading a missing workflow reports absence
    Given no workflow state file exists for a feature
    When the workflow is loaded by name
    Then the error states no workflow was found for that feature

  Scenario: Loading a corrupted workflow reports the cause
    Given a workflow state file containing invalid JSON
    When the workflow is loaded by name
    Then the error names the state file path
    And the error includes the underlying parse failure

  Scenario: Loading an unreadable workflow is not reported as absence
    Given a workflow state file that exists but cannot be read
    When the workflow is loaded by name
    Then the error names the state file path
    And the error does not state that no workflow was found
