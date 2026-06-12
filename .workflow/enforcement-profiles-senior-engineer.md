# Senior-Engineer Report: enforcement-profiles

**Date:** 2026-06-12
**Feature:** enforcement-profiles
**Source-only** (no `*_test.go` / `tests/` — qa-senior owns tests).

## Summary

Added named enforcement profiles — `strict` (default), `guided`, `outcome` — that
scale only the *process* axis (prewrite step-gating, the stop-and-ask review
prompt, mandatory subagent-orchestration evidence, plan-advisor default) while
the *verification* axis (validate-step gates + claim verification) stays welded
on under every profile. Default = **strict** reproduces today's behavior
byte-for-byte. A single resolver (`ProfileDefaults`) is the one source of truth
for every governed knob; precedence is explicit-knob > profile-default > hardcoded.

## Files Touched

| File | Lines | Reason |
|---|---|---|
| `internal/config/enforcement_profile.go` | 42 | NEW. Profile consts, `NormalizeEnforcementProfile` (unknown/empty→strict), `validateEnforcementProfile` (rejects non-empty unknown, names the field). |
| `internal/config/profile_defaults.go` | 40 | NEW. `ProfileKnobs` + `ProfileDefaults(profile)` — the single source of truth for per-profile knob defaults. |
| `internal/config/workflow_config.go` | 22 | `EnforcementProfile` toml field + unexported `RawStepConfirmationMode` shadow (the explicit-vs-defaulted signal). |
| `internal/config/config.go` | 91 | In `Load`: capture raw step_confirmation_mode + validate raw profile BEFORE `applyDefaults` normalizes it (else error is hidden). |
| `internal/config/defaults.go` | 27 | `applyDefaults` normalizes `EnforcementProfile`. |
| `internal/workflow/profile.go` | 27 | NEW. `EffectiveProfile(wf,cfg)` (per-feature→global→strict) + `DisplayProfile(wf)` (cfg-free, for status). |
| `internal/workflow/order.go` | 57 | `NewWithOrder(feature, order, profile)` — profile-aware: pins profile, sets `strict-subagents-v1` ONLY when strict. |
| `internal/workflow/state.go` | 78 | `EnforcementProfile` json field on `Workflow`; `New` passes `ProfileStrict`. |
| `internal/hookpolicy/prewrite.go` | 71 | outcome bypasses the `IsAllowedInStep` ordering block; strict/guided unchanged; no-active-workflow block untouched. |
| `cmd/centinela/start.go` | 90 | `--profile` flag; pins flag-or-global profile into the workflow at start. |
| `cmd/centinela/hook_autostart.go` | 55 | Auto-started features inherit the global profile (graceful strict fallback on config error). |
| `cmd/centinela/hook_context_review_mode.go` | 31 | `effectiveConfirmationMode`: raw-explicit knob > profile default > hardcoded; outcome→auto suppresses. |
| `cmd/centinela/complete.go` | 100 | NO logic change — only a comment pinning "verification is constant, no profile branch here". |
| `internal/ui/render_status.go` | 56 | Read-only `Profile` line via `workflow.DisplayProfile(wf)`. |

## Architecture Compliance

**G2 boundaries (verified):**
- `internal/config` imports nothing internal (still a leaf). New config files import only `fmt`/`strings`.
- `internal/workflow` now imports `internal/config` (order.go, state.go, profile.go) — allowed; no cycle (config never imports workflow).
- `internal/hookpolicy` imports config + workflow — already did; unchanged direction.
- `cmd/centinela` wires only (flag, resolver reads). No new import of `internal/verify` or `internal/gates` anywhere.

**G1 line counts:** every new/modified source file ≤100. Largest: `complete.go` 100, `config.go` 91, `start.go` 90. Split achieved by keeping consts+normalize (`enforcement_profile.go`) and the resolver (`profile_defaults.go`) in separate files.

**Verification constant — confirmed:** `complete.go`'s `current == "validate"` block runs `executeValidation()` + `runClaimVerification()` unconditionally; NO profile branch was added around them. `internal/verify` and `internal/gates` are byte-for-byte untouched. A comment at the gate documents that no profile branch belongs there.

## Key Design Decisions

**applyDefaults raw-capture fix (chosen approach):** added an unexported
`RawStepConfirmationMode` shadow field on `WorkflowConfig` (toml:"-"), set in
`Load` from the decoded value BEFORE `applyDefaults` overwrites the empty
`StepConfirmationMode` with `every_step`. The confirmation resolver consults the
raw field: raw-non-empty ⇒ explicit (wins); raw-empty ⇒ profile default ⇒
hardcoded. Chosen over "stop normalizing in applyDefaults" because other call
sites still read the normalized `StepConfirmationMode` and removing that
normalization would ripple; the shadow field is additive and localized.

**Profile validation timing:** `validateEnforcementProfile` runs against the RAW
decoded value in `Load` *before* `applyDefaults`, because `applyDefaults`
normalizes an unknown value to `strict` — validating after would silently swallow
the bad value. Empty stays valid (→ strict default); only an explicitly-set
unsupported string errors.

