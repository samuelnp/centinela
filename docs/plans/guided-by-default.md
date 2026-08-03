# Plan: guided-by-default

**Feature:** `guided-by-default` · **Phase 13: Lighter Centinela** ·
**Archetype:** canonical (plan → code → tests → validate → docs)
**Date:** 2026-08-03 · **Brief:** [docs/features/guided-by-default.md](../features/guided-by-default.md)
**Spec:** [specs/guided-by-default.feature](../../specs/guided-by-default.feature)

---

## 0. The invariant this plan is subordinate to

`cmd/centinela/complete_validate_gates.go:11-12`:

> Verification is CONSTANT across every profile — NO profile branch belongs
> here; profiles scale process, never proof.

**Every slice below is a PROCESS change. No slice adds a profile branch to a
gate, a test run, a verifier check, or claim verification.** Slice 4 exists
solely to make that mechanically checkable rather than asserted.

### Per-behavior process/proof ledger (authoritative for review)

| # | Behavior changed | Process or Proof | Where |
|---|---|---|---|
| 1 | Tail default profile for **new** starts: strict → guided | PROCESS | `start_resolve.go`, `profile.go` |
| 2 | Confirmation cadence becomes `after_plan` by default | PROCESS (falls out of #1 via `ProfileDefaults`) | none — already wired |
| 3 | Orchestration evidence bundle not required under guided | PROCESS (record of authorship) | none — already wired via `order.go` |
| 4 | Greenfield cascade rungs demoted to advisory under guided/outcome | PROCESS | `start_guard.go`, `hook_setup.go` |
| 5 | Self-graded `overall >= 9` threshold deleted, **all profiles** | Theater removal — neither side | `quality.go`, `quality_features.go`, `promote_scores.go` |
| 6 | Gatekeeper report + grounded `adversarial-v1` verdict | **PROOF — untouched** | `validate_gatekeeper.go` |
| 7 | `centinela validate` gate set + suite + coverage + fmt | **PROOF — untouched** | `complete_validate_gates.go` |
| 8 | `VerificationFresh` ×2, claim verification | **PROOF — untouched** | `complete_validate_gates.go` |
| 9 | Production-readiness gate (`BLOCKING` blocks) | **PROOF — untouched, config-driven** | `validate.go` |
| 10 | Prewrite step-gating | PROCESS, **unchanged** for strict *and* guided | `hookpolicy/prewrite.go` |

---

## 1. Scope

**In**

- A `ProfileContract` pin (`guided-default-v1`) on the workflow state, and a
  guided tail default that fires only for pinned workflows.
- Deletion of the `qualityThreshold = 9` gate in all three of its enforcement
  sites, in every profile.
- A `RequireRoadmapGrading` profile knob that makes the greenfield cascade's
  grading rungs advisory under guided/outcome and unchanged under strict.
- An explicit `enforcement_profile = "strict"` in this repo's `centinela.toml`.
- A `centinela doctor` advisory for projects that inherit the new default.
- Documentation of the new default and the process/proof line.

**Out** — see brief. Notably: no gate logic changes; no `outcome` changes; no
removal of the quality artifacts or `--scores`; no automatic user-config
rewrites; the two inert `ProfileKnobs` fields stay inert (deferred).

## 2. Dependencies & assumptions

- `truthful-validators` shipped: `quality_shape.go` / `quality_fields.go`
  two-stage decoding is present and **must not regress**. Slice 2 deletes only
  the threshold comparisons, never a shape or range check.
- `adversarial-validate-verifier` (`ValidateContract`) and
  `unified-plan-specialist` (`PlanContract`) supply the state-dated back-compat
  precedent copied verbatim in Slice 1.
- `EffectiveProfile`'s tiers 1–3 (per-feature pin → explicit global →
  driver-model capability) are correct and unchanged; only tier 4 moves.
- `.workflow/roadmap.json` is loaded by `roadmap.Load()` and does **not** require
  `ROADMAP.md`, so slimming the cascade cannot break roadmap reads.

## 3. Slices

