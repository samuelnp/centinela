# model-capability-profiles — qa-senior

## Summary

Comprehensive test suite for the model-capability-profiles feature: a pinned
driver model's declared capability class selects the DEFAULT enforcement profile
as a new, lowest-priority precedence tier, preserving byte-identical back-compat.
All 24 spec scenarios are mapped 1:1 to acceptance tests; every new function and
branch is covered. Full suite green (1465 passed), coverage gate passes at 95.5%.

## Files written

### Colocated unit tests (drive the coverage gate)
- `internal/config/capability_class_test.go` (56) — `CapabilityClassFor` (builtin
  both forms, user override, user-beats-builtin, unknown, empty/whitespace, trim,
  case-sensitivity), `AllowedCapabilityClasses` order.
- `internal/config/capability_profile_test.go` (52) — `ProfileForCapability`
  (3 defaults, override, class normalization), `DefaultProfileForModel` ok/!ok.
- `internal/config/driver_model_test.go` (37) — `DriverModelFrom` flag>env>config,
  trim, all-empty, nil cfg.
- `internal/config/capability_validate_test.go` (46) — `validateCapabilities`
  valid/unknown-class-value/empty-key/unknown-class-key/unknown-profile/empty/nil.
- `internal/config/capability_load_test.go` (54) — integration through `config.Load`
  (bad class fails, normalized class loads, `RawEnforcementProfile` captured).
- `internal/workflow/profile_capability_test.go` (60) — `EffectiveProfile` tier-3
  capability cases + tier-3 miss + zero-config-strict + higher-tiers-beat-capability.
- `internal/workflow/start_resolve_test.go` (62) — `ResolveStart` all tiers; the
  explicit-global-NOT-pinned correctness fix; driver flag>env>config surfacing.
- `internal/workflow/profile_provenance_test.go` (64) — `ProfileProvenance` all
  five exact notes (Unicode → U+2192) + nil-cfg fallback.
- `internal/ui/render_status_provenance_test.go` (38) — `RenderStatusWithConfig`
  frontier + zero-config provenance suffix; worktree-row branch.

### Acceptance tests (spec-traceability, all ≤100 lines)
- `tests/acceptance/model_capability_profiles_precedence_test.go` (76) — 8 scenarios.
- `tests/acceptance/model_capability_profiles_resolution_test.go` (66) — 5 scenarios.
- `tests/acceptance/model_capability_profiles_validation_test.go` (70) — 6 scenarios.
- `tests/acceptance/model_capability_profiles_provenance_test.go` (54) — 5 scenarios.

### Edge-case doc
- `.workflow/model-capability-profiles-edge-cases.md` — enumerates the 13 guaranteed
  edge cases mapped to the tests that cover them.

## Fix to the existing test

`tests/acceptance/enforcement_profiles_confirm_test.go` built
`cfg.Workflow.EnforcementProfile = ProfileOutcome` WITHOUT `RawEnforcementProfile`.
Under the new (correct) precedence, an explicit global needs the raw signal to
engage tier 2 — so the test was failing. Fixed by adding
`cfg.Workflow.RawEnforcementProfile = config.ProfileOutcome` to BOTH
`TestEP_OutcomeSuppressesReviewPrompt` and `TestEP_ExplicitConfirmationOverridesProfile`,
mirroring what `config.Load` does. Assertions were NOT weakened — the tests now
model a real loaded config.

## Scenario → test mapping (24/24, titles match the .feature exactly)

| Spec scenario | Test |
|---|---|
| Zero config resolves to strict byte-identically | TestMCP_ZeroConfigStrict |
| Frontier built-in driver model defaults to outcome | TestMCP_FrontierDefaultsOutcome |
| Capable local driver model declared in config defaults to guided | TestMCP_CapableLocalDefaultsGuided |
| Limited local driver model declared in config defaults to strict | TestMCP_LimitedLocalDefaultsStrict |
| Unknown driver model with no capability falls back to strict | TestMCP_UnknownDriverFallsToStrict |
| Capability profiles override remaps a class to a different profile | TestMCP_CapabilityProfilesOverride |
| Explicit global enforcement_profile beats the capability default | TestMCP_ExplicitGlobalBeatsCapability |
| Per-feature profile flag beats the capability default | TestMCP_FlagBeatsCapability |
| Driver model flag overrides env overrides config | TestMCP_DriverFlagOverridesEnvOverridesConfig |
| Driver model env overrides config when no flag is given | TestMCP_DriverEnvOverridesConfig |
| Driver model falls back to config when no flag and no env | TestMCP_DriverFallsToConfig |
| Driver model is empty when nothing is configured | TestMCP_DriverEmptyWhenUnconfigured |
| An opaque model id with no capability is accepted and pins without error | TestMCP_OpaqueModelAcceptedAndPins |
| Capability class values are normalized by trim and lowercase | TestMCP_ClassValuesNormalized |
| An unknown capability class value fails config load | TestMCP_UnknownClassValueFailsLoad |
| An empty model id key in capabilities fails config load | TestMCP_EmptyModelIDFailsLoad |
| An unknown class key in capability_profiles fails config load | TestMCP_UnknownClassKeyFailsLoad |
| An unknown profile value in capability_profiles fails config load | TestMCP_UnknownProfileValueFailsLoad |
| Absent capability tables are valid and change nothing | TestMCP_AbsentTablesValid |
| Status shows the profile came from a frontier driver model | TestMCP_StatusFrontierProvenance |
| Status shows strict default for an unknown driver model | TestMCP_StatusUnknownDriverProvenance |
| Status shows the global provenance when an explicit global profile wins | TestMCP_StatusGlobalProvenance |
| Status shows the per-feature flag provenance when --profile was passed | TestMCP_StatusFlagProvenance |
| Status shows the strict default provenance for a zero-config feature | TestMCP_StatusDefaultProvenance |

## Verification

- `go build ./...` — Success.
- `gofmt -l internal cmd tests` — empty (clean).
- `go test ./...` — 1465 passed in 24 packages (incl. the fixed acceptance test).
- `./scripts/check-coverage.sh` — coverage gate passed: 95.5% >= 95.0%.
- All new `_test.go` files ≤100 lines (max 76).
- All target functions (capability.go, driver_model.go, capability_validate.go,
  start_resolve.go, profile_provenance.go, profile.go capability branch,
  render_status.go) at 100% statement coverage.

## Handoff

Next role: validation-specialist.
