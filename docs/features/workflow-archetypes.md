# Feature: workflow-archetypes

- surface: internal
- status: planned
- roadmap: Phase 4 — Loop Velocity
- fixes: forcing a diagnosis/bugfix/exploration through a feature-shaped plan→docs pipeline it doesn't fit

## Problem

Centinela has exactly one workflow shape: plan → code → tests → validate → docs.
That fits a net-new product feature, but real engineering also does **bugfixes**
(reproduce → fix → ship, no design doc), **refactors** (change structure, prove
behavior unchanged, no user-facing docs), and **spikes** (timeboxed exploration
you may throw away). Forcing those through the 5-step ceremony — a Gherkin spec
for a one-line fix, a KB doc + HTML portal for an internal refactor, a full
validate gate for a throwaway prototype — is friction that pushes people to skip
Centinela for exactly the work where a light rail would still help.

## The idea: named tracks that reuse the canonical steps

An **archetype** is a named preset that selects a *subset and ordering of the
existing canonical steps* (`plan`, `code`, `tests`, `validate`, `docs`). It is
NOT a set of new step names — reusing the canonical names means every existing
mechanism (step-gating file-type matrix, required-role orchestration, per-step
artifact validation, the ship gate) works unchanged.

| Archetype | Step order | Drops | Ship-gated? |
|-----------|------------|-------|:-----------:|
| **canonical** (default) | plan → code → tests → validate → docs | — | yes |
| **hotfix** | code → tests → validate | plan, docs | yes |
| **refactor** | plan → code → tests → validate | docs | yes |
| **spike** | plan → code | tests, validate, docs | **no** |

- **hotfix** — urgent fix: reproduce/fix in `code`, prove with `tests`, ship
  through the `validate` gate. No upfront design ceremony, no docs portal.
- **refactor** — restructure with a `plan` (what/why), change in `code`, prove
  equivalence with `tests`, `validate`. No user-facing docs (internal change).
- **spike** — timeboxed exploration: a light `plan`, then `code`. **No `validate`
  step at all**, so no ship gate — explore freely, throw away or promote later.

Selected at `centinela start --archetype <name>`, or per-feature via a roadmap.json
`archetype` field; pinned in the feature's workflow state. Orthogonal to
enforcement-profiles (which scales strictness): any archetype runs under any
profile.

## The safety property (why spike is not a hole)

The ship gate in `complete.go` fires on `current == "validate"` — it keys on the
**step**, not on any archetype label. So:

- An archetype that **includes** `validate` (canonical, hotfix, refactor) is
  gated identically — gates + claim verification run, the enforcement-profiles
  invariant is untouched.
- `spike` simply **has no `validate` step**, so the gate never fires. This is not
  a bypass branch — there is no `if archetype == "spike" skip gate` code. A
  feature cannot relabel itself to dodge verification, because verification is
  attached to the step, not the name.
- A spike you later decide to ship still passes through `validate` at merge time
  (the merge-steward validates), so "no ship gate" means "not expected to ship
  as-is," not "ships unverified."

## Goal

- `centinela start --archetype hotfix|refactor|spike|canonical` plus a
  roadmap.json per-feature `archetype` override; default = canonical (zero
  behavior change for everyone today).
- A single `ArchetypeStepOrder(name)` mapping consumed by the existing
  step-order selection seam (`workflowOrderForFeature`).
- The active archetype shown in `centinela status`.
- NO changes to the ship gate, the step-gating matrix, the role policy, or the
  per-step validators — archetypes compose them.

## Non-goals (v1)

- **No new step names** (reproduce/characterize/prove-equivalent). Evocative
  names are description; the step identifiers stay canonical. New named steps
  with their own validators are a possible later enhancement.
- **No coupling to enforcement-profiles.** An archetype sets the step *sequence*;
  a profile sets *strictness*. They remain independent.
- **No change to what any gate or claim verification checks.** Archetypes change
  which steps exist, never how a present step is verified.
- **No auto-detection** of archetype from the change (e.g. "this looks like a
  bugfix"). Chosen explicitly by flag/config.
