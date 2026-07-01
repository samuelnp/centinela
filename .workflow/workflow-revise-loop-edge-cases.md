# Edge Cases: workflow-revise-loop

## Covered

- **Empty / whitespace-only `--reason`** — rejected at two layers: the cmd
  (`runRevise` trims and errors) and the domain (`RewindTo` trims and errors).
  `TestRunReviseEmptyReasonRejected`, `TestRewindToRejections`,
  acceptance `TestRL_WhitespaceReason` / `TestRL_MissingReason`.
- **Forward target** (target index >= current) — rejected, state untouched.
  `TestRewindToRejections` (forward-target), `TestRL_ForwardTarget`.
- **Equal target** (target == current step) — rejected (not strictly before).
  `TestRewindToRejections` (equal-target), `TestRL_EqualTarget`.
- **Unknown step name** — rejected, error names the offending value.
  `TestRewindToRejections` (unknown-step), `TestRL_UnknownStep`.
- **Revising a `done` workflow** — rejected before any mutation.
  `TestRewindToDoneRejected`, `TestRL_DoneWorkflow`.
- **No state mutation on any rejection** — every negative reloads persisted
  state and asserts `currentStep` unchanged; the domain test asserts
  `len(Revisions)` unchanged.
- **Archetype-awareness** — `reopenedSteps` / `RewindTo` use the workflow's own
  `OrderedSteps()`, never `DefaultStepOrder`; pinned for hotfix order
  `[code,tests,validate]` (`TestReopenedStepsCanonicalAndHotfix`,
  `TestRL_ArchetypeHotfixOrder`) — no plan/docs steps appear.
- **Target is the last step** — `reopenedSteps` returns an empty slice.
  `TestReopenedStepsCanonicalAndHotfix`.
- **Re-opened step `CompletedAt` cleared** — target and every re-opened step get
  `CompletedAt=nil`. `TestRewindToReopensDownstream`,
  `TestRL_CompletedAtClearedOnReopen`.
- **Steps before the target keep their `done` state** — plan stays done on a
  validate→code rewind. `TestRewindToReopensDownstream`.
- **Idempotent invalidation** — removing already-absent evidence is not an error
  and does not count. `TestInvalidateRemovesBothAndIdempotent`,
  `TestInvalidateArtifactIdempotent`, `TestRL_IdempotentInvalidation`.
- **Safety: never deletes source / test / docs** — only `.workflow/<feature>-*`
  is touched; sibling source, test, and docs files survive.
  `TestInvalidateSafetyNeverTouchesSource`, `TestRL_SafetyNoSourceDeletion`,
  integration `TestReviseLoopEndToEnd`.
- **Real (non-absence) removal error surfaces** — a non-empty directory at an
  evidence path makes `os.Remove` fail and the error propagates (named path).
  `TestRemoveBothSurfacesRealError`, `TestInvalidateDownstreamErrorSurfaces`.
- **Per-step invalidation policy** — validate adds gatekeeper +
  production-readiness; tests adds the `-edge-cases.md` artifact; internal code
  excludes ux-ui-specialist while user-facing code includes it.
  `TestInvalidationTargets*`, `TestRL_InternalFeatureNoUXInvalidation`.
- **Dedup across re-opened steps** — a role/artifact named by multiple steps is
  invalidated and counted once. `TestInvalidateDownstreamCountsAndDedups`.
- **Multiple rewinds accumulate** — the append-only `Revisions` log grows;
  earlier entries are preserved. `TestRL_MultipleRewindsAccumulate`,
  `TestRevisionsSummary`.
- **Back-compat JSON** — empty `Revisions` is omitted from serialized state and
  round-trips intact when present. `TestRevisionsRoundTrip`.
- **Re-gating** — after invalidation the orchestration gate reports the
  re-opened step's evidence missing (blocks), and passes once a valid evidence
  pair is regenerated. `TestRL_ReGatingBlocksWithoutEvidence`,
  `TestRL_ReGatingAdvancesAfterRegenerated`.
- **Telemetry** — a `step-revised` event records both endpoints (from/to) and
  the model. `TestRecordRevisedEvent`.
- **Status surface** — the `Revisions` row appears only when the log is
  non-empty, showing count + latest reason. `TestRenderStatusRevisionsRow`,
  `TestRenderStatusNoRevisionsRow`.

## Residual Risks

- **Spec vs. implementation note (documented, not a defect):** the spec line
  "`my-feature-senior-engineer.json` is invalidated" under the *Internal feature
  code-step* scenario is inaccurate for a `--to code` rewind — `code` is the
  **target**, so its evidence is intentionally preserved (the happy-path
  scenario relies on this invariant). The acceptance test asserts the
  genuinely-true behavior (current=code, ux-ui never referenced, the re-opened
  tests step's qa-senior evidence shed); the ux-ui exclusion for the code step
  itself is pinned at the unit level (`TestInvalidationTargetsCodeInternalVsUserFacing`).
- **Re-gating asserted at the gate primitive** (`orchestration.ValidateStep`)
  rather than by driving a full `centinela complete`, because `complete`'s
  forward path is unchanged and already covered; keeps the assertion crisp and
  avoids re-testing unrelated gate plumbing.
- **Concurrent revises** of the same feature are out of scope (single-writer CLI
  model); last-write-wins on the JSON state, consistent with the rest of the
  workflow tooling.
