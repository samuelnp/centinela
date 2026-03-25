Feature: Harden main release automation
  As a maintainer
  I want main pushes to drive semantic releases automatically
  So release artifacts stay consistent with versioned source

  Scenario: Push to main bumps semantic version and creates tag
    Given conventional commits exist since the latest v-tag
    When a push lands on main
    Then the workflow should update Makefile VERSION
    And commit the new release version
    And push a new vX.Y.Z tag

  Scenario: Tag workflow publishes six platform artifacts
    Given a pushed vX.Y.Z tag
    When the release workflow runs
    Then it should build linux amd64 and arm64 binaries
    And it should build darwin amd64 and arm64 binaries
    And it should build windows amd64 and arm64 binaries
    And it should publish SHA256SUMS and all binaries

  Scenario: Bot loop prevention is preserved
    Given the push actor is github-actions[bot]
    When version bump workflow evaluates execution
    Then it should skip the bump job to avoid recursive runs
