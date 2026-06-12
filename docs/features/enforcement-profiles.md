# Feature: enforcement-profiles

- surface: internal
- status: planned
- roadmap: Phase 4 — Loop Velocity (capability-spectrum keystone)
- fixes: one-size enforcement either burdens a strong model with ceremony or under-scaffolds a weak one — both end in governance being switched off

## Problem

Centinela enforces one fixed amount of process: step-gated writes, stop-and-ask
between every step, and (in strict orchestration) seven required subagent
evidence files per feature. That is right for a small local model that needs
maximum rails, but it taxes a frontier model that no longer does — and it
under-serves a quick fix that doesn't warrant the ceremony. When the process
cost exceeds the perceived benefit, users disable Centinela entirely. (We
measured this on ourselves: the dogfood note in the roadmap records ~370K tokens
and five confirmations for a ~250-line internal fix.)

## The core idea: decouple two axes

Centinela conflates two independent things:

1. **How much *process* is enforced** — step-gating, confirmations, mandatory
   subagent evidence. This SHOULD scale with model capability and task size.
2. **Whether *outcomes* are verified** — all validate gates + claim
   verification at completion. This must NEVER vary: no model's claims are
   trusted, regardless of capability.

`enforcement-profiles` introduces named presets that turn the first axis up or
down while the second stays welded on.

## The three profiles

| Profile | Step-gating (prewrite) | Stop-and-ask | Subagent evidence | Gates + claim verification |
|---------|:---:|:---:|:---:|:---:|
| **strict** (default) | ON (block out-of-step writes) | every step | required | **constant** |
| **guided** | ON | after plan | NOT required | **constant** |
| **outcome** | OFF (write in any order) | suppressed (auto) | NOT required | **constant** |

- **strict** — maximum scaffolding, for a small/local model that needs physical
  rails. The most rigorous mode, and the **back-compat default**: today's
  behavior is exactly this row (gating + every-step confirmation + mandatory
  subagent evidence), so an unconfigured project upgrades with no change.
- **guided** — a lighter opt-in: rails stay on (writes still step-gated) but the
  seven-file subagent ceremony is not mandatory and review prompts only after
  planning. For a capable model driving the steps itself.
- **outcome** — a capable agent works freely and fast: writes in any order, no
  inter-step prompts, no mandatory subagent ceremony — but `centinela complete`
  / merge still runs every gate and claim verification green before anything
  ships. "Verification stays constant" is the whole point.

## Hard invariant (must not regress)

`internal/verify/verify.go` and `internal/gates/gates.go` are NOT modified by
this feature. The validate-step hard block in `complete.go` (gates +
`verify.Verify`) runs identically under all three profiles. A profile can only
relax *process*, never *verification*.

## Goal

- A `[workflow] enforcement_profile = "strict" | "guided" | "outcome"` global
  setting, plus a per-feature override via `centinela start --profile <p>`
  (persisted in the feature's workflow state).
- The prewrite hook, confirmation prompt, and orchestration-evidence
  requirement all read the *effective* profile (per-feature override, else
  global, else the back-compat default).
- Zero behavior change for existing projects that set neither the profile nor
  the underlying knobs.

## Non-goals (v1)

- **No change to what the gates or claim verification check.** Outcome
  verification is constant by definition.
- **No collapse of the 5-step state machine.** Outcome mode relaxes ordering
  and ceremony; it does NOT remove the per-step artifact existence checks
  (a plan file, a spec, tests, a gatekeeper report are still produced). Folding
  the intermediate gates into a single ship-gate is a possible future
  refinement, explicitly out of scope here.
- **No new model-awareness.** Profiles are chosen by config/flag here;
  auto-selecting a profile from a model's declared capability is the dependent
  feature `model-capability-profiles`, not this one.
- **No removal of existing knobs.** `step_confirmation_mode`, `plan_advisor_mode`
  remain; a profile sets their *defaults*, an explicit knob still wins.
