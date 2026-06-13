# deterministic-artifact-scaffolds — qa-senior

## Test Inventory

Colocated unit tests (drive the coverage gate):

- `internal/orchestration/required_plan_inputs_test.go` (83) — `RequiredPlanInputs`: brief + every `docs/features/*.md` + plan path; sorted; deduped; `normalizeFeatureDocPath` slash/dot/backslash normalization.
- `internal/evidence/plan_inputs_test.go` (60) — `PlanInputs` delegates to `RequiredPlanInputs` for both plan roles, nil for every other `AllRoles()` role.
- `internal/evidence/fill_companion_test.go` (83) — `FillSlot`/`FillMarker`; `companionSkeleton` per report-bearing role + unknown-role fallback; `DefaultCompanionTemplate` role-aware + one-line fallback; fill marker never in marshaled JSON.
- `internal/evidence/artifact_bodies_test.go` (74) — `analyzedSpecsList` sorted-present / empty; gatekeeper + prodready keep Status/Date + fill; edge-cases + changelog use fill.
- `internal/evidence/skeleton_not_poisoned_test.go` (45) — `Skeleton`/`SchemaSkeleton`(repair)/`docsSpecialistPair` inputs stay empty.
- `cmd/centinela/evidence_init_prefill_test.go` (54) — `runEvidenceInit` pre-fills plan path + a docs/features path for big-thinker; leaves senior-engineer inputs empty.

Acceptance tests (1:1 scenario traceability, all under `tests/acceptance/`):

- `deterministic_artifact_scaffolds_helper_test.go` (74) — shared harness mirroring the init pre-fill without importing package main.
- `deterministic_artifact_scaffolds_prefill_test.go` (99), `_validate_test.go` (87), `_emptyfields_test.go` (69), `_fill_test.go` (79), `_artifacts_test.go` (83).

## Coverage Gaps

None blocking. Every new function is at 100%: `PlanInputs`, `FillSlot`, `companionSkeleton`, `DefaultCompanionTemplate`, `analyzedSpecsList`, `gatekeeperBody`, `prodReadyBody`, `edgeCasesBody`, `changelogBody`, `RequiredPlanInputs`, `normalizeFeatureDocPath`, `validatePlanSnapshotInputs`. `runEvidenceInit` at 84.2% (untested lock/write error arms only). Total coverage gate: 95.5% >= 95.0%.

## Acceptance Wiring

All 21 `.feature` scenarios map 1:1 via `// Scenario: <exact title>` markers; the `comm -23` title-diff is empty. `centinela.toml` validate.commands already runs `go test ./tests/acceptance/...`. `.workflow/deterministic-artifact-scaffolds-edge-cases.md` enumerates the covered edges and residual risks.

## Semantic-change note

`RequiredPlanInputs` was extended during the code step to ALSO include the plan path `docs/plans/<feature>.md` (previously `docs/features/*.md` only), aligning code with the locked spec and `evidence-contract.md`. Consequence: `validatePlanSnapshotInputs` now also requires the plan path. New tests assert BOTH the init pre-fill AND the validator-required set include the plan path. Full suite: 1506 passing (was 1465).

A spec/impl discrepancy is documented in the edge-cases file: the empty-specs gatekeeper scenario text says no fill rows, but the approved impl renders one cosmetic fill placeholder row; the acceptance test asserts the truthful shipped behavior (no real spec paths listed).

## Handoff

handoffTo: validation-specialist
