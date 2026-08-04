# roadmap-state-hygiene — senior-engineer

Implements slices S1–S6 of `docs/plans/roadmap-state-hygiene.md` in order.
Acceptance-test authoring is deliberately left to the tests step (no files under
`tests/acceptance/` were written here).

## Files Touched

### New — S1, the pathspec leaf and the git primitives
- `internal/roadmapstate/paths.go` (68) — `Paths()`, `IsStatePath`, `Covers`.
  Stdlib-only leaf. THE definition of roadmap state, shared by the commit
  pathspec, the tree digest and the freshness range so they cannot drift apart.
- `internal/roadmapstate/message.go` (35) — `Message(verb, subject)`,
  `SubjectLimit`; conventional `chore(roadmap): …`, single-line, rune-safe
  truncation at 60.
- `internal/gitutil/commit.go` (97) — `CommitPaths`, `ErrNothingToCommit`,
  `existingPaths` (drops a gitignored-and-untracked entry so one ignored path
  cannot take the whole mutation commit down).
- `internal/gitutil/commit_guard.go` (77) — `IsRepo`, `HasHead`,
  `InProgressOperation`, `CommitBlockedReason`.

### New — S2, the mutation choke point
- `internal/roadmap/statesync.go` (82) — `Committer`, `SyncOptions`,
  `SyncReport`, `Sync`, `ErrNoChange`.
- `internal/roadmap/regenerate.go` (43) — `RegenerateMarkdown`,
  `PromoteArtifactPaths`, `WriteRoadmapJSON`.
- `internal/ui/render_roadmap_sync.go` (41) — committed / left-uncommitted /
  warning lines.
- `cmd/centinela/roadmap_sync.go` (56) — `syncRoadmapState`,
  `roadmapCommitter`, `roadmapAutoCommitEnabled`.

### New — S3, freshness
- `internal/workflow/freshness_range.go` (48) —
  `revisionRangeIsRoadmapStateOnly`, fail-closed.

### New — S4, backlog aging + nudge
- `internal/roadmap/backlog_age.go` (69) — `Aged`, `AgeBacklog`, `StaleOnly`,
  `DefaultStaleDays`.
- `internal/roadmap/backlog_stats.go` (32) — `BacklogStats`, `SummarizeBacklog`.
- `internal/roadmap/backlog_nudge.go` (31) — `Nudge`, `NudgeFor`.
- `internal/ui/render_backlog_list.go` (76), `internal/ui/render_backlog_nudge.go` (21).
- `cmd/centinela/roadmap_backlog.go` (76), `cmd/centinela/roadmap_backlog_json.go` (62).

### New — S5, `roadmap resolve`
- `internal/roadmap/resolve.go` (78) — `PhaseConflictError`, `Merged`, `Resolve`.
- `internal/roadmap/resolve_threeway.go` (67), `resolve_backlog.go` (69),
  `resolve_index.go` (49), `resolve_order.go` (67), `resolve_side.go` (87),
  `resolve_stages.go` (53: `StageRunner`, `Stages`, `ReadStages`, `Conflicted`).
- `internal/ui/render_resolve.go` (28), `cmd/centinela/roadmap_resolve.go` (88).

### Modified
- `internal/roadmap/rawrender.go`, `rawphase_render.go` — S6: every phase
  renders canonically; `phaseKeyOrder` preserves unknown phase keys.
- `internal/roadmap/rawio.go` — `parseRawRoadmap` extracted from
  `readRawRoadmap` so `resolve` parses index stages the same way.
- `internal/roadmap/firstincomplete.go` — `FirstIncompleteSchedulable`.
- `internal/treestate/digest.go`, `untracked.go` — the `.workflow/` constant is
  replaced by `roadmapstate.Covers` / `IsStatePath`.
- `internal/workflow/validate_freshness.go` — `compareStamp` consults the
  revision range before declaring a moved HEAD stale.
