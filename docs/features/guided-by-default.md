# Feature Brief: guided-by-default

**Slug:** `guided-by-default`
**Phase:** Phase 13: Lighter Centinela
**Archetype:** canonical (5 steps)
**Depends on:** `truthful-validators` (shipped)
**Date:** 2026-08-03

## Roadmap contract (verbatim)

> Flip the default enforcement profile strict to guided, slim the greenfield
> cascade to PROJECT.md + roadmap.json, delete the self-graded quality threshold
> gate (depends on truthful-validators)

## Source material

Phase 13 ("Lighter Centinela") is retrospective-driven: *verify the work, not
artifacts about the work; process weight proportional to risk*. The originating
`retrospective.md` is **no longer present in the tree** (it was an untracked
root file and has since been removed; `git log` on `main` is now at `458a054`
v0.55.0, several features past the retrospective). The items this feature
answers, as carried in the Phase 13 header and the roadmap entry:

- **WS2.1 / failure #1 — inverted ceremony-to-code ratio.** A cold start pays
  for PROJECT.md, ROADMAP.md, `.workflow/roadmap.json`,
  `roadmap-analysis.{md,json}`, `roadmap-quality.{md,json}` and a
  production-readiness prompt doc before one line of code is allowed.
- **WS2.3 / failure #3 — the self-graded quality gate is theater.** An agent
  writes `overall: 9` into `.workflow/roadmap-quality.json` and a `>= 9` check
  lets it through. The number is assigned by the same party it constrains, so it
  gates nothing and costs a full evaluator pass.
- **Failure #9 — rubber-stamp confirmations.** `every_step` produces four
  confirmation prompts per feature; a prompt asked every time is answered
  without reading.
- **WS2.4 — confirmation default `after_plan`.** See "Scope" below: this is not
  a separate change here, it *falls out* of the flip.

## Problem

Centinela's zero-config default is `strict`. Strict is calibrated for a
low-capability driver on an unknown codebase, but it is what every new project
and every dogfooding run gets. The cost lands entirely on **process** —
confirmation prompts, orchestration evidence bundles, and a six-rung greenfield
setup cascade — while the things that actually establish correctness (gates,
tests, the adversarial verifier, claim verification) are already constant across
all three profiles. The result is a framework whose *felt* weight is dominated by
paperwork that proves nothing, which pushes users toward disabling governance
wholesale instead of dialing process down and keeping proof up.

## THE LINE: process may relax, proof may not

`cmd/centinela/complete_validate_gates.go:11` states the invariant this feature
must not break:

```go
// Verification is CONSTANT across every profile — NO profile branch belongs
// here; profiles scale process, never proof.
```

This is verified as still true today. `workflow.ValidateArtifacts` has exactly
**one** profile-conditioned branch — `validateOrchestration`, which early-returns
unless `wf.OrchestrationMode == StrictOrchestrationMode`. Everything else
(`validateGatekeeper`, `validateProductionReadiness`, `validateTests`,
`validateDocsOutput`, `validatePlan`) and everything in `runValidateGates`
(`VerificationFresh` ×2, `executeValidation`, `runClaimVerification`) is
profile-blind. **This feature adds no profile branch to any of them.**

| Changed behavior | Side of the line |
|---|---|
| Tail default profile for new starts: strict → guided | PROCESS |
| Confirmation cadence `every_step` → `after_plan` (guided default) | PROCESS |
| Orchestration evidence bundle not required under guided | PROCESS — a record of *who* produced the work, not of whether it passes |
| Greenfield cascade rungs (roadmap analysis, roadmap quality, ROADMAP.md, production-readiness *prompt doc*) become advisory under guided/outcome | PROCESS |
| Self-graded `overall >= 9` threshold deleted | NEITHER — a self-assigned number was never evidence; removing theater is not relaxing proof |
| Gatekeeper report required + grounded verdict (`adversarial-v1`) | PROOF — **unchanged, all profiles** |
| `centinela validate` gates (file_size, import_graph, security, spec_traceability, build, docstring, coverage, fmt) | PROOF — **unchanged, all profiles** |
| Full test suite + double `VerificationFresh` + claim verification | PROOF — **unchanged, all profiles** |
| Production-readiness *gate* (`**Status:** BLOCKING` blocks `complete`) | PROOF — **unchanged, config-driven, never profile-driven** |
| Prewrite step-gating | PROCESS, but **unchanged** — only `outcome` drops it; strict and guided both keep it |

## Verified current state (all citations re-checked at `458a054`)

| Claim from the roadmap/brief | Status |
|---|---|
| `internal/workflow/profile.go:EffectiveProfile` falls back to strict | TRUE — `return config.ProfileStrict` (line 23), 4-tier precedence: per-feature pin → explicit global → driver-model capability → strict |
| `internal/config/profile_defaults.go` strict sets `ConfirmEveryStep` | TRUE — and guided already sets `ConfirmAfterPlan` + `RequireSubagentEvidence: false` |
| `cmd/centinela/hook_setup.go` enforces the greenfield cascade | TRUE — rungs: PROJECT.md → ROADMAP.md → roadmap.json → roadmap-analysis.{md,json} → roadmap-quality.{md,json} → production-readiness-prompt.md → checkpoint |
| `cmd/centinela/start_guard.go` enforces the cascade | TRUE — `workflowOrderForFeature` calls `ValidateAnalysis`, `ValidateQuality`, `HasBootstrapPhase`, `BootstrapComplete`, dependency guard |
| `internal/roadmap/quality.go qualityThreshold = 9` | TRUE — line 12. Enforced in **three** places: `ValidateQuality` (`Threshold != 9`), `quality_features.go:18` (`Overall < 9`), and `promote_scores.go:35` (`--scores` overall `< 9`) |
| `truthful-validators` reworked quality.go's shape errors | TRUE — two-stage decode (`quality_shape.go`, `quality_fields.go`) distinguishes absent / wrong-kind / wrong-type from out-of-range. **This must not regress.** |

