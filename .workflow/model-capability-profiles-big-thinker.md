# Plan: model-capability-profiles

### Big-Thinker Report: model-capability-profiles

**Date:** 2026-06-12

#### Problem

`enforcement-profiles` gave us three named strictness presets but the choice is
manual (`--profile` / `[workflow] enforcement_profile`). `configurable-model-routing`
made the model fully configurable per runner/tier — including weak local models —
yet Centinela still defaults every project to `strict`. The model that actually
drives a workflow has no way to *inform* the default amount of process. A
frontier-model user is taxed with strict ceremony unless they remember a flag; a
local-model user gets the right rails only by luck.

This feature attaches a **capability class** to each model id and lets the
capability of a pinned **driver model** select the *default* enforcement
profile — strictly as a new, lower-priority tier of the existing precedence,
preserving byte-identical back-compat when nothing is configured.

#### Scope (In/Out)

**In:**
- `[orchestration.capabilities]` (model id → capability class) with built-in
  defaults for the three known Anthropic-tier models.
- `[orchestration.capability_profiles]` (class → default profile) override.
- Driver-model resolution: `--model` flag → `CENTINELA_MODEL` env →
  `[orchestration] driver_model` → none.
- Pin `DriverModel` into workflow state at `start`.
- New `EffectiveProfile` precedence tier for the capability-derived default,
  guarded by a new `RawEnforcementProfile` "explicitly-set" signal.
- `centinela status` surfaces driver model + profile provenance.
- Validation of the new config (known classes, known profiles, non-empty ids).
- Parity test keeping config-leaf class/profile string sets in sync with the
  canonical sets.

**Out:**
- Any change to `internal/gates` or `internal/verify` (verification is constant).
- Live re-resolution / mid-feature rewrite (pinned at start).
- Per-role capability (profile is per-workflow; one driver model keys it).
- Runtime model/runner auto-detection (declared, not sniffed).
- Telemetry-based calibration (`capability-calibration`, Phase 7).

#### Dependencies & Assumptions

- **Depends on `enforcement-profiles` (shipped):** reuses `ProfileStrict/Guided/
  Outcome`, `NormalizeEnforcementProfile`, `ProfileDefaults`, `EffectiveProfile`,
  the `--profile` start pin, and the hard back-compat guarantee.
- **Depends on `configurable-model-routing` (shipped):** reuses concrete model
  ids, the config-leaf ↔ orchestration-domain split, and the
  parity-test-keeps-string-sets-in-sync pattern (`orchestration_models.go`).
- **Assumption — opaque model strings:** as with routing, Centinela does NOT
  verify the model exists; it validates *shape* (known class, known profile,
  non-empty id). Capability is a declared property.
- **Assumption — one driver model per workflow:** the enforcement profile is a
  per-workflow scalar, so a single driver model (not a per-role table) is the
  correct key. Per-role routing is a separate, shipped concern.
