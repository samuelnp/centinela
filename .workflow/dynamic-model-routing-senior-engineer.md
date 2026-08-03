# dynamic-model-routing — senior-engineer

**Date:** 2026-07-30 · **Step:** code · **Handoff:** qa-senior

Implemented slices 1–5 of `docs/plans/dynamic-model-routing.md` in order, exactly
as planned. No deviations from the file-by-file plan except two additions noted
under Trade-Offs (a `ui` renderer for the `route show` table, and the optional
`route show` hint that open question 4 left to the line budget).

## Files Touched

| File | Change | Lines |
|------|--------|-------|
| `internal/config/orchestration.go` | +`RoutingMode`, +`Floors` fields on `OrchestrationConfig` | 90 |
| `internal/config/orchestration_routing_mode.go` | NEW — mode consts, `DynamicRoutingEnabled`, `validateRoutingMode`, `OrchestrationFloors`, `aliasPlanFloor`, `validateOrchestrationFloors` | 95 |
| `internal/config/file_size_exceptions.go` | wired both validators into `validateConfig` | 75 |
| `internal/orchestration/route_floors.go` | NEW — `tierRank`, `TierBelow`, `DefaultFloors`, `EffectiveFloor` | 48 |
| `internal/orchestration/route_decision.go` | NEW — `RouteRequest`, `ParseRouteTarget` (rules 2–3), `ValidateRoute` (rules 4–7), `NormalizeRole`, `allowedTiersList` | 88 |
| `internal/orchestration/route_overlay.go` | NEW — `ApplyRoutes` (wholesale replace, non-mutating, corrupt-tier tolerant) | 25 |
| `internal/orchestration/route_directive.go` | NEW — `RoutingDirective` (the R4 one-liner) | 49 |
| `internal/workflow/state.go` | +`ModelRoutes map[string]ModelRoute` `json:"modelRoutes,omitempty"`; `New()` relocated out | 97 |
| `internal/workflow/order.go` | received `New()` (no signature change) | 65 |
| `internal/workflow/model_routes.go` | NEW — `ModelRoute`, `SetModelRoute`, `ModelRouteFor`, `RouteTiers`, `RoleScheduledStep`, `RoleStepUnderway` | 98 |
| `internal/telemetry/event.go` | +`TypeRouteDecision`, +`Role`/`Tier`/`PrevTier` (all `omitempty`) | 54 |
| `internal/telemetry/constructors.go` | +`RecordRouteDecision` | 72 |
| `internal/ui/render_route.go` | NEW — `RouteRow`, `RenderRouteTable` (aligned table + optional hint) | 74 |
| `cmd/centinela/route.go` | NEW — cobra group + `loadRoutingConfig` (refusal rule 1) | 38 |
| `cmd/centinela/route_set.go` | NEW — `route set <feature> <role> <tier> [--reason]` | 65 |
| `cmd/centinela/route_request.go` | NEW — `buildRouteRequest`, `currentEffectiveTier` | 42 |
| `cmd/centinela/route_show.go` | NEW — `route show`, `scheduledRoles`, `routeRows` | 80 |
| `cmd/centinela/start_routing.go` | NEW — `printRoutingDirective` | 29 |
| `cmd/centinela/start.go` | +1 call | 93 |
| `cmd/centinela/hook_orchestration.go` | dynamic overlay + routing directive, both branch-gated | 92 |
| `internal/scaffold/assets/centinela.toml` | commented `[orchestration] routing_mode` + `[orchestration.floors]` block | 121 |
| `docs/architecture/workflow-enforcement.md` | "Dynamic Model Routing (optional)" section | 220 |
| `internal/scaffold/assets/docs/architecture/workflow-enforcement.md` | same section, lockstep — files remain byte-identical | 220 |
| `docs/guides/configuration-reference.md` | `[orchestration]` table rows + full "Dynamic model routing" reference | 331 |

**Colocated tests written in this step** (per-package coverage is per-package):
`internal/config/orchestration_routing_mode_test.go`, `orchestration_floors_test.go`,
`internal/orchestration/route_{floors,decision,overlay,directive}_test.go`,
`internal/workflow/model_routes_test.go`, `model_routes_underway_test.go`,
`internal/telemetry/route_decision_test.go`, `internal/ui/render_route_test.go`,
`cmd/centinela/route_{helper,set,refusals,show,edges}_test.go`,
`cmd/centinela/hook_orchestration_dynamic_test.go`.

## Architecture Compliance

**Layer boundaries (G2, n-tier).**

- `internal/config` still imports **nothing internal**. Floors key/value validation
  reuses the existing LOCAL `allowedModelRoles`/`allowedModelTiers`/`allowedTiersList`
  sets, so the two vocabularies cannot drift; the retired plan-role alias mirrors
  `aliasPlanRole` in a string-valued sibling (`aliasPlanFloor`).
