# Edge Cases: spec-conflict-false-positives

## Covered

### False positives eliminated (the hotfix's core contract)

- Byte-identical spec copied by an idle bystander worktree —
  `internal/worktree/spec_conflicts_baseline_test.go:TestDetectSpecConflicts_BystanderCarriesMainBaseline_NoFlag`
- Merging worktree itself carries only main's untouched baseline —
  `internal/worktree/spec_conflicts_baseline_test.go:TestDetectSpecConflicts_MergingCarriesMainBaseline_NoFlag`
- Byte-identical copies of one file across main + two worktrees, including two
  companion scenarios sharing one Given clause —
  `internal/worktree/spec_conflicts_falsepos_test.go:TestDetectSpecConflicts_IdenticalCopiesAndCompanionScenarios_NoFlag`
- Two different files on main sharing a Given clause —
  `internal/worktree/spec_conflicts_falsepos_test.go:TestDetectSpecConflicts_CrossFileSameGivenOnMain_NoFlag`
- Main vs. the merging worktree disagreeing (supersession, not a conflict) —
  `internal/worktree/spec_conflicts_falsepos_test.go:TestDetectSpecConflicts_MainVersusMergingIsSupersession_NoFlag`
- Two different files owned by two different worktrees sharing a scenario name —
  `internal/worktree/spec_conflicts_falsepos_test.go:TestDetectSpecConflicts_DifferentFilesDifferentOwners_NoFlag`
- Agreeing scenarios in differently-named files never conflict —
  `internal/worktree/spec_conflicts_more_test.go:TestDetectSpecConflicts_SameGivenSameThen_NoFlag`
- Intra-feature scenarios (same owner) are never compared to each other —
  `internal/worktree/spec_conflicts_more_test.go:TestDetectSpecConflicts_SameOwnerIsNotConflict`
- End-to-end, real binary: main + merging worktree with byte-identical specs
  (incl. companion scenarios sharing a Given) + a bystander worktree carrying
  main's copies → real `centinela merge` succeeds, pre-check reports nothing —
  `tests/acceptance/spec_conflict_binary_test.go:TestAcceptance_SpecConflict_IdenticalSpecsAcrossWorktreesDoNotBlockRealMerge`
- Same scenario, same acceptance surface (function-driven) —
  `tests/acceptance/parallel_feature_worktrees_test.go:TestParallelWorktrees_IdenticalAndSupersedingSpecsDoNotBlock`

### True positives still caught

- Two worktrees diverging on `Then` for the same (file, scenario) —
  `internal/worktree/spec_conflicts_test.go:TestDetectSpecConflicts_TwoWorktreesDivergentThen_Flags`
- Two worktrees diverging on `Given` for the same (file, scenario) —
  `internal/worktree/spec_conflicts_more_test.go:TestDetectSpecConflicts_TwoWorktreesDivergentGiven_Flags`
- Both worktrees edited away from a present main baseline, in different
  directions —
  `internal/worktree/spec_conflicts_baseline_test.go:TestDetectSpecConflicts_BothEditedBaseline_Flags`
- No main baseline exists; two worktrees introduce the scenario differently —
  `internal/worktree/spec_conflicts_baseline_test.go:TestDetectSpecConflicts_NoBaselineBothIntroduced_Flags`
- A missing `Then` counts as a divergence —
  `internal/worktree/spec_conflicts_collect_test.go:TestDetectSpecConflicts_IncompleteScenarioIgnored`
- End-to-end, real binary: two worktrees genuinely diverging on the same
  (file, scenario) block the real `centinela merge`, name both worktrees,
  report the scenario once per error print, main HEAD does not advance, and
  the worktree under merge is kept — `tests/acceptance/spec_conflict_binary_test.go:TestAcceptance_SpecConflict_TwoWorktreesDivergeBlocksRealMergeOnce`
- Same true positive via `runMerge` in-process —
  `cmd/centinela/merge_test.go:TestRunMerge_SpecConflict_Blocks`
- Same true positive via the pure detector + formatter —
  `tests/acceptance/parallel_feature_worktrees_test.go:TestParallelWorktrees_SpecConflictDetectedPreMerge`

### Dedup and output bounding (the 720KB regression)

