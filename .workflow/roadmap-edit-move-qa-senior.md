# roadmap-edit-move — qa-senior

## Test Inventory

### Colocated — `internal/roadmap` (move the same-package coverage gate)
| File | Lines | Covers |
|------|-------|--------|
| `edit_move_helpers_test.go` | 77 | `canonRoadmap`/`phaseSlice`/`orderIn`/`contains` shared helpers |
| `edit_fields_test.go` | 76 | field-only edits, `--depends-on` unchanged/clear/replace, not-found byte-identical |
| `edit_rename_test.go` | 91 | cross-phase dependent rewrite + validate PASS, collision (owning phase), invalid slug, same-name idempotent |
| `edit_reject_test.go` | 48 | unknown-dep, self-cycle, multi-hop cycle — all byte-identical |
| `edit_error_test.go` | 64 | missing-file (edit/move/reorder), archetype branch, malformed target/sibling decode |
| `move_test.go` | 74 | append default, anchor first/last/middle, self-anchor deviation (byte-identical) |
| `move_guard_test.go` | 63 | Backlog/Baseline src+target, unknown phase/anchor, not-found; draft preserved |
| `move_reorder_error_test.go` | 40 | post-mutation typed re-decode failures, unknown anchor after remove |
| `reorder_test.go` | 55 | within/cross-phase reposition, no-op no-write byte-identical |
| `reorder_guard_test.go` | 37 | Backlog anchor, not-found, unknown anchor, missing --before/--after |
| `rawdeps_rewrite_test.go` | 49 | `rewriteDependents` multi-dependent/multi-phase exact bytes, no-match no-op |
| `anchor_insert_test.go` | 59 | `anchorPos` before/after/append/unknown, `insertFeatureAt` head/tail/out-of-range |
| `raw_error_test.go` | 69 | featureName/phaseOrder/schedulablePhaseIndex/applyRename/applyReorder error branches |

### Colocated — `cmd/centinela`
| File | Lines | Covers |
|------|-------|--------|
| `roadmap_edit_test.go` | 100 | field edit, `--depends-on` cobra `Changed()` sentinel (clear vs preserve), error propagation |
| `roadmap_move_test.go` | 63 | success, `--to-phase` required, `--before/--after` mutually exclusive, error propagation |
| `roadmap_reorder_test.go` | 65 | success, anchor required, mutually exclusive, error propagation |

### tests/ tier trio (required by the gate; does NOT move coverage)
| File | Lines | Covers |
|------|-------|--------|
| `tests/unit/roadmap_edit_move_unit_test.go` | 84 | edit/rename/move/reorder + rejected-op byte-identical, via package API |
| `tests/integration/roadmap_edit_move_integration_test.go` | 75 | move keeps dependency valid; move preserves `roadmap-quality.json` byte-identical |
| `tests/acceptance/roadmap_edit_move_test.go` | 93 | temp-built binary: rename rewrite, `update` alias, edit rejections byte-identical |
| `tests/acceptance/roadmap_edit_move_ops_test.go` | 100 | temp-built binary: move append/anchor, reorder within/no-op, move guard refusals |

All files ≤100 lines (G1). No production code modified.

## Coverage Gaps

- Gated metric (full `go test ./...`): **97.2% ≥ 95%** (PASS, >2% margin).
- Per-package: `internal/roadmap` **96.0%**, `cmd/centinela` **96.6%** — both above the 95% floor.
- Residual uncovered lines in the new code are **unreachable defensive decode-guards**: the `decodePhase` error
  branches in `requireSchedulablePhaseIdx`, `insertFeatureAt`, `replaceFeatureAt`, and the `compactBytes` error
  paths — in every case the phase/entry was already successfully decoded upstream, so the guard cannot fire.
  Forcing them would require injecting an impossible mid-operation mutation; not pursued.

## Acceptance Wiring

- Acceptance + unit + integration files carry `// Acceptance: specs/roadmap-edit-move.feature` and per-test
  `// Scenario: <exact name>` comments mapping the majority of the 31 scenarios.
- Acceptance drives a **temp-built** binary (`rmcBin`/`rmcProject`/`rmcRun`) in `t.TempDir()` — no network,
  no git push, no installed binary. Edit/move/reorder only run dependency validation, so a plain roadmap body
  (no analysis/quality artifacts) suffices.

## Deferred Findings

All are **data-safe** (roadmap.json left byte-identical; no silent mutation) and are spec-vs-code deviations for
the validation-specialist to reconcile (adjust spec wording, or add a one-line guard) — none is a correctness/data bug:

1. **`move` self-anchor is not a no-op.** `move x --to-phase P --after x` removes `x` before resolving the anchor,
   so it errors `anchor feature "x" not found in phase "P"` instead of exiting 0 (spec: "move self-anchor no-op").
   File untouched. Tested as `TestMove_SelfAnchorByteIdentical`.
2. **`edit --name <same>` is not byte-identical.** Edit always re-renders the target's phase one-per-line, so a
   same-name edit differs from a `json.Indent`-canonical on-disk file (spec: "same-name byte-identical"). It IS
   idempotent once settled and rewrites no dependents. Only rejected ops and the reorder no-op guard are truly
   byte-identical. Tested as `TestEdit_RenameSameNameStable`.
3. **Error-wording drift.** Backlog/Baseline targets say "non-schedulable" (spec example: "unknown phase");
   unknown anchors/slugs say "not found" (spec example: "unknown feature"). Exit code + byte-identity match the spec.

## Handoff

→ **validation-specialist**. Gate status green: `go vet ./...` clean, `go test ./...` all pass (3300+ tests),
coverage 97.2% ≥ 95%, `check-fmt.sh` clean. Please reconcile the three deferred spec deviations above (all data-safe)
against `specs/roadmap-edit-move.feature` before shipping — either soften the spec's byte-identical/exit-0/error-substring
expectations for those edge cases, or request a one-line code adjustment (e.g. a self-anchor short-circuit in `Move`).
