### QA-Senior Report: adversarial-validate-verifier
**Date:** 2026-07-28

#### Test Inventory

| Tier | File | Scenarios |
|------|------|-----------|
| unit (pre-existing, colocated) | `internal/gatereport/{verdict,block,check,stamp}_test.go` | Status parsing, alias normalization, fence parsing, admissibility matrix |
| unit (pre-existing, colocated) | `internal/treestate/{digest,treestate}_test.go` | digest exclusion of `.workflow/`, stamp computation (stubbed Runner) |
| unit (pre-existing, colocated) | `internal/workflow/{validate_gatekeeper,validate_freshness,validate_freshness_stale,stamp}_test.go` | legacy vs adversarial content gate, freshness matrix (stubbed Runner) |
| unit (pre-existing, colocated) | `internal/orchestration/{policy_more,policy_user_facing,models,directives}_test.go` | `RequiredRoles("validate") == [gatekeeper]`, reasoning tier, delegation contract string |
| acceptance (new) | `tests/acceptance/adversarial_validate_verifier_happy_test.go` | SAFE advances; WARNING advances + ledger entry |
| acceptance (new) | `tests/acceptance/adversarial_validate_verifier_verdict_test.go` | CRITICAL blocks w/ echoed finding; BLOCKING/UNSAFE aliases; missing/unparseable Status |
| acceptance (new) | `tests/acceptance/adversarial_validate_verifier_grounding_test.go` | empty commands; artifact-new-stub-then-complete; partial commands (no passing validate); no-bash-harness fail-closed |
| acceptance (new) | `tests/acceptance/adversarial_validate_verifier_freshness_test.go` | revision skew; uncommitted patch; `.workflow/`-only churn advances (D3a); artifact stamp preserves commands |
| acceptance (new) | `tests/acceptance/adversarial_validate_verifier_legacy_test.go` | legacy in-flight completes; fresh workflow refuses hand-authored legacy; `RequiredRolesForFeature` exactly `[gatekeeper]` |
| acceptance (new) | `tests/acceptance/adversarial_validate_verifier_directive_test.go` | hook directive names gatekeeper + reasoning tier + paths-only contract (AC6) |
| acceptance (new) | `tests/acceptance/adversarial_verifier_prompt_test.go` | refutation stance, input contract, no-summaries prohibition, CRITICAL, verification fence, stamp instruction, fail-closed clause |
| acceptance (new, helpers) | `tests/acceptance/adversarial_validate_verifier_{helper,report}_test.go` | shared binary build (sync.Once), fixture/report builders |
| integration (new) | `tests/integration/adversarial_validate_verifier_integration_test.go` (+`_helper_test.go`) | `workflow.ValidateArtifacts` + `workflow.VerificationFresh` driven through a REAL git repo and the REAL exec Runner (colocated tests only stub the Runner) |

Spec scenario → test function map (all 18 scenarios in `specs/adversarial-validate-verifier.feature` execute):

