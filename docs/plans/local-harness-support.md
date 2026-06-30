# Implementation Plan: local-harness-support

> Composes `host-harness-adapters` (adapter registry) + `model-capability-profiles`
> (capability→profile precedence) + `configurable-model-routing` (opaque model
> strings) into a first-class **local model** target: wire an OpenCode provider
> block at a local endpoint AND default a declared local driver model to
> `limited → strict` with explicit provenance — with **zero regression** to the
> existing zero-config path.

## Design summary

Four orthogonal seams, each a strictly-additive lowest-precedence tier:

1. **Config leaf** — a new `LocalConfig` under `OrchestrationConfig` (`toml:"local"`),
   shape-validated (known provider, all-or-nothing, non-empty model). Pure leaf,
   imports nothing internal.
2. **Driver-model resolution** — `[orchestration.local].model` becomes the
   **lowest** driver-model candidate (below `[orchestration] driver_model`), so
   declaring an endpoint auto-keys the workflow to that model.
3. **Capability default** — a `LocalDefaultClass` hook returns `limited` for the
   declared local model **only when** it has no explicit/builtin class. It is
   layered as a fallback inside `DefaultProfileForModel`, BELOW
   `CapabilityClassFor` (explicit `[orchestration.capabilities]` + builtin map).
   `CapabilityClassFor` itself is **left untouched** (the precedence primitive).
4. **OpenCode provider wiring** — `buildOpenCodeConfig` gains a `mergeProvider`
   step driven by a setup-local `LocalProvider` value (mapped from `LocalConfig`
   in `cmd/`, so `internal/setup` keeps importing nothing internal). Centinela
   owns only the provider **key it writes**; an existing same-named/foreign
   provider is never clobbered (mirrors `mergeOpenCodeAgents`). Idempotent: a
   real change is the only trigger of `changed=true`.

The hard invariant holds because every new tier is gated on a non-empty
`[orchestration.local]` block: with no local block, `LocalConfig` is the zero
value, `DriverModelFrom` skips the empty candidate, `LocalDefaultClass` returns
`("", false)`, and `mergeProvider(nil)` no-ops — so config resolution and the
managed opencode output are **byte-for-byte** identical to pre-feature.

## File-by-file rollout

### Slice 1 — `LocalConfig` type + validation (config leaf, no behavior change)

