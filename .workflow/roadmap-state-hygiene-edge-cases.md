# Edge Cases: roadmap-state-hygiene

The 38 cases below are the starting list recorded in the senior-engineer's
evidence JSON (`.workflow/roadmap-state-hygiene-senior-engineer.json`,
`edgeCases`). Each is mapped to the executable assertion that covers it —
either the new acceptance tier this step adds, or a pre-existing colocated
unit/`cmd` test the engineer wrote alongside the S1-S6 implementation.

## Covered

1. Commit fails after a successful mutation (signing/hook): warn on stderr,
   exit 0, roadmap.json/ROADMAP.md still correct —
   `tests/acceptance/roadmap_state_hygiene_hostile_test.go:TestRsh_HostileGitEnvironmentWarnsNeverFails`
   (case "git commit fails"); `internal/gitutil/commit_errors_test.go:TestCommitPathsReportsASigningFailure`.
2. Merge/rebase/cherry-pick/revert in progress: commit skipped with the named
   reason, exit 0 —
   `tests/acceptance/roadmap_state_hygiene_hostile_test.go:TestRsh_HostileGitEnvironmentWarnsNeverFails`
   (merge/rebase cases); `internal/gitutil/commit_guard_test.go` (cherry-pick/revert).
3. Unrelated staged and unstaged files present: pathspec commit leaves them
   byte-identical —
   `tests/acceptance/roadmap_state_hygiene_unrelated_test.go:TestRsh_UnrelatedStagedAndUnstagedChangesSurvive`.
4. roadmap.json already dirty from a hand edit: folded into the mutation's
   read-modify-write and committed together —
   `internal/roadmap/statesync_test.go`, `internal/roadmap/statesync_edges_test.go`.
5. Not a git repository or no HEAD: regeneration only, one warning, exit 0 —
   `tests/acceptance/roadmap_state_hygiene_hostile_test.go:TestRsh_HostileGitEnvironmentWarnsNeverFails`.
6. `disable_auto_commit` set: ROADMAP.md still regenerated, commit skipped —
   `tests/acceptance/roadmap_state_hygiene_disable_test.go:TestRsh_DisableAutoCommitSkipsCommitNotRegeneration`;
   `tests/acceptance/roadmap_state_hygiene_freshness_test.go:TestRsh_UncommittedRegeneratedMarkdownStaysFresh`.
7. No-op mutation: git status over the pathspec is empty, no empty commit —
   `tests/acceptance/roadmap_state_hygiene_noop_test.go:TestRsh_NoopMutationCreatesNoCommit`.
8. Declared pathspec entry never written (promote artifacts absent): dropped,
   not fatal — `internal/gitutil/commit_test.go`, `internal/gitutil/commit.go:existingPaths`.
9. Mutation deleted a tracked roadmap-state path: kept via `git ls-files` so
   the removal is committed —
   `internal/gitutil/commit_errors_test.go:TestCommitPathsCommitsADeletedTrackedPath`.
10. Mutation inside `.worktrees/<feature>`: commit lands on that worktree's
    HEAD; primary checkout untouched —
    `tests/acceptance/roadmap_state_hygiene_worktree_test.go:TestRsh_MutationInWorktreeNeverTouchesPrimaryCheckout`.
11. Promote also rewrites analysis/quality artifacts: pathspec supplied by the
    mutation, asserted against `PromoteArtifactPaths` —
    `tests/acceptance/roadmap_state_hygiene_promote_test.go:TestRsh_PromoteCommitsAnalysisAndQualityArtifacts`.
12. Deferral after artifact stamp: the recorded..HEAD range is all roadmap
    state, verification stays fresh —
    `tests/acceptance/roadmap_state_hygiene_freshness_test.go:TestRsh_DeferralAfterStampStaysFresh`;
    `internal/workflow/freshness_realgit_test.go`.
13. Uncommitted regenerated ROADMAP.md (`disable_auto_commit`): dropped from
    the tree digest, stamp survives — see case 6.
14. Source file committed in the same range as roadmap state: mixed range
    stales with today's message —
    `tests/acceptance/roadmap_state_hygiene_freshness_test.go:TestRsh_SourceChangeAfterStampStillStales`;
    `internal/workflow/validate_freshness_test.go:TestVerificationStalesOnASourceChangeBesideRoadmapState`.
