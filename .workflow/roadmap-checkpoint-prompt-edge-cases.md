# Edge-Case Report: roadmap-checkpoint-prompt

**Date:** 2026-05-23
**Role:** qa-senior (edge-case hard-path coverage)

This report enumerates the hard paths the test suite now guards. Each entry
names the risk, the asserting test(s), and the tier.

## 1. Stale-vs-fresh boundary (`latest.After(at)`)

`Decide` suppresses when the latest artifact mtime is **at or before** the
marker `at`, and re-fires (`DecisionStale`) when an artifact is **strictly
after** it.

- Marker `at` **equal** to the latest artifact mtime -> Suppressed
  (`TestDecide_FreshMarker_Suppressed`, unit). This pins the boundary: equality
  is fresh, not stale.
- Marker `at` **after** all artifacts -> Suppressed
  (`TestDecide_FreshMarkerAfterArtifacts_Suppressed`, unit).
- An artifact one hour **after** the marker -> Stale
  (`TestDecide_StaleMarker_RoadmapNewer`, `TestDecide_StaleMarker_AnalysisNewer`,
  unit; `TestCheckpoint_Stale_RoadmapNewer`, `TestCheckpoint_Stale_AnalysisArtifactNewer`,
  integration; `TestAccept_Checkpoint_StaleRoadmap`,
  `TestAccept_Checkpoint_StaleAnalysisArtifact`, acceptance — driven via real
  `os.Chtimes` so the on-disk mtime is genuinely later).

## 2. Malformed / unparseable marker (no crash, fail toward re-emit)

The marker is untrusted input. Two corruption classes are covered:

- **Malformed JSON** (`{not valid json`): `parseMarkerAt` fails to unmarshal,
  `Decide` returns `DecisionStale` and re-emits. Asserted unit
  (`TestDecide_MalformedMarkerJSON_Stale`), integration
  (`TestCheckpoint_MalformedMarker_ReEmits`), acceptance
  (`TestAccept_Checkpoint_MalformedMarkerReEmits`). The integration/acceptance
  variants additionally assert the command does NOT crash (the test helper
  fails on any non-zero exit / error return).
- **Unparseable `at`** (`"yesterday"`): JSON parses but `time.Parse(RFC3339)`
  fails -> `DecisionStale` -> re-emit. Asserted unit
  (`TestDecide_UnparseableAt_Stale`), integration
  (`TestCheckpoint_UnparseableAt_ReEmits`), acceptance
  (`TestAccept_Checkpoint_UnparseableAtReEmits`).

Both classes fail toward re-emitting rather than silently swallowing the
prompt — a corrupted marker can never permanently suppress the checkpoint.

## 3. Precedence ordering (earlier directives win)

The checkpoint is the LAST branch in `runHookSetup`. Earlier missing-artifact
directives must short-circuit before the checkpoint is ever evaluated.

- Missing `ROADMAP.md` -> `roadmap required`, NOT checkpoint
  (`TestCheckpoint_Precedence_MissingRoadmap`, integration;
  `TestAccept_Checkpoint_PrecedenceMissingRoadmap`, acceptance).
- Invalid `.workflow/roadmap.json` -> `roadmap json` directive, NOT checkpoint
  (`TestCheckpoint_Precedence_InvalidRoadmapJSON`, integration;
  `TestAccept_Checkpoint_PrecedenceInvalidRoadmapJSON`, acceptance).

Both assert the checkpoint directive is ABSENT, so a future reordering of the
hook chain that lets the checkpoint leak through is caught.

## 4. mtime-granularity no-op re-fire trade-off (KNOWN, documented)

`WriteMarker` serializes `at` at **RFC3339 second precision**
(`now.UTC().Format(time.RFC3339)`), while filesystem mtimes carry **sub-second**
precision. If the marker is written in the SAME wall-clock second as a roadmap
artifact, that artifact's sub-second mtime is `After()` the second-truncated
`at`, and the prompt re-fires even though nothing meaningfully changed.

This is a deliberate trade-off, not a bug: in real use the user reviews the
roadmap and runs `centinela roadmap iterate` seconds-to-minutes after the
artifacts were last touched, so `at >= latest mtime` holds and the prompt stays
suppressed. The anti-spam tests reproduce the real-world ordering deterministically
by backdating artifact mtimes to a whole second strictly before the marker
write (`TestCheckpoint_AntiSpam_IterateThenSilent`, integration;
`TestAccept_Checkpoint_AntiSpamIterateThenSilent`, acceptance).

**Surfaced for validation-specialist:** if exact-second-collision suppression is
ever required, `WriteMarker` would need sub-second precision (e.g.
`time.RFC3339Nano`) OR `Decide` would need to compare at second granularity.
No production code was changed to paper over this; it is reported as a coverage
boundary.

## 5. Anti-spam idempotency

After the user persists the "iterate" choice, an unchanged disk must NOT re-fire
the prompt on the next hook invocation. Verified end-to-end through the REAL
`centinela roadmap iterate` subcommand followed by a second `hook setup`:

- Integration: `TestCheckpoint_AntiSpam_IterateThenSilent` (calls
  `runRoadmapIterate` then `runHookSetup`).
- Acceptance: `TestAccept_Checkpoint_AntiSpamIterateThenSilent` (execs the built
  binary: `roadmap iterate` then `hook setup`).

## 6. First-feature selection edge cases

`FirstIncompleteBootstrap` resolves which feature the panel names.

- `nil` roadmap -> `("", false)` (`TestFirstIncompleteBootstrap_NilRoadmap_False`).
- No Phase 0 bootstrap phase -> `("", false)`
  (`TestFirstIncompleteBootstrap_NoBootstrapPhase_False`; suppresses end-to-end
  in `TestCheckpoint_Suppressed_NoPhaseZero` / `TestAccept_Checkpoint_SuppressNoPhaseZero`).
- All bootstrap features `done` -> `("", false)`
  (`TestFirstIncompleteBootstrap_AllDone_False`; end-to-end suppression in
  `TestCheckpoint_Suppressed_BootstrapComplete`).
- First done, second incomplete -> picks the second
  (`TestFirstIncompleteBootstrap_SkipsDonePicksSecond`; end-to-end in
  `TestCheckpoint_MultiFeature_PicksSecond` / `TestAccept_Checkpoint_MultiFeaturePicksSecond`).
- `in-progress` (workflow file present, not `done`) counts as incomplete
  (`TestFirstIncompleteBootstrap_InProgressIsNonDone`). Note the interaction:
  the feature IS the first incomplete, but `Decide` then suppresses because the
  workflow file exists (`TestCheckpoint_Suppressed_WorkflowFileExists`).

## 7. Degenerate "no artifacts on disk" guard

A valid, parseable marker but with NONE of the required artifacts present:
`LatestMtime` returns `found=false` and `Decide` suppresses rather than emitting
a meaningless decision (`TestDecide_StaleMarkerButNoArtifacts_Suppressed`, unit).
This only reachable in practice via a corrupted checkout; the suite pins the
fail-safe-to-suppress behavior.