- `cmd/centinela/complete.go` — prints the Backlog nudge on `done`.
- The ten mutation commands — one `syncRoadmapState(...)` line each; the
  "Remember to sync ROADMAP.md" string no longer exists in the codebase.
- `centinela.toml` (`import_graph` leaf layer + rationale), `PROJECT.md` (G2).

## Architecture Compliance

- **G1** — every source and test file is ≤100 lines (verified by the gate over
  the 91-file changed set).
- **G2** — `internal/roadmapstate` is a new stdlib-only leaf, declared in the
  gate's leaf layer and justified in PROJECT.md. It exists precisely to avoid
  the cycle `treestate → roadmap → workflow → treestate`.
- **G7** — the CLI stays thin. `syncRoadmapState` wires config + committer and
  prints; every decision (what to regenerate, whether to commit, what the
  reason is) is in `internal/roadmap`. `runRoadmapBacklog` does no aging or
  counting. `resolveRoadmapState` sequences domain calls and never merges.
- The domain never runs git: `roadmap.Committer` and `roadmap.StageRunner` are
  the two injected seams.
- Presentation lives in `internal/ui` only.

## Type-Safety Notes

- No `any`, no untyped maps escaping a package. Merge stages are
  `json.RawMessage` throughout so unknown fields survive verbatim.
- `AgeDays` is a `*int` in the JSON payload so an unknown clock is `null`, never
  a zero that reads as "brand new".
- Sentinels are typed: `gitutil.ErrNothingToCommit`, `roadmap.ErrNoChange`,
  `*roadmap.PhaseConflictError` (matched with `errors.As`).

## Trade-Offs

1. **`gitutil.PathsChangedSince` was not added.** The plan listed it, but S3
   reads the range through the already-injected `treestate.Runner`, which keeps
   the freshness path testable without a real repo. Adding an unused second
   implementation of the same query would be the drift this feature attacks.
2. **`MergeInProgress(repo) bool` became `InProgressOperation(repo) string`.**
   The spec distinguishes "merge in progress" from "rebase in progress" in the
   operator-facing reason; one function returning the reason serves both, and a
   bool wrapper would be dead API.
3. **The one-time normalization commit of `.workflow/roadmap.json` was NOT
   made.** The renderer change is in place, so the next real mutation
   normalizes the file with a truthful message. Doing it here would mean either
   hand-writing bytes into roadmap.json (the manual editing this feature exists
   to remove) or fabricating a no-op mutation whose commit message lies about
   what it did — and R4 says a whole-file normalization collides with every
   in-flight branch, of which this repo currently has several. It was verified
   lossless in a throwaway clone of the real file: 15 phases / 176 features
   round-tripped semantically identical, `roadmap.json` diff 400 deletions →
   69 insertions, `ROADMAP.md` +1 line. Recommend landing it as its own
   mechanical commit on main immediately after this merges.
4. **`--no-verify` on the mutation commit** is scoped to generated governance
   state and justified in `CommitPaths`'s doc comment; no user-facing flag.
5. **Staleness is strictly `> N` days.** A finding deferred exactly N days ago
   is not yet "older than N days"; tested at the boundary.
6. **`resolve` re-sorts the whole Backlog** by `deferredAt` then name. That is
   a larger one-time diff than a pure append would be, but it makes the merge
   result a function of its inputs alone.
7. **`indexFindings` refuses a nameless Backlog entry.** `featureName` returns
   `""` rather than erroring for a missing name, which would silently dedupe two
   different nameless findings from opposite sides into one — exactly the data
   loss `resolve` exists to prevent.
8. **`promote` syncs BEFORE its grading report.** `reportGradingAfterPromote`
   can return "promote wrote files but validate failed" — after roadmap.json is
   already on disk. Syncing first means that error path still leaves ROADMAP.md
   in sync instead of drifted.
