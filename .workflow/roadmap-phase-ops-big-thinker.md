# Big-Thinker Report: roadmap-phase-ops

**Date:** 2026-07-09

## Problem

The Roadmap Editing Suite can now create, remove, edit, move, and reorder
**features** (roadmap-crud-add-remove + roadmap-edit-move, shipped), but it cannot
manage the **phases** that contain them. An operator — and the Magallanes Plan page
that drives the CLI — still has to hand-edit `roadmap.json` to add a new phase,
rename one, or remove an empty (or forcibly-pruned) phase. This is the last and
highest-complexity slice of the suite, because inserting or removing a phase shifts
every later phase index, so the format-preserving raw layer's `dirty`-index
bookkeeping (keyed by phase index) has to be reindexed in lockstep or an unrelated
phase renders corrupt.

## Scope

**In:** `roadmap phase add <name> [--note --after <phase>]`, `roadmap phase rename
<old> <new>`, `roadmap phase remove <name> [--force]`; the raw-layer structural
helpers `insertPhaseAt`/`removePhaseAt`/`renamePhaseAt` that **reindex the `dirty`
map** on insert/remove; Backlog/Baseline reserved-name guards on all three commands;
and the `--force` non-empty-remove path that prunes the phase, its features, and
those features' analysis/quality entries in one atomic write, then re-validates and
refuses (byte-identical) if the result would not pass.

**Out (shipped):** feature-level add/remove/draft/promote (roadmap-crud-add-remove);
feature-level edit/move/reorder (roadmap-edit-move). No new persisted-schema fields.

## Dependencies & Assumptions

- **Raw layer + dirty-index render contract** already on disk: `rawio.go`,
  `rawrender.go` (`internal/roadmap/rawrender.go` confirmed present),
  `rawfeature_mutate.go` (confirmed present) — `decodePhase`/`setPhase`/
  `knownPhaseList`/`compactBytes`, `readRawRoadmap`/`writeRawRoadmap`/
  `finalizeMutation`. The reindex work MUST preserve this contract, not fork it.
- **Reserved-name guards** `isBacklogPhaseName`/`isBaselinePhaseName` exist and are
  reused verbatim; Backlog/Baseline are managed only via defer/promote and must be
  untouchable by add/rename/remove.
- **Validators** `ValidateDependencies`/`ValidateAnalysis`/`ValidateQuality` and
  `toRoadmap` exist and run against the in-memory post-op roadmap before the write.
- **Analysis/quality entry writers** from `promote_artifacts.go`/`artifacts_shared.go`
  are reused to prune coverage on the `--force` path (no reimplementation).
- Assumes atomic temp+rename write (last-writer-wins) already governs the raw layer;
  a rejected op writes nothing. No schema migration; `Roadmap`/`Phase` reused as-is.

## Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| Dirty-map reindex off-by-one after insert/remove corrupts an *unrelated* later phase (the core new complexity) | **High** | Unit-test insert/remove at first/middle/last WITH a concurrently-dirtied later phase; assert exact rendered bytes prove the reindex (key ≥ pos → +1 on insert; drop idx, key > idx → −1 on remove). Reindex lives ONLY in insertPhaseAt/removePhaseAt. |
| Orphaned analysis/quality coverage left behind by `--force` remove of scored features | Med | Prune phase + features + their analysis/quality entries in the SAME atomic write; re-run ValidateAnalysis/ValidateQuality before committing. |
| Reserved-phase corruption — a command touching Backlog/Baseline breaks defer/promote invariants | Med | Guard all three commands (add/rename/remove) with isBacklog/isBaseline on both source and target names. |
| Rejected op not byte-identical (partial or reformatting write leaks on refusal) | Med | Validate fully in memory before any write; on any refusal return before writeRawRoadmap; add byte-identical assertions on every refusal path in tests. |
| Per-package coverage dips below the 95% gate (target ≥97%) | Low | Colocated `internal/roadmap/*_test.go`; cover each refusal branch and the reindex matrix. |

## Rollout

Ship smallest-blast-radius slice first, growing structural risk last:

1. **`phase rename`** — `renamePhaseAt` reuses the existing dirty mechanism, no
   structural index shift. Lowest risk; validates the guard + validate + write spine.
2. **`phase add`** — `insertPhaseAt`: the first reindex (keys ≥ pos shift +1), plus
   `--after`/before-Backlog anchor resolution and reserved/duplicate/empty guards.
3. **`phase remove` (empty)** — `removePhaseAt`: the second reindex (drop idx, keys >
   idx shift −1); non-empty refusal path (name the feature count).
4. **`phase remove --force`** — prune phase+features+analysis/quality entries, re-run
   dependency+analysis+quality validation, refuse byte-identical on break (e.g. a
   surviving feature still `dependsOn` a removed one). Highest complexity, ships last.

## Deferred Findings

none

## Handoff

feature-specialist
