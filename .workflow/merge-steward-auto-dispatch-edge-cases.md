# Edge-Case Report: merge-steward-auto-dispatch

**Date:** 2026-05-15

---

## Risk Matrix

| # | Case | Impact | Likelihood | Why |
|---|------|--------|------------|-----|
| 1 | Corrupt/unparseable marker JSON silently blocks merge indefinitely | High | Medium | Hook skips on LoadPending error; operator has no recovery path (no cleanup flag); `--continue` also fails but marker persists |
| 2 | --continue with unknown `handoffTo` value (typo: "compleet") silently escalates instead of erroring | Medium | Medium | readStewardHandoff accepts any non-empty string; ValidateEvidence only checks non-empty, not enum; typo hides as ESCALATE |
| 3 | Concurrent `centinela merge` on same feature causes marker file corruption | High | Low | WritePending uses direct os.WriteFile, not atomic write; race condition on concurrent processes |
| 4 | Marker exists but worktree deleted out-of-band; --continue clears marker but branch leaks | Medium | Low | ResolveMerge calls Remove (idempotent if worktree gone) but never DeleteBranch; inconsistent state |
| 5 | Re-run merge while marker exists, now merge succeeds; old ESCALATE marker lingers in .workflow/ | Medium-High | Medium | runMerge doesn't clear pending marker on clean merge; hook re-emits stale directive; steward invoked unnecessarily |
| 6 | Marker written then --continue run from subdirectory; marker not found (silent state loss) | High | Low | PendingPath is cwd-relative; no repo-root detection; marker written in wrong .workflow/ |
| 7 | Hook re-emission stops on valid ESCALATE evidence; operator has no signal merge is still blocked | Medium | Medium | Hook checks `validator(feature) succeeds` but doesn't distinguish APPLY/ESCALATE; silent on ESCALATE handoffTo |
| 8 | Marker for feature whose worktree was manually deleted; directive re-emitted but steward can't work | Medium-High | Low | Hook globs markers without checking Exists(worktree); steward finds missing worktree; bad UX or crash |
| 9 | --continue with APPLY evidence but main tree dirty; operator `git reset --hard` during cleanup, loses changes | Medium-High | Low | isDirty check blocks finalize (correct), but error message unclear; operator may corrupt state during stash/reset |
| 10 | Marker Directive() re-renders invalid reason (e.g., "unknown-reason") producing malformed directive string | Low | Very Low | Directive() assumes reason is valid (git-text-conflict or post-merge-validate-failed); no validation if neither |
| 11 | Evidence file missing `handoffTo` field; ValidateEvidence catches it but future relaxation would escalate silently | Low | Very Low | readStewardHandoff accepts empty string from unmarshal; current orchestration validator prevents it but brittle |
| 12 | Evidence schema incomplete (missing outputs); error message unclear which field is missing | Low | Low | orchestration.ValidateEvidence returns "incomplete evidence fields" without field details; not worktree-specific |
| 13 | Marker exists for escalated merge; operator fixes code and re-runs merge; old evidence causes re-escalation | Medium | Medium | ResolveMerge keeps evidence file on ESCALATE; re-running `--continue` re-reads old evidence; must manually delete or restart |
| 14 | --continue with no marker; error message doesn't suggest recovery ("run centinela merge <feature>") | Low | Low | Error is clear but lacks recovery hint; UX friction |
| 15 | Marker re-written during steward work (reason changes text-conflict → post-merge-validate-failed); steward operating on stale reason | Low | Medium | WritePending idempotent rewrite; if merge re-runs and hits different failure, marker reason changes; steward may miss shift |

**Matrix Total: 15 risks, 12 with Impact ≥ Medium**

---

## Missing or Weak Scenarios

### From Spec (merge-steward-auto-dispatch.feature, 13 scenarios)

