# Gatekeeper Report: model-capability-profiles

**Date:** 2026-06-12
**Status:** SAFE

## Analyzed Specs

- `specs/model-capability-profiles.feature` (new feature under review)
- `specs/enforcement-profiles.feature` (shared `EffectiveProfile` resolver + back-compat guarantee)
- `specs/configurable-model-routing.feature` (shared `OrchestrationConfig`)
- `specs/configurable-subagent-models.feature` (skim — orchestration models)
- `specs/configurable-step-confirmation-mode.feature` (skim — confirmation-mode resolver)

## Domain code inspected

- `internal/workflow/profile.go` (`EffectiveProfile`, `DisplayProfile`)
- `internal/workflow/start_resolve.go` (`ResolveStart`, `StartDecision`)
- `internal/workflow/profile_provenance.go` (status provenance)
- `internal/config/config.go` (`Load` raw-capture)
- `internal/config/orchestration.go` (`OrchestrationConfig` new fields)
- `internal/config/capability.go` / `capability_validate.go` / `driver_model.go`
- `cmd/centinela/start.go`, `cmd/centinela/hook_autostart.go`

## Findings

### 1. EffectiveProfile precedence change — BACK-COMPAT SAFE
**Affected spec/scenario:** enforcement-profiles → "An unconfigured project keeps today's behavior (default strict)", "A per-feature profile overrides the global setting", "Gates and claim verification run under every profile".
**Risk:** The new lowest tier (capability default from `wf.DriverModel`) and the gating of the explicit-global tier behind `cfg.Workflow.RawEnforcementProfile != ""` could change resolution for existing zero-config / explicit-global projects.
**Verdict:** No regression. New tier 3 only engages when `wf.DriverModel != ""` AND the model has a capability class (`DefaultProfileForModel` returns `ok=true`). Zero-config workflows have empty `DriverModel`, so resolution falls straight through to `ProfileStrict` — byte-identical. The explicit-global tier still wins over capability and strict.
**Evidence:**
- `internal/workflow/profile_test.go::TestEffectiveProfile_Precedence` asserts: per-feature override wins; explicit global wins when no per-feature pin; `&config.Config{}` (unconfigured) → strict.
- `config.Load` (config.go:73) captures `RawEnforcementProfile = EnforcementProfile` BEFORE `applyDefaults` normalizes the empty knob to strict, so a real loaded config sets the raw signal exactly when the user explicitly set a profile. Confirmed only `Load`-constructed configs reach production (see finding 3).
- Full suite: `go test ./...` → 1465 passed in 24 packages. `centinela validate` → all gates passed.

### 2. Start no longer pins an explicit global enforcement_profile — SAFE
**Affected spec/scenario:** enforcement-profiles status/provenance + "A per-feature profile overrides the global setting".
**Risk:** `start.go`/`hook_autostart.go` now pin `wf.EnforcementProfile` ONLY for an explicit `--profile`; an explicit global is left to resolve live.
**Verdict:** Intentional and correct. `ResolveStart` mirrors runtime `EffectiveProfile` precedence (PinnedProfile → explicit global → capability → strict) for the start-time orchestration-evidence mode, so feature creation still sees the right profile. Leaving the global unpinned makes status provenance read "global" (profile_provenance.go:24) vs "--profile" — a presentation refinement, not a behavior change. No existing scenario asserts that an explicit global is *pinned onto the workflow record*; they assert the *effective* profile, which is preserved.

### 3. Struct-literal Config bypasses raw-capture — CONTAINED, no production regression
**Affected:** any code building `config.Config{}` directly (not via `config.Load`) and setting `EnforcementProfile` without `RawEnforcementProfile`.
**Risk:** Such a config would NOT engage the explicit-global tier (raw empty → treated as "unset"), silently falling through to capability/strict.
**Verdict:** No production exposure. Grep of non-test `.EnforcementProfile =` assignments shows only `wf.EnforcementProfile` (the workflow tier-1 pin, in start.go/hook_autostart.go) and `defaults.go` (post-Load normalization). No production code builds a global-profile Config literal outside `Load`. The one affected test (`tests/acceptance/enforcement_profiles_confirm_test.go`) and the unit helper `cfgWithProfile` were correctly updated to also set `RawEnforcementProfile`, keeping those scenarios green.

### 4. OrchestrationConfig changes are additive — SAFE
**Affected spec/scenario:** configurable-model-routing (all), configurable-subagent-models.
**Risk:** Shared `OrchestrationConfig` struct modified.
**Verdict:** Only NEW fields added (`Capabilities`, `CapabilityProfiles`, `DriverModel`). No existing field (`Models`, `ModelMap`, `UIPaths`) changed type or semantics. The model-routing resolver lives in `internal/orchestration/` and is untouched by this feature. No driver/capability terms appear in the routing specs (verified: no overlap). spec-traceability gate confirms all 24 model-routing-adjacent scenarios still covered. Absent capability tables validate and change nothing (capability_validate.go is no-op on nil maps).

### 5. Built-in haiku id form — COSMETIC SPEC NOTE (non-blocking)
**Affected:** model-capability-profiles.feature Background (line 14) + scenario line 89 use bare `claude-haiku-4-5`; the built-in map (capability.go:29-30) keys the claude form as `claude-haiku-4-5-20251001` and only the opencode form as bare `anthropic/claude-haiku-4-5`.
**Risk:** A literal `claude-haiku-4-5` (claude runner, undated) has no built-in class → resolves to strict via the unknown-fallback tier.
**Verdict:** Not a conflict. The dated id matches `internal/orchestration` tierModels (resolve_test.go:17), so the built-in keys are internally consistent with the routing layer. Scenario line 89's outcome is driven by `--profile outcome` (tier 1) winning, so it never exercises haiku→limited resolution; the assertion passes regardless. This is a spec-prose naming imprecision in the Background table only, not a behavioral defect. Suggestion (optional, non-blocking): align the Background example id with the dated built-in key in a future docs pass.

## Recommendation

**SAFE** — The EffectiveProfile precedence change is additive and back-compat-preserving: zero-config resolves to strict byte-identically (proven by `TestEffectiveProfile_Precedence` + `TestEffectiveProfile_NilInputs`), explicit-global still wins over capability/strict (gated correctly by the `Load`-captured `RawEnforcementProfile` signal), and the new capability tier engages only for a pinned, capability-bearing driver model. Full suite (1465 tests) and `centinela validate` both pass; no existing enforcement-profiles or configurable-model-routing scenario regresses.
