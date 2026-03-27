Feature: Generate human-readable project documentation
  As a maintainer
  I want Centinela to render HTML docs from project artifacts
  So I can share a complete and understandable project overview

  Scenario: Generate report from roadmap and workflow artifacts
    Given roadmap and feature artifacts exist
    When centinela docs generate is run
    Then an HTML document should be written
    And it should include roadmap and workflow sections
    And it should include Mermaid graph blocks

  Scenario: Validation fails when critical inputs are missing
    Given roadmap artifacts are missing
    When centinela docs validate is run
    Then command should fail with actionable guidance
