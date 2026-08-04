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

---

# Post-verification fixes (adversarial verifier: WARNING, 7 findings)

Four findings closed on this branch. Each has a regression test at the tier that
caught it, and each was dogfooded before/after against a binary built from the
pre-fix HEAD (`git archive HEAD`, read-only on git).

## F2 HIGH — detached HEAD committed to a dangling commit

`gitutil.CommitBlockedReason` checked repo/HEAD/in-progress but not detachment,
so `git commit` SUCCEEDED onto a commit nothing references, printed
`✓ Committed`, and the record died at the next checkout. The brief's E1 named
"detached" as a warn-and-skip cause; the implementation missed it.

- `internal/gitutil/commit_guard.go` (97) — new `DetachedHead` predicate
  (`git symbolic-ref -q HEAD`), consulted AFTER `InProgressOperation` so a
  rebase — which also detaches — still reports the more specific operation.
- `internal/gitutil/commit_detached_test.go` (48) — real detached repo: the
  reason, the refusal, that HEAD did not move, and that reattaching clears it.

Dogfood: BEFORE `✓ Committed` then the deferral is GONE after `git checkout main`
(grep count 0). AFTER `⚠ ... detached HEAD`, exit 0, deferral still present
(count 1) because it stayed in the working tree, which survives the checkout.

## F1 HIGH — concurrent mutations destroyed a deferral and said otherwise

(a) **Truthful message.** `SyncReport` gained `InWorkingTree`, a VERIFIED
read-back (`roadmap.StateInSync`: re-load roadmap.json, re-render, byte-compare
against ROADMAP.md on disk). `RenderRoadmapSync` may only say "it is in your
working tree" when that check passes; otherwise it prints, in red, that the
state was NOT committed and no longer matches what the command wrote, and names
the file to re-check. `syncRoadmapState` routes that case to stderr too.

(b) **Serialization.** `internal/filelock` is a new stdlib-only leaf holding the
flock primitives EXTRACTED from `internal/evidence` (which has used them since
the evidence CLI shipped) — reuse, not a second implementation; `evidence.Lock`
now delegates to it. `internal/roadmap/mutate_lock.go` takes that lock across
the whole read → modify → write, at all ten exported mutation entry points. A
timeout returns an error and writes NOTHING: exiting non-zero without mutating
is honest and repeatable, whereas proceeding would reintroduce the clobber.

The lock deliberately does NOT live at `.workflow/roadmap.json.lock`. Centinela
does not manage a consumer project's `.gitignore`, so a sibling would leave an
untracked file in every project that ran a mutation — dirtying the tree is the
exact failure this feature exists to remove. `mutate_lock_path.go` puts it in
the git directory (never reported by `git status`, shared by every user of the
checkout, worktree-aware via the `gitdir:` pointer), falling back to the OS temp
dir for the supported no-repository path.

- `internal/filelock/filelock.go` (72), `flock_unix.go` (28),
  `flock_windows.go` (35) — moved from `internal/evidence`.
- `internal/roadmap/mutate_lock.go` (30), `mutate_lock_path.go` (85).
- `internal/roadmap/{defer,add,remove,edit,move,reorder,phase_add,phase_rename,
  phase_remove,promote}.go` — +5 lines each; `promote_backlog.go` (38) split out
  so `promote.go` (70) stays under the cap.
- `internal/evidence/lock.go` (41) — delegates; same message and timeout.

Concurrency test methodology, at BOTH tiers:
- `internal/roadmap/mutate_lock_test.go` (89) — 8 goroutines released
  simultaneously through a barrier channel, each calling `Defer` on ONE
  roadmap.json; asserts all 8 slugs survive, the document still parses, and the
  Backlog holds exactly 8. Plus: a held lock makes a mutation fail with
  "nothing was written" and leaves the file byte-identical.
- `cmd/centinela/roadmap_concurrent_test.go` (69) — three real OS PROCESSES
  (the test binary re-execs itself as a worker via an env var). This is the only
  tier that exercises git's own `index.lock`, and it was a failed `git add`
  under that lock that made the loss silent. Asserts all three slugs survive and
  that no worker claimed the tree holds a record that is gone.
- Negative control, run both ways: with `lockRoadmapState` neutralized (byte-copy
  backup, `cmp`-verified restore — no `git checkout`), the goroutine test fails
  (8 defers collapse to 1) and the process test fails 3/3.

