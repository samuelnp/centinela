### QA-Senior Report: merge-steward-auto-dispatch

**Date:** 2026-05-15

#### Summary

Authored a full test suite (44 new test functions across 12 files plus 3
test helper files) covering the spec's 13 scenarios, the edge-case
tester's high-priority proposals, and the two in-place patch
regressions. Coverage-counted tests live in-package
(`internal/worktree` as `package worktree_test`, `cmd/centinela` as
`package main`, `internal/ui` as `package ui`); acceptance tests build
the CLI binary and exercise it end-to-end, mirroring
`tests/acceptance/enforce_actionable_orchestration_evidence_test.go`.
`go test ./... -count=1` is green; `centinela validate` passes all gates
(G1 file size, full suite, 95% coverage — actual **95.1%**).

#### Test Inventory

| Tier | File | Scenarios |
|------|------|-----------|
| unit | `internal/worktree/merge_pending_test.go` | PendingPath; WritePending atomic + no `.tmp` leftover (patch 2); idempotent rewrite (reason replaced not appended, spec #13); LoadPending missing→(nil,nil); corrupt JSON→error |
| unit | `internal/worktree/merge_pending_directive_test.go` | `Directive()` both reasons; unknown-reason still well-formed; ClearPending idempotent |
| unit | `internal/worktree/merge_pending_errors_test.go` | WritePending `.workflow`-blocked; LoadPending path-is-dir read error; ClearPending non-absence removal error; ResolveMerge isDirty git error |
| integration | `internal/worktree/finalize_test.go` | ResolveMerge APPLY+clean→finalize; ESCALATE→blocked+note; invalid evidence→refuse; no-marker→clear error; dirty-main→blocked even on APPLY; corrupt marker→error |
| integration | `internal/worktree/finalize_edge_test.go` | Worktree-gone still finalizes; escalation note includes report+diff; StewardDirective names prompt/feature/resume |
| integration | `internal/ui/render_merge_test.go` | RenderMergeStewardNeeded / RenderMergeEscalated content |
| integration | `cmd/centinela/merge_dispatch_test.go` | runMerge clean-merge clears stale marker (**patch 1 regression**); dispatchSteward writes marker + directive + non-zero |
| integration | `cmd/centinela/merge_continue_test.go` | --continue no-marker→clean error; APPLY→finalize; ESCALATE→stderr+non-zero; missing evidence→refuse |
| integration | `cmd/centinela/merge_evidence_test.go` | readStewardHandoff decode/missing/corrupt; schema-invalid evidence refuses (spec #11) |
| integration | `cmd/centinela/hook_merge_test.go` | Re-emits while marker+no-evidence (spec #4); silent on valid evidence (spec #5, spec-correct); silent no-marker (spec #6); multiple markers; corrupt marker skipped |
| integration | `cmd/centinela/merge_branches_test.go` | runMerge --continue routing; dispatchSteward WritePending error surfaced |
| acceptance | `tests/acceptance/merge_steward_auto_dispatch_test.go` | Clean merge no-dispatch (regression guard, spec #1); text-conflict dispatch (marker+directive naming prompt/feature/resume, spec #2); --continue APPLY finalizes (spec #7); --continue ESCALATE blocks (spec #8) |

Helpers: `internal/worktree/finalize_helper_test.go`,
`cmd/centinela/merge_steward_helper_test.go`,
`tests/acceptance/merge_steward_auto_dispatch_helper_test.go`.

#### Regression Guards (patches under test)

- **Patch 1** — `cmd/centinela/merge.go` `runMerge`: clean merge calls
  `worktree.ClearPending`. Guarded by
  `TestRunMerge_CleanMergeClearsStaleMarker`: seed stale marker → clean
  merge → marker gone, worktree removed, exit 0.
- **Patch 2** — `internal/worktree/merge_pending.go` `WritePending`
  atomic write. Guarded by `TestWritePending_AtomicNoLeftoverTmp`: no
  `.tmp` survives and the final marker is well-formed JSON.
- **Spec-correct decision (NOT changed):** the hook stays silent once
  valid steward evidence exists.
  `TestRunHookMerge_SilentWhenValidEvidencePresent` asserts the
  spec-correct contract (silent on valid evidence); re-emission is
  asserted only while a marker exists AND no valid evidence
  (`TestRunHookMerge_ReEmitsWhileMarkerNoEvidence`). No test asserts the
  hook re-emits on valid ESCALATE evidence.

#### Coverage Gaps

All 13 spec scenarios have an executable assertion (unit/integration
cover scenarios 1–13 at function level; acceptance covers 1, 2, 7, 8
end-to-end through the built binary). No spec scenario is un-asserted.
Residual uncovered code is limited to mid-operation filesystem-failure
branches in `WritePending` (temp write / `os.Rename` failure) and one
`ResolveMerge` teardown error path — not reliably reproducible without
fault injection; total coverage remains above the 95% gate (95.1%).

#### Acceptance Wiring

`centinela.toml`:

```toml
[validate]
commands = [
  "go test ./...",
  "./scripts/check-coverage.sh"
]
```

`go test ./...` recurses into `tests/acceptance/`, so the acceptance
tests run inside `centinela validate`. No change required (verified: the
validate run compiled and passed the acceptance package).

#### Handoff

- Next role: validation-specialist
- Edge-case report: `.workflow/merge-steward-auto-dispatch-edge-cases.md`
  (produced separately by the edge-case-tester subagent)
