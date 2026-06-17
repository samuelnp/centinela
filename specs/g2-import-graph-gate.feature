Feature: G2 Import-Graph Gate
  As a maintainer or agent working in a Centinela-governed Go project
  I want `centinela validate` to mechanically check the import graph against
  a configured per-layer allow matrix
  So that forbidden cross-layer imports fail validation instead of relying on
  human review

  Background:
    Given the project has a `centinela.toml` at its root
    And the project contains Go source files organized into layers

  # ---------------------------------------------------------------------------
  # Happy path — clean import graph
  # ---------------------------------------------------------------------------

  Scenario: All imports respect the layer matrix — gate passes
    Given `centinela.toml` contains a `[gates.import_graph]` block with `enabled = true`
    And the block defines layers:
      | layer    | paths                          | allow    |
      | config   | internal/config/**             |          |
      | domain   | internal/workflow/**           | config   |
      | domain   | internal/gates/**              | config   |
      | ui       | internal/ui/**                 | domain   |
      | cmd      | cmd/**                         | domain   |
    And every Go package imports only packages in its layer's allow list
      or packages in the same layer
    When `centinela validate` runs
    Then the gate result named `import_graph` is `Pass`
    And the output contains "import_graph" with status "PASS"
    And the exit code is 0

  # ---------------------------------------------------------------------------
  # Negative path — forbidden cross-layer edge
  # ---------------------------------------------------------------------------

  Scenario Outline: A package imports a layer it is not allowed to import — gate fails
    Given `centinela.toml` contains a valid `[gates.import_graph]` block
    And layer "<importer_layer>" does not include "<imported_layer>" in its allow list
    And the package "<importer_pkg>" (in layer "<importer_layer>") imports "<imported_pkg>" (in layer "<imported_layer>")
    When `centinela validate` runs
    Then the gate result named `import_graph` is `Fail`
    And the output contains the violating edge formatted as:
      "<importer_pkg> -> <imported_pkg> (<importer_layer> may not import <imported_layer>)"
    And the exit code is 1

    Examples:
      | importer_pkg                          | importer_layer | imported_pkg                       | imported_layer |
      | github.com/samuelnp/centinela/internal/config | config | github.com/samuelnp/centinela/internal/ui  | ui             |
      | github.com/samuelnp/centinela/internal/config | config | github.com/samuelnp/centinela/internal/workflow | domain  |
      | github.com/samuelnp/centinela/internal/ui     | ui     | github.com/samuelnp/centinela/cmd/centinela | cmd          |

  Scenario: Multiple forbidden edges are all listed in the failure output
    Given `centinela.toml` contains a valid `[gates.import_graph]` block
    And two packages each import a package in a forbidden layer
    When `centinela validate` runs
    Then the gate result is `Fail`
    And the output contains one details line per violating edge
    And the exit code is 1

  # ---------------------------------------------------------------------------
  # Gate disabled / no config block → gate omitted
  # ---------------------------------------------------------------------------

  Scenario: No `[gates.import_graph]` block present — gate is omitted
    Given `centinela.toml` does not contain a `[gates.import_graph]` block
    When `centinela validate` runs
    Then no gate result named `import_graph` appears in the output
    And the exit code reflects only the other gates' results

  Scenario: Gate explicitly disabled with `enabled = false` — gate is omitted
    Given `centinela.toml` contains a `[gates.import_graph]` block with `enabled = false`
    When `centinela validate` runs
    Then no gate result named `import_graph` appears in the output
    And the exit code reflects only the other gates' results

  # ---------------------------------------------------------------------------
  # Unmapped package → Warn
  # ---------------------------------------------------------------------------

  Scenario: A package matches no configured layer — gate warns
    Given `centinela.toml` contains a valid `[gates.import_graph]` block
    And at least one Go package in the module matches no layer path glob
    And no forbidden imports are present among the mapped packages
    When `centinela validate` runs
    Then the gate result named `import_graph` is `Warn`
    And the output identifies the unmapped package(s) by import path
    And the exit code is 0

  # ---------------------------------------------------------------------------
  # Malformed config → Fail with a config-error message
  # ---------------------------------------------------------------------------

  Scenario Outline: Malformed `[gates.import_graph]` config — gate fails with config error
    Given `centinela.toml` contains a `[gates.import_graph]` block with `enabled = true`
    And the config is malformed because <malformation>
    When `centinela validate` runs
    Then the gate result named `import_graph` is `Fail`
    And the output contains a message starting with "import_graph config:"
    And the config-error message does NOT contain the import-violation arrow format "→"
    And the exit code is 1

    Examples:
      | malformation                                                       |
      | a layer has an empty `paths` list                                  |
      | an allow-list references a layer name not defined in the config    |
      | the module path is set to an empty string                          |

  # ---------------------------------------------------------------------------
  # Empty matrix (block present but no layers) → Warn (not a silent Pass)
  # ---------------------------------------------------------------------------

  Scenario: Block present with no layers defined — gate warns rather than silently passing
    Given `centinela.toml` contains a `[gates.import_graph]` block with `enabled = true`
    And the block defines zero layers
    When `centinela validate` runs
    Then the gate result named `import_graph` is `Warn`
    And the output indicates the layer matrix is empty
    And the exit code is 0

  # ---------------------------------------------------------------------------
  # Load error / uncompilable code → Fail (never a false Pass)
  # ---------------------------------------------------------------------------

  Scenario: The module contains uncompilable code — gate fails with load error
    Given `centinela.toml` contains a valid `[gates.import_graph]` block
    And the Go module contains a syntax error that prevents `go/packages` from loading
    When `centinela validate` runs
    Then the gate result named `import_graph` is `Fail`
    And the output contains the load error message from `go/packages`
    And the exit code is 1

  # ---------------------------------------------------------------------------
  # Standard-library and third-party imports are ignored
  # ---------------------------------------------------------------------------

  Scenario: A package imports standard-library and third-party packages — not flagged
    Given `centinela.toml` contains a valid `[gates.import_graph]` block
    And a package imports "fmt", "os", and "golang.org/x/tools/go/packages"
    And none of those are in the module's own path prefix
    When `centinela validate` runs
    Then those imports are not evaluated against the layer matrix
    And the gate result is `Pass` (assuming no internal violations exist)

  # ---------------------------------------------------------------------------
  # Test files — _test.go packages map to the package-under-test's layer
  # ---------------------------------------------------------------------------

  Scenario: An external test package (_test suffix) imports across a forbidden layer boundary
    Given `centinela.toml` contains a valid `[gates.import_graph]` block
    And an external test package "github.com/samuelnp/centinela/internal/config_test"
      is mapped to the "config" layer (same as the package under test)
    And that test package imports a package in the "ui" layer which is forbidden for "config"
    When `centinela validate` runs
    Then the gate result named `import_graph` is `Fail`
    And the violating edge lists the test package as the importer

  # ---------------------------------------------------------------------------
  # Intra-layer imports are always allowed (self-import)
  # ---------------------------------------------------------------------------

  Scenario: A package imports another package in the same layer — always allowed
    Given `centinela.toml` contains a valid `[gates.import_graph]` block
    And layer "domain" contains both "internal/workflow" and "internal/gates"
    And "internal/workflow" imports "internal/gates"
    When `centinela validate` runs
    Then that intra-layer import is not flagged as a violation
    And the gate result is `Pass` (assuming no other violations)

  # ---------------------------------------------------------------------------
  # Diff-aware filter is ignored — whole-module load
  # ---------------------------------------------------------------------------

  Scenario: A violation exists outside the current diff set — gate still fails
    Given `centinela.toml` contains a valid `[gates.import_graph]` block
    And a forbidden import edge exists in a file NOT in the current git diff
    When `centinela validate` runs with a diff filter active
    Then the gate result named `import_graph` is `Fail`
    And the violating edge is reported regardless of whether the file is in the diff
