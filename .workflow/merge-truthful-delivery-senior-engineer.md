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

- Deliberate non-changes, deferred to Backlog: `runValidateForMerge` runs in
  the invoking CWD (`merge-validate-primary-tree`); `merge --continue` still
  passes "." (`merge-continue-primary-tree`); `docsPortalRegen` writes
  relative to the invoking CWD (`merge-portal-regen-primary-tree`).
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