9. **Three existing tests were updated, none weakened.** `gitStub` now answers
   `diff --name-only` with a source path (a moved HEAD without that answer was
   previously undefined, and is now legitimately fresh); two `rawphase_struct`
   byte-exact renders and one `rawdeps_rewrite` assertion encoded the
   pre-S6 "untouched phases are not reformatted" behavior that AC14 reverses.
   Every one still asserts the same invariant against the new format.

## Verification

- `go build ./...`, `go vet ./...`, `go test ./... -run xxxNONE` — clean.
- Full `go test ./...` — 0 test failures, 0 package failures.
- `./scripts/check-fmt.sh` — exit 0. `centinela docs lint` — 52/52 documented.
- `centinela validate` on the final tree — **All gates passed**: G1 ✓,
  cross-compile ✓ (6 targets), roadmap_drift ✓, docstring ✓, and all three
  validate commands ✓ (the two ⚠ are the pre-existing unmapped-package
  import_graph notice and spec-traceability, which the tests step closes).
- Coverage from that run's profile: `internal/roadmapstate` 100.0%,
  `internal/treestate` 100.0%, `internal/ui` 99.8%, `internal/workflow` 97.4%,
  `internal/gitutil` 97.6%, `cmd/centinela` 96.5%, `internal/roadmap` 95.6%.
  Restricted to the files this feature created or changed: **97.9% (617/630)**.
  `internal/roadmap`'s untouched remainder sits at 94.8%, which is why that
  package total is under 97%.
- Dogfooded with a scratch binary in throwaway repos: pathspec commit with an
  unrelated staged file preserved; `disable_auto_commit`; mid-merge skip;
  `generate` committing nothing; `roadmap resolve` on a REAL two-branch
  conflict; `backlog --stale/--older-than/--json`; and a real-git freshness
  test (`internal/workflow/freshness_realgit_test.go`) proving a roadmap-state
  commit keeps a stamp fresh while a source commit stales it.

## Handoff

To **qa-senior**. What the tests step still owes:

- The executable acceptance tier under `tests/acceptance/` mapping 1:1 to the
  scenarios in `specs/roadmap-state-hygiene.feature` (`// Scenario: <name>`),
  plus `validate.commands` coverage for it.
- `.workflow/roadmap-state-hygiene-edge-cases.md` (the 38 cases recorded in the
  evidence JSON are the starting list).
- Integration coverage for the eight mutation commands this step wired but only
  smoke-tested through `defer` and `promote`: `add`, `remove`, `edit`, `move`,
  `reorder`, `phase add|rename|remove` — each must produce exactly one commit
  whose changed paths are roadmap state only and whose message starts with
  `chore(roadmap): <verb>`.
- AC6 (worktree locality) has no automated test yet: it needs a repo with a
  linked worktree asserting the primary checkout's HEAD, tree and index are
  untouched.
- If the package-level 97% target must be met for `internal/roadmap`, ~23
  statements of PRE-EXISTING error branches need covering (`writeAtomic`,
  `promoteDraftInPlace`, `forceRemovePhase`, `applyReorder`, …); none of them
  are code this feature introduced.

Two roadmap-record follow-ups this step deliberately did NOT make, because both
mutate the shared `.workflow/roadmap.json` that several sessions are writing
right now, and both are cleaner as their own commit on main after this merges:

- Close the Backlog finding `rawio-reformat-diff-churn` — S6 fixes it, and the
  plan says fold it in rather than re-defer it.
- Two gaps found while reading the code but NOT introduced by this feature, and
  worth a `roadmap defer --source roadmap-state-hygiene/senior-engineer` once
  the roadmap file is safe to touch:
  - `roadmap edit --name` renames a feature and rewrites every dependent, but
    leaves that slug's `roadmap-analysis.json` / `roadmap-quality.json` entries
    under the OLD name, orphaning them (`internal/roadmap/edit_rename.go`).
  - `featureName` returns `""` for an entry with no `name` instead of erroring,
    so any caller keying on it silently collapses nameless entries;
    `resolve` now guards this locally, other callers do not
    (`internal/roadmap/rawmutate.go:77`).
