# Plan: enforcement-profiles

Introduce a named strictness preset that scales *process* while keeping
*verification* welded on. Built on the existing config/normalize pattern and the
existing step_confirmation_mode machinery — no changes to verify.go or gates.go.

## Layer compliance (G2)

- Config in `internal/config/` (leaf). Profile constants + normalizer beside the
  existing `step_confirmation_mode` ones (`workflow_mode.go`).
- Effective-profile resolution + the prewrite relaxation in `internal/hookpolicy`
  and `internal/workflow` (domain). Per-feature override field in the workflow
  state struct.
- `cmd/centinela/` only wires: the `--profile` flag on `start`, and reading the
  effective profile where hooks already load config + workflow state.
- Untouched: `internal/verify`, `internal/gates`.

## Precedence model (back-compat is sacred)

Effective value of each governed knob resolves as:

1. An explicitly-configured knob in `centinela.toml` (e.g. `step_confirmation_mode`) — **always wins.**
2. Else the effective profile's default for that knob.
3. Else the current hard-coded default.

Effective profile resolves as: per-feature override (workflow state) → global
`[workflow] enforcement_profile` → **default that reproduces today's behavior.**

This guarantees: a project that sets neither the profile nor the knobs gets
byte-identical behavior to today. A project that already sets
`step_confirmation_mode = "every_step"` keeps it under any profile.

## What each profile sets (defaults only; explicit knobs override)

| Knob / behavior | **strict (default)** | guided | outcome |
|---|---|---|---|
| prewrite step-gating | on | on | **off** |
| step_confirmation_mode default | every_step | after_plan | auto |
| require subagent (orchestration) evidence | **yes** | no | no |
| plan_advisor_mode default | always | missing_info | off |

**Back-compat default — RESOLVED (default = `strict`).** The big-thinker verified
today's behavior in code: step-gating ON, `step_confirmation_mode` defaults to
`every_step`, and orchestration evidence is REQUIRED because `NewWithOrder`
unconditionally sets `OrchestrationMode="strict-subagents-v1"`. That is exactly
the *strict* row. So the behavior-preserving default is **`strict`**: an
unconfigured project upgrades with byte-identical behavior. This also keeps a
genuinely three-rung spectrum — `guided` becomes a distinct *lighter* opt-in
(rails on, but no mandatory subagent evidence and confirmation after_plan), and
`outcome` drops the rails. (Supersedes the roadmap's offhand "guided is today's
default"; ROADMAP prose to be corrected — the behavior, not the label, is what
must be preserved, and it is.)

**applyDefaults hazard — RESOLVED.** `Load` runs `toml.Decode → applyDefaults →
validateConfig`; an absent `step_confirmation_mode` decodes to `""` then
`applyDefaults` overwrites it with `every_step`, destroying the "was it explicit?"
signal the precedence model needs. Fix: capture the RAW decoded value before
normalization (a `RawStepConfirmationMode` shadow set in `Load`, or move that one
normalization into the resolver) so the resolver can tell explicit-every_step
from defaulted. Resolve confirmation lazily: raw-non-empty ⇒ explicit (wins) →
profile default → hardcoded.

## Implementation

### 1. Config (`internal/config/workflow_mode.go`, `workflow_config.go`, `defaults.go`)
- Add `EnforcementProfile string \`toml:"enforcement_profile"\`` to `WorkflowConfig`.
- Constants `ProfileStrict/Guided/Outcome` + `NormalizeEnforcementProfile` (unknown → the back-compat default).
- Apply in `applyDefaults`. Add `validateEnforcementProfile` (reject unknown values explicitly, like the severity check) wired into `validateConfig`.
- Provide a pure resolver: `func ProfileDefaults(profile string) ProfileKnobs` returning the per-profile knob defaults, so step-gating/confirmation/advisor all consult ONE source of truth.

