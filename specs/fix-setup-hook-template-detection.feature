Feature: Setup hook project detection and directives

  Scenario: Roadmap guidance after template rename
    Given a centinela project has PROJECT.md but no PROJECT.md.template
    And ROADMAP.md is missing
    When centinela hook setup runs
    Then it prints roadmap-required guidance

  Scenario: Plain directive line accompanies boxed guidance
    Given setup guidance is required
    When centinela hook setup runs
    Then output includes a CENTINELA directive line before panel content