Dogfood, 5 runs each: BEFORE destroyed a deferral **5/5**, every time with
`git status` EMPTY and the loser printing "left uncommitted". AFTER **0/5**,
both records present, 2 commits, `git status` clean (no lock file in the tree).

## F3 MEDIUM — S6 normalization landed as its own commit

Commit `adb8485` `chore(roadmap): canonicalize roadmap.json rendering`.

Produced by the SHIPPED renderer (a semantically null `phase add` + `phase
remove` in a throwaway clone), then committed here on a pathspec so the rest of
the working tree stayed untouched (index empty before and after).

- Diff size: **1 file changed, 66 insertions(+), 399 deletions(-)**.
  `ROADMAP.md` is byte-identical — it renders from the typed document, not the
  raw layout.
- Semantic equivalence: key-sorted JSON before/after is byte-identical (`cmp`);
  15 phases, 176 features, both top-level keys, `intro` preserved, and every
  phase/feature key round-trips (`archetype, deferredAt, dependsOn, description,
  features, fixes, name, note, source, summary`). A second render is
  byte-identical (idempotent).
- Payoff, measured: each of the five deferrals recorded afterwards is a
  **2 files changed, 3 insertions(+), 1 deletion(-)** commit instead of dragging
  a 399-line reformat behind it.

## F4 MEDIUM — `resolve` dropped a one-sided edit in a modify/delete

`survivor` decided on presence alone, so "deletion wins over an UNTOUCHED side"
was applied to a CHANGED side too: ours edits a finding, theirs deletes it, and
the edit vanished with exit 0 and a `kept 0` summary. It now refuses that pair
by slug — the same contract as a both-sides phase edit — via a new
`FindingConflictError`. A reformatted-but-unchanged side is still a delete: the
comparison is over compacted JSON.

- `internal/roadmap/resolve.go` (92), `resolve_backlog.go` (92).
- `internal/roadmap/resolve_modifydelete_test.go` (77) — both directions, the
  untouched-side case still working, the whitespace case, and that a refusal
  returns no document.
- `cmd/centinela/roadmap_resolve_refuse_test.go` (63) — exit non-zero, markers
  byte-identical, nothing staged, no ROADMAP.md written.

