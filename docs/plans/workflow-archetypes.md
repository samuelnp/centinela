# Plan: workflow-archetypes

Add named workflow tracks (hotfix/refactor/spike) as presets that select a
subset/ordering of the EXISTING canonical steps. Reuse every existing mechanism;
change only the step-order selection seam + state + start flag + status render.
Mirrors the just-shipped enforcement-profiles pattern (flag + state field +
roadmap override + normalize/validate + status surface).

## Layer compliance (G2)

- Archetype constants + `ArchetypeStepOrder(name)` + `NormalizeArchetype` /
  `validateArchetype` live in `internal/workflow` (domain) — step orders are a
  workflow concern, and `internal/config` must stay leaf. (enforcement_profile
  lives in config because it's a config knob read widely; archetype is a
  start-time step-order selector owned by workflow.)
- `cmd/centinela/start_guard.go` (`workflowOrderForFeature`) consumes it.
- `internal/roadmap` gains an optional `Archetype` field on the Feature struct.
- Untouched: complete.go ship gate, classify.go matrix, orchestration/policy.go,
  workflow/validate.go validators, internal/verify, internal/gates.

## The four archetypes (canonical step subsets)

```
canonical : plan, code, tests, validate, docs      (default — unchanged)
hotfix    : code, tests, validate
refactor  : plan, code, tests, validate
spike     : plan, code                              (no validate → ungated)
```

`ArchetypeStepOrder(name) ([]string, ok)` returns the order; unknown → not-ok.

## Why this needs almost no new wiring

- **Step-gating** (`IsAllowedInStep`): every step in every archetype is a
  canonical name already in the matrix. hotfix starting at `code` → the prewrite
  allows TypeCode/TypePlan writes from step one; no plan artifact is required
  because `validatePlan` only runs for the `plan` step (absent in hotfix). ✔
- **Ship gate** (`complete.go` `current == "validate"`): fires for canonical/
  hotfix/refactor (they contain `validate`); never for spike. **No edit to
  complete.go.** ✔
- **Required roles** (`RequiredRoles`): keyed on canonical names — hotfix
  requires senior-engineer (code) + qa-senior (tests) + validation-specialist
  (validate); spike requires big-thinker+feature-specialist (plan) +
  senior-engineer (code). Lighter ceremony falls out of fewer steps. ✔ (Under
  guided/outcome profiles, that evidence isn't mandatory anyway.)
- **Per-step validators** (`ValidateArtifacts`): canonical-named steps validate
  exactly as today. A hotfix's `validate` step still needs a gatekeeper report;
  spike's terminal `code` step has no validator (today `code` has none). ✔

## Implementation

### 1. Archetype core (`internal/workflow/archetype.go`)
- Consts `ArchetypeCanonical="canonical"`, `Hotfix`, `Refactor`, `Spike`.
- `NormalizeArchetype(s)`: empty → canonical; unknown stays as-is for the
  validator to reject (don't silently coerce unknown → canonical, or a typo
  would run the wrong track).
- `ArchetypeStepOrder(name) ([]string, bool)`: the table above. canonical →
  DefaultStepOrder.
- `validateArchetype(name)`: empty or one of the four = ok; else error naming the
  flag/field. Called from start (flag) and roadmap load (field).

### 2. State (`internal/workflow/state.go`, `order.go`)
- Add `Archetype string \`json:"archetype,omitempty"\`` to Workflow.
- `NewWithOrder` already takes the order; set `wf.Archetype` after construction
  in start (keep NewWithOrder signature stable — order already encodes the
  sequence; archetype is metadata for display + status).

### 3. Selection seam (`cmd/centinela/start_guard.go`, `start.go`)
- `centinela start --archetype <name>` flag (mirror `--profile`).
- Precedence in `workflowOrderForFeature` (or a small resolver beside it):
  explicit `--archetype` flag → roadmap.json Feature.Archetype → bootstrap-phase
  order (unchanged) → canonical default.
- Resolve archetype → step order via `ArchetypeStepOrder`; pass to NewWithOrder;
  persist `wf.Archetype`. Bootstrap features keep BootstrapStepOrder UNLESS an
  explicit archetype is given (document this precedence; bootstrap is itself a
  kind of archetype).

### 4. Roadmap field (`internal/roadmap/roadmap.go`)
- Add optional `Archetype string \`json:"archetype,omitempty"\`` to Feature.
- A loader accessor `FeatureArchetype(r, feature) string`. Validate on load
  (reject unknown so a bad roadmap fails fast, consistent with roadmap validate).

### 5. Status surface (`internal/ui` + status command)
- Show `Archetype <name>` and the actual step list in `centinela status`
  (read-only). For spike, annotate "spike — no ship gate" so it's visibly
  not-for-shipping.

## Test plan

- Unit (colocated, per-package):
  - `ArchetypeStepOrder` returns the right order for each of the 4; unknown → !ok.
  - `NormalizeArchetype` (empty→canonical, known passthrough, unknown passthrough
    for validator); `validateArchetype` accepts empty+4, rejects unknown naming
    the field.
  - selection precedence: flag > roadmap field > bootstrap > canonical (table).
  - **safety test:** a spike's resolved order does NOT contain "validate"; a
    hotfix/refactor/canonical order DOES — pin this, since the gate depends on it.
  - state round-trip: Archetype persists in/out of .workflow JSON.
  - roadmap: Feature.Archetype parsed; unknown rejected on load.
  - status render shows archetype + the spike annotation.
- Integration (`tests/integration`): start --archetype hotfix → state has order
  [code,tests,validate] and archetype hotfix; start --archetype spike → order
  [plan,code], no validate step.
- Acceptance (`tests/acceptance/workflow_archetypes_test.go`): per-scenario, with
  the `// Acceptance:` + `// Scenario:` comments closing the spec-traceability
  gate on this feature's own spec.

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| spike perceived as a verification bypass | High | It isn't: gate keys on the `validate` step, not a label; spike has no validate step; merge still validates. The safety test pins "spike order has no validate." Docs + status make it explicit. |
| hotfix/refactor skip plan/docs → a real feature mislabeled loses needed steps | Med | Archetype is an explicit opt-in at start; canonical is the default; status shows the active track so a wrong choice is visible immediately. |
| Bootstrap-phase order vs explicit archetype conflict | Med | Defined precedence: explicit flag wins; otherwise bootstrap logic unchanged. Covered by a precedence test. |
| Coupling to enforcement-profiles | Low | Archetype = sequence, profile = strictness; no shared code; an orthogonality test asserts any archetype × any profile. |
| G1 >100 lines | Low | archetype.go small; the order table is data. |

## Rollout

1. Archetype core (consts, ArchetypeStepOrder, normalize/validate) — pure, no wiring.
2. State field + status render.
3. start --archetype flag + selection precedence in workflowOrderForFeature.
4. roadmap.json Feature.Archetype + load validation.
5. The safety test + orthogonality test + acceptance closing the dogfood.