15. Unresolvable recorded revision or any git error on the range: fails
    closed, reported stale —
    `tests/acceptance/roadmap_state_hygiene_freshness_test.go:TestRsh_UnreadableRevisionRangeFailsClosed`;
    `internal/workflow/freshness_range_test.go:TestRevisionRangeStateOnlyDecisionTable`.
16. Empty recorded revision: never exempt, git is not consulted —
    `internal/workflow/freshness_range_test.go:TestRevisionRangeStateOnlyRejectsAnEmptyRecordedRevision`.
17. Empty revision range (identical trees): fresh — `freshness_range_test.go`
    decision table, "empty range" case.
18. Roadmap-lookalike paths (`docs/ROADMAP.md`, `ROADMAP.md.bak`,
    `workflow/roadmap.json`): never exempt — `internal/roadmapstate/paths_test.go`.
19. Backlog finding with no `deferredAt`: age unknown, sorted first, counted
    stale, `ageDays` null in JSON —
    `tests/acceptance/roadmap_state_hygiene_backlog_list_test.go:TestRsh_BacklogUnknownDeferredAtSortsFirstAndCountsStale`;
    `tests/acceptance/roadmap_state_hygiene_backlog_json_test.go:TestRsh_BacklogJSONEmitsMachineShape`.
20. Backlog finding with an unparseable `deferredAt`: identical to missing,
    never an error — same test as case 19 (fixture carries both a missing
    and an unparseable `"not-a-date"` entry side by side).
21. Finding deferred exactly N days ago: NOT stale (strict `>`) —
    `internal/roadmap/backlog_age_test.go` (boundary case).
22. Empty or absent Backlog phase: explicit empty state, exit 0 —
    `tests/acceptance/roadmap_state_hygiene_backlog_json_test.go:TestRsh_EmptyBacklogIsAnExplicitEmptyState`.
23. Non-empty Backlog with nothing past the threshold: `--stale` names the
    threshold and total, exit 0 — `internal/ui/render_backlog_list_test.go`.
24. Nudge suppressed while any SCHEDULABLE feature is incomplete (Backlog
    entries never count as schedulable work) —
    `tests/acceptance/roadmap_state_hygiene_nudge_test.go:TestRsh_NoNudgeWhileScheduleableWorkRemains`
    and `TestRsh_CompletingLastFeatureNudgesAboutBacklog` (a non-empty Backlog
    phase does not block the nudge once schedulable work is done).
25. Nudge suppressed and `complete` unaffected when roadmap.json cannot be
    parsed —
    `tests/acceptance/roadmap_state_hygiene_nudge_empty_test.go:TestRsh_UnreadableRoadmapSuppressesNudge`.
26. Resolve: both sides added the same slug, deduped keeping the earlier
    `deferredAt`, counted once —
    `tests/acceptance/roadmap_state_hygiene_resolve_dedup_test.go:TestRsh_ResolveKeepsOneEntryForASharedSlug`.
27. Resolve: one-sided deletion wins over an unchanged side; both-sides
    deletion stays deleted —
    `tests/acceptance/roadmap_state_hygiene_resolve_dedup_test.go:TestRsh_ResolveHonoursAOneSidedDeletion`
    (one-sided); `internal/roadmap/resolve_more_test.go` (both-sides).
28. Resolve: both sides changed a schedulable phase, refused by name with
    markers and index intact —
    `tests/acceptance/roadmap_state_hygiene_resolve_refuse_test.go:TestRsh_ResolveRefusesOnARealPhaseConflict`.
29. Resolve: both sides changed a top-level key (`intro`), refused by name —
    `internal/roadmap/resolve_test.go`.
30. Resolve: one side of the conflict is not valid JSON, refused naming that
    side, nothing staged —
    `tests/acceptance/roadmap_state_hygiene_resolve_refuse_test.go:TestRsh_ResolveRefusesAnUnparseableSide`.
31. Resolve: a Backlog entry with no name is refused rather than keyed on the
    empty slug — `internal/roadmap/resolve_edges_test.go`.
