Feature: Strict orchestration smoke simulation
  Scenario: Required roles are enforced per step
    Given strict orchestration mode is enabled
    When step completion is attempted
    Then required role evidence must exist
