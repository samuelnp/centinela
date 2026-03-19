Feature: Fix setup next-step suggestion

  Scenario: After writing PROJECT.md, Claude guides user to roadmap
    Given PROJECT.md does not exist
    When a user prompt is submitted
    Then Claude fills in PROJECT.md
    And the closing instruction says "next, let's define your roadmap"
    And the closing instruction does not mention "centinela start"