32. Resolve: absent merge stage (one side added/deleted the file) reads as an
    empty roadmap — `internal/roadmap/resolve_stages_test.go`.
33. Resolve: conflicted path with no readable content on either side is
    refused — `internal/roadmap/resolve_stages_test.go`.
34. Resolve: whitespace-only differences are not semantic conflicts —
    `internal/roadmap/resolve_side.go:compactJSON`, exercised by
    `internal/roadmap/resolve_test.go`.
35. Resolve: only ROADMAP.md conflicted, regenerated from the already-merged
    roadmap.json and staged —
    `tests/acceptance/roadmap_state_hygiene_resolve_misc_test.go:TestRsh_ResolveRegeneratesAMarkdownOnlyConflict`.
36. Resolve: nothing conflicted, no-op exit 0, no file modified or staged —
    `tests/acceptance/roadmap_state_hygiene_resolve_misc_test.go:TestRsh_ResolveIsANoopOutsideAConflict`.
37. Canonical render is idempotent and preserves unknown top-level, per-phase
    and per-feature fields —
    `tests/acceptance/roadmap_state_hygiene_canonical_test.go:TestRsh_RenderingIsIdempotent`.
38. Canonical render of a phase with no `features` key normalizes to an empty
    array and stays valid JSON — `internal/roadmap/rawrender_canonical_test.go`.

## Additional cases found and covered this step

- All 9 roadmap-mutating commands (`add`, `remove`, `edit`, `move`, `reorder`,
  `promote`, `phase add`, `phase rename`, `phase remove`) drive exactly one
  commit whose changed paths are roadmap-state-only, through the REAL compiled
  binary — not just `defer`/`promote` as the code step smoke-tested —
  `tests/acceptance/roadmap_state_hygiene_mutations_test.go:TestRsh_EveryMutationRegeneratesAndCommits`.
- `roadmap resolve` against a REAL two-branch `git merge` conflict (not a
  stubbed `StageRunner`) unions the Backlog by slug with no markers surviving,
  and separately refuses when a schedulable phase genuinely diverged on both
  sides — `roadmap_state_hygiene_resolve_union_test.go`,
  `roadmap_state_hygiene_resolve_refuse_test.go`.

## Residual Risks

- **`internal/roadmap`'s pre-existing coverage remainder.** The
  senior-engineer's report notes ~23 statements of PRE-EXISTING error
  branches (`writeAtomic`, `promoteDraftInPlace`, `forceRemovePhase`,
  `applyReorder`, …) not introduced by this feature keep the package total
  under the 97% target even though every new S1-S6 file is fully covered.
  Out of scope for this step; tracked as a pre-existing gap, not a new one.
- **Two confirmed-but-unrecorded pre-existing gaps** the senior-engineer found
  but deliberately did not defer via `centinela roadmap defer` (deferring
  would mutate the shared `.workflow/roadmap.json` and, via S6's canonical
  renderer, reformat the ENTIRE live file — the same reformatting-collision
  risk R4 flags). Confirmed by direct code read during this step:
  - `roadmap edit --name` renames a feature and rewrites every dependent, but
    leaves that slug's `roadmap-analysis.json`/`roadmap-quality.json` entries
    under the OLD name — `internal/roadmap/edit_rename.go:applyRename` never
    touches the analysis/quality files.
  - `featureName` (`internal/roadmap/rawmutate.go`) returns `""` for an entry
    with no `name` field instead of erroring; `resolve`'s `indexFindings`
    guards this locally, but no other caller does. Mitigation: recorded here
    rather than run through `roadmap defer` in this worktree, per the explicit
    instruction to avoid a live-repo mass-reformat commit unrelated to this
    feature's diff. Should be deferred with
    `centinela roadmap defer <slug> --source roadmap-state-hygiene/qa-senior`
    by an operator once the roadmap file is safe to touch (e.g. from the
    primary checkout, ideally right after the S6 normalization commit the
    engineer also recommended landing separately).
- **`roadmap edit --description` to the identical existing value** is the
  no-op vehicle this suite uses (AC7); a rename-only no-op
  (`--name` equal to the current slug) takes the same `editIsNoop` branch but
  is not separately re-driven through the acceptance tier — covered at the
  `internal/roadmap/edit.go` unit level instead.
