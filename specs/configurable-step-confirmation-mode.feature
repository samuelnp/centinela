Feature: Configurable workflow step confirmation prompting
  As a maintainer
  I want to configure when step confirmation prompts are shown
  So I can choose strict review mode or automated progression guidance

  Scenario: Default mode prompts for every completed step
    Given a workflow with a step that has valid artifacts
    When hook context runs without explicit step confirmation mode
    Then a review-required prompt should be rendered

  Scenario: After-plan mode prompts only at plan step
    Given workflow step confirmation mode is after_plan
    When hook context runs for plan and code steps with valid artifacts
    Then review-required prompt should render only for plan step

  Scenario: Auto mode suppresses review-required prompts
    Given workflow step confirmation mode is auto
    When hook context runs for a step with valid artifacts
    Then review-required prompt should not be rendered