Budget rule: **every file, including `_test.go`, stays ≤100 lines.** Files near
the cap today (`start_guard.go` 99, `hook_setup.go` 91, `profile.go` 34) get
extractions rather than growth.

### Slice 1 — Pin the profile contract; flip the tail default for new starts

*Smallest correct slice: makes the default flip real without touching any
in-flight workflow.*

| File | Change | Budget |
|---|---|---|
| `internal/workflow/profile_contract.go` | **new** — `ProfileContractGuidedDefault = "guided-default-v1"`, `(*Workflow).UsesGuidedDefault()`, doc comment stating state-dated back-compat | ≤45 |
| `internal/workflow/state.go` | +`ProfileContract string \`json:"profileContract,omitempty"\`` with the legacy-means-strict comment | 97 → ≤100 |
| `internal/workflow/order.go` | `NewWithOrder` sets `ProfileContract: ProfileContractGuidedDefault` | 65 → ≤70 |
| `internal/workflow/profile.go` | tail: `if wf.UsesGuidedDefault() { return ProfileGuided }; return ProfileStrict` | 34 → ≤45 |
| `internal/workflow/profile_provenance.go` | tail note `"default"` → `"default (guided)"` / `"default (strict, legacy workflow)"` | 37 → ≤50 |
| `internal/workflow/start_resolve.go` | `default:` branch → `ProfileGuided`; the no-capability fallback in the `DriverModel != ""` branch stays **strict** (a declared-but-unmapped model keeps maximum scaffolding, per `getting-started.md`) | 51 → ≤60 |
| `internal/workflow/profile_contract_test.go` | **new** | ≤90 |
| `internal/workflow/profile_test.go` | extend | ≤100 |

**Tests, both directions**
- Pinned contract + zero config → `guided`. ✅
- **No** pin (legacy state file) + zero config → `strict`. ❌-direction, the
  back-compat guard.
- `--profile strict` on a pinned workflow → `strict` (tier 1 wins).
- Explicit global `enforcement_profile = "strict"` + pinned workflow → `strict`,
  provenance `global` (tier 2 wins).
- `driver_model = "haiku"` (limited → strict) + pinned workflow → `strict`
  (tier 3 wins).
- `driver_model = "sonnet"` (capable → guided) → `guided` with capability
  provenance, not `default (guided)` — the two must stay distinguishable.
- `EffectiveProfile(nil, cfg)` still returns `strict` — keeps the existing
  `TestEP_UnconfiguredKeepsStrictBehavior` acceptance test honest for
  workflow-less callers.

**Risk:** two "guided" results with different provenance get conflated in
status output. Mitigated by the provenance assertions above.

---

### Slice 2 — Delete the self-graded quality threshold gate (all profiles)

| File | Change | Budget |
|---|---|---|
| `internal/roadmap/quality.go` | delete `const qualityThreshold`; delete the `q.Threshold != qualityThreshold` check (the field still decodes and is ignored, so existing artifacts stay valid) | 67 → ≤60 |
| `internal/roadmap/quality_features.go` | delete the `Overall < qualityThreshold` refusal; **keep** unknown-feature, `validateScoreRange`, empty-summary and coverage checks | 44 → ≤40 |
| `internal/roadmap/promote_scores.go` | delete the `Overall < qualityThreshold` refusal; keep `validateScoreRange` | 39 → ≤35 |
| `cmd/centinela/roadmap_promote.go` | `--scores` help text: drop "overall >= 9" | 99 → ≤100 |
| `cmd/centinela/start_guard_draft.go` | rewrite `draftStartError`: the refusal survives on its own merits (a draft has no agreed phase/definition), not on the deleted gate. **Must still contain the substring "draft"** — the acceptance suite asserts it | 13 → ≤25 |
| `cmd/centinela/roadmap_validate.go` | on success, print an **advisory** score summary (any feature under 9 listed, exit 0) | 34 → ≤50 |
| `internal/roadmap/quality_test.go`, `quality_features_shape_test.go`, `quality_shape_test.go`, `promote_scores` tests | update threshold expectations; **add** regression asserts that every shape/range error from `truthful-validators` still fires | each ≤100 |

