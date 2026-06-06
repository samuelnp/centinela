### Senior-Engineer Report: configurable-model-routing
**Date:** 2026-06-06

Implements configurable model routing as one feature: tier-remap via
`[orchestration.model_map]` and union-typed role-override via
`[orchestration.models]`, with a runner-agnostic all-runners reference line
(no `CENTINELA_RUNNER` signal, no `internal/setup/` changes).

#### Files Touched
| Path | Reason |
|------|--------|
| internal/orchestration/resolve.go | Added `RunnerCodex` enum + codex column (empty) to `tierModels`; `ModelReference` now renders all three runners (codex falls back to tier name when empty). |
| internal/orchestration/model_routing.go | New: `RoleModel`/`RoleModels`/`ModelMap` types + the 4-step precedence resolver `ResolveModel(role, models, modelMap, runner)` and `RoleTier`. |
| internal/orchestration/models.go | Added `AllowedRunnerKeys()` so the config leaf's set can be parity-checked against the domain. |
| internal/config/orchestration.go | `Models` evolved to `map[string]RoleModelValue` (union); added `ModelMap`; accessors `OrchestrationModelTiers`/`OrchestrationModelOverrides`/`OrchestrationModelMap`; kept `OrchestrationModels` as a back-compat alias. |
| internal/config/orchestration_models.go | Validate union entries: tier-string form OR runner→model table form (known runner key, non-empty model). |
| internal/config/orchestration_model_map.go | New: `RoleModelValue.UnmarshalTOML` (string \| table), `allowedRunnerKeys` set, and `model_map` shape validator (known tier, known runner, non-empty model, normalized). |
| internal/config/file_size_exceptions.go | Wired `validateOrchestrationModelMap` into `validateConfig`. |
| cmd/centinela/hook_orchestration.go | Thin: maps config tables → domain types via `orchestrationRouting`; unchanged emission shape. |
| cmd/centinela/orchestration_annotate.go | Thin: per-role annotation now lists each runner's resolved ID (runner-agnostic). |
| internal/orchestration/resolve_test.go | Minimal existing-test fix to the new 4-arg `ResolveModel` signature (keeps `go build`/package tests green). |

#### Precedence (how each of the 4 steps is implemented — `model_routing.go`)
1. **Role override** — `models[role].Overrides[runner]` non-empty → return it.
2. **Role→tier then tier-map** — `roleTier(role, override)` resolves the effective tier (explicit valid tier override else built-in default); `modelMap[tier][runner]` non-empty → return it.
3. **Built-in default** — `tierModels[tier][runner]` non-empty → return it.
4. **Missing** — return `string(tier)` with `ok=false`. Codex (empty column) hits this and never leaks a claude/opencode ID.

#### Architecture Compliance
- Boundary checks passed: `internal/config` imports nothing from `internal/orchestration` (leaf preserved); cmd/ maps config→domain types with no decision logic (G7).
- G1 file size: every modified source file ≤ 100 lines (max 82, `orchestration.go`).
- `go build ./...` and `go vet ./internal/... ./cmd/...` both clean.

#### Type-Safety Notes
- Union handled by a typed `RoleModelValue` with an explicit `UnmarshalTOML(any)` switching on `string` vs `map[string]any` (no `any` leaks into the domain).
- Resolver takes typed `RoleModels`/`ModelMap`/`Runner`; returns `(string, bool)` — opaque model strings, no reflection or dynamic lookups.

#### Trade-Offs
- Parallel `RoleModels`/`ModelMap` domain types (built by cmd/ from config accessors) rather than importing config into orchestration — keeps the leaf boundary intact at the cost of a small mapping in the hook.
- Per-role annotation enumerates all runners (locked decision #1) instead of one ID — chosen because the hook has no runtime runner signal.

#### Handoff
- Next role: qa-senior
- Outstanding TODOs (tests step, hook blocks me from editing `tests/` now):
  - Update `tests/unit/configurable_subagent_models_resolve_unit_test.go` to the 4-arg `ResolveModel` signature.
  - Update predecessor acceptance tests asserting the OLD `(model: <tier>)` annotation to the new per-runner `model: <id> (<runner>)` format (`TestOrchestrationHook_ConfiguredTierAnnotated`, `_NormalizedTierAccepted`, `_AbsentTableAllDefaults`).
  - Add parity coverage for `allowedRunnerKeys` ↔ `orchestration.AllowedRunnerKeys()`.
  - Author the 7-AC Gherkin-mapped tests + edge cases (codex rule-4, mixed forms, role-override-beats-map).
