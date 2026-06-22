Feature: spec reconstruction — deterministic Gherkin skeletons from the Inventory
  As a team adopting Centinela on a brownfield codebase that already ran centinela analyze
  I want centinela reconstruct to derive a behavioral spec corpus skeleton from the Inventory
  So that a spec-first repo gets confirmable .feature skeletons and brief stubs with no LLM call and no clobbering of hand-authored specs

  Scenario: A valid inventory reconstructs feature skeletons and brief stubs into the review dir
    Given an inventory with behavioral packages at .workflow/analysis.json
    When the operator runs centinela reconstruct
    Then the exit code is zero
    And at least one specs feature skeleton is written to the review dir
    And at least one docs features brief stub is written to the review dir
    And the skeletons carry TODO confirm markers for behavior the scan cannot know

  Scenario: Every generated feature parses with the spec traceability scenario parser
    Given an inventory with behavioral packages at .workflow/analysis.json
    When the operator runs centinela reconstruct
    Then every generated feature file has a Feature line and at least one Scenario line
    And the spec traceability scenario parser reads each generated file without error

  Scenario: Unknowable behavior is emitted as an explicit TODO confirm and never fabricated
    Given an inventory with a target whose behavior cannot be inferred from structure
    When the operator runs centinela reconstruct
    Then the generated feature contains a Feature line and one TODO confirm scenario stub
    And no Given When or Then line asserts a fabricated concrete behavior

  Scenario: Re-running reconstruct on an unchanged inventory produces byte-identical output
    Given a fixed inventory at .workflow/analysis.json
    When the operator runs centinela reconstruct twice into the same review dir
    Then both runs exit zero
    And every reconstructed file is byte-identical between the two runs

  Scenario: A hand-authored spec is never clobbered and is reported as skipped
    Given an inventory whose target slug already has a hand-authored specs feature file
    When the operator runs centinela reconstruct
    Then the exit code is zero
    And the existing hand-authored specs feature file is left byte-for-byte unchanged
    And the summary reports that target as skipped

  Scenario: Running reconstruct without an inventory fails with guidance and writes nothing
    Given the project directory has no analysis inventory
    When the operator runs centinela reconstruct
    Then the exit code is non-zero
    And the error message tells the operator to run centinela analyze first
    And no feature skeleton or brief stub is written

  Scenario: The summary reports targets selected files written and total TODO markers
    Given an inventory with behavioral packages at .workflow/analysis.json
    When the operator runs centinela reconstruct
    Then the exit code is zero
    And the stdout summary reports the number of targets selected
    And the stdout summary reports the number of files written
    And the stdout summary reports the total count of TODO confirm markers

  Scenario: An empty doc-only inventory selects zero targets and writes no empty feature
    Given an inventory with no behavioral packages
    When the operator runs centinela reconstruct
    Then the exit code is zero
    And the summary reports zero targets selected
    And no empty feature file is written

  Scenario: A polyglot inventory with an empty Go graph still selects manifest and package targets
    Given an inventory whose Go graph is empty and whose targets come from packages and manifests
    When the operator runs centinela reconstruct
    Then the exit code is zero
    And at least one feature skeleton is written for a manifest or package derived target
