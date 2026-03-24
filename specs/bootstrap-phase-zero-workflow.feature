Feature: Bootstrap phase-zero workflow
  As a maintainer
  I want bootstrap setup to be mandatory only for greenfield projects
  So project initialization is safe without blocking existing-project adoption

  Scenario: Greenfield project requires bootstrap before non-bootstrap features
    Given Project Stage is greenfield
    And roadmap defines Phase 0: Bootstrap with incomplete features
    When I run "centinela start non-bootstrap-feature"
    Then centinela should block start with bootstrap guidance

  Scenario: Existing project does not require bootstrap phase
    Given Project Stage is existing
    And roadmap has no Phase 0: Bootstrap
    When I run "centinela start feature-x"
    Then centinela should start the workflow

  Scenario: Bootstrap feature uses three-step workflow
    Given Project Stage is greenfield
    And feature belongs to Phase 0: Bootstrap
    When workflow is created for that feature
    Then steps should be plan, code, and validate

  Scenario: Non-bootstrap tests step ignores placeholder files
    Given feature is in tests step
    And tests folders only contain .gitkeep files
    When I run "centinela complete <feature>"
    Then completion should fail requiring real test artifacts
