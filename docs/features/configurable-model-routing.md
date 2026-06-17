# Feature: Configurable Model Routing

> Follow-up to [`configurable-subagent-models`](configurable-subagent-models.md),
> which shipped the role→tier knob but **deliberately rejected** concrete /
> cross-provider model IDs (design decision #1, and its deferred-follow-ups list:
> "Raw model-ID override escape hatch"). This feature delivers that follow-up and
> generalizes it across runners.

## Problem

Today Centinela's model selection is two fixed layers:

1. **role → tier** — user-configurable in `[orchestration.models]` (shipped).
2. **tier → concrete model** — a **hardcoded, Anthropic-only** table in
   `internal/orchestration/resolve.go:21`
   (`reasoning=claude-opus-4-7`, `balanced=claude-sonnet-4-6`, `fast=claude-haiku`).

The second layer cannot be changed without editing Go. So a user running under
**opencode** (or, soon, **codex**) — where multiple providers are first-class —
**cannot** point Centinela at the models they actually want: Kimi (Moonshot) for
reasoning, DeepSeek for coding, etc. The semantic tiers are useful for cognitive
load, but they are welded to one vendor.

Two distinct needs are unmet:

- **Vendor substitution** — "use *my* reasoning model for the reasoning tier,
  whatever provider it's on." A tier remap.
- **Per-role precision** — big-thinker and senior-engineer both default to the
  `reasoning` tier, so a tier remap alone can't give *coding* its own model
  without a per-role override. "DeepSeek-Coder for senior-engineer, Kimi for
  big-thinker" requires a direct role→model binding.

**Who is hurting:** operators on opencode/codex (and Claude Code users routing
through a gateway like OpenRouter) who want to choose price/quality/latency per
role across providers, and who today are stuck on the built-in Anthropic IDs.

Like the rest of the directive system, this is **advisory**: Centinela injects
the resolved model into the delegate directive; the orchestrator (host session)
complies. It is not hard-enforced (consistent with `configurable-subagent-models`).

## Outcome

A project can override, in `centinela.toml`, **both**:

- which concrete model backs each **tier**, per runner (base layer), and
- which concrete model a specific **role** uses, per runner (override layer),

while every unconfigured role/tier/runner keeps falling back to the built-in
Anthropic defaults. The orchestration hook resolves each role for the active
runner and annotates the delegate directive with the concrete model ID.

```toml
# Base layer — remap which concrete model backs each tier, per runner.
# Runner keys: claude | opencode | codex. Omitted runners keep the built-in default.
[orchestration.model_map.reasoning]
opencode = "moonshotai/kimi-k2"
codex    = "o3"

[orchestration.model_map.balanced]
opencode = "deepseek/deepseek-chat"

# Override layer — pin a specific role to a concrete model (wins over its tier).
# A role value may still be a tier name (back-compat with configurable-subagent-models).
[orchestration.models]
senior-engineer = { opencode = "deepseek/deepseek-coder" }
qa-senior       = "balanced"   # still accepts a plain tier string
```

Resulting directive under the opencode runner (illustrative):

```
CENTINELA DIRECTIVE: orchestrator only for "x"/"plan"; delegate to
[big-thinker (model: moonshotai/kimi-k2), feature-specialist (model: deepseek/deepseek-chat)].
```

## Design Decisions (locked with the user)

1. **Both layers** — tier→model remap *and* direct role→model override. The
   override layer wins over the tier layer for the roles it names.
2. **Runner-keyed, all runners equal** — every override is keyed by runner
   (`claude` | `opencode` | `codex`). There is no Anthropic-privileged path; the
   built-in table is just the default values for each runner. Codex is included
   now so the resolver is ready when `codex-support` lands.
3. **Concrete model strings are opaque to Centinela** — Centinela validates the
   *shape* (non-empty, known runner key, known tier/role key) but does **not**
   verify the model exists on the provider. Availability is the runner's job.
   This keeps Centinela from chasing every provider's model catalog.
4. **Back-compat with `[orchestration.models]`** — a role value may be a plain
   tier string (old behavior) **or** a runner→model table (new override). The
   loader accepts both forms; no migration required for existing configs.
5. **Tiers stay** — the semantic tiers remain the zero-config default and the
   churn-shield. This feature opens the previously-welded tier→model table and
   adds a role-level escape hatch; it does not remove tiers.
