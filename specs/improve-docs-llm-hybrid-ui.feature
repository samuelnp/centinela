Feature: Generate polished project docs with LLM-first hybrid flow
  As a maintainer
  I want documentation generation to prioritize LLM synthesis and polished UI
  So project docs are navigable, visual, and useful for stakeholders

  Scenario: HTML documentation includes navigation and examples
    Given roadmap and feature artifacts exist
    When centinela docs generate is run
    Then an HTML document should be written
    And it should include navigation anchors and section links
    And it should include documentation examples and summary cards

  Scenario: Mermaid graphs focus on feature understanding
    Given roadmap analysis and specs exist
    When centinela docs generate is run
    Then it should include Mermaid graph blocks for project features
    And it should not include workflow-specific Mermaid handoff graphs

  Scenario: Prompt guidance is LLM-first with command fallback
    Given documentation generator prompt exists
    Then it should instruct LLM narrative synthesis first
    And it should keep centinela docs generate as a fallback path
