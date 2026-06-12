# Senior-Engineer Implementation Report: model-capability-profiles

**Role:** senior-engineer · **Step:** code · **Handoff to:** qa-senior

## Summary

Implemented the model-capability-profiles feature exactly per the approved plan: a
driver model's declared capability class selects the DEFAULT enforcement profile as
a new lowest-priority precedence tier, with byte-identical zero-config back-compat.
All capability vocabulary, validation, and resolution live in the `internal/config`
leaf (no cross-layer edges); `internal/workflow` consumes it; `cmd` stays thin.

## Files created

| File | Lines | Purpose |
|------|-------|---------|
| `internal/config/capability.go` | 97 | classes, built-in model→class map, defaultProfileForClass, CapabilityClassFor, ProfileForCapability, DefaultProfileForModel |
| `internal/config/driver_model.go` | 24 | DriverModelFrom: flag → CENTINELA_MODEL env → config, trim-first-non-empty |
| `internal/config/capability_validate.go` | 47 | validateCapabilities: unknown class, empty model-id key, unknown class key, unknown profile value |
| `internal/workflow/start_resolve.go` | 49 | StartDecision + ResolveStart (start-time precedence, explicit-only pin) |
| `internal/workflow/profile_provenance.go` | 34 | ProfileProvenance for status (mirrors EffectiveProfile tiers) |

## Files modified

| File | Change |
|------|--------|
| `internal/config/orchestration.go` | +Capabilities, +CapabilityProfiles, +DriverModel fields on OrchestrationConfig |
| `internal/config/workflow_config.go` | +RawEnforcementProfile `toml:"-"` (explicit-vs-defaulted signal) |
| `internal/config/config.go` | capture RawEnforcementProfile before applyDefaults |
| `internal/config/file_size_exceptions.go` | call validateCapabilities from validateConfig |
| `internal/workflow/state.go` | +DriverModel `json:"driverModel,omitempty"` field |
| `internal/workflow/profile.go` | new capability precedence tier in EffectiveProfile |
| `cmd/centinela/start.go` | --model flag; ResolveStart wiring; explicit-only profile pin; pin DriverModel |
| `cmd/centinela/hook_autostart.go` | ResolveStart wiring (empty flags; env/config resolved inside) |
| `internal/ui/render_status.go` | RenderStatusWithConfig + profileLine (provenance suffix) |
| `cmd/centinela/status_model.go` | load cfg, thread to RenderStatusWithConfig |
| `internal/workflow/profile_test.go` | updated helper to set RawEnforcementProfile (mirrors Load) |
| `cmd/centinela/review_mode_profile_test.go` | updated cfgRaw helper to set RawEnforcementProfile |

## Slice → code mapping

- **Slice 1 (capability vocabulary, config leaf, pure):** capability.go,
  capability_validate.go, driver_model.go, orchestration.go fields,
  file_size_exceptions.go wiring.
- **Slice 2 (RawEnforcementProfile + explicit-only pin):** workflow_config.go,
  config.go capture, ResolveStart's explicit-only PinnedProfile.
- **Slice 3 (driver-model resolution + pin):** DriverModelFrom, --model flag,
  wf.DriverModel field, pin in both start sites.
- **Slice 4 (capability tier in EffectiveProfile):** profile.go new tier 3.
- **Slice 5 (status provenance):** profile_provenance.go, render_status.go,
  status_model.go.

## Key decisions

1. **Provenance tier-1 note is `--profile` (not "explicit").** The status spec
   (specs/model-capability-profiles.feature lines 189–192) asserts
   `Profile  outcome  (--profile)`. The spec is authoritative over the prompt's
   alternative "explicit" suggestion; used `--profile`.
2. **capability.go is 97 lines** (prompt suggested ≤90). Within the G1 ≤100 hard
   rule. Kept resolution cohesive in one file; provenance lives separately in
   profile_provenance.go.
3. **Status threads cfg via renderStatusBody loading config** — cmd wiring only;
   all provenance logic is in internal/workflow.ProfileProvenance. RenderStatus
   keeps a nil-cfg overload for callers without config.
4. **Tier-2 (global) gating on RawEnforcementProfile != ""** is load-bearing: it
   lets the capability tier engage only when no explicit global profile was set,
   while preserving byte-identical zero-config strict.

## Verification

- `go build ./...` — clean.
- `go vet ./...` — no issues.
- `gofmt -l internal cmd` — no output (formatted).
- `go test ./...` — 1369 passed, 1 failed (see Known issue).
- File sizes — every created/modified .go ≤100 lines.
- Manual precedence checks (throwaway test, all PASS):
  zero-config → strict; {DriverModel:"claude-opus-4-7"} + empty profile + no raw →
  outcome; same + explicit global guided (raw set) → guided.

## Known issue for qa-senior (tests step)

`tests/acceptance/enforcement_profiles_confirm_test.go` (`TestEP_OutcomeSuppressesReviewPrompt`)
constructs a config with `EnforcementProfile = outcome` but NOT `RawEnforcementProfile`.
Under the new (correct) precedence, tier 2 requires `RawEnforcementProfile != ""`, so
that cfg now resolves to strict and the test fails. The fix is a one-line addition
mirroring config.Load and the cmd-level test I already updated:

    cfg.Workflow.RawEnforcementProfile = config.ProfileOutcome

I could NOT apply it during the code step: the prewrite hook classifies `tests/`
paths as TypeTests and blocks them during the code step (colocated `internal/`/`cmd/`
tests are TypeCode and were editable — those I fixed). This is the legitimate
code↔tests step boundary; qa-senior owns this file in the tests step. Do NOT weaken
the assertion — add the raw signal so it asserts the right thing.