Dogfood on a REAL two-branch conflict: BEFORE `✓ ... kept 0 findings`, exit 0,
`IMPORTANT-NEW-DETAIL` gone, markers removed. AFTER exit 1 naming `finding-A`,
the edit still present, markers intact, and `git ls-files -s` byte-identical
across the refusal (the three staged entries there are git's own merge stages).

## Deferrals recorded in-branch

Five, all `--source roadmap-state-hygiene/gatekeeper`, each its own
pathspec-scoped `chore(roadmap): defer <slug>` commit:
`roadmap-resolve-summary-arithmetic-unreconcilable`,
`roadmap-state-exemption-covers-all-workflow-dir`,
`roadmap-phase-rename-commit-subject-loses-old-name`,
`roadmap-edit-rename-leaves-analysis-quality-stale`,
`roadmap-featurename-empty-not-an-error`.
None recorded for F1-F4 — those are fixed here.

## Divergences and one thing NOT done

- The lock is a new stdlib leaf rather than a copy of `evidence`'s: `evidence`
  is a domain package `internal/roadmap` must not import, so the primitive moved
  down instead of being duplicated. `internal/filelock` is declared in the
  `import_graph` leaf layer and justified in PROJECT.md G2 alongside
  `internal/roadmapstate`.
- `centinela evidence validate <feature>` exits 1 on ONE issue, and it is the
  GATEKEEPER's `.json` (0 inputs, 0 outputs, 0 edgeCases) — the verifier wrote
  its `.md` but never populated its evidence record. That is another role's
  artifact; filling it in would be fabricating a verifier's inputs, so it is
  left for the verifier or the coordinator. This role's own evidence is
  complete (18 inputs, 53 outputs, 48 edge cases, handoffTo qa-senior).
- `.workflow/roadmap-state-hygiene-edge-cases.md` gained the four cases the
  verifier proved were missing (E17 concurrency, the lock's tree-cleanliness,
  E1b detached HEAD, E12a modify/delete). E17 was previously CLAIMED and untrue;
  it is now true and covered at two tiers.

## Spec / brief / acceptance updates forced by the truthful-message fix

The phrase "left uncommitted" was mandated by AC4, AC5, two `.feature`
scenarios and six acceptance assertions. It is exactly the claim the verifier
proved can be false, so the wording changed everywhere rather than only in the
code — leaving the spec demanding a phrase we deliberately removed would be the
same class of untrue artifact the verifier penalised.

- `tests/acceptance/roadmap_state_hygiene_{disable,hostile}_test.go` — assert
  the CLAIM ("not committed", the reason, and "in your working tree" only
  because the read-back verified it), not the old phrasing. Nothing weakened:
  both still require exit 0, the record on disk, ROADMAP.md regenerated and no
  commit.
- `specs/roadmap-state-hygiene.feature` — AC4/AC5 lines reworded; four
  scenarios ADDED for the fixed behaviours (lost record never reported as
  uncommitted, concurrent mutations, detached HEAD, resolve modify/delete).
- `docs/features/roadmap-state-hygiene.md` — AC4/AC5 reworded, E17 corrected
  from a false claim to the implemented one, E19 (detached HEAD) and E20
  (modify/delete) added.
- New acceptance coverage for all four scenarios, binary-driven against real
  repos: `roadmap_state_hygiene_{detached,concurrent,modifydelete}_test.go`
  (45/63/56 lines). The concurrency one runs three real `centinela roadmap
  defer` PROCESSES and also asserts the tree is clean afterwards (the lock must
  never become an untracked file). `spec-traceability-gate` reports all 39
  scenarios covered.

---

# Round-2 fixes (verifier: CRITICAL, NO-SHIP — 2 silent-loss paths)

## CRITICAL — the lock was keyed on an unresolved absolute path

`filepath.Abs` never resolves symlinks, so ONE roadmap.json reached as
`/real/proj/...` and as `/link/proj/...` hashed to two different lock names.
Two lock files were created side by side in the SAME `.git`, excluding nothing,
and the write race was silently back — with a green `✓ Committed` printed for a
record that had been overwritten and `git status` clean. macOS symlinks `/tmp`
and `/var`, so this is an everyday path spelling.

- `internal/roadmap/mutate_lock_canon.go` (43, new) — `canonicalPath` resolves
  the FILE when it exists, else its DIRECTORY (so a not-yet-created roadmap.json
  still canonicalizes), falling back to the absolute path on any failure: a lock
  under an unresolved name is still a lock, and nothing here may skip locking.
  `canonicalDir` gives the lock DIRECTORY one spelling too, so two processes
  land on the same lock path, not merely the same inode.
- `internal/roadmap/mutate_lock_path.go` (82) — hashes the canonical path and
  canonicalizes the resolved git dir at both return sites.
- Tests: `mutate_lock_symlink_test.go` (65) — lock-path invariance across a
  symlink AND explicitly across the system `/tmp -> /private/tmp` shape;
  `mutate_lock_symlink_race_test.go` (48) — 8 concurrent mutations, half
  entering by the real path and half by the symlink, all must survive.
  `mutate_lock_path_test.go` (74) + `mutate_lock_fallback_test.go` (37) now
  compare against canonical directories (t.TempDir() is itself a symlink on
  macOS — the fix is why they had to change).
- Negative control (byte-copy backup, `cmp`-verified restore): with
  canonicalization disabled the race test loses 4 of 8 records and both
  invariance tests fail.

## MEDIUM — `resolve` discarded a one-sided incoming edit

`survivor` compared against base only in the delete branch; when both sides
still held the slug it delegated to `earlier()`, which orders by `deferredAt`
alone. An edit does not move `deferredAt`, so the comparison tied and OURS won —
a finding THEIRS edited and OURS left untouched resolved to the BASE version,
behind a green `✓ Resolved … kept 3 findings`.

- `internal/roadmap/resolve_survivor.go` (65, split out) — the both-present
  branch now three-ways against base (`threeWay(was, mine, yours)`) exactly like
  a phase; `earlier()` decides ONLY the case the base never had (both sides
  independently deferring one slug). Both-sides-different-edits is refused.
- `internal/roadmap/resolve_errors.go` (37, split out) — `FindingConflictError`
  gained `Detail` so one type covers both irreconcilable shapes without losing
  the specific wording ("was modified on one side and deleted on the other" /
  "was modified differently on both sides").
- `internal/roadmap/resolve.go` (67), `resolve_backlog.go` (45),
  `resolve_order.go` (72, `earlier`'s contract narrowed in its doc comment).
- Tests: `resolve_onesided_edit_test.go` (79) — theirs-edit wins, ours-edit
  wins, both-edited refuses by name, identical edits merge, and `earlier()`
  still decides independent captures.
  `tests/acceptance/roadmap_state_hygiene_incoming_edit_test.go` (83) — both
  scenarios through a REAL `git merge` conflict.
- Two scenarios added to the `.feature`; `spec-traceability-gate` reports all
  **41** scenarios covered.

## Reproductions (isolated: baseline = current tree with ONLY the one fix reverted)

**Symlink race**, 4 concurrent processes x 6 rounds, half via a symlinked path:

| | records destroyed | lock files in `.git` | loser printed |
|---|---|---|---|
| BEFORE (Abs-keyed) | **10 of 24** | 2 | green `✓ Committed roadmap state:` for an absent record |
| AFTER (canonical) | **0 of 24** | 1 | — |

`git status` was clean in every BEFORE round — the "lost record, clean tree,
green checkmark" signature.

**One-sided incoming edit**, real `git merge` conflict, base `summary:"orig"`,
theirs `EDITED-BY-THEIRS`, ours untouched:

| | exit | their edit present | base "orig" present | both additions unioned |
|---|---|---|---|---|
| BEFORE | 0 (`✓ Resolved … kept 3 findings: 1 from our side, 1 from theirs`) | **0** | 1 | yes |
| AFTER | 0 (same summary line) | **1** | 0 | yes |

## Deferrals recorded in-branch

Five, `--source roadmap-state-hygiene/gatekeeper`, each a
`2 files changed, 3 insertions(+), 1 deletion(-)` commit:
`roadmap-sync-readback-false-overwrite-alarm`,
`roadmap-statesync-readback-lacks-record-identity`,
`roadmap-rebase-warning-omits-abort-hazard`,
`roadmap-resolve-write-outside-mutation-lock`,
`roadmap-resolve-keyorder-false-conflict`.
None for the two fixed here. Ten gatekeeper findings are now in the Backlog
across both rounds.

## Divergences

- `FindingConflictError` gained a `Detail` field rather than a second error
  type, so round-1's assertions and the modify/delete acceptance test keep their
  exact wording while the new shape gets its own.
- Three files were split to stay under the 100-line cap
  (`resolve_errors.go`, `resolve_survivor.go`, `mutate_lock_fallback_test.go`);
  no behaviour moved with them.
- `mutate_lock_path_test.go`'s directory assertions had to compare against
  canonical paths. That is the fix working: `t.TempDir()` returns
  `/var/folders/...`, whose real location is `/private/var/folders/...`, which
  is exactly the class of divergence the CRITICAL finding was about.

---

# Round-3 fix (verifier: WARNING → SHIP, one new MEDIUM)

## MEDIUM — `resolve` dropped a one-sided edit to the Backlog phase SHELL

`backlogPhase` built the merged phase by unmarshalling theirs' object and then
ours' over the top, so any key BOTH sides carried resolved to ours regardless of
the base. `features` was overwritten afterwards and was safe; every other key —
`note` above all, a first-class `Phase` field rendered as the ROADMAP.md phase
blockquote — was not. An incoming `note` edit vanished with exit 0 behind a
green `✓ Resolved`. Same silent-loss class as the two CRITICALs, one level up:
every OTHER phase (`mergePhase`) and every top-level key (`mergeRest`) already
three-wayed correctly; only the Backlog shell was exempt.

- `internal/roadmap/resolve_shell.go` (64, new) — `mergeShellKeys` three-ways
  every shell key against base and refuses a genuine both-sides divergence with
  `PhaseConflictError`; `side.phaseObject` decodes a side's phase, refusing a
  non-object body by side and phase name.
- `internal/roadmap/resolve_order.go` (67) — `backlogPhase` now takes the base
  side and delegates the shell to `mergeShellKeys`.
- `internal/roadmap/resolve_backlog.go` (46) — passes the base through.
- Tests: `resolve_shell_test.go` (94) — theirs-edits/ours-untouched,
  the inverse, both-edit-differently refusing, and unknown-key survival with
  `features` still owned by the merged list; `resolve_shell_errors_test.go` (78)
  — malformed side on each of the three stages, and the no-phase case.
  `tests/acceptance/roadmap_state_hygiene_shell_test.go` (94) — all three
  through a REAL `git merge` conflict, including that the merged note reaches
  the regenerated ROADMAP.md.
- Negative control (byte-copy backup, `cmp`-verified restore): with the shell
  three-way reverted, `TestResolveKeepsAOneSidedShellEditFromTheirs` and
  `TestResolveRefusesABothSidesShellEdit` both fail.

## Attribution FIXED, not deferred

The three-way rewrite made correct attribution cheap, so Finding 2 is closed
rather than left deferred, and its Backlog entry was removed.

`countContribution` counted by presence of the SLUG and skipped anything the
base already had, which told two lies at once: the counts never summed to
`Kept`, and a slug both sides deferred was credited to ours even when the
earlier capture kept was THEIRS. It now credits the surviving ENTRY to the side
whose BYTES it is — a comparison that cannot misattribute — and `Merged` gained
`FromBase`, so the arithmetic partitions `Kept` exactly.

- `internal/roadmap/resolve_survivor.go` (71), `resolve.go` (68),
  `internal/ui/render_resolve.go` (34): the line is now
  `kept N findings: A unchanged, B from our side, C from theirs.` with
  `A + B + C == N`.

## Coverage gap the verifier named (process tier PLUS symlink)

My round-2 symlink test was in-process and my process test had no symlink, so
neither exercised the combination that actually broke. Closed:
`tests/acceptance/roadmap_state_hygiene_symlink_race_test.go` (79) runs four
real `centinela roadmap defer` PROCESSES, half entering through a symlinked
directory, and asserts all four records survive, that exactly ONE lock file
exists, and that the tree is clean.

## Reproductions (baseline = current tree with ONLY these changes reverted)

**Phase shell `note`, both directions:**

| direction | | exit | note kept | summary |
|---|---|---|---|---|
| theirs edits, ours untouched | BEFORE | 0 | **ORIGINAL-NOTE** (their edit gone) | `kept 3 findings: 1 from our side, 1 from theirs` (2 ≠ 3) |
| | AFTER | 0 | **THEIRS-EDITED-NOTE** | `kept 3 findings: 1 unchanged, 1 from our side, 1 from theirs` (3 = 3) |
| ours edits, theirs untouched | BEFORE | 0 | OURS-EDITED-NOTE (right by luck — ours always won) | — |
| | AFTER | 0 | OURS-EDITED-NOTE | — |
| both edit differently | BEFORE | **0** | OURS-EDITED-NOTE (theirs silently lost) | green `✓ Resolved … kept 1 findings: 0 from our side, 0 from theirs` |
| | AFTER | **1** | markers left, refused | `phase "Backlog" changed on both sides` |

**Provenance misattribution** (both sides deferred one slug, theirs' capture is
the earlier one):

| | summary | what actually survived |
|---|---|---|
| BEFORE | `kept 1 findings: 1 from our side, 0 from theirs` | `THEIR-DIFFERENT-ANALYSIS` |
| AFTER | `kept 1 findings: 0 unchanged, 0 from our side, 1 from theirs` | `THEIR-DIFFERENT-ANALYSIS` |

## Backlog changes

- Recorded: `roadmap-mutation-commit-races-git-index-lock` (LOW — the commit
  runs after the roadmap lock is released, so concurrent mutations race git's
  `index.lock`; no record is lost and the read-back keeps every message honest,
  but a mutation can exit 0 having committed nothing).
- REMOVED `roadmap-resolve-summary-arithmetic-unreconcilable` — fixed above, so
  restating it would have recorded a debt that no longer exists.

## Divergences

- The both-sides shell divergence is refused as a `PhaseConflictError` naming
  `Backlog` rather than a new error type: it IS a phase changed on both sides,
  and the existing message ("roadmap resolve only merges divergent Backlog
  findings") is exactly right — the findings merged, the shell did not.
- `Merged` gained a field (`FromBase`) rather than changing the meaning of the
  two existing counts, so a reader of the struct cannot silently get the old
  semantics. Three assertions that encoded the unreconcilable arithmetic were
  updated to require reconciliation.

