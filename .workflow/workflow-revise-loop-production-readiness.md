### Production Readiness Report: workflow-revise-loop
**Date:** 2026-06-30
**Status:** WARNING

#### Files Reviewed
- `internal/workflow/rewind.go` — `Revision` type, `RewindTo`, `RevisionsSummary`, `reopenedSteps`
- `internal/workflow/state.go` — `Revisions` field on `Workflow`, `Save` (shared persistence)
- `internal/evidence/invalidate.go` — `Invalidate`, `InvalidateArtifact`, `removeBoth`
- `internal/evidence/invalidation_targets.go` — `InvalidationTargets` policy
- `cmd/centinela/revise.go` — `reviseCmd`, `runRevise`
- `cmd/centinela/revise_invalidate.go` — `invalidateDownstream`
- `internal/telemetry/event.go` — `TypeStepRevised`, `From` field
- `internal/telemetry/constructors.go` — `RecordRevised`
- `internal/ui/render_status.go` — Revisions row rendering
- (context) `cmd/centinela/telemetry_model.go`, `internal/telemetry/record.go`, `cmd/centinela/complete.go`

#### Findings
| Check | File | Severity | Issue | Suggested Fix |
|-------|------|----------|-------|---------------|
| C3 | internal/workflow/state.go | WARNING | `Save` persists via plain `os.WriteFile` (no write-temp-then-rename). A crash mid-write of `.workflow/<feature>.json` can leave a truncated/malformed state file that breaks the step state machine. Pre-existing shared infra (`saveWorkflow = workflow.Save`, also used by `complete`), not introduced here — but `revise` deletes evidence files immediately before this save, so a corrupt save is slightly more consequential on this path. | Make `workflow.Save` atomic: write to a temp file in `.workflow/` then `os.Rename` over the target. Benefits every command, not just revise. Track via `centinela start workflow-revise-loop-hardening`. |
| C1 | cmd/centinela/revise.go | INFO (not a finding) | `_ = reviseCmd.MarkFlagRequired("to"/"reason")` discards the returned error. Acceptable: cobra returns an error only when the named flag does not exist, i.e. a programmer error caught at `init()`, never a runtime condition. Both flags are registered two lines above. | No change required. |

#### Summary
CRITICAL: 0, WARNING: 1

#### Recommendation
The feature is robust on the checks that matter for a destructive, state-mutating command. **C1 error handling** is sound: every state-mutating or file-deleting path checks its error — `removeBoth` surfaces non-absence `os.Remove` failures with the offending path and treats absence as idempotent success, `invalidateDownstream` propagates the first failure, and the `saveWorkflow` error is wrapped. The discarded `MarkFlagRequired` errors are programmer-error-only and acceptable. **C4 input validation** is strong and defence-in-depth: `--to`/`--reason` are cobra-required, empty reason is rejected at both the cmd boundary and inside `RewindTo`, unknown steps are rejected naming the valid set, and the transition is enforced strictly backward and refused on a `done` workflow. **C3 atomicity / crash-ordering** is deliberately fail-safe: `RewindTo` validates fully before mutating any in-memory state (all-or-nothing), and `runRevise` deletes evidence *before* it persists the rewound state — so a crash between deletion and save leaves evidence gone but the on-disk state un-rewound, which the next `complete`/`validate` detects as missing evidence and forces a gate re-run (the operation is idempotent on re-invocation). The dangerous inverse — state rewound on disk while stale evidence survives, which could let a re-opened step pass on old certification — is structurally impossible because the save is last. **Invalidation scope** is safe: `Invalidate`/`InvalidateArtifact` only ever touch `.workflow/<feature>-*` paths; source, tests, and docs are never removed. **C2** (the lone telemetry file handle defers `Close`, no goroutines) and **C5** (no secrets) pass. The single WARNING is the non-atomic `workflow.Save`: it is pre-existing shared infrastructure rather than a regression in this feature, so it does not block, but hardening it to write-temp-then-rename is recommended via `centinela start workflow-revise-loop-hardening`.
