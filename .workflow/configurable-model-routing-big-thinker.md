### Big-Thinker Report: configurable-model-routing
**Date:** 2026-06-05

> Plan file: [`docs/plans/configurable-model-routing.md`](../docs/plans/configurable-model-routing.md).
> Feature brief: [`docs/features/configurable-model-routing.md`](../docs/features/configurable-model-routing.md).
> Predecessor: `configurable-subagent-models` (shipped role→tier; deferred concrete model IDs).

#### Problem

Centinela resolves a subagent's model in two layers: `role → tier` (user-config,
shipped) and `tier → concrete model` — a hardcoded, Anthropic-only table in
`internal/orchestration/resolve.go`. The second layer cannot change without
editing Go, so operators on opencode/codex (and Claude Code users behind an
OpenRouter-style gateway) cannot point a tier — or a single role — at the model
they actually want (Kimi for reasoning, DeepSeek-Coder for senior-engineer).
This feature opens the tier→model table to config (`[orchestration.model_map]`)
and adds a role-level override (union-typed `[orchestration.models]`), keyed by
runner, while every unconfigured role/tier/runner keeps the built-in Anthropic
default. It stays advisory: the hook annotates the delegate directive and the
orchestrator complies — no evidence-schema change, no `centinela complete` gate.

#### Scope

- **In:** `internal/config/` new `model_map` table + union-typed `models`
  (custom TOML unmarshal, shape validation, local runner/tier/role sets +
  parity test); `internal/orchestration/` `RunnerCodex` enum value, 4-step
  precedence `ResolveModel(role, models, modelMap, runner)`, codex column in
  `tierModels`, all-runners `ModelReference`; thin
  `cmd/centinela/hook_orchestration.go` (emit the all-runners reference line,
  include codex in the model-reference wording); Gherkin spec + config docs.
- **Out:** any `CENTINELA_RUNNER`/`--runner` signal or `internal/setup/` runner
  detection (locked decision #1); out-of-band role emission; recording/verifying
  the used model in evidence; provider-availability checks; codex's concrete
  default IDs (filled by `codex-support` — only the column/key ship here).

#### Dependencies & Assumptions

- Builds on `configurable-subagent-models`: `Tier`/`Role` enums,
  `DefaultTierForRole`, `NormalizeTier`, `AllowedTiers`, `AllowedRoleSlugs`,
  `RequiredRolesForFeature`, the `Runner` enum, `tierModels`,
  `ResolveModel`/`ModelReference` already exist and are unit-tested.
- `internal/config/` is a leaf and MUST NOT import `internal/orchestration/`;
  runner/tier/role string sets are duplicated locally and reconciled by a
  cross-package parity test (existing pattern).
- TOML decoder (already wired in `config.Load`) supports a custom unmarshal hook
  on a wrapper type for the union field — exact interface to be confirmed in code.
- Hook stays runner-agnostic at emit time (locked decision #1): `runner` is a
  resolver parameter and the reference line enumerates all three runners, so no
  runtime runner signal is needed.
- Hard rule: every NEW source file and `_test.go` in `internal/`/`cmd/` ≤100
  lines — resolver precedence and config union/validator split across files.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Union-typed `models` (string \| table) breaks existing plain-string configs | High | Medium | Custom unmarshal accepting both forms; back-compat acceptance test (AC#4); keep the tier accessor path stable. |
| Resolver/config files exceed the 100-line hard rule | Medium | High | Split precedence helper, union type + custom unmarshal, and `model_map` validator into separate ≤100-line files; split `_test.go` by concern. |
| Opaque model strings → typo points at a non-existent model | Medium | Medium | Shape-only validation (non-empty, known runner/tier/role); availability is the runner's job; advisory directive degrades, not crashes. |
| Codex column empty before `codex-support` → wrong-vendor ID leaks under codex | Low | High | Rule 4 returns the tier name + `ok=false`; never emit a claude/opencode ID for codex (AC#7). |
| All-runners reference line noisier with a third runner | Low | Medium | Keep one compact line factored by tier; append codex column; dedupe tiers in stable order. |
| Config leaf drifts from domain allowed-sets (new runner key) | Medium | Medium | Extend parity test to `allowedRunnerKeys` ↔ domain runner set; fails loudly on drift. |
| Advisory only — orchestrator may ignore the model hint | Medium | Medium | Accept by design (consistent with directive model); revisit an evidence-check later. |

#### Rollout

- **Step 1 — codex runner key + codex column (data-only, smallest correct
  slice).** Add `RunnerCodex` to the enum, an empty codex entry to `tierModels`,
  render codex in `ModelReference`, update the hook's reference wording. Test
  codex → `ok=false` (rule 4) and a three-runner reference line. No new config
  surface; everything still defaults.
- **Step 2 — tier remap via `[orchestration.model_map]` (base layer).** Add
  `ModelMap` + accessor + shape validation; thread `modelMap` into `ResolveModel`
  for rule 2 (tier-map override) and rule 3 (default). Covers AC#1, #3, #6, #7.
- **Step 3 — role override via union-typed `[orchestration.models]`.** Union
  wrapper + custom TOML unmarshal (string OR runner→model table) + table
  validation; apply rule 1 ahead of step 2. Covers AC#2, #4, #5.
- **Step 4 — docs + Gherkin spec + parity test + acceptance wiring.** Scenarios
  1:1 with the 7 AC (incl. codex rule-4 fallback, mixed forms); config reference
  for both tables; extend parity test; add acceptance to `validate.commands`.

#### Handoff

- Next role: feature-specialist
- Outstanding questions:
  1. Confirm the TOML decoder interface for the union field (custom
     `UnmarshalTOML` on a wrapper vs. decode-to-`map[string]any` then narrow);
     pick the one that stays ≤100 lines and preserves plain-string back-compat.
  2. Finalize the `ResolveModel` shape for the union `models` argument (a typed
     `RoleModels` carrying optional tier string + optional runner→model map, vs.
     two parallel maps).
  3. Author Gherkin 1:1 with the 7 acceptance criteria, including codex rule-4
     fallback and the mixed-forms edge case.
  4. Decide whether the delegate line shows resolved IDs per runner inline or
     relies solely on the factored `ModelReference` line (locked decision #1
     favors the factored line; confirm directive readability).
