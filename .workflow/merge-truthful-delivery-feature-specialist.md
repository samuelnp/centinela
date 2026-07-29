# merge-truthful-delivery — feature-specialist

### Feature-Specialist Report: merge-truthful-delivery
**Date:** 2026-07-09

## Behavior Summary

`centinela merge <f>` and `centinela deliver <f> --via merge` become truthful:
before merging, the primary working tree is resolved via `git worktree list
--porcelain` (new `worktree.PrimaryTree(cwd)`, first `worktree <path>` entry,
through the existing `gitRunner` seam) and every repo-scoped call
(`DetectSpecConflicts`, `Merge`, `ClearPending`) receives that path instead of
`"."`. Run from inside `.worktrees/<f>/`, the command prints a one-line notice
naming the primary tree and merges there. After a git-clean merge + validate
pass, verification (new `merge_verify.go`) compares `rev-parse HEAD`
before/after: `MergeOutcome.RefAdvanced` when it moved, else
`MergeOutcome.AlreadyMerged` when `merge-base --is-ancestor <branch> HEAD`
holds; neither → hard error ("merge reported success but <branch> did not
advance") and the success line is never printed. Worktree removal is verified
on disk before being claimed. Refusals: unresolvable primary tree, bare
primary, detached HEAD, dirty primary tree (existing guard, right tree now).
Conflict/validate-fail steward outcomes are untouched.

## Acceptance Criteria (Gherkin)

Spec: `specs/merge-truthful-delivery.feature` — 13 scenarios, titles map 1:1
to `// Scenario:` comments in acceptance tests:

1. merge from inside the feature worktree advances main and removes the
   worktree (THE regression; asserts on `git rev-parse main`, not the message)
2. merge run from inside a worktree prints a one-line primary-tree notice
3. merge run from the primary tree behaves as before, with verification applied
4. merge refuses when the primary tree cannot be resolved
5. merge refuses when the primary working tree is bare
6. merge refuses when the primary tree is in detached HEAD state
7. dirty primary tree is still refused
8. the success message is never printed when the ref did not advance
9. re-running deliver for an already-merged branch reports honestly
10. already-merged re-run with the worktree already gone is an idempotent
    honest success
11. removal is only claimed when the worktree directory is actually gone
12. a text conflict still dispatches the Merge Steward with no success claim
13. a validate failure after a clean text merge still dispatches the steward

## UX States

| State   | Trigger | Surface |
|---------|---------|---------|
| loading | n/a (synchronous CLI) | n/a |
| empty   | already-merged branch (nothing to merge) | exit 0, truthful "already merged" line; worktree cleanup still verified |
| error   | unresolvable/bare/detached/dirty primary tree; ref did not advance; worktree survives removal | non-zero exit, error message; success line never printed |
| success | RefAdvanced (or AlreadyMerged) AND worktree verified gone | existing `ui.RenderSuccess` line (already-merged wording variant); notice line when invoked from inside a worktree |

## Edge Cases

- porcelain variations: single tree, path with spaces, trailing blank lines,
  empty/garbage output, `gitRunner` failure → parsed or clearly refused
- bare primary tree entry → refused before any merge
- detached HEAD in the primary tree → refused (advancing detached HEAD ≠ main)
- already-merged re-run → RefAdvanced=false, AlreadyMerged=true, exit 0,
  honest wording; idempotent when the worktree is also already gone
- git exits 0 but HEAD unmoved and branch not an ancestor → hard error
- worktree directory survives `Remove` → error, removal never claimed
- dirty primary tree → refused (existing `isDirty`, now the right tree)
- conflict / validate-fail outcomes → steward dispatch unchanged, no success
- run from the primary tree → no notice, verification still applies

## Out-of-Scope

- PR delivery path (`deliver --via pr`) — untouched
- Merge Steward conflict/validate-fail dispatch and `merge --continue` resume
- Running merge-time `centinela validate` against the merged primary tree
  instead of the invoking CWD (NEW discovery — deferred, see below)
- Roadmap-mutation atomicity (retrospective WS1.4) and other WS1 items

## Deferred Findings

- `merge-validate-primary-tree` — `cmd/centinela/merge.go:runValidateForMerge`
  ignores its `repo` argument and runs `executeValidation()` in the invoking
  CWD, so merge-time validate can check the worktree instead of the merged
  primary tree. Recorded via `centinela roadmap defer`.

## Handoff

- Next role: senior-engineer
- Open clarifications: none — the plan pins file layout
  (`internal/worktree/primary.go`, `merge_verify.go`, edits to `merger.go` +
  `cmd/centinela/merge.go`), the ≤100-line splits, and the real-git
  regression test shape (external `package worktree_test`, t.TempDir, no
  network).
