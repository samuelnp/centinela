# merge-truthful-delivery — big-thinker

### Big-Thinker Report: merge-truthful-delivery
**Date:** 2026-07-09

#### Problem

Operators and orchestrating agents running `centinela deliver --via merge`
from inside a feature worktree — the documented operational model — get a
fabricated success: `cmd/centinela/merge.go:47` passes `repo="."` into
`worktree.Merge`, so `gitRunner` (`internal/worktree/provision.go:12`,
`cmd.Dir = repo`) executes `git merge --no-ff <branch>` with HEAD already on
the feature branch. Git self-no-ops ("Already up to date", exit 0), `Remove`
silently no-ops because `Exists` stats `./.worktrees/<feature>` against the
wrong base, and `merge.go:65` prints "Merged … into main and removed its
worktree" unconditionally. Observed 4/4 deliveries in the 2048-rust field
test (`retrospective.md` §2 row 2, appendix incidents 1–2), including one
near branch-loss during manual recovery. This violates the framework's own
north star: state transitions must be truthful. Why now: it is WS1.1 (P0,
pure bug fix) and the retrospective's named success metric ("assert on the
ref, not the message").

#### Why this is the right minimal fix

- **Resolve-then-merge beats refuse-only.** Refusing when
  `IsInsideWorktree(".")` (retrospective option b) would be safe but hostile:
  the operational model runs centinela from inside the worktree (hooks
  resolve the active workflow from shell CWD), so refusal forces a
  cd-to-primary that fights the rest of the framework. `PrimaryTree(cwd)`
  keeps the ergonomics and refuses only when resolution genuinely fails.
- **Resolution belongs in `internal/worktree`, not a `git -C` at the cmd
  layer.** Every worktree API is already repo-parameterized; the bug is one
  caller feeding `"."`. Fixing the caller once and adding `PrimaryTree` gives
  every future caller the correct base, whereas cmd-level `git -C` would
  leave the layer's contract silently CWD-relative.
- **Verification is the actual fix; resolution is just plumbing.** Even with
  perfect resolution, success must be earned: `rev-parse HEAD` before/after +
  `merge-base --is-ancestor` makes the success line depend on the ref, which
  also converts any future regression of this family into a hard error
  instead of a silent lie. `--no-ff` on an already-merged branch prints
  "Already up to date" without moving HEAD, so `RefAdvanced=false,
  AlreadyMerged=true` cleanly encodes idempotent re-runs.
- **Proof/process boundary intact.** This changes delivery truthfulness only;
  `cmd/centinela/complete.go` verification is untouched and no profile
  branch is introduced anywhere.

#### Scope

- In: `internal/worktree/primary.go` (`PrimaryTree` via
  `git worktree list --porcelain`, first-entry parse, refuse on bare/garbage),
  `internal/worktree/merge_verify.go` (HEAD before/after, ancestor check,
  removal verification), `MergeOutcome.RefAdvanced/AlreadyMerged`,
  `cmd/centinela/merge.go` wiring (resolved repo into `DetectSpecConflicts`,
  `Merge`, `ClearPending`; worktree notice; success only on
  `RefAdvanced || AlreadyMerged`), real-git regression test from inside a
  worktree asserting main's SHA advanced and `.worktrees/<f>` is gone.
- Out (deferred, recorded below): running merge-time `centinela validate`
  and the portal regen against the resolved primary tree
  (`runValidateForMerge` at `merge.go:69` ignores its repo argument;
  `docsPortalRegen` writes relative to the invoking CWD); the
  `merge --continue` steward path (`merge_continue.go:25` passes `"."` to
  `ResolveMerge` — same wrong-CWD family); PR delivery; steward dispatch
  logic; roadmap-mutation atomicity (WS1.4).

#### Dependencies & Assumptions

- Reuses the overridable `gitRunner` seam (`provision.go`), `isDirty`
  (`merger_git.go`), `Remove`/`Exists`, `ClearPending`, `IsInsideWorktree`
  (`path.go` — first non-test caller), and the `docsPortalRegen` test seam.
- Assumes git's documented porcelain contract: the main worktree is the
  first `worktree <path>` entry of `git worktree list --porcelain`.
