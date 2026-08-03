# beg-docstring-debt — senior-engineer

## Files Touched

| Path | Reason |
|------|--------|
| `internal/orchestration/evidence.go` | Doc comments for `Evidence` and `ValidateEvidence`, the two exported identifiers the docstring gate reports on this branch |

## Problem

`docstring-gate` landed on main enforcing at `severity = "fail"`, scoped to the
merge-base changed-file set. `binding-evidence-gates` modifies
`internal/orchestration/evidence.go` (the merge-steward handoff check lives
there), which pulls that file into scope and surfaces two pre-existing
undocumented exports. The gate is a ratchet working exactly as designed: it does
not open the legacy backlog, but it does require that a file you touch be clean.

Fixed by writing the comments — not by narrowing scope, adding a per-file
exception, or lowering severity.

## Architecture Compliance

- Boundary checks passed: documentation only — no identifier, signature, or
  control flow changed; no imports added; no new package edges.
- G1 file size: `evidence.go` remains within the 100-line cap.
- G7 outer-layer rule: unaffected.

## Type-Safety Notes

- No type surface changed. The comments state the invariants the gates depend
  on (Outputs must name real files; ValidateEvidence must never fail open) so
  the next reader learns them from the declaration rather than from a gate
  failure.

## Trade-Offs

- **Document vs. exempt.** The gate supports per-file exceptions; using one here
  would have recorded "this file may stay undocumented" for two identifiers that
  are central to how completion is judged. Two comments is cheaper and honest.
- **Scope.** Only the two identifiers the gate reports are documented. The
  repo-wide backlog stays with `docstring-full-scan-debt-paydown`, which is
  where the docstring-gate feature deliberately put it.

## Handoff

- Next role: qa-senior.
- Outstanding: none. `centinela docs lint` reports all 22 exported identifiers
  across the 16 changed Go files documented; build and gofmt are clean. There is
  no behavior change to test — the gate itself is the verification, and it runs
  in `centinela validate` at the validate step.
