Feature: Roadmap quality evaluator enforcement
  As a maintainer
  I want roadmap quality scores enforced before feature start
  So planning quality is consistently high across all roadmap features

  Scenario: Greenfield start is blocked when any feature overall score is below 9
    Given a greenfield project with ROADMAP.md and roadmap analysis artifacts
    And roadmap quality JSON includes a feature with overall score 8
    When centinela start is attempted for a roadmap feature
    Then start should fail with roadmap quality guidance

  Scenario: Roadmap validate fails when quality coverage is incomplete
    Given roadmap quality JSON missing one roadmap feature
    When centinela roadmap validate is run
    Then validation should fail with missing quality feature error

  Scenario: Roadmap validate passes with complete high-quality scoring
    Given roadmap quality JSON covers all roadmap features with overall scores 9 or 10
    When centinela roadmap validate is run
    Then validation should pass
