Feature: Full-sync migration for managed docs and setup assets

  Scenario: Preview full migration plan
    Given managed docs and setup assets are outdated or missing
    When the user runs "centinela migrate"
    Then Centinela shows a unified preview for docs and setup assets
    And no files are modified

  Scenario: Apply full sync migration
    Given a full migration plan contains create and update actions
    When the user runs "centinela migrate --apply"
    Then Centinela creates missing managed assets and updates outdated ones
    And Centinela preserves unrelated user configuration keys

  Scenario: Setup-only migration with agent scope
    Given only OpenCode assets require migration
    When the user runs "centinela migrate setup --agent opencode --apply"
    Then Centinela applies only OpenCode setup changes
    And Claude settings are left unchanged

  Scenario: Hook guidance includes setup migration
    Given setup migration is required
    When the context migration hook runs on prompt submit
    Then Centinela shows migration-needed guidance for setup assets
    And instructs the assistant to ask for user approval before apply
