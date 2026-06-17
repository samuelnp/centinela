# Plan: Configurable Model Routing

> Implements [`docs/features/configurable-model-routing.md`](../features/configurable-model-routing.md).
> Follow-up to [`configurable-subagent-models`](../features/configurable-subagent-models.md)
> (which shipped role→tier and deferred concrete/cross-provider model IDs).

## Locked decisions (do not relitigate)

1. **Runner identity at emit time = "emit all-runners reference line."** The
   directive stays runner-agnostic. The hook emits EACH runner's resolved
   concrete model ID (claude / opencode / codex) on the delegate/reference line
   and lets the orchestrator pick its own row. NO runner signal is injected at
   wiring time. NO changes to `internal/setup/` for runner detection. There is
   no `CENTINELA_RUNNER` / `--runner` work in v1. This is option (a) from the
   predecessor's runner-resolution analysis, now extended to three runners.
2. **Ship as ONE feature.** Tier-remap via `[orchestration.model_map]` AND
   role-override via union-typed `[orchestration.models]` land together. Do NOT
   split into `model-routing-tier-remap` / `model-routing-role-override`.

## Problem framing

Centinela's model selection has two layers: `role → tier` (user-configurable,
shipped) and `tier → concrete model` (a hardcoded, Anthropic-only table at
`internal/orchestration/resolve.go`). The second layer is welded to Go source,
so operators on opencode/codex — or Claude Code users behind an OpenRouter-style
gateway — cannot point a tier (or a single role) at the model they actually
want (Kimi for reasoning, DeepSeek-Coder for senior-engineer, etc.). Two unmet
needs: **vendor substitution** (remap a whole tier per runner) and **per-role
precision** (big-thinker and senior-engineer share the `reasoning` tier, so a
tier remap alone cannot give coding its own model). This feature opens the
tier→model table to config and adds a role-level override, while every
unconfigured role/tier/runner keeps falling back to the built-in Anthropic
defaults. Like the rest of the directive system it is **advisory**: the hook
annotates the delegate directive; the orchestrator complies. No evidence-schema
change, no gate at `centinela complete`.

## Scope

### In (v1)
- `internal/config/`:
  - New `OrchestrationConfig.ModelMap map[string]map[string]string`
    (`toml:"model_map"`) — tier → runner → concrete model string.
  - `OrchestrationConfig.Models` evolves from `map[string]string` to a
    **union-typed** form per role: a plain tier **string** (back-compat) OR a
    `map[string]string` (runner → concrete model). Implemented with a custom
    TOML unmarshal on a dedicated wrapper type.
  - Shape validation: unknown runner key, unknown role key, unknown tier,
    empty model string all fail at load with an error naming the offending key.
    Normalize (trim + lowercase) tier names and runner keys before validating.
  - Local string sets (`allowedRunnerKeys`) added beside the existing
    `allowedModelTiers` / `allowedModelRoles`; parity test keeps them in sync
    with the domain's allowed sets (config leaf may not import orchestration).
- `internal/orchestration/`:
  - `RunnerCodex Runner = "codex"` enum value.
  - 4-step precedence resolver. Signature evolves to
    `ResolveModel(role, models, modelMap, runner) (modelID string, ok bool)`
    where `models` carries the union value and `modelMap` is tier→runner→model.
  - `tierModels` table gains a `RunnerCodex` column (values empty until
    `codex-support`; a missing codex entry falls through to rule 4).
  - `ModelReference` (and/or per-role annotation) updated to render all three
    runners' IDs so the directive is runner-agnostic.
- `cmd/centinela/hook_orchestration.go` (thin): pass `model_map` + union models
  through to the resolver; emit the all-runners reference line; update the
  `ModelReference` line wording from "claude / opencode" to include codex.
- Docs + Gherkin spec covering all 7 acceptance criteria; config-reference docs
  for the two tables.

### Out (v1, inherited boundaries)
- No `CENTINELA_RUNNER` / `--runner` signal; no `internal/setup/` changes; no
  edits to `.claude/settings.json` or the opencode plugin for runner detection.
- Out-of-band roles (gatekeeper, production-readiness, edge-case-tester,
  merge-steward) are still NOT emitted by the directive hook.
- No recording of the model actually used in role evidence; no validation of
  the model at `centinela complete`.
- No provider-availability / model-existence checks (the runner's job; strings
  are opaque to Centinela — shape-validated only).
- Codex's concrete default model IDs (filled by `codex-support`, Phase 8). This
  feature ships the codex **column and runner key**, not the values.

## Precedence (role → concrete model, for a given runner)

1. **Role override** — `[orchestration.models].<role>.<runner>` present → use it.
2. **Role → tier** (explicit `[orchestration.models].<role> = "<tier>"` or the
   built-in default tier for the role) **→ tier-map override**
   `[orchestration.model_map].<tier>.<runner>` → use it.
