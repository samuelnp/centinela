# version-lock-profile-contract — qa-senior

## Test Inventory

| Tier | File | Scenarios |
|------|------|-----------|
| unit (colocated) | `internal/workflow/state_version_lock_test.go` | the golden shape lock, updated to record `profileContract` |

**No new tests, deliberately.** The change records a decision in the golden
list that the existing test enforces. Adding a second test over the same
reflected field list would assert the same thing twice; the honest coverage
statement is that the lock caught the drift and now passes.

## Red → green evidence

On the merged tree before this change: `--- FAIL: TestWorkflowStructFieldsAreVersionLocked`,
reporting `profileContract` in the struct and absent from the list. After:
`ok github.com/samuelnp/centinela/internal/workflow`. That pair is produced by
the same command CI runs, not by a mock.

## Coverage Gaps

None for this change. No statements were added or removed, so per-package
coverage is unmoved.

## Acceptance Wiring

`centinela.toml` `validate.commands` runs `go test ./... -coverprofile=coverage.out`,
covering every tier plus colocated package tests in one profiled run. Unchanged.

## Regression Guards

The lock is the guard: any future `Workflow` field added without updating the
list — or without the SchemaVersion bump the comment says a meaning-change owes
— fails `centinela validate` locally, in precommit, and in CI.

## Deferred Findings

none

## Handoff

- Next role: validation-specialist (the gatekeeper runs the validate step).
- Note for the verifier: this branch carries all of `durable-workflow-state`,
  its merge of main (which brought `guided-by-default`), and this one-line
  golden-list update. The claim to test is narrow — that the shape record now
  matches the struct and that holding SchemaVersion at 1 is right for an
  additive `omitempty` field.
