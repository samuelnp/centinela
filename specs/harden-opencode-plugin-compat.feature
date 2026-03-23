Feature: OpenCode plugin compatibility hardening
  As a maintainer
  I want the generated plugin to tolerate payload variations
  So hook enforcement remains stable across OpenCode changes

  Scenario: File path extraction supports fallback keys
    Given tool input with alternate file path keys
    When plugin resolves file path
    Then it should still detect the intended path

  Scenario: Prompt append tolerates output shape differences
    Given prompt append output with prompt or context variants
    When hook text is appended
    Then plugin should avoid runtime errors and append where possible
