# workflow-archetypes — qa-senior

**Date:** 2026-06-12
**Handoff →** validation-specialist

## Test Inventory

### Unit (colocated, per-package)

| File | Lines | Covers |
|------|------:|--------|
| `internal/workflow/archetype_safety_test.go` | 49 | **SAFETY TEST.** Pins resolved order per archetype + the `contains("validate")` property the ship gate keys on. spike == `[plan,code]`, no validate; hotfix/refactor/canonical contain validate. |
| `internal/workflow/archetype_test.go` | 51 | `NormalizeArchetype` (empty→canonical, known/unknown passthrough, case+space), `ValidateArchetype` (empty+4 ok, unknown error names value+field), `DisplayArchetype` (each + spike annotation + nil). |
| `internal/workflow/archetype_order_test.go` | 41 | `ArchetypeStepOrder` (4 known orders + unknown→nil,false) and clone-not-alias mutation safety. |
| `internal/workflow/archetype_state_test.go` | 48 | State Save/Load round-trip of `Archetype` + StepOrder; orthogonality (spike order × strict profile). |
| `internal/roadmap/archetype_test.go` | 54 | `Feature.Archetype` JSON parse; `FeatureArchetype` accessor (hit/empty/missing/nil); `ValidateDependencies` rejects unknown archetype naming the feature. |
| `internal/ui/render_status_archetype_test.go` | 27 | `RenderStatus` archetype line + spike "no ship gate" annotation; unpinned→canonical. |
| `cmd/centinela/start_archetype_test.go` | 64 | `archetypeOrderByName` (validate+resolve); `resolveArchetypeOrder` precedence: flag > roadmap > canonical fallthrough. |

### Integration (`tests/integration`)

| File | Lines | Covers |
|------|------:|--------|
| `workflow_archetypes_integration_test.go` | 59 | Real `centinela start --archetype hotfix` → persisted `[code,tests,validate]`; `--archetype spike` → `[plan,code]`, no validate. Archetype + order flow from the built binary's start path. |

### Acceptance (`tests/acceptance`) — spec-traceability closure

All three files carry `// Acceptance: specs/workflow-archetypes.feature`. Each
`// Scenario:` comment (exact spec text) sits directly above one real test func.

| File | Lines | Scenarios |
|------|------:|-----------|
| `workflow_archetypes_test.go` | 62 | 1–4 (hotfix/refactor/spike/canonical orders via `ArchetypeStepOrder`) |
| `workflow_archetypes_gate_test.go` | 64 | 5–8 (ship-gate-reaches-validate, spike-never-gated, flag-overrides-roadmap, archetype-pinned-in-state) |
| `workflow_archetypes_state_test.go` | 44 | 9–11 (orthogonality, unknown rejected, status shows archetype) |

## Coverage Gaps — 11-scenario → test mapping (must be NONE)

| # | Scenario (spec) | Test func | Honest channel |
|---|-----------------|-----------|----------------|
| 1 | The hotfix archetype resolves to a code-tests-validate order | `TestWA_HotfixOrder` | `ArchetypeStepOrder` |
| 2 | The refactor archetype resolves to a plan-code-tests-validate order | `TestWA_RefactorOrder` | `ArchetypeStepOrder` |
| 3 | The spike archetype resolves to a plan-code order with no validate step | `TestWA_SpikeOrder` | `ArchetypeStepOrder` (no validate) |
| 4 | The default archetype is the canonical five-step order | `TestWA_CanonicalDefault` | `NormalizeArchetype`+`ArchetypeStepOrder` |
| 5 | A ship-gated archetype runs gates and claim verification | `TestWA_ShipGatedArchetypeReachesValidate` | order contains validate ⇔ gate fires |
| 6 | A spike never reaches the ship gate | `TestWA_SpikeNeverReachesShipGate` | order omits validate ⇔ gate never fires |
| 7 | An explicit archetype flag overrides the roadmap archetype | `TestWA_FlagOverridesRoadmapArchetype` | precedence (also `resolveArchetypeOrder` in cmd) |
| 8 | The active archetype is pinned in the workflow state | `TestWA_ArchetypePinnedInState` | `Save`/`Load` round-trip |
| 9 | Archetype and enforcement profile are independent | `TestWA_ArchetypeIndependentOfProfile` | `NewWithOrder(spike, strict)` |
| 10 | An unknown archetype value is rejected | `TestWA_UnknownArchetypeRejected` | `ValidateArchetype` names field |
| 11 | The status output shows the active archetype | `TestWA_StatusShowsArchetype` | `RenderStatus`/`DisplayArchetype` |

**Gaps: none.** Dogfood confirms: `spec-traceability-gate All 11 scenarios have acceptance coverage.`

## Acceptance Wiring

- `validate.commands` already includes `go test ./tests/acceptance/...` and the
  spec-traceability gate — both green for this feature's spec (11/11 covered).
- Matcher normalization (trim / collapse spaces / strip one trailing period /
  lowercase) — scenario comments copied verbatim from `specs/workflow-archetypes.feature`.
- 11 `// Scenario:` comments ↔ 11 test funcs, each comment directly above its func
  (verified: no orphan comments).

## Verification

- `gofmt -l cmd internal tests` → empty.
- `go vet ./...` → clean (No issues found).
- `go test ./...` → **1342 passed in 24 packages**, 0 fail.
- `./scripts/check-coverage.sh` → **coverage gate passed: 95.3% >= 95.0%**.
  Archetype functions: `ArchetypeStepOrder`/`NormalizeArchetype`/`ValidateArchetype`/
  `DisplayArchetype`/`FeatureArchetype`/`resolveArchetypeOrder`/`archetypeOrderByName`
  all 100.0%; `archetypeLine` 100%.
- Dogfood: `cent validate` → `spec-traceability-gate All 11 scenarios have acceptance coverage.`
- All 11 new test files ≤100 lines (max 64).

## Safety-test result

spike resolves to exactly `[plan, code]` and contains **no** `validate` element;
hotfix/refactor/canonical all **contain** `validate`. The ship gate keys on the
presence of the `validate` step (step-keyed, complete.go), so spike is ungated by
absence — there is no archetype bypass branch.

## Handoff → validation-specialist

Source + tests green; coverage ≥95% per package; spec-traceability 11/11 covered;
edge-cases report at `.workflow/workflow-archetypes-edge-cases.md`. Ready for the
gatekeeper report + `centinela validate` gate run.
