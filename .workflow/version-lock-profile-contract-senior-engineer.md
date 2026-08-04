# version-lock-profile-contract — senior-engineer

## Files Touched

| Path | Reason |
|------|--------|
| `internal/workflow/state_version_lock_test.go` | Record `profileContract` in the golden field list; skip by exportedness rather than by json tag; state why SchemaVersion holds at 1 |
| `internal/workflow/state_version_compat_test.go` | Carry `profileContract` through the legacy round-trip fixture |

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

Version 1 has not shipped yet, so it can still absorb the field — that is the
whole reason, and it is the one recorded beside the list.

An earlier draft of this report argued "additive and back-compat-by-absence",
which a verifier disproved empirically: a real v0.55.6 binary drops
`profileContract` AND `schemaVersion` on a Load→Save round-trip, turning a
guided-pinned workflow strict mid-run. That is a behaviour change, not merely an
incomplete read. No bump could prevent it either, because no released binary
carries the version check at all. The comment now says so explicitly and warns
against reading it as "additive fields are free" once v1 ships.

## Architecture Compliance

- Boundary checks passed: two colocated test files; no production code, no
  imports, no package edges changed.
- G1 file size: `state_version_lock_test.go` 83 lines, `state_version_compat_test.go`
  86 lines (both ≤100).
- G7 outer-layer rule: unaffected.

## Type-Safety Notes

- No type surface changed. The golden list stays `[]string`, now compared
  against fields selected by EXPORTEDNESS with encoding/json's own name
  fallback — deciding on the tag alone let an exported untagged field reach the
  state file while passing the lock. Two exotic shapes still evade it (an
  embedded unexported struct type, and `json:"-,"`); deferred as
  `workflow-shape-lock-marshal-derived`, whose real fix is deriving the list
  from `json.Marshal` output rather than from reflection over tags.

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
