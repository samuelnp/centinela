### Big-Thinker Report: workflow-archetypes
**Date:** 2026-06-12

#### Problem
Centinela has exactly one workflow shape (plan → code → tests → validate → docs). That fits a net-new feature but forces bugfixes, refactors, and throwaway spikes through ceremony they don't need (a Gherkin spec for a one-line fix, a docs portal for an internal refactor, a full validate gate for a prototype), pushing people to skip Centinela on exactly the work a light rail would still help. The fix is **archetypes**: named presets that select a *subset/ordering of the existing canonical steps* — reusing the canonical names so every existing mechanism (step-gating matrix, required-role policy, per-step validators, the ship gate) works unchanged. canonical (default), hotfix (code→tests→validate), refactor (plan→code→tests→validate), spike (plan→code, ungated).

#### Scope
- **In:** `internal/workflow/archetype.go` (consts + `ArchetypeStepOrder` + `NormalizeArchetype`/`validateArchetype`); `Workflow.Archetype` state field (state.go); `--archetype` flag + precedence resolver beside `workflowOrderForFeature` (start.go/start_guard.go); fix start.go:84 hardcoded "plan" current-step print; optional `Feature.Archetype` on roadmap with load-validation; status render of archetype + spike "no ship gate" annotation; unit/integration/acceptance tests incl. the safety test.
- **Out:** No new step *names*; no edits to complete.go ship gate, classify.go matrix, orchestration/policy.go roles, workflow/validate.go validators, internal/verify, internal/gates; no coupling to enforcement-profiles; no auto-detection of archetype.

#### Dependencies & Assumptions
- `workflowOrderForFeature` (start_guard.go:12) is the single step-order selection seam; both start.go and hook_autostart.go route through it. It already imports `internal/workflow` + `internal/roadmap`.
- The ship gate is `if current == "validate"` in complete.go:51 — keyed on the step, never on an archetype label.
- `internal/roadmap` ALREADY imports `internal/workflow` (roadmap.go:7); `internal/workflow` does NOT import roadmap. So roadmap-side archetype validation calling `workflow.validateArchetype` adds no new edge and no cycle.
- All hooks guard `if wf.CurrentStep == <name> { ... } else continue` and degrade when a step is absent.
- `centinela merge` validates independently of workflow step (merge.go:37 → runValidateForMerge → executeValidation).

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|---|---|---|---|
| spike perceived as a verification bypass | High | Low | Not a bypass: gate keys on the `validate` step (complete.go:51), spike has no validate step, no `if archetype==spike` branch exists. Merge re-validates (merge.go:48). Safety test pins "spike order excludes validate; others include it." |
| A real feature mislabeled hotfix/refactor loses plan/docs | Med | Med | Explicit opt-in at start; canonical is default; status shows active track + step list so a wrong choice is visible immediately. |
| Bootstrap order vs explicit archetype conflict | Med | Low | Precedence: explicit flag/field wins; otherwise BootstrapStepOrder path in start_guard.go unchanged. Existing bootstrap tests pass an empty flag → unaffected. Covered by a precedence test. |
| start.go:84 prints "Current step: plan" unconditionally | Low | High (today) | Replace literal with `order[0]`; otherwise hotfix start mis-reports "plan" while state CurrentStep is "code". |
| Coupling to enforcement-profiles | Low | Low | Archetype=sequence, profile=strictness; no shared code; orthogonality test (any archetype × any profile). |

#### Rollout
- Step 1: Archetype core (consts, `ArchetypeStepOrder`, `NormalizeArchetype`, `validateArchetype`) — pure data, no wiring.
- Step 2: `Workflow.Archetype` field + set after `NewWithOrder` in start; status render + spike annotation; fix start.go:84.
- Step 3: `--archetype` flag + precedence resolver in/beside `workflowOrderForFeature` (flag → roadmap field → bootstrap → canonical).
- Step 4: `Feature.Archetype` on roadmap + validate-on-load.
- Step 5: safety test + orthogonality test + per-scenario acceptance closing the spec-traceability gate.

