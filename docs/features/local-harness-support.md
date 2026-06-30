# Feature Brief: local-harness-support

## Problem

Centinela's harness and model integration assumes frontier **cloud** harnesses.
`host-harness-adapters` generalized the *setup* surface (a `HarnessAdapter`
registry: claude/opencode/aider), `configurable-model-routing` made the
*tier→concrete-model* table per-runner and opaque-string, and
`model-capability-profiles` connected a *driver model's capability class* to the
default enforcement profile. But none of those let a **local-model** user
actually point Centinela at a local backend and get correct governance by
default:

- The OpenCode adapter writes `instructions` + `agent` into `opencode.json` but
  **never wires a provider**. A user running Ollama or an OpenAI-compatible
  server (llama.cpp, vLLM, LM Studio) must hand-edit `opencode.json` to add the
  provider/baseURL — Centinela's managed setup does not touch it.
- `builtinModelCapability` only knows the six Anthropic ids. A local model id has
  **no capability class**, so `DefaultProfileForModel` returns `ok=false` and the
  capability tier never engages. The local user lands on `strict` only because
  `strict` is also the hardcoded fallback — by luck, not by provenance. There is
  no way to say "this is a small local model → it is `limited` → default to
  `strict`" without hand-mapping every model id in `[orchestration.capabilities]`.

**Who is hurting:** local-model operators — the fastest-growing segment, and the
one that needs physical scaffolding the most — cannot adopt Centinela without
manual, undocumented `opencode.json` surgery and per-model capability mapping.

**Why now:** `host-harness-adapters` + `model-capability-profiles` +
`configurable-model-routing` all shipped; this feature composes them into a
first-class local target. It is the last piece that lets the strict-profile +
deterministic-scaffold governance story actually run on a small local model.

## Acceptance Bar (from roadmap)

A small local model completes a governed feature end-to-end using the `strict`
profile and deterministic scaffolds, with all gates and claim verification
passing.

## User Stories

- As an **Ollama user**, I want `centinela init` to wire OpenCode at my local
  Ollama endpoint, so I do not hand-edit `opencode.json` provider/baseURL.
- As a **llama.cpp / vLLM / LM Studio user**, I want to declare a generic
  OpenAI-compatible endpoint (base URL + model id + optional API-key env var),
  so any OpenAI-compatible local server is a one-block config.
- As a **local-model operator**, I want my declared local driver model to default
  to the `limited` capability → `strict` profile **with explicit provenance**
  (`centinela status` shows the source), so I get maximum scaffolding by
  declaring an endpoint, not by luck or by hand-mapping every model id.
- As an **existing user**, I want a config with no `[orchestration.local]` block
  to resolve **exactly** as today (zero regression to the capability/profile
  precedence and the managed opencode output).

## Acceptance Criteria (→ Gherkin in `specs/local-harness-support.feature`)

1. Given `[orchestration.local] provider = "ollama"`, `endpoint =
   "http://localhost:11434/v1"`, `model = "qwen2.5-coder"`, when the OpenCode
   adapter plans setup, then `opencode.json` gains a managed `provider` block for
   that endpoint and the model is reachable — Claude/Aider files untouched.
2. Given `provider = "openai-compatible"` with an `endpoint` + `model` + optional
   `api_key_env`, when planned, then a generic OpenAI-compatible provider block is
   written (npm `@ai-sdk/openai-compatible`, the given baseURL).
3. Given a declared `[orchestration.local].model` with **no** explicit capability
   class, when the driver model resolves, then it defaults to `limited` → `strict`
   profile, and `centinela status` attributes the profile to the local default.
4. Given an explicit `--profile` or global `enforcement_profile`, when a local
   model is declared, then the explicit source still wins (back-compat invariant
   from `model-capability-profiles` is preserved).
5. Given a config with **no** `[orchestration.local]` block, when anything
   resolves, then behavior is byte-for-byte identical to pre-feature (managed
   opencode output and capability/profile precedence both unchanged).
6. Given a malformed `[orchestration.local]` (unknown `provider`, empty
   `endpoint` when a provider is set, empty `model`), when config loads, then it
   fails loudly naming the offending key.
7. Re-running `init`/`migrate` with the same local config is idempotent (managed
   markers respected); the provider block is only rewritten on real change.