### Already fixed / not needed — do not re-do

- **WS2.4 (`after_plan` default) needs no separate change.**
  `ProfileDefaults(guided).ConfirmationMode` is *already* `ConfirmAfterPlan`, and
  `hook_context_review_mode.go:effectiveConfirmationMode` already resolves
  explicit config → profile default. Flipping the default profile delivers WS2.4
  for free. This repo's own `centinela.toml` already sets
  `step_confirmation_mode = "after_plan"` explicitly.
- **The plan advisor is already at guided behavior for everyone.**
  `NormalizePlanAdvisorMode("")` returns `missing_info`, and
  `planadvisor.Directive` reads `cfg.Workflow.PlanAdvisorMode` directly. The
  `ProfileKnobs.PlanAdvisorMode` field is **never read by any production caller**
  — strict's documented `always` is inert. (Deferred, see below.)
- **Proof parity already has an acceptance test.**
  `tests/acceptance/enforcement_profiles_invariant_test.go` asserts a failing
  validate command blocks under all three profiles. This feature extends it
  rather than inventing a new mechanism.

## Blast radius of the default flip

Every workflow created since `enforcement-profiles` shipped carries
`OrchestrationMode` derived at start, but `wf.EnforcementProfile` is populated
**only** from an explicit `--profile` (`start.go`: `wf.EnforcementProfile =
decision.PinnedProfile`). A zero-config workflow therefore has an empty pin and
re-derives live through `EffectiveProfile` on every read. Changing only the tail
constant would silently loosen **in-flight** workflows mid-run.

**Decision — follow the `PlanContract` / `ValidateContract` precedent exactly.**
Back-compat is *state-dated, never clock-dated*: a new `ProfileContract` field
(`"guided-default-v1"`) is pinned on every workflow created after this ships;
`EffectiveProfile`'s tail returns guided **iff** that contract is pinned, and
strict otherwise. An unpinned (legacy or in-flight) workflow is untouched, and a
fresh workflow cannot dodge the new default by hand-editing state. No user
config is ever rewritten automatically; `centinela doctor` emits an advisory
telling projects that want the old behavior to pin it.

## Acceptance criteria

1. **AC1 — new starts default to guided.** With no `centinela.toml`, no
   `--profile` and no `driver_model`, a workflow created by `centinela start`
   resolves to `guided`; `centinela status` shows provenance `default`.
2. **AC2 — no in-flight workflow changes profile.** A workflow state file
   without `profileContract` resolves to `strict` under identical conditions.
3. **AC3 — explicit configuration still wins.** `--profile strict`, a global
   `enforcement_profile`, and a `driver_model` capability class each still
   outrank the new tail default, in that order.
4. **AC4 — proof parity.** For an identical tree, `strict` and `guided` produce
   an identical pass/fail outcome for: a failing `validate` command, a missing
   gatekeeper report, an ungrounded gatekeeper verdict, and a `BLOCKING`
   production-readiness report.
5. **AC5 — the threshold gate is gone, in every profile.** A feature scored
   `overall: 3` promotes and validates successfully; the *shape* and *range*
   validators from `truthful-validators` still reject `overall: 0`,
   `overall: 11`, a missing `scores` object, and a non-integer score.
6. **AC6 — slim greenfield cold start.** Under guided, a tree containing only
   `PROJECT.md` and a `.workflow/roadmap.json` with a `Phase 0: Bootstrap`
   entry completes `centinela start <phase-0-feature>` end to end.
7. **AC7 — strict keeps the full cascade.** The same tree under an explicit
   `enforcement_profile = "strict"` still refuses, naming the missing roadmap
   analysis artifact.
8. **AC8 — `start` still refuses what it should.** Under guided, `start` still
   refuses a Backlog finding, a Draft feature, a feature with unmet
   dependencies, a roadmap with no `Phase 0: Bootstrap`, and a non-bootstrap
   feature while bootstrap is incomplete.
9. **AC9 — Centinela does not downgrade itself.** This repo's `centinela.toml`
   carries `enforcement_profile = "strict"` explicitly, asserted by a test.

## Out of scope

- Any change to gate logic, test execution, claim verification, the verifier
  contract, or the production-readiness gate.
- The `outcome` profile's behavior (unchanged).
- Removing `.workflow/roadmap-quality.{md,json}` or the `--scores` CLI surface.
- Wiring the inert `ProfileKnobs.PlanAdvisorMode` / `.StepGating` fields
  (deferred).
- Automatic rewriting of any user's `centinela.toml`.

## Deferred findings

- `profile-plan-advisor-knob-inert`
- `profile-step-gating-knob-inert`
