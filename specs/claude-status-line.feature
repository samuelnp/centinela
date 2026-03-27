Feature: Claude status line shows Centinela workflow state
  As a developer using Claude Code
  I want a compact Centinela status line
  So that I can always see step, blockers, and next action

  Scenario: No active workflow
    Given no active workflow exists
    When I run "centinela hook statusline"
    Then the output should include "WF:none"
    And the output should include "BLOCK:NO_WORKFLOW"

  Scenario: Active workflow in plan step
    Given feature "alpha" is in "plan" step
    When I run "centinela hook statusline"
    Then the output should include "STEP:plan"
    And the output should include "NEXT:write-plan"

  Scenario: Active workflow in tests step without edge cases report
    Given feature "alpha" is in "tests" step
    And ".workflow/alpha-edge-cases.md" does not exist
    When I run "centinela hook statusline"
    Then the output should include "BLOCK:MISSING_EDGE_CASES"

  Scenario: Claude setup wires statusLine command
    Given a project with ".claude/settings.json"
    When I run "centinela init --agent claude"
    Then settings should include a "statusLine" command
    And existing hooks should remain configured
