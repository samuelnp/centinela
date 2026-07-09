### Big-Thinker Report: roadmap-edit-move

**Date:** 2026-07-01

#### Problem

`roadmap-crud-add-remove` (shipped) can create and delete features but cannot
**change** one in place, **move** it between phases, or **reorder** it within a
phase. Operators — and the Magallanes Plan page driving the CLI — need to
rename/retarget a feature, fix its `dependsOn`, relocate it to the correct phase,
and order features among siblings, all without hand-editing `roadmap.json`. This
feature adds the mutate/relocate half of the editing suite, reusing the raw-feature
helpers feature 2 introduced, while keeping every write atomic, byte-stable, and
graph-validated so a rejected op changes nothing.

#### Scope

- **In:** `roadmap edit`/`update <slug>` (in-place `--name`/`--description`/
  `--depends-on`/`--archetype`, with a `--depends-on` sentinel distinguishing
  "clear deps" from "unchanged"; rename rewrites dependents' `dependsOn` across all
  phases; cycle/unknown-dep re-validation before write); `roadmap move <slug>
  --to-phase <P> [--before|--after <anchor>]` (general phase→phase relocation,
  preserving draft/quality state); `roadmap reorder <slug> --before|--after
  <anchor>` (reposition by anchor).
- **Out:** add/remove/draft/promote (roadmap-crud-add-remove, shipped); phase
  add/rename/remove (roadmap-phase-ops, successor). No persisted-schema change.

#### Dependencies & Assumptions

- Builds directly on **roadmap-crud-add-remove (shipped)**: reuses
  `rawfeature_find.go` (`findFeature`/`featurePhase`), `rawfeature_mutate.go`
  (`removeFeatureAt`/`replaceFeatureAt`/`insertFeatureAt`), `rawtyped.go`
  (`toRoadmap`), `rawdeps.go` (`featureDependents`), the atomic `rawio.go`/
  `rawrender.go`, `ValidateDependencies`, `validateSlug`, `validateNoCollision`,
  and the Backlog/Baseline guards. Confirmed present: `rawfeature_find.go`,
  `rawdeps.go`.
- Dependencies are keyed by feature **name**, not phase — moving a depended-on
  feature keeps the graph valid; only **rename** must rewrite references.
- Draft/finalized status and quality entries are keyed by name, so a move that
  preserves the name preserves them.
- Every mutation is one read-modify-write through the existing atomic raw layer;
  last-writer-wins on concurrency is inherited and unchanged.

#### Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| Rename dependent-rewrite miss → dangling `dependsOn` ref | High | `rewriteDependents(old,new)` across all phases + post-write `ValidateDependencies`; test multiple dependents across multiple phases |
| Anchor off-by-one on `--before`/`--after` insertion index | Med | Test first/last/middle and cross-phase anchors explicitly |
| Non-byte-preserving move (untouched phases mutated) | Med | Assert exact bytes on untouched phases round-trip |
| Rejected op not byte-identical (partial write on refusal) | Med | Validate-before-write; single atomic write; assert `roadmap.json` unchanged on every rejection path |
| Per-package coverage regression | Low | Colocated `internal/roadmap/*_test.go`; aim ≥97% |

#### Rollout

1. `edit` field-only (apply provided flags, `--depends-on` sentinel, re-validate).
2. `edit --name` rename branch + `rewriteDependents` (split `edit_rename.go`).
3. `move` (Backlog/Baseline guard, anchor resolution, draft/quality preserved).
4. `reorder` (within/cross-phase reposition, no-op byte-identical).

#### Deferred Findings

none

#### Handoff

feature-specialist — implement `internal/roadmap/{edit,edit_rename,move,reorder}.go`
+ `rewriteDependents`/`insertFeatureAt` raw helpers and the three thin cobra
commands, keeping each source + `_test.go` ≤100 lines.
