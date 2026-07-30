# Edge Cases: token-diet

## Covered

- Empty `inputs` list on otherwise-complete evidence trips the generic
  "incomplete evidence fields" branch before the plan-snapshot check ever
  runs (see Residual Risks). The per-path "missing feature-doc snapshot
  inputs" message for a *reachable* empty-inputs case is asserted at the unit
  tier: `internal/orchestration/plan_snapshot_test.go:TestValidatePlanSnapshotInputsNamesEachMissingPath`.
  The real, observable CLI behavior for a wholly-empty list is asserted at
  the acceptance tier: `tests/acceptance/token_diet_plan_evidence_test.go:TestTD_EvidenceEmptyInputsRejectedNamingBothPaths`.
- `evidence validate <feature>` always passes nil `uiPaths` (no CLI flag
  exists for them), so a ux-ui-specialist evidence file can never pass the
  "real UI file" output check through that specific subcommand alone. Tests
  call the exported `orchestration.ValidateEvidence` directly with uiPaths,
  the same way `centinela validate`'s internal call path does:
  `tests/acceptance/token_diet_ux_tag_test.go`,
  `tests/acceptance/token_diet_ux_role_test.go`.
- Digest-state directory (`.workflow/`) unwritable: hook still exits 0,
  still prints the roadmap line, surfaces no error to the host session.
  `tests/acceptance/token_diet_hook_noroadmap_test.go:TestTD_UnwritableStateDirectoryNeverBreaksHostSession`
  (skipped under a root euid, where permission bits are not enforced).
- Corrupt / unreadable digest-state file fails open (renders, exit 0, no
  crash), alongside four degenerate stdin payload shapes (empty, non-JSON,
  JSON without `session_id`) against a *current* digest state.
  `tests/acceptance/token_diet_hook_failopen_test.go:TestTD_EveryUncertaintyFailsOpenAndRenders`.
- Missing or invalid `.workflow/roadmap.json`: no roadmap line printed, and
  critically no digest-state file is written either (nothing to compare a
  next prompt against once a real roadmap reappears).
  `tests/acceptance/token_diet_hook_noroadmap_test.go:TestTD_MissingOrInvalidRoadmapPrintsNothingWritesNoState`.
- Roadmap reformatting (whitespace / key-order churn, identical projection)
  must NOT force a re-render — the digest hashes the projection, not the raw
  bytes. `tests/acceptance/token_diet_hook_render2_test.go:TestTD_RoadmapReformattingAloneDoesNotForceReRender`.
- Two independent `git worktree` checkouts of the same repo keep independent
  digest state (`SummaryStatePath()` is worktree-relative, not shared via
  `.git`). `tests/acceptance/token_diet_hook_git_test.go:TestTD_TwoWorktreesKeepIndependentDigestState`.
- The digest-state file never dirties `git status --porcelain`, verified
  against the repo's REAL `.gitignore` (copied into the fixture verbatim, not
  a hand-rolled pattern) so the assertion proves the actual edited file, not
  just that gitignoring works in general.
  `tests/acceptance/token_diet_hook_git_test.go:TestTD_DigestStateFileNeverDirtiesWorkingTree`.
- A runner with no built-in model entry (`codex`) falls back to the tier
  name and never leaks another runner's concrete id.
  `tests/acceptance/token_diet_directive_test.go:TestTD_RunnerWithNoBuiltinEntryFallsBackToTierName`.
- An operator's retired dated pin (`claude-haiku-4-5-20251001`) keeps
  resolving to the same capability class and default enforcement profile as
  its replacement alias (`haiku`) — the capability map is a strict superset,
  never destructively rewritten.
  `tests/acceptance/token_diet_capability_test.go:TestTD_OperatorRetiredPinKeepsDefaultEnforcementProfile`,
  `tests/acceptance/token_diet_directive_test.go:TestTD_TierRemapInConfigOverridesBuiltinAlias`.
- Degenerate UX edge-case entries (`":"`, `": text"`, `""`) normalize to the
  empty tag and satisfy nothing, without panicking.
  `tests/acceptance/token_diet_ux_normalize_test.go:TestTD_DegenerateEntriesMatchNothingAndNeverPanic`.
- Path normalization is symmetric for both required plan-doc anchors
  (`docs/features/`, `docs/plans/`) across bare / `./` / backslash / absolute
  forms. `tests/acceptance/token_diet_plan_normalize_test.go:TestTD_PathNormalizationSymmetricForBriefsAndPlans`.
- A retired legacy plan role (`big-thinker`, `feature-specialist`) on an
  UNPINNED in-flight workflow still gets the shrunken 2-entry prefill, while
  the same roles are refused outright on a `planner-v1`-pinned workflow
  (pre-existing behavior, unaffected by this feature — see
  `TestUPS_EvidenceInitRefusesRetiredRoleOnPinnedWorkflow`).
  `tests/acceptance/token_diet_plan_init_test.go:TestTD_RetiredLegacyPlanRoleGetsSameShrunkenSet`.

## Residual Risks

- The "empty inputs list is rejected naming both paths" Gherkin scenario
  describes `validatePlanSnapshotInputs`'s own contract (verified directly,
  colocated). At the full `ValidateEvidence`/CLI level, `len(Inputs)==0`
  always trips the generic "incomplete evidence fields" branch first
  (`internal/orchestration/evidence.go:37`), so the specific per-path message
  is never the one an operator actually sees for a *wholly* empty list. This
  ordering predates token-diet and is not a regression it introduces;
  documented here rather than deferred as a finding.
- `centinela evidence validate <feature>` has no flag to supply `uiPaths`, so
  it can never fully pass a ux-ui-specialist evidence file end-to-end; only
  `centinela validate`'s internal call path (which threads configured
  `ui_paths`) can. Pre-existing CLI-surface gap, unrelated to this feature's
  UX-tag normalization change; not deferred as a token-diet finding.
- Slice C's per-prompt saving is small (~56 bytes) by design (per
  `docs/plans/token-diet.md` §1); the digest-state seam is the reusable part.
  No test asserts the byte count — that would pin an incidental UI string.
