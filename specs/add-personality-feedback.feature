Feature: Personality-first centinela feedback
  As a centinela user
  I want command and hook messages to include a recognizable voice
  So workflow feedback is easier to parse and feels less robotic

  Scenario: Success and info messages include persona wording
    Given centinela renders a success line
    When the line is displayed in the terminal
    Then it includes the CENTINELA label
    And it includes a persona expression for success tone

  Scenario: Warning and error messages include stricter persona wording
    Given centinela renders warning and error panels
    When those messages are produced by hooks
    Then warning output includes a warning expression
    And error output includes an error expression

  Scenario: Persona output keeps actionable content intact
    Given a blocked write message is rendered
    When centinela shows next action guidance
    Then the reason and next action text remain present
    And the message keeps existing channel and title metadata
