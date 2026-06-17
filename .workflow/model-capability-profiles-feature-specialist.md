### Feature-Specialist Report: model-capability-profiles

**Date:** 2026-06-12

#### Behavior Summary

A driver model is resolved at `centinela start` (precedence `--model` flag >
`CENTINELA_MODEL` env > `[orchestration] driver_model` config > none) and pinned
into workflow state as an opaque id. When the workflow has no explicitly-set
profile, that driver model's declared capability class (`frontier`/`capable`/
`limited`) selects the default enforcement profile as a NEW lowest-priority
precedence tier in `EffectiveProfile`: tier 1 explicit `--profile` (or an
explicitly-set global pinned at start) > tier 2 explicit global
`[workflow] enforcement_profile` (gated by `RawEnforcementProfile != ""`) > tier
3 capability default from `wf.DriverModel` (engages only when the id maps to a
known class via the built-in map or `[orchestration.capabilities]`) > tier 4
`strict`. Capability classes map 1:1 to profiles by default (frontier→outcome,
capable→guided, limited→strict), remappable via `[orchestration.capability_profiles]`.
The observable surface is the terminal output of `centinela status`, which gains
a provenance suffix on the existing `Profile` line naming which tier produced the
active profile. Zero-config behavior is byte-identical to today (strict), which is
the absolute back-compat invariant.

#### Resolved Open Questions (locked decisions)

1. **`status` provenance wording — LOCKED.** The existing `Profile  <value>` line
   gains a parenthesised provenance suffix. Exact strings the senior-engineer
   must implement (rendered after the profile value, separated by two spaces;
   the suffix is muted-styled like the archetype note):
   - tier 1, explicit `--profile`: `Profile  outcome  (--profile)`
   - tier 1/2, explicit global: `Profile  guided  (global)`
   - tier 3, capability hit: `Profile  outcome  (driver: claude-opus-4-7 → frontier)`
     (pattern: `(driver: <id> → <class>)`)
   - tier 3 miss / unknown id with a driver pinned: `Profile  strict  (driver: some/unknown-local-model → no capability, default strict)`
     (pattern: `(driver: <id> → no capability, default strict)`)
   - tier 4, zero-config: `Profile  strict  (default)`
   The arrow is the Unicode `→` (U+2192), matching the report's own usage. Provenance
   is presentation-only and is derived by a `ProfileProvenance(wf, cfg)` helper in
   `internal/workflow` (logic stays out of the `internal/ui` outer layer).

2. **`--model` validation — CONFIRMED opaque.** `--model` accepts ANY string with
   no existence check, consistent with `configurable-model-routing`'s opaque-model
   stance. A model with no built-in and no declared capability simply yields no
   capability tier (`DefaultProfileForModel` returns `ok=false`) and falls through
   to strict. No validation error is ever raised for an unknown driver model id.

3. **`CENTINELA_MODEL` env naming + precedence — CONFIRMED.** Env var name is
   `CENTINELA_MODEL` (matches the `CENTINELA_RUNNER` naming convention from
   routing). Resolution precedence is `--model` flag > `CENTINELA_MODEL` env >
   `[orchestration] driver_model` config > none (empty). Empty at every layer ⇒
   empty pinned driver model ⇒ capability tier never engages.

4. **Unknown/typo'd driver model — CONFIRMED safe fallback to strict.** An id with
   no capability match yields no capability tier and resolves to strict (more
   rails, never fewer — the safe direction). It is surfaced in status as
   `(driver: <id> → no capability, default strict)` so the user can see the typo
   produced the strict default rather than silently mis-firing.

#### Gherkin Scenarios

All scenarios live in `specs/model-capability-profiles.feature` (24 scenarios +
a shared Background pinning the built-in map). Titles are unique and stable for
the spec-traceability gate (`// Scenario: <name>` comments in
`tests/acceptance/model_capability_profiles_test.go`). Grouped by precedence tier
and concern:

Back-compat (tier 4):
- Zero config resolves to strict byte-identically

Capability default (tier 3):
- Frontier built-in driver model defaults to outcome
- Capable local driver model declared in config defaults to guided
- Limited local driver model declared in config defaults to strict
- Unknown driver model with no capability falls back to strict
- Capability profiles override remaps a class to a different profile

Explicit global wins (tier 2):
- Explicit global enforcement_profile beats the capability default

Per-feature flag wins (tier 1):
- Per-feature profile flag beats the capability default

Driver-model resolution (flag > env > config):
- Driver model flag overrides env overrides config
- Driver model env overrides config when no flag is given
- Driver model falls back to config when no flag and no env
- Driver model is empty when nothing is configured

Opaque ids + normalization:
- An opaque model id with no capability is accepted and pins without error
- Capability class values are normalized by trim and lowercase

Config validation:
- An unknown capability class value fails config load
- An empty model id key in capabilities fails config load
- An unknown class key in capability_profiles fails config load
- An unknown profile value in capability_profiles fails config load
- Absent capability tables are valid and change nothing

Status provenance:
- Status shows the profile came from a frontier driver model
- Status shows strict default for an unknown driver model
- Status shows the global provenance when an explicit global profile wins
- Status shows the per-feature flag provenance when --profile was passed
- Status shows the strict default provenance for a zero-config feature

#### UX States

The only observable surface is the terminal output of `centinela status` (and the
non-erroring success of `centinela start --model <id>`). There is no async load.

| State    | Trigger | Surface |
|----------|---------|---------|
| loading  | n/a — synchronous CLI, no async data fetch | n/a |
| empty    | feature started with no driver model and no explicit profile | `Profile  strict  (default)` — no driver line; identical to pre-feature output |
| error    | invalid `[orchestration.capabilities]`/`[orchestration.capability_profiles]` at `config.Load` (unknown class, empty id, unknown profile) | `centinela` exits non-zero with a config error naming the offending value (e.g. `genius`, `turbo`); workflow does not start. NOTE: a bad `--model` id is NOT an error — opaque ids are always accepted. |
| success  | feature started with a driver model that maps to a capability | `Profile  <profile>  (driver: <id> → <class>)`; explicit profile/global shows `(--profile)`/`(global)`; unknown driver shows `(driver: <id> → no capability, default strict)` |

#### Out-of-Scope

- No change to `internal/gates` or `internal/verify` — capability moves the
  *process* axis only, verification is constant.
- No live re-resolution — driver model + derived profile pinned at start; mutating
  config or `CENTINELA_MODEL` mid-feature does not retro-rewrite the pinned value.
- No per-role capability — one driver model keys the per-workflow profile;
  per-role routing remains `configurable-model-routing`.
- No runtime model/runner auto-detection — capability is declared, never sniffed.
- No telemetry-based calibration — that is `capability-calibration` (Phase 7).
- No existence validation of model ids — `--model` and all config model ids are
  opaque strings; only config *shape* (known class, known profile, non-empty key)
  is validated.

#### Handoff

- **Next role:** senior-engineer
- **Open clarifications:** none — all four open questions are locked above. The
  senior-engineer implements the precedence in `EffectiveProfile` (tier 3 between
  the explicit-global check and the strict default), the `ProfileProvenance`
  helper for status, the explicit-only start pin (`wf.EnforcementProfile` pinned
  only when `--profile` given OR `RawEnforcementProfile != ""`), and the
  config-leaf capability vocabulary/validation per the plan's file budgets.
