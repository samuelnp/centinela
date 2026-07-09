# roadmap-phase-ops — senior-engineer

## Files Touched

New internal/roadmap source (all ≤100 lines):
- `rawphase_struct.go` (81) — `insertPhaseAt`/`removePhaseAt` (the dirty-map reindex), `renamePhaseAt`, `phaseIndexByName`.
- `phase_add.go` (43) — `PhaseAdd`; validation + insert + `finalizeMutation`.
- `phase_add_position.go` (29) — `insertPosition` (resolve `--after` / before-Backlog / last).
- `phase_rename.go` (44) — `PhaseRename`; reserved/collision/empty guards, same-name no-op.
- `phase_remove.go` (43) — `PhaseRemove`; reserved/unknown guard, empty-vs-non-empty branch.
- `phase_remove_force.go` (47) — `forceRemovePhase` + `revalidateArtifacts` (prune analysis/quality).
- `artifact_prune.go` (43) — `removeFeatureEntries` (mirror of `appendFeatureEntry`, drops entries by name).

New cmd/centinela thin commands:
- `roadmap_phase.go` (15) — parent `phase` command (registration only).
- `roadmap_phase_add.go` (37) — `phase add <name> [--note --after]`.
- `roadmap_phase_rename.go` (30) — `phase rename <old> <new>`.
- `roadmap_phase_remove.go` (34) — `phase remove <name> [--force]`.

Modified (minimal, backward-compatible):
- `rawmutate.go` — added `Note string json:"note,omitempty"` to `rawPhase` so a rename that
  re-encodes via `setPhase` preserves an existing phase note; omitempty keeps note-less phases
  byte-identical.
- `rawphase_render.go` — `renderDirtyPhase` now emits the `note` key (after `name`, before
  `features`) so an inserted/renamed dirty phase preserves its note in the render contract.

## Architecture Compliance

- Every source file ≤100 lines (max is `rawphase_struct.go` at 81). No G1 exception needed.
- Layering: all logic lives in `internal/roadmap`; the four cobra files are thin (flag parse →
  delegate → render success). Business rules (guards, positions, reindex, prune, validation) are
  entirely in the internal package.
- Reuse-only: no forked raw layer. Uses `readRawRoadmap`/`writeRawRoadmap`/`finalizeMutation`,
  `decodePhase`/`setPhase`, `backlogPhaseIndex`/`phaseBytes`, `isNonSchedulablePhase`,
  `ValidateDependencies`/`ValidateAnalysis`/`ValidateQuality`, `compactBytes`, `writeArtifact`,
  and the render contract. The dirty-map reindex lives ONLY in `insertPhaseAt`/`removePhaseAt`.
- Deterministic/atomic: one mutation = one atomic temp+rename write; a rejected op writes nothing;
  untouched regions round-trip via the existing render path.

## Type-Safety Notes

- Go strict typing throughout; no `interface{}`/`any`. Phase entries flow as `json.RawMessage`
  for byte-preservation (same idiom as the shipped feature layer).
- `insertPhaseAt`/`removePhaseAt` bounds-check `pos`/`idx` and return errors rather than panicking.
- The `--force` prune revalidates the in-memory typed `Roadmap` for dependencies BEFORE any write,
  so the surviving-dependsOn case refuses byte-identically (proven in dogfood: roadmap+analysis+
  quality shas all unchanged).

## Trade-Offs

- **Reindex is the sole locus of index bookkeeping.** `insertPhaseAt` shifts every dirty key ≥ pos
  by +1 and marks the new phase dirty; `removePhaseAt` drops key idx and shifts keys > idx by −1.
  An isolated unit-level check (concurrently-dirty later phase + earlier insert/remove) confirmed
  the shifted phase's bytes render at the correct index and the doc round-trips as valid JSON.
- **Force-remove validation ordering.** Dependency validation runs in-memory on the mutated typed
  roadmap before any byte hits disk — the decisive byte-identity gate. Analysis/quality coverage is
  consistent by construction (exactly the removed features' entries are pruned), and
  `revalidateArtifacts` re-runs `ValidateAnalysis`+`ValidateQuality` against the final on-disk state
  as a post-write correctness check. The `.md` provenance companions are intentionally NOT pruned
  (append-only provenance; not part of the coverage set the validators read).
- **Note preservation.** Extending `rawPhase` with an omitempty `Note` and the render key list is the
  smallest correct change; without it a rename of a noted phase (or an inserted `--note` phase) would
  silently drop the note. Note-less phases stay byte-identical.

## Deferred Findings

None.

## Handoff → qa-senior

Implementation is complete and green (build/vet/test-compile clean). Tests are the NEXT step — write
them per the plan's Tests section:
- `rawphase_struct_test.go` — insert/remove at first/middle/last WITH a concurrently-dirtied phase;
  assert EXACT rendered bytes (this is the make-or-break reindex proof; I verified it ad-hoc but it
  needs a committed test).
- `phase_add_test.go` / `phase_add_position` coverage — `--after` first/middle/last/before-Backlog,
  `--after Backlog`, empty roadmap, `--note`; duplicate/reserved/empty/unknown-anchor refusals,
  byte-identical untouched phases and on-reject.
- `phase_rename_test.go` — in-place rename, same-name no-op byte-identity, unknown/collision/reserved
  (either side)/empty refusals.
- `phase_remove_test.go` / `phase_remove_force_test.go` — empty-phase remove, non-empty refusal
  (feature count), `--force` prunes roadmap+analysis+quality and revalidates PASS, `--force` refused
  byte-identical when a surviving feature dependsOn a removed one, reserved refusals with/without
  `--force`, remove-only-phase → `{"phases":[]}`.
- `artifact_prune` coverage — missing artifact file is a no-op.
- `cmd/centinela/roadmap_phase_*_test.go` — flag parsing (`--note`/`--after`/`--force`).
- tests/ tier trio with `// Acceptance:`/`// Scenario:` traceability; acceptance drives a temp-built
  binary against a temp fixture (NO network).
Target ≥97% per-package coverage (colocated `internal/roadmap/*_test.go`, each ≤100 lines incl. tests).