6. **Advisory** — resolved model appended to the existing delegate directive. No
   evidence-schema change, no validation at `centinela complete`.

### Resolution precedence (role → concrete model, for a given runner)

1. **Role override** `[orchestration.models].<role>.<runner>` — if present, use it.
2. **Role → tier** (explicit `[orchestration.models].<role> = "<tier>"`, or the
   built-in default tier for the role) **→ tier map override**
   `[orchestration.model_map].<tier>.<runner>` — if present, use it.
3. **Built-in tier→model default** for `<runner>` (today's Anthropic table).
4. On any missing mapping for the active runner: emit the **tier name** (not a
   wrong vendor ID) with `ok=false` so the caller can warn. Never crash the hook,
   never emit a Claude-only ID under another runner.

### Default tier → model (built-in, per runner — unchanged base)

| Tier | claude | opencode | codex |
|------|--------|----------|-------|
| reasoning | `claude-opus-4-7` | `anthropic/claude-opus-4-7` | _(default tbd in codex-support)_ |
| balanced | `claude-sonnet-4-6` | `anthropic/claude-sonnet-4-6` | _tbd_ |
| fast | `claude-haiku-4-5-20251001` | `anthropic/claude-haiku-4-5` | _tbd_ |

Codex column is wired but its concrete defaults land with `codex-support`; until
then a missing codex entry falls back to rule 4 (emit tier name + warn).

## User Stories

- As an opencode user, I want to set the reasoning tier to `moonshotai/kimi-k2`
  so all my reasoning-heavy roles use Kimi without editing prompts or Go.
- As a developer, I want to pin `senior-engineer` to `deepseek/deepseek-coder`
  while big-thinker stays on my reasoning model, so coding and architecture can
  use different models even though they share a tier.
- As a multi-runner operator, I want each override keyed by runner so the same
  `centinela.toml` resolves correctly under Claude Code, opencode, and codex.
- As an operator, I want a malformed override (unknown runner key, unknown role,
  unknown tier, empty model) to fail loudly at config-load time so mistakes are
  caught before a run.
- As an existing user, I want my current `big-thinker = "reasoning"` tier config
  to keep working untouched after I upgrade.

## Acceptance Criteria (→ Gherkin in `specs/configurable-model-routing.feature`)

1. Given `[orchestration.model_map.reasoning] opencode = "moonshotai/kimi-k2"`
   and the active runner is opencode, when the plan directive is emitted, then a
   role defaulting to the reasoning tier is annotated `model: moonshotai/kimi-k2`.
2. Given `[orchestration.models] senior-engineer = { opencode = "deepseek/deepseek-coder" }`,
   when the directive is emitted under opencode, then `senior-engineer` is
   annotated with `deepseek/deepseek-coder` (override beats its tier).
3. Given a role with a tier override but **no** matching `model_map` entry for the
   runner, when emitted, then it falls back to the built-in default model for
   that tier+runner.
4. Given a plain string role value (`qa-senior = "balanced"`), when loaded, then
   it is accepted and behaves exactly as in `configurable-subagent-models`
   (back-compat).
5. Given an **unknown runner key** (e.g. `gemini`), an **unknown role**, an
   **unknown tier**, or an **empty model string**, when config loads, then it
   fails with an error naming the offending key.
6. Given an **absent** `[orchestration.model_map]` and `[orchestration.models]`,
   when emitted, then every role resolves to its built-in default (zero-config
   safe, identical to pre-feature behavior).
7. Given the active runner has **no** mapping (override, tier-map, or default)
   for a resolved tier, when emitted, then the directive carries the **tier name**
   plus a warning — never another runner's concrete ID.

## Edge Cases

- Empty `[orchestration.model_map]` / `[orchestration.models]` tables → all
  defaults (same as absent).
- Casing/whitespace on tier names and runner keys (`"Reasoning"`, `" opencode "`)
  → normalized (trim + lowercase), then validated.
- Role value is a runner→model table **and** also names a tier elsewhere → the
  per-role concrete override wins (precedence rule 1).
- A `model_map` entry for a runner that the current run isn't using → ignored for
  this run, still validated for shape at load.
- Mixed forms across roles (one role a tier string, another a runner table) →
  both valid in the same `[orchestration.models]`.
- Codex runner before `codex-support` ships → no codex defaults; falls to rule 4
  (tier name + warn), never a wrong-vendor ID.
- Out-of-band roles (gatekeeper, edge-case-tester, merge-steward) are still not
  emitted by the directive hook in v1 (inherited scope boundary).

## Data Model

No persisted runtime entities. Pure configuration + resolution:

- **`internal/config/` (leaf, no internal imports):**
  - `OrchestrationConfig.ModelMap map[string]map[string]string`
    (`toml:"model_map"`) — tier → runner → concrete model.
  - `OrchestrationConfig.Models` evolves from `map[string]string` to a type that
    accepts **either** a string (tier) **or** a `map[string]string` (runner →
    model) per role. Custom TOML unmarshal handles the union; local string-set
    validation mirrors the existing `orchestration_models.go` validator.
- **`internal/orchestration/` (domain):**
  - Runner enum extended with `RunnerCodex`.
  - Resolver implements the 4-step precedence above:
    `ResolveModel(role, models, modelMap, runner) (modelID string, ok bool)`.
  - The built-in `tierModels` table gains a codex column (values filled by
    `codex-support`).
- **`cmd/centinela/hook_orchestration.go`** stays thin: calls the resolver, joins
  the annotation onto the directive. No decision logic in `cmd/`.

## Integration Points

- **`internal/config/`** — new `model_map` table; union-typed `models`; validation
  for unknown runner/role/tier and empty model strings. Parity test keeps the
  config-leaf string sets in sync with the domain's allowed sets (same pattern as
  `orchestration_models.go`).
