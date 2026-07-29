# Feature Brief — merge-truthful-delivery

> WS1.1 of the 2048-rust field-test retrospective (`retrospective.md`, section 2
> row 2, appendix incidents 1–2). Pure bug fix: make `centinela deliver --via
> merge` / `centinela merge` truthful — a success message must mean main
> actually advanced and the worktree is actually gone.

## Problem — what pain, who
`centinela deliver --via merge`, run from **inside a feature worktree** (the
documented operational model), prints "Merged <feature> into main and removed
its worktree" while **main never advances and the worktree is never removed**.
Observed 4× in the 2048-rust field test, with one near branch-loss during
manual recovery. Root cause chain:
- `cmd/centinela/merge.go:47` passes `repo="."` into `worktree.Merge`.
- `internal/worktree/provision.go:gitRunner` sets `cmd.Dir = repo`, so from a
  worktree HEAD **is** the feature branch and `git merge --no-ff <branch>`
  self-no-ops ("Already up to date", exit 0).
- `Remove`/`Exists` stat `./.worktrees/<feature>` relative to the wrong CWD →
  silent no-op (Remove is idempotent-by-design).
- `merge.go:65` prints success unconditionally — no ref verification.
- `deliver.go:54-55` routes `--via merge` into the same path.
- `internal/worktree/path.go:IsInsideWorktree` exists but is never consulted.

## Scope (this feature ONLY)
- **In:** resolve the primary working tree before merging
  (`PrimaryTree(cwd)` parsing `git worktree list --porcelain` via the existing
  `gitRunner` seam); refuse clearly when it cannot be resolved; verify the
  target ref advanced (`rev-parse HEAD` before/after in the primary tree) or
  the branch was already merged (`merge-base --is-ancestor`); verify the
  worktree directory is actually gone after `Remove`; success message only
  when `RefAdvanced || AlreadyMerged`, worded truthfully for already-merged.
- **In (added after adversarial verification):** the same primary tree is
  threaded through `WritePending`/`LoadPending`/`ResolveMerge`/`hook merge`, so
  a stall started from a worktree CWD is resumable from either CWD and
  `merge --continue` prints success only through the same verified reporter;
  removal is verified against `git worktree list --porcelain` (the registry),
  not the `.worktrees/<feature>` path convention; a merge that lands but
  cannot remove the worktree reports both halves and offers `--force-remove`;
  the post-merge `centinela validate` and the portal regen run in the primary
  tree, with the gate scope forced full (the diff base already contains the
  merge, so a diff-aware run gates nothing).
- **Out:** PR delivery path; steward conflict/validate-fail dispatch logic;
  roadmap-mutation atomicity (WS1.4).

## User Stories
- As an operator finishing a feature from inside `.worktrees/<f>/`,
  `centinela deliver --via merge` merges into main **in the primary tree**,
  removes the worktree, and only then reports success.
- As an operator, if the merge machinery cannot determine the primary tree,
  the command refuses with a clear error instead of guessing `"."`.
- As an operator re-running deliver after a branch already landed, I get an
  honest "already merged" message — not a fabricated "Merged … into main".
- As an orchestrating agent, I can trust the exit code + message: success is
  asserted on the ref, not on the absence of git errors.

## Acceptance Criteria (THIS feature → Gherkin)
1. `merge`/`deliver --via merge` from inside `.worktrees/<f>/` advances
   `git rev-parse main` in the primary tree and removes `.worktrees/<f>` —
   the retrospective's named success metric.
2. `repo` is resolved via `PrimaryTree(".")`; the resolved path is passed to
   `DetectSpecConflicts`, `Merge`, and `ClearPending`. When resolution fails
   (not a git repo, bare primary, unparseable porcelain) the command exits
   non-zero with a clear message and merges nothing.
3. When run from inside a worktree, a one-line notice names the primary tree
   the merge will operate on (`IsInsideWorktree(".")`).
4. `MergeOutcome` gains `RefAdvanced bool` and `AlreadyMerged bool`. After a
   git-clean merge, if the primary tree's HEAD did not advance AND the branch
   is not already an ancestor → hard error ("merge reported success but
   <branch> did not advance"); the success line is never printed.
5. Already-merged branch re-run: `RefAdvanced=false, AlreadyMerged=true` →
   exit 0 with truthful wording (already merged; worktree cleanup still runs).
6. After `Remove`, the worktree directory is verified gone before removal is
   claimed; a still-present directory is an error, not a success.
7. Dirty primary tree is still refused (existing `isDirty` guard, now
   pointing at the right tree).

## Edge Cases (THIS feature ONLY)
- Primary tree resolution: bare primary entry (`bare` attribute) → refused;
  porcelain variations (path with spaces, trailing blank lines, single-tree
  output, empty/garbage output) → parsed or clearly refused; `gitRunner`
  failure propagated.
- Detached HEAD in the primary tree → refused before merging (advancing a
  detached HEAD is not "main advanced").
- Already-merged re-run where the worktree was ALSO already removed →
  idempotent success with the truthful wording.
- Text conflict / validate failure outcomes keep today's steward dispatch —
  no success verification applies (nothing claims success).
- `--continue` after a worktree-initiated stall: marker in the primary tree,
  dirty-guard on the primary tree, ancestry proof before any success line;
  resumable from the worktree CWD and from the primary CWD alike.
- Worktree busy (untracked/modified files) after a verified advance →
  truthful two-part outcome + `--force-remove` recovery, idempotent on re-run.
- Worktree registered outside `.worktrees/<feature>` → really removed and
  verified against the registry, never claimed by a passing `os.Stat`.
- Run from the primary tree itself (not a worktree): behavior unchanged
  except verification now applies; no worktree notice printed.

## Data Model
No persisted-schema change. `worktree.MergeOutcome` gains two booleans
(`RefAdvanced`, `AlreadyMerged`); consumed only by `cmd/centinela/merge.go`.

## Integration Points
- Reuses the overridable `gitRunner` seam (`internal/worktree/provision.go`),
  `IsInsideWorktree`/`DetectFeatureFromCwd` (`path.go`), `isDirty`
  (`merger_git.go`), `Remove` (`remove.go`), `ClearPending`, and the
  `docsPortalRegen` test seam in `cmd/centinela/merge.go`.
- `deliver.go` needs no change: it delegates to `runMerge`, which is fixed.

## Risks
- **Porcelain parse fragility** (Med): first `worktree <path>` block is the
  primary tree by git contract; unit-test variations, error on no match.
- **False AlreadyMerged** (Med): `merge-base --is-ancestor` must be checked
  only after a git-clean merge with an unmoved HEAD; test both branches.
- **Real-git test flakiness** (Low): t.TempDir repos with explicit
  `user.name`/`user.email`/`init.defaultBranch=main`; no network.
- **Per-package coverage** (Low): colocated `internal/worktree/*_test.go` and
  `cmd/centinela/*_test.go`; ≥97% target on both packages.

## Decomposition
Coherent single slice (~3 small new files + 2 surgical edits). No split.
Deferred sibling: run merge-time validate against the merged primary tree.
