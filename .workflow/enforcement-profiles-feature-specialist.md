### Feature-Specialist Report: enforcement-profiles
**Date:** 2026-06-11

#### Behavior Summary
Centinela gains three named enforcement profiles — `strict`, `guided`, `outcome` — that scale only the *process* axis (prewrite step-gating, the stop-and-ask review prompt, and the mandatory subagent-orchestration evidence requirement) while the *verification* axis (validate-step gates + claim verification at `complete`) stays welded on under every profile. The default is `strict`, which reproduces today's behavior byte-for-byte: an unconfigured project upgrades with step-gating ON, `step_confirmation_mode` = `every_step`, and subagent evidence REQUIRED. `guided` keeps the rails (writes still step-gated) but drops the mandatory subagent ceremony and defaults confirmation to `after_plan`. `outcome` removes ordering rails (writes allowed in any active step), suppresses the review prompt (auto), and drops mandatory evidence — yet `complete` still runs gates + claim verification. Each governed knob resolves by precedence: an explicit knob in `centinela.toml` wins, else the effective profile's default, else the hardcoded default; the effective profile itself resolves per-feature override → global `[workflow] enforcement_profile` → the `strict` default.

#### Gherkin Scenarios
All map to `specs/enforcement-profiles.feature` (12 scenarios):
- **Outcome profile allows writing code during the plan step** — Given outcome + plan step, When prewrite hook evaluates a code write, Then allowed. → `EvaluatePrewrite`/`IsAllowedInStep` bypass under outcome.
- **Strict and guided profiles still block out-of-step writes** — Given strict + plan step, When prewrite evaluates a code write, Then blocked. → `EvaluatePrewrite` retains gating for strict/guided.
- **A write with no active workflow is always blocked** — Given no active workflow, When prewrite evaluates a plan/code write, Then blocked regardless of profile. → the no-active-workflow branch in `EvaluatePrewrite` is profile-independent.
- **Outcome profile suppresses the stop-and-ask review prompt** — Given outcome, When the review-prompt decision is made, Then no prompt. → `shouldRenderReviewPrompt` resolves outcome→auto.
- **An explicit confirmation mode overrides the profile default** — Given `step_confirmation_mode=every_step` + outcome profile, When the review-prompt decision is made, Then a prompt is rendered. → precedence resolver: raw-explicit knob beats profile default in `shouldRenderReviewPrompt`.
- **Strict profile requires subagent orchestration evidence** — Given strict, When the feature is created, Then orchestration mode requires evidence. → `NewWithOrder` sets `strict-subagents-v1` for strict.
- **Guided profile does not require subagent orchestration evidence** — Given guided, When created, Then orchestration mode does not require evidence. → `NewWithOrder` leaves mode empty for guided.
- **Outcome profile does not require subagent orchestration evidence** — Given outcome, When created, Then orchestration mode does not require evidence. → `NewWithOrder` leaves mode empty for outcome.
- **A per-feature profile overrides the global setting** — Given global=guided + feature started with outcome override, When effective profile resolved, Then outcome. → `EffectiveProfile(wf,cfg)`.
- **An unconfigured project keeps today's behavior (default strict)** — Given neither profile nor confirmation mode set, When effective knobs resolved, Then profile=strict, step-gating on, confirmation every_step, subagent evidence required. → `EffectiveProfile` default + `ProfileDefaults(strict)` + `NewWithOrder`.
- **Gates and claim verification run under every profile** — Given a feature on validate with a failing gate/claim, When completion attempted under strict/guided/outcome, Then blocked in every case. → complete-step gate (`executeValidation` + `runClaimVerification`) take no profile parameter.
- **An unknown profile value is rejected at config load** — Given a toml with an unsupported `enforcement_profile`, When loaded, Then load fails naming the field. → `validateEnforcementProfile` in `validateConfig`.

#### UX States
| State | Trigger | Surface |
|---|---|---|
| Write blocked (out-of-step) | prewrite hook under strict/guided, write not allowed in current step | prewrite hook block output (deny + reason) |
| Write allowed | prewrite hook under outcome, or in-step write under any profile | prewrite hook allow (no block) |
| No-workflow block | plan/code write with no active workflow, any profile | prewrite hook block output |
| Review prompt rendered | `shouldRenderReviewPrompt` true (every_step effective, or explicit knob) | the stop-and-ask review prompt after step artifacts |
| Review prompt suppressed | outcome (auto) or non-prompting effective mode | n/a (no prompt surfaced) |
| Effective profile visible | `centinela status <feature>` | status output line showing the effective profile |
| Config-load error | unknown `enforcement_profile` value | config load error naming the `enforcement_profile` field |

#### Out-of-Scope
- No change to what the gates or claim verification check — `internal/verify` and `internal/gates` are untouched; outcome verification is constant by definition.
- No collapse of the 5-step state machine — outcome relaxes ordering/ceremony only; per-step artifact existence checks (`ValidateArtifacts`) still run under all profiles.
- No model-capability auto-selection — profiles are chosen by config/flag here; auto-selecting from a model's declared capability is the dependent `model-capability-profiles` feature.
- No removal of existing knobs — `step_confirmation_mode` and `plan_advisor_mode` remain; a profile only sets their *defaults*, and an explicit knob still wins.

#### Handoff
- Next role: senior-engineer
- Open clarifications: none blocking. The senior-engineer must (a) capture the RAW decoded `step_confirmation_mode` before `applyDefaults` overwrites it so the precedence resolver can distinguish explicit-`every_step` from defaulted, (b) make `NewWithOrder` profile-aware (strict→`strict-subagents-v1`, guided/outcome→empty) and thread the effective profile to its callers, and (c) keep `outcome`'s relaxation scoped to the `IsAllowedInStep` ordering block only — `ValidateArtifacts` and the validate-step gate/verify path must take no profile parameter.
