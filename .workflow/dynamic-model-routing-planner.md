### Planner Report: dynamic-model-routing

**Date:** 2026-07-30

#### Problem

Model routing is static and project-wide: `[orchestration.models]` fixes each role's tier for every feature, so a trivial rename pays the same reasoning-tier senior-engineer and qa-senior as a cross-package redesign. The orchestrator — the only party that sees a task's actual complexity at `centinela start` — has no sanctioned way to route a role down (or up) for one feature. Because every gate is machine-run and model-independent, a wrong downgrade costs a bounced step, never a shipped defect: routing is a pure cost knob the framework doesn't expose. This feature exposes it behind an opt-in `routing_mode = "dynamic"`, with floors, an auditable `route` CLI, and a hook directive that hands the decision to the orchestrator — while static mode stays byte-identical.

#### Scope

**In (v1):**
- `[orchestration] routing_mode = "static"|"dynamic"` (default static, absent = static, anything else = config error at load).
- `[orchestration.floors]` role→minimum-tier; shipped defaults gatekeeper=reasoning, planner=balanced, others floorless; validated at config load; an explicit config entry replaces the default (a project may consciously lower it).
- `centinela route set <feature> <role> <tier> [--reason]` + `centinela route show <feature>` (dynamic mode only), with the full refusal matrix: unknown role/tier, role not scheduled in this workflow's steps, below floor (error names the floor), downgrade below static default without `--reason`, downgrade once the role's step is underway/completed, workflow already done. Upgrades anytime.
- Routes persisted in the feature's workflow state (`modelRoutes`, absent-safe); un-routed roles fall back to the existing static resolution chain unchanged.
- Dynamic-mode routing directive at `centinela start` and in the orchestration hook while current-step roles are un-routed; routed tiers flow into the existing `model:` annotations.
- Telemetry `route-decision` event (role, tier, prevTier, reason, model, timestamp).
- Scaffold toml commented example, configuration-reference guide section, workflow-enforcement doc paragraph + scaffold-assets mirror in lockstep.

**Out (fixed by brief):** `route suggest`/heuristics, archetype profile tables, calibration feedback, auto-escalation on bounce, changing tier→model maps (already covered by `[orchestration.models]`/`model_map`).

#### Dependencies & Assumptions

- Builds directly on shipped configurable-subagent-models / configurable-model-routing / token-diet surfaces: `orchestration.ResolveModel` 4-step precedence, `RoleModels`/`ModelMap`, `ModelReference`, `annotateRoles`.
- `workflow.RequiredEvidenceRoles(feature, step)` is the single contract-aware source of truth for role↔step mapping (legacy contracts, archetype subsets/reorders, user-facing gating) — route validation reuses it; no parallel table.
- The orchestration hook (plus `start`) is the ONLY surface where role→model resolution reaches the orchestrator; overlaying routes there covers all consumption. `resolveEmitModel`/`resolveEmitModelFrom` stamp the DRIVER model on telemetry — unrelated to role routing; routes must NOT touch that chain (route events are stamped with it like every other event).
- Config leaf may not import `internal/*`; it already holds local `allowedModelRoles`/`allowedModelTiers` sets with a cross-package parity test — floors validation reuses those exact sets, so no new vocabulary can drift.
- Telemetry stays a config-only leaf; values resolved at cmd/ emit sites (existing pattern).
- Assumption: this feature's own workflow stays un-routed (repo config remains static), so running under the pre-change installed binary is safe.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Static directive output drifts (token-diet just dieted these panels; tests pin content: `cmd/centinela/hook_orchestration_{routing,plan,contract,docs,user_facing}_test.go`, `tests/acceptance/token_diet_directive_test.go`) | High | Low | Dynamic behavior is a code branch, not a string change: the static path executes zero new statements. All existing hook tests kept verbatim as guards; new acceptance test asserts absent-key vs `"static"` outputs are byte-equal and contain no routing line. Expected test breakage: NONE — a failing pinned test is a defect, not a test to update. |
| Workflow JSON forward/back-compat | Medium | Low | `modelRoutes` is `omitempty`; nil map = no routes; colocated test unmarshals legacy JSON and asserts a Save round-trip adds no key. Known residual: an OLD binary re-saving a routed workflow silently drops `modelRoutes` (unknown field ignored on read, omitted on write) — documented; routes re-settable; telemetry keeps the audit. |
| Config↔domain vocabulary drift (floors keys/tiers vs `AllowedRoleSlugs()`/`AllowedTiers()`) | Medium | Medium | Floors reuse the SAME local config sets as `[orchestration.models]`; a new `tests/unit` parity test sweeps every domain slug/tier through floors validation (extends the existing parity pattern in a NEW file so the shipped test isn't edited during the code step). |
| This feature's own workflow runs under the pre-change binary (no `route` verb, no routing code) | Medium | High | No self-routing; new behavior is opt-in and this repo's config stays static. Dogfood `route` via a scratch `go build` binary before the validate gate; `make install` + `centinela doctor` before the gatekeeper runs. |
| 100-line cap collisions on hot files (`internal/workflow/state.go` is at exactly 100; `cmd/centinela/start.go` at 92) | Medium | High | Pre-planned: relocate `New()` from state.go to order.go; extract `start_routing.go` and `route_request.go` helpers; every new file budgeted ≤~90 lines (see plan). |
| "Underway" proxy (evidence-artifact existence for the current step) misjudges an in-flight subagent that hasn't written yet | Low | Medium | Gates are model-independent — a wrong downgrade bounces, never ships; every decision + reason is audited. Rewound steps keep old evidence on disk ⇒ a downgrade-for-a-cheaper-rerun stays refused, matching the no-auto-escalation intent. |
| Routes consumed nowhere (merge-steward, roles skipped by archetype or non-user-facing gating) | Low | Medium | Refusal rule: `route set` rejects any role not scheduled in THIS workflow's steps — no dangling no-op routes. |

#### Rollout

1. **Config surface** — `routing_mode` + `floors` fields, validation, accessors (inert; nothing reads them).
2. **Domain + state** — `internal/orchestration` route floors/decision/overlay/directive (pure) + `workflow.ModelRoutes` shape and role/step helpers (inert).
3. **`route` CLI + telemetry** — `route set/show`, `route-decision` event (routes persistable; emission unchanged).
4. **Emission integration** — overlay + routing directive in `hook orchestration` and `start`, gated on dynamic mode (the only behavior switch, last).
5. **Docs/scaffold** — scaffold toml example, configuration-reference, workflow-enforcement + assets mirror.

Full file-by-file detail with line budgets, resolved open points (R1–R6), and the ordered refusal matrix: `docs/plans/dynamic-model-routing.md`.

#### Behavior Summary

With `routing_mode` absent or `"static"`, every command and hook emits exactly today's bytes and `centinela route` refuses with a pointer to the config key. With `"dynamic"`: `start` and the orchestration hook append one lean directive line while any current-step role is un-routed — naming the un-routed roles, their floors, static defaults, and the `route set` command — instructing the orchestrator to decide and record before delegating. `route set` persists role→{tier, reason, decidedAt} into the workflow JSON after the refusal matrix passes (floors enforced naming the floor; `--reason` required only below the static default; downgrades refused once the role's step is underway or completed; upgrades anytime); every accepted decision appends a `route-decision` telemetry event carrying the previous effective tier. Model resolution consults routes first, then `[orchestration.models]`, then built-ins — un-routed roles are always safe. `route show` renders the effective table (role, tier, source routed/static, floor, reason, decidedAt). Pre-existing workflow JSON without the field behaves identically to today.

