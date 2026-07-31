# Plan: dynamic-model-routing

**Date:** 2026-07-30 · **Archetype:** canonical · **Handoff:** senior-engineer

Brief: `docs/features/dynamic-model-routing.md` (design decisions there are FIXED).
Spec: `specs/dynamic-model-routing.feature`.

## Resolved open points (rationale in planner report)

| # | Open point | Resolution |
|---|------------|------------|
| R1 | Workflow-state JSON shape | `ModelRoutes map[string]ModelRoute` on `Workflow`, JSON `modelRoutes,omitempty`; `ModelRoute{Tier string; Reason string ,omitempty; DecidedAt string}` (RFC3339 UTC string, mirrors `StepState.CompletedAt`). Map keyed by role slug = one effective route per role; upgrades overwrite in place; history lives in telemetry (append-only), not state. Absent field → nil map → no routes. |
| R2 | "Step underway" for a role | Derive the role's scheduled step by scanning `wf.OrderedSteps()` × `workflow.RequiredEvidenceRoles(feature, step)` (contract-aware; handles archetype subsets/reorder and legacy contracts). Underway = scheduled step index < current index, OR workflow done, OR scheduled step == current step AND either evidence artifact (`orchestration.MarkdownPath/JSONPath`) already exists on disk. Pure status ("in-progress") cannot work: the first step is in-progress from `start`, which would make the sanctioned start-time routing window impossible. Evidence existence = delegation has begun. A role never required by this workflow's steps → `route set` refused ("not scheduled"). |
| R3 | CWD | `route set/show` are CWD-relative like every workflow command (`workflow.FilePath` is relative). Worktree mode ⇒ run from inside the worktree, same as `evidence`/`complete`. No special resolution code. |
| R4 | Dynamic directive shape | ONE extra line, emitted only while ≥1 current-step role is un-routed: `CENTINELA DIRECTIVE: routing (dynamic): unrouted [<roles>]; floors: <role>>=<tier>,…; static: <role>=<tier>,…; decide: centinela route set <feature> <role> <tier> [--reason "…"]`. Floors listed only for current-step roles that have one. Line disappears once all current-step roles are routed (routes already show in the existing `model:` annotations). Static path emits nothing new — code branch, not string branch. |
| R5 | Floors enforcement home | Config leaf validates `[orchestration.floors]` keys/values against its LOCAL `allowedModelRoles`/`allowedModelTiers` sets (config imports nothing internal; parity test extended). Semantics — tier ordering, shipped defaults (gatekeeper=reasoning, planner=balanced), refusal matrix — live in `internal/orchestration` (owns Tier/Role vocabulary already). `cmd/` stays a thin shuffler (G7). |
| R6 | Migration/scaffold surface | No data migration (absent key = static; absent routes = none). Scaffolded `centinela.toml` gains a commented `[orchestration]` routing_mode + floors block. `docs/architecture/workflow-enforcement.md` gets a short routing paragraph — WITH its `internal/scaffold/assets` mirror updated in lockstep. Agent prompts unchanged: routing is the orchestrator's decision, not the subagents'. `docs/guides/configuration-reference.md` (not mirrored) gets the full reference. |

## Refusal matrix (route set) — evaluation order

1. `routing_mode` ≠ dynamic → error naming `[orchestration] routing_mode`.
2. Workflow `done` → refused ("workflow already complete").
3. Unknown role → error listing allowed role slugs. Unknown tier → error listing tiers.
4. Role not scheduled in this workflow's steps (archetype subset, non-user-facing skips ux-ui/docs, merge-steward) → refused.
5. New tier below the role's effective floor → error NAMING the floor.
6. Downgrade (new < current effective tier, where current = existing route else static default) AND role's step underway/completed (R2) → refused. Upgrades always pass this check.
7. New tier below the STATIC default AND `--reason` empty → error demanding `--reason`.

`route show` refuses only rule 1.

## File-by-file plan

Hard caps: ≤100 lines per source file incl. `_test.go` in `cmd/` + `internal/`; per-package coverage ≥95% floor, aim ≥97% via colocated tests written IN the code step (tests/ tier is blocked until the tests step).

### Slice 1 — config surface (inert; nothing reads it yet)

