# Senior-Engineer Report — configurable-subagent-models (code step)

## Scope
Implemented Slices 1–3 of the plan (domain, config leaf, single-emission-site
wiring). No tests written (owned by qa-senior). `centinela complete` not run.

## Files

### Slice 1 — Domain (`internal/orchestration`), pure, no wiring
- **`internal/orchestration/models.go`** (74 lines) — `Tier` type +
  `TierReasoning`/`TierBalanced`/`TierFast`; `defaultTierForRole` map encoding the
  locked role→tier defaults (7 step roles + out-of-band gatekeeper/edge-case-tester/
  merge-steward); `DefaultTierForRole(role)` (falls back to balanced for unknown
  roles); `NormalizeTier(s)` (trim+lowercase then validate); `AllowedTiers()`;
  `AllowedRoleSlugs()` (the 10 slugs config may key on).
- **`internal/orchestration/resolve.go`** (74 lines) — `Runner` type
  (`RunnerClaude`/`RunnerOpenCode`/`RunnerUnknown`); `tierModels` per-runner
  tier→model table (the single place a model refresh edits);
  `ResolveModel(role, models, runner)` resolving role→tier (config override via
  map else default)→model, returning the tier name + `ok=false` on any missing
  mapping (never panics); `ModelReference(tiers)` rendering ONE compact line with
  both runner IDs, deduped and in stable AllowedTiers() order.

### Slice 2 — Config leaf (`internal/config`), no internal imports
- **`internal/config/orchestration.go`** (36 lines) — added
  `Models map[string]string` (`toml:"models"`) to `OrchestrationConfig`; added
  nil-safe `OrchestrationModels(cfg)` accessor.
- **`internal/config/orchestration_models.go`** (49 lines) —
  `validateOrchestrationModels(cfg)`: rejects unknown role keys (error names the
  key) and invalid tiers (error names the key + lists allowed tiers); tiers
  normalized (trim/lowercase) before validation; absent/empty table is valid.
  Uses LOCAL `allowedModelTiers`/`allowedModelRoles` string sets — the leaf does
  NOT import `internal/orchestration`.
- **`internal/config/file_size_exceptions.go`** (39 lines) — `validateConfig`
  now calls `validateOrchestrationModels(cfg)`.

### Slice 3 — Wire the single emission site (`cmd/centinela`), thin
- **`cmd/centinela/hook_orchestration.go`** (47 lines) — loads config via
  `config.Load()`; on error falls back to an empty override map (zero-config-safe,
  never aborts). Per workflow it delegates annotation to `annotateRoles`, then
  prints the existing two lines plus one
  `CENTINELA DIRECTIVE: model reference: <ModelReference(tiers)>`.
- **`cmd/centinela/orchestration_annotate.go`** (27 lines, new thin helper) —
  `annotateRoles` builds annotated names (`<role> (model: <tier>)`), the evidence
  file list, and the deduped tiers in play. Only delegates to `orchestration.*`
  (DefaultTierForRole/NormalizeTier/MarkdownPath/JSONPath) — no decision logic in
  cmd/ (G7).

## Sample emitted directive (plan step, default config)
```
CENTINELA DIRECTIVE: orchestrator only for "payments"/"plan"; delegate to [big-thinker (model: reasoning), feature-specialist (model: balanced)].
Required evidence before centinela complete payments: .workflow/payments-big-thinker.md, .workflow/payments-big-thinker.json, .workflow/payments-feature-specialist.md, .workflow/payments-feature-specialist.json
CENTINELA DIRECTIVE: model reference: reasoning: claude-opus-4-7 (claude) / anthropic/claude-opus-4-7 (opencode); balanced: claude-sonnet-4-6 (claude) / anthropic/claude-sonnet-4-6 (opencode)
```

Runner-agnostic resolution confirmed: `ResolveModel(RoleBigThinker, nil, RunnerOpenCode)` → `anthropic/claude-opus-4-7` (opencode form, not the Claude ID).

## Layers / G1
- **G1**: every touched file ≤100 lines (max 74). The `cmd/` change was split into
  a thin `orchestration_annotate.go` to keep the hook lean.
- **Config is the leaf**: `internal/config` imports nothing internal; allow-lists
  are local string sets (a parity test, owned by qa-senior, will bind them to the
  domain's `AllowedTiers()`/`AllowedRoleSlugs()`).
- **Resolution lives only in `internal/orchestration`** (domain).
- **cmd stays thin (G7)**: only calls `orchestration.*` helpers + the config
  accessor; no decision logic.

## Build / Vet
- `go build ./...` — PASS
- `go vet ./internal/orchestration/ ./internal/config/ ./cmd/centinela/` — PASS
- `go test -run xxxNoMatch ./...` (touched pkgs) — existing tests still compile.

## Edge cases handled
- Missing tier→model mapping → `ResolveModel` returns the tier name + ok=false; no panic.
- Unknown runner (RunnerUnknown) → hook never resolves a single ID; the both-runner
  reference line is always emitted.
- Absent/empty `[orchestration.models]` → all defaults apply (zero-config-safe).
- Casing/whitespace in tier values normalized before validation and resolution.
- Config load error in the hook → falls back to defaults, directive still emitted.

## Handoff
qa-senior — write unit/integration/acceptance tests per the plan's "Files Added".
