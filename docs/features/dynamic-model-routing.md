# Feature: dynamic-model-routing

**Phase:** 13 — Lighter Centinela
**Archetype:** canonical
**Depends on:** none

## Problem

Model routing is static: `[orchestration.models]` maps each role to a tier
(`reasoning`/`balanced`/`fast`) or per-runner override, project-wide. Every
feature pays the same model cost regardless of complexity — a trivial config
rename spawns the same reasoning-tier senior-engineer and qa-senior as a
cross-package redesign. The orchestrator, which sees the task's actual
complexity at start time, has no sanctioned way to route a role down (or up)
for one feature. Because Centinela's gates are machine-run and
model-independent, a wrong downgrade costs a bounced step, never a shipped
defect — routing is a pure cost knob the framework currently doesn't expose.

## Expected outcome

1. **Static stays the default, byte-for-byte.** New key
   `[orchestration] routing_mode = "static" | "dynamic"`, default `static`.
   With `static` (or the key absent), behavior and all emitted directives are
   exactly today's.
2. **Dynamic mode hands the decision to the orchestrator.** At `centinela
   start` (and in step directives while unset), hooks emit a routing
   directive: the step's roles, allowed tiers, configured floors, and the
   static defaults as reference — instructing the orchestrator to decide and
   record before delegating.
3. **`centinela route` command group** (dynamic mode only):
   - `route set <feature> <role> <tier> [--reason "..."]` — persists the
     choice into the feature's workflow state. Refused: unknown role/tier,
     tier below the role's floor (error names the floor), downgrade below the
     static default without `--reason`, and any downgrade for a step already
     underway or completed. Upgrades are allowed anytime.
   - `route show <feature>` — effective table: role, routed tier (or static
     fallback), floor, reason, decidedAt.
4. **Un-routed roles fall back to static config** — partial decisions are
   always safe; model resolution consults the workflow's routes first, then
   `[orchestration.models]`, then built-in defaults (existing chain).
5. **Configurable floors, strict defaults.** `[orchestration.floors]`
   role → minimum tier; shipped defaults `gatekeeper = "reasoning"`,
   `planner = "balanced"`; other roles floorless. Floors are validated at
   config load (unknown role/tier rejected) and enforced at `route set`.
6. **Every route decision is audited.** Tier, previous effective tier,
   reason, timestamp recorded in telemetry alongside the existing per-step
   cost samples, so insights can attribute spend to routing decisions.

## Out of scope

- Heuristic/deterministic `route suggest` and archetype profile tables — in
  dynamic mode the orchestrator is the suggester.
- Calibration feedback (correlating bounces with tiers to tune floors) —
  future layer on top of the recorded telemetry.
- Changing which models the tiers map to per runner (existing
  `[orchestration.models]` override form already covers that).
- Auto-escalation on gate bounce — the orchestrator may route up manually;
  nothing automatic.

## Constraints

- No governance weakening: floors default strict; the gatekeeper stays
  reasoning-tier unless a project explicitly lowers its floor in config.
- Workflow-state schema change must not break existing in-flight workflow
  JSON (absent routes field = no routes).
- Hook directive output in static mode must remain byte-identical (token-diet
  just dieted these panels; do not regrow them in static mode).
- 100-line file cap, per-package coverage ≥97% on touched packages, scaffold
  mirror lockstep for any docs/architecture edits.
