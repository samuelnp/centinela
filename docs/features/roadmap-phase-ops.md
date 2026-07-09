# Feature Brief — roadmap-phase-ops

> Feature 4 of 4 in the Roadmap Editing Suite. Umbrella design:
> [docs/plans/roadmap-editing-suite-design.md](../plans/roadmap-editing-suite-design.md).
> This brief isolates the acceptance criteria + edge cases for THIS feature only
> (phase-level operations), separate from the feature-level commands already shipped.

## Problem — what pain, who
The suite can now create, remove, edit, move, and reorder **features**
(roadmap-crud-add-remove + roadmap-edit-move, shipped) but cannot manage the
**phases** that contain them: an operator (and the Magallanes Plan page) still
cannot add a new phase, rename one, or remove an empty phase without hand-editing
`roadmap.json`. This feature adds phase-level structural operations, completing
the editing suite. It is the highest-complexity raw-layer change because inserting
or removing a phase shifts every later phase index, so the format-preserving raw
layer's dirty-index bookkeeping must be reindexed.

## Scope (this feature ONLY)
- **In:** `roadmap phase add <name> [--note --after <phase>]`, `roadmap phase
  rename <old> <new>`, `roadmap phase remove <name> [--force]`; the raw-layer
  structural helpers (`insertPhaseAt`, `removePhaseAt`, `renamePhaseAt`) that
  reindex the `dirty` map on insert/remove.
- **Out (shipped):** feature add/remove/draft/promote (roadmap-crud-add-remove);
  feature edit/move/reorder (roadmap-edit-move). No new persisted-schema fields.

## User Stories
- As an operator, `roadmap phase add "Phase 13: Hardening" --after "Phase 12: …"`
  inserts a new empty phase at the chosen position.
- As an operator, `roadmap phase rename "Phase 13: Hardening" "Phase 13: Security"`
  renames a phase in place without disturbing its features or other phases.
- As an operator, `roadmap phase remove "Phase 13: Security"` deletes an **empty**
  phase; a non-empty phase is refused unless I pass `--force`.

## Acceptance Criteria (THIS feature → Gherkin)
1. `roadmap phase add <name> [--note <n>] [--after <phase>]` inserts a new phase
   with an empty `features: []` (and optional `note`); with `--after` it lands
   immediately after the named phase, otherwise before the Backlog phase (or last
   if no Backlog); untouched phases round-trip byte-identical.
2. `phase add` refuses: a duplicate phase name, the reserved names `Backlog`/
   `Baseline` (managed via defer/promote), an empty name, and an unknown `--after`
   anchor — each leaving `roadmap.json` byte-identical.
3. `roadmap phase rename <old> <new>` renames in place; refuses when `<old>` is
   unknown, `<new>` collides with an existing phase, or either side is a reserved
   `Backlog`/`Baseline` name; features and all other phases are untouched
   (byte-identical).
4. `roadmap phase remove <name>` deletes an **empty** phase; a non-empty phase is
   **refused** (naming the feature count) unless `--force` is given, which removes
   the phase and its features; an unknown phase errors; `roadmap.json` byte-identical
   on any refusal.
5. All three commands are read-modify-write via the existing atomic raw layer; a
   rejected op writes nothing, and after any structural insert/remove the on-disk
   render is still valid JSON with correctly-ordered phases.
6. `roadmap validate` stays green after each operation. Specifically, `phase
   remove --force` of a phase containing scored (non-draft) features removes the
   phase, its features, **and** those features' analysis and quality entries in the
   SAME atomic write, then re-runs dependency + analysis + quality validation and
   **refuses (byte-identical) if the result would not pass** — e.g. when a surviving
   feature in another phase still `dependsOn` a removed one. No orphaned coverage is
   ever left behind, and no partial write occurs.

## Edge Cases (THIS feature ONLY)
- Empty roadmap (`{"phases":[]}`): `phase add` succeeds (first phase); `phase
  rename`/`remove` of any name → "not found".
- `phase add --after` the Backlog/Baseline phase → allowed target position, but the
  new phase is inserted as a normal schedulable phase (not inside Backlog).
- `phase remove` the ONLY phase → leaves `{"phases":[]}` (valid empty roadmap).
- `phase remove --force` a phase with scored (non-draft) features → the phase, its
  features, AND those features' analysis/quality entries are all removed in one
  atomic write; dependency + analysis + quality validation re-runs before the write
  and the op is refused (byte-identical) if a surviving feature would be left
  depending on a removed one or coverage would otherwise break.
- `phase rename` to the SAME name → no-op, byte-identical.
- Reindex correctness: inserting/removing a phase in the MIDDLE must not corrupt
  later phases already marked dirty in the same operation (the raw `dirty` map is
  keyed by phase index and must be reindexed).
- Reserved-name protection: neither `add`, `rename`, nor `remove` may create,
  rename into, or delete `Backlog`/`Baseline` via these commands.
- Empty/missing/malformed `roadmap.json` → error, file untouched.
- Concurrent writers → inherited atomic temp+rename, last-writer-wins.

## Data Model
No persisted-schema change. Reuses `Roadmap`/`Phase`. New raw-layer helpers in a
`rawphaseops.go` (+ `rawphase_struct.go` if >100 lines): `insertPhaseAt(pos, raw)`,
`removePhaseAt(idx)`, `renamePhaseAt(idx, newName)` — the first two MUST reindex the
`dirty` map so a mixed feature+phase mutation renders correctly.

## Integration Points
- Reuses `rawio.go`/`rawrender.go` (the dirty-index render contract),
  `decodePhase`/`setPhase`/`knownPhaseList`, the Backlog/Baseline name guards, and
  `ValidateDependencies`/`ValidateAnalysis`/`ValidateQuality` for the post-op check.
- Consumer: Magallanes drives phase add/rename/remove via the CLI; `roadmap --json`
  and ROADMAP.md reflect the new phase structure.

## Risks
- **Dirty-map reindex bug** (High): the core new complexity — an off-by-one in
  reindexing after insert/remove corrupts an unrelated phase. Mitigation: unit-test
  insert/remove at first/middle/last positions with a concurrently-dirtied phase;
  assert exact rendered bytes.
- **Orphaned coverage on `--force` remove** (Med): removing scored features can
  break analysis/quality coverage → run `validate` after and refuse/repair.
- **Reserved-phase corruption** (Med): accidentally letting a command touch
  Backlog/Baseline breaks defer/promote invariants; guard all three commands.
- **Per-package coverage** (Low): colocated `internal/roadmap/*_test.go`; ≥97%.

## Decomposition
Final coherent slice of the suite; depends on the raw-feature helpers shipped by
roadmap-crud-add-remove / roadmap-edit-move. No further split.
