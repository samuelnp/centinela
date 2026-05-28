### Gatekeeper Report: roadmap-parallel-readiness
**Date:** 2026-05-28
**Status:** WARNING

#### Analyzed Specs
Targeted scan of the specs most likely to depend on the contracts this feature
changed (`roadmap.Feature` schema, `Load()` now validating deps, Option B in
`analysis.go`, `RenderSessionRehydration` plural signature, `planadvisor` and
`docgen` read-paths). Full executable suite (`go test ./...` — 21 packages,
0 failures) and the coverage gate (95.1% ≥ 95.0%) both already PASS, so this
analysis focuses on spec-prose contracts that runtime tests do not encode.

Reviewed:
- `specs/roadmap-parallel-readiness.feature` (new)
- `specs/session-context-rehydration.feature`
- `specs/roadmap-senior-pm-analysis.feature`
- `specs/enrich-plan-advisor-context.feature`
- `specs/roadmap-checkpoint-prompt.feature`
- `specs/clarify-roadmap-missing-artifacts.feature`
- `specs/roadmap-quality-overall-threshold.feature`
- `specs/fix-roadmap-write-blocked.feature`

Also confirmed:
- `git diff --name-only main -- docs/architecture/ internal/scaffold/assets/`
  returned empty → scaffold-mirror parity unaffected.
- `centinela validate` reports "All gates passed" (G1 file size ✓, tests ✓,
  coverage ✓) — no runtime regression on any existing acceptance spec.

#### Findings

**Finding 1 — Rehydration spec prose now describes a contract the implementation no longer satisfies**
- **Affected spec:** `specs/session-context-rehydration.feature`
- **Affected scenarios:**
  - "SessionStart injects the rehydration payload on each supported source"
    (Scenario Outline, ~lines 55–72): asserts `the output should name the
    next feature to plan as "next-feature"`. The implementation now emits a
    plural `Ready to start now:` frontier listing every ready feature, not a
    single "next feature to plan" line.
  - "Next feature is the first incomplete across all phases, not just
    Phase 0" (~lines 75–81): asserts a single-next, declaration-order
    behavior. `FirstIncomplete` is no longer the source of truth for what
    SessionStart emits; readiness is derived from `dependsOn` + status.
- **Risk:** Spec prose contradicts the shipped behavior; a future reader (or
  LLM advisor) treating this spec as authoritative will encode the wrong
  contract.
- **Suggestion:** In the docs step (or a small follow-up), refresh these two
  scenarios to assert the plural `Ready to start now:` frontier (carry the
  pointer-PATH and no-inlining clauses forward — those still hold). The
  third scenario in that file ("Every roadmap feature done yields a
  graceful roadmap-complete state with no next feature") is still aligned
  and needs no change.

**Finding 2 — Senior-PM analysis spec still attributes cycle detection to the analysis JSON**
- **Affected spec:** `specs/roadmap-senior-pm-analysis.feature`
- **Affected scenario:** "Roadmap validate fails on dependency cycle"
  (~lines 11–14): `Given roadmap analysis JSON depends on a cyclic feature
  graph … Then validation should fail with cycle error`.
- **Risk:** Option B moved cycle/unknown-dep validation onto `roadmap.json`
  (`ValidateDependencies` inside `Load()`); the analysis JSON no longer
  carries `dependsOn`. The intent (a cycle blocks operations) is preserved,
  but the spec attributes it to the wrong source. Tests in
  `internal/roadmap/dependencies_test.go` exercise the new location.
- **Suggestion:** Rephrase the Given to "Given roadmap.json declares a
  cyclic feature graph" (and update the second scenario "complete analysis"
  similarly — deps now live in roadmap.json, analysis carries qualitative
  rationale only).

**Finding 3 — Plan-advisor spec still references analysis-side dependencies**
- **Affected spec:** `specs/enrich-plan-advisor-context.feature`
- **Affected scenario:** "Advisor prioritizes dependency context before
  same-phase siblings" (~lines 6–9): `Given roadmap analysis defines
  dependencies for the active feature`.
- **Risk:** `internal/planadvisor/roadmap_context.go` now reads `dependsOn`
  from `roadmap.json`; the runtime behavior is unchanged (deps still drive
  advisor priority) but the spec attributes the data source to analysis.
- **Suggestion:** Replace the Given with "Given roadmap.json declares
  dependencies for the active feature".

**No findings** for `roadmap-checkpoint-prompt`, `clarify-roadmap-missing-artifacts`,
`roadmap-quality-overall-threshold`, `fix-roadmap-write-blocked` — these touch
roadmap setup/validation flows but do not encode the specific contracts that
changed.

#### Recommendation
- **WARNING.** The implementation is correct, all gates pass, and no
  scenario is functionally broken (qa-senior updated the executable tests
  to the new contracts in the corrective code-step pass). But three
  existing `.feature` files contain prose that contradicts the shipped
  behavior in 4 scenarios. Document the drift now and refresh the wording
  in the docs step (or schedule a small follow-up). Not BLOCKING because
  the change is the deliberate result of the Decision Record (Option B +
  plural rehydration) and runtime contracts are intact.
</content>
