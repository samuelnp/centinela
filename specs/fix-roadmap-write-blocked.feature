Feature: Roadmap phase writes are not blocked

  Scenario: Writing a feature brief without an active workflow is allowed
    Given no feature workflow is active
    When Claude writes docs/features/caesar-cipher.md
    Then the prewrite hook allows the write

  Scenario: Writing ROADMAP.md without an active workflow is allowed
    Given no feature workflow is active
    When Claude writes ROADMAP.md
    Then the prewrite hook allows the write

  Scenario: Feature brief writes are still allowed during plan step
    Given a feature workflow is active in the plan step
    When Claude writes docs/features/my-feature.md
    Then the prewrite hook allows the write
