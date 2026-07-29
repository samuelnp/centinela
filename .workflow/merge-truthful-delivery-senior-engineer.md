# merge-truthful-delivery — senior-engineer

## Files Touched

- `internal/worktree/primary.go` (new, 56) — `PrimaryTree(cwd)` parses
  `git worktree list --porcelain` via the `gitRunner` seam; first entry is
  the primary. Refuses bare primaries and unparseable output, never guesses.
- `internal/worktree/merge_verify.go` (new, 63) — `headSHA`, `currentBranch`
  (detached-HEAD refusal), `isAncestor`, `verifyRemoved`, `verifyAdvance`.
  `verifyAdvance` sets `RefAdvanced` on a moved HEAD, else `AlreadyMerged`
  via ancestor check; neither → hard error.
- `internal/worktree/merger.go` (edit, 78) — `MergeOutcome` gains
  `RefAdvanced`/`AlreadyMerged`; guard order: dirty → self-merge (feature
  branch checked out in primary would make `isAncestor` trivially true) →
  before-SHA capture → merge → validate → `verifyAdvance` → `Remove` →
  `verifyRemoved`.
- `cmd/centinela/merge.go` (edit, 83) — resolves the repo once via
  `worktree.PrimaryTree(".")`, refuses when unresolvable, one-line
  primary-tree notice when invoked inside a worktree, resolved repo passed
  to `DetectSpecConflicts`/`Merge`/`ClearPending`.
- `cmd/centinela/merge_report.go` (new, 25) — success line ONLY on
  `RefAdvanced || AlreadyMerged`, honest already-merged wording,
  defense-in-depth refusal otherwise.

## Architecture Compliance

All git plumbing stays in `internal/worktree` behind the existing
`gitRunner` seam; cmd layer only orchestrates and renders. No existing
exported signature changed — the untouchable test tier still compiles
(`go test ./... -run xxxNONE` verified). All files ≤100 lines.

## Type-Safety Notes

New outcome facts are explicit booleans on `MergeOutcome`, set in exactly
one place (`verifyAdvance`); no stringly-typed status. Errors carry repo and
branch context.

## Trade-Offs

- ~~Deliberate non-changes deferred to Backlog~~ — SUPERSEDED by the
  post-verification round below: all three (`runValidateForMerge`,
  `merge --continue`, `docsPortalRegen` running against the invoking CWD)
  are fixed here. None of those slugs ever reached `.workflow/roadmap.json`
  in this worktree, so there is nothing to un-defer.
- Self-merge refusal (vs treating it as AlreadyMerged): correctness over
  convenience — a self-merge no-op is indistinguishable from the fabricated
  success this feature exists to kill.
- Process note: the implementing subagent stalled after the code and defers
  landed; the orchestrator independently re-verified build/vet/fmt/test-tier
  compile and completed this report from the verified tree state.

## Handoff

qa-senior: priority is the real-git regression test — from inside a feature
worktree, assert `git rev-parse main` advanced AND `.worktrees/<f>` is gone.
Unit-test `PrimaryTree` porcelain parsing (multi-worktree, bare, spaces,
empty) and every `verifyAdvance` branch via the `gitRunner` var; cmd tests
stub `docsPortalRegen`. Spec scenarios map 1:1 (14, incl. self-merge guard).

---

## Post-Verification Round (adversarial verifier: CRITICAL, findings 1-4)

The verifier could not refute the headline fix, but proved the feature had
displaced the same falsehood into `merge --continue` and left two more
untruths. What changed:

**Finding 1 (CRITICAL) — `merge --continue` fabricated the success line.**
`runMerge` now resolves `PrimaryTree(".")` BEFORE the `--continue` branch and
threads it through `dispatchSteward` -> `WritePending`, `runMergeContinue` ->
`ResolveMerge`, the steward-evidence lookup and `centinela hook merge` ->
`LoadPending`. The marker therefore always lives in the primary tree, which
makes a worktree-initiated stall discoverable — and resumable — from the
worktree CWD and the primary CWD alike. `ResolveMerge` now proves the branch
really landed (`verifyLanded`: `merge-base --is-ancestor` in the primary tree)
and removal (`verifyRemoved`) before returning, and `merge_continue.go`'s
unconditional `fmt.Println` was replaced by the same `reportMergeSuccess` the
direct path uses. An APPLY verdict is the steward's claim; ancestry is the
proof. `MergeOutcome.BaseSHA` is persisted in the marker so the resume can
word "merged now" vs "already merged" from evidence rather than assumption.

**Finding 2 — `verifyRemoved` could pass on a live worktree.**
New `internal/worktree/registry.go` parses `git worktree list --porcelain`
for the branch's live (non-prunable) worktree. `verifyRemoved` consults the
registry first and refuses when the branch is still registered anywhere;
`Remove` targets the registered path, so a relocated worktree is really
removed instead of silently skipped. A registry read failure is a refusal,
never an assumed removal.

**Finding 3 — busy worktree left the operator stranded.**
`finishMerge` flags `RemoveFailed` on the outcome, and `reportMergeFailure`
composes the truthful two-part message ("<f> merged into main — verified;
worktree removal failed: <cause>; re-run `<cmd> --force-remove` to retry
removal"). A plain re-run repeats the same truth via the already-merged path;
`--force-remove` (`worktree.WithForceRemove`) completes the outstanding step.

**Finding 4 — the merge-time validate gated nothing.**
`runValidateForMerge` and `docsPortalRegen` now run inside the primary tree
via the new `inDir` helper, and the gate scope is forced full
(`config.FlagForceFull`) because the diff base already contains the merge.

Findings 5-8 were deferred by actually running the `roadmap defer` commands
(`merge-assert-branch-ancestor`, `render-warn-gate-details`,
`merge-dirty-error-names-cause`, `merge-distinguish-missing-branch`).
`merge-validate-primary-tree` was NOT deferred: it is fixed above.
