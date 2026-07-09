# Plan — roadmap-edit-move

> Feature 3 of 4. Brief: [docs/features/roadmap-edit-move.md](../features/roadmap-edit-move.md).
> Umbrella design: [docs/plans/roadmap-editing-suite-design.md](roadmap-editing-suite-design.md).

## Goal
Add `roadmap edit`/`update`, `move`, and `reorder` — in-place mutate + relocate
of an existing feature — reusing the raw-feature helpers that
`roadmap-crud-add-remove` (shipped) introduced. No persisted-schema change.

## Deliverables

### `internal/roadmap/rawdeps.go` (extend — feature 2 added `featureDependents`)
- `rewriteDependents(oldName, newName string) error` — rewrite every feature's
  `dependsOn` entry equal to `oldName` → `newName`, across all phases, preserving
  formatting of otherwise-untouched phases.

### `internal/roadmap/rawfeature_mutate.go` (extend if needed)
- `insertFeatureAt(phaseIdx, pos int, entry json.RawMessage) error` — anchor-based
  insertion (reuse/add alongside feature 2's append/remove/replace).

### `internal/roadmap/edit.go` (~60 → split `edit_rename.go` for the rename branch)
```go
type EditRequest struct{ Slug, NewName, Description, Archetype string; DependsOn []string; SetDeps bool }
func Edit(path string, req EditRequest) error
```
`findFeature`; apply only the provided fields to the decoded `Feature`
(`SetDeps` sentinel = cobra `Changed("depends-on")`); if `NewName!="" && !=Slug`:
`validateSlug`, collision check, `replaceFeatureAt`, `rewriteDependents(Slug,NewName)`;
`toRoadmap()`+`ValidateDependencies` (catches cycle/unknown-dep) → single atomic write.

### `internal/roadmap/move.go` (~50)
```go
type MoveRequest struct{ Slug, ToPhase, BeforeAnchor, AfterAnchor string }
func Move(path string, req MoveRequest) error
```
Refuse if source or target is Backlog/Baseline; `findFeature`; `removeFeatureAt(src)`;
resolve anchor → `insertFeatureAt(targetIdx, pos)`; `toRoadmap()`+`ValidateDependencies`;
write. Feature's `Draft`/quality entries (keyed by name) are preserved by the move.

### `internal/roadmap/reorder.go` (~45)
```go
type ReorderRequest struct{ Slug, BeforeAnchor, AfterAnchor string }
func Reorder(path string, req ReorderRequest) error
```
Within-phase (or across, if anchor is elsewhere) reposition via
`removeFeatureAt`+`insertFeatureAt`; a no-op reorder leaves bytes identical.

### Thin cobra commands
- `cmd/centinela/roadmap_edit.go` — `edit|update <slug> --name --description --depends-on(StringSlice) --archetype` (Aliases `update`).
- `cmd/centinela/roadmap_move.go` — `move <slug> --to-phase [--before|--after]`.
- `cmd/centinela/roadmap_reorder.go` — `reorder <slug> [--before|--after <anchor>]`.

## Reuse (shipped by roadmap-crud-add-remove — do NOT reimplement)
- `rawfeature_find.go` (`findFeature`, `featurePhase`), `rawfeature_mutate.go`
  (`appendFeatureToPhase`/`removeFeatureAt`/`replaceFeatureAt`), `rawtyped.go`
  (`toRoadmap`), `rawdeps.go` (`featureDependents`), `mutate_validate.go`,
  the atomic `rawio.go`/`rawrender.go`, `ValidateDependencies`, `validateSlug`,
  `validateNoCollision`, the Backlog/Baseline guards.

## Constraints
- Every source + `_test.go` ≤ 100 lines (split per the file list).
- Commands thin; logic in `internal/roadmap`. Strict typing, no `any`.
- Deterministic/byte-stable; untouched phases round-trip byte-identical; a rejected
  op writes nothing; one mutation = one atomic write.

## Tests (colocated for coverage — aim ≥97%)
- `edit_test.go` / `edit_rename_test.go` — field edits (only provided flags change),
  rename rewrites dependents across phases, rename collision (names phase),
  same-name no-op, cycle/unknown-dep rejection byte-identical.
- `move_test.go` — cross-phase move + anchor (first/last/middle), Backlog/Baseline
  refusal, unknown phase/anchor, untouched-phase byte-identical, self-anchor no-op,
  draft/quality preserved.
- `reorder_test.go` — within-phase + cross-phase reposition, no-op byte-identical.
- `rawdeps_rewrite_test.go` — `rewriteDependents` multi-dependent/multi-phase, exact bytes.
- `cmd/centinela/roadmap_edit_test.go`, `roadmap_move_test.go`, `roadmap_reorder_test.go` — flag parsing incl. the `--depends-on` Changed sentinel.
- tests/ tier trio (`tests/{unit,integration,acceptance}/roadmap_edit_move_*_test.go`) with
  `// Acceptance:`/`// Scenario:` traceability; acceptance drives a temp-built binary (no network).

## Verification (end-to-end)
1. `go test ./...` green; `./scripts/check-coverage.sh` ≥95% (target ≥97%); `check-fmt.sh`.
2. Dev binary in a temp project: `roadmap add a --phase P; roadmap add b --phase P`;
   `roadmap edit a --name a2` → `b`'s dependsOn (if on `a`) rewritten, validate PASS;
   `roadmap move a2 --to-phase Q --before <x>` relocates, untouched phases byte-identical;
   `roadmap reorder` repositions; each rejected op leaves roadmap.json byte-identical;
   cycle via `edit --depends-on` refused.
3. `centinela validate` passes in the worktree.
