# Plan: merge-steward-auto-dispatch

## Mechanism (decided)

State-file-driven dispatch via CENTINELA DIRECTIVE. `centinela merge`
writes a pending marker and prints a directive; a UserPromptSubmit hook
re-surfaces it until valid steward evidence exists; `centinela merge
--continue` validates the evidence and gates finalization. This matches
the existing directive/evidence pattern (`hook_setup`, `hook_context`,
`hook_orchestration`) and reuses the orchestration evidence validator —
the Go binary never calls an LLM.

## Steps

1. `internal/worktree/pending.go` — `PendingMarker` type +
   `WritePending(repo string, o MergeOutcome)`, `LoadPending`,
   `ClearPending`, `PendingPath`. Marker holds feature, reason,
   conflicted paths, worktree path, RFC3339 timestamp. ≤100 lines.
2. `internal/worktree/steward.go` — add `StewardDirective(o)` returning
   the structured `CENTINELA DIRECTIVE:` string (prompt path, inputs,
   evidence paths, `centinela merge --continue` resume cmd).
3. `internal/worktree/finalize.go` — `ResolveMerge(repo, feature,
   validateEvidence func) (Resolution, error)`: re-check clean tree,
   parse `.workflow/<feature>-merge-steward.json`, classify
   APPLY/complete vs ESCALATE/user/invalid. On APPLY: `Remove` worktree
   + `ClearPending`. On ESCALATE: keep both, return escalation detail.
4. `cmd/centinela/merge.go` — add `--continue` bool flag. Conflict path
   now calls `worktree.WritePending`, prints `StewardDirective`, exits
   non-zero. `--continue` path calls `worktree.ResolveMerge` (passing an
   orchestration-validator adapter) and renders APPLY success or
   ESCALATE block (escalation note + diff to stderr).
5. Evidence adapter in `cmd/centinela/` (thin): wrap
   `orchestration.ValidateEvidence(path, feature, "merge",
   RoleMergeSteward, nil)` so the worktree layer stays free of the
   orchestration import.
6. `cmd/centinela/hook_merge.go` — new UserPromptSubmit hook: for each
   pending marker without valid steward evidence, print the dispatch
   directive + a `ui.RenderStep`-style block. Register `cmdMergeHook`
   in `internal/setup/hooks.go` (`ensurePrompt`).
7. UI: `internal/ui/render_merge.go` — `RenderMergeStewardNeeded` and
   `RenderMergeEscalated` (read-only rendering, no state mutation).
8. Tests — unit: pending read/write/clear, finalize APPLY vs ESCALATE
   vs invalid-evidence, directive string. Integration: hook re-emits
   while pending. Acceptance: text-conflict → directive → stub APPLY
   evidence → `--continue` finalizes; stub ESCALATE evidence →
   `--continue` stays blocked, exit non-zero, worktree kept.
9. `.workflow/merge-steward-auto-dispatch-edge-cases.md` capturing brief
   edge cases + any found during code.
10. Docs step: update `docs/architecture/merge-steward-prompt.md`
    (auto-dispatch + `--continue` flow), `evidence-contract.md` note,
    `workflow-enforcement.md`, README merge section.
