# Plan — roadmap-phase-ops

> Feature 4 of 4. Brief: [docs/features/roadmap-phase-ops.md](../features/roadmap-phase-ops.md).
> Umbrella design: [docs/plans/roadmap-editing-suite-design.md](roadmap-editing-suite-design.md).

## Goal
Add `roadmap phase add`/`rename`/`remove` — phase-level structural operations —
reusing the format-preserving raw layer. The core new complexity: inserting or
removing a phase shifts every later phase index, so the raw `dirty` map (keyed by
phase index) must be **reindexed**. No persisted-schema change.

## Deliverables

### `internal/roadmap/rawphase_struct.go` (~80 → split if needed) — the hard part
- `insertPhaseAt(pos int, raw json.RawMessage) error` — splice a new phase at `pos`;
  reindex `dirty` (every entry with key ≥ pos shifts +1).
- `removePhaseAt(idx int) error` — drop phase `idx`; reindex `dirty` (drop `idx`,
  every key > idx shifts −1).
- `renamePhaseAt(idx int, newName string) error` — decode phase, set name, `setPhase`
  (no structural shift; reuses the existing dirty mechanism).
- Helpers: `phaseIndexByName(name)`, reuse `decodePhase`/`setPhase`/`knownPhaseList`
  and the Backlog/Baseline name guards (`isBacklogPhaseName`/`isBaselinePhaseName`).

### `internal/roadmap/phase_add.go` (~50)
```go
func PhaseAdd(path, name, note, afterPhase string) error
```
Validate name non-empty, not a reserved (`Backlog`/`Baseline`) name, no duplicate;
resolve insert position (`--after <phase>` → just after it; else just before Backlog,
or last); build the phase entry `{name, note?, features: []}`; `insertPhaseAt`;
`toRoadmap()`+`ValidateDependencies`; single atomic write.

### `internal/roadmap/phase_rename.go` (~45)
```go
func PhaseRename(path, oldName, newName string) error
```
Reject unknown `oldName`, reserved either side, `newName` collision, empty `newName`;
same-name → no-op byte-identical; `renamePhaseAt`; validate; write.

### `internal/roadmap/phase_remove.go` (+ `phase_remove_force.go` if >100 lines)
```go
func PhaseRemove(path, name string, force bool) error
```
Reject unknown/reserved `name`. If the phase is non-empty and `!force` → refuse
(name the feature count). If `force` → remove the phase, its features, AND each
removed feature's analysis + quality entries (reuse the artifact-writers from
promote/finalize to prune), in the SAME atomic write; then re-run
`ValidateDependencies`+`ValidateAnalysis`+`ValidateQuality` and refuse
(byte-identical) if a surviving feature would depend on a removed one or coverage
would break. Empty phase → straightforward `removePhaseAt`.

### Thin cobra commands
- `cmd/centinela/roadmap_phase.go` — parent `phase` command (registration only).
- `cmd/centinela/roadmap_phase_add.go` — `phase add <name> [--note --after]` → `PhaseAdd`.
- `cmd/centinela/roadmap_phase_rename.go` — `phase rename <old> <new>` → `PhaseRename`.
- `cmd/centinela/roadmap_phase_remove.go` — `phase remove <name> [--force]` → `PhaseRemove`.

## Reuse (do NOT reimplement)
- Raw layer: `rawio.go` (`readRawRoadmap`/`writeRawRoadmap`/`finalizeMutation`),
  `rawrender.go` (the `dirty`-index render contract that reindex MUST preserve),
  `decodePhase`/`setPhase`/`knownPhaseList`, `compactBytes`.
- Guards/validators: `isBacklogPhaseName`/`isBaselinePhaseName`, `ValidateDependencies`,
  `ValidateAnalysis`/`ValidateQuality`, `toRoadmap`.
- Analysis/quality entry writers from `promote_artifacts.go`/`artifacts_shared.go`
  (for the `--force` prune path).

## Constraints
- Every source + `_test.go` ≤ 100 lines (split per the file list).
- Commands thin; logic in `internal/roadmap`. Strict typing, no `any`.
- Deterministic/byte-stable; untouched phases round-trip byte-identical; a rejected
  op writes nothing; one mutation = one atomic write.
- The `dirty` reindex lives ONLY in `insertPhaseAt`/`removePhaseAt`.

## Tests (colocated for coverage — aim ≥97%)
- `phase_add_test.go` — insert with/without `--after` (first/middle/last/before-Backlog),
  duplicate/reserved/empty-name refusals, unknown anchor, byte-identical untouched phases + on-reject.
- `phase_rename_test.go` — rename in place, unknown/collision/reserved refusals, same-name no-op.
- `phase_remove_test.go` / `phase_remove_force_test.go` — empty-phase remove, non-empty refusal
  (no --force), `--force` removes phase+features+analysis/quality entries and revalidates,
  `--force` refused when a surviving feature depends on a removed one (byte-identical),
  remove-only-phase → empty roadmap.
- `rawphase_struct_test.go` — insert/remove at first/middle/last WITH a concurrently-dirtied
  phase; assert exact rendered bytes prove the `dirty` reindex is correct.
- `cmd/centinela/roadmap_phase_*_test.go` — flag parsing + `--force`/`--after`/`--note`.
- tests/ tier trio (`tests/{unit,integration,acceptance}/roadmap_phase_ops_*_test.go`) with
  `// Acceptance:`/`// Scenario:` traceability; acceptance drives a temp-built binary (no network).

## Verification (end-to-end)
1. `go test ./...` green; `./scripts/check-coverage.sh` ≥95% (target ≥97%); `check-fmt.sh`.
2. Dev binary in a temp project: `phase add "Phase 9: X" --after "<p>"`; `phase rename`;
   add features then `phase remove` (refused) → `--force` (removes + prunes coverage) →
   `roadmap validate` PASS; each rejected op leaves roadmap.json byte-identical; the
   dirty-reindex case (mutate a feature in a later phase in the same run) renders correctly.
3. `centinela validate` passes in the worktree.
