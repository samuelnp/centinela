# Edge Cases: model-capability-profiles

A pinned driver model's declared capability class selects the DEFAULT
enforcement profile as a new, LOWEST-priority precedence tier. The hard paths
cluster around (a) byte-identical back-compat when nothing is configured,
(b) keeping the capability tier strictly below explicit `--profile` and explicit
global, (c) the explicit-global "honored but not pinned" correctness fix,
(d) lenient model-id matching vs strict class vocabulary, and (e) parse-time
validation of the two new config tables.

## Covered

- **Back-compat: zero workflow + zero config → strict (byte-identical).**
  Risk: the new capability tier could change the default for projects that
  declare nothing. Tested: `TestEffectiveProfile_ZeroConfigStrict` and acceptance
  `TestMCP_ZeroConfigStrict` ("Zero config resolves to strict byte-identically");
  also `TestMCP_AbsentTablesValid` proves absent tables load and still resolve strict.

- **Unknown / typo'd driver model → strict (the safe fallback).**
  Risk: a model-id typo could silently engage a wrong profile; strict (more rails)
  must be the fallback. Tested: `TestEffectiveProfile_CapabilityTier` (driver set,
  no class → strict) and acceptance `TestMCP_UnknownDriverFallsToStrict`. Status
  surfaces it distinctly: `TestMCP_StatusUnknownDriverProvenance` asserts
  `driver: <id> → no capability, default strict`.

- **Explicit `--profile` beats the capability default.**
  Risk: a frontier driver could override a deliberate per-feature pin. Tested:
  `TestEffectiveProfile_HigherTiersBeatCapability` and acceptance
  `TestMCP_FlagBeatsCapability` (outcome flag + strict global + limited driver →
  outcome).

- **Explicit global enforcement_profile beats capability — and is NOT pinned.**
  Risk: the start pin used to bake the global into `wf.EnforcementProfile`, making
  it indistinguishable from a `--profile` pin and short-circuiting the capability
  tier. The correctness fix leaves an explicit global UNPINNED (resolved live at
  tier 2) while still honoring it. Tested: `TestResolveStart` ("explicit global not
  pinned": PinnedProfile empty yet EffectiveProfile=global) and acceptance
  `TestMCP_ExplicitGlobalBeatsCapability`; status reads `(global)` via
  `TestMCP_StatusGlobalProvenance`. The previously-failing
  `enforcement_profiles_confirm_test.go` was fixed to set `RawEnforcementProfile`,
  modelling a real loaded config under the new precedence.

- **`capability_profiles` remaps a class to any profile.**
  Risk: a hardcoded class→profile map would block customization. Tested:
  `TestProfileForCapability` (frontier→guided override) and acceptance
  `TestMCP_CapabilityProfilesOverride` (opus frontier remapped to guided).

- **Driver resolution precedence: flag > env (`CENTINELA_MODEL`) > config.**
  Risk: ambient env could silently override a flag, or config could shadow env.
  Tested: `TestDriverModelFrom` (all permutations + trimming + nil cfg) and
  acceptance `TestMCP_DriverFlagOverridesEnvOverridesConfig`,
  `TestMCP_DriverEnvOverridesConfig`, `TestMCP_DriverFallsToConfig`,
  `TestMCP_DriverEmptyWhenUnconfigured`.

- **Opaque model ids are accepted and pin without error.**
  Risk: `--model` could reject an unknown local id. Tested:
  `TestMCP_OpaqueModelAcceptedAndPins` (a made-up id pins, no error, resolves
  strict via the no-capability fallback).

- **Empty tables and nil cfg are valid (no-op).**
  Risk: an over-eager validator could reject an empty or absent table. Tested:
  `TestValidateCapabilities` (nil cfg, absent tables, empty maps all → nil).

- **Model-id matching is trim-lenient but case-sensitive; class is trim+lower.**
  Risk: inconsistent normalization could match the wrong model or reject a valid
  class. Tested: `TestCapabilityClassFor` (trimmed match hits; `Claude-Opus-4-7`
  with different case MISSES — ids are opaque, not lowercased) and
  `TestProfileForCapability` / `TestMCP_ClassValuesNormalized`
  (`"  Frontier  "` normalizes to frontier).

- **User `[orchestration.capabilities]` override beats the built-in map.**
  Risk: built-ins could shadow a deliberate user re-classification. Tested:
  `TestCapabilityClassFor` ("user beats builtin": opus declared `limited` →
  limited).

- **Config validation rejects malformed capability config at `config.Load`.**
  Risk: a typo'd class/profile could be silently ignored. Tested via real
  `config.Load` over temp tomls: `TestMCP_UnknownClassValueFailsLoad` (names
  `genius`), `TestMCP_EmptyModelIDFailsLoad` (names "empty"),
  `TestMCP_UnknownClassKeyFailsLoad` (names `genius`),
  `TestMCP_UnknownProfileValueFailsLoad` (names `turbo`); plus
  `TestLoad_BadCapabilityClassRejected` and `TestLoad_RawEnforcementProfileCaptured`
  (the raw shadow is captured exactly like `RawStepConfirmationMode`).

- **Status provenance strings are locked (Unicode arrow U+2192).**
  Risk: drifting wording would silently break the status contract. Tested with
  exact-string assertions: `TestProfileProvenance` (all five notes incl. nil-cfg
  fallback) and acceptance `TestMCP_Status*Provenance` (frontier, unknown, global,
  `--profile`, default); UI rendering checked end-to-end in
  `TestRenderStatusWithConfig_Provenance`.

- **Nil-guards on `EffectiveProfile` / `ProfileProvenance` / `ResolveStart`.**
  Risk: a nil wf or cfg from a partial status path could panic. Tested:
  `TestEffectiveProfile_NilInputs`, `TestProfileProvenance` (nil cfg → default),
  and `DriverModelFrom` nil-cfg cases.
