Feature: Migrate managed docs to latest Centinela templates

  Scenario: Preview migration plan for legacy managed docs
    Given managed docs exist without version headers
    When the user runs "centinela migrate docs"
    Then Centinela prints planned updates without modifying files

  Scenario: Apply migration after user approval
    Given a migration plan contains outdated managed docs
    When the user runs "centinela migrate docs --apply"
    Then Centinela updates docs to current versions
    And Centinela preserves keep blocks and custom sections

  Scenario: Prompt hook requests confirmation flow
    Given outdated managed docs are detected
    When the context hook runs on prompt submit
    Then Centinela shows migration summary
    And instructs the assistant to ask for user approval before apply
