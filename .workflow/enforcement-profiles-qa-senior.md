# QA-Senior Report: enforcement-profiles

**Date:** 2026-06-12
**Feature:** enforcement-profiles
**Handoff →** validation-specialist

## Summary

Tests-only delivery for the enforcement-profiles feature (strict default / guided
/ outcome). Added colocated per-package unit tests restoring the four touched
packages to ≥95% coverage (overall 95.2%), one integration test, and a 12-scenario
acceptance suite that closes the spec-traceability gate on
`specs/enforcement-profiles.feature`. The core safety claim — verification is
constant across every profile — is proven by a behavioral invariant test at both
the in-process and full-binary level.

## Test Inventory

### Unit (colocated, same package as code under test)
| File | Lines | Targets |
|---|---|---|
| `internal/config/enforcement_profile_test.go` | 41 | `NormalizeEnforcementProfile` (each value/unknown/case/space), `validateEnforcementProfile` (empty+known OK, unknown errors naming field) |
| `internal/config/profile_defaults_test.go` | 46 | `ProfileDefaults` per-profile knobs; unknown→strict; only-strict-requires-evidence |
| `internal/config/enforcement_profile_load_test.go` | 53 | `Load` unknown-profile rejection, known accepted, `RawStepConfirmationMode` raw-capture |
| `internal/workflow/profile_test.go` | 49 | `EffectiveProfile` precedence + nil inputs; `DisplayProfile` |
| `internal/workflow/order_profile_test.go` | 41 | `NewWithOrder` orchestration mode per profile + pinned field; `New` defaults strict |
| `internal/hookpolicy/prewrite_profile_test.go` | 58 | outcome bypass; strict+guided block; no-workflow block under any profile |
| `cmd/centinela/review_mode_profile_test.go` | 56 | `effectiveConfirmationMode` precedence; outcome suppress; explicit beats profile |
| `cmd/centinela/start_profile_test.go` | 51 | `start --profile` persistence; global inheritance |
| `cmd/centinela/invariant_verification_test.go` | 48 | **INVARIANT**: `executeValidation` blocks under every profile (+clean control) |
| `internal/ui/render_status_profile_test.go` | 24 | `RenderStatus` Profile line (pinned + strict default) |

### Integration
| File | Lines | Targets |
|---|---|---|
| `tests/integration/enforcement_profiles_integration_test.go` | 53 | `start --profile outcome` end-to-end → persisted state → prewrite allows code in plan |

### Acceptance (`// Acceptance: specs/enforcement-profiles.feature`)
| File | Lines | Scenarios |
|---|---|---|
| `tests/acceptance/enforcement_profiles_test.go` | 47 | 1, 2, 3 |
| `tests/acceptance/enforcement_profiles_confirm_test.go` | 52 | 4, 5 |
| `tests/acceptance/enforcement_profiles_evidence_test.go` | 44 | 6, 7, 8, 9 |
| `tests/acceptance/enforcement_profiles_invariant_test.go` | 41 | 11 (full binary) |
| `tests/acceptance/enforcement_profiles_config_test.go` | 45 | 10, 12 |

All files ≤100 lines (G1). gofmt clean, go vet clean.

## Coverage Gaps — 12-scenario → test mapping (MUST be none)

| # | Scenario | Acceptance test |
|---|---|---|
| 1 | Outcome profile allows writing code during the plan step | `TestEP_OutcomeAllowsCodeWriteDuringPlan` |
| 2 | Strict and guided profiles still block out-of-step writes | `TestEP_StrictAndGuidedBlockOutOfStepWrites` |
| 3 | A write with no active workflow is always blocked | `TestEP_NoActiveWorkflowAlwaysBlocked` |
| 4 | Outcome profile suppresses the stop-and-ask review prompt | `TestEP_OutcomeSuppressesReviewPrompt` |
| 5 | An explicit confirmation mode overrides the profile default | `TestEP_ExplicitConfirmationOverridesProfile` |
| 6 | Strict profile requires subagent orchestration evidence | `TestEP_StrictRequiresSubagentEvidence` |
| 7 | Guided profile does not require subagent orchestration evidence | `TestEP_GuidedNoSubagentEvidence` |
| 8 | Outcome profile does not require subagent orchestration evidence | `TestEP_OutcomeNoSubagentEvidence` |
| 9 | A per-feature profile overrides the global setting | `TestEP_PerFeatureOverridesGlobal` |
| 10 | An unconfigured project keeps today's behavior (default strict) | `TestEP_UnconfiguredKeepsStrictBehavior` |
| 11 | Gates and claim verification run under every profile | `TestEP_GatesRunUnderEveryProfile` |
| 12 | An unknown profile value is rejected at config load | `TestEP_UnknownProfileRejectedAtLoad` |

Gaps: **none** — the spec-traceability gate reports "All 12 scenarios have acceptance coverage."

## The Invariant Test

`TestExecuteValidation_BlockedUnderEveryProfile` (cmd, in-process) and
`TestEP_GatesRunUnderEveryProfile` (acceptance, full binary) each set up a temp
project with `enforcement_profile = {strict|guided|outcome}` and a `[validate]
commands = ["exit 1"]`, then assert completion is BLOCKED in all three cases. The
assertion is "profile does not change the verdict": the failing-command run errors
identically, while `TestExecuteValidation_PassesUnderEveryProfileWhenClean`
(control, `exit 0`) passes identically — isolating the block to the failure, not
the profile. This rides the same `executeValidation()` gate `complete.go` runs at
the validate step; `complete.go` adds no profile branch. **Result: PASS for all
three profiles at both levels.**

## Acceptance Wiring

- Each acceptance file carries the header `// Acceptance: specs/enforcement-profiles.feature`.
- Each of the 12 `// Scenario: <name>` comments copies the spec text verbatim and
  sits directly above a real test func with a genuine assertion (no orphan comments).
- `validate.commands` already includes `go test ./tests/acceptance/...`.
- Dogfood: a fresh binary's `validate` reports the spec-traceability gate as
  ✓ "All 12 scenarios have acceptance coverage."

## Handoff → validation-specialist

- `gofmt -l cmd internal tests` → empty.
- `go vet ./...` → clean.
- `go test ./...` → 1313 passed, 24 packages.
- `./scripts/check-coverage.sh` → 95.2% ≥ 95.0% (gate passes).
- Spec-traceability gate → all 12 scenarios covered.
- Ready for gatekeeper + validate.
