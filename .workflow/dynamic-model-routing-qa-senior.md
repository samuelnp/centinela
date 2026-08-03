# dynamic-model-routing — qa-senior

**Date:** 2026-07-30 · **Step:** tests · **Handoff:** validation-specialist

This step was FIX + TEST. The edge-case tester reproduced four code defects
against the shipped binary; each is fixed here with a red-then-green test, and
four cheaper findings were fixed alongside. The rest are deferred with slugs.

## Test Inventory

| Tier | File | Covers |
|------|------|--------|
| unit (`tests/unit/`) | `dynamic_model_routing_overlay_unit_test.go` | `TestApplyRoutes_RefusesRouteBelowFloor` (E1), `TestApplyRoutes_HonorsConfigLoweredFloor`, `TestApplyRoutes_PreservesPerRunnerOverride` (E4), `TestEffectiveFloor_LegacyPlanRolesInheritPlannerFloor` (E2), `TestEffectiveFloor_LegacyValidateRoleInheritsGatekeeperFloor` (E2) |
| unit | `dynamic_model_routing_underway_unit_test.go` | `TestRoleStepUnderway_StaleEvidenceFromPriorRun` (E6), `_JSONStubFromThisRunCounts`, `_OutOfOrderCurrentStep` (E7) |
| unit | `dynamic_model_routing_config_unit_test.go` | floors↔domain allow-list parity (roles + tiers), unknown role/tier refusals, `TestFloorsAndModels_KeyCasingParity` (E9), routing_mode accepted values + typo |
| integration (`tests/integration/`) | `dynamic_model_routing_integration_test.go` | lifecycle set→persist→reload→overlay through real state + real config; `TestRouting_HandWrittenRouteCannotWeakenGatekeeper` (E1); `TestRouting_FloorAboveStaticIsLabelledRoutesOnly` (E3); legacy JSON round-trip (`omitempty` keeps the key absent); dynamic→static→dynamic round-trip (E11 semantics pinned) |
| acceptance (`tests/acceptance/`) | `dynamic_model_routing_fixtures_test.go` | binary built once from `./cmd/centinela` into a temp dir (never the installed binary); local temp fixtures only, no git remote, no network |
| acceptance | `dynamic_model_routing_happy_test.go` | spec scenarios 1–5 |
| acceptance | `dynamic_model_routing_refusals_test.go` | spec scenarios 6–11 (+ hotfix archetype role scoping) |
| acceptance | `dynamic_model_routing_static_test.go` | spec scenarios 12–13, incl. byte-identity **with routes present in state** |
| acceptance | `dynamic_model_routing_regressions_test.go` | E1, E2, E3, E4, E6, E7, E8 through the real binary |
| colocated (code-step packages) | `internal/orchestration/route_{overlay,floors,directive}_test.go`, `internal/workflow/model_routes_underway_test.go`, `internal/config/orchestration_floors_test.go`, `internal/ui/render_route_test.go`, `cmd/centinela/route_show_test.go`, `hook_orchestration_dynamic_test.go` | extended for the new branches (per-package coverage is per-package; the `tests/` tier does not move it) |

## Coverage Gaps

**None of the 13 spec scenarios is unasserted.** Every acceptance file carries
`// Acceptance: specs/dynamic-model-routing.feature`, and the set of
`// Scenario: <name>` comments is an exact match for the feature file's 13
scenario names (verified by diffing the two sorted lists — 13 want, 13 have,
zero missing, zero extra).

Three scenarios are asserted with a deliberate scope note:

- Scenario 13 says "`centinela complete f` … behave exactly as before". Running
  `complete` would execute the full gate + suite inside an acceptance test, so
  the assertion is made as: `route show` is all-static, `status` works, and the
  dynamic-mode hook output contains the static-mode output verbatim for the same
  legacy state. The behavioral claim is covered without a nested suite run.
- Scenario 1's telemetry assertion reads `.workflow/telemetry/events.jsonl`
  directly (telemetry is on by default).
- Scenario 11's "not user-facing" precondition is established by writing no
  `docs/features/f.md`, which is what `IsUserFacingFeature` actually reads.

## Acceptance Wiring

`centinela.toml` `[validate] commands` runs:

```
go test ./... -coverprofile=coverage.out
COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh
```

`./...` includes `tests/acceptance`, so the acceptance tier executes inside the
mandated validate run — confirmed by the comment block in `centinela.toml` and
by the profiled run below (45 packages, `tests/acceptance` among them).

## Regression Guards (red → green)

The red state was captured by stashing every source fix (keeping the new tests)
and running the suite against the pre-fix tree. Result: **7 of 7 regression
guards failed; all 15 spec-scenario acceptance tests passed** — the defects were
in behavior the spec never pinned, which is exactly why they shipped.

