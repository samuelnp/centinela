Feature: docs-step-markdown-first
  The docs step produces markdown documentation humans actually read.
  The deterministic HTML pipeline (internal/docgen, docs generate, KB html,
  portal index.html) is deleted; the gate moves to real updated doc files
  under docs/ or README.md, and `centinela docs context <feature>` prints
  the curated feature-scale inputs the documentation specialist needs.

  Background:
    Given a Centinela project with an active workflow in the docs step

  Scenario: User-facing docs step passes with a real updated doc file
    Given a feature "uf" whose brief declares "surface: user-facing"
    And a non-empty changelog at ".workflow/uf-changelog.md"
    And documentation-specialist evidence whose outputs include the existing file "docs/guides/getting-started.md"
    When docs artifacts are validated for "uf"
    Then validation passes

  Scenario: README.md alone satisfies the real-file rule in a project without docs/guides
    Given a feature "uf" whose brief declares "surface: user-facing"
    And a non-empty changelog at ".workflow/uf-changelog.md"
    And the project has no "docs/guides" directory
    And documentation-specialist evidence whose outputs include the existing file "README.md"
    When docs artifacts are validated for "uf"
    Then validation passes

  Scenario: User-facing docs step fails when no output is a real doc file
    Given a feature "uf" whose brief declares "surface: user-facing"
    And a non-empty changelog at ".workflow/uf-changelog.md"
    And documentation-specialist evidence whose outputs list only ".workflow/uf-documentation-specialist.md"
    When docs artifacts are validated for "uf"
    Then validation fails
    And the error names a required real updated file under "docs/" or "README.md"

  Scenario: Documentation-specialist outputs must be real files
    Given a feature "uf" whose brief declares "surface: user-facing"
    And documentation-specialist evidence whose outputs list the missing path "docs/guides/ghost.md"
    When the documentation-specialist evidence is validated
    Then validation fails
    And the error says actionable outputs must be real files

  Scenario: User-facing docs step fails without a changelog entry
    Given a feature "uf" whose brief declares "surface: user-facing"
    And no file exists at ".workflow/uf-changelog.md"
    When docs artifacts are validated for "uf"
    Then validation fails
    And the error names ".workflow/uf-changelog.md"

  Scenario: Internal docs step keeps the one-line changelog contract
    Given a feature "int" whose brief declares no user-facing surface
    And a non-empty changelog at ".workflow/int-changelog.md"
    And no documentation-specialist evidence exists
    When docs artifacts are validated for "int"
    Then validation passes

  Scenario: docs context prints the curated feature-scale inputs
    Given a feature "f" with a brief, a plan, and a spec file
    And a changelog draft at ".workflow/f-changelog.md"
    When I run "centinela docs context f"
    Then the command exits 0
    And stdout contains a "Feature brief" section with the brief content
    And stdout contains a "Plan" section with the plan content
    And stdout contains a "Spec scenarios" section with the spec content
    And stdout contains a "Changelog draft" section with the changelog content

  Scenario: docs context reports every missing required input at once
    Given a feature "f" with a brief but no plan and no spec file
    When I run "centinela docs context f"
    Then the command exits non-zero
    And the error names "docs/plans/f.md"
    And the error names "specs/f.feature"

  Scenario: docs context treats the changelog draft as optional
    Given a feature "f" with a brief, a plan, and a spec file
    And no file exists at ".workflow/f-changelog.md"
    When I run "centinela docs context f"
    Then the command exits 0
    And the "Changelog draft" section suggests "centinela artifact new f changelog"

  Scenario: The HTML pipeline is gone
    When I run "centinela docs generate"
    Then the command fails as an unknown command
    And no code path regenerates "docs/project-docs/index.html" at merge time
