Feature: Release tag resolution after workflow_run
  As a maintainer
  I want Release to resolve tags from main after Version Bump
  So CI-created bump tags always publish artifacts

  Scenario: workflow_run resolves latest v-tag from main
    Given Version Bump completed successfully
    When Release starts from workflow_run
    Then it should fetch tags and origin/main
    And resolve the latest v-tag merged into origin/main

  Scenario: no main tag skips release gracefully
    Given no v-tag is merged into origin/main
    When Release resolves a workflow_run tag
    Then it should set skip output to true
    And not fail the workflow

  Scenario: manual tag push path remains unchanged
    Given a push to refs/tags/vX.Y.Z
    When Release resolves the tag
    Then it should use GITHUB_REF_NAME
