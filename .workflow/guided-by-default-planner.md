# guided-by-default — planner

**Date:** 2026-08-03 · **Role:** planner (planner-v1, both lenses)

## Problem

Centinela's zero-config default enforcement profile is `strict`, calibrated for a
low-capability driver on an unknown codebase — yet it is what every new project
and every dogfooding run inherits. Its entire cost lands on **process**: four
confirmation prompts per feature, a required orchestration-evidence bundle, and a
six-rung greenfield setup cascade (PROJECT.md → ROADMAP.md → roadmap.json →
roadmap-analysis → roadmap-quality → production-readiness prompt) that must be
paid before a line of code may be written. Meanwhile the things that actually
establish correctness — gates, the full test suite, double verification
freshness, claim verification, the adversarial verifier's grounded verdict, the
production-readiness gate — are *already* constant across all three profiles. One
of those cascade rungs, the `overall >= 9` roadmap quality threshold, is pure
theater: the number is assigned by the same agent it constrains. The net effect
is a framework whose felt weight is dominated by paperwork that proves nothing,
which pushes users to disable governance wholesale instead of dialing process
down while keeping proof up.

## Scope

- **In:** a `ProfileContract` pin (`guided-default-v1`) plus a guided tail
  default that fires only for pinned workflows; deletion of `qualityThreshold`
  at all three enforcement sites in every profile; a `RequireRoadmapGrading`
  profile knob making the cascade's grading rungs advisory under guided/outcome;
  an explicit `enforcement_profile = "strict"` in this repo's `centinela.toml`;
  a `centinela doctor` advisory; proof-parity acceptance tests; docs.
- **Out:** any change to gate logic, test execution, claim verification, the
  verifier contract, or the production-readiness gate; `outcome`'s behavior;
  removal of the quality artifacts or the `--scores` CLI surface; wiring the
  inert `ProfileKnobs.PlanAdvisorMode` / `.StepGating` fields; any automatic
  rewrite of a user's `centinela.toml`.

## Dependencies & Assumptions

- `truthful-validators` (shipped) supplies the two-stage shape decoding in
  `quality_shape.go` / `quality_fields.go`. Slice 2 deletes only the three
  threshold comparisons and must not regress a single shape or range error.
- `adversarial-validate-verifier` (`ValidateContract`) and
  `unified-plan-specialist` (`PlanContract`) supply the state-dated back-compat
  precedent — empty pin means legacy means old behavior — copied verbatim.
- `EffectiveProfile` tiers 1–3 (per-feature pin → explicit global → driver-model
  capability) are correct and unchanged; only tier 4 moves.
- `roadmap.Load()` reads `.workflow/roadmap.json` and never requires
  `ROADMAP.md`, so slimming the cascade cannot break roadmap reads.
- `complete_validate_gates.go:11` — "Verification is CONSTANT across every
  profile — NO profile branch belongs here; profiles scale process, never proof"
  — is verified still true: `ValidateArtifacts` has exactly one
  profile-conditioned branch (`validateOrchestration`), and this feature adds none.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Existing projects silently lose enforcement | High | Medium | `ProfileContract` pin: legacy/in-flight workflows stay strict; only new starts flip; doctor advisory; no config is ever rewritten |
| "Guided" read as "less verification" by a future contributor | High | Low | Proof-parity acceptance table + source-level guard asserting `complete_validate_gates.go` contains no profile branch; ledger in brief, plan and docs |
| Threshold deletion regresses truthful-validators' shape errors | High | Medium | Only the three `Overall < threshold` / `Threshold != 9` comparisons are removed; every shape and range error gets an explicit regression test |
| Slimmed cascade breaks greenfield cold start | High | Medium | End-to-end guided cold-start acceptance test; `HasBootstrapPhase`/`BootstrapComplete` retained (both derive from roadmap.json, zero extra ceremony) |
| `start_guard.go` (99 lines) / `hook_setup.go` (91) exceed the 100-line cap | Medium | High | `start_guard_cascade.go` and `hook_setup_cascade.go` extractions planned up front |
| Two distinct "guided" provenances conflated in status output | Low | Medium | Provenance assertions distinguishing `default (guided)` from the capability tier |
| Centinela's own dogfooding silently downgrades | High | Low | Explicit `enforcement_profile = "strict"` pin plus a self-governance test that fails if the pin is dropped |
| Reordering `ResolveStart` before `resolveArchetypeOrder` changes error precedence | Low | Medium | `ResolveStart` normalizes rather than errors; existing archetype-error tests must still pass |

## Rollout

- **Step 1 — Slice 1:** `ProfileContract` pin + guided tail default for new
  starts. Smallest correct slice; shippable alone; touches no in-flight workflow.
