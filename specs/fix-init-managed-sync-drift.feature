Feature: Greenfield init leaves no pending migration
  As a Centinela user setting up a project
  I want centinela init to write managed assets in their migrated form
  So that centinela migrate does not report spurious pending updates

  Scenario: Fresh init reports no pending migrations
    Given a new repository
    When I run "centinela init"
    And I run "centinela migrate"
    Then no managed setup assets are reported as requiring migration

  Scenario: Init writes the managed-version header
    When I run "centinela init"
    Then AGENTS.md begins with the centinela managed-version marker
    And .opencode/plugins/centinela.js begins with the managed-version marker

  Scenario: The managed-sync path is idempotent
    Given centinela init has written the OpenCode managed assets
    When the opencode sync plan is rebuilt
    Then no item is marked as create or update
