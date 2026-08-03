# guided-by-default — qa-senior

**Date:** 2026-08-03

## Test Inventory

Senior-engineer landed slices 1, 2, 3, 5 in full plus slice 4's colocated
source-guard half (`cmd/centinela/complete_validate_gates_invariant_test.go`),
and 15 unit-tier files enumerated in
`.workflow/guided-by-default-senior-engineer.md`. This step adds the
acceptance-tier half of slice 4 (the proof-parity table) plus binary-driven
coverage for every remaining spec scenario.

| Tier | File | Scenarios |
|------|------|-----------|
| acceptance | `tests/acceptance/enforcement_profiles_invariant_helpers_test.go` | shared fixture builders (`gbdSeedWorkflow`, `gbdMakeSafe`) — no scenario of its own |
| acceptance | `tests/acceptance/enforcement_profiles_invariant_report_test.go` | missing gatekeeper report; ungrounded verdict — both profiles |
| acceptance | `tests/acceptance/enforcement_profiles_invariant_freshness_test.go` | stale verification (revision skew) — both profiles |
| acceptance | `tests/acceptance/enforcement_profiles_invariant_prodready_test.go` | BLOCKING production readiness — both profiles |
| acceptance | `tests/acceptance/enforcement_profiles_invariant_gatefail_test.go` | failing validate command — both profiles |
| acceptance | `tests/acceptance/enforcement_profiles_invariant_clean_test.go` | clean tree completes both profiles; guided-skips-evidence divergence |
| acceptance | `tests/acceptance/guided_by_default_self_governance_test.go` | this repo's centinela.toml pins strict explicitly |
| acceptance | `tests/acceptance/guided_by_default_default_flip_test.go` | zero-config → guided; legacy workflow → strict |
| acceptance | `tests/acceptance/guided_by_default_precedence_test.go` | --profile, global, driver-model tiers outrank the default |
| acceptance | `tests/acceptance/guided_by_default_cascade_test.go` | guided cold start; strict full cascade; setup hook advise/halt |
| acceptance | `tests/acceptance/guided_by_default_start_refusals_test.go` | Backlog/draft/no-bootstrap-phase refusals; missing roadmap json |
| acceptance | `tests/acceptance/guided_by_default_quality_test.go` | low score advisory; ignored threshold; scores shape/type faults |
| acceptance | `tests/acceptance/guided_by_default_quality_promote_test.go` | promote with low score; out-of-range refusal; artifact not rewritten |
| acceptance | `tests/acceptance/guided_by_default_doctor_test.go` | doctor advises inherited default; silent when pinned |
| unit (senior-engineer, unchanged) | `internal/workflow/profile_contract_test.go`, `internal/config/project_profile_test.go`, `internal/roadmap/quality_advisory_test.go`, `internal/doctor/check_profile_default_test.go`, `cmd/centinela/start_guard_cascade_test.go`, `start_guard_guided_refusals_test.go`, `hook_setup_profile_test.go`, `roadmap_promote_range_test.go` | tier/provenance tables; unmet-dependency and bootstrap-incomplete refusals |

## Coverage Gaps

All 30 scenarios in `specs/guided-by-default.feature` now have a matching
`// Scenario:` acceptance marker (verified by direct parse of the spec against
`tests/acceptance/*.go` headers — zero gaps). Two AC8-outline examples (unmet
dependencies, non-bootstrap-while-incomplete) are exercised only at unit tier;
see `.workflow/guided-by-default-edge-cases.md` → Residual Risks for why that
is accepted rather than deferred.

## Acceptance Wiring

```toml
[validate]
commands = [
  "go test ./... -coverprofile=coverage.out",
  "COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh",
  "./scripts/check-fmt.sh"
]
```

`go test ./...` includes `tests/acceptance/...` (no separate acceptance
invocation exists or is needed — the single profiled run is this repo's
documented single-run design).

## Deferred Findings

None genuinely new. `profile-display-unused-and-now-misleading` was already
deferred by senior-engineer (source `guided-by-default/senior-engineer`) and
is not re-raised. The two inert `ProfileKnobs` fields remain out of scope per
the brief.

## Handoff

- Next role: gatekeeper (this workflow is pinned to the `adversarial-v1`
  validate contract; `centinela evidence init` resolved `handoffTo`
  accordingly rather than the legacy `validation-specialist`).
- Edge-case report: `.workflow/guided-by-default-edge-cases.md` (produced via
  `centinela artifact new`, filled with negative-path `file:test` references).
