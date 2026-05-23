# Edge Cases — session-context-rehydration (tests step)

Hard paths guarded by the unit / integration / acceptance suites.

## Active-workflows panel classification (Half A)
- **Evidence-JSON rejection via `feature == basename`.** `<feature>-<role>.json`
  evidence files (e.g. `alpha-qa-senior.json`) parse a `feature` that differs
  from the file basename, so `ActiveWorkflows` skips them. A role name like
  `qa-senior` never surfaces as an active row.
  Covered: unit `TestActiveWorkflows_RejectsNoiseKeepsRealNonDone`,
  integration `TestRunHookContext_EvidenceJSONNotActive`,
  acceptance `TestAcceptance_EvidenceJSONNotActive`.
- **Ad-hoc roadmap JSONs rejected.** `roadmap.json` / `roadmap-quality.json`
  have no matching `feature` field, so they are never treated as workflows.
  Covered: unit reject-noise test, acceptance `TestAcceptance_AdHocRoadmapJSONsNotActive`.
- **Done workflows excluded.** `currentStep == "done"` is filtered out while a
  genuine non-done sibling still renders.
  Covered: unit reject-noise test, acceptance `TestAcceptance_DoneExcludedNonDoneShown`.
- **Dedupe to a single row.** Multiple evidence JSONs for one feature plus the
  genuine state file collapse to exactly one panel row.
  Covered: unit `TestActiveWorkflows_DedupesToSingleRow`,
  acceptance `TestAcceptance_DuplicatesDedupeToSingleRow` (scoped to the
  ACTIVE WORKFLOWS panel — per-feature review reminders legitimately repeat names).
- **Mtime recency ordering.** Survivors sort by file modification time
  descending (most-recently-touched first).
  Covered: unit `TestActiveWorkflows_SortsByMtimeDescending`.
- **Cap + `+N more` hint.** Above the cap of 5, only the 5 most-recent rows show
  and a `+N more` hint appears; at-or-below the cap there is no hint; `max <= 0`
  means no cap.
  Covered: unit `TestCapActive_*` and `TestRenderContextCapped_MoreHintBranches`,
  integration `TestRunHookContext_CapShowsPlusNMore`,
  acceptance `TestAcceptance_CapShowsRecentPlusNMore` / `TestAcceptance_AtOrBelowCapNoMoreHint`.

## SessionStart rehydration (Half B)
- **Cross-phase FirstIncomplete.** When every Phase 0 feature is done, the walk
  continues into later phases; the next feature is the first incomplete across
  ALL phases, not just Phase 0.
  Covered: unit `TestFirstIncomplete_CrossesPhases`,
  integration `TestRunHookSession_NextIsFirstIncompleteAcrossPhases`,
  acceptance `TestAcceptance_NextIsFirstIncompleteAcrossPhases`.
- **Roadmap-complete empty state.** Every feature done yields a graceful
  "Roadmap complete" line with NO next-feature name and NO `docs/features/<next>.md`
  pointer; exit 0, no crash.
  Covered: unit `TestRenderSessionRehydration_CompleteHasNoNextPointer` and
  `TestFirstIncomplete_AllDoneReturnsFalse`,
  integration `TestRunHookSession_AllDoneRoadmapComplete`,
  acceptance `TestAcceptance_AllDoneRoadmapComplete`.
- **Missing / invalid roadmap silent no-crash.** Absent `.workflow/roadmap.json`
  or malformed JSON both exit 0 emitting nothing — no rehydration payload.
  Covered: unit `TestFirstIncomplete_NilAndEmpty`,
  integration `TestRunHookSession_MissingAndInvalidAreSilent`,
  acceptance `TestAcceptance_MissingAndInvalidRoadmapSilent`.
- **Pointers-as-paths-only.** The payload lists `PROJECT.md` and
  `docs/features/<next>.md` as PATHS and never inlines their file contents
  (asserted by the absence of inlined brief headings such as `## Problem`).
  Covered: unit `TestRenderSessionRehydration_SuccessPayloadHasPointersNotContents`,
  integration `TestRunHookSession_ValidRoadmapEmitsRehydration`,
  acceptance `TestAcceptance_SessionStartPayloadOnEachSource`.
- **Source-independence.** The SessionStart payload is identical across the
  `startup | clear | compact | resume` sources.
  Covered: acceptance `TestAcceptance_SessionStartPayloadOnEachSource`.

## Stable spec literals (must not drift)
`CENTINELA DIRECTIVE: session rehydration`, `+N more`, pointer paths
`PROJECT.md` and `docs/features/<next>.md`, cap = 5.