**Tests, both directions**
- `overall: 3` → `ValidateQuality` passes, `--scores 3,3,3,3,3,3` promotes. ✅
- `overall: 0` and `overall: 11` → still rejected by range. ❌
- `"threshold": 7` (or absent) in `roadmap-quality.json` → accepted. ✅
- Missing `scores` object → still `required object field "scores" is missing`. ❌
- `"scores": []` → still reported as a JSON *array* shape fault, never as a
  range fault. ❌ (direct `truthful-validators` regression guard)
- `"overall": "nine"` → still a type fault naming the field. ❌
- An existing `.workflow/roadmap-quality.json` on disk is **not rewritten** by
  any command in this slice.
- `start` on a draft still refuses, and the message still says "draft". ❌

**Answer to "what replaces it":** nothing blocking. `centinela roadmap validate`
reports the scores advisorily; `roadmap promote --scores` still records them.
`start` still refuses Backlog findings, drafts, unmet dependencies, a missing
`Phase 0: Bootstrap`, and non-bootstrap features while bootstrap is incomplete —
none of which depend on a self-assigned number.

---

### Slice 3 — Profile-scoped greenfield cascade

| File | Change | Budget |
|---|---|---|
| `internal/config/profile_defaults.go` | +`RequireRoadmapGrading bool` — strict `true`, guided/outcome `false` | 41 → ≤50 |
| `internal/config/project_profile.go` | **new** — `ProjectDefaultProfile(cfg)`: explicit global → driver-model capability → `ProfileGuided`. For surfaces that have **no** workflow (`hook setup`, pre-`start` guards) | ≤40 |
| `cmd/centinela/start_guard.go` | `workflowOrderForFeature` takes the resolved profile; calls `ValidateAnalysis`/`ValidateQuality` only when `RequireRoadmapGrading`; otherwise emits one advisory line and continues. Keeps `HasBootstrapPhase`, `BootstrapComplete`, Backlog, draft and dependency refusals in **all** profiles | 99 → split, each ≤100 |
| `cmd/centinela/start_guard_cascade.go` | **new** — extracted advisory/required cascade decision, keeps the file above under budget | ≤60 |
| `cmd/centinela/start.go` | compute `workflow.ResolveStart(...)` **before** `resolveArchetypeOrder` so the profile is available to the guard (today `decision` is computed after) | ≤100 |
| `cmd/centinela/hook_setup.go` | rungs 4/5/6 (roadmap analysis, roadmap quality, production-readiness prompt doc) early-return only under `RequireRoadmapGrading`; otherwise one consolidated advisory line and continue to the checkpoint. Rungs 1–3 (PROJECT.md, ROADMAP.md, roadmap.json) unchanged in shape — `ROADMAP.md` demoted to advisory under guided | 91 → split, each ≤100 |
| `cmd/centinela/hook_setup_cascade.go` | **new** — rung list derived from the profile | ≤60 |
| new `_test.go` per changed file | | each ≤100 |

**Required in every profile:** `PROJECT.md` and a parseable
`.workflow/roadmap.json`.
**Advisory under guided/outcome:** `ROADMAP.md`, `roadmap-analysis.{md,json}`,
`roadmap-quality.{md,json}`, `docs/architecture/production-readiness-prompt.md`.
**Never touched:** the production-readiness *gate* in
`workflow.validateProductionReadiness` — it is config-driven and stays proof.

**Tests, both directions**
- Guided cold start: tree with only `PROJECT.md` + `roadmap.json` (with
  `Phase 0: Bootstrap`) → `centinela start <phase-0-feature>` succeeds. ✅
- Same tree, `enforcement_profile = "strict"` → refuses, naming the roadmap
  analysis artifact. ❌
- Guided + Backlog slug → refuses. ❌
- Guided + draft → refuses, message contains "draft". ❌
- Guided + unmet dependency → refuses naming the dependency. ❌
- Guided + roadmap with no `Phase 0: Bootstrap` → refuses. ❌
- Guided + non-bootstrap feature while bootstrap incomplete → refuses. ❌
- `hook setup` on a guided project missing analysis/quality → prints an
  advisory and still reaches the checkpoint (does not early-return). ✅
