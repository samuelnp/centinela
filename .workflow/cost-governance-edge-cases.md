# Edge Cases: cost-governance

## Covered

| # | Edge case | Handling | Test |
|---|-----------|----------|------|
| 1 | Malformed transcript line | skipped, not fatal | `cost:TestSumFromSumsAndSkipsGarbage` |
| 2 | Missing transcript file | no-op, zero, no error | `cost:TestSumFromMissingFileIsNoOp`, `acceptance:TestAccCostMissingTranscriptNoOp` |
| 3 | Repeated capture (same transcript) | cursor at EOF → only the delta counts | `cost:TestSumFromOffsetReadsOnlyDelta`, `hook_cost:TestHookCostCapturesAndAttributes` |
| 4 | Transcript truncated/rotated (offset > size) | recount from start | `cost:TestSumFromTruncationResetsToStart` |
| 5 | New session (different transcript path) | cursor resets to 0 | `cost:TestOffsetForMatchesPathOnly` |
| 6 | Malformed cursor file | read as zero | `cost:TestLoadCursorMalformedIsZero` |
| 7 | Zero budget for a scope | never reports over | `cost:TestBudgetStatusOverUnderAndRemaining`, `unit:TestCostBudgetSoftGate` |
| 8 | Over-budget step | validate exits 0 with ⚠ (soft) | `acceptance:TestAccCostCaptureReportAndSoftGate`, `cmd:TestEmitCostWarningOverBudget` |
| 9 | Within budget | silent | `cmd:TestEmitCostWarningNotOverIsSilent` |
| 10 | `[cost]` disabled / unconfigured | capture + warning are no-ops | `cmd:TestHookCostDisabledIsNoOp`, `config:TestCostIsActive` |
| 11 | No transcript_path on stdin | no sample recorded | `cmd:TestHookCostNoTranscriptIsNoOp` |
| 12 | No active feature (empty repo) | resolver returns nil → no-op | `cmd:TestActiveWorkflowNoneActive` |
| 13 | Non-worktree (root) mode | active feature resolved from .workflow | `cmd:TestActiveWorkflowRootMode` |
| 14 | Zero token delta | RecordCostSample writes nothing | `telemetry:TestRecordCostSampleZeroIsNoOp` |
| 15 | Old telemetry lines (no token fields) | read as 0 (omitempty back-compat) | covered by `telemetry.Read` leniency |

## Residual Risks

- Cross-feature attribution within a single Stop is approximate (transcript is
  per-session); the byte cursor prevents double-counting but a Stop spanning two
  features attributes the delta to whichever is active at fire time. Acceptable
  for a visibility-only soft gate.
- Tier budgets are keyed by model id (tier-name resolution deferred to a later
  feature); v1's soft gate evaluates step + feature budgets.
- Defensive I/O error branches (SaveCursor MkdirAll/Marshal, SumFrom open/seek
  failures) are non-deterministic to trigger and left uncovered by design; all
  return safe zero/no-op values.