- **`internal/orchestration/`** — `RunnerCodex`, precedence resolver, codex column.
- **`cmd/centinela/hook_orchestration.go`** — single emission site; also updates
  the `ModelReference` line (currently "claude / opencode") to include codex.
- **Runner identity at emit time** — the open question from
  `configurable-subagent-models` (the hook has no runtime runner signal) is now
  load-bearing: with per-runner concrete IDs, emitting a runner-agnostic
  reference line (all runners' IDs) is the safe default, OR the runner is injected
  at wiring time (`CENTINELA_RUNNER` / `--runner`) in `internal/setup/` for
  `.claude/settings.json`, the opencode plugin, and the future codex adapter.
  **Big-thinker must resolve this** — it determines whether the directive shows
  one ID or all.
- **`codex-support` (Phase 8)** — provides codex's default model IDs; this feature
  ships the codex *column and runner key* so that work is a data fill, not a
  refactor.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Runner unknown at emit time → wrong concrete ID under the wrong runner | High | High | Default to emitting per-runner reference (all IDs) and let orchestrator pick; OR inject a runner signal at wiring time. Never guess. Big-thinker decides. |
| Union-typed `models` (string \| table) breaks existing configs | High | Medium | Custom unmarshal + back-compat acceptance test (#4) covering the plain-string form. |
| Opaque model strings → typo points at a non-existent model | Medium | Medium | Validate shape only; document that availability is the runner's responsibility; advisory directive means a bad ID degrades, not crashes. |
| Config surface grows (two tables, three runner keys) | Medium | Medium | Strong validation with precise errors; thorough docs + examples; zero-config remains the default. |
| Codex defaults unknown today | Low | High | Ship the column empty; rule 4 fallback; fill in `codex-support`. |
| Advisory only — orchestrator may ignore the model hint | Medium | Medium | Accept by design (consistent with the directive model); revisit evidence-check later. |

## Decomposition

Sized for one feature. If big-thinker finds it too large, the natural split is:

- **`model-routing-tier-remap`** — `[orchestration.model_map]` (base layer) +
  per-runner resolution + codex runner key. Delivers vendor substitution.
- **`model-routing-role-override`** — union-typed `[orchestration.models]`
  (role→concrete model) on top, with back-compat. Delivers per-role precision.

Explicitly **out of scope** for v1 (inherited from `configurable-subagent-models`):

- Model selection for out-of-band agents (gatekeeper, production-readiness,
  edge-case-tester, merge-steward) — different injection point.
- Recording the model actually used in role evidence + validating it at
  `centinela complete`.
- Resolving provider availability / model existence (pushed to the runner).
