Feature: Semantic version release automation
  As a maintainer
  I want automated version bumping and release artifact publishing
  So users can install stable compiled binaries from releases

  Scenario: Push to main creates semantic bump and tag
    Given conventional commit messages since the last release tag
    When a push lands on main
    Then the workflow should bump Makefile VERSION
    And create and push a new vX.Y.Z tag

  Scenario: Tag push builds and publishes release artifacts
    Given a pushed vX.Y.Z tag
    When release workflow runs
    Then it should build platform binaries
    And upload checksums and assets to GitHub Release

  Scenario: Installer script fetches latest release binary
    Given a user platform and architecture
    When scripts/install.sh runs
    Then it should download matching artifact
    And verify checksum before installation