| File | Change |
|------|--------|
| `internal/config/orchestration.go` (edit) | Add field `Local LocalConfig \`toml:"local"\`` to `OrchestrationConfig`. |
| `internal/config/orchestration_local.go` (**new**, ~70 lines) | `type LocalConfig struct { Provider, Endpoint, Model, APIKeyEnv string }` (all `toml`-tagged). `allowedLocalProviders = {"ollama","openai-compatible"}`. `func normProvider(s) string` (trim+lower). Accessor `LocalProviderConfig(cfg) (LocalConfig, bool)` returning the trimmed block + whether it is set (provider non-empty). |
| `internal/config/local_validate.go` (**new**, ~45 lines) | `validateLocalConfig(cfg) error`: if provider empty → all fields must be empty (all-or-nothing absent is valid); else provider must be in the allow-list (normalized) else error listing allowed values; `endpoint` non-empty after trim; `model` non-empty after trim; `api_key_env` optional (reference only, never resolved — runner's job). Each error names the offending key. |
| `internal/config/file_size_exceptions.go` (edit) | Add `if err := validateLocalConfig(cfg); err != nil { return err }` into the `validateConfig` chain (next to `validateCapabilities`). |

Validation rules (AC#6, edge cases): unknown provider → `orchestration.local.provider %q unsupported (allowed: ollama, openai-compatible)`; endpoint empty when provider set → `orchestration.local.endpoint must not be empty`; model empty → `orchestration.local.model must not be empty`. Provider is normalized (trim+lower); `endpoint`/`model`/`api_key_env` are opaque → trimmed only, never existence-checked (configurable-model-routing decision #3).

### Slice 2 — Driver-model candidate + local capability default

| File | Change |
|------|--------|
| `internal/config/driver_model.go` (edit) | Append `cfg.Orchestration.Local.Model` as the **last** (lowest-precedence) candidate after `cfg.Orchestration.DriverModel`. Empty → trimmed away → no change to zero-config. |
| `internal/config/local_capability.go` (**new**, ~35 lines) | `func LocalDefaultClass(modelID string, cfg *Config) (string, bool)`: returns `("",false)` when id empty / cfg nil / `Local.Model != id` / the id already has a class via `CapabilityClassFor`; otherwise `(CapabilityLimited, true)`. This makes it the strictly-lowest tier. |
| `internal/config/capability.go` (edit, ~3 lines) | In `DefaultProfileForModel`, after the `CapabilityClassFor` miss, fall through to `LocalDefaultClass`; if it hits, return `ProfileForCapability(CapabilityLimited, cfg), true`. `CapabilityClassFor` stays untouched. |
| `internal/workflow/profile_provenance.go` (edit) | After the `CapabilityClassFor` miss branch, before the strict fallback: if `config.LocalDefaultClass(wf.DriverModel, cfg)` hits → note `fmt.Sprintf("local default: %s → limited → strict", wf.DriverModel)`. |

`EffectiveProfile` and `ResolveStart` need **no edit** — both already route through
`DefaultProfileForModel`, so they pick up the local default for free. AC#4 holds:
explicit `--profile` (tier 1) and global `enforcement_profile` (tier 2) are checked
before the driver-model tier in all three sites, so they still win.

### Slice 3 — OpenCode managed provider wiring

| File | Change |
|------|--------|
| `internal/setup/opencode_provider.go` (**new**, ~55 lines) | `type LocalProvider struct { Provider, Endpoint, Model, APIKeyEnv string }`. `buildLocalProvider(lp LocalProvider) (key string, block map[string]any)`: both kinds use npm `@ai-sdk/openai-compatible`; `options.baseURL = Endpoint`; for `openai-compatible` add `options.apiKey = "{env:"+APIKeyEnv+"}"` when set; `models = { Model: {} }`. key = the provider name (`ollama` / `openai-compatible`). |
| `internal/setup/opencode_provider_merge.go` (**new**, ~35 lines) | `mergeProvider(raw map[string]json.RawMessage, lp *LocalProvider) bool`: nil → false. Unmarshal `raw["provider"]`; if the key already exists → return false (no clobber — owns only its own key). Else marshal the built block under its key, re-marshal `provider`, return true. |
| `internal/setup/opencode_config_build.go` (edit) | `buildOpenCodeConfig(path string, local *LocalProvider)`; call `if mergeProvider(raw, local) { changed = true }` after `mergeOpenCodeAgents`. |
| `internal/setup/opencode_config.go` (edit) | `InjectOpenCodeConfig(path string, local *LocalProvider)` threads `local` into `buildOpenCodeConfig`. |
| `internal/setup/sync_hooks.go` (edit) | `planOpenCodeConfig(local *LocalProvider)` threads it into `buildOpenCodeConfig`. |
| `internal/setup/adapter_opencode.go` (edit) | `openCodeAdapter` gains field `local *LocalProvider`; `PlanItems()` calls `planOpenCodeConfig(a.local)`. Registry keeps the zero-value (`local=nil`). |
| `internal/setup/sync.go` (edit) | Add `BuildSyncPlanWithLocal(agent string, local *LocalProvider) (SyncPlan, error)`; existing `BuildSyncPlan(agent)` delegates with `nil` (byte-identical). For `opencode`/`both`, substitute `openCodeAdapter{local: local}` for the registry's zero adapter; also thread `local` into the `InjectOpenCodeConfig` apply path in `sync.go`. |
| `cmd/centinela/init_agent.go`, `migrate_setup.go`, `migrate.go`, `hook_migrate.go` (edit) | Load cfg (already loaded in most), call `localProviderFrom(cfg)` helper → map `config.LocalConfig`→`setup.LocalProvider` only when set, pass to `BuildSyncPlanWithLocal`. |
| `cmd/centinela/local_provider.go` (**new**, ~25 lines) | `func localProviderFrom(cfg *config.Config) *setup.LocalProvider`: returns nil unless `LocalProviderConfig` is set; thin mapping only (no decision logic — G7 safe). |

Idempotency (AC#7) and no-clobber (edge case) come from `mergeProvider`'s
add-if-absent + value-compare discipline; re-running emits no change once the
key exists. `internal/setup` still imports nothing internal — the
`config.LocalConfig → setup.LocalProvider` mapping lives in `cmd/`.

### Slice 4 — Status provenance + hermetic acceptance test

| File | Change |
|------|--------|
| status surfacing | Already delivered by Slice 2's `ProfileProvenance` edit; `RenderStatusWithConfig` renders `Profile  strict  (local default: <id> → limited → strict)` with no further change. Verify the `cmd/centinela` status path threads `cfg` (it already does via `RenderStatusWithConfig`). |
| acceptance test (Slice handled in tests step) | **Hermetic**: a local stub HTTP backend (`httptest.Server`) standing in for the OpenCode runner endpoint, OR a pure file-level assertion that the wired `opencode.json` provider points at the configured baseURL — **no real network call, no `ollama`/`vLLM` process** (acceptance-test-network-hang lesson). The end-to-end bar asserts a governed run under `strict` with the provider wired passes gates against the stub. |

## Test strategy (planned for the tests step, by qa-senior)

- **Unit (`internal/config`)**: `LocalConfig` parse + every `validateLocalConfig`
  branch (unknown provider, endpoint-without-model, model-without-endpoint,
  casing/whitespace normalization); `DriverModelFrom` local candidate precedence;
  `LocalDefaultClass` (hit, miss-when-explicit-mapped, miss-when-no-local);
  `DefaultProfileForModel` local fallback → strict; **back-compat**: zero-config
  resolution byte-identical (table of existing cases unchanged).
- **Unit (`internal/setup`)**: `buildLocalProvider` shape per kind; `mergeProvider`
  add / idempotent re-add / no-clobber-existing-key; **golden** opencode.json with
  and without a local block — the no-local golden must equal the pre-feature
  golden byte-for-byte (mirror the host-harness golden parity guard).
- **Unit (`internal/workflow`)**: `ProfileProvenance` local-default note.
- **Acceptance (`tests/acceptance`)**: hermetic end-to-end governed run under
  `strict` with the provider wired against a local stub backend; idempotent
  re-`init`/`migrate` (no second write); Claude/Aider files untouched (scope
  assertion). All colocated `_test.go` files stay ≤100 lines (G1 applies to tests).

## G1 / layer compliance

- Every new source file is budgeted ≤100 lines (largest: `opencode_provider.go` ~55).
- `internal/config` imports nothing internal (leaf preserved).
- `internal/setup` imports nothing internal (the `LocalProvider` mapping is in `cmd/`).
- `cmd/centinela` stays a thin orchestrator — `localProviderFrom` is a pure mapping,
  no validation/decision logic (G7 preserved); all rules live in `internal/config`.

## Out of scope (this feature)

- Codex local wiring (`codex-support`).
- Any new capability class (`limited` is reused).
- Auto-detecting a running local server / verifying the model exists (runner's job).
- Model-routing table changes (already per-runner via `configurable-model-routing`).
- Pointing **Aider/Claude** harnesses at a local endpoint — the local block wires
  only OpenCode's provider surface (deferred: `aider-local-provider-wiring`).
</content>
</invoke>