- **Step 2 — Slice 2:** delete the quality threshold at all three sites,
  unconditionally, in every profile. Independent of Slice 1.
- **Step 3 — Slice 3:** `RequireRoadmapGrading` knob; cascade grading rungs
  become advisory under guided/outcome, unchanged under strict. Depends on 1 and 2.
- **Step 4 — Slice 4:** proof-parity acceptance table + the source-level
  no-profile-branch guard. Must land in the same PR as Slices 1 and 3.
- **Step 5 — Slice 5:** repo self-pin, doctor advisory, scaffold comment, docs.
- **Can wait (deferred):** wiring `ProfileKnobs.PlanAdvisorMode` and
  `.StepGating` so the knob struct stops describing behavior no caller implements.

### Process/proof ledger (authoritative for review)

| Behavior changed | Side of the line |
|---|---|
| Tail default for new starts: strict → guided | PROCESS |
| Confirmation cadence becomes `after_plan` | PROCESS (falls out of the flip) |
| Orchestration evidence bundle not required under guided | PROCESS — a record of *who* produced the work |
| Cascade grading rungs advisory under guided/outcome | PROCESS |
| Self-graded `overall >= 9` threshold deleted | Theater removal — neither side |
| Gatekeeper report + grounded `adversarial-v1` verdict | **PROOF — untouched** |
| `centinela validate` gate set, suite, coverage, fmt | **PROOF — untouched** |
| `VerificationFresh` ×2 and claim verification | **PROOF — untouched** |
| Production-readiness gate (`BLOCKING` blocks) | **PROOF — untouched, config-driven** |
| Prewrite step-gating | PROCESS, **unchanged** for strict *and* guided |

## Behavior Summary

After this change a project with no `centinela.toml` gets `guided` on every new
`centinela start`: one confirmation stop after the plan instead of four, no
required orchestration-evidence bundle, and a greenfield cold start needing only
`PROJECT.md` plus `.workflow/roadmap.json`. Workflows started before the change
keep `strict` byte-for-byte, because back-compat is state-dated on a pinned
`ProfileContract`, never clock-dated. No self-assigned quality score gates
anything, in any profile — `roadmap validate` reports scores advisorily and
`promote --scores` still records them, while `start` continues to refuse Backlog
findings, drafts, unmet dependencies, a missing bootstrap phase, and
non-bootstrap features while bootstrap is incomplete. Every gate, the full test
suite, both freshness checks, claim verification, the adversarial verifier's
grounded verdict and the production-readiness gate behave **identically** under
`strict` and `guided`, asserted by an acceptance table rather than by convention.
Centinela's own repository pins `strict` explicitly so its dogfooding is unchanged.

## Acceptance Criteria (Gherkin)

All scenarios live in `specs/guided-by-default.feature`.

- **AC1** *A new workflow in a zero-config project defaults to guided* — happy
  path for the flip.
- **AC2** *A workflow started before the flip keeps strict* — the blast-radius
  negative.
- **AC3** *An explicit per-feature / global / capability profile still outranks
  the new default* — three precedence negatives, plus *A capability-derived
  guided profile is distinguishable from the default*.
- **AC4** *Gate failures block completion identically under strict and guided*,
  *A missing verifier report…*, *An ungrounded verifier verdict…*, *A blocking
  production-readiness report…* — four `Scenario Outline`s over both profiles,
  the proof-parity core; paired with *A tree that satisfies every gate completes
  under every profile* so parity is not achieved by everything failing, and with
  *Guided relaxes process, and only process*, which names the single legitimate
  divergence (the orchestration evidence bundle) so parity is not over-claimed.
- **AC5** *A low self-assigned quality score no longer blocks anything*,
  *Promoting a feature with a low overall score succeeds*, *The declared
  threshold field is no longer enforced*, plus five negatives (*out-of-range
  scores*, *missing scores object*, *wrong JSON kind*, *non-integer score*,
  *existing artifacts survive untouched*) — the truthful-validators regression
  guard.
- **AC6/AC7** *A guided greenfield project cold-starts from PROJECT.md and
  roadmap.json alone* against *A strict greenfield project still requires the
  full cascade*, plus the two setup-hook scenarios.
- **AC8** *Guided still refuses what the roadmap says it must* (five-case
  outline), *A missing roadmap json is still refused under guided*, and *The
  production-readiness gate is untouched by cascade slimming*.
- **AC9** *Centinela's own repository pins its profile explicitly* and the two
  doctor-advisory scenarios.

## UX States