| File | Change | Budget |
|------|--------|--------|
| `internal/config/orchestration.go` (88) | +2 fields on `OrchestrationConfig`: `RoutingMode string \`toml:"routing_mode"\``, `Floors map[string]string \`toml:"floors"\`` | →~91 |
| `internal/config/orchestration_routing_mode.go` (NEW) | `RoutingModeStatic/Dynamic` consts; `DynamicRoutingEnabled(cfg) bool` (trim+lower; ""/absent/"static"→false); `validateRoutingMode` (anything else rejected); `OrchestrationFloors(cfg) map[string]string` (normalized copy; retired plan-role keys alias to planner exactly like `aliasPlanRole` does for models); `validateOrchestrationFloors` (unknown role/tier → error, reusing `allowedModelRoles`/`allowedModelTiers`/`allowedTiersList`) | ~85 |
| `internal/config/file_size_exceptions.go` (69) | wire the two validators into `validateConfig` | →~75 |
| `internal/config/orchestration_routing_mode_test.go` (NEW, split if needed) | absent/static/dynamic/garbage mode; floors valid/unknown-role/unknown-tier/retired-key-alias; accessor defaults | ≤100 |

### Slice 2 — domain + workflow state (inert)

| File | Change | Budget |
|------|--------|--------|
| `internal/orchestration/route_floors.go` (NEW) | `tierRank` (reasoning=3, balanced=2, fast=1); `TierBelow(a, b Tier) bool`; `DefaultFloors()` (gatekeeper→reasoning, planner→balanced); `EffectiveFloor(role, cfgFloors map[string]string) (Tier, bool)` — an explicit config entry REPLACES the default (projects may lower it), else default, else no floor | ~55 |
| `internal/orchestration/route_decision.go` (NEW) | `RouteRequest{Role Role; NewTier, CurrentTier, StaticTier Tier; Floor Tier; HasFloor bool; Reason string; Scheduled, StepUnderway, WorkflowDone bool}`; `ValidateRoute(req) error` = matrix rules 2–7 (pure; no IO) | ~90 |
| `internal/orchestration/route_overlay.go` (NEW) | `ApplyRoutes(models RoleModels, routes map[string]string) RoleModels` — a routed tier replaces the role's entry WHOLESALE (routes-first per brief: beats a config role-override table); invalid tier strings ignored (state corruption never breaks the hook) | ~35 |
| `internal/orchestration/route_directive.go` (NEW) | `RoutingDirective(feature string, roles []Role, routed map[string]string, floors map[string]string, models RoleModels) string` — the R4 one-liner; "" when all roles routed | ~55 |
| `internal/workflow/state.go` (100 — AT CAP) | +`ModelRoutes map[string]ModelRoute \`json:"modelRoutes,omitempty"\`` (+3 lines); RELOCATE `New()` (bottom ~5 lines) to `order.go` (already hosts `NewWithOrder`) to stay ≤100 | →≤100 |
| `internal/workflow/model_routes.go` (NEW) | `ModelRoute` type; `(wf) SetModelRoute(role string, r ModelRoute)` (lazy map init); `(wf) ModelRouteFor(role) (ModelRoute, bool)`; `RouteTiers(wf) map[string]string` nil-safe; `RoleScheduledStep(wf, role) (string, bool)`; `RoleStepUnderway(wf, role) bool` per R2 (os.Stat on evidence paths) | ~85 |
| colocated `_test.go` (NEW, one per file) | incl. back-compat: unmarshal a legacy workflow JSON WITHOUT `modelRoutes` → nil map, Save round-trip adds no key | ≤100 each |

### Slice 3 — telemetry + `route` command group

| File | Change | Budget |
|------|--------|--------|
| `internal/telemetry/event.go` (50) | +`TypeRouteDecision = "route-decision"`; +fields `Role`, `Tier`, `PrevTier` (all `omitempty`; `Reason` field already exists) | →~57 |
| `internal/telemetry/constructors.go` (56) | +`RecordRouteDecision(cfg, feature, role, tier, prevTier, reason, model string)` | →~66 |
| `cmd/centinela/route.go` (NEW) | cobra group wiring only (evidence.go idiom) | ~25 |
| `cmd/centinela/route_set.go` (NEW) | `route set <feature> <role> <tier>` + `--reason`; load cfg+wf; matrix rule 1; delegate to `buildRouteRequest` + `orchestration.ValidateRoute`; persist via `SetModelRoute`+`workflow.Save`; `telemetry.RecordRouteDecision` (model = `resolveEmitModel(wf, cfg)`); success line shows `role → tier (was <prev>)` | ~90 |
| `cmd/centinela/route_request.go` (NEW) | `buildRouteRequest(wf, cfg, role, tier, reason)` — floors via `config.OrchestrationFloors` + `orchestration.EffectiveFloor`; static tier via `orchestration.RoleTier(role, orchestrationRouting(cfg) models)`; current effective = existing route else static; scheduled/underway via workflow helpers | ~60 |
| `cmd/centinela/route_show.go` (NEW) | effective table over the workflow's scheduled roles: role, tier, source (routed/static), floor, reason, decidedAt; matrix rule 1 | ~80 |
| colocated `_test.go` (NEW, split per 100-line cap) | success + every refusal + show table + static-mode refusal | ≤100 each |

