# version-lock-profile-contract — senior-engineer

## Files Touched

| Path | Reason |
|------|--------|
| `internal/workflow/state_version_lock_test.go` | Record `profileContract` in the golden field list, with the reasoning for holding SchemaVersion at 1 |

## Problem

`guided-by-default` merged to main while `durable-workflow-state` was in flight
and added `ProfileContract string \`json:"profileContract,omitempty"\`` to the
`Workflow` struct. `durable-workflow-state` shipped `TestWorkflowStructFieldsAreVersionLocked`,
a golden list of every JSON key the struct models, which fails when the shape
changes without a deliberate decision.

Neither PR could see this: the test does not exist on main, and the field does
not exist on the branch. It fails only on the merged tree — the parallel-merge
semantic conflict this repo has hit before, and the reason the full gate is run
on the merged tree before pushing.

The test did exactly its job. This change is the decision it demanded, made
explicitly rather than by silently re-recording whatever the struct happens to
contain.

## The decision: no SchemaVersion bump

`profileContract` is `omitempty` and back-compat-by-absence: a file without the
key loads with the zero value, which is the same thing the field means when
unset. The documented migration contract in `schema_version.go` is that
defaulting IS the migration for additive fields, so version 1 still describes
this shape honestly.

A bump is owed by a change that alters the MEANING of an existing key or removes
one — where an older binary reading the file would be *wrong* rather than merely
incomplete. That is the line recorded in the comment beside the list, so the
next person facing this question has the rule rather than a precedent to copy.

## Architecture Compliance

- Boundary checks passed: one colocated test file; no production code, no
  imports, no package edges changed.
- G1 file size: `state_version_lock_test.go` is 66 lines (≤100).
- G7 outer-layer rule: unaffected.

## Type-Safety Notes

- No type surface changed. The golden list stays `[]string` compared against
  reflected struct tags, so it cannot drift from the real marshalled shape
  without failing.

## Trade-Offs

- **Record vs. bump.** Bumping to 2 would be the conservative reflex, but it
  would be dishonest: nothing about reading a v1 file is unsafe for this shape,
  and a version that moves on every additive field teaches readers to ignore it.
- **Comment placement.** The rule lives beside the list rather than in the
  commit message, because the next person to hit this reads the list.

## Handoff

- Next role: qa-senior.
- Outstanding: none. `go build ./...`, `gofmt`, and the version-lock test pass;
  the full suite runs at the validate step. No behavior changed, so there is
  nothing new to test — the test that caught this is the coverage.
