### Feature-Specialist Report: deterministic-artifact-scaffolds

**Date:** 2026-06-12

#### Behavior Summary

`centinela evidence init` and `centinela artifact new` become shape-complete by
construction. For the two plan roles (big-thinker, feature-specialist), `evidence
init` stamps `inputs` with `RequiredPlanInputs(feature)` — every `docs/features/*.md`
plus `docs/plans/<feature>.md` — so the agent never hand-loops ~80 paths and the
big-thinker / feature-specialist plan-snapshot rule passes with zero manual
`append`. All other roles get empty `inputs`. `outputs` and `edgeCases` stay empty
for every role (genuine fill slots; pre-filling `outputs` would fail the real-file
validator). Pre-fill is applied in the `runEvidenceInit` command path via
`evidence.PlanInputs(feature, role)`, never inside `Skeleton`, so the repair and
docs-specialist templates are not poisoned. Markdown bodies (companion `.md` and
`artifact new` `.md`) swap italic-prose placeholders for an explicit fill marker
(`evidence.FillSlot`), and the gatekeeper artifact pre-fills its "Analyzed Specs"
section by globbing the existing `specs/*.feature`. No fill marker ever lands in an
evidence JSON list field; `**Status:**` / `**Date:**` lines are unchanged so
`centinela validate` still parses them.

#### Gherkin Scenarios

The executable contract lives at `specs/deterministic-artifact-scaffolds.feature`.
Each scenario maps 1:1 to a Go test (`// Scenario: <title>`).

- Value scenario — *Init pre-fill lets big-thinker pass plan-snapshot validation
  with zero appends*: `evidence init demo big-thinker` then `evidence validate demo`
  passes the big-thinker plan-snapshot rule without any manual `evidence append …
  inputs`. Mirrored for feature-specialist.
- *Init pre-fills plan-snapshot inputs for big-thinker / feature-specialist*:
  `inputs` equal `RequiredPlanInputs(feature)` (every `docs/features/*.md` +
  `docs/plans/<feature>.md`).
- *Init leaves inputs empty for every non-plan role* (scenario outline): senior-
  engineer, ux-ui-specialist, qa-senior, validation-specialist,
  documentation-specialist, gatekeeper → empty `inputs`.
- *Skeleton stays empty so repair and docs templates are not poisoned*:
  `Skeleton`, `SchemaSkeleton` (repair), and the `docsSpecialistPair` keep empty
  `inputs`; pre-fill is command-path only.
- *PlanInputs is the only source shared with the validator* /
  *PlanInputs returns nil for a non-plan role*.
- *Init leaves outputs empty* and *Init leaves edgeCases empty* for every role.
- *Init pre-fill is idempotent under force re-run* and *…includes a feature brief
  created after the first init* (`--force`, sorted, de-duplicated).
- *FillSlot renders the canonical marker*: `FillSlot` returns the canonical
  `<` + `FILL:` substance-slot string (markdown only).
- *Companion skeleton seeds role-appropriate FILL slots* (outline, per-role header)
  and *Unknown role falls back to the one-line companion placeholder*.
- *No FILL marker ever lands in an evidence JSON list field*.
- *Gatekeeper artifact pre-fills Analyzed Specs from existing specs* and the
  *empty-list when no specs exist* counterpart.
- *Artifact bodies use FILL slots for substance sections* and *Artifact Status and
  Date lines stay parseable by validate*.
- Back-compat — *Pre-existing minimal evidence JSON still validates*.

#### UX States

This feature is a CLI; the surfaces are `centinela evidence init` / `artifact new`
terminal output plus the generated `.workflow/*.json` / `.workflow/*.md` /
artifact files. Loading is n/a (synchronous, sub-second).

| State   | Trigger                                                                 | Surface |
|---------|-------------------------------------------------------------------------|---------|
| loading | n/a — synchronous file write                                            | n/a |
| empty   | plan role in a repo with only its own brief; or gatekeeper with no `specs/*.feature` | `inputs` = the single brief + plan path; "Analyzed Specs" renders an empty list (no placeholder rows) |
| error   | role file already exists without `--force`                              | existence-guard message on stderr, non-zero exit; nothing overwritten |
| success | `init`/`artifact new` writes a shape-complete pair                       | "wrote .workflow/<feature>-<role>.json and companion .md"; pre-filled `inputs` + fill slots in the `.md` |

#### Out-of-Scope

- No `outputs` pre-fill in evidence JSON and no `PredictedOutputs` API (`outputs`
  stays a fill slot; pre-seeding non-existent files fails the real-file validator).
- No fill marker in any JSON list field (markdown bodies only).
- No `--minimal` opt-out; no profile/model-capability gating (pre-fill is
  unconditional).
- No schema change; no new overwrite flag — `--force` remains the only overwrite
  path.
- No loosening of the orchestration validator; pre-filled `inputs` pass by
  construction (same `RequiredPlanInputs` source).

#### Handoff

- Next role: senior-engineer
- Locked answers to the big-thinker open questions:
  1. Companion section headers (aligned to the live `*-prompt.md` docs):
     big-thinker = Problem / Scope / Dependencies & Assumptions / Risks / Rollout /
     Handoff; feature-specialist = Behavior Summary / Acceptance Criteria (Gherkin) /
     UX States / Edge Cases / Out-of-Scope / Handoff; senior-engineer = Files Touched /
     Architecture Compliance / Type-Safety Notes / Trade-Offs / Handoff;
     ux-ui-specialist = Flow Review / Accessibility / Visual Hierarchy / State
     Coverage / Handoff; qa-senior = Test Inventory / Coverage Gaps / Acceptance
     Wiring / Handoff; validation-specialist = Gates Run / Synthesis / Decision;
     gatekeeper = Analyzed Specs / Findings / Recommendation;
     documentation-specialist = KB Pages / project-docs Entries / Outcome.
     The `.feature` outline pins one representative header per role for the test.
  2. Gatekeeper "Analyzed Specs" globs ALL `specs/*.feature` (gatekeeper reviews
     cross-spec conflicts); empty list when none.
  3. `--force` is the only overwrite path (no new flag).
- Open clarifications: none — design was settled by big-thinker; this report is the
  acceptance contract.
