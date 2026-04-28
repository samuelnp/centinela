Feature: Enrich plan advisor context
  As a maintainer
  I want plan-advisor to use roadmap and related feature context
  So planning questions are informed by what the project already knows

  Scenario: Advisor prioritizes dependency context before same-phase siblings
    Given roadmap analysis defines dependencies for the active feature
    When the plan-advisor hook runs during plan
    Then related dependency context should be considered before sibling context

  Scenario: Advisor reuses related edge-case lessons
    Given related features have edge-case reports
    When the plan-advisor hook runs during plan
    Then it should surface planning questions informed by those related edge cases

  Scenario: Advisor remains concise while using richer context
    Given roadmap, specs, and related feature artifacts exist
    When the plan-advisor hook runs during plan
    Then it should summarize relevant context without dumping raw file contents
    And it should still respect the configured question limit