#### Handoff
- Next role: feature-specialist
- Outstanding questions (resolutions):
  1. **Spike "no hole" is AIRTIGHT, end-to-end.** The ship gate is `if current == "validate"` (cmd/centinela/complete.go:51) — it keys on the step, not the archetype; there is no `if archetype == "spike"` skip branch. spike skips the gate solely because its order [plan,code] contains no validate step. The merge path closes the loop: `runMerge` (cmd/centinela/merge.go:37) passes `runValidateForMerge` into `worktree.Merge`, which calls it after a clean text merge (internal/worktree/merger.go:48); `runValidateForMerge` (merge.go:53) → `executeValidation()` (cmd/centinela/validate.go:46) runs gates + validate-commands and is **workflow-step-agnostic** (loads config, no step parameter). So a promoted spike is validated at merge regardless of having no validate step. Property holds.
  2. **Steps that may be absent + graceful degradation** (drives senior-engineer guards):
     - **start.go:84** `ui.RenderStep("Current step", "plan")` — HARDCODED, the one real bug. Must become `order[0]`. Does NOT degrade; hotfix would mis-print "plan".
     - hook_plan_advisor.go:33 (`if CurrentStep != "plan" continue`) — degrades; never fires for hotfix. OK.
     - hook_context.go:53 (plan brief-needed), :60 (tests edge-cases), :69 (docs) — each guarded by `CurrentStep ==/!= name`, skip if absent. OK.
     - hook_context_review_mode.go:17 (`ConfirmAfterPlan → CurrentStep == "plan"`) — under after_plan mode a hotfix (no plan) simply never auto-prompts; harmless. OK (minor UX edge, not a break).
     - hook_statusline_rules.go:13/28/40 — per-step `if`; falls through to `"implement-" + CurrentStep` default for any step. OK.
     - hook_orchestration.go:41-43 — `RequiredRolesForFeature` returns nil for unknown step → `len==0 continue`; for hotfix's first step "code" returns [senior-engineer] correctly. OK.
     - workflow/validate.go ValidateArtifacts switch — no default case; absent steps (e.g. spike terminal "code") have no validator, exactly today's behavior. OK.
     - orchestration/policy.go RequiredRoles + evidence/schema_init.go stepForRole — keyed by name, unknown → nil/""; reused names in new positions resolve correctly. OK.
     - order.go:38 `CurrentStep: order[0]` and complete.go nextIdx logic — position-agnostic, iterate the actual order. OK.
     - **No code indexes a positional step or assumes plan/docs/tests exist** beyond the start.go:84 literal. Everything else is name-keyed and degrades by skipping.
  3. **Option B (compose canonical names) confirmed correct.** Requires NO change to IsAllowedInStep / RequiredRoles / ValidateArtifacts. hotfix's first step "code" is allowed by the prewrite (classify.go IsAllowedInStep "code" → TypeCode/TypePlan/TypeRoadmap), so writes work from step one; no plan artifact is required because validatePlan only runs for the "plan" step. Orchestration/autostart key on the step name, not "plan first." No misbehavior found.
  4. **Archetype core belongs in `internal/workflow`, G2-clean.** Step orders are a workflow concern and config must stay leaf. start_guard.go already imports internal/workflow; roadmap already imports internal/workflow (roadmap.go:7) so a `Feature.Archetype` validated via `workflow.validateArchetype` adds NO new import edge and NO cycle (workflow does not import roadmap). Matches PROJECT.md G2.
  5. **Bootstrap precedence safe.** Resolver order: explicit `--archetype` flag → `Feature.Archetype` → bootstrap (BootstrapStepOrder, unchanged) → canonical. Existing bootstrap tests call `workflowOrderForFeature` with no flag set, so they hit the unchanged bootstrap branch and pass. A precedence test pins flag-overrides-bootstrap.

**Spec note:** All 11 scenarios are implementable as-is against this design (resolution, ship-gate, override, persistence, orthogonality, unknown-rejection, status). The acceptance harness must add `// Acceptance:` + `// Scenario:` comments to close the spec-traceability gate on this feature's own spec — flag to qa-senior, not a spec change.
