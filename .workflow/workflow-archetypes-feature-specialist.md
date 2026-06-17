### Feature-Specialist Report: workflow-archetypes
**Date:** 2026-06-12

#### Behavior Summary
An archetype is a named start-time preset that selects a subset and ordering of the canonical step names (plan, code, tests, validate, docs) — never new step identifiers. `centinela start --archetype hotfix|refactor|spike|canonical` (default canonical) resolves a step order via `ArchetypeStepOrder`, with precedence explicit flag > roadmap.json `Feature.Archetype` > bootstrap order > canonical; the chosen archetype is pinned in workflow state and shown in `centinela status`. Because every step is a canonical name, every existing mechanism (step-gating matrix, required-role orchestration, per-step validators, the ship gate keyed on `current == "validate"` at complete.go:51) works unchanged. The only safety-relevant property is that the ship gate fires for any order containing the validate step and is simply absent for spike (plan, code) — not a bypass branch — while a promoted spike is still re-validated step-agnostically at merge. Archetype (sequence) and enforcement profile (strictness) are independent.

#### Gherkin Scenarios
Reference: `specs/workflow-archetypes.feature` (11 scenarios).
1. **The hotfix archetype resolves to a code-tests-validate order** — Given a hotfix-started feature / When its order is resolved / Then order is code,tests,validate and plan+docs absent. Maps to `ArchetypeStepOrder("hotfix")`.
2. **The refactor archetype resolves to a plan-code-tests-validate order** — Given refactor / When resolved / Then plan,code,tests,validate; docs absent. Maps to `ArchetypeStepOrder("refactor")`.
3. **The spike archetype resolves to a plan-code order with no validate step** — Given spike / When resolved / Then plan,code and contains no validate step. Maps to `ArchetypeStepOrder("spike")`; safety-relevant omission asserted directly.
4. **The default archetype is the canonical five-step order** — Given no archetype / When resolved / Then archetype is canonical and order is plan,code,tests,validate,docs. Maps to `NormalizeArchetype("")` + `ArchetypeStepOrder("canonical")`.
5. **A ship-gated archetype runs gates and claim verification** (safety) — Given a feature whose resolved order contains the validate step / When that validate step is reached and completed / Then the ship gate fires, running gates and claim verification before advancing. Observable: order contains "validate" + completing that step triggers the `current == "validate"` gate (complete.go:51). Tightened from internal phrasing to observable order-contains + gate-fires.
6. **A spike never reaches the ship gate** (safety) — Given a spike whose order omits validate / When worked to its final step / Then the validate gate is never triggered. Observable: resolved order omits "validate", complete walks to terminal without firing the gate; no `if archetype==spike` branch.
7. **An explicit archetype flag overrides the roadmap archetype** — Given roadmap assigns refactor + started with explicit hotfix flag / When resolved / Then hotfix. Maps to the precedence resolver (flag > roadmap).
8. **The active archetype is pinned in the workflow state** — Given hotfix-started / When state reloaded / Then persisted archetype is hotfix. Maps to `Workflow.Archetype` JSON round-trip.
9. **Archetype and enforcement profile are independent** — Given spike + strict profile / When workflow created / Then order is spike order and profile is strict. Maps to orthogonality: `ArchetypeStepOrder` vs `EnforcementProfile`, no shared code.
10. **An unknown archetype value is rejected** — Given unsupported name / When validated / Then validation fails with an error naming the archetype. Maps to `validateArchetype`.
11. **The status output shows the active archetype** — Given spike-started / When status rendered / Then output names the spike archetype. Maps to `RenderStatus` archetype line (mirrors existing Profile line, render_status.go:18).

#### UX States
| State | Trigger | Surface |
|-------|---------|---------|
| Resolved first step shown | `centinela start --archetype hotfix` | start output `Current step` must show `order[0]` (=code), NOT hardcoded "plan" — fix start.go:84 |
| Archetype + spike annotation | `centinela status <feature>` | status names the active archetype; spike annotated "no ship gate" alongside the step list |
| Ship gate fires | complete the validate step (canonical/hotfix/refactor) | gates + claim verification run before advancing (complete.go:51) |
| Ship gate absent | spike worked to terminal `code` | no validate step exists; gate never fires; merge re-validates if promoted |
| Unknown-archetype error | `--archetype bogus` or bad roadmap field | validation fails fast with an error naming the offending archetype value |
| Default (no flag) | `centinela start` with no `--archetype` | canonical order; zero behavior change — n/a annotation |

#### Out-of-Scope
- No new step names (no reproduce/characterize/prove-equivalent); identifiers stay canonical.
- No coupling to enforcement-profiles; archetype sets sequence, profile sets strictness.
- No change to what any gate or claim verification checks; archetypes change which steps exist, not how a present step is verified.
- No auto-detection of archetype from the change; chosen explicitly by flag/roadmap field.

#### Handoff
- Next role: senior-engineer
- Open clarifications:
  - **start.go:84 bug** — `ui.RenderStep("Current step", "plan")` is hardcoded; must become `ui.RenderStep("Current step", order[0])`, else hotfix mis-prints "plan" while state CurrentStep is "code". (re-flagged for senior-engineer)
  - Status render: add an Archetype line mirroring the existing Profile line (render_status.go:18) plus a spike "no ship gate" annotation.
  - Resolver precedence: explicit flag > roadmap `Feature.Archetype` > bootstrap order (unchanged) > canonical; pin `wf.Archetype` after `NewWithOrder`.
