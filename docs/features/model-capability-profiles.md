# Feature: model-capability-profiles

- surface: internal
- status: planned
- roadmap: Phase 6 — Capability-Adaptive Governance (keystone)
- fixes: enforcement strictness is chosen by hand (flag/config); a model's
  declared capability has no way to *inform* the default. Frontier models keep
  getting taxed with strict ceremony unless a human remembers `--profile`.

## Problem

`enforcement-profiles` decoupled *how much process* from *whether outcomes are
verified*, and gave us three presets (`strict` / `guided` / `outcome`). But the
profile is still selected by intuition: a human passes `--profile outcome` or
sets `[workflow] enforcement_profile`. Nothing connects the *model actually
driving the workflow* to the amount of scaffolding it gets.

`configurable-model-routing` made the model fully configurable per runner/tier —
including local models on Ollama / llama.cpp / OpenAI-compatible endpoints. So a
project can now route to a weak local model OR a frontier model, but Centinela
still defaults everyone to `strict`. The frontier user must remember a flag; the
local-model user gets the right rails only by luck. The model knows what it can
do; the framework should let that knowledge pick the default.

## The core idea: capability selects the *default* profile

Attach a **capability class** to each model id, and map each class to a default
enforcement profile:

| Capability | Default profile | Rationale |
|------------|-----------------|-----------|
| `frontier` | `outcome`  | follows instructions + uses tools reliably; rails are pure tax |
| `capable`  | `guided`   | drives steps itself, still benefits from light ordering rails |
| `limited`  | `strict`   | needs maximum physical scaffolding |

A **driver model** (the one model the workflow is keyed off) is resolved at
`start` and pinned into workflow state. Its capability picks the *default*
profile — but only as a new, lower-priority tier of the existing precedence: an
explicit `--profile` or an explicit global `enforcement_profile` still wins.

## Hard invariant (must not regress)

Back-compat with `enforcement-profiles` is absolute: a project that sets **no**
driver model resolves exactly as today (explicit profile → global → strict).
The capability-derived default engages **only** when a driver model is known AND
maps to a capability AND no profile was set explicitly. Verification
(`internal/verify`, `internal/gates`) is untouched — capability moves the
*process* axis only, never the verification axis.

## Goal

- `[orchestration.capabilities]` — concrete model id → capability class, with
  built-in defaults for the three known Anthropic tiers (opus=frontier,
  sonnet=capable, haiku=limited), so opt-in is one line: name a driver model.
- `[orchestration.capability_profiles]` — optional override of class → default
  profile.
- A **driver model** selector: `centinela start --model <id>` → `CENTINELA_MODEL`
  env → `[orchestration] driver_model` config → none.
- A new tier in `EffectiveProfile`: capability default from the pinned driver
  model, below the two explicit sources, above the strict default.
- `centinela status` surfaces the driver model and profile provenance.

## Non-goals (v1)

- **No change to gates or claim verification.** Capability moves process only.
- **No live re-resolution.** The driver model + derived profile are *pinned* at
  start for reproducibility; changing config mid-feature does not retro-rewrite.
- **No per-role capability.** Enforcement profile is per-*workflow*, so one
  driver model keys it. Per-role model routing stays `configurable-model-routing`.
- **No telemetry-based calibration.** Measuring whether the chosen profile was
  right is `capability-calibration` (Phase 7), which depends on this feature.
- **No auto-detection of the runner/model at runtime.** The directive hook has
  no runtime model signal; the driver model is declared, not sniffed.

## Dependencies

- `enforcement-profiles` (shipped) — profiles, `ProfileDefaults`,
  `EffectiveProfile`, the `--profile` pin, the back-compat guarantee.
- `configurable-model-routing` (shipped) — concrete model ids, runner keys, the
  config-leaf/orchestration-domain split and its parity-test pattern.
