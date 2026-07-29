# Plan — merge-truthful-delivery

> Brief: [docs/features/merge-truthful-delivery.md](../features/merge-truthful-delivery.md).
> Field evidence: `retrospective.md` §2 row 2, §4 WS1.1, appendix incidents 1–2.

## Goal
`centinela merge` / `deliver --via merge` run from ANY CWD (worktree or
primary) merges in the **primary working tree**, verifies the ref actually
advanced (or was already merged), verifies the worktree is actually gone, and
only then prints success. No redesign of conflict/steward paths.

## Deliverables

### `internal/worktree/primary.go` (new, ~50)
```go
func PrimaryTree(cwd string) (string, error)
```
Runs `gitRunner(cwd, "worktree", "list", "--porcelain")`. The primary tree is
the **first** `worktree <path>` entry. Errors (never guesses): runner failure,
empty/garbage output, no `worktree ` line, first block carrying the `bare`
attribute ("primary working tree is bare"). Tolerates paths with spaces
(value = everything after the first space) and trailing blank lines.

### `internal/worktree/merge_verify.go` (new, ~55)
```go
func headSHA(repo string) (string, error)            // rev-parse HEAD
func isAncestor(repo, branch string) bool            // merge-base --is-ancestor <branch> HEAD
func verifyRemoved(repo, feature string) error       // stat: dir must be gone
func verifyAdvance(o *MergeOutcome, repo, before string) error
```
`verifyAdvance` re-reads HEAD, sets `RefAdvanced = before != after`; when not
advanced, sets `AlreadyMerged = isAncestor(repo, o.Branch)`; when neither →
`fmt.Errorf("merge reported success but %q did not advance %s", branch, ...)`.

### `internal/worktree/merger.go` (edit, stays ≤100)
- `MergeOutcome` gains `RefAdvanced bool`, `AlreadyMerged bool`.
- Guard before merging: `headSHA(repo)` captured (also fails fast on a repo
  with no commits); refuse detached HEAD in the primary tree via
  `gitRunner(repo, "symbolic-ref", "-q", "HEAD")` error.
- **Self-merge guard**: refuse when the primary tree's checked-out branch IS
  the feature branch (compare `symbolic-ref --short HEAD` to `o.Branch`) —
  otherwise the merge self-no-ops and `isAncestor` is trivially true (a
  commit is its own ancestor), so `AlreadyMerged` would fabricate the exact
  false success this feature exists to kill.
- After the validate pass: `verifyAdvance(&out, repo, before)`; hard error on
  neither-flag. Then `Remove(repo, feature, false)` + `verifyRemoved` —
  removal is only claimed when the directory is actually gone.
- Keep the existing `isDirty` refusal and conflict/validate-fail early
  returns exactly as-is (they never claim success).

### `cmd/centinela/merge.go` (edit, stays ≤100)
- `repo, err := worktree.PrimaryTree(".")` — on error, refuse: no merge, no
  success line ("cannot resolve primary working tree: …").
- `if worktree.IsInsideWorktree(".")` → one-line notice naming `repo`.
- Pass `repo` to `DetectSpecConflicts`, `Merge`, `ClearPending`.
- Success line ONLY when `outcome.RefAdvanced || outcome.AlreadyMerged`;
  already-merged wording: branch was already merged, worktree cleaned up.
- If the message block outgrows 100 lines, split a `mergeSuccessLine(outcome)`
  helper into `cmd/centinela/merge_report.go`.

### Post-verification round (adversarial verifier, findings 1–4)
- `internal/worktree/registry.go` (new): `registeredWorktree` /
  `findRegisteredWorktree` parse `git worktree list --porcelain` for the
  branch's live (non-prunable) worktree. `verifyRemoved` asks the registry
  first; `Remove` removes the registered path. A registry read failure is a
  refusal, never an assumed removal.
- `internal/worktree/merge_finish.go` (new): `MergeOption` /
  `WithForceRemove` / `finishMerge`, which flags `RemoveFailed` so a merge
  that landed but could not clean up reports both halves.