8. **Acceptance bar:** an end-to-end governed run under `strict` with the local
   provider wired passes all gates and claim verification — exercised by a
   hermetic acceptance test (a local bare/stub backend, **no** real network
   call; see the acceptance-test-network-hang lesson).

## Edge Cases

- **Unknown provider value** (not `ollama` / `openai-compatible`) → load error
  listing allowed providers.
- **Endpoint set but no model** (or model but no endpoint) → load error; the
  block is all-or-nothing per provider.
- **Casing/whitespace** on `provider` → normalized (trim + lowercase) then
  validated; the opaque `model`/`endpoint` strings are trimmed only.
- **User already hand-wrote a `provider` in `opencode.json`** → managed merge must
  not clobber an unrelated/user provider; only Centinela's managed block is
  owned (mirror `mergeInstructions` / managed-marker discipline).
- **`api_key_env` names a missing env var** → Centinela does not verify presence
  (availability is the runner's job, per `configurable-model-routing` decision
  #3); it only writes the reference.
- **Local model id also mapped in `[orchestration.capabilities]`** → the explicit
  user mapping wins over the local default (precedence: explicit > local-default).
- **No `[orchestration.local]`** → no provider block emitted, capability tier not
  engaged (existing zero-config path).
- **G1 file size** → new config + adapter code stays ≤100 lines/file; provider
  builders split into their own small files.

## Data Model

No persisted runtime entities. Configuration + resolution only:

- **`internal/config/` (leaf):** a new `LocalConfig` under `OrchestrationConfig`
  (`toml:"local"`): `Provider`, `Endpoint`, `Model`, `APIKeyEnv` (all strings).
  Shape validation (known provider, all-or-nothing, non-empty model). A
  local-default capability hook: when `[orchestration.local].model` is set and
  has no explicit class, treat it as `limited` (lowest precedence, explicit
  `[orchestration.capabilities]` still wins).
- **`internal/setup/` (infrastructure):** a `mergeProvider` step in
  `buildOpenCodeConfig` driven by `LocalConfig`; a small provider-block builder
  per provider kind. The opencode adapter's `PlanItems` gains the provider when a
  local block is present. Managed-marker / idempotent like the rest of sync.

## Integration Points

- `internal/config/orchestration.go` (+ new `orchestration_local.go`) — the
  `LocalConfig` type, parsing, validation, and the local-default capability hook
  that composes with `capability.go`'s precedence.
- `internal/setup/opencode_config_build.go` + a new `opencode_provider.go` —
  wire the managed provider block; adapter `PlanItems` includes it.
- `internal/ui` / `cmd/centinela` status — surface the local driver model and the
  `limited (local default)` profile provenance (extends the existing
  `model-capability-profiles` status line; no new decision logic in `cmd/`).
- **Out of scope:** Codex (separate `codex-support` feature); any new
  capability class; auto-detecting a running local server or verifying the model
  exists (runner's job); model-routing table changes (already per-runner).

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Managed provider merge clobbers a user's hand-written provider | High | Medium | Own only Centinela's managed block via markers; golden + idempotency tests like `mergeInstructions` |
| Regression to capability/profile precedence (back-compat invariant) | High | Medium | Local default is the lowest tier, engaged ONLY when a local model is declared and unmapped; explicit sources still win; back-compat acceptance test |
| Acceptance test does a real network call to a local server and hangs CI | High | Medium | Hermetic test with a local bare/stub backend; no network push (acceptance-test-network-hang lesson) |
| Provider config surface grows / opaque-string typos | Medium | Medium | Shape-only validation with precise errors; document availability is the runner's job |
| G1 file-size violations splitting provider builders | Low | Medium | One small file per concern (config type, validation, provider builder) |

## Decomposition

Sized for one feature, sequenced internally:

1. `LocalConfig` type + validation in the config leaf (no behavior change yet).
2. Local-default capability hook composing with `capability.go` precedence.
3. OpenCode managed provider wiring (`mergeProvider` + provider builder) with
   golden + idempotency guards.
4. Status provenance + the hermetic end-to-end acceptance test (the bar).

If big-thinker finds this too large, the natural split is
`local-provider-wiring` (1+3) and `local-capability-default` (2+4).