- **Scenario 6 (hook re-emit)**: Spec checks that directive is re-emitted while marker exists without valid evidence. But spec does NOT cover: ESCALATE evidence present → should hook re-emit or go silent? Current impl goes silent (BUG—issue #7 above). Proposed fix: hook must re-emit on ESCALATE.

- **Scenario 11 (invalid evidence)**: Spec covers schema-invalid evidence (validator rejects). But spec does NOT cover: valid JSON schema but unknown `handoffTo` value (e.g., "retry" or typo "compleet"). Current code treats non-"complete" as ESCALATE without validating it's "user". Proposed: add enum check.

- **Scenario 13 (dirty main on APPLY)**: Spec covers dirty-main blocks finalize. But spec does NOT cover: operator runs `git reset --hard` during cleanup and corrupts the merge state. No guard. Proposed: clearer error message or pre-check.

- **Concurrent merges**: Spec is silent on concurrent `centinela merge` on the same feature. Current impl has race on WritePending. Proposed: atomic writes or mutex.

- **Marker + worktree out-of-sync**: Spec has "worktree kept on conflict" but does NOT cover: worktree deleted out-of-band while marker exists. Hook will re-emit, steward cannot work. Proposed: hook checks Exists and emits warning if gone.

- **Clean merge after stalled merge**: Spec scenario 1 (clean merge) and scenario 13 (marker rewritten) are separate. But spec does NOT cover: clean merge succeeds and old marker from prior conflict lingers. Proposed: runMerge must ClearPending on success.

### New Scenarios Not Yet Tested

- **Corrupt marker JSON**: LoadPending fails, hook skips, `--continue` fails, marker persists forever. No recovery.
- **readStewardHandoff missing field**: Evidence valid schema but `handoffTo` absent; unmarshals to ""; current code escalates, future code might not error.
- **Subdirectory cwd**: Marker written in wrong `.workflow/`; --continue doesn't find it.
- **ESCALATE evidence then retry**: Evidence file persists; re-running `--continue` on same stale evidence re-escalates. Operator confused about recovery path.

---

## Proposed/Added Tests

### Unit Tests

**Target file: `internal/worktree/merge_pending_test.go` (new)**
- `TestLoadPending_CorruptJSON_Error`: Corrupt marker JSON returns error, not nil.
- `TestWritePending_Concurrent_NoCorruption`: Two concurrent WritePending calls don't corrupt the file (use sync.Mutex or atomic writes).
- `TestPendingMarker_Directive_InvalidReason_Safe`: Directive() with invalid reason does not produce malformed string.

**Target file: `cmd/centinela/merge_evidence_test.go` (new)**
- `TestReadStewardHandoff_InvalidValue_Error`: readStewardHandoff with `handoffTo: "compleet"` returns enum error.
- `TestReadStewardHandoff_MissingField_Error`: readStewardHandoff with missing `handoffTo` returns error, not "".

**Target file: `internal/worktree/finalize_test.go` (new)**
- `TestResolveMerge_CorruptPending_Error`: ResolveMerge with corrupt marker returns error.
- `TestResolveMerge_DirtyMain_RefusesEvenWithAPPLY`: isDirty check blocks finalize before evidence validation.
- `TestResolveMerge_WorktreeAlreadyGone_ClearsMarkerAndExits`: Remove is idempotent; marker cleared even if worktree gone.
- `TestResolveMerge_NoMarker_ClearError`: --continue with no marker returns actionable error.
- `TestResolveMerge_EvidenceIncompleteSchema_Refuses`: Evidence valid JSON but missing outputs → error.

**Target file: `cmd/centinela/hook_merge_test.go` (new)**
- `TestHookMerge_CorruptMarker_SkipsAndContinues`: Hook doesn't crash on corrupt marker; continues to next.
- `TestHookMerge_EscalateEvidence_ContinuesEmitting`: Hook does NOT silence on valid ESCALATE evidence; re-emits.
- `TestHookMerge_WorktreeGone_AlertsOperator`: Hook detects worktree missing and emits warning.
- `TestHookMerge_MultipleMarkers_EmitsAllDirectives`: Hook handles multiple pending markers (one per feature).

**Target file: `cmd/centinela/merge_steward_test.go` (new/expanded)**
- `TestRunMerge_CleanAfterPriorConflict_ClearsPendingMarker`: Clean merge clears stale marker.

### Integration Tests

**Target file: `cmd/centinela/merge_test.go` (expanded)**
- `TestRunMergeContinue_FromSubdir_FindsMarker`: Run `centinela merge --continue` from subdirectory; marker found (requires repo-root detection).
- `TestRunMergeContinue_DirtyMainCleanup_GuardAgainstLoss`: Main dirty on `--continue`; error message guides safe cleanup.
- `TestRunMergeContinue_APPLYWorktreeGone_StillFinalizes`: Evidence valid APPLY but worktree already removed; finalize still succeeds, branch cleanup documented.

**Target file: `cmd/centinela/merge_steward_test.go` (new)**
- `TestHookAndContinue_EscalateWorkflow`: Text-conflict → marker + hook re-emit → steward escalates → hook still re-emits → operator re-runs merge → clean success.

### Acceptance Tests

**Target file: `integration_test.go` or `acceptance_test.go` (new)**
- **Scenario: Concurrent conflicts (both feature branches conflict)**: Two parallel `centinela merge` calls on different features; no marker corruption.
- **Scenario: Marker + worktree out-of-sync**: Text conflict → marker + worktree kept. Operator deletes worktree out-of-band. Hook runs → warns worktree missing. Operator manually restarts merge.
- **Scenario: ESCALATE then retry**: Feature hits conflict, steward escalates. Operator fixes and re-runs `centinela merge <feature>` (fresh, not --continue). Marker re-written, evidence cleared, merge re-attempted.
- **Scenario: Stale evidence recovery**: Feature escalates. Old evidence in `.workflow/`. Operator forgets to clean; runs `--continue` → re-escalates. Error message guides: "Delete evidence or restart with `centinela merge <feature>`."

---

## Residual Risks

### Risk #1: Race Condition on WritePending (High Impact)
**Status:** Not fully mitigated by tests alone. Requires code fix.
**Mitigation:**
- Implement atomic writes: write to `.workflow/<feature>-merge-pending.json.tmp`, then `os.Rename()` to final path.
- Add test with goroutines to catch race under `-race` flag.
- Document: "centinela merge is designed for sequential invocation per feature; concurrent calls are unsupported."

### Risk #2: Hook Silent on ESCALATE (Medium Impact)
**Status:** Bug in current implementation.
**Mitigation:**
- Enhance hook logic: after validating evidence, read `handoffTo` value from evidence JSON.
- If `handoffTo == "user"` (ESCALATE), continue re-emitting.
- If `handoffTo == "complete"` (APPLY), go silent.
- Add test `TestHookMerge_EscalateEvidence_ContinuesEmitting`.

### Risk #3: Stale Marker After Clean Merge (Medium-High Impact)
**Status:** Bug in current implementation.
**Mitigation:**
- In `cmd/centinela/merge.go::runMerge()`, after successful merge and worktree removal, explicitly call `worktree.ClearPending(".", feature)`.
- Add test `TestRunMerge_CleanAfterPriorConflict_ClearsPendingMarker`.

### Risk #4: cwd-Relative Paths (High Impact, Architectural)
**Status:** Fundamental design choice; affects all state files.
**Mitigation:**
- Option A: Require centinela to always run from repo root (detect via `.git` or `centinela.toml`). Auto-chdir or error if not in root.
- Option B: All state paths must be absolute or repo-relative with explicit root detection.
- Current code assumes cwd=repo root; document this assumption or enforce it.
- Add test: `TestMerge_FromSubdir_RequireRootOrError`.

### Risk #5: Evidence Enum Validation (Medium Impact)
**Status:** Current code relies on default ESCALATE for unknown values (conservative but lacks validation).
**Mitigation:**
- Add enum check in `readStewardHandoff()` after unmarshal:
  ```go
  if e.HandoffTo != "complete" && e.HandoffTo != "user" {
    return "", fmt.Errorf("invalid handoffTo: %q; must be 'complete' or 'user'", e.HandoffTo)
  }
  ```
- Add test `TestReadStewardHandoff_InvalidValue_Error`.

### Risk #6: Corrupt Marker Recovery (High Impact, UX)
**Status:** No recovery path; operator must manually delete `.workflow/<feature>-merge-pending.json`.
**Mitigation:**
- Add a `centinela merge --force-cleanup <feature>` flag that removes marker + evidence without finalization.
- Or: enhance error message: "... run `rm .workflow/<feature>-merge-pending.json` to clear, then retry."
- Add test `TestLoadPending_CorruptJSON_SuggestsCleanup`.

### Risk #7: Branch Leakage (Medium Impact, State Consistency)
**Status:** ResolveMerge does not call DeleteBranch.
**Mitigation:**
- Decide: should ResolveMerge clean the branch, or is that caller's responsibility?
- If ResolveMerge: add `DeleteBranch(repo, feature)` after Remove (soft failure—log but don't block).
- If not: document that branch cleanup is out-of-scope for finalization.
- Add test to clarify expectation.

### Risk #8: Escalation Evidence Persistence (Medium Impact, UX)
**Status:** Evidence file not cleared on ESCALATE; operator must manually delete or restart.
**Mitigation:**
- Option A: Document in escalation message: "Run `centinela merge <feature>` fresh (not `--continue`) to retry after fixing."
- Option B: Add `--retry` flag that clears old evidence and restarts merge.
- Option C: Automatically clear evidence on ESCALATE when operator re-runs `centinela merge` (not `--continue`).
- Add test to document behavior.

### Risk #9: Handoff Field Missing (Low Impact, Defensive)
**Status:** orchestration.ValidateEvidence catches it; future relaxation could miss it.
**Mitigation:**
- Add defensive check in `readStewardHandoff()` before reading handoffTo:
  ```go
  if e.HandoffTo == "" {
    return "", fmt.Errorf("missing handoffTo in steward evidence")
  }
  ```
- This is belt-and-suspenders (orchestration already validates), but improves resilience.

---

## Summary for QA Senior

**Top 5 Hardest Risks (by Impact × Likelihood):**

1. **Race on WritePending**: Direct os.WriteFile without atomic write. Two concurrent `centinela merge` calls corrupt marker JSON. Impact=High, Likelihood=Low but catastrophic if it happens.

2. **Stale Marker After Clean Merge**: Operator fixes conflict manually, re-runs `centinela merge`, succeeds, but old marker lingers. Hook re-emits stale directive, steward invoked unnecessarily. Impact=Medium-High, Likelihood=Medium.

3. **Hook Silent on ESCALATE**: Valid ESCALATE evidence written by steward; hook stops re-emitting. Operator has no signal merge is still blocked. Running `--continue` shows escalation but operator was uninformed. Impact=Medium, Likelihood=Medium.

4. **Corrupt Marker Blocks Forever**: JSON truncated or invalid. LoadPending fails. Hook skips silently. `--continue` also fails. Marker persists. No recovery flag. Operator must manually delete file. Impact=High, Likelihood=Medium (rare but unrecoverable).

5. **cwd-Relative Paths Silent State Loss**: Marker written in wrong `.workflow/` if operator runs from subdirectory. --continue doesn't find it. State silently lost. No automatic repo-root detection. Impact=High, Likelihood=Low but easy to trigger.

**Test Checklist (43 test functions, target files listed):**
- 5 unit tests in merge_pending_test.go (corrupt, concurrent, invalid reason)
- 4 unit tests in merge_evidence_test.go (unknown handoffTo, missing field)
- 6 unit tests in finalize_test.go (corrupt marker, dirty main, no marker, worktree gone, incomplete schema)
- 4 unit tests in hook_merge_test.go (corrupt skip, escalate emit, worktree gone, multiple markers)
- 5 integration tests in merge_test.go (subdir, dirty cleanup, worktree gone, hook+continue)
- 2+ acceptance tests (concurrent conflicts, marker+worktree out-of-sync, escalate retry)

