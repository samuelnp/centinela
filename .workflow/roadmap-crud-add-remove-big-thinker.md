# Big-Thinker Report: roadmap-crud-add-remove

**Date:** 2026-07-01
**Feature:** 2 of 4 in the Roadmap Editing Suite (umbrella: `docs/plans/roadmap-editing-suite-design.md`)
**Handoff:** feature-specialist

## Problem

Centinela's roadmap can capture findings (`defer` into the validate-exempt Backlog)
and `promote` them behind a ≥9 quality gate, but there is no way to deliberately
*author* the roadmap: no `add`, no `remove`, no direct-finalize. An operator — and
the Magallanes Plan page consuming `roadmap --json` — needs to create a new feature
in a real phase, delete a planned one, and finalize a drafted item, all without
breaking `roadmap validate` (which demands an analysis entry and a quality entry
with overall ≥9 for every *schedulable* feature via the single coverage set in
`backlog.go` `NonBacklogFeatureSet`). Naively adding a scoreless feature to a real
phase makes `validate` fail and blocks greenfield starts. This feature delivers the
create/remove half plus the `Draft` lifecycle that makes direct authoring gate-safe.

## Scope

**In:**
- `Draft bool json:"draft,omitempty"` on `Feature` (`types.go`) and its hooks:
  the single coverage-set exemption in `NonBacklogFeatureSet`, `Summary()`
  exclusion, `readiness.go classifyFeature` → `State:"draft"` (excluded from
  `ReadySet`), `mdgen_feature.go` ` *(draft)*` marker, `start_guard` refusal.
- `internal/roadmap/draft.go` (`IsDraftFeature`, `DraftFeatures`).
- Generalized raw-feature helpers layered on `rawDoc`: `rawfeature_find.go`
  (`findFeature`, `featurePhase`), `rawfeature_mutate.go` (`appendFeatureToPhase`
  full-feature/dup-guarded, `removeFeatureAt` generalizing `removeBacklogFeature`,
  `replaceFeatureAt`), `rawtyped.go` (`toRoadmap`), `rawdeps.go`
  (`featureDependents`).
- `roadmap add` (creates a draft in a chosen schedulable phase; runs
  `ValidateDependencies` on the draft).
- `roadmap remove`/`rm` (guarded against depended-on and in-progress/done).
- Generalized `promote`: in-place draft finalize (clear flag + append
  analysis/quality artifacts, NO move) vs Backlog move-and-score, **branched by the
  slug's current location, not a flag**.
- `FeatureView.Draft` + `readiness:"draft"` in the JSON view (`view_types.go`,
  `view.go`) — the draft dimension feature 1 deliberately deferred.

**Out (successor features):** `edit`/`update`, `move`, `reorder`
(→ roadmap-edit-move); `phase add`/`rename`/`remove` (→ roadmap-phase-ops).

## Dependencies & Assumptions

- Builds on **roadmap-json-contract** (shipped): `view.go`/`view_types.go` are
  present; this feature *extends* `FeatureView` and `BuildView`, not creates them.
- Reuses the existing format-preserving raw I/O layer verbatim: `rawio.go`
  (`readRawRoadmap`/`writeRawRoadmap`/`writeAtomic`), `rawmutate.go`, `rawmove.go`
  (`appendToPhase`/`removeBacklogFeature` — generalize, do not fork),
  `rawrender.go` (dirty-index render contract).
- Reuses `ValidateDependencies` (`dependencies.go`), `FeatureStatus` (`roadmap.go`),
  `ParseScores` (`promote_scores.go`), and the promotion artifact append path
  (`promote_artifacts.go` → refactor into a shared `appendScoreArtifacts`).
- Assumes the existing non-schedulable predicate `isNonSchedulablePhase`
  (Backlog + Baseline) is the correct exclusion boundary; drafts live *inside*
  schedulable phases, so they are visible to `BuildView`/`DeriveReadiness` and the
  draft check must be applied per-feature there, not per-phase.
