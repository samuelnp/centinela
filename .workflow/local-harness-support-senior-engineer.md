# local-harness-support — senior-engineer

## Files Touched

Four additive, lowest-precedence seams. All source files ≤100 lines.

**Slice 1 — config leaf (`internal/config`, imports nothing internal):**
- `orchestration_local.go` (new, 53) — `LocalConfig` type, `allowedLocalProviders`, `normProvider`, `LocalProviderConfig` accessor.
- `local_validate.go` (new, 38) — `validateLocalConfig`: all-or-nothing, allow-list, non-empty endpoint/model; each error names its key.
- `orchestration.go` (edit, 86) — added `Local LocalConfig \`toml:"local"\``.
- `file_size_exceptions.go` (edit) — wired `validateLocalConfig` into the `validateConfig` chain next to `validateCapabilities`.

**Slice 2 — driver-model candidate + capability default:**
- `driver_model.go` (edit, 27) — `cfg.Orchestration.Local.Model` appended as the lowest driver-model candidate (below `driver_model`).
- `local_capability.go` (new, 25) — `LocalDefaultClass`: returns `(limited,true)` only for the declared local model with no explicit/builtin class; else `("",false)`.
- `capability.go` (edit, 100) — `DefaultProfileForModel` falls through to `LocalDefaultClass` after a `CapabilityClassFor` miss. `CapabilityClassFor` untouched.
- `internal/workflow/profile_provenance.go` (edit, 37) — emits `local default: <id> → limited → strict` after the capability miss, before the strict fallback.

**Slice 3 — OpenCode managed provider wiring (`internal/setup`, imports nothing internal):**
- `opencode_provider.go` (new, 34) — `LocalProvider` type, `buildLocalProvider` (npm `@ai-sdk/openai-compatible`, `options.baseURL`, `options.apiKey={env:...}` for openai-compatible, `models{model:{}}`).
- `opencode_provider_merge.go` (new, 56) — `mergeProvider`: nil → no-op; owns only its own key; foreign same-key block (no managed marker) never overwritten; managed block rewritten only on a real value diff (normalized compare).
- `opencode_config_build.go`, `opencode_config.go`, `sync_hooks.go`, `adapter_opencode.go`, `sync_types.go`, `sync.go` (edits) — threaded `*LocalProvider` through build → plan → apply; `SyncItem.Local` carries it to the apply path.
- `cmd/centinela/local_provider.go` (new, 24) — `localProviderFrom(cfg)` pure mapping `config.LocalConfig → setup.LocalProvider` (keeps `internal/setup` internal-free).
- `cmd/centinela/init_agent.go`, `migrate.go`, `migrate_setup.go`, `hook_migrate.go` (edits) — call `BuildSyncPlanWithLocal(agent, localProviderFrom(cfg))`.

**Slice 4 — status provenance:** delivered by the Slice-2 `ProfileProvenance` edit; the status path already threads `cfg` via `RenderStatusWithConfig` → `ProfileProvenance`. No further change.

## Architecture Compliance

- `internal/config` imports nothing internal (leaf preserved).
- `internal/setup` imports nothing internal — the `LocalConfig → LocalProvider` mapping lives in `cmd/centinela/local_provider.go`.
- `cmd/` stays a thin orchestrator: `localProviderFrom` is a pure mapping with no validation/decision logic.
- Every source file ≤100 lines (largest new: `opencode_provider_merge.go` at 56; `capability.go` trimmed to exactly 100).

## Type-Safety Notes

No `interface{}` decision logic; `LocalProvider`/`LocalConfig` are concrete structs. `mergeProvider` round-trips JSON via typed `map[string]json.RawMessage` and normalizes for value-compare. No dynamic typing shortcuts.

## Trade-Offs

- **Managed marker for idempotency vs no-clobber:** `mergeProvider` embeds a `centinela:managed-version` marker (sibling of npm/options/models, inert to OpenCode) so it can both rewrite its own block on a real endpoint change (AC#7 "real change") AND never overwrite a user's same-key block (edge case). This is stricter than the plan's "key-exists → skip", which would have failed the "rewritten on a real change" scenario.
- **`SyncItem.Local` field:** added so `ApplySync` (which takes only a plan) threads the same provider the plan computed, keeping `ApplySync`'s signature stable.
- **Zero-regression:** `BuildSyncPlan(agent)` delegates to `BuildSyncPlanWithLocal(agent, nil)`; `mergeProvider(nil)`/`LocalDefaultClass(nil)`/empty driver candidate all no-op. Existing `internal/config`, `internal/setup`, `internal/workflow`, `cmd/centinela` test suites pass unchanged.

## Handoff

To **qa-senior** (tests step). Production code compiles (`go build ./...` OK), `go vet` clean on touched packages, existing touched-package tests green. Tests to add: `validateLocalConfig` branches, `DriverModelFrom` local precedence, `LocalDefaultClass`/`DefaultProfileForModel` local fallback, `ProfileProvenance` note, `buildLocalProvider` shape per kind, `mergeProvider` add/idempotent/no-clobber/real-change, and a no-local golden equal to the pre-feature golden (byte-identical). All new `_test.go` files must stay ≤100 lines (G1).
