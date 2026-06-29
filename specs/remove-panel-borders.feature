Feature: Remove panel border boxes
  As a Centinela user
  I want system messages rendered without a border box
  So that the output is less visually noisy

  Background:
    Given Centinela renders system messages through renderSystemPanel

  Scenario: A CLI command panel renders without a border
    When a command prints a system panel (e.g. roadmap phase overview)
    Then the output contains the channel, title, and body
    And the output contains no rounded border characters

  Scenario: A hook directive panel renders without a border
    When the prewrite hook renders a blocked-write directive
    Then the output contains "BLOCKED WRITE" and the next action
    And the output contains no rounded border characters

  Scenario: Branding is preserved
    When any system panel renders
    Then it still shows the 🛡️👁️ persona label and its channel tag

  Scenario: Single-line CLI output is unchanged
    When a command prints a one-line result via RenderSuccess
    Then it renders as a single branded line with no border (unchanged)
