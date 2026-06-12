# workflow-archetypes — senior-engineer

**Date:** 2026-06-12

## Files Touched

| File | Reason | Lines |
|------|--------|-------|
| `internal/workflow/archetype.go` (new) | Archetype consts, `NormalizeArchetype`, `ValidateArchetype`, `DisplayArchetype`. | 57 |
| `internal/workflow/archetype_order.go` (new) | `ArchetypeStepOrder` returning cloned canonical-step subsets; spike has no validate. | 29 |
| `internal/workflow/state.go` | Added `Archetype string \`json:"archetype,omitempty"\`` to `Workflow`. | 82 |
| `internal/ui/render_status.go` | Added the `Archetype <name>` line + spike annotation via `archetypeLine`; logic stays in workflow. | 67 |
| `internal/roadmap/roadmap.go` | Added optional `Feature.Archetype` field. | 88 |
| `internal/roadmap/archetype.go` (new) | `FeatureArchetype(r, feature)` accessor (split out to keep roadmap.go ≤100). | 18 |
| `internal/roadmap/dependencies.go` | `ValidateDependencies` now rejects unknown feature archetypes on load (calls `workflow.ValidateArchetype`). | 38 |
| `cmd/centinela/start.go` | `--archetype` flag (mirrors `--profile`); resolver call; persist `wf.Archetype`; **start.go:84 bug fix**. | 93 |
| `cmd/centinela/start_guard.go` | `resolveArchetypeOrder` precedence resolver (flag > roadmap field > bootstrap/canonical); `workflowOrderForFeature` left intact. | 88 |

## Architecture Compliance

### Layer boundaries (G2)
- Archetype core lives in `internal/workflow` (domain), not `internal/config` — config stays leaf. `internal/workflow` imports only `internal/config` (existing), never `internal/roadmap`.
- `internal/roadmap` calls `workflow.ValidateArchetype` — roadmap already imports workflow (`roadmap.go:7`), so **no new import edge, no cycle** (workflow does not import roadmap).
- `cmd/centinela` consumes both, as before. The selection seam `workflowOrderForFeature` is wrapped, not modified.
- `internal/ui` carries no archetype logic: it calls `workflow.DisplayArchetype` for the name + annotation and only renders.

### G1 line counts (all ≤100)
`archetype.go` 57 · `archetype_order.go` 29 · `state.go` 82 · `render_status.go` 67 · `roadmap.go` 88 · `roadmap/archetype.go` 18 · `dependencies.go` 38 · `start.go` 93 · `start_guard.go` 88. (`roadmap.go` initially hit 104; the accessor was split into `internal/roadmap/archetype.go`.)

### Deliberately NOT touched (verified)
- `cmd/centinela/complete.go` — the ship gate `if current == "validate"` is untouched; spike safety relies on it being **step-keyed**, not archetype-keyed. There is no `if archetype == "spike"` branch anywhere.
- `cmd/centinela/classify.go` (`IsAllowedInStep`), `internal/orchestration/policy.go` (`RequiredRoles`), `internal/workflow/validate.go` (`ValidateArtifacts`) — all keyed on canonical step names; reused names in new positions resolve correctly.
- `internal/verify` and `internal/gates` — no new imports, no edits.
- `internal/workflow/order.go` `NewWithOrder` signature unchanged; `wf.Archetype` is set after construction as metadata.

### start.go:84 fix
`ui.RenderStep("Current step", "plan")` was hardcoded — a hotfix start mis-printed "plan" while state `CurrentStep` was "code". Now `ui.RenderStep("Current step", order[0])`, so the printed current step matches the resolved order's first step for every archetype.

## Type-Safety Notes
- No `interface{}`/`any`; archetypes are plain string consts validated through `ValidateArchetype`.
- `NormalizeArchetype` passes unknown values through unchanged so the validator (not a silent coercion) rejects typos.
- `ArchetypeStepOrder` returns **clones** (`cloneOrder` / fresh literals) so callers never alias `DefaultStepOrder`'s backing array.
- Errors wrap with `%w` (roadmap) and use `fmt.Errorf` naming the offending value + "archetype" field, matching the enforcement-profile house style.

## Trade-Offs
- **Resolver wraps rather than rewrites `workflowOrderForFeature`.** This keeps the bootstrap branch and all existing bootstrap tests intact (they hit the unchanged function with no flag), at the cost of a second `roadmap.Load()` when a roadmap-archetype is present. Cheap and read-only; favored correctness/back-compat over micro-optimization.
- **Roadmap archetype validation added to `ValidateDependencies`** (called by `Load`) rather than a new validator, so a bad roadmap fails fast everywhere `Load` runs, consistent with `centinela roadmap validate` — no new call site needed.
- **Spike is ungated by absence.** No bypass branch exists; the gate simply never sees a `validate` step. Documented at the spike order with a pointer to the safety argument. A promoted spike is still re-validated step-agnostically at merge.

## Handoff → qa-senior

Source only; qa-senior owns all `*_test.go` and `tests/` artifacts. Suggested coverage (colocated, per-package, ≤100 lines each):

- `ArchetypeStepOrder`: correct order for canonical/hotfix/refactor/spike; unknown → `(nil, false)`; returned slice is a copy (mutating it must not affect `DefaultStepOrder`).
- `NormalizeArchetype`: empty→canonical, known passthrough, unknown passthrough.
- `ValidateArchetype`: empty + 4 known → nil; unknown → error naming the value and "archetype".
- **Safety test (pin this):** spike's resolved order does NOT contain "validate"; hotfix/refactor/canonical orders DO. The ship gate depends on this property.
- Precedence: flag > roadmap `Feature.Archetype` > bootstrap > canonical (table-driven against `resolveArchetypeOrder`).
- State round-trip: `Workflow.Archetype` persists in/out of `.workflow` JSON.
- Roadmap: `Feature.Archetype` parsed; unknown rejected on `Load` (via `ValidateDependencies`), error names the feature.
- Orthogonality: spike + strict → order is spike order AND profile is strict (no shared code).
- Status render: `RenderStatus` shows the archetype line; spike shows the "spike — no ship gate" annotation.
- **start.go:84 status fix:** start output `Current step` shows `order[0]` (hotfix → "code"), not a hardcoded "plan".
- Integration: `start --archetype hotfix` → state order `[code,tests,validate]`, archetype hotfix; `start --archetype spike` → order `[plan,code]`, no validate step.
- Acceptance: per-scenario `tests/acceptance/workflow_archetypes_test.go` with `// Acceptance:` + `// Scenario:` comments closing the spec-traceability gate (11 scenarios).
