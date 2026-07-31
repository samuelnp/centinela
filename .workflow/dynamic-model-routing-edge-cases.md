# Edge Cases: dynamic-model-routing

**Date:** 2026-07-30 · **Step:** tests · **Role:** edge-case-tester
**Method:** scratch binary `go build -o /tmp/centinela-ec ./cmd/centinela` driven against
hand-built temp fixtures (`centinela.toml` + `.workflow/<f>.json`). Every row marked
**VERIFIED** was reproduced against that binary; rows marked **INSPECTED** are
code-reading only (race/atomicity, not cheaply reproducible).

---

## Risk Matrix

| # | Case | Impact | Likelihood | Why / evidence |
|---|------|--------|------------|----------------|
| **E1** | **Floors are bypassed entirely by writing `modelRoutes` straight into `.workflow/<f>.json`** — VERIFIED | **Critical** | **High** | The state file is agent-writable: `workflow.ClassifyFile` returns `TypeOther` for `.workflow/*.json`, and `hookpolicy.EvaluatePrewrite` allows `TypeOther` unconditionally in every step and every profile. Probe: `hook prewrite` on `.workflow/g.json` → **exit 0 (allowed)**, while a code file in the plan step → exit 2. With `{"modelRoutes":{"gatekeeper":{"tier":"fast"}}}` hand-written, `hook orchestration` emitted `delegate to [gatekeeper (model: haiku (claude) …)]` and `route show` displayed `gatekeeper fast routed / floor reasoning` with **no warning**. The identical value via `route set g gatekeeper fast` is refused. Floors are validated **only** in `ValidateRoute`; `ApplyRoutes` (the consumer) never consults `EffectiveFloor`. This is the trust anchor: the adversarial verifier's own model is silently downgraded by the very agents it judges. |
| **E2** | **Legacy plan contract escapes the shipped `planner >= balanced` floor** — VERIFIED | High | Medium | On a workflow without `planContract: planner-v1`, `RequiredEvidenceRoles` returns `big-thinker` + `feature-specialist`. `EffectiveFloor` keys on the role slug, and `defaultFloors` has only `planner`/`gatekeeper` — there is **no reverse alias** (config aliases legacy *keys onto* planner, never planner's floor *onto* legacy roles). Probe: `route set lg big-thinker fast --reason x` → **exit 0**, plan role downgraded to haiku. Same shape at validate: a legacy contract schedules `validation-specialist` and no `gatekeeper` row at all, so the whole validate step is floorless. |
| **E3** | **Floors constrain `route set` only — never the static/default resolution** — VERIFIED | High | Medium | `[orchestration.floors] qa-senior = "reasoning"` + static `balanced` → `route show` renders `qa-senior balanced static reasoning` side by side, unflagged, and the hook delegates at sonnet. Worse for the anchor: `[orchestration.models] gatekeeper = "fast"` (still valid config) → verifier delegated at **haiku** while the directive literally prints the self-contradiction `floors: gatekeeper>=reasoning; static: gatekeeper=fast`. The brief sanctions lowering the floor *in `[orchestration.floors]`*; today a project can lower the effective tier through a **different key** the floor mechanism does not police, while the routing surface keeps advertising a floor it never enforces. |
| **E4** | **A no-op route silently discards the project's per-runner concrete-model override** — VERIFIED | Med-High | High | `ApplyRoutes` replaces the role's `RoleModel` **wholesale**, dropping `Overrides`. Probe: `[orchestration.models] senior-engineer = { claude = "my-pinned-sonnet-4-9", … }`; before → `model: my-pinned-sonnet-4-9 (claude)`; after `route set h senior-engineer reasoning` (same effective tier, success line reads "was reasoning", `route show` unchanged) → `model: opus (claude)`. The directive *invites* this: it prints `static: senior-engineer=reasoning`, so the natural way to silence the un-routed line is to confirm the same tier — which blows away the pin. `route show` renders tiers only, so the loss is invisible on the routing surface. |
| **E5** | **An older binary silently deletes every route on any load→save** — VERIFIED | Medium | Med-High | `Load` → `json.Unmarshal` drops unknown fields; `Save` → `MarshalIndent` of the struct. Probe with a struct lacking `ModelRoutes`: `modelRoutes` present → old-binary re-save → **`modelRoutes: None`**. Any old-binary `complete`/`revise`/`evidence` wipes routes mid-feature; this project routinely runs a stale installed binary against a newer worktree. Usually fail-safe (fallback to static = stronger), but an **upgrade** route (e.g. gatekeeper routed back up over a config-lowered static) is silently reverted → fail-unsafe, and telemetry then disagrees with state. |
| **E6** | **Stale evidence from a prior run closes the sanctioned routing window at step 1** — VERIFIED | Medium | Medium | `RoleStepUnderway` = current step + `os.Stat` on `.workflow/<f>-<role>.{md,json}`. Probe: fresh workflow on `plan`, leftover `.workflow/pr-planner.md` → `route set pr planner balanced --reason …` → *"its \"plan\" step is already underway or completed"*; delete the file → same command succeeds. Reachable by re-starting a feature slug whose artifacts were never cleaned, by a rewind that left a role's stub, and — routinely — because the house convention is to run `evidence init`/`artifact new` **first**, which creates both stubs and therefore shuts the window before any routing decision is made. |
| **E7** | **`currentStep` outside the step order disarms every "underway" refusal** — VERIFIED | Medium | Low-Med | `stepIndexIn` returns `-1` for an unknown step, so `stepIndexIn(step) < -1` is never true and `step != wf.CurrentStep` short-circuits to *not underway*. Probe: `currentStep: "foo"` → `route set k senior-engineer fast --reason …` **succeeds** even though the code step is long completed. Rule 5 (floor) still holds, rule 6 does not. Same reachability as E1, plus any future archetype/rewind that leaves `currentStep` off `OrderedSteps`. |
| **E8** | **`route show` reports a corrupt route as effective; the overlay ignores it** — VERIFIED | Medium | Low | `routeRows` renders `route.Tier` raw. Probe with `{"tier":"ultra"}` → table shows `senior-engineer ultra routed`, while `hook orchestration` for the same workflow emits `model: opus … reasoning` (ApplyRoutes dropped it). `{"tier":""}` renders a **blank** Tier cell with source `routed`. The audit surface and the effective behavior diverge exactly where an operator would go to check. `decidedAt` is likewise printed unvalidated (raw `x`). |
| **E9** | **Floors and models disagree on config key casing** — VERIFIED | Low | Medium | `validateOrchestrationFloors` lowercases the key before the allow-list check; `validateOrchestrationModels` does not. Probe: `[orchestration.floors] Gatekeeper = "FAST"` → accepted **and applied**; `[orchestration.models] Gatekeeper = "fast"` → `unknown role key "Gatekeeper"`. Two validators, one vocabulary, opposite answers. The shipped parity test compares *membership*, not normalization, so it cannot catch this. |
| **E10** | **`route set` + `complete` concurrently = lost update; `Save` is neither locked nor atomic** — INSPECTED | Medium | Low | `workflow.Save` is a bare `os.WriteFile` of the whole struct; the `internal/evidence` lock is not used for state. `route set` does Load → validate → Save with a wide window; a concurrent `complete` (also Load→Save) clobbers the route, or the route clobbers the step advance. Non-atomic write also means a killed process leaves a truncated file — probed: every command for that feature then fails with `invalid workflow file … invalid character ' '` (clean error, no panic, but the feature is bricked until hand-repaired). |
| **E11** | **dynamic→static leaves stale routes invisible and silently re-armable** — VERIFIED | Low-Med | Medium | Flipping to `static` makes the overlay skip (probe: the pinned override is correctly restored — good), but `route show` and `route set` both refuse under static, so there is **no way to see** that routes still sit in state. Flipping back to `dynamic` re-arms every stale route with no confirmation, including routes decided under a floor config that has since been tightened. |
| **E12** | **Floors on roles this project can never schedule are silently inert** — VERIFIED | Low | Medium | `[orchestration.floors] documentation-specialist = "reasoning"` on a non-user-facing feature, or `validation-specialist` on an `adversarial-v1` workflow: config validates, `route show` shows nothing, no warning. An operator believes a floor is in force that can never fire. Mirror case: floors are never re-checked after a route is recorded — a route legal under yesterday's floor stays applied under today's stricter one. |
| **E13** | Upgrading a role whose step is already completed is accepted — VERIFIED | Low | Medium | Probe at the `validate` step: `route set k senior-engineer reasoning` → exit 0, state written, telemetry event emitted — for a step that ran days ago. Nothing consumes it; it pollutes cost attribution. Rule 6 guards downgrades only, and rule 4 asks "scheduled", not "still ahead". |
| **E14** | `route show` succeeds on a `done` workflow while `route set` refuses — VERIFIED | Low | Low | `route show dn` renders the full static table, exit 0. Harmless but inconsistent with rule 2's "routing decisions are closed". |
| **E15** | Whitespace-only `--reason` stored raw on a non-downgrade — VERIFIED | Low | Low | Rule 7 trims correctly (`--reason "   "` and `--reason ""` both refused on a downgrade below static). But on an upgrade the raw `"   "` is persisted into `modelRoutes[…].reason` and renders as `—`. Cosmetic; the audit line carries a blank reason. |
| **E16** | `route set` calls `workflow.Save` directly, bypassing the `saveWorkflow` test seam — INSPECTED | Low | Low | `complete.go` defines `var saveWorkflow = workflow.Save` and `revise.go` uses it; `route_set.go` does not. The save-failure branch of `runRouteSet` is therefore unstubbable from `cmd/centinela` tests. |

### Verified-correct (probed, no action needed)

Role/tier are case- and whitespace-insensitive (`SENIOR-ENGINEER`, `"  fast  "` accepted).
`routing_mode` accepts `Dynamic`/`DYNAMIC`/`" dynamic "`; a typo (`dinamyc`) is rejected at
config load naming the allowed values. Unknown role, unknown tier, unknown feature, `done`
workflow, and static-mode refusals all exit 1 with the message the spec names. Hotfix
archetype (`stepOrder: [code, validate]`) scopes the scheduled-role set correctly —
`planner`/`qa-senior` refused as unscheduled, so a role's step index is never undefined for
a well-formed subset. Legacy JSON without `modelRoutes` loads to a nil map and `route show`
is all-static; `omitempty` keeps the key absent after a round-trip. Repeat `route set` with
the same tier is idempotent in state, and telemetry `prevTier` is accurate across repeats,
upgrades, and upgrade→downgrade-back (`reasoning→balanced→balanced→reasoning→balanced` all
recorded exactly). Corrupt/truncated state yields a clean error, never a panic.

---

## Missing or Weak Scenarios

Against the shipped colocated tests (`route_{set,show,refusals,edges}_test.go`,
`hook_orchestration_dynamic_test.go`, `model_routes*_test.go`, `route_*_test.go`):

1. **No test asserts the overlay respects floors.** `ApplyRoutes` is tested for
   wholesale-replace, non-mutation, and corrupt-tier tolerance — never for a route that
   `route set` would have refused (E1). The refusal matrix is exercised only through the
   command, i.e. only on the path that already enforces it.
2. **No legacy-contract routing test.** `TestRunRouteShow_LegacyWorkflowIsAllStatic` covers
   rendering, not the floor gap on `big-thinker`/`feature-specialist`/`validation-specialist` (E2).
3. **No floor-vs-static-default test.** Nothing pins what should happen when the floor is
   above the tier a role resolves to without a route (E3).
4. **No override-preservation test.** `TestApplyRoutes_ReplacesWholesaleWithoutMutating`
   asserts the *opposite* of E4 without asking whether a same-tier route should clobber a
   per-runner pin.
5. **No new→old schema-compat test.** Back-compat is tested old→new only; the field-drop
   direction (E5) is untested.
6. **`RoleStepUnderway` has no stale-evidence or out-of-order-`currentStep` case** (E6, E7);
   `model_routes_underway_test.go` covers past/current/done/unscheduled only.
7. **No `route show` fidelity test** — nothing asserts the table agrees with what the hook
   actually emits (E8).
8. **Casing parity between `[orchestration.floors]` and `[orchestration.models]` is untested**
   (E9); the existing parity test checks membership only.
9. **No concurrency or atomicity test** for workflow-state writes (E10).
10. **No dynamic→static→dynamic round-trip test** (E11).

---

## Proposed / Added Tests

### Unit (`tests/unit/`)

| Test | Asserts |
|------|---------|
| `TestApplyRoutes_RefusesRouteBelowFloor` | `ApplyRoutes` (or a new floors-aware sibling) drops/clamps a `gatekeeper: fast` route when the effective floor is `reasoning` — E1, the single highest-value test in this list. |
| `TestEffectiveFloor_LegacyPlanRolesInheritPlannerFloor` | `EffectiveFloor(RoleBigThinker, nil)` and `RoleFeatureSpecial` return `balanced, true` — E2. |
| `TestApplyRoutes_PreservesPerRunnerOverride` | A route whose tier equals the role's effective tier keeps `RoleModel.Overrides` intact — E4. |
| `TestRoleStepUnderway_OutOfOrderCurrentStep` | `currentStep: "foo"` does **not** report every role as freely routable — E7. |
| `TestRoleStepUnderway_StaleEvidenceFromPriorRun` | Pins the intended semantics when a stub predates the current run (e.g. compare stub mtime against `wf.StartedAt`) — E6. |
| `TestFloorsAndModels_KeyCasingParity` | Every `AllowedRoleSlugs()` entry, plus its `Title`/`UPPER` variants, is accepted-or-rejected **identically** by `[orchestration.floors]` and `[orchestration.models]` — E9. Extends the pattern in `configurable_subagent_models_config_unit_test.go`. |
| `TestRouteRows_CorruptTierRendersAsUnknown` | A route with `tier: "ultra"` or `""` is not rendered as an effective routed tier — E8. |

### Integration (`tests/integration/`)

| Test | Asserts |
|------|---------|
| `TestRouting_HandWrittenRouteCannotWeakenGatekeeper` | Write `modelRoutes.gatekeeper = fast` directly to `.workflow/<f>.json`, run the orchestration path, assert the emitted gatekeeper model is still reasoning-tier — E1 end-to-end. |
| `TestRouting_FloorAboveStaticIsSurfaced` | `[orchestration.floors] qa-senior = "reasoning"` with static `balanced` produces a warning (or a config-load refusal) rather than a silent under-floor delegation — E3. |
| `TestRouting_NewToOldSchemaRoundTrip` | Unmarshal→marshal through a struct lacking `ModelRoutes` loses the field; pins the need for a schema/version guard — E5. |
| `TestRouting_DynamicStaticDynamicRoundTrip` | Routes recorded, mode flipped to `static` (correctly ignored), flipped back: assert the re-arm is explicit/visible — E11. |
| `TestRouting_ConcurrentRouteSetAndComplete` | Two processes racing Load→Save on one feature: neither the route nor the step advance is lost — E10. |

### Acceptance (`tests/acceptance/dynamic_model_routing_*_test.go`, built binary via `runCent`)

All carry `// Acceptance: specs/dynamic-model-routing.feature` + `// Scenario:` comments.
Local fixtures only — no network.

| Test | Scenario |
|------|----------|
| `TestDMR_HandWrittenGatekeeperRouteIsNotHonored` | E1 through the real `hook orchestration` + `hook prewrite`. **Priority 1.** |
| `TestDMR_LegacyPlanRoleDowngradeIsRefused` | `route set <f> big-thinker fast` on a pre-`planner-v1` workflow — E2. |
| `TestDMR_SameTierRouteKeepsPinnedModel` | Pinned `senior-engineer.claude`, `route set … reasoning`, hook still emits the pinned id — E4. |
| `TestDMR_StaleEvidenceDoesNotCloseTheRoutingWindow` | E6, with the artifact stub predating `startedAt`. |
| `TestDMR_RouteShowMatchesHookDirective` | For one fixture (including a corrupt tier), the `route show` tier and the hook's `model:` annotation agree — E8. |
| `TestDMR_StaticByteIdentity` (already planned) | Keep — it guards the whole opt-in premise. Extend it to assert byte-identity **with routes present in state**, not only with empty state. |
| `TestDMR_HotfixArchetypeScopesRoles` | Verified-good today; pin it so an archetype change cannot regress the "not scheduled" refusal. |

---

## Residual Risks

- **R1 — The state file is the routing authority and is agent-writable.** Even with E1 fixed
  at the overlay, nothing detects tampering of `modelRoutes` (or `currentStep`, E7). Real
  containment is a state-integrity mechanism (managed-marker/checksum, or classifying
  `.workflow/<feature>.json` as a protected file type in `hookpolicy`) — beyond this
  feature's test scope. **Deferred → `route-state-integrity-guard`.**
- **R2 — Workflow state has no lock and no atomic write.** Pre-existing, but dynamic routing
  adds a second frequent writer to a file `complete`/`revise`/`evidence` already touch,
  raising collision odds. **Deferred → `workflow-state-atomic-save`.**
- **R3 — No schema version on `.workflow/*.json`.** Any field added by a newer binary is
  silently dropped by an older one (E5); routing is merely the first field where the loss is
  behavioural. **Deferred → `workflow-state-schema-version`.**
- **R4 — Floors are a route-time check, not an invariant.** Until floors are evaluated where
  the model is *resolved*, `[orchestration.models]`, built-in defaults, and legacy role sets
  can each land below a declared floor (E2, E3, E12). The proposed tests pin the intended
  invariant; making the code honor it everywhere is a code-step change.
- **R5 — Cost-only blast radius, with one exception.** Per the brief a wrong tier costs a
  bounced step, never a shipped defect — **except** for the gatekeeper, where the tier *is*
  the quality of the adversarial judgment. Every high-impact row above (E1, E2, E3, E5)
  converges on that one role; that is why they are rated above their raw likelihood.
- **R6 — Telemetry noise.** Repeat, no-op, and past-step routes each append a
  `route-decision` event (E13, E15). Cost attribution built on this log must reduce by
  (feature, role, last event) rather than count events.

## Deferred Findings

- `route-state-integrity-guard`
- `workflow-state-atomic-save`
- `workflow-state-schema-version`
