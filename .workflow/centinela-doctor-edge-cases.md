# Edge Cases: centinela-doctor

## Covered

Each case below is exercised by a colocated unit test (`internal/doctor/*_test.go`,
`internal/ui/render_doctor_test.go`, `cmd/centinela/doctor_test.go`) and/or an
end-to-end acceptance test (`tests/acceptance/centinela_doctor_*_test.go`).

- **No `.claude/` directory** → hooks check WARN ("run `centinela setup`"), no
  crash, no repair. (`TestHooksCheckNoClaudeDirWarns`, `TestDoctorNoClaudeDirWarns`)
- **Hooks missing/stale** → ERROR with itemized details + safe re-wire repair;
  `--fix` re-wires; second `--fix` byte-identical (idempotent).
  (`TestHooksCheckMissingEntriesError`, `TestHooksRepairFixesAndIsIdempotent`,
  `TestDoctorMissingHooksErrorThenFix`, `TestDoctorHookFixIdempotent`)
- **No roadmap.json** → roadmap check OK / not-applicable. (`TestRoadmapCheckNoFileNotApplicable`)
- **Corrupt roadmap.json** → roadmap check ERROR ("cannot load"), no panic.
  (`TestRoadmapCheckCorruptLoadError`)
- **ROADMAP.md drift** → ERROR; `--fix` regenerates. (`TestRoadmapCheckDriftError`,
  `TestDoctorRoadmapDriftRepaired`)
- **Phase-name glyph** (`"✅ Phase 0: Bootstrap"`) → ERROR naming the phase +
  prefix breakage; `--fix` strips glyph & regenerates; idempotent re-run.
  (`TestRoadmapCheckGlyphError`, `TestRoadmapRepairStripsGlyphRegeneratesIdempotent`,
  `TestDoctorRoadmapGlyphStripped`, `TestDoctorRoadmapGlyphFixIdempotent`)
- **Glyph helper edges**: clean name, leading glyph, digit-leading, empty,
  whitespace-only, glyph-without-"Phase", multi-glyph. (`TestHasLeadingGlyph`,
  `TestStripLeadingGlyph`)
- **Abandoned worktree** (merged/missing branch) → ERROR, report-only with exact
  `git worktree remove` command; `--fix` never removes it.
  (`TestWorktreesCheckAbandonedReportsCommand`, `TestDoctorFixDoesNotRemoveWorktree`)
- **No worktrees** → worktrees check OK. (`TestWorktreesCheckNoneOK`, `TestDoctorNoWorktreesOK`)
- **Orphaned `.workflow` state** (no branch + no worktree) → ERROR, report-only
  `rm` command; `--fix` never deletes; a live branch or live worktree exempts it.
  (`TestWorkflowStateOrphanReported`, `TestWorkflowStateLiveBranchNotOrphan`,
  `TestOrphanedWorkflowsSkipsLiveWorktree`, `TestDoctorFixDoesNotDeleteWorkflowState`)
- **Orphaned `*.json.tmp`** → ERROR listing paths + safe sweep; `--fix` removes;
  idempotent; remove-failure propagates. (`TestEvidenceCheckOrphanErrorAndRepair`,
  `TestRepairEvidenceErrorPropagates`, `TestDoctorEvidenceFixIdempotent`)
- **Config**: `verify_timeout` below floor → WARN; missing gate dir → WARN;
  unknown TOML keys → WARN naming each key; all report-only (`--fix` no-op);
  unparseable `centinela.toml` → ERROR, other checks still run.
  (`TestConfigCheck*`, `TestUnknownConfigKeys*`, `TestDoctorConfigParseErrorDegradesToError`)
- **Version**: installed behind Makefile → WARN reporting both + `make install`;
  binary not on PATH → WARN, no crash; no comparable Makefile VERSION → OK;
  report-only. (`TestVersionCheck*`, `TestDoctorVersionBehindWarn`,
  `TestDoctorVersionBinaryNotFoundWarn`)
- **Exit codes**: any ERROR → exit 1; OK/WARN only → exit 0. (`TestExitError`,
  `TestDoctorAnyErrorExitsOne`, `TestDoctorOnlyWarnExitsZero`)
- **Deterministic, non-TTY output**: fixed check order across runs; byte-identical
  repeated runs; no ANSI escapes; every line is a glyph/detail/summary line.
  (`TestRunOrderStableAcrossRuns`, `TestDoctorDeterministicOutput`, `TestDoctorNonTTYNoANSI`)
- **`--fix` partial failure**: a failing safe repair marks that check ERROR while
  the others still apply; exit 1; per-check post-fix report is accurate.
  (`TestFixPartialFailureMarksCheckError`, `TestDoctorFixAttemptsAllEvenWhenOneFails`,
  `TestDoctorFixPartialReportPerCheck`)
- **Multiple simultaneous problems** reported in one pass; multiple safe repairs
  applied in one `--fix` invocation. (`TestDoctorMultipleProblemsSinglePass`,
  `TestDoctorFixMultipleInOnePass`, `TestDoctorFixNeverDestructive`)
- **Run from a worktree vs repo root**: `NewContext`/`resolveRoot` climbs out of
  `.worktrees/<feature>` to the canonical root; checks read root-relative state.
  (`TestResolveRoot*`, `TestNewContextChdirsToRootAndLoadsConfig`,
  `TestDoctorResolvesRepoRootFromWorktree`, `TestDoctorRunsFromRepoRoot`)
- **Not in a git repo**: git-dependent checks (worktrees, workflow-state) degrade
  to WARN/OK, no panic; non-git checks still diagnose.
  (`TestWorktreesCheckNoGitWarns`, `TestDoctorNotInGitRepoDegrades`)
- **No active workflow**: doctor runs to completion (read-only diagnostic).
  (`TestDoctorRunsWithoutActiveWorkflow`)
- **Renderer**: glyph per status, detail bullets, report-only `→ run:` line only
  when a `Command` is set (not for safe Apply-only repairs). (`TestRenderDiagnosis*`)

## Residual Risks

- **Hook BuildSyncPlan I/O failure** and **`os.Chdir`/`filepath.Abs` failures**
  in `NewContext`/`resolveRoot` are defensive error branches not unit-covered;
  they require an unreadable filesystem state that is not hermetically
  reproducible. `internal/doctor` package coverage is 97.5% with these as the
  only gaps; behavior on failure is a returned error, never a panic.
- **Real `centinela --version` parsing** is exercised via a stubbed
  `versionRunner` (unit) and a PATH-injected fake binary (acceptance) rather than
  the host's installed binary, keeping tests deterministic. The live binary path
  is the report-only WARN case and cannot mutate state.
- **OpenCode hook wiring** is treated by the hooks check identically to Claude
  wiring (one `setup.BuildSyncPlan("both")` plan); a project that intentionally
  omits OpenCode files still surfaces them as fixable — accepted as conservative
  per the feature brief (WARN/ERROR over silent drift).
