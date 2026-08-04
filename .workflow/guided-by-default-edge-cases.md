# Edge Cases: guided-by-default

## Covered

- **Missing gatekeeper report blocks both profiles identically** —
  `tests/acceptance/enforcement_profiles_invariant_report_test.go:TestEP_MissingGatekeeperReportBlocksEveryProfile`
- **Narrated verdict with an empty commands array (no commands-run record)
  blocks both profiles identically** —
  `tests/acceptance/enforcement_profiles_invariant_report_test.go:TestEP_UngroundedVerdictBlocksEveryProfile`
- **Stale verification (revision skew after a stamped SAFE verdict) blocks
  both profiles identically** —
  `tests/acceptance/enforcement_profiles_invariant_freshness_test.go:TestEP_StaleVerificationBlocksEveryProfile`
- **A failing `validate.commands` entry blocks both profiles identically** —
  `tests/acceptance/enforcement_profiles_invariant_gatefail_test.go:TestGBD_GateFailureBlocksBothProfiles`
  and (3-profile superset) `tests/acceptance/enforcement_profiles_invariant_test.go:TestEP_GatesRunUnderEveryProfile`
- **A BLOCKING production-readiness report blocks both profiles identically** —
  `tests/acceptance/enforcement_profiles_invariant_prodready_test.go:TestEP_BlockingProductionReadinessBlocksEveryProfile`
- **The one legitimate divergence, asserted as divergent**: an identical
  SAFE/stamped/passing tree with no orchestration-evidence bundle completes
  under guided but is refused under strict —
  `tests/acceptance/enforcement_profiles_invariant_clean_test.go:TestEP_GuidedSkipsOrchestrationEvidenceStrictRequires`
- **A clean tree (no refusal anywhere) completes under both profiles**, so
  parity is not achieved by everything failing —
  `tests/acceptance/enforcement_profiles_invariant_clean_test.go:TestEP_CleanTreeCompletesUnderEveryProfile`
- **A pre-flip (unpinned `profileContract`) workflow state file stays strict**
  even with zero config —
  `tests/acceptance/guided_by_default_default_flip_test.go:TestGBD_LegacyWorkflowKeepsStrict`
- **A strict greenfield project still demands the full cascade** (roadmap
  senior PM analysis) where a guided one would only advise —
  `tests/acceptance/guided_by_default_cascade_test.go:TestGBD_StrictGreenfieldStillRequiresFullCascade`
- **The setup hook still halts (directive + no checkpoint) on a strict
  project missing roadmap analysis**, byte-identical to pre-flip behavior —
  `tests/acceptance/guided_by_default_cascade_test.go:TestGBD_SetupHookHaltsOnStrictProject`
- **Guided still refuses a Backlog finding, a draft feature, and a roadmap
  with no `Phase 0: Bootstrap`** —
  `tests/acceptance/guided_by_default_start_refusals_test.go:TestGBD_GuidedStillRefusesWhatRoadmapRequires`
- **A missing `.workflow/roadmap.json` is refused under guided** (required in
  every profile; only the grading rungs are profile-scoped) —
  `tests/acceptance/guided_by_default_start_refusals_test.go:TestGBD_MissingRoadmapJSONRefusedUnderGuided`
- **A missing `scores` object, a `scores` array (wrong JSON kind), and a
  non-integer `overall` are still reported as shape/type faults, never
  silently passed or mis-attributed as a range fault** —
  `tests/acceptance/guided_by_default_quality_test.go:TestGBD_MissingScoresObjectIsShapeFault`,
  `TestGBD_WrongKindScoresIsShapeFault`, `TestGBD_NonIntegerScoreIsTypeFault`
- **Out-of-range scores (0 and 11) are still refused** even though the
  `overall >= 9` gate itself is gone —
  `tests/acceptance/guided_by_default_quality_promote_test.go:TestGBD_OutOfRangeScoreRefused`
- **A project that already pins its enforcement profile gets no doctor
  advisory** — the negative path for the advisory itself —
  `tests/acceptance/guided_by_default_doctor_test.go:TestGBD_DoctorSilentWhenProfilePinned`
- **This repo's own `centinela.toml` losing the explicit `strict` pin fails a
  test**, not just a review —
  `tests/acceptance/guided_by_default_self_governance_test.go:TestGBD_RepoPinsEnforcementProfileExplicitly`
- **`complete_validate_gates.go` gaining a profile branch fails a test**
  (source-level AST guard, landed by senior-engineer) —
  `cmd/centinela/complete_validate_gates_invariant_test.go`

## Residual Risks

- **Two of the five `Guided still refuses...` outline examples (unmet
  dependencies, non-bootstrap feature while bootstrap is incomplete) are only
  covered at unit level**, not acceptance/binary-driven —
  `cmd/centinela/start_guard_guided_refusals_test.go`. Accepted: both paths
  are unconditional (not profile-scoped) and read `roadmap.UnmetDependencies`
  / `roadmap.BootstrapComplete` directly, so the acceptance-tier risk is low;
  the covered scenario markers for the outline's Backlog/draft/no-bootstrap-
  phase examples above give traceability without re-deriving the full
  example matrix at the binary layer.
- **`workflow.DisplayProfile` is dead code that would mislead if adopted**
  (returns `strict` for every guided-by-default workflow because it reads
  `EnforcementProfile`, not `ProfileContract`). Already flagged by
  senior-engineer as deferred finding `profile-display-unused-and-now-
  misleading` (source `guided-by-default/senior-engineer`) — not re-raised
  here; no new test added since there is no production caller to regress.
- **The two inert `ProfileKnobs` fields** (`PlanAdvisorMode`, `StepGating`)
  remain unwired, per the brief's explicit "Out" scope — not a gap
  introduced by this feature, not re-tested here.
- **Production-readiness parity is asserted only for the BLOCKING status,
  not WARNING.** `ProductionReadinessWarning` is profile-agnostic (reads only
  `cfg.Gates.ProductionReadinessEnabled` and the report file), so a WARNING
  parity test would exercise the same code path already covered on the
  BLOCKING side; not duplicated here.
- **No deferred findings recorded**: no genuinely new gap was found beyond
  what senior-engineer already deferred. `centinela roadmap defer` was not
  invoked.
