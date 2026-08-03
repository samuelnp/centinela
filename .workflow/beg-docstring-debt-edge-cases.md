# Edge Cases: beg-docstring-debt

## Covered

- The docstring gate over this branch's changed-file set, verified by running
  `centinela docs lint` against the fix: "All 22 exported identifiers across 16
  changed Go file(s) are documented." Before the fix the same command reported
  `Evidence` and `ValidateEvidence` in `internal/orchestration/evidence.go`.
- Documentation-only change: `go build ./...` and `gofmt` clean, and the full
  suite runs unchanged at the validate step. There is no behavior to assert — a
  test pinning the text of a doc comment would assert nothing about the program,
  and the gate is itself the executable check, running inside
  `centinela validate`.

## Residual Risks

- **The gate is scope-dependent, so this can recur.** Any future change touching
  a file with pre-existing undocumented exports surfaces them the same way. That
  is the ratchet working as designed, not a defect; the remedy is to document
  them, and the repo-wide backlog stays tracked as
  `docstring-full-scan-debt-paydown`.
- **Doc comments can go stale.** No gate checks that a comment still describes
  its identifier, only that one exists. The invariants written here (Outputs
  must name real files; ValidateEvidence must never fail open) are enforced by
  tests elsewhere, so drift would be misleading rather than load-bearing.
- **Not a substitute for the deferred paydown.** This documents exactly the two
  identifiers the gate reports on this branch; the remaining measured backlog
  sits outside any changed-file set and is untouched.
