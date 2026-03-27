Feature: README bootstrap tutorial
  Scenario: New user sees end-to-end setup flow
    Given a user is reading the Quick Start section
    When they follow the bootstrap tutorial
    Then they should see how PROJECT.md is created
    And they should see how roadmap artifacts are generated
    And they should see how to start the first feature

  Scenario: Natural-language control examples are clear
    Given a user prefers chat prompts over raw CLI commands
    When they read the tutorial examples
    Then they should see intent phrases for roadmap status
    And they should see intent phrases for starting a feature
    And they should see intent phrases for continuing current feature work
