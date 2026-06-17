# deferred-findings-roadmap-capture — qa-senior

## Summary

Complete test inventory for the `deferred-findings-roadmap-capture` feature.
Coverage gate: **95.0% >= 95.0% — PASS**.
All 1504 tests pass. Format gate clean. G1 compliant (all _test.go in internal/ and cmd/ ≤ 100 lines).

---

## Test Inventory

### Colocated Unit Tests (internal/roadmap/)

| File | Tests | Coverage target |
|------|-------|-----------------|
| backlog_test.go | isBacklogPhaseName variants, IsBacklogPhaseName, NonBacklogFeatureSet, BacklogFeatures, IsBacklogFeature | backlog.go |
| defer_test.go | HappyPath, PreservesExistingEntries, EmptySummary (regression), WhitespaceSummary | defer.go |
| defer_collisions_test.go | InvalidSlug, SlugCollisionInBacklog (regression), SlugCollisionInNonBacklog (regression), NoSourceField, AppendToExistingBacklog | defer.go + defer_validate.go |
| defer_more_test.go | Defer missing file, corrupt file, JSON special chars | defer.go |
| defer_validate_test.go | validateSlug valid/invalid, validateSummary empty/whitespace, validateNoCollision clean/collision | defer_validate.go |
| promote_test.go | LoadBacklogFinding found/not-found, Promote HappyPath | promote.go |
| promote_regression_test.go | PreflightMissingAnalysisJSON (regression), PreflightMissingQualityJSON (regression), UnknownPhase (regression), SlugNotInBacklog, EmptiedBacklogKept | promote.go + promote_preflight.go |
| promote_more_test.go | Promote missing file, LoadBacklogFinding missing file, summary override, deferredAt provenance | promote.go |
| promote_more2_test.go | corrupt backlog entry, LoadBacklogFinding corrupt JSON, double-promote refused, summary fallback | promote.go |
| promote_preflight_test.go | preflightArtifacts all present, missing analysis/quality JSON and MD, checkArtifactJSON corrupt/invalid | promote_preflight.go |
| promote_artifacts_test.go | provenanceBullet with/without source, appendFeatureEntry append/missing/corrupt | promote_artifacts.go |
| promote_artifacts_more_test.go | appendPromotionArtifacts writes all four files, empty features key | promote_artifacts.go |
| promote_artifacts_errors_test.go | missing quality JSON, missing analysis.md (platform-dependent) | promote_artifacts.go |
| promote_write_error_test.go | Promote write error (roadmap.json is directory), Defer write error | promote.go + defer.go |
| promote_json_error_test.go | LoadBacklogFinding corrupt entry, appendToPhase corrupt phase | promote.go |
| rawio_test.go | readRawRoadmap happy/missing/corrupt/invalidPhases, writeAtomic, compactBytes, writeRawRoadmap roundtrip | rawio.go |
| rawio_more_test.go | writeAtomic creates parent dir, compactBytes no HTML escape, readRawRoadmap no-phases-key | rawio.go |
| rawio_errors_test.go | writeAtomic success, writeRawRoadmap render error | rawio.go |
| rawmutate_test.go | phaseFeatureNames, appendBacklog new/existing phase, featureName, encodePhase | rawmutate.go |
| rawmutate_more_test.go | decodePhase invalid JSON, setPhase dirty, featureName invalid, phaseFeatureNames corrupt | rawmutate.go |
| rawmutate_errors_test.go | appendBacklog decode error, phaseFeatureNames empty, decodePhase valid, setPhase valid | rawmutate.go |
| rawrender_test.go | render untouched/sorted-keys, indentValue, backlogPhaseIndex present/absent | rawrender.go |
| rawrender_more_test.go | phaseBytes dirty path, render with dirty phase, backlogPhaseIndex lowercase | rawrender.go |
| rawphase_render_test.go | renderDirtyPhase one-per-line/empty/invalid, writePhaseKey | rawphase_render.go |
| rawmove_test.go | findInBacklog found/not-found/no-phase, removeBacklogFeature | rawmove.go |
| rawmove_append_test.go | appendToPhase happy/unknown/Backlog/duplicate | rawmove.go |
| rawmove_more_test.go | knownPhaseList, removeNoMatch idempotent, phaseName invalid | rawmove.go |
| rawmove_errors_test.go | findInBacklog corrupt phase, appendToPhase corrupt phase | rawmove.go |
| artifactio_test.go | writeArtifact sorted keys, byte-stable, writeFeatureArray multiple entries, appendLine creates/appends | artifactio.go |
| artifactio_more_test.go | writeArtifact invalid features array, appendLine no trailing newline | artifactio.go |
| mapkeys_test.go | sortedKeys ascending, empty map | mapkeys.go |
| promote_scores_test.go | ParseScores valid/boundary/overall-threshold/out-of-range/wrong-count/non-numeric/empty | promote_scores.go |