- A scenario name repeated within one file is reported once, not once per
  copy —
  `internal/worktree/spec_conflicts_format_test.go:TestDetectSpecConflicts_RepeatedScenarioReportedOnce`
- Each other in-flight worktree is its own pairing, reported separately —
  `internal/worktree/spec_conflicts_format_test.go:TestDetectSpecConflicts_TwoOtherWorktrees_OnePerPair`
- `FormatSpecConflicts` caps at 10 entries plus an "and N more" suffix for 14
  true conflicts —
  `internal/worktree/spec_conflicts_format_test.go:TestFormatSpecConflicts_CapsLongReports`
- Duplicate scenario names within one file cannot manufacture extra pairings
  (`indexByKey` keeps the first record) —
  `internal/worktree/coverage_merge_helpers_test.go:TestIndexByKey_FirstRecordWins`
- End-to-end, real binary: blocked merge output stays under a 4096-byte sanity
  cap and prints exactly one conflict entry per error print (arrows == prints)
  —
  `tests/acceptance/spec_conflict_binary_test.go:TestAcceptance_SpecConflict_TwoWorktreesDivergeBlocksRealMergeOnce`

### Malformed / partial input handling

- No specs directory at all (missing worktree or missing `specs/`) —
  `internal/worktree/spec_conflicts_test.go:TestDetectSpecConflicts_NoSpecsDirectory_NoError`
- Empty conflict list formats to an empty string —
  `internal/worktree/spec_conflicts_test.go:TestFormatSpecConflicts_Empty`
- Non-`.feature` files and nested directories inside `specs/` are ignored —
  `internal/worktree/spec_conflicts_collect_test.go:TestDetectSpecConflicts_IgnoresNonFeatureEntries`
- An unreadable spec file (chmod 0000) is skipped, not fatal —
  `internal/worktree/spec_conflicts_collect_test.go:TestDetectSpecConflicts_UnreadableSpecSkipped`
- A dangling symlink named `*.feature` is skipped —
  `internal/worktree/coverage_merge_helpers_test.go:TestReadSpecsFrom_UnreadableEntrySkipped`
- A worktree directory that carries no specs contributes nothing —
  `internal/worktree/spec_conflicts_collect_test.go:TestDetectSpecConflicts_WorktreeWithoutSpecsIgnored`
- Invalid feature slug rejected before any spec I/O —
  `cmd/centinela/merge_test.go:TestRunMerge_InvalidSlug_Errors`

## Residual Risks

- **Scenario deletion is invisible.** A worktree that deletes a scenario
  another worktree edited is not detected as a conflict (the deleting side
  simply has no record to compare). Captured as Backlog item
  `spec-conflict-scenario-deletion-detection`
  (`.workflow/roadmap.json`, deferred by senior-engineer).
- **Only the first `Given` and first `Then` per scenario are compared.**
  Divergence in a `When` step, an `And` step, or a second `Then` is invisible
  to this detector. Captured as Backlog item
  `spec-conflict-deep-gherkin-diff` (deferred by senior-engineer).
- **Detection only sees worktrees on disk.** `DetectSpecConflicts` returns
  nil immediately when the merging feature's own `.worktrees/<feature>/specs`
  is absent or empty (`internal/worktree/spec_conflicts.go:38`), so a feature
  merged after its worktree was already removed (e.g. merged via a raw PR
  outside `centinela merge`) contributes no pre-check at all. This is a
  pre-existing property of the worktree-based workflow (the detector has no
  other source of "what this feature changed" than the on-disk checkout), not
  a regression introduced by this hotfix, and every acceptance/binary test in
  this suite exercises the on-disk path exclusively. Not deferred as a new
  finding — it is the same constraint the senior-engineer report flagged for
  reviewer attention (`.workflow/spec-conflict-false-positives-senior-engineer.md`,
  Handoff item 2); accepted as-is for this hotfix's scope.
- **Cross-file semantic contradictions are no longer detected**, by design —
  the old Given-bucketing pairer that caught these was also the entire source
  of false positives. `centinela validate` on the merged tree remains the
  real semantic gate for this class. Documented as a deliberate trade-off in
  the senior-engineer report, not re-deferred here.
