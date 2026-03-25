Feature: Release trigger after automated version bump
  As a maintainer
  I want releases to run after CI-created tags
  So every semantic bump on main produces published artifacts

  Scenario: Version bump workflow completion triggers release
    Given Version Bump finishes successfully
    When release workflow evaluates triggers
    Then it should run from workflow_run
    And resolve the v-tag from the bump commit

  Scenario: Manual tag push still triggers release
    Given a maintainer pushes a vX.Y.Z tag
    When release workflow runs
    Then it should build and publish artifacts for all supported targets

  Scenario: Release assets include checksum manifest
    Given release artifacts are built
    When publishing GitHub Release
    Then SHA256SUMS should be uploaded with binaries
