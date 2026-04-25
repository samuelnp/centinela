Feature: Non-interactive status output

  Scenario: Show status without a TTY
    Given a workflow exists for a feature
    And the command is executed in a non-interactive shell
    When the user runs `centinela status <feature>`
    Then the command exits successfully
    And it prints the workflow status without trying to open `/dev/tty`
