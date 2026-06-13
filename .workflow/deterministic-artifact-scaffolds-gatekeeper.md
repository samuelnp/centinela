### Gatekeeper Report: deterministic-artifact-scaffolds
**Date:** 2026-06-12
**Status:** WARNING

#### Analyzed Specs
- specs/deterministic-artifact-scaffolds.feature (the new feature spec under review)
- specs/enforce-plan-snapshot-inputs.feature (plan-snapshot input contract — primary regression surface)
- specs/evidence-cli.feature (typed evidence CLI: init / set / append / artifact new behavior)
- specs/enforce-actionable-orchestration-evidence.feature (actionable-outputs validator — FILL-marker leak surface)

Domain files inspected:
- internal/orchestration/plan_snapshot.go (RequiredPlanInputs + validatePlanSnapshotInputs)
- internal/evidence/plan_inputs.go (PlanInputs pre-fill, delegates to RequiredPlanInputs)
- internal/evidence/companion.go, companion_skeletons.go, fill.go (FILL markers, per-role skeletons)
- internal/evidence/artifact_derive.go, artifact_gatekeeper.go, artifact_templates.go (Analyzed Specs pre-fill)
- docs/architecture/evidence-contract.md (documented per-role rules)

#### Findings

**Item A — Extending `RequiredPlanInputs` to also require `docs/plans/<feature>.md`: NO REGRESSION (doc-aligned tightening). VERDICT: SAFE.**

- The validator (`validatePlanSnapshotInputs`) computes its required set from the
  exact same `RequiredPlanInputs(feature)` that the init pre-fill (`PlanInputs`)
  uses, so a pre-filled init validates by construction (Slice-1 spec scenario
  "Init pre-fill lets big-thinker pass plan-snapshot validation with zero appends").
- The tightening already matched the locked contract: docs/architecture/evidence-contract.md
  (big-thinker / feature-specialist per-role rules, lines 56-72) already documents
  `inputs` MUST include "the current feature's plan at `docs/plans/<feature>.md`".
  The code was previously LOOSER than the documented contract; this change closes
  that gap. It is a correctness fix, not a new requirement invented by this feature.
- enforce-plan-snapshot-inputs.feature asserts two behaviors: (1) omitting any
  required path FAILS, (2) including the full snapshot set PASSES. The stricter set
  is a superset; its "passes with full snapshot coverage" scenario uses evidence
  that follows the documented contract (which includes the plan path), so it still
  passes. The "fails without full coverage" scenario is strengthened, not broken.
- Independently confirmed via internal/orchestration/evidence_snapshot_test.go:
  evidence listing only `docs/features/f.md` (omitting plan + sibling brief) now
  correctly errors "missing feature-doc snapshot inputs"; evidence that adds
  `docs/plans/f.md` + sibling brief PASSES. required_plan_inputs_test.go proves the
  set is sorted, de-duplicated, and includes the plan path. Full suite green
  (1506 passing); `centinela validate` passes all gates; spec-traceability-gate
  reports all 21 scenarios covered.
- Note: existing on-disk evidence for already-completed features is not re-validated
  retroactively; the rule fires only on plan-step big-thinker / feature-specialist
  evidence at validate/complete time. No silent breakage of unrelated features.

**Item B — Empty Analyzed Specs renders one `<FILL:>` row vs. spec wording "no `<FILL:` placeholder rows": COSMETIC SPEC-PROSE MISMATCH, not a functional conflict. VERDICT: WARNING (acceptable, but spec prose should be corrected).**

- The spec scenario (specs/deterministic-artifact-scaffolds.feature line 125) reads
  "an empty list with no `<FILL:` placeholder rows". The implementation
  (`analyzedSpecsList()` in artifact_derive.go) and the acceptance test
  (TestDAS_GatekeeperEmptyAnalyzedSpecs) instead emit exactly ONE `<FILL:>` row and
  the test asserts `<FILL:` IS present. So the test does NOT match the literal spec prose.
- This is markdown-only with zero JSON-validator exposure: `analyzedSpecsList()`
  feeds only `gatekeeperBody()` (a markdown stub). The FILL marker can never reach an
  evidence JSON list field — guarded by the "No FILL marker ever lands in an evidence
  JSON list field" scenario and the actionable-outputs validator in
  enforce-actionable-orchestration-evidence.feature. Confirmed green.
- The chosen behavior is the better engineering choice (a single truthful FILL slot
  prompts the human gatekeeper to confirm "no specs", rather than a phantom spec path
  or a silently-empty section). The actual defect is the SPEC TEXT, which over-specifies
  "no `<FILL:` rows" when the intended/implemented contract is "no phantom `specs/` paths".
- Why WARNING not SAFE: the spec-traceability-gate only checks that a scenario maps to
  some test, not that the test enforces the prose; so this drift passes the gate silently.
  Recommend a one-line spec edit (line 125: "...empty of real spec paths; a single
  `<FILL:>` prompt row is allowed") to realign the locked spec with the truthful behavior.
  No code change required.

#### Recommendation
- WARNING — no regression and no validator exposure; ship after a one-line spec-prose
  fix on line 125 so the locked `.feature` text matches the truthful one-FILL-row behavior
  the acceptance test already enforces.