3. **Built-in tier→model default** for `<runner>` (today's Anthropic table).
4. Missing mapping for the active runner → return the **tier name** with
   `ok=false` so the caller can warn. Never crash; never emit another runner's
   concrete ID.

## Dependencies & assumptions

- Builds on `configurable-subagent-models`: `Tier`, `Role`, `DefaultTierForRole`,
  `NormalizeTier`, `AllowedTiers`, `AllowedRoleSlugs`, `RequiredRolesForFeature`,
  the `Runner` enum, `tierModels`, and `ResolveModel`/`ModelReference` already
  exist and are unit-tested.
- `internal/config/` is a leaf and MUST NOT import `internal/orchestration/`.
  The runner/tier/role string sets are duplicated locally and reconciled by a
  cross-package parity test (existing pattern in `orchestration_models_test.go`).
- TOML library in use supports a custom `UnmarshalTOML`/`UnmarshalText`-style
  hook on a wrapper type for the union field (verify the exact interface against
  the loader during code; the decoder is already wired in `config.Load`).
- The hook remains runner-agnostic at emit time (locked decision #1); the
  resolver takes `runner` as a parameter and the reference line enumerates all
  runners, so no runtime runner signal is required.
- Advisory model holds: a bad/opaque model string degrades the directive, it
  does not crash the hook.
- Hard rule: every NEW source file (and `_test.go` in `internal/`/`cmd/`) ≤100
  lines. The resolver and config validators are split accordingly (see Rollout).

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Union-typed `models` (string \| table) breaks existing plain-string configs | High | Medium | Custom unmarshal accepting both forms; back-compat acceptance test (AC#4) over the plain-string form; keep `OrchestrationModels` accessor stable for the tier path. |
| Resolver/config files exceed the 100-line hard rule (4-step precedence + union unmarshal + validation + parity) | Medium | High | Split: resolver precedence in its own file, runner enum/table edits in `resolve.go`/`models.go`; config union type + custom unmarshal in a new file separate from the validator; keep each `_test.go` ≤100 lines (split by concern). |
| Opaque model strings → typo points at a non-existent model | Medium | Medium | Shape-only validation (non-empty, known runner/tier/role); document availability as the runner's responsibility; advisory directive degrades, not crashes. |
| Codex column empty before `codex-support` → wrong-vendor ID leaks under codex | Low | High | Rule 4 fallback returns the tier name + `ok=false`; never emit a claude/opencode ID for codex; covered by AC#7. |
| All-runners reference line grows noisier with a third runner | Low | Medium | Keep one compact reference line factored by tier (existing `ModelReference` shape), append codex column; dedupe tiers in stable order. |
| Config leaf drifts from domain allowed-sets (new runner key) | Medium | Medium | Parity test extended to cover `allowedRunnerKeys` ↔ domain runner set; fails loudly on drift. |
| Advisory only — orchestrator may ignore the model hint | Medium | Medium | Accept by design (consistent with the directive model); revisit an evidence-check later. |
| Scaffold mirror drift (resolver/docs mirrored under `internal/scaffold/assets`) | Low | Medium | If any edited architecture doc is mirrored, update the mirror; not expected for source files, but check before validate. |

## Rollout (smallest correct slice first)

- **Slice 1 — codex runner key + codex column (data-only, no behavior change).**
  Add `RunnerCodex` to the enum and an (empty) codex entry to `tierModels`.
  Update `ModelReference` to render the codex column. Update
  `hook_orchestration.go`'s reference wording to include codex. Unit-test that
  codex resolves to `ok=false` (rule 4) and the reference line lists three
  runners. Smallest correct slice: ships the runner-agnostic three-runner line
  with zero new config surface; everything else still defaults.

- **Slice 2 — tier remap via `[orchestration.model_map]` (base layer).**
  Add `ModelMap` to `OrchestrationConfig` + accessor. Validate shape (known
  tier, known runner key, non-empty model; normalize first). Thread `modelMap`
  into `ResolveModel` so rule 2's tier-map override and rule 3's default are
  honored per runner. Acceptance: AC#1 (opencode reasoning → Kimi), AC#3
  (tier override but no map entry → built-in default), AC#6 (absent tables →
  all defaults), AC#7 (no mapping → tier name + warn).

- **Slice 3 — role override via union-typed `[orchestration.models]` (override
  layer).** Introduce the union wrapper type + custom TOML unmarshal accepting
  string OR runner→model table; validate the table form (known runner, non-empty
  model). Make `ResolveModel` apply rule 1 (role override beats its tier) ahead
  of slice 2's logic. Acceptance: AC#2 (senior-engineer pinned to
  deepseek-coder beats its tier), AC#4 (plain-string back-compat), AC#5
  (unknown runner/role/tier/empty model fail at load naming the key).

- **Slice 4 — docs + spec + parity.** Gherkin `.feature` for all 7 AC; config
  reference for both tables with the precedence rules and the codex caveat;
  extend the parity test to the new runner-key set. Wire acceptance execution
  into `validate.commands` (tests step).

### File-split guidance (≤100 lines each, tests included)
- `internal/orchestration/resolve.go` — keep runner enum + `tierModels` +
  `ModelReference`; move the 4-step precedence body into a small helper file if
  `resolve.go` would exceed 100 lines.
- `internal/config/orchestration.go` — struct + accessors only.
- `internal/config/orchestration_models.go` — keep the existing tier validator;
  add the union type + custom unmarshal and the `model_map`/role-table validator
  in a **separate** new file (e.g. `orchestration_model_map.go`) to stay ≤100.
- Split `_test.go` files by concern (precedence vs. validation vs. parity vs.
  reference rendering) so none exceeds 100 lines.

## Handoff to feature-specialist

- Confirm the exact TOML decoder interface for the union field against
  `config.Load` (custom `UnmarshalTOML` on a wrapper vs. decoding into
  `map[string]any` then narrowing) and pick the one that keeps the file ≤100
  lines and preserves the plain-string back-compat path.
- Decide the final `ResolveModel` signature/shape for the union `models`
  argument (e.g. a typed `RoleModels` carrying both the optional tier string and
  the optional runner→model map) vs. passing two parallel maps.
- Author the Gherkin scenarios 1:1 with the 7 acceptance criteria, including the
  codex rule-4 fallback and the mixed-forms edge case.
- Decide whether the per-role annotation on the delegate line shows the resolved
  ID per runner inline, or relies solely on the factored `ModelReference` line
  (locked decision #1 favors the factored line; confirm directive readability).