| State | Trigger | Surface |
|-------|---------|---------|
| success | `centinela start` on a zero-config project | `Workflow started` + `Profile guided (default)` in `centinela status` |
| success (legacy) | `centinela status` on a pre-flip workflow | `Profile strict (default, legacy workflow)` |
| advisory | guided project missing roadmap analysis/quality | one-line `hook setup` advisory; cascade continues to the checkpoint |
| advisory | `centinela doctor` on a project inheriting the new default | INFO diagnosis; exit code unchanged |
| advisory | `centinela roadmap validate` with a feature scored below 9 | listed in the score summary; exit 0 |
| error | guided `start` on a Backlog / draft / dependency-blocked feature | unchanged refusal messages |
| error | strict project missing roadmap analysis | unchanged directive |
| error (proof) | any profile: failing gate, missing or ungrounded gatekeeper report, BLOCKING readiness | unchanged refusal, identical across profiles |
| empty | no workflows present | doctor advisory suppressed |

## Edge Cases

1. A workflow state file with no `profileContract` (legacy or in-flight) must
   resolve to `strict`, not guided — back-compat is state-dated, never clock-dated.
2. A hand-edited state file cannot dodge the flip: absence of the pin only ever
   yields the *stricter* outcome.
3. A `driver_model` that maps to no capability class keeps resolving to `strict`,
   not to the new guided tail default — a declared-but-unmapped model must keep
   maximum scaffolding.
4. `EffectiveProfile(nil, cfg)` (workflow-less callers) must keep returning
   `strict`, so the existing unconfigured-project acceptance test stays honest.
5. Capability-derived `guided` and default-derived `guided` must remain
   distinguishable in status provenance.
6. `"threshold": 7` or an absent `threshold` field in an existing
   `roadmap-quality.json` must now be accepted, not rejected.
7. `overall: 0` and `overall: 11` must still be refused by the range check after
   the threshold check is deleted.
8. `"scores": []`, a missing `scores` object, and `"overall": "nine"` must still
   produce shape/type faults naming the field — never an out-of-range message.
9. Existing `.workflow/roadmap-quality.{md,json}` files must not be rewritten,
   migrated, or invalidated by any command in this feature.
10. The draft-start refusal must survive the threshold deletion on its own
    merits and its message must still contain the substring "draft" (the
    acceptance suite asserts it).
11. A guided greenfield tree with no `ROADMAP.md` at all must still cold-start;
    `ROADMAP.md` is generated from `roadmap.json`.
12. A guided project must still be refused for a missing or unparseable
    `.workflow/roadmap.json` — that rung is required in every profile.
13. `hook setup` under strict must early-return byte-identically to today.
14. `start_guard.go` and `hook_setup.go` must not cross the 100-line cap; the
    extraction files are part of the change, not a follow-up.
15. Reordering `ResolveStart` ahead of `resolveArchetypeOrder` in `start.go`
    must not change which error an operator sees for a bad archetype.
16. The production-readiness *gate* must remain config-driven; only its *setup
    prompt-doc rung* is demoted.

## Out-of-Scope

- Gate logic, test execution, claim verification, verifier contract, and the
  production-readiness gate — all untouched.
- The `outcome` profile's behavior.
- Deleting `.workflow/roadmap-quality.{md,json}` or the `roadmap promote
  --scores` surface.
- Wiring the inert `ProfileKnobs.PlanAdvisorMode` / `.StepGating` fields
  (deferred as `profile-plan-advisor-knob-inert` and
  `profile-step-gating-knob-inert`).
- Automatic rewriting of any user's `centinela.toml`.
- WS2.4 (`after_plan` confirmation default) as separate work: it is already
  `ProfileDefaults(guided).ConfirmationMode` and falls out of the flip for free.

### Deferred Findings

- `profile-plan-advisor-knob-inert` — `ProfileKnobs.PlanAdvisorMode` is never
  read by any production caller; `planadvisor.Directive` reads
  `cfg.Workflow.PlanAdvisorMode` directly and `NormalizePlanAdvisorMode`
  already defaults to `missing_info`, so strict's documented `always` is inert.
- `profile-step-gating-knob-inert` — `ProfileKnobs.StepGating` is never read;
  `hookpolicy/prewrite.go` branches on `EffectiveProfile == ProfileOutcome`
  directly, so a future profile with `StepGating=false` would silently still
  step-gate.

## Handoff

- **Next role:** `senior-engineer`.
- Outstanding questions:
  1. Should `status` render `default (guided)` or keep the bare `default` note?
     The plan assumes the annotated form; if `ui` snapshot tests pin the exact
     string, update them rather than reverting the annotation.
  2. `hook_setup.go`'s `ROADMAP.md` rung is demoted to advisory under guided
     because `ROADMAP.md` is generated from `roadmap.json`. If the
     `roadmap_drift` warn-mode gate becomes noisy for cold-start projects with
     no `ROADMAP.md` at all, prefer auto-suggesting `centinela roadmap generate`
     in the advisory over re-promoting the rung.
