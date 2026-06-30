### Big-Thinker Report: local-harness-support
**Date:** 2026-06-30

#### Problem

Three shipped features each generalized one axis of harness/model integration but
none lets a local-model operator actually adopt Centinela: `host-harness-adapters`
generalized the *setup* surface (the `HarnessAdapter` registry), `configurable-
model-routing` made the tier→model table per-runner and opaque, and
`model-capability-profiles` connected a driver model's capability class to the
default enforcement profile. Yet the OpenCode adapter writes `instructions` +
`agent` into `opencode.json` but **never a `provider` block**, so an Ollama /
OpenAI-compatible (llama.cpp / vLLM / LM Studio) user must hand-edit
`opencode.json`; and `builtinModelCapability` only knows the six Anthropic ids, so
a local model id returns `ok=false` from `DefaultProfileForModel` and lands on
`strict` only because `strict` is also the hardcoded fallback — by luck, with no
provenance. Local-model operators (the fastest-growing, most-scaffolding-needing
segment) cannot adopt the strict-profile + deterministic-scaffold story without
undocumented manual surgery. This feature composes the three predecessors into a
first-class local target.

#### Scope

- **In:**
  - A `[orchestration.local]` config block (`provider`, `endpoint`, `model`,
    `api_key_env`) in the config leaf, with shape-only validation (known provider,
    all-or-nothing, non-empty model; opaque strings trimmed, never existence-checked).
  - The declared `[orchestration.local].model` as the **lowest** driver-model
    candidate, so declaring an endpoint auto-keys the workflow.
  - A `LocalDefaultClass` capability hook → `limited` for the declared local
    model, layered as the strictly-lowest tier inside `DefaultProfileForModel`
    (explicit `[orchestration.capabilities]` and the builtin map both still win).
  - A managed `provider` block written into `opencode.json` for `ollama` and
    `openai-compatible` (npm `@ai-sdk/openai-compatible`, the given baseURL),
    idempotent and non-clobbering of a user/foreign provider key.
  - `centinela status` provenance: `Profile strict (local default: <id> → limited
    → strict)`.
  - A hermetic end-to-end acceptance test (local stub backend; no real network).
- **Out:**
  - Codex local wiring (`codex-support`).
  - Any new capability class (`limited` reused).
  - Auto-detecting a running local server / verifying the model exists (runner's job).
  - Model-routing table changes (already per-runner).
  - Pointing **Aider/Claude** harnesses at a local endpoint (NEW discovery —
    deferred as `aider-local-provider-wiring`; the local block wires only
    OpenCode's provider surface).

#### Dependencies & Assumptions

- `host-harness-adapters` (shipped) — the `HarnessAdapter` registry + `BuildSyncPlan`;
  the opencode adapter and `mergeOpenCodeAgents` managed-merge discipline this
  extends.
- `model-capability-profiles` (shipped) — `CapabilityClassFor` /
  `DefaultProfileForModel` precedence, `DriverModelFrom`, `EffectiveProfile` /
  `ResolveStart` / `ProfileProvenance`. The local default must be the lowest tier
  and must NOT regress these.
- `configurable-model-routing` (shipped) — opaque model strings: validate shape
  only, never check existence (decision #3); availability is the runner's job.
- Layer rules (n-tier, PROJECT.md): `internal/config` imports nothing internal
  (leaf); `internal/setup` imports nothing internal (so the
  `LocalConfig→LocalProvider` mapping lives in `cmd/`); `cmd/` is a thin
  orchestrator (no decision logic — G7); every new file ≤100 lines (G1, incl. tests).
- Assumption: OpenCode accepts `@ai-sdk/openai-compatible` with a `baseURL` for both
  Ollama and generic OpenAI-compatible servers (Ollama exposes an OpenAI-compatible
  `/v1` endpoint), so one builder covers both kinds.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Managed provider merge clobbers a user's hand-written/foreign provider | High | Medium | `mergeProvider` owns only the key it writes (add-if-absent, like `mergeOpenCodeAgents`); golden + no-clobber + idempotency tests |
| Regression to capability/profile precedence (back-compat invariant) | High | Medium | Local default gated on a non-empty local block; `CapabilityClassFor` left untouched; the new tier sits below explicit + builtin in all three resolution sites; byte-identical zero-config tests |
| Acceptance test makes a real network call and hangs CI | High | Medium | Hermetic stub backend (`httptest`) or pure file-level assertion; no real `ollama`/network call (acceptance-test-network-hang lesson) |
| Threading `local` through `BuildSyncPlan` bloats the interface / touches all adapters | Medium | Medium | Additive `BuildSyncPlanWithLocal` + an optional `local` field on `openCodeAdapter`; `PlanItems()` signature and other adapters unchanged; `BuildSyncPlan` delegates with nil |
| Provider config surface grows / opaque-string typos | Medium | Medium | Shape-only validation with key-named errors; document availability is the runner's job |
| G1 file-size violations splitting provider builders | Low | Medium | One small file per concern (type, validate, builder, merge); largest ~55 lines |

#### Rollout

- **Step 1 — Config leaf (no behavior change):** `LocalConfig` type under
  `OrchestrationConfig` + `validateLocalConfig` wired into `validateConfig`.
  Smallest correct slice; everything downstream keys off this.
- **Step 2 — Resolution composition:** `DriverModelFrom` gains the local model as
  lowest candidate; `LocalDefaultClass` hook + `DefaultProfileForModel` fallback +
  `ProfileProvenance` note. `EffectiveProfile`/`ResolveStart` get it for free.
- **Step 3 — OpenCode provider wiring:** `buildLocalProvider` + `mergeProvider`,
  threaded via `BuildSyncPlanWithLocal` and the `cmd/` `LocalConfig→LocalProvider`
  mapping; golden + idempotency guards.
- **Step 4 — Status provenance (already from Step 2) + the hermetic end-to-end
  acceptance bar.** Can land last; it is the acceptance gate, not a prerequisite.

If this proves too large for one feature, the natural split is
`local-provider-wiring` (Steps 1+3) and `local-capability-default` (Steps 2+4),
already flagged in the brief.

#### Deferred Findings

- `aider-local-provider-wiring` — the `[orchestration.local]` block wires only
  OpenCode's `provider` surface; pointing the Aider (or Claude) harness at a local
  endpoint is unaddressed and is a separate, lower-priority concern. Recorded via
  `centinela roadmap defer`.

#### Handoff

- Next role: feature-specialist.
- Outstanding questions:
  - Confirm OpenCode's exact `provider` JSON shape (npm package + `options.baseURL`
    + `models`) against a current OpenCode release before the senior-engineer pins
    `buildLocalProvider` — the brief assumes `@ai-sdk/openai-compatible` for both
    kinds.
  - Decide the acceptance-test stub form (httptest server vs. pure file-level
    provider assertion) — both are hermetic; the file-level assertion is the
    cheapest and safest for the "no real network" lesson.