1. Grounded SAFE advances → `TestAVV_SafeVerdictAdvances`
2. WARNING advances + ledger → `TestAVV_WarningVerdictAdvancesAndLandsInLedger`
3. CRITICAL blocks, finding echoed → `TestAVV_CriticalVerdictBlocksWithFindingEchoed`
4. Legacy alias outline (BLOCKING/UNSAFE) → `TestAVV_LegacyAliasesBlockAsCritical`
5. Missing Status blocks → `TestAVV_MissingStatusLineBlocks`
6. Unparseable Status blocks → `TestAVV_UnparseableStatusBlocks`
7. Empty commands-run record fails → `TestAVV_EmptyCommandsRecordBlocks`
8. `artifact new` stub then `complete` FAILS → `TestAVV_ArtifactNewStubThenCompleteFails`
9. Commands without a passing `centinela validate` refused → `TestAVV_PartialCommandsWithoutPassingValidateRefused`
10. No-bash-harness fails closed → `TestAVV_NoBashHarnessFailsClosed`
11. Revision skew demands fresh verification → `TestAVV_RevisionSkewDemandsFreshVerification`
12. Uncommitted patch stales the digest → `TestAVV_UncommittedPatchStalesTheDigest`
13. `.workflow/`-only churn survives freshness (D3a) → `TestAVV_WorkflowOnlyChurnSurvivesFreshness` + `TestRealGitFreshnessRoundTrip` (real git)
14. `artifact stamp` records revision/digest, preserves commands → `TestAVV_ArtifactStampRecordsRevisionAndPreservesCommands`
15. Legacy in-flight workflow still completes → `TestAVV_LegacyInFlightWorkflowStillCompletes`
16. Fresh workflow refuses hand-authored legacy → `TestAVV_FreshWorkflowRefusesHandAuthoredLegacyFormat`
17. No validation-specialist required + `RequiredRolesForFeature` → `TestAVV_NoValidationSpecialistRequired`
18. Hook directive names verifier/tier/contract → `TestAVV_HookDirectiveNamesVerifierAndReasoningTier`

#### Coverage Gaps

- **Spec/implementation gap** (not a missing test — a missing behavior):
  the WARNING scenario's second `Then` clause, "the WARNING finding should
  be surfaced in the complete output," is not implemented by
  `cmd/centinela/complete.go`. Dogfooding the real binary confirms only the
  memory-ledger write happens (verified and asserted); no stdout banner
  exists for a gatekeeper WARNING, unlike the analogous
  `workflow.ProductionReadinessWarning` /
  `ui.RenderProductionReadinessWarning` pair for the production-readiness
  role. This predates this feature (the pre-feature `complete.go` had no
  such print either) and is not new-feature scope creep to silently patch
  from the tests step. Recorded in
  `.workflow/adversarial-validate-verifier-edge-cases.md` under Residual
  Risks. Not roadmap-deferred (defer is for NEW out-of-scope gaps; this is
  in-scope and should be triaged by the validate step / a follow-up, not
  backlogged).
- `verify-crosscheck-verifier-commands` (arg/exit-code truthfulness
  cross-check) was already deferred on the roadmap during the plan step —
  no new test asserts it, by design (v1 enforces shape/existence, not
  truthfulness, per the plan).
- No other `.feature` scenario lacks an executable assertion — see the
  scenario map above.

#### Acceptance Wiring

`grep -A6 "\[validate\]" centinela.toml`:
```
[validate]
# `go test ./...` already runs the acceptance tier (tests/acceptance is under
# ./...) and satisfies the acceptance-execution gate; check-coverage.sh re-runs
# the full suite under `set -e` (fails on any test failure) for the coverage
# gate. A separate `go test ./tests/acceptance/...` would run the (uncacheable,
# ~200s) acceptance tier a redundant extra time on every validate/complete —
# removed to keep `centinela complete`'s claim-verification within the runner's
```
Acceptance execution is present (via inclusion in `go test ./...`, not a
dedicated separate command — deliberate, documented, and pre-existing).
`validate.commands` was NOT modified by this step.

#### Deferred Findings

None recorded from this step. The one in-scope gap found (WARNING stdout
surfacing) is documented above and in the edge-cases ledger rather than
deferred, since `centinela roadmap defer` is reserved for NEW,
out-of-scope findings.

#### Handoff

- Next role: validation-specialist (per `centinela evidence schema
  qa-senior`; the validate step's REQUIRED evidence role for
  `adversarial-v1` workflows is `gatekeeper`, but the qa-senior evidence
  schema's `handoffTo` default is unchanged by this feature — see Coverage
  Gaps for the related, but distinct, WARNING-surfacing gap).
- Edge-case report: `.workflow/adversarial-validate-verifier-edge-cases.md`
  (produced directly in this step per the current workflow; enumerates
  every negative path exercised and residual risks).