| # | Defect | Fix | Red evidence | Green |
|---|--------|-----|--------------|-------|
| **E1** | CRITICAL — floors lived only in `ValidateRoute`, the `route set` write path. `.workflow/*.json` is `TypeOther` for `ClassifyFile`, so any agent in any step could hand-write `{"modelRoutes":{"gatekeeper":{"tier":"fast"}}}` and the hook would delegate the adversarial verifier at haiku. | Enforcement moved to the **resolution** path: new `orchestration.HonoredRouteTier(role, tier, floors)` parses the tier AND checks it against `EffectiveFloor`; `ApplyRoutes` now takes `floors` and ignores any route that fails — the same fail-safe posture as the pre-existing unparseable-tier handling (never fatal, never honored). The role key is normalized first, so `"GateKeeper"` cannot launder the attack. `cmd/centinela/hook_orchestration.go` passes the effective floors into the overlay. | `TestDMR_HandWrittenGatekeeperRouteIsNotHonored` failed pre-fix with `delegate to [gatekeeper (model: haiku …)]` | passes; the hook emits `gatekeeper (model: opus (claude)`, `route show` flags the row `ignored`, and the directive re-lists `unrouted [gatekeeper]` |
| **E2** | HIGH — no reverse alias: `big-thinker`/`feature-specialist` inherited no planner floor and a legacy validate contract had no gatekeeper-equivalent floor for `validation-specialist`, leaving both legacy steps floorless. | `floorSuccessor` maps each retired role onto its successor (`big-thinker`/`feature-specialist` → `planner`, `validation-specialist` → `gatekeeper`); `EffectiveFloor` tries the role's own config entry, then the successor's, then the role's default, then the successor's default. Mirrors the D8 alias in `internal/config` in the direction a legacy workflow needs. | `TestDMR_LegacyPlanRoleDowngradeIsRefused` failed pre-fix: `route set lg big-thinker fast` **exited 0** | passes; refused naming the `balanced` floor, `validation-specialist fast` also refused, and a hand-written `big-thinker: fast` no longer reaches the hook |
| **E4** | MED-HIGH — `ApplyRoutes` replaced the role entry wholesale, so a *same-tier* route (a visible no-op in `route show`, and the exact confirmation the directive's `static: senior-engineer=reasoning` invites) silently destroyed a per-runner `[orchestration.models]` pin. | A route whose tier equals `RoleTier(role, models)` is skipped, leaving the static entry — including its `Overrides` table — intact. A route to a **different** tier still replaces wholesale, per the brief's routes-first rule. | `TestDMR_SameTierRouteKeepsPinnedModel` failed pre-fix at the post-route assertion: pin `my-pinned-sonnet-4-9` → `model: opus (claude)` | passes in both directions: same-tier keeps the pin, `→ balanced` correctly moves to `sonnet` |
| **E3** | HIGH — the surface advertised a floor the static path does not enforce: `floors: gatekeeper>=reasoning; static: gatekeeper=fast` is self-contradicting. | **Display only** — no new enforcement over static config (that would break the brief's contract that an explicit `[orchestration.models]` choice is sanctioned). The directive segment is now `floors (routes only):` and the `route show` column header is `Route floor`. | `TestDMR_DirectiveLabelsFloorsAsRouteOnly` failed pre-fix (bare `floors:`) | passes; the line still prints `static: gatekeeper=fast` beside it, now without claiming to govern it |

Also fixed (cheap and low-risk, as authorized):

| # | Fix | Red evidence |
|---|-----|--------------|
| **E6** | `roleEvidenceFromThisRun` compares the stub's mtime against `wf.StartedAt`: a stub from an earlier run of the same slug no longer closes the sanctioned start-time routing window. A workflow with no `startedAt` (legacy JSON) keeps the conservative reading. | `TestRoleStepUnderway_StaleEvidenceFromPriorRun` + `TestDMR_StaleEvidenceDoesNotCloseTheRoutingWindow` both failed pre-fix |
| **E7** | An out-of-order `currentStep` made `stepIndexIn` return `-1` for both operands, so `scheduled < current` was never true and every underway refusal was disarmed at once. An unreadable cursor now reads as **underway** (fail-safe). | `TestRoleStepUnderway_OutOfOrderCurrentStep` + `TestDMR_OutOfOrderCurrentStepStillRefusesDowngrades` failed pre-fix (`route set k senior-engineer fast` exited 0) |
| **E8** | `routeRows` renders the tier the overlay would emit and flags an unhonored route as source `ignored`; `unroutedRoles` treats an unhonored route as still un-routed, so the directive keeps asking. Display and behavior can no longer diverge. | `TestDMR_RouteShowMatchesHookDirective` failed pre-fix (table showed `ultra` as effective while the hook emitted opus) |
| **E9** | `validateOrchestrationFloors` now checks the **raw** role key exactly as `validateOrchestrationModels` does, so `Gatekeeper` is rejected by both instead of accepted-and-applied by one. | `TestFloorsAndModels_KeyCasingParity` failed pre-fix on every `UPPER`/`Title` variant |

**Files changed (all ≤100 lines; G1 scans `internal/`, `cmd/`, not `tests/`):**
`internal/orchestration/{route_overlay,route_floors,route_directive}.go`,
`internal/workflow/model_routes.go` + new `model_routes_underway.go` (the
underway logic was split out to stay under the cap),
`internal/config/orchestration_routing_mode.go`, `internal/ui/render_route.go`,
`cmd/centinela/{hook_orchestration,route_request,route_show}.go`.

**Pinned static-output tests were NOT modified** and pass verbatim:
`hook_orchestration_{routing,plan,contract,docs,user_facing}_test.go`,
`hook_orchestration_test.go`, `tests/acceptance/token_diet_directive_test.go`,
`internal/workflow/state_test.go` and friends. The colocated tests that WERE
edited are this feature's own code-step tests, and only where the fix changed
the intended semantics (overlay signature, directive label, table header).

## Suite + Coverage (one profiled run)

| Command | Result |
|---------|--------|
| `go test ./... -coverprofile=coverage.out` | **3906 pass, 0 fail, 45 packages** (exit 0) |
| `COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh` | **coverage gate passed: 97.2% >= 95.0%** |

Per-package on touched packages: `internal/config` 98.0%, `internal/orchestration`
98.4%, `internal/workflow` 97.2%, `internal/ui` 99.8%, `cmd/centinela` 96.9%
(up from 96.1% at code-step handoff), `internal/telemetry` 87.8% (pre-existing
package level; untouched by this step). Every new/edited function in the routing
path reports 100% except `cmd/centinela/route_request.go:35` (83.3%) and the
pre-existing config-error branch in `hook_orchestration.go:26` (96.6%).

## Deferred Findings

Already recorded by the edge-case tester and still valid (not re-deferred):

- `route-state-integrity-guard` — R1/E1's residue: even with the floor enforced
  at resolution, nothing detects tampering of `modelRoutes` or `currentStep`.
  Real containment is a state-integrity mechanism (checksum/managed marker, or
  classifying `.workflow/<feature>.json` as a protected type in `hookpolicy`).
- `workflow-state-atomic-save` — E10: `workflow.Save` is a bare `os.WriteFile`
  with no lock; routing adds a second frequent writer.
- `workflow-state-schema-version` — E5: an older binary silently drops
  `modelRoutes` on any load→save.

Newly deferred by this step:

```
centinela roadmap defer route-lifecycle-hygiene \
  --summary "Recorded routes are invisible and unmanageable outside the dynamic window: flipping routing_mode to static hides them from route show/set while leaving them re-armable (E11), route show succeeds on a done workflow while route set refuses (E14), and a past-step upgrade is accepted and emits telemetry nothing consumes (E13). Needs a route clear/list surface and a scheduling-aware rule 4." \
  --source dynamic-model-routing/qa-senior
```

Judged cosmetic and **not** deferred: E12 (floors on roles a workflow can never
schedule are inert with no warning), E15 (whitespace-only `--reason` stored raw
on a non-downgrade), E16 (`route_set.go` bypasses the `saveWorkflow` test seam —
the save-failure branch is still covered via a chmod'd state file).

## Docs Debt for the Next Steps

`docs/guides/configuration-reference.md:168` describes `floors` as "Role →
minimum tier enforced by `centinela route set`". After E1 that is incomplete —
floors are now enforced at model resolution too, and the display wording changed
(`floors (routes only):`, `Route floor`). Left to the docs step rather than
edited here, since the tests step must not touch `docs/`.

## Handoff

**Next role:** validation-specialist.

- All 13 spec scenarios asserted at the acceptance tier against a purpose-built
  binary; no test in this suite touches a network URL or a git remote.
- The single mandated profiled run is green and the coverage gate passes; do not
  take my word for it — re-run both commands.
- The highest-value thing to re-verify independently is E1: hand-write
  `{"modelRoutes":{"gatekeeper":{"tier":"fast"}}}` into a temp `.workflow/g.json`
  under `routing_mode = "dynamic"` and confirm `hook orchestration` still emits
  `gatekeeper (model: opus (claude)`.