### Slice 4 — emission integration (behavior switch)

| File | Change | Budget |
|------|--------|--------|
| `cmd/centinela/hook_orchestration.go` (72) | inside the wf loop, when `config.DynamicRoutingEnabled`: `models = orchestration.ApplyRoutes(models, workflow.RouteTiers(wf))` before `annotateRoles` (per-wf copy — do not mutate the shared map across workflows), and after the model-reference line print `orchestration.RoutingDirective(...)` if non-empty. Static path: ZERO new statements executed | →~85 |
| `cmd/centinela/start_routing.go` (NEW) | `printRoutingDirective(wf, cfg)` called from `runStart` (start.go is at 92 — only the 1-line call goes there); emits the R4 line for the first step's roles in dynamic mode | ~40 |
| `cmd/centinela/start.go` (92) | +1 call | →~93 |
| `cmd/centinela/hook_orchestration_dynamic_test.go` (NEW) | dynamic overlay shows routed model id in `model:` annotation; routing line present while unrouted, absent when all routed; STATIC BYTE-IDENTITY: capture hook output with (a) no `routing_mode`, (b) `routing_mode = "static"` → assert `outA == outB` and neither contains `"routing"` | ≤100 |

### Slice 5 — docs + scaffold (validate/docs steps)

| File | Change |
|------|--------|
| `internal/scaffold/assets/centinela.toml` | commented `[orchestration] routing_mode` + `[orchestration.floors]` example block |
| `docs/architecture/workflow-enforcement.md` + `internal/scaffold/assets/docs/architecture/workflow-enforcement.md` | short "Dynamic model routing" paragraph — BOTH files, lockstep (scaffold-mirror parity) |
| `docs/guides/configuration-reference.md` | routing_mode, floors, `route set/show`, refusal matrix, telemetry event |

## Test strategy

**Colocated (code step)** — every new file above ships a colocated `_test.go` (per-package coverage is per-package; `tests/` tier doesn't move it). Touched packages and targets: `internal/config`, `internal/orchestration`, `internal/workflow`, `internal/telemetry`, `cmd/centinela` — all ≥97%.

**Existing tests that must KEEP PASSING UNTOUCHED** (they are the static byte-identity guard):
- `cmd/centinela/hook_orchestration_routing_test.go`, `hook_orchestration_plan_test.go`, `hook_orchestration_test.go`, `hook_orchestration_contract_test.go`, `hook_orchestration_docs_test.go`, `hook_orchestration_user_facing_test.go`
- `tests/acceptance/token_diet_directive_test.go` (built-binary directive pin)
- `internal/workflow/state_test.go` + friends (`New()` relocation changes no signature)

**Expected breakage: NONE.** If any pinned static-output test fails, the static path was touched — that is a defect, not a test to update.

**tests step (tests/ tier, blocked until then):**
- `tests/unit/dynamic_model_routing_config_unit_test.go` — parity: every `orchestration.AllowedRoleSlugs()` accepted as a floors key, every `AllowedTiers()` as a floors value (extends the configurable-subagent-models parity pattern; new file, so the shipped parity test is not edited during code step).
- `tests/integration/` — route lifecycle against real workflow JSON on disk: set → show → overlay → Save/Load round-trip; legacy JSON without `modelRoutes`.
- `tests/acceptance/dynamic_model_routing_*_test.go` — built binary (token-diet idiom: `runCent`): happy set+show, each refusal exit-code + message, static byte-identity via two `hook orchestration` runs, un-routed fallback. All carry `// Acceptance: specs/dynamic-model-routing.feature` + `// Scenario:` comments (spec-traceability gate). Local repos only — no network (acceptance-hang memory).
- `.workflow/dynamic-model-routing-edge-cases.md`.

## Rollout commits (within the feature branch)

1. `feat(config): [orchestration] routing_mode + floors — validated, inert`
2. `feat(orchestration): route floors/decision/overlay/directive domain + workflow ModelRoutes state`
3. `feat(route): centinela route set/show + route-decision telemetry`
4. `feat(hooks): dynamic-mode routing directive + route overlay in start/orchestration hook`
5. `docs/scaffold: routing_mode + floors reference, scaffold toml, enforcement doc + mirror`

Smallest correct slice first: 1–2 are pure additions with zero behavior change; 3 makes routes persistable; 4 is the only behavior switch and is gated on `routing_mode = "dynamic"`; 5 is prose.

## Dogfood note

The installed binary predates `route`. Build a scratch binary (`go build -o /tmp/... ./cmd/centinela`) to dry-run `route set/show` and the dynamic directive before the validate gate (dogfood memory). This feature's own workflow is NOT routed (config here stays static) — no self-dependency.
