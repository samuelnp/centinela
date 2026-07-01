# Senior-Engineer Report — roadmap-edit-move

Adds `roadmap edit`/`update`, `move`, and `reorder`: in-place mutate + relocate of
an existing feature, reusing the shipped raw-feature layer (`roadmap-crud-add-remove`).
No persisted-schema change.

## Files Touched

### internal/roadmap (logic)
| File | Lines | Role |
|------|-------|------|
| `rawdeps.go` | 74 | +`rewriteDependents(old,new)` — rewrite every feature's `dependsOn` old→new across all phases; only phases holding a dependent go dirty |
| `rawfeature_mutate.go` | 83 | +`insertFeatureAt(phaseIdx,pos,entry)` — anchor-based insert (alongside append/remove/replace) |
| `rawfeature_anchor.go` | 36 | +`anchorPos(phaseIdx,before,after)` — resolve before/after anchor to an insert index (empty = append) |
| `mutate_shared.go` | 50 | +`finalizeMutation` (toRoadmap+ValidateDependencies+atomic write), `schedulablePhaseIndex`, `requireSchedulablePhaseIdx` |
| `edit.go` | 49 | `EditRequest`/`Edit` — apply only provided fields, then single atomic write |
| `edit_rename.go` | 43 | `applyEditFields` (Description/Archetype non-empty, DependsOn on SetDeps), `applyRename` (validateSlug+collision+rewriteDependents) |
| `move.go` | 46 | `MoveRequest`/`Move` — refuse Backlog/Baseline src+target, remove+anchor-insert, verbatim entry bytes |
| `reorder.go` | 54 | `ReorderRequest`/`Reorder` — remove+insert, no-op detection skips write |
| `reorder_apply.go` | 66 | `applyReorder`, `phaseOrder`, `sameOrder` (no-op snapshot compare) |

### cmd/centinela (thin cobra)
| File | Lines | Role |
|------|-------|------|
| `roadmap_edit.go` | 49 | `edit\|update <slug> --name --description --depends-on --archetype` (`SetDeps = Changed("depends-on")`) |
| `roadmap_move.go` | 50 | `move <slug> --to-phase --before --after` (before/after mutually exclusive) |
| `roadmap_reorder.go` | 47 | `reorder <slug> --before\|--after` (one required) |

## Architecture Compliance
- Every source file ≤ 100 lines (max 83). No G1 exceptions needed.
- Commands are thin: parse/guard flags, call `internal/roadmap`, print `ui.RenderSuccess`. All mutation logic lives in `internal/roadmap`.
- No new cross-layer import edges; reuses shipped raw layer (`rawio`/`rawrender`/`rawfeature_*`/`mutate_validate`/`dependencies`) unchanged.
- One mutation = one atomic write via the shared `finalizeMutation` → `writeRawRoadmap` (temp-file+rename). Rejected ops write nothing.

## Type-Safety Notes
- No `any` in production paths. `rewriteDependents` decodes to the typed `Feature` (not `map[string]any`); anchor/index helpers use `int`; request structs are fully typed.
- Post-mutation `ValidateDependencies` (unknown-dep, cycle, archetype) runs before every write.

## Trade-Offs
- Edited/dirtied phases re-render one-object-per-line (shipped `renderDirtyPhase`); untouched phases pass through `json.Indent`. "Untouched phase byte-identical" holds for a phase already in `json.Indent` canonical form on disk — a phase previously written in one-per-line form is re-indented on the next write (pre-existing rawrender behavior, not introduced here).
- Reorder no-op is guaranteed byte-identical by *not writing* when the phase-order snapshot is unchanged, rather than relying on render reproducing prior bytes.
- `Edit` re-encodes the target feature via the `Feature` struct (field order/`omitempty`) matching `roadmap add`; only that feature's phase is dirtied.
- Description/Archetype use empty="not provided" (matches `add`); only DependsOn has an explicit clear-vs-omit sentinel per the plan.

## Verification
- `go build ./...`, `go vet ./...` clean; `go test ./... -run xxxNONE` compiles every package (no sibling signature break).
- Dogfood (dev binary `/tmp/cen-f3`, temp project): rename `a→a2` rewrote `b.dependsOn`→`a2`; `move a2 --to-phase Q2 --before d` relocated; reorder no-op sha-identical; rejected edit (cycle) + rejected move (unknown phase) sha-identical; untouched phase byte-identical from canonical file; `update` alias works.

## Deferred Findings
None.

## Handoff → qa-senior
Colocated `_test.go` (≤100 lines each) per plan §Tests: `edit`/`edit_rename` (field-only edits, cross-phase rename rewrite, collision, same-name no-op, cycle/unknown-dep reject byte-identical), `move` (anchor first/last/middle, Backlog/Baseline refusal, unknown phase/anchor, untouched byte-identical, self-anchor no-op, draft/quality preserved), `reorder` (within+cross-phase, no-op byte-identical), `rawdeps_rewrite` (multi-dependent/multi-phase exact bytes), and `cmd` flag-parsing incl. the `--depends-on` Changed sentinel, plus the tests/ tier trio driving a temp-built binary.