- `hook setup` on a strict project missing analysis → early-returns with the
  existing directive, byte-identical to today. ❌-direction regression guard.

**Risk:** `start_guard.go` and `hook_setup.go` are both near the 100-line cap;
growth is a G1 violation. Mitigated by the two new extraction files.

---

### Slice 4 — Proof parity, made mechanical

*This slice is the reason the flip is safe to ship. It asserts the invariant
rather than trusting it.*

| File | Change | Budget |
|---|---|---|
| `tests/acceptance/guided_by_default_parity_test.go` | **new** — table over `{strict, guided}`: identical tree ⇒ identical outcome for (a) failing `validate` command, (b) missing gatekeeper report, (c) ungrounded gatekeeper verdict, (d) `**Status:** BLOCKING` production readiness | ≤100 |
| `tests/acceptance/guided_by_default_parity_helpers_test.go` | **new** — shared fixture builder | ≤100 |
| `tests/acceptance/enforcement_profiles_invariant_test.go` | extend the existing three-profile gate assertion with the gatekeeper-report cases | ≤100 |
| `cmd/centinela/complete_validate_gates_invariant_test.go` | **new** — source-level guard: `complete_validate_gates.go` contains no `Profile` identifier, so a future edit that adds a profile branch fails a test rather than a review | ≤60 |

**Tests, both directions**
- Each of the four refusals fires under **both** profiles (proof preserved). ❌
- A clean tree passes under **both** profiles (no false parity via
  everything-fails). ✅
- Guided **does** skip the orchestration-evidence requirement while strict does
  not — the one legitimate divergence, asserted explicitly so parity is not
  over-claimed. ✅/❌

---

### Slice 5 — Self-governance, doctor advisory, docs

