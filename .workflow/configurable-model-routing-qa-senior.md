### QA-Senior Report: configurable-model-routing
**Date:** 2026-06-06

#### Test Inventory
| Tier | File | Scenarios |
|------|------|-----------|
| unit | tests/unit/configurable_model_routing_resolve_unit_test.go | AC1 tier-map, AC2 override wins, AC3 no-map-default, AC7 codex rule-4 |
| unit | tests/unit/configurable_model_routing_parity_unit_test.go | allowedRunnerKeys parity (model_map + override forms), unknown-runner reject |
| unit | tests/unit/configurable_subagent_models_resolve_unit_test.go | predecessor: migrated to 4-arg ResolveModel (override-beats-default, exact IDs, unknown-runner fallback, nil-map) |
| integration | tests/integration/configurable_model_routing_integration_test.go | config->resolver e2e: tier remap (AC1), role override (AC2), absent defaults (AC6) |
| acceptance | tests/acceptance/configurable_model_routing_test.go | AC1 tier remap, AC2 override beats tier, AC7 codex no-leak, edge override-beats-map, edge empty-tables |
| acceptance | tests/acceptance/configurable_model_routing_config_test.go | AC5 (unknown runner/role/tier, empty model), AC4 back-compat, mixed forms, key normalization |
| acceptance | tests/acceptance/configurable_subagent_models_test.go | predecessor: updated annotation format (AC1/AC6 defaults, out-of-band absent) |
| acceptance | tests/acceptance/configurable_subagent_models_config_test.go | predecessor: updated normalized-tier annotation format |

Coverage-bearing colocated package tests (same package as the new code, so they
count toward the per-package coverage gate):
- internal/orchestration/model_routing_test.go — all 4 ResolveModel precedence
  branches, empty-override fall-through, RoleTier override/default/invalid,
  AllowedRunnerKeys order.
- internal/config/orchestration_model_map_test.go — model_map validator (valid,
  unknown runner/tier, empty model, key normalization, absent/empty),
  allowedRunnerKeysList.
- internal/config/orchestration_models_union_test.go — RoleModelValue.UnmarshalTOML
  (string, table, non-string value error, wrong-type error), union override
  validation (unknown role/runner, empty model), accessor.
- internal/config/orchestration_routing_test.go — nil-safe accessors, mixed
  forms coexist, model_map accessor.
- cmd/centinela/hook_orchestration_routing_test.go — orchestrationRouting through
  the hook: model_map remap, role override, absent defaults, plain tier string,
  config-error fallback.

#### Coverage Gaps
None. All 7 acceptance criteria and all 12+ enumerated edge cases have at least
one executable assertion (see .workflow/configurable-model-routing-edge-cases.md
for the AC->test and edge->test matrices). The directive's runner-agnostic
all-runners annotation is asserted per runner column (claude/opencode/codex),
including the codex rule-4 fallback that must never leak another runner's ID.

#### Acceptance Wiring
`centinela.toml` already runs the acceptance tier via the full Go suite:
```toml
[validate]
commands = [
  "go test ./...",
  "./scripts/check-coverage.sh"
]
```
`go test ./...` executes tests/acceptance/* (the binary-driven hook scenarios)
in addition to unit/integration. No change to validate.commands was needed.

#### Regression Guards
- Predecessor breakage fixed: the old 3-arg ResolveModel call site
  (tests/unit/configurable_subagent_models_resolve_unit_test.go) and the three
  acceptance tests asserting the old `(model: <tier>)` directive string were
  migrated to the new 4-arg signature and per-runner `model: <id> (<runner>)`
  format — these now guard the new contract.
- TestRunHookOrchestration_ConfigErrorFallsBack guards the zero-config-safe hook
  path so a malformed config can never abort the directive.
- TestAllowListParity_* guards against the config-leaf runner-key set drifting
  from the domain AllowedRunnerKeys() (the leaf cannot import orchestration).
- TestRouting_CodexRule4NoLeak / TestResolveModel_CodexRule4NoLeak guard against
  a claude/opencode model ID leaking into the empty codex column.

#### Results
- `go test ./...` : 1109 passed, 24 packages, exit 0.
- Coverage gate (`scripts/check-coverage.sh`, MIN_COVERAGE=95.0): total 95.4% >= 95.0% — PASS.
- Touched-package coverage: internal/config 100.0%, internal/orchestration 96.9%,
  cmd/centinela 93.2% (pre-existing baseline 93.0%; the touched functions
  orchestrationRouting/resolvedPerRunner/annotateRoles are 100%). The validate
  gate enforces the total, which passes.
- Every new/edited _test.go in internal/ and cmd/ and under tests/ is <=100 lines.

#### Handoff
- Next role: validation-specialist.