- **Layer rule (G2):** `internal/config` and `internal/orchestration` are LEAF
  (no internal imports). `internal/workflow` is DOMAIN (may import both). So the
  capability→profile *string* mapping and driver-model resolution can live in
  the config leaf alongside the profile constants they reference; `EffectiveProfile`
  (workflow) consumes them. (See "Divergence" below — this is where I depart
  from the orchestrator's proposed placement.)

#### Divergence from proposed design (with reasons)

1. **Capability→profile mapping lives in `internal/config`, NOT
   `internal/orchestration`.** The orchestrator proposed putting
   `ProfileForCapability` / `DefaultProfileForModel` in `internal/orchestration`.
   But those functions return an *enforcement-profile string*, and the profile
   constants (`ProfileStrict/Guided/Outcome`, `NormalizeEnforcementProfile`)
   live in `internal/config`. `internal/orchestration` today knows nothing about
   enforcement profiles and shouldn't start to — that would couple the routing
   leaf to the governance vocabulary. Both packages are leaves, so neither can
   import the other; placing capability logic in `config` lets it reuse the
   profile constants directly with zero new cross-package coupling. The model-id
   string is opaque, so `config` needs no `orchestration` import either. **Net:
   all capability config + resolution lives in `internal/config`.** This also
   keeps `EffectiveProfile`'s new tier a single `config.` call.

2. **Capability classes: exactly 3, mapped 1:1 to profiles — ADOPTED.** A richer
   taxonomy (separate instruction-following / tool-use / context axes) is
   tempting and the roadmap mentions those dimensions, but v1 only needs to
   *select a profile*, and there are exactly three profiles. Three classes
   named for the spectrum (`frontier`/`capable`/`limited`) is the simplest thing
   that delivers the value; multi-axis scoring is a future feature
   (`capability-calibration` will have the telemetry to justify it). Keeping the
   class names distinct from the profile names (not reusing `strict` etc.) is
   deliberate: it lets `capability_profiles` remap a class to any profile without
   a confusing identity mapping, and it reads correctly ("this model is
   limited" → "default to strict").

3. **Pin DriverModel at start — ADOPTED.** Resolving live would make the
   effective profile depend on ambient env (`CENTINELA_MODEL`) and mutable
   config at *every* hook invocation, so the same feature could silently flip
   strictness between writes. Pinning at start matches how `EnforcementProfile`
   and `Archetype` are already pinned, and gives reproducibility. The pinned
   value is the *resolved model id*, not the derived profile — we still derive
   the profile through `EffectiveProfile` so an explicit profile set *after*
   start (impossible today, but future-proof) still wins, and so `status` can
   show provenance. Pinning the id (not the profile) also keeps the workflow
   state honest about *why* a profile applies.

4. **Ship built-in known-model capabilities — ADOPTED, with a twist.** Built-ins
   cover the three Anthropic tier models so opt-in is one line (`driver_model =
   "claude-opus-4-7"`). But the built-in map is keyed on the *exact* concrete ids
   already in `tierModels` (both the `claude` and the `anthropic/...` opencode
   forms), so it works under both runners without the user re-declaring. Unknown
   ids (any local model) simply have no built-in → the user declares them in
   `[orchestration.capabilities]`. An unknown driver model with no declared
   capability ⇒ no capability tier ⇒ falls through to strict (safe).

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **Regression of `enforcement-profiles` back-compat** (a no-config project stops resolving to strict) | **High** | **Low** | Capability tier engages ONLY when driver model is non-empty AND maps to a capability AND `RawEnforcementProfile`==""·`wf.EnforcementProfile`=="". A dedicated acceptance test asserts byte-identical `EffectiveProfile` for the zero-config matrix across all three workflow profiles. |
| Driver model accidentally overrides an explicit global `enforcement_profile` | High | Med | Add `RawEnforcementProfile` captured before `applyDefaults` (exact `RawStepConfirmationMode` precedent); capability tier sits strictly *below* the explicit-global check. Test: explicit global + driver model ⇒ global wins. |
| Capability tier overrides a per-feature `--profile` | High | Low | `wf.EnforcementProfile` is checked first in `EffectiveProfile` and is non-empty whenever `--profile` was passed OR a global was set at start. Capability is consulted only when `wf.EnforcementProfile`==strict-default AND was not explicitly pinned. (See precedence note on the start-pin ambiguity + its mitigation below.) |
| Start already pins global `enforcement_profile` into `wf.EnforcementProfile`, masking "user set nothing" | High | Med | **Pin only an *explicit* profile** at start: change `start.go` to pin `wf.EnforcementProfile` only when `--profile` given OR `RawEnforcementProfile`!="". When neither, leave `wf.EnforcementProfile` empty so `EffectiveProfile` can reach the capability + global tiers live. Back-compat preserved: empty pin → `EffectiveProfile` still ends at strict. Covered by test. |
| Config surface grows (two new tables + a scalar) | Med | Med | Mirror the routing validators' precise errors; zero-config stays the default; thorough examples. |
| Parity drift: config-leaf class/profile sets diverge from canon | Med | Med | Parity test (same pattern as `orchestration_models_test` runner/tier parity) asserts the leaf sets equal the canonical sets. |
| Local-model id typo → silent strict (no capability match) | Low | Med | Strict is the *safe* fallback (more rails, never fewer). Document; `status` shows "driver: <id> (no capability declared → strict)". |
| File-size (G1 ≤100 lines) on touched files | Med | Med | New code in fresh small files; `EffectiveProfile` gains ~6 lines (currently 9). Budgets below. |

#### Rollout (slices — smallest correct slice first)

- **Slice 1 — capability vocabulary + resolution (config leaf, pure).** Add the
  capability classes, the class→profile default map, `[orchestration.capabilities]`
  + `[orchestration.capability_profiles]` parsing, validation, built-in map, and
  `DefaultProfileForModel(modelID, cfg) (profile string, ok bool)`. No behavior
  change yet (nothing calls it). Fully unit-testable in isolation. Ships value as
  a verified, dormant resolver.
- **Slice 2 — `RawEnforcementProfile` + explicit-only start pin.** Capture the raw
  global profile (precedent: `RawStepConfirmationMode`); change `start.go` to pin
  `wf.EnforcementProfile` only when explicitly set. Pure refactor guarded by a
  back-compat test (zero-config still resolves strict). No capability wiring yet.
- **Slice 3 — driver-model resolution + pin at start.** `DriverModel(cfg)` (flag
  → env → config), `--model` flag on `start`, `wf.DriverModel` field, pin at
  start. Still no profile effect (capability tier not added). Verifies the pin in
  isolation.
- **Slice 4 — wire the capability tier into `EffectiveProfile`.** Add the single
  new precedence tier reading `wf.DriverModel` → `config.DefaultProfileForModel`.
  This is the behavior-activating slice; its acceptance test is the full
  precedence matrix. Back-compat asserted by Slice 2's test still passing.
- **Slice 5 — `status` provenance.** Surface driver model + which precedence tier
  produced the active profile. Read-only, presentation-only.

Slices 1–3 ship with zero behavior change; the feature "turns on" only at slice
4. Each slice is independently testable and leaves the tree green.

#### Handoff

**Next role:** feature-specialist.

Outstanding questions for the feature-specialist to lock:
1. **`status` provenance string wording** — exact text for each tier (e.g.
   `Profile  outcome  (driver: claude-opus-4-7 → frontier)` vs
   `Profile  strict  (global)`). Presentation, not logic; pick concise phrasing.
2. **Should `--model` validate the id against built-ins/declared capabilities, or
   accept any opaque string?** Lean: accept any string (consistent with routing's
   opaque-model stance); a model with no capability simply yields no capability
   tier. Confirm.
3. **`CENTINELA_MODEL` env precedence vs `--model`** — confirm flag > env > config
   (matches `CENTINELA_RUNNER`-style wiring intent in routing). Locked here as
   flag > env > config; specialist confirms naming `CENTINELA_MODEL`.
4. **Gherkin scenarios** — the precedence matrix (driver-only, driver+global,
   driver+`--profile`, no-driver) maps 1:1 to acceptance scenarios.

---

## Implementation Plan

### Layer placement (G2-safe)

All capability vocabulary, validation, and resolution live in **`internal/config`**
(leaf) — they reference the profile constants already in that package and treat
the model id as an opaque string, so they need NO `internal/orchestration` import.
`EffectiveProfile` (in `internal/workflow`, domain) consumes them via the single
`config.DefaultProfileForModel` call. `cmd/centinela` wires the `--model` flag and
the pin. No new cross-layer edges; `import_graph` gate stays green.

### New / changed files (package · ≤100-line budget)

| File | Pkg | New/Changed | Budget | Contents |
|------|-----|-------------|--------|----------|
| `internal/config/capability.go` | config | new | ≤70 | classes, class→profile defaults, `CapabilityClassFor`, `ProfileForCapability`, `DefaultProfileForModel`, built-in model→class map |
| `internal/config/capability_validate.go` | config | new | ≤80 | parse-time validation of `[orchestration.capabilities]` + `[orchestration.capability_profiles]` |
| `internal/config/driver_model.go` | config | new | ≤40 | `DriverModel(cfg) string` — flag-less resolution (env → config); flag injected by cmd |
| `internal/config/orchestration.go` | config | changed | +6 lines (≤90) | add `Capabilities map[string]string` + `CapabilityProfiles map[string]string` to `OrchestrationConfig` |
| `internal/config/workflow_config.go`* | config | changed | +1 field | add `RawEnforcementProfile string` (not a toml field; set in Load like `RawStepConfirmationMode`) |
| `internal/config/config.go` | config | changed | +1 line | capture `cfg.Workflow.RawEnforcementProfile = cfg.Workflow.EnforcementProfile` before `applyDefaults` |
| `internal/config/file_size_exceptions.go` | config | changed | +2 lines | call `validateCapabilities(cfg)` from `validateConfig` |
| `internal/workflow/state.go` | workflow | changed | +3 lines | add `DriverModel string json:"driverModel,omitempty"` |
| `internal/workflow/order.go` | workflow | changed | +1 field set | accept/pass driver model (or set on wf in start) |
| `internal/workflow/profile.go` | workflow | changed | +~7 lines (≤30) | new capability tier in `EffectiveProfile` |
| `cmd/centinela/start.go` | cmd | changed | +~10 lines (≤100) | `--model` flag; resolve driver model; explicit-only profile pin; pin `wf.DriverModel` |
| `internal/ui/render_status.go` | ui | changed | +~4 lines (≤80) | driver-model + provenance line |
| `internal/workflow/profile.go` (provenance helper) | workflow | changed | (same file) | `ProfileProvenance(wf, cfg) string` for status, OR a small `internal/ui` helper reading workflow |

\* exact file holding `WorkflowConfig` to be confirmed by specialist (the struct
with `EnforcementProfile`/`RawStepConfirmationMode` fields); add the raw field there.

Test files (qa-senior owns, listed for completeness): `capability_test.go`,
`capability_validate_test.go`, `driver_model_test.go`,
`capability_parity_test.go` (config), `profile_test.go` additions (workflow),
plus `tests/acceptance/model_capability_profiles_test.go` for the precedence
matrix.

### Config structs & function signatures

```go
// internal/config/orchestration.go — OrchestrationConfig gains:
type OrchestrationConfig struct {
    UIPaths            []string                     `toml:"ui_paths"`
    Models             map[string]RoleModelValue    `toml:"models"`
    ModelMap           map[string]map[string]string `toml:"model_map"`
    Capabilities       map[string]string            `toml:"capabilities"`         // model id → class
    CapabilityProfiles map[string]string            `toml:"capability_profiles"`  // class → profile
    DriverModel        string                       `toml:"driver_model"`
}

// internal/config/capability.go
const (
    CapabilityFrontier = "frontier"
    CapabilityCapable  = "capable"
    CapabilityLimited  = "limited"
)

// AllowedCapabilityClasses returns the valid classes in stable order.
func AllowedCapabilityClasses() []string

// builtinModelCapability maps the known concrete model ids (both claude and
// anthropic/... forms from tierModels) to their class.
//   claude-opus-4-7, anthropic/claude-opus-4-7         → frontier
//   claude-sonnet-4-6, anthropic/claude-sonnet-4-6      → capable
//   claude-haiku-4-5-20251001, anthropic/claude-haiku-4-5 → limited
var builtinModelCapability = map[string]string{ /* ... */ }

// defaultProfileForClass maps a class to its default enforcement profile.
//   frontier→outcome, capable→guided, limited→strict
func defaultProfileForClass(class string) string

// CapabilityClassFor returns the declared class for a model id: user
// [orchestration.capabilities] override first, then built-in, else ("", false).
func CapabilityClassFor(modelID string, cfg *Config) (string, bool)

// ProfileForCapability returns the default profile for a class, honoring a
// [orchestration.capability_profiles] override, else defaultProfileForClass.
func ProfileForCapability(class string, cfg *Config) string

// DefaultProfileForModel resolves a model id straight to its default profile:
// CapabilityClassFor → ProfileForCapability. ok=false when the id has no class
// (no built-in, none declared) — caller must NOT engage the capability tier.
func DefaultProfileForModel(modelID string, cfg *Config) (profile string, ok bool)

// internal/config/driver_model.go
// DriverModel returns the configured driver model id: env CENTINELA_MODEL wins
// over [orchestration] driver_model. The --model flag is layered in cmd (it
// passes its value as an explicit override before calling this, or cmd resolves
// flag→DriverModel(cfg)). Empty when nothing is configured.
func DriverModel(cfg *Config) string

// internal/config/capability_validate.go
// validateCapabilities rejects unknown classes in [orchestration.capabilities]
// values, empty model-id keys, unknown class keys / unknown profile values in
// [orchestration.capability_profiles]. Absent/empty tables are valid.
func validateCapabilities(cfg *Config) error
```

### Exact new `EffectiveProfile` precedence

```go
// internal/workflow/profile.go
func EffectiveProfile(wf *Workflow, cfg *config.Config) string {
    // 1. Per-feature pin (--profile, or an explicit global captured at start).
    if wf != nil && wf.EnforcementProfile != "" {
        return config.NormalizeEnforcementProfile(wf.EnforcementProfile)
    }
    // 2. Explicit global [workflow] enforcement_profile (raw == set).
    if cfg != nil && cfg.Workflow.RawEnforcementProfile != "" {
        return config.NormalizeEnforcementProfile(cfg.Workflow.EnforcementProfile)
    }
    // 3. NEW: capability default from the pinned driver model.
    if wf != nil && wf.DriverModel != "" && cfg != nil {
        if profile, ok := config.DefaultProfileForModel(wf.DriverModel, cfg); ok {
            return config.NormalizeEnforcementProfile(profile)
        }
    }
    // 4. Strict back-compat default.
    return config.ProfileStrict
}
```

**Precedence (highest → lowest):**
1. `wf.EnforcementProfile` (explicit `--profile`, or an explicitly-set global
   pinned at start).
2. Explicit global `[workflow] enforcement_profile` (`RawEnforcementProfile != ""`).
3. Capability default from `wf.DriverModel` (engages only with a known class).
4. `strict` default.

**Critical change enabling this:** today `start.go` pins
`wf.EnforcementProfile = cfg.Workflow.EnforcementProfile` even when the user set
nothing (it's normalized to `"strict"`), which would short-circuit at tier 1 and
suppress capability. **Slice 2 changes the start pin to pin only an explicit
profile** (`--profile` given OR `RawEnforcementProfile != ""`); otherwise leave
`wf.EnforcementProfile` empty so tiers 2–4 are reachable. Back-compat holds: with
no driver model, an empty pin still ends at tier 4 = strict.

### New centinela.toml sections

```toml
[orchestration]
# The model the workflow is keyed off for its default enforcement profile.
# Lowest-priority source: --model flag > CENTINELA_MODEL env > this value.
driver_model = "claude-opus-4-7"

# Concrete model id → capability class. Declare local/unknown models here.
# Built-in defaults already cover the three Anthropic tier models.
[orchestration.capabilities]
"deepseek/deepseek-coder"    = "limited"
"moonshotai/kimi-k2"         = "capable"
"claude-opus-4-7"            = "frontier"   # (already built-in; shown for clarity)

# Optional: remap a capability class to a different default profile.
[orchestration.capability_profiles]
frontier = "outcome"
capable  = "guided"
limited  = "strict"
```

Built-in `model → class` (no config required):

| Model id (claude / opencode form) | Class |
|-----------------------------------|-------|
| `claude-opus-4-7` / `anthropic/claude-opus-4-7` | frontier |
| `claude-sonnet-4-6` / `anthropic/claude-sonnet-4-6` | capable |
| `claude-haiku-4-5-20251001` / `anthropic/claude-haiku-4-5` | limited |

Default `class → profile` (no config required): frontier→outcome, capable→guided,
limited→strict.

### Validation rules (parse-time, at `config.Load`)

- `[orchestration.capabilities]`: each value must be a known class (`frontier` /
  `capable` / `limited`, normalized trim+lower); each key (model id) must be
  non-empty. Absent/empty table valid.
- `[orchestration.capability_profiles]`: each key must be a known class; each
  value must be a known enforcement profile (reuse `validateEnforcementProfile`
  logic / `NormalizeEnforcementProfile`). Absent/empty table valid.
- `driver_model`: opaque, validated only as a string (no existence check); empty
  is valid. (Consistent with routing's opaque-model decision.)
- Wired in via `validateConfig` (file_size_exceptions.go) calling
  `validateCapabilities(cfg)`, after the existing orchestration validators.

### Parity test

`internal/config/capability_parity_test.go` asserts the config-leaf
`AllowedCapabilityClasses()` set equals the canonical profile↔class mapping and
that every class has a `defaultProfileForClass` entry mapping to a real profile —
mirroring the runner/tier parity test pattern. (Profiles already live in config,
so there is no cross-package set to mirror as with routing; the parity assertion
is class-completeness + profile-validity.)

### Back-compat acceptance (the load-bearing test)

`tests/acceptance/model_capability_profiles_test.go` precedence matrix:

| `--profile` | global `enforcement_profile` | `driver_model` (class) | Expected `EffectiveProfile` |
|:-----------:|:----------------------------:|:----------------------:|:---------------------------:|
| — | — | — | **strict** (byte-identical back-compat) |
| — | — | opus (frontier) | outcome |
| — | — | unknown-local (no class) | strict |
| — | — | local declared `limited` | strict |
| — | — | local declared `capable` | guided |
| — | guided | opus (frontier) | **guided** (explicit global wins) |
| outcome | strict | haiku (limited) | **outcome** (--profile wins) |

The first row is the back-compat guarantee and must pass unchanged before and
after slice 4.