### 2. Per-feature override (`internal/workflow/state.go`, `cmd/centinela/start.go`)
- Add `EnforcementProfile string \`json:"enforcementProfile,omitempty"\`` to `Workflow`.
- `centinela start --profile <p>` normalizes and persists it. Empty → inherit global.
- A resolver `EffectiveProfile(wf, cfg) string` (per-feature → global → default).

### 3. Step-gating relaxation (`internal/hookpolicy/prewrite.go`, `internal/workflow/classify.go`)
- `EvaluatePrewrite` already has `cfg` + workflow state; compute `EffectiveProfile`. If `outcome`, bypass the `IsAllowedInStep` block (writes allowed in any active step). strict/guided keep today's gating.
- Keep the "no active workflow → block plan/code" behavior unchanged.

### 4. Confirmation prompt (`cmd/centinela/hook_context_review_mode.go`)
- `shouldRenderReviewPrompt` already reads `step_confirmation_mode`. Change it to resolve the *effective* mode: explicit config knob → profile default → hardcoded. outcome → auto (suppress); strict/guided → every_step (unless explicitly overridden).

### 5. Orchestration evidence requirement (`internal/workflow/order.go`, `validate_orchestration.go`, `cmd/centinela/hook_orchestration.go`)
- `NewWithOrder` currently always sets `OrchestrationMode="strict-subagents-v1"`. Gate that on the effective profile: only strict sets it; guided/outcome leave it empty so `validateOrchestration` early-returns (evidence optional). Directives in `hook_orchestration.go` already skip non-strict workflows — so they naturally go quiet for guided/outcome (informational only).
- IMPORTANT: this means guided/outcome features do NOT hard-require the 7 subagent evidence files — the per-step artifact checks (plan file, spec, tests, gatekeeper, docs) still apply.

### 6. Surface the profile (`internal/ui`, status output)
- Show the effective profile in `centinela status` so it's visible which rails are active. (Read-only render.)

## Test plan

- Unit (colocated, per-package coverage):
  - `NormalizeEnforcementProfile` (each value, unknown→default, case/space); `validateEnforcementProfile` rejects bad value.
  - `ProfileDefaults` returns the right knob set per profile.
  - `EffectiveProfile` precedence (per-feature override > global > default).
  - prewrite: outcome bypasses step-gating (a code write during plan step is allowed); strict/guided still block it; no-active-workflow still blocks.
  - confirmation: explicit knob beats profile; outcome→auto suppresses; strict→every_step prompts.
  - orchestration: strict sets strict-subagents-v1 (evidence required); guided/outcome leave it empty (validateOrchestration early-returns).
  - **invariant test:** under EACH profile, the validate-step complete path still calls gates + verify (assert the hard block fires on a failing test/gate regardless of profile).
- Integration (`tests/integration`): start a feature with `--profile outcome`, write code during the plan step (allowed), advance, and confirm complete still runs gates+verify.
- Acceptance (`tests/acceptance/enforcement_profiles_test.go`): per-scenario, with the `// Acceptance:` + `// Scenario:` comments closing the new spec-traceability gate on this feature's own spec.

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| Back-compat break: an upgrade silently changes gating/confirmation | High | Precedence model + a behavior-preserving default profile; an explicit invariant test that an un-configured project's resolved knobs equal today's hardcoded defaults. |
| outcome perceived as "no governance" | Med | Docs + status output make clear gates + claim verification are unchanged; outcome relaxes process, not proof. |
| Profile and step_confirmation_mode conflict confusingly | Med | One documented precedence rule; `centinela status` shows the *effective* values. |
| Scope creep into model-capability auto-selection | Med | Explicitly out of scope; this feature only reads config/flag. |
| G1: touched files exceed 100 lines | Med | Profile knob table + resolver in their own small files. |

## Rollout

1. Config struct + constants + normalize + validate + `ProfileDefaults` (no behavior change yet).
2. `EffectiveProfile` resolver + workflow-state field + `start --profile`.
3. Wire the three consumers (prewrite gating, confirmation, orchestration-evidence) through the resolver, behind the back-compat default.
4. Status surface + the invariant test proving verification is constant.