#### Gherkin Scenarios

`specs/dynamic-model-routing.feature` — 13 scenarios, each mapped to executable assertions (built-binary acceptance tests in the token-diet `runCent` idiom plus colocated/unit tests, all carrying `// Acceptance:` + `// Scenario:` traceability):

- **Happy:** downgrade with reason recorded (workflow JSON + telemetry event with prevTier); hook directive reflects routed tier while un-routed roles use static config; upgrade anytime without reason even mid-step; `route show` effective table; routing directive emitted while unset and silent once all current-step roles are routed.
- **Negative:** floor refusal names the floor (gatekeeper→reasoning); downgrade below static default without `--reason` refused, nothing persisted; downgrade refused once the role's step is underway (evidence exists) and for a completed step; unknown role/tier refused listing the allowed vocabularies; unscheduled role refused (ux-ui-specialist on a non-user-facing feature).
- **Static:** absent key vs `routing_mode = "static"` produce byte-identical hook output with no routing line; `route set` refused naming `[orchestration] routing_mode`.
- **Compat:** legacy workflow JSON without `modelRoutes` loads cleanly — all-static `route show`, hooks/`complete` unchanged.

#### UX States

1. `start` (dynamic): existing output + `CENTINELA DIRECTIVE: routing (dynamic): unrouted [planner]; floors: planner>=balanced; static: planner=reasoning; decide: centinela route set <feature> <role> <tier> [--reason "…"]` (one line; floors shown only for roles that have one).
2. Orchestration hook (dynamic, un-routed current-step roles): the same single line after the model-reference line. Once all current-step roles are routed: line absent; routed tiers visible in the existing `model:` annotations.
3. `route set` success: `ui.RenderSuccess` line — `route set: senior-engineer → balanced (was reasoning)`.
4. `route set` refusals: non-zero exit with `fmt.Errorf` per matrix (floor named; allowed sets listed; `--reason` demanded; underway step named; routing_mode pointer in static mode) — the repo's plain-error idiom; no new hardcoded-string surface beyond it (Go CLI; i18n gate handled per existing repo idiom).
5. `route show`: aligned table — Role | Tier | Source | Floor | Reason | Decided — static-fallback rows marked `static`, floorless roles `—`.
6. Static mode: all outputs byte-identical to today; `route *` errors name `[orchestration] routing_mode`.

#### Out-of-Scope

`route suggest` / heuristic or archetype-profile suggestions; calibration feedback correlating bounces with tiers; auto-escalation on gate bounce; per-runner tier→model remapping (already shipped); routing out-of-band roles (merge-steward, edge-case-tester) not scheduled by the workflow's steps; insights views attributing spend to routes (the telemetry records land now; the view is a future layer per brief); any UI beyond the CLI.

#### Deferred Findings

- `scaffold-toml-orchestration-examples` — the scaffolded centinela.toml carries zero `[orchestration]` examples for already-shipped knobs (models, model_map, capabilities, local backend); deferred via `centinela roadmap defer --source dynamic-model-routing/planner`.

#### Handoff

**Next role:** senior-engineer.

Outstanding questions (non-blocking, decide at implementation):
1. `ApplyRoutes` replaces a role's entry wholesale, so a routed tier beats a config role-override table (routes-first per brief) — confirm no consumer depends on override-beats-everything under dynamic mode (none found; all existing routing tests run static).
2. Exact `route show` column formatting and success-line wording — follow `ui` package idiom; keep out of pinned static surfaces.
3. `RoleStepUnderway` treats EITHER evidence artifact (`.md` or `.json`, via `orchestration.MarkdownPath`/`JSONPath`) as "delegation begun" — evidence init writes both stubs first, so either suffices; confirm during implementation.
4. Whether `route show` should append the routing-directive hint when roles are un-routed — nice-to-have only if free within the 100-line budget.