- `internal/worktree/merge_verify.go`: `verifyLanded` — the `--continue`
  proof is ancestry in the primary tree (an APPLY verdict is a claim).
- `internal/worktree/finalize.go`: `ResolveMerge` verifies target branch,
  ancestry and removal, and returns the verified `MergeOutcome`.
- `MergeOutcome.BaseSHA` is persisted in the pending marker so `--continue`
  can word "merged now" vs "already merged" from evidence.
- `cmd/centinela`: `repo` threaded into `dispatchSteward`, `runMergeContinue`,
  `hook merge` and the steward-evidence lookup; `--force-remove` flag;
  `reportMergeFailure` for the truthful two-part outcome; `inDir` +
  `config.FlagForceFull` so the post-merge validate and the portal regen run
  on the merged primary tree instead of the invoking CWD.

### Unchanged
`deliver.go` (routes into fixed `runMerge`), `path.go` (its
`IsInsideWorktree` finally gets a caller).

## Constraints
- Every source + `_test.go` ≤100 lines (G1 applies to test files too).
- Logic in `internal/worktree`; `cmd/centinela/merge.go` stays thin.
- Strict typing; CLI strings via existing `ui.Render*` patterns (no i18n
  surface in this repo's CLI).
- Coverage is per-package: colocate tests in `internal/worktree` and
  `cmd/centinela`; gate 95%, aim ≥97%.

## Tests (colocated + tiers)
- `internal/worktree/primary_test.go` — porcelain parsing via `gitRunner`
  stub: two-tree output, single tree, bare primary refused, path with spaces,
  trailing blank lines, empty/garbage output, runner error propagated.
- `internal/worktree/merge_verify_test.go` — stubbed `gitRunner`:
  RefAdvanced; AlreadyMerged (unmoved HEAD + ancestor); neither → hard error
  with the exact message; detached-HEAD refusal; `verifyRemoved` failure when
  the dir survives.
- `internal/worktree/merge_realgit_test.go` (external `package worktree_test`,
  real git, t.TempDir) — THE regression: init repo (main), `Create` worktree,
  commit on the branch inside it, resolve `PrimaryTree` from inside the
  worktree, run `Merge` against it; assert `git rev-parse main` ADVANCED and
  `.worktrees/<f>` REMOVED. Split helpers into a sibling `_helper_test.go`
  to hold the 100-line cap.
- `cmd/centinela/merge_truthful_test.go` — stub `docsPortalRegen`, real git
  temp repo, chdir into the worktree, `runMerge`; assert main SHA moved,
  worktree gone, notice printed. Second case: already-merged re-run exits 0
  with the truthful wording. Third: unresolvable primary tree → refusal.
- tests/ tier trio (`tests/{unit,integration,acceptance}/merge_truthful_
  delivery_*_test.go`) with `// Acceptance: specs/merge-truthful-delivery.feature`
  + `// Scenario:` traceability; acceptance drives a temp-built binary against
  a local t.TempDir repo — **no network, no real push** (a local bare origin
  if a remote is ever needed).

## Edge cases the tests must pin
Bare primary refused; detached HEAD refused; already-merged re-run honest
(incl. worktree already gone → idempotent); porcelain variations; dirty
primary tree still refused (existing guard, right tree now); conflict and
validate-fail outcomes still dispatch the steward with no success claim;
run from primary tree (no notice, verification still applies).

## Verification (end-to-end)
1. `go test ./...` green; `./scripts/check-coverage.sh` ≥95% (target ≥97%
   on `internal/worktree` + `cmd/centinela`); `check-fmt.sh`.
2. Dogfood a /tmp-built binary in a throwaway repo: start a feature worktree,
   commit, run `deliver --via merge` FROM INSIDE the worktree → main SHA
   moved, worktree removed, notice shown; re-run → honest already-merged.
3. `centinela validate` passes in this worktree.
