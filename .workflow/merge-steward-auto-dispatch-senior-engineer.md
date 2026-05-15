### Senior-Engineer Report: merge-steward-auto-dispatch

**Date:** 2026-05-15

#### Summary

Activated the dormant Merge Steward contract via state-file dispatch.
`centinela merge` now writes a pending marker and emits a structured
`CENTINELA DIRECTIVE:` on conflict/validate-failure; a new
UserPromptSubmit hook re-surfaces it until valid steward evidence
exists; `centinela merge --continue` re-validates that evidence through
the existing orchestration validator and gates finalization. The Go
binary never calls an LLM. Clean-merge behavior is byte-for-byte
unchanged (regression guard preserved).

#### Files Touched

| Path | Reason |
|------|--------|
| `internal/worktree/merge_pending.go` | New. `PendingMarker` type + `WritePending`/`LoadPending`/`ClearPending`/`PendingPath`; `Directive()` re-renders the dispatch line from a stored marker. Idempotent rewrite (truncating `WriteFile`). |
| `internal/worktree/finalize.go` | New. `ResolveMerge` + injected `StewardEvidenceValidator` adapter type; clean-tree re-check, APPLY/ESCALATE/invalid classification, worktree+marker teardown on APPLY. |
| `internal/worktree/steward.go` | Added `StewardPromptPath` const + `StewardDirective()` producing the two-line directive (imperative + details) mirroring `hook_orchestration.go`. |
| `cmd/centinela/merge.go` | Added `--continue` bool flag; dispatches to continue or normal merge; conflict path calls `dispatchSteward`. Stays a thin 53-line orchestrator. |
| `cmd/centinela/merge_dispatch.go` | New. `dispatchSteward`: write marker, print directive block + line, exit non-zero, keep worktree. |
| `cmd/centinela/merge_continue.go` | New. `stewardEvidenceValidator` adapter (wraps `orchestration.ValidateEvidence`); `runMergeContinue` renders APPLY success vs ESCALATE block (note+diff to stderr). |
| `cmd/centinela/merge_evidence.go` | New. `readStewardHandoff` decodes the verdict (`handoffTo`) from already-validated evidence. |
| `cmd/centinela/hook_merge.go` | New. `centinela hook merge` UserPromptSubmit hook: globs `*-merge-pending.json`, re-emits directive unless valid steward evidence present. Silent when no marker. |
| `internal/ui/render_merge.go` | New. `RenderMergeStewardNeeded` + `RenderMergeEscalated` — pure rendering, no state mutation. |
| `internal/setup/hooks.go` | Registered `cmdMerge` via `ensurePrompt` following the exact existing pattern. |
| `internal/setup/hooks_test.go` | Updated prompt-hook count assertion 6→7 (consequence of the new required hook). |
| `cmd/centinela/merge_steward_test.go` | Updated assertion to the new contract wording; added pending-marker presence check. |

#### Architecture Compliance

- **G2 boundary (verified via `go list -deps`):** `internal/worktree`
  has **zero** dependency on `internal/orchestration`. The validator is
  injected as `type StewardEvidenceValidator func(feature string)
  (verdict string, err error)` defined in `finalize.go`. The adapter
  `stewardEvidenceValidator` lives in `cmd/centinela/` — the only layer
  permitted to import both packages — and wraps
  `orchestration.ValidateEvidence(path, feature, "merge",
  RoleMergeSteward, nil)`. Verdict = the evidence `handoffTo`
  (`complete`=APPLY, `user`=ESCALATE), read post-validation.
- **G7 outer-layer:** `cmd/centinela/merge.go` is 53 lines of pure
  wiring; all state/decision logic (`ResolveMerge`, marker I/O,
  directive string) lives in `internal/worktree/`.
- **G1 file size:** every touched source file ≤ 100 lines. Largest:
  `merge_pending.go` 87, `finalize.go` 80, `hooks.go` 66.
- `internal/ui/render_merge.go` renders only; no mutation (PROJECT.md
  rule 2).

#### Type-Safety Notes

- No `interface{}`/`any`. The injected validator is a concrete named
  func type. Evidence-verdict decode uses a minimal typed struct, not a
  `map[string]any`.
- `LoadPending` returns `(*PendingMarker, error)` — absence is `(nil,
  nil)`, distinct from corruption (typed error), so callers cannot
  conflate "no pending merge" with "broken marker".
- `Resolution` is an explicit struct (`Finalized`/`Escalated`/`Verdict`/
  `EscalationNote`) rather than stringly-typed status.

#### Trade-Offs

- **Verdict from `handoffTo`, not a separate field.** The evidence
  contract already pins `complete`=APPLY / `user`=ESCALATE; reusing it
  avoids widening the schema or re-parsing the markdown report.
- **`--continue` flag over a `merge finalize` subcommand.** Confirmed by
  big-thinker/feature-specialist; matches the `git rebase --continue`
  mental model and keeps one command.
- **Hook globs all `*-merge-pending.json`** (not just active
  workflows) — merges are out-of-band, so a feature need not have a
  live workflow to have a pending merge.
- **Two existing-test assertions updated** (hook count 6→7; steward
  error wording). These pinned the *old* contract this feature
  deliberately supersedes; no new test files written (qa-senior owns
  step 3).

#### Handoff

- Next role: qa-senior
- Outstanding TODOs: full unit/integration/acceptance suite for the 13
  spec scenarios (step 3). Edge cases recorded in the JSON evidence.