- Ordering, shipped defaults, and the refusal matrix live in `internal/orchestration`
  (a leaf that owns the Tier/Role vocabulary already). No new imports added there —
  the package still imports stdlib only.
- `internal/workflow` (domain) → `internal/orchestration` (leaf) is the existing,
  allowed edge; `model_routes.go` adds no new package dependency.
- `internal/ui` gained `render_route.go` with a **ui-owned** `RouteRow` of plain
  strings, so **no new import edge** was introduced into the presentation layer
  (the PROJECT.md G2 allow-list for `internal/ui` is untouched).
- `cmd/` imports `internal/*` freely (allowed) and stays a thin shuffler.

**G7 (no business logic in the outer layer).** Every decision is a domain call:
`orchestration.ParseRouteTarget` / `ValidateRoute` / `EffectiveFloor` / `RoleTier` /
`RoutingDirective` / `ApplyRoutes`, and `workflow.RoleScheduledStep` /
`RoleStepUnderway` / `RouteTiers`. `route_request.go` and `route_show.go` only
gather and project those results — the same shape as the pre-existing
`orchestrationRouting` shuffler.

**G1 (≤100 lines).** Every added and edited `.go` file — sources and `_test.go`
alike — is ≤100 lines; the largest is `internal/workflow/model_routes.go` at 98.
`state.go` was at exactly the cap: `New()` was relocated to `order.go` (which
already hosts `NewWithOrder`) as planned, leaving state.go at 97 with the new
field. No G1 exception was added.

**Scaffold-mirror lockstep.** `docs/architecture/workflow-enforcement.md` and
`internal/scaffold/assets/docs/architecture/workflow-enforcement.md` were edited
with the identical insert and verified byte-identical by `diff`.

## Type-Safety Notes

- Routing decisions travel as the domain's own named types (`orchestration.Role`,
  `orchestration.Tier`) end to end; the only `string` hops are the JSON/TOML
  boundaries and the workflow-state map key, each converted through a validating
  normalizer (`NormalizeRole`, `NormalizeTier`).
- `RouteRequest` is an explicit struct of resolved facts rather than a bag of
  positional booleans, so a new rule cannot silently reorder arguments.
- `EffectiveFloor` returns `(Tier, bool)` instead of a sentinel zero tier, so
  "no floor" is not confusable with a weak floor.
- `tierRank` maps unknown tiers to rank 0, making `TierBelow` total: a corrupt
  value can never compare as *stronger* than a real one.
- `ModelRoutes` is `omitempty` with a lazily-initialized map; nil is a valid,
  meaningful state (no routes) and needs no migration.
- `ui.RouteRow` carries only pre-rendered strings, so the presentation layer
  cannot re-derive a decision.

## Trade-Offs

### Resolved open questions (from the planner handoff)

**1 — Routed tier beats the config role-override table in `ApplyRoutes`.**
Confirmed, and confirmed safe. `ApplyRoutes` replaces the role's `RoleModel`
*wholesale* (`RoleModel{Tier: …}`), which drops any `[orchestration.models].<role>`
runner→model override for that role. I audited every consumer of `RoleModels`:
`ResolveModel`, `RoleTier`, and `annotateRoles`/`resolvedPerRunner` — all reached
only through `hook orchestration`, `start`, and `route show`, all of which are
either static-mode (never overlaid) or dynamic-mode (overlay intended). No consumer
depends on "override beats everything", and every existing routing test runs
static, so none changed. The precedence is the brief's routes-first rule: a route
is a deliberate per-feature decision and must win over a project-wide default.
The alternative (merging the tier into the existing entry, keeping the override)
would make a route silently a no-op for exactly the roles a project pinned — the
worst possible failure mode for a cost knob. Documented in the `route_overlay.go`
doc comment.