**NewWithOrder signature change:** `NewWithOrder(feature, order)` →
`NewWithOrder(feature, order, profile string)`. It normalizes the profile, pins
it on `Workflow.EnforcementProfile`, and sets `OrchestrationMode` to
`strict-subagents-v1` ONLY when `ProfileDefaults(profile).RequireSubagentEvidence`
(true only for strict). Callers updated: `start.go` (flag/global), `hook_autostart.go`
(global), `state.go::New` (strict). `validateOrchestration` / `strictOrchestrationEnabled`
needed no change — they early-return when `OrchestrationMode != strict-subagents-v1`,
so guided/outcome naturally make evidence optional.

## Type-Safety Notes

- No `interface{}`/`any`. `ProfileKnobs` is a concrete struct. Profiles are
  string consts (matching the existing `step_confirmation_mode` / `plan_advisor_mode`
  pattern) — normalized at every boundary so an invalid string can never reach a
  consumer.
- Error wrapping uses `%q` for the offending value and names `workflow.enforcement_profile`.

## Trade-Offs

- **Status uses `DisplayProfile(wf)` (pinned field only), not `EffectiveProfile`.**
  `RenderStatus` has no `cfg`. For a pre-existing workflow with an empty pinned
  field it shows `strict` (the default) rather than re-resolving the current
  global config. Acceptable: new starts always pin the field; the global is only
  a fallback that also defaults to strict. Avoided threading `cfg` through the UI
  (which would put resolution logic in the outer layer).
- **`hook_autostart` loads config a second time** to inherit the global profile.
  Cheap, and keeps the strict fallback explicit on any config error (hooks must
  never break the host session).

## Verification (from the worktree)

- `gofmt -l cmd internal` → empty (clean).
- `go vet ./...` → no issues.
- `go build ./cmd/centinela` → success.
- `go test ./...` → **24/24 packages ok, exit 0** (incl. tests/unit, tests/integration,
  tests/acceptance). Existing assertions that default-started features carry
  `OrchestrationMode == strict-subagents-v1` still pass because default = strict.
- Dogfood (fresh `/tmp/cent-ep`): `status enforcement-profiles` shows `Profile strict`;
  an unknown `enforcement_profile = "wat"` is rejected at load with the field-naming
  error; a throwaway in-package probe confirmed outcome bypass (code write allowed
  in plan), strict still blocks, no-workflow still NeedInit-blocks, and orch modes
  (strict→`strict-subagents-v1`, outcome→empty).

## Handoff → qa-senior

**Test files I mechanically fixed (signature-only, logic untouched) — re-review:**
- `internal/workflow/order_test.go` — `NewWithOrder("f", BootstrapStepOrder)` → `..., "")`.
- `cmd/centinela/start_guard_test.go` — `NewWithOrder("setup", order)` → `..., "")`.

No other test files call `NewWithOrder` directly. Tests that call `workflow.New(feature)`
are unaffected (signature unchanged; `New` now passes `ProfileStrict` internally).

**Coverage qa-senior must add (per-package colocated `_test.go` ≤100 lines, no `-coverpkg`):**

Unit (colocated with the code under test):
- `config`: `NormalizeEnforcementProfile` (each value, unknown→strict, case/space); `validateEnforcementProfile` (empty OK, known OK, unknown errors naming the field); `ProfileDefaults` returns the right `ProfileKnobs` per profile; `Load` raw-capture (explicit `every_step` vs defaulted distinguishable via `RawStepConfirmationMode`).
- `workflow`: `EffectiveProfile` precedence (per-feature override > global > strict default; nil wf, nil cfg); `DisplayProfile` (empty→strict); `NewWithOrder` sets orch mode strict→`strict-subagents-v1`, guided/outcome→empty, and pins `EnforcementProfile`.
- `hookpolicy`: outcome bypasses gating (code write in plan allowed); strict AND guided still block out-of-step writes; no-active-workflow still `NeedInit`-blocks under every profile.
- `cmd/centinela`: `effectiveConfirmationMode` / `shouldRenderReviewPrompt` — explicit raw knob beats profile (outcome+every_step → prompt); outcome (no explicit) → auto suppresses; strict → every_step prompts; after_plan only on plan step.
- **Invariant test:** under {strict, guided, outcome}, completing the validate step with a failing gate/claim is still blocked (assert no profile relaxes the complete gate).

Integration (`tests/integration`): start a feature with `--profile outcome`, write code during the plan step (allowed by prewrite), advance, confirm complete still runs gates + verify. (Note: a true `centinela start` in a greenfield project trips the roadmap-analysis gate; in integration use the existing harness pattern that writes the workflow JSON directly or provides roadmap artifacts.)

Acceptance (`tests/acceptance/enforcement_profiles_test.go`): one test per the 12 `.feature` scenarios, each with the `// Acceptance:` + `// Scenario:` comments to close the spec-traceability gate on this feature's own spec. Map: outcome-allows-plan-write, strict/guided-block, no-workflow-always-block, outcome-suppresses-prompt, explicit-mode-overrides, strict-requires-evidence, guided-no-evidence, outcome-no-evidence, per-feature-overrides-global, unconfigured-keeps-today (strict + gating on + every_step + evidence required), gates+verify-under-every-profile, unknown-profile-rejected-at-load.

Also add `.workflow/enforcement-profiles-edge-cases.md` (tests step requires it).
