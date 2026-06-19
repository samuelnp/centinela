Feature: Language-aware G2 import-graph gate
  As a maintainer of a non-Go project governed by Centinela
  I want the import-graph gate to select a graph provider by language
  So that layer-dependency rules are enforced instead of hard-failing on `go list`

  Scenario: Go repository is enforced via the go provider
    Given a Go module with a forbidden cross-layer import
    And the import_graph gate is enabled with no explicit provider
    When the import_graph gate runs
    Then the gate auto-selects the "go" provider
    And the gate fails reporting the forbidden edge

  Scenario: Node repository is enforced via the node provider
    Given a Node project whose package.json is the active manifest
    And the import_graph gate is configured with provider "node"
    And the dependency-cruiser command emits a forbidden cross-layer edge
    When the import_graph gate runs
    Then the gate fails reporting the forbidden edge

  Scenario: Python repository is enforced via the python provider
    Given a Python package whose pyproject.toml is the active manifest
    And the import_graph gate is configured with provider "python"
    And the import walker emits a forbidden cross-layer edge
    When the import_graph gate runs
    Then the gate fails reporting the forbidden edge

  Scenario: Project with no recognized manifest skips with a warning
    Given a directory with no go.mod, package.json, or pyproject.toml
    And the import_graph gate is enabled with no explicit provider
    When the import_graph gate runs
    Then the gate is reported as a warning
    And the warning explains no provider matched the project
    And validate still exits zero

  Scenario: Custom-script provider enforces an unsupported language
    Given the import_graph gate is configured with provider "script"
    And script_command points at a program emitting the import-graph JSON
    And that program reports a forbidden cross-layer edge
    When the import_graph gate runs
    Then the gate fails reporting the forbidden edge

  Scenario: Missing external tool warns instead of failing
    Given the import_graph gate is configured with provider "node"
    And the dependency-cruiser tool is not installed
    When the import_graph gate runs
    Then the gate is reported as a warning
    And the warning explains the external tool is not installed
    And validate still exits zero

  Scenario: Empty layer matrix warns before any provider is selected
    Given the import_graph gate is enabled with an empty layer matrix
    When the import_graph gate runs
    Then the gate is reported as a warning
    And no provider is selected