- Concurrency guarantee is unchanged: atomic temp+rename, one-feature-per-line,
  last-writer-wins (no locking upgrade).

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|-----------|
| Draft exemption leaks beyond the single `NonBacklogFeatureSet` hook (semantics drift; validate silently stops demanding artifacts for wrong features) | High | Med | Exactly ONE `if f.Draft { continue }` in `NonBacklogFeatureSet`; unit-test the coverage set both ways (draft excluded, non-draft included). Every other consumer derives draftness from the persisted `f.Draft` field, never re-implements the exemption. |
| Draft dimension handled inconsistently across the four readers (`NonBacklogFeatureSet`, `classifyFeature`/`ReadySet`, `Summary`, `buildFeatureView`) — e.g. code edits `BuildView` but not `classifyFeature`, so an unscored draft classifies as "ready", leaks into `ReadySet`, and `start` is allowed | High | Med | `classifyFeature` MUST check `f.Draft` FIRST → `State:"draft"` before the ready/blocked default; `buildFeatureView` must flow the "draft" state through (today it copies Readiness only for ready/blocked). Cross-hook table test: one draft feature asserted absent from ReadySet, absent from Summary committed counts, `readiness:"draft"` in the view, refused by start. |
| Non-byte-preserving raw mutation (touched phase re-renders differently, or an untouched phase drifts) | Med | Med | `appendFeatureToPhase` renders a FULL feature via the existing `compactBytes` (deterministic struct-field key order + omitempty); assert exact rendered bytes and that untouched phases round-trip byte-identical (dirty-index contract). |
| A **rejected** mutation does not leave `roadmap.json` byte-identical | Med | Med | Mutate `rawDoc` in memory only; run all validation (slug, collision, phase kind, `toRoadmap`+`ValidateDependencies`, artifact preflight) BEFORE the single `writeRawRoadmap`. Every reject path returns before any byte hits disk — mirror the existing `Promote` preflight ordering. Test asserts pre/post file bytes identical on each reject. |
| `promote` branch ambiguity: Backlog-move vs in-place draft finalize chosen wrongly (e.g. off a `--phase` flag instead of location) | Med | Med | Branch strictly on the slug's CURRENT location: Backlog finding → move-into-`--phase`; draft in a schedulable phase → in-place finalize (no move); non-draft non-Backlog → clear error. Test all three branches; keep the cobra `Changed("scores")` sentinel. |
| `add`'s generalized helper regresses the existing name-only `appendToPhase` used by Backlog promote | Med | Low | Generalize rather than fork: `appendToPhase` becomes a thin caller of `appendFeatureToPhase(target, Feature{Name:slug, DependsOn:[]string{}})`; keep existing promote tests green as the regression fence. |
| Per-package coverage regression (tests/-tier files contribute zero to the 95% gate) | Low | Med | Colocated `internal/roadmap/*_test.go` and `cmd/centinela/*_test.go`; aim ≥97% so parallel merges don't tip main red. Every new source file ≤100 lines (incl. `_test.go`). |

## Rollout — smallest correct slice first

1. **Draft field + all hooks + `add`.** Add `Draft` to `Feature`; wire the single
   `NonBacklogFeatureSet` exemption, `Summary`, `classifyFeature`, `mdgen`,
   `start_guard`; add `draft.go`; generalize the raw helpers
   (`rawfeature_find`/`rawfeature_mutate`/`rawtyped`); ship `add` (validate-then-
   mutate-then-`ValidateDependencies`-then-write-once). This is the coherent floor:
   authoring a gate-safe draft is testable end-to-end (`add` → `validate` PASS).
2. **`remove`.** Add `rawdeps.go` (`featureDependents`) and the two guards
   (dependents incl. draft dependents, in-progress/done). Independent of promote.
3. **Generalized `promote` (in-place finalize).** Extract `appendScoreArtifacts`
   shared path; add the location-branch; finalize clears `Draft` via
   `replaceFeatureAt` and appends artifacts without moving.
4. **View extension alongside step 1** (it is part of the draft dimension):
   `FeatureView.Draft` + `readiness:"draft"` in `buildFeatureView`/`BuildView`.

## Deferred Findings

None. No genuinely new gap surfaced during planning; the brief fences the scope and
the umbrella design already records the successor features (roadmap-edit-move,
roadmap-phase-ops) in the Backlog.

## Handoff

**Next role:** feature-specialist.

**Outstanding questions / call-outs for the specialist:**
- The single highest-value invariant to encode as tests is the **draft-dimension
  consistency across four readers** (coverage set, readiness/ReadySet, Summary,
  view). Source of truth is the persisted `f.Draft`; only `NonBacklogFeatureSet`
  contains the exemption *predicate*.
- `buildFeatureView` currently copies `Readiness` only for `ready|blocked`; the
  specialist should specify that the `draft` state also flows through (set
  `fv.Readiness = "draft"` and `fv.Draft = true`), otherwise a draft renders with
  an empty readiness in the JSON contract.
- Confirm `add`'s full-feature entry key ordering matches `compactBytes`/`Feature`
  struct field order for byte-stability, and that `appendToPhase` is refactored to
  delegate (no fork) so Backlog promote stays byte-identical.
- No design change recommended before coding — the plan's file list and reuse map
  are consistent with the delivered raw layer and feature-1 view layer.