### Colocated Unit Tests (internal/ui/)

| File | Tests | Coverage target |
|------|-------|-----------------|
| render_backlog_test.go | renderBacklogSection with findings, no Backlog phase, empty Backlog | render_backlog.go |
| render_promote_test.go | RenderPromoteEvaluatorContext with source, no source, all 6 dimensions | render_promote.go |

### Colocated Unit Tests (cmd/centinela/)

| File | Tests | Coverage target |
|------|-------|-----------------|
| roadmap_defer_test.go | resolveDeferSource explicit/feature-only/empty, runRoadmapDefer happy/empty-summary | roadmap_defer.go |
| roadmap_defer_more_test.go | resolveDeferSource worktree CWD auto-detection, repo root nil | roadmap_defer.go |
| roadmap_promote_test.go | runRoadmapPromote no-scores/explicit-empty-scores(regression)/no-phase/low-score | roadmap_promote.go |
| roadmap_promote_more_test.go | runRoadmapPromote scored success, reportPromoteResult success, printEvaluatorContext not-in-backlog | roadmap_promote.go |
| roadmap_promote_errors_test.go | reportPromoteResult no roadmap, validate fails | roadmap_promote.go |
| roadmap_promote_final_test.go | promoteScored Promote() error, reportPromoteResult quality validate fails | roadmap_promote.go |
| start_guard_backlog_test.go | workflowOrderForFeature Backlog refused with "promote" error (regression) | start_guard.go |
| start_guard_dep_block_test.go | workflowOrderForFeature dep-blocked error path | start_guard.go |

---

## Regression Tests

Four defects fixed in the code pass under dedicated regression tests:

| Defect | Regression test |
|--------|-----------------|
| Partial-write preflight: roadmap.json written before artifact pre-check | TestPromote_PreflightMissingAnalysisJSON, TestPromote_PreflightMissingQualityJSON |
| Duplicate-entry refusal: same slug could be deferred twice | TestDefer_SlugCollisionInBacklog, TestDefer_SlugCollisionInNonBacklog |
| Sorted key order: writeArtifact must emit keys in ascending order | TestWriteArtifact_SortedKeys |
| --scores "" treated as no-scores path instead of error | TestRunRoadmapPromote_ExplicitEmptyScoresIsError, TestDfrc_PromoteEmptyScoresError |

---

## Integration Tests

**tests/integration/deferred_findings_roadmap_capture_integration_test.go**

`TestDeferThenPromote_FullFlow` — end-to-end Go-API test covering:
1. Defer a finding -> verify Backlog entry with slug, summary, source, deferredAt
2. Promote with passing scores -> verify moved to Phase 5, Backlog retained, artifacts updated

---

## Acceptance Tests (specs/deferred-findings-roadmap-capture.feature — 25 scenarios)