- Assumes `deliver.go` needs no change (it delegates to `runMerge`).
- Coverage is per-package (no -coverpkg): new logic needs colocated tests in
  `internal/worktree` and `cmd/centinela`; gate 95%, target ≥97%. G1 (≤100
  lines) applies to `_test.go` files too.
- Builds on parallel-feature-worktrees, merge-steward-auto-dispatch,
  delivery-artifact-generation (their flows must keep passing).

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Porcelain parse fragility (spaces in paths, blank lines, bare primary, garbage) | Medium | Low | First-block contract is documented; value = everything after first space; refuse (never guess) on no match/bare; unit-test each variation |
| Primary tree has the FEATURE branch checked out → self-no-op mislabeled `AlreadyMerged` (a branch is trivially its own ancestor) | High | Low | Refuse before merging when primary HEAD's symbolic ref equals the feature branch (extends the planned detached-HEAD guard) |
| Primary tree on a non-main branch → truthful-but-surprising merge target; message hardcodes "main" | Medium | Low | Verification is branch-agnostic (HEAD advance); success/notice wording should name the actual target branch |
| `git worktree remove` refuses on untracked files in the worktree → happy path now errors where it used to lie | Medium | Medium | Honest failure is the point; `complete` auto-commits each step so a delivered worktree is normally clean; test the dirty-worktree message quality |
| Post-removal CWD is a deleted directory (process ran inside the worktree) → portal regen and any later relative op fail | Low | High | Regen is best-effort (notice, never fails the merge); deferred sibling moves regen to the primary tree |
| False `AlreadyMerged` from checking ancestor before merging instead of after an unmoved HEAD | Medium | Low | Only evaluate `isAncestor` after a git-clean merge with `before == after`; test both branches |
| Real-git test flakiness / hang (network, missing identity) | Medium | Low | t.TempDir repos, explicit user.name/email + `-b main`, no remotes ever (a prior acceptance test hung the suite on a network push) |
| Regression of conflict/steward flows (they must keep dispatching with zero success claims) | High | Low | Conflict and validate-fail early returns unchanged; existing merger/steward tests plus a re-run of the golden flows |

#### Rollout

- Step 1: `PrimaryTree` + porcelain unit tests (pure addition, no callers).
- Step 2: `merge_verify.go` + `MergeOutcome` flags, wired into `Merge` with
  the detached-HEAD/self-branch guard; stubbed-runner tests for
  advanced / already-merged / neither(hard error) / removal-verify paths.
- Step 3: `cmd/centinela/merge.go` wiring: resolve once, pass everywhere,
  notice when inside a worktree, success gated on the flags, honest
  already-merged wording.
- Step 4: real-git regression (external `package worktree_test` + cmd-level
  `runMerge` test from inside a worktree) + tests/ tier trio with spec
  traceability; dogfood a /tmp-built binary in a throwaway repo.
- Can wait: everything listed under Out (deferred below).

#### Deferred Findings

- `merge-validate-primary-tree` — merge-time `centinela validate` and
  `docsPortalRegen` run in the invoking CWD, not the merged primary tree
  (`runValidateForMerge` discards its repo arg). Already present in the
  Backlog (recorded by the parallel plan role); independently confirmed here.
- `merge-continue-primary-tree` — `merge --continue` passes `"."` to
  `ResolveMerge` (`merge_continue.go:25`); pending-marker load/clear and the
  finalize `Remove` hit the invoking CWD — same wrong-CWD family as this
  fix. Recorded via `centinela roadmap defer`.

#### Handoff

- Next role: feature-specialist (brief `docs/features/merge-truthful-delivery.md`,
  plan `docs/plans/merge-truthful-delivery.md`, spec
  `specs/merge-truthful-delivery.feature` exist; this analysis endorses them).
- Outstanding questions: (1) should the success/notice line name the actual
  primary-tree branch instead of hardcoding "main"? (2) confirm the
  self-branch guard (primary HEAD == feature branch → refuse) is added next
  to the detached-HEAD guard so `AlreadyMerged` can never mask a self-merge;
  (3) dirty-worktree removal failure needs an actionable error message.
