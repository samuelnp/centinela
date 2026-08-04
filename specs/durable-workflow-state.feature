Feature: Durable workflow state
  As Centinela, whose every gate reads .workflow/<feature>.json
  I want that file written atomically, versioned, and guarded against stale overwrites
  So that a killed process, a lagging binary, or a concurrent command can never
  destroy or silently truncate a feature's workflow state

  Background:
    Given the workflow state directory ".workflow" exists

  # --- Atomic, durable writes -------------------------------------------------

  Scenario: A killed write leaves the previous state intact
    Given a workflow "alpha" whose state file records current step "code"
    And a process advances "alpha" to step "tests"
    When that process is killed after the replacement bytes are written but before the rename
    Then ".workflow/alpha.json" still parses as valid JSON
    And it still records current step "code"
    And no zero-byte or truncated state file is left on disk

  Scenario: A completed write replaces the state file in one step
    Given a workflow "alpha" whose state file records current step "code"
    When "alpha" is advanced to step "tests" and the save succeeds
    Then ".workflow/alpha.json" records current step "tests"
    And no temporary file remains beside it

  Scenario: An abandoned temporary file is not mistaken for orphaned evidence
    Given a crashed write left a workflow temporary file in ".workflow"
    When "centinela doctor" runs its evidence check
    Then the temporary file is not reported as an orphaned evidence file
    And the evidence check reports "no orphaned evidence temp files"

  Scenario: The replaced state file keeps its readable file mode
    Given a workflow "alpha" whose state file is mode 0644
    When "alpha" is saved again
    Then ".workflow/alpha.json" is still mode 0644

  Scenario: A write that cannot be completed reports the state file path
    Given the workflow state directory cannot be written to
    When a workflow "alpha" is saved
    Then the command fails with an error naming ".workflow/alpha.json"
    And no partial state file is created

  # --- Schema version ---------------------------------------------------------

  Scenario: A versionless legacy workflow file loads unchanged
    Given a state file for "legacy" written before schema versions existed
    And it carries no "schemaVersion" field
    When "legacy" is loaded
    Then the load succeeds
    And the workflow is treated as schema version 1
    And no migration step is required of the operator

  Scenario: A newly started workflow is stamped with the current schema version
    When a workflow "beta" is started
    Then ".workflow/beta.json" carries "schemaVersion" set to 1

  Scenario: A same-version file round-trips without losing fields
    Given a state file for "gamma" at schema version 1 carrying a recorded model route
    When "gamma" is loaded and saved again by the same binary
    Then the recorded model route is still present in the state file
    And every other recorded field is unchanged

  Scenario: A future-version state file is refused on save with an actionable message
    Given a state file for "delta" carrying schema version 99
    When a command loads "delta" and tries to save it
    Then the save is refused
    And the error names ".workflow/delta.json"
    And the error names the schema version the file carries
    And the error names the schema version this binary understands
    And the error says upgrading Centinela is the fix
    And ".workflow/delta.json" is left byte-for-byte unchanged

  Scenario: A future-version state file this binary can still model keeps governing
    Given a state file for "delta" carrying schema version 99
    And its body still parses, so this binary can read its step
    And "delta" is the only workflow in the project
    When the prewrite hook evaluates a write allowed in "delta"'s current step
    Then the write is allowed
    And the hook does not report that no workflow has been started

  Scenario: A future-version state file this binary cannot model refuses the write
    Given a state file for "delta" carrying schema version 99
    And its body does not parse, so this binary cannot read its step
    And "delta" is the only workflow in the project
    When the prewrite hook evaluates a write to a governed file
    Then the write is refused, because ".workflow/*.json" is itself an
      ungoverned write target and passing would let an agent unblock itself by
      writing a future version over its own state file
    And the refusal names upgrading Centinela as the remedy
    And the hook does not report that no workflow has been started, and no
      duplicate workflow is auto-started

  # --- Concurrent writers -----------------------------------------------------

  Scenario: A stale save is refused rather than silently overwriting a newer one
    Given a workflow "epsilon" is loaded by a long-running command
    And a second command loads "epsilon", records a model route, and saves it
    When the long-running command saves its own copy of "epsilon"
    Then that save is refused
    And the error explains that the state file changed since it was read
    And the error tells the operator to re-run the command
    And the recorded model route is still present in the state file

  Scenario: Re-running a refused command after the conflict succeeds
    Given a save of "epsilon" was refused because the state file changed
    When the same command is run again
    Then it loads the current state file
    And its save succeeds
    And both the model route and the step advance are present in the state file

  Scenario: A workflow that was never loaded saves without a conflict check
    Given no state file exists for "zeta"
    When a workflow "zeta" is created and saved
    Then the save succeeds
    And ".workflow/zeta.json" records "zeta" at its first step
