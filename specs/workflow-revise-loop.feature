Feature: Workflow revise loop
  centinela revise <feature> --to <step> --reason "<why>" provides a
  controlled, auditable backward transition. Rewinding re-opens every
  downstream step to pending, deletes only those steps' certification
  evidence (.workflow/<feature>-<role>.{json,md} and -edge-cases.md),
  and records a Revision entry on the workflow state. Source, test, and
  docs files are never touched. The next centinela complete is then forced
  to re-run the gates on the corrected tree before re-advancing.

  # Acceptance: workflow-revise-loop
  # Scenario mapping to tests/acceptance/workflow_revise_loop_test.go

  Scenario: Happy path — validate step is revised back to code
    # Scenario: happy-path-validate-to-code
    Given a canonical feature "my-feature" whose current step is validate
    And the feature has senior-engineer evidence at .workflow/my-feature-senior-engineer.json
    And the feature has gatekeeper evidence at .workflow/my-feature-gatekeeper.md
    And the feature has validation-specialist evidence at .workflow/my-feature-validation-specialist.json
    And the feature source file at internal/myfeature/handler.go exists
    When the user runs: centinela revise my-feature --to code --reason "bug found in handler"
    Then the command exits 0
    And the current step of my-feature is code
    And the step code is in-progress
    And the steps tests, validate, and docs are pending
    And .workflow/my-feature-gatekeeper.json does not exist
    And .workflow/my-feature-gatekeeper.md does not exist
    And .workflow/my-feature-validation-specialist.json does not exist
    And .workflow/my-feature-validation-specialist.md does not exist
    And .workflow/my-feature-edge-cases.md does not exist
    And .workflow/my-feature-senior-engineer.json still exists
    And internal/myfeature/handler.go still exists
    And the workflow state contains a revision from validate to code with reason "bug found in handler"

  Scenario: Re-gating — complete at re-opened step is blocked until evidence is regenerated
    # Scenario: re-gating-blocks-advance-without-evidence
    Given a canonical feature "my-feature" whose current step is code after a rewind from validate
    And .workflow/my-feature-qa-senior.json does not exist
    When the user runs: centinela complete my-feature
    Then the command exits non-zero
    And the error output references missing qa-senior evidence
    And the current step of my-feature remains code

  Scenario: Re-gating — complete advances once evidence is regenerated
    # Scenario: re-gating-advances-after-evidence-regenerated
    Given a canonical feature "my-feature" whose current step is code after a rewind from validate
    And the senior-engineer evidence has been regenerated at .workflow/my-feature-senior-engineer.json
    And all gates for the code step pass
    When the user runs: centinela complete my-feature
    Then the command exits 0
    And the current step of my-feature advances past code

  Scenario: Negative — revise without --reason is rejected
    # Scenario: missing-reason-rejected
    Given a canonical feature "my-feature" whose current step is validate
    When the user runs: centinela revise my-feature --to code
    Then the command exits non-zero
    And the error output indicates --reason is required
    And the current step of my-feature remains validate
    And no workflow state is mutated

  Scenario: Negative — empty or whitespace-only --reason is rejected
    # Scenario: whitespace-reason-rejected
    Given a canonical feature "my-feature" whose current step is validate
    When the user runs: centinela revise my-feature --to code --reason "   "
    Then the command exits non-zero
    And the error output indicates reason must not be empty
    And the current step of my-feature remains validate

  Scenario: Negative — revise to a forward step is rejected
    # Scenario: forward-target-rejected
    Given a canonical feature "my-feature" whose current step is code
    When the user runs: centinela revise my-feature --to tests --reason "jump forward"
    Then the command exits non-zero
    And the error output indicates the target step is not strictly before the current step
    And the current step of my-feature remains code

  Scenario: Negative — revise to the current step is rejected
    # Scenario: equal-target-rejected
    Given a canonical feature "my-feature" whose current step is validate
    When the user runs: centinela revise my-feature --to validate --reason "same step"
    Then the command exits non-zero
    And the error output indicates the target step is not strictly before the current step
    And the current step of my-feature remains validate

  Scenario: Negative — revise to an unknown step name is rejected
    # Scenario: unknown-step-rejected
    Given a canonical feature "my-feature" whose current step is validate
    When the user runs: centinela revise my-feature --to deploy --reason "unknown"
    Then the command exits non-zero
    And the error output names "deploy" as an unrecognised step
    And the current step of my-feature remains validate

  Scenario: Negative — revising a done workflow is rejected
    # Scenario: done-workflow-rejected
    Given a canonical feature "my-feature" whose current step is done
    When the user runs: centinela revise my-feature --to validate --reason "reopen"
    Then the command exits non-zero
    And the error output states that a completed workflow cannot be revised
    And no workflow state is mutated

  Scenario: Audit — revision count and reason are visible in status
    # Scenario: audit-visible-in-status
    Given a canonical feature "my-feature" that has been revised twice
    And the most recent revision had reason "second rewind"
    When the user runs: centinela status my-feature
    Then the output contains "Revisions" with the count 2
    And the output contains the reason "second rewind" inline

  Scenario: Archetype-awareness — rewind respects hotfix step order
    # Scenario: archetype-hotfix-order-respected
    Given a hotfix feature "hotfix-one" with step order code, tests, validate
    And the current step is validate
    When the user runs: centinela revise hotfix-one --to code --reason "hotfix regression"
    Then the command exits 0
    And the current step of hotfix-one is code
    And the steps tests and validate are pending
    And no plan or docs step is present in the workflow state

  Scenario: Safety — invalidation never touches source, test, or docs files
    # Scenario: safety-no-source-deletion
    Given a canonical feature "my-feature" whose current step is validate
    And the following files exist:
      | path                                      |
      | internal/myfeature/service.go             |
      | tests/unit/myfeature/service_test.go      |
      | docs/features/workflow-revise-loop.md     |
      | docs/plans/workflow-revise-loop.md        |
    When the user runs: centinela revise my-feature --to code --reason "safety check"
    Then the command exits 0
    And all of the above files still exist unchanged
    And only .workflow/my-feature-* evidence files were removed

  Scenario: Idempotency — invalidating already-absent evidence is not an error
    # Scenario: idempotent-invalidation
    Given a canonical feature "my-feature" whose current step is validate
    And .workflow/my-feature-gatekeeper.json does not exist
    And .workflow/my-feature-gatekeeper.md does not exist
    When the user runs: centinela revise my-feature --to code --reason "idempotent"
    Then the command exits 0
    And no error is reported about missing evidence files

  Scenario: Multiple rewinds accumulate in the revision log
    # Scenario: multiple-rewinds-accumulate
    Given a canonical feature "my-feature" that has been revised once from validate to code with reason "first rewind"
    And the feature has since been advanced back to validate
    When the user runs: centinela revise my-feature --to code --reason "second rewind"
    Then the command exits 0
    And the workflow state contains 2 revision entries
    And the first entry has reason "first rewind"
    And the second entry has reason "second rewind"

  Scenario: Internal feature code-step invalidation excludes ux-ui-specialist
    # Scenario: internal-feature-no-ux-invalidation
    Given an internal CLI feature "my-feature" whose current step is tests
    And .workflow/my-feature-ux-ui-specialist.json does not exist
    When the user runs: centinela revise my-feature --to code --reason "internal fix"
    Then the command exits 0
    And the current step of my-feature is code
    And no ux-ui-specialist evidence is referenced in the invalidation output
    And .workflow/my-feature-qa-senior.json is invalidated
    And .workflow/my-feature-senior-engineer.json is preserved because code is the target step

  Scenario: Re-opened step CompletedAt is cleared
    # Scenario: completed-at-cleared-on-reopen
    Given a canonical feature "my-feature" whose tests step has a non-null CompletedAt timestamp
    And the current step is validate
    When the user runs: centinela revise my-feature --to code --reason "clear timestamps"
    Then the command exits 0
    And the tests step CompletedAt is null in the workflow state
    And the validate step CompletedAt is null in the workflow state
    And the code step CompletedAt is null in the workflow state