**2 — `route show` table + success-line wording, per the `internal/ui` idiom.**
Rendering was moved out of `cmd/` into `internal/ui/render_route.go`
(`RenderRouteTable`), matching the repo's `Render<Noun>` convention and PROJECT.md
rule 2 ("ui renders; it does not decide"). To avoid widening the `internal/ui` G2
allow-list, the row type `ui.RouteRow` is **owned by ui** and holds plain strings —
so the presentation layer gained no new internal import. Columns are
`Role | Tier | Source | Floor | Reason | Decided`, width-computed for alignment,
routed rows in `StyleGreen`, empty cells rendered `—` (the planner's UX-state 5).
The success line is `ui.RenderSuccess("route set: <role> → <tier> (was <prev>)")`,
matching UX-state 3 verbatim. Refusals stay plain `fmt.Errorf` per the repo's
CLI idiom.

**3 — Either evidence artifact counts as the role's step being underway.**
Confirmed and implemented: `roleEvidenceExists` stats **both**
`orchestration.MarkdownPath` and `JSONPath` and returns true on the first hit.
`centinela evidence init` writes both stubs in one call, so in practice they
appear together; accepting either also covers a hand-written `.md` report and the
`artifact stamp` path. Deliberately conservative: a false "underway" only refuses
a *downgrade* (upgrades are never gated), which is the safe direction.

**4 — Optional routing hint in `route show`.** Included — it fit in budget
(`route_show.go` is 80 lines, `render_route.go` 74). `route show` appends the same
`orchestration.RoutingDirective` line for the **current** step's un-routed roles,
so an operator reading the table immediately sees which decisions are still open
and the exact command to close them. It reuses the domain function, so the hint
can never drift from the hook's wording, and it is omitted entirely once the
current step is fully routed.

### Other trade-offs

- **Refusal rules split across two domain functions.** Rules 2–3 (workflow done;
  unknown role/tier) live in `ParseRouteTarget` because they must run *before* the
  typed `RouteRequest` can be built; rules 4–7 live in `ValidateRoute`. This keeps
  every rule in the domain and preserves the plan's documented evaluation order
  (rule 1 is the config guard in `loadRoutingConfig`). The alternative — a single
  string-taking entry point — would have pushed either untyped fields or a second
  parse into the domain struct.
- **`static:` reference is computed from the pre-overlay models.** The hook passes
  `models`, not `wfModels`, to `RoutingDirective`, so the reference column always
  shows what the decision is measured against. Functionally equivalent today (only
  un-routed roles appear on the line) but it keeps the intent explicit if the line
  ever grows.
- **Per-workflow overlay copy.** `ApplyRoutes` never mutates its input, and the
  hook assigns to a loop-local `wfModels`, so one feature's routes cannot leak into
  another active workflow's directive.
- **Known residual (from the planner risk table, unchanged):** an OLD binary that
  re-saves a routed workflow silently drops `modelRoutes` (unknown field on read,
  omitted on write). Routes are re-settable and telemetry keeps the audit trail.

## Deferred Findings

none — no new out-of-scope discoveries surfaced during implementation. The
planner's `scaffold-toml-orchestration-examples` deferral is already recorded and
remains valid (this feature only adds the routing block, not examples for the
already-shipped `models`/`model_map`/`capabilities`/`local` knobs).

## Verification Performed

| Command | Result |
|---------|--------|
| `go build ./...` | pass |
| `go vet ./...` | pass |
| `go test ./... -run xxxNONE` | pass — **zero test-compile breaks** |
| `go test ./internal/... ./cmd/...` | 2753 pass, 0 fail (42 packages) |
| `go test ./tests/...` | 1101 pass, 0 fail (unit + integration + acceptance) |
| `go test ./...` (full sweep) | exit 0 |
| `go test ./... -coverprofile` → `go tool cover -func` | **97.2%** aggregate (gate: 95%) |
| `./scripts/check-fmt.sh` | pass |
| G1 scan over every touched `.go` | none over 100 lines |

Per-package coverage on touched packages: config 98.0%, orchestration 98.4%,
workflow 97.0%, ui 99.8%, cmd/centinela 96.1%, telemetry 88.0% (pre-existing
package level; every new symbol in it is at 100%).

**Pinned static-output tests were not modified and all pass verbatim:**
`hook_orchestration_{routing,plan,contract,docs,user_facing}_test.go`,
`hook_orchestration_test.go`, `tests/acceptance/token_diet_directive_test.go`,
`internal/workflow/state_test.go` and friends.

### Dogfood (scratch binary `go build -o /tmp/centinela-dmr ./cmd/centinela`)

Static mode — absent key vs explicit `routing_mode = "static"`, `diff` clean:

```
CENTINELA DIRECTIVE: orchestrator only for "f"/"plan"; delegate to [planner (model: opus (claude), …)].
Required evidence before centinela complete f: .workflow/f-planner.md, .workflow/f-planner.json
CENTINELA DIRECTIVE: model reference: reasoning: opus (claude) / anthropic/claude-opus-4-7 (opencode) / reasoning (codex)
BYTE-IDENTICAL: static absent == static explicit

$ centinela route set f qa-senior fast --reason x
dynamic routing is off — set [orchestration] routing_mode = "dynamic" in centinela.toml   (exit 1)
```

Dynamic mode — directive, happy path, refusals, table, overlay:

```
CENTINELA DIRECTIVE: routing (dynamic): unrouted [planner]; floors: planner>=balanced; static: planner=reasoning; decide: centinela route set f <role> <tier> [--reason "…"]

$ route set f senior-engineer balanced --reason "config-only change"
 🛡️👁️  CLI  route set: senior-engineer → balanced (was reasoning)          exit 0
$ route set f gatekeeper balanced --reason "save cost"
tier "balanced" is below the gatekeeper floor "reasoning" — raise the tier or lower the floor in [orchestration.floors]   exit 1
$ route set f qa-senior fast
routing "qa-senior" to "fast" is a downgrade below the static default "balanced" — rerun with --reason "…"   exit 1
$ route set f ux-ui-specialist fast --reason x
role "ux-ui-specialist" is not scheduled in this workflow's steps — nothing would consume the route   exit 1
$ route set f qa-senior turbo --reason x
unknown tier "turbo" (allowed: reasoning, balanced, fast)   exit 1
$ route set f senior-engineer fast --reason "late savings"      # code step, evidence on disk
cannot downgrade "senior-engineer" to "fast": its "code" step is already underway or completed   exit 1
$ route set f senior-engineer reasoning                          # upgrade, no reason, mid-step
 🛡️👁️  CLI  route set: senior-engineer → reasoning (was balanced)          exit 0

$ route show f
Model routing — f
  Role             Tier       Source  Floor      Reason              Decided
  planner          reasoning  static  balanced   —                   —
  senior-engineer  balanced   routed  —          config-only change  2026-07-30T17:39:49Z
  qa-senior        balanced   static  —          —                   —
  gatekeeper       reasoning  static  reasoning  —                   —

routing (dynamic): unrouted [planner]; floors: planner>=balanced; …

.workflow/f.json → "modelRoutes": {"senior-engineer": {"tier":"balanced","reason":"config-only change","decidedAt":"2026-07-30T17:39:49Z"}}
telemetry        → {"type":"route-decision","feature":"f","reason":"config-only change","role":"senior-engineer","tier":"balanced","prevTier":"reasoning"}

# overlay reaches the model annotation, and the routing line disappears once routed:
CENTINELA DIRECTIVE: orchestrator only for "f"/"code"; delegate to [senior-engineer (model: sonnet (claude), model: anthropic/claude-sonnet-4-6 (opencode), model: balanced (codex))].
```

Legacy JSON (no `modelRoutes`, no pinned contracts) — loads cleanly, all-static,
contract-aware role set:

```
$ route show g
  Role                   Tier       Source  Floor  Reason  Decided
  big-thinker            reasoning  static  —      —       —
  feature-specialist     balanced   static  —      —       —
  senior-engineer        reasoning  static  —      —       —
  qa-senior              balanced   static  —      —       —
  validation-specialist  fast       static  —      —       —
```

## Handoff

**Next role:** qa-senior.

**Compile-break inventory: NONE.** `go test ./... -run xxxNONE` is clean; no
exported signature changed (`New()` moved files only). No existing test was edited
or deleted.

**Outstanding TODOs for the tests step** (all `tests/`-tier, blocked during code
and therefore not written here — the plan's Test Strategy section is authoritative):

1. `tests/unit/dynamic_model_routing_config_unit_test.go` — cross-package parity:
   every `orchestration.AllowedRoleSlugs()` accepted as an `[orchestration.floors]`
   key and every `AllowedTiers()` as a value. New file, so the shipped
   configurable-subagent-models parity test is not edited.
2. `tests/integration/` — route lifecycle against real workflow JSON on disk:
   `set → show → overlay → Save/Load` round-trip, plus legacy JSON without
   `modelRoutes`.
3. `tests/acceptance/dynamic_model_routing_*_test.go` — built-binary (`runCent`
   token-diet idiom) coverage of all 13 spec scenarios, each carrying
   `// Acceptance: specs/dynamic-model-routing.feature` + `// Scenario:` for the
   spec-traceability gate. **Local repos only — no network** (acceptance-hang
   memory). The static byte-identity scenario should assert two `hook orchestration`
   runs (absent key vs `"static"`) are byte-equal and contain no `routing` line.
4. `.workflow/dynamic-model-routing-edge-cases.md` — the edge cases are already
   enumerated in this role's evidence JSON `edgeCases` array.

**Watch-outs for qa-senior:**

- `cmd/centinela` package tests share the `routeSetReason` cobra flag global;
  `routeRepo` (in `route_helper_test.go`) resets it via `t.Cleanup` — reuse it.
- The routing hint line contains the substring `unrouted`, so any assertion that
  "no routed row exists" must scope itself to the table, not the whole output.
- `RequiredEvidenceRoles` reads the workflow from **disk**, so a test workflow must
  be `Save`d before `RoleScheduledStep`/`route show` will see its pinned contracts.
