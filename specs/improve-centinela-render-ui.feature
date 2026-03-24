Feature: Explicit and attractive centinela render output
  As a centinela user
  I want system output to be clearly differentiated from LLM responses
  So I can quickly understand source and required action

  Scenario: Prewrite block output is explicitly system-branded
    Given a write is blocked by workflow policy
    When centinela renders the blocked message
    Then output shows a clear CENTINELA system header
    And output includes a compact reason and action hint

  Scenario: Context output separates status from action-required notices
    Given active workflows exist
    When context hook renders output
    Then workflow status appears in a branded section
    And review/brief/edge-case reminders appear in dedicated warning sections

  Scenario: Postwrite tag is compact but explicit
    Given a workflow file write succeeds
    When postwrite hook emits the tag
    Then tag output clearly identifies itself as centinela metadata
    And tag keeps feature step and progress in one line