| Spec scenario | Acceptance test function | File |
|--------------|--------------------------|------|
| Happy-path defer | TestDfrc_DeferHappyPath | deferred_findings_defer_test.go |
| Defer appends to existing Backlog | TestDfrc_DeferAppendsToExistingBacklog | deferred_findings_defer_test.go |
| Defer no --source from repo root | TestDfrc_DeferNoSourceField | deferred_findings_defer_test.go |
| Defer empty summary rejected | TestDfrc_DeferEmptySummaryRejected | deferred_findings_defer_rejections_test.go |
| Defer duplicate Backlog slug | TestDfrc_DeferDuplicateBacklogSlug | deferred_findings_defer_rejections_test.go |
| Defer duplicate non-Backlog slug | TestDfrc_DeferDuplicateNonBacklogSlug | deferred_findings_defer_rejections_test.go |
| Defer invalid slug | TestDfrc_DeferInvalidSlug | deferred_findings_defer_rejections_test.go |
| Backlog shown in roadmap output | TestDfrc_BacklogShownInRoadmapOutput | deferred_findings_render_test.go |
| No Backlog section when missing | TestDfrc_NoBacklogSectionWhenMissing | deferred_findings_render_test.go |
| No Backlog section when empty | TestDfrc_NoBacklogSectionWhenEmpty | deferred_findings_render_test.go |
| Backlog not in ready output | TestDfrc_BacklogNotInReadyOutput | deferred_findings_render_test.go |
| start refuses Backlog feature | TestDfrc_StartRefusesBacklogFeature | deferred_findings_render_test.go |
| validate passes Backlog exempt | TestDfrc_ValidatePassesBacklogExempt | deferred_findings_validate_exempt_test.go |
| validate fails uncovered non-Backlog | TestDfrc_ValidateFailsUncoveredNonBacklog | deferred_findings_validate_exempt_test.go |
| Pre-Backlog phase not exempt | TestDfrc_PreBacklogPhaseNotExempt | deferred_findings_validate_exempt_test.go |
| Promote no --scores prints context | TestDfrc_PromoteNoScoresPrintsContext | deferred_findings_promote_test.go |
| Promote with scores moves entry | TestDfrc_PromoteWithScoresMovesEntry | deferred_findings_promote_test.go |
| Promote preserves unknown fields | TestDfrc_PromotePreservesUnknownFields | deferred_findings_promote_test.go |
| Promote low overall rejected | TestDfrc_PromoteLowOverallRejected | deferred_findings_promote_rejections_test.go |
| Promote out-of-range score rejected | TestDfrc_PromoteOutOfRangeScoreRejected | deferred_findings_promote_rejections_test.go |
| Promote unknown phase rejected | TestDfrc_PromoteUnknownPhaseRejected | deferred_findings_promote_rejections_test.go |
| Promote slug not in Backlog | TestDfrc_PromoteSlugNotInBacklog | deferred_findings_promote_rejections_test.go |
| Promote malformed --scores rejected | TestDfrc_PromoteMalformedScoresRejected | deferred_findings_promote_rejections_test.go |
| --scores "" is usage error (regression) | TestDfrc_PromoteEmptyScoresError | deferred_findings_promote_rejections_test.go |
| Prompt parity byte-identical | TestDfrc_PromptParityByteIdentical | deferred_findings_prompt_parity_test.go |

Note: "Defer auto-resolves --source from worktree CWD" is tested at the unit level
(TestResolveDeferSource_WorktreeCWD in roadmap_defer_more_test.go). The acceptance-tier
binary test is omitted because the test harness cannot place the binary inside a
.worktrees/<feature>/ CWD without filesystem workarounds.

---

## Gate Status

| Check | Result |
|-------|--------|
| `rtk go test ./...` | 1504 passed |
| `./scripts/check-coverage.sh` | 95.0% >= 95.0% PASS |
| `./scripts/check-fmt.sh` | PASS |
| G1 all _test.go <= 100 lines | PASS |
| Three test tiers present | PASS (unit + integration + acceptance) |
| Regression tests for 4 defects | PASS |
| 25 acceptance scenarios traced | PASS |