| File | Change | Budget |
|---|---|---|
| `centinela.toml` | add `enforcement_profile = "strict"` under `[workflow]`, with a comment naming the flip. `step_confirmation_mode = "after_plan"` stays (explicit knob outranks strict's `every_step`, deliberately) | — |
| `internal/doctor/check_profile_default.go` | **new** — INFO diagnosis when `.workflow/*.json` workflows exist and no explicit `enforcement_profile`: "the default is now guided; pin `enforcement_profile` to keep strict". Never ERROR, never auto-fixed | ≤70 |
| `internal/doctor/check_profile_default_test.go` | **new** | ≤90 |
| `internal/scaffold/assets/…/centinela.toml` | comment documenting the guided default and how to pin strict; the key stays **unset** so new projects actually get guided | — |
| `docs/guides/configuration-reference.md` | default column `strict` → `guided`; the capability table row `limited → strict` is unchanged | — |
| `docs/guides/configuration.md`, `docs/guides/getting-started.md` | update default statements | — |
| `docs/architecture/workflow-enforcement.md` (+ `internal/scaffold` mirror if listed) | add the process/proof ledger | — |
| `tests/acceptance/guided_by_default_self_governance_test.go` | **new** — this repo's `centinela.toml` carries an explicit `enforcement_profile`; fails if the pin is ever dropped | ≤60 |

**Tests, both directions**
- Doctor advisory fires on a project with workflows and no explicit profile. ✅
- Doctor advisory silent when the profile is explicit, and silent on a
  workflow-less directory. ❌
- Doctor advisory is INFO — `doctor.ExitError` stays false. ❌
- Repo `centinela.toml` parses and yields `strict`. ✅

**Note (scaffold mirror):** `docs/architecture/*` edits must be mirrored into
`internal/scaffold/assets`; only 8 docs are covered by the parity test, so check
the mirror explicitly for any file edited here.

---

## 4. Risks

| Risk | Impact | Likelihood | Mitigation |
|---|---|---|---|
| Existing projects silently lose enforcement | High | Medium | Slice 1's `ProfileContract` pin: in-flight and legacy workflows stay strict; only new starts flip. Slice 5's doctor advisory names the change. No user config is ever rewritten. |
| "Guided" is read as "less verification" by a future contributor | High | Low | Slice 4's parity table + the source-level guard on `complete_validate_gates.go`; ledger in brief, plan and `workflow-enforcement.md` |
| Deleting the threshold regresses `truthful-validators`' shape errors | High | Medium | Slice 2 deletes only the three `Overall < threshold` / `Threshold != 9` comparisons; every shape/range error gets an explicit regression test |
| Slimmed cascade breaks greenfield cold start | High | Medium | Slice 3's end-to-end cold-start acceptance test under guided; `HasBootstrapPhase` / `BootstrapComplete` retained (both derive from `roadmap.json`, zero extra ceremony) |
| `start_guard.go` / `hook_setup.go` exceed the 100-line cap | Medium | High | Extraction files planned up front (`*_cascade.go`) |
| Two distinct "guided" provenances conflated in status | Low | Medium | Provenance assertions in Slice 1 |
| Centinela's own dogfooding silently downgrades | High | Low | Slice 5 pin + self-governance test |
| Reordering `ResolveStart` before `resolveArchetypeOrder` in `start.go` changes error precedence (a bad `--profile` now reported before a bad archetype) | Low | Medium | `ResolveStart` normalizes rather than errors, and config load already rejects unknown profiles; assert the existing archetype-error tests still pass |

## 5. Rollout sequence

1. **Slice 1** — the flip itself, safely pinned. Shippable alone.
2. **Slice 2** — delete the threshold. Independent of 1; unconditional.
3. **Slice 3** — profile-scoped cascade. Depends on 1 (needs the resolved
   profile) and on 2 (the quality rung's blocking check is already gone).
4. **Slice 4** — parity tests. Must land in the same PR as 1 and 3.
5. **Slice 5** — self-governance, doctor, docs. Last; depends on all above.

Can wait (deferred): wiring `ProfileKnobs.PlanAdvisorMode` and `.StepGating` so
the knob struct stops describing behavior no caller implements.

## 6. Behavior summary

After this change, a project with no `centinela.toml` gets `guided` on every new
`centinela start`: one confirmation stop after the plan instead of four, no
required orchestration-evidence bundle, and a greenfield cold start that needs
only `PROJECT.md` plus `.workflow/roadmap.json`. Workflows started before the
change keep `strict` byte-for-byte. No self-assigned quality score gates
anything, in any profile. Every gate, the full test suite, the double freshness
check, claim verification, the adversarial verifier's grounded verdict and the
production-readiness gate behave **identically** under `strict` and `guided` —
asserted by an acceptance table, not by convention. Centinela's own repository
pins `strict` explicitly so its dogfooding is unchanged.

## 7. UX states (CLI)

| State | Trigger | Surface |
|---|---|---|
| success | `centinela start` on a zero-config project | `Workflow started` + `Profile guided (default)` in `centinela status` |
| success (legacy) | `centinela status` on a pre-flip workflow | `Profile strict (default, legacy workflow)` |
| advisory | guided project missing roadmap analysis/quality | one-line `hook setup` advisory; the cascade continues |
| advisory | `centinela doctor` on a project inheriting the new default | INFO diagnosis, exit code unchanged |
| advisory | `centinela roadmap validate` with a feature scored below 9 | listed in the score summary, exit 0 |
| error | guided `start` on a Backlog / draft / dependency-blocked feature | unchanged refusal messages |
| error | strict project missing roadmap analysis | unchanged directive |
| error (proof) | any profile: failing gate, missing/ungrounded gatekeeper report, BLOCKING readiness | unchanged refusal, identical across profiles |
| empty | no workflows present | doctor advisory suppressed |

## 8. Handoff

**Next role:** `senior-engineer`.

Open questions for implementation:
1. Confirm whether `status` should render `default (guided)` or keep the bare
   `default` note — the plan assumes the annotated form; if `ui` snapshot tests
   pin the exact string, update them rather than reverting the annotation.
2. `hook_setup.go`'s `ROADMAP.md` rung: the plan demotes it to advisory under
   guided since `ROADMAP.md` is generated from `roadmap.json`. If the
   `roadmap_drift` gate's warn-mode output becomes noisy for cold-start
   projects with no `ROADMAP.md` at all, prefer auto-suggesting
   `centinela roadmap generate` in the advisory over re-promoting the rung.
