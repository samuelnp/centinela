Feature: Senior PM roadmap analysis enforcement
  As a maintainer
  I want roadmap dependency analysis to be required before feature start
  So feature sequencing is coherent and dependency-safe

  Scenario: Greenfield start is blocked without roadmap analysis
    Given a greenfield project with ROADMAP.md and .workflow/roadmap.json
    When centinela start is attempted for a non-bootstrap feature
    Then start should fail with roadmap analysis guidance

  Scenario: Roadmap validate fails on dependency cycle
    Given roadmap analysis JSON depends on a cyclic feature graph
    When centinela roadmap validate is run
    Then validation should fail with cycle error

  Scenario: Roadmap validate passes with complete analysis
    Given roadmap analysis JSON covers all roadmap features with valid dependencies
    When centinela roadmap validate is run
    Then validation should pass
