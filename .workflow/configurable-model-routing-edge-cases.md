# Edge Cases — configurable-model-routing (tests step)

Enumerates every edge case the test suite covers, mapped to its test(s). All
three tiers (unit / integration / acceptance) plus colocated package tests
(coverage-bearing) are listed.

## Acceptance criteria coverage (7 AC)

| AC | What | Tests |
|----|------|-------|
| AC1 | Tier remap: `model_map.reasoning.opencode` used for opencode | `tests/acceptance/.../TestRouting_TierRemapForRunner`, `tests/unit/.../TestRouting_ResolveTierMapForRunner`, `tests/integration/.../TestRouting_ConfigToResolver_TierRemap`, `cmd/.../TestRunHookOrchestration_ModelMapRemap`, `internal/orchestration/.../TestResolveModel_TierMapOverride` |
| AC2 | Role override beats the role's tier for the runner | `tests/acceptance/.../TestRouting_RoleOverrideBeatsTier`, `tests/unit/.../TestRouting_ResolveRoleOverrideWins`, `tests/integration/.../TestRouting_ConfigToResolver_RoleOverride`, `cmd/.../TestRunHookOrchestration_RoleOverride`, `internal/orchestration/.../TestResolveModel_RoleOverrideWins` |
| AC3 | Tier override but no `model_map` entry for runner → built-in default | `tests/unit/.../TestRouting_ResolveNoMapEntryUsesDefault`, `internal/orchestration/.../TestResolveModel_TierOverrideNoMapEntry` |
| AC4 | Plain tier string in `[orchestration.models]` still loads (back-compat) | `tests/acceptance/.../TestRoutingConfig_PlainTierStringBackCompat`, `internal/config/.../TestRoleModelValue_UnmarshalString`, `cmd/.../TestRunHookOrchestration_PlainTierString` |
| AC5 | Malformed config fails loudly naming the key (4 cases) | unknown runner: `TestRoutingConfig_UnknownRunnerInModelMap` / `TestModelMap_UnknownRunnerRejected`; unknown role: `TestRoutingConfig_UnknownRoleInModels` / `TestUnion_UnknownRoleTableRejected`; unknown tier: `TestRoutingConfig_UnknownTierInModelMap` / `TestModelMap_UnknownTierRejected`; empty model: `TestRoutingConfig_EmptyModelInModelMap` / `TestModelMap_EmptyModelRejected` + `TestUnion_OverrideEmptyModelRejected` |
| AC6 | Absent tables → all built-in defaults | `tests/acceptance/.../TestOrchestrationHook_AbsentTableAllDefaults`, `tests/integration/.../TestRouting_ConfigToResolver_AbsentDefaults`, `cmd/.../TestRunHookOrchestration_AbsentTablesDefault` |
| AC7 | Active runner with no mapping → tier name + ok=false, never another runner's ID | `tests/acceptance/.../TestRouting_CodexRule4NoLeak`, `tests/unit/.../TestRouting_ResolveCodexFallback`, `cmd/.../TestRunHookOrchestration_ModelMapRemap` (codex column), `internal/orchestration/.../TestResolveModel_CodexRule4NoLeak` |

## Edge cases

1. **Empty tables == absent tables** — `[orchestration.model_map]` and
   `[orchestration.models]` present but empty resolve to all built-in defaults.
   `tests/acceptance/.../TestRouting_EmptyTablesDefault`,
   `internal/config/.../TestModelMap_AbsentAndEmptyValid`.
2. **Casing/whitespace normalization** on tier + runner keys (`" Reasoning "`,
   `" Opencode "`) — trimmed + lowercased before validation.
   `tests/acceptance/.../TestRoutingConfig_KeyNormalization`,
   `internal/config/.../TestModelMap_KeyNormalization`,
   `internal/orchestration/.../TestRoleTier_OverrideAndDefault` (Fast→fast).
3. **Mixed forms in `[orchestration.models]`** — one role a tier string, another
   a runner→model table, both valid in the same table.
   `tests/acceptance/.../TestRoutingConfig_MixedFormsLoad`,
   `internal/config/.../TestMixedForms_TierAndTableCoexist`.
4. **Role override beats `model_map`** for the same runner+tier (precedence rule
   1 over rule 2). `tests/acceptance/.../TestRouting_OverrideBeatsModelMap`,
   `internal/orchestration/.../TestResolveModel_RoleOverrideWins`.
5. **Codex before codex-support** (empty column) → rule-4 tier-name fallback,
   never a claude/opencode ID leak. `TestRouting_CodexRule4NoLeak` (acceptance),
   `TestResolveModel_CodexRule4NoLeak` (domain).
6. **Empty override string** in a role table → ignored, falls through to the
   tier path (does not return ""). `internal/orchestration/.../TestResolveModel_EmptyOverrideFallsThrough`.
7. **Invalid tier override string** → falls back to the role's default tier
   instead of erroring at resolve time. `internal/orchestration/.../TestRoleTier_OverrideAndDefault`.
8. **Nil models / nil model_map** → no panic, built-in defaults.
   `tests/unit/configurable_subagent_models_resolve_unit_test.go` (NilModelsMap),
   `internal/orchestration/resolve_test.go` (NilMapNoPanic).
9. **Non-string model value / non-string-non-table role value** in the union
   unmarshal → typed error. `internal/config/.../TestRoleModelValue_UnmarshalNonStringValueErrors`,
   `TestRoleModelValue_UnmarshalWrongTypeErrors`.
10. **Config-load failure is zero-config-safe in the hook** — a malformed config
    falls back to defaults rather than aborting the directive.
    `cmd/.../TestRunHookOrchestration_ConfigErrorFallsBack`.
11. **Out-of-band roles not emitted** (gatekeeper, merge-steward, etc.) — scope
    boundary preserved. `tests/acceptance/.../TestOrchestrationHook_OutOfBandRolesAbsent`.
12. **Allow-list parity** — config-leaf `allowedRunnerKeys` cannot drift from the
    domain `AllowedRunnerKeys()`. `tests/unit/.../TestAllowListParity_AllRunnerKeysAcceptedInModelMap`,
    `TestAllowListParity_AllRunnerKeysAcceptedInOverride`,
    `TestAllowListParity_UnknownRunnerKeyRejected`,
    `internal/orchestration/.../TestAllowedRunnerKeys_ThreeStable`.

## Predecessor tests fixed

- `tests/unit/configurable_subagent_models_resolve_unit_test.go` — migrated all
  5 calls to the new 4-arg `ResolveModel(role, models, modelMap, runner)` with
  `RoleModels` literals.
- `tests/acceptance/configurable_subagent_models_test.go` — updated
  `TestOrchestrationHook_ConfiguredTierAnnotated` and `_AbsentTableAllDefaults`
  to the new per-runner `model: <id> (<runner>)` annotation format.
- `tests/acceptance/configurable_subagent_models_config_test.go` — updated
  `TestOrchestrationHook_NormalizedTierAccepted` to the per-runner format.
