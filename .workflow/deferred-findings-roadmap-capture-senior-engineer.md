### Senior-Engineer Report: deferred-findings-roadmap-capture
**Date:** 2026-06-12

#### Files Touched
| Path | Reason |
|------|--------|
| internal/roadmap/roadmap.go | edit — extend `Feature` (Summary/Source/DeferredAt omitempty) + new `Source` type |
| internal/roadmap/backlog.go | new — `BacklogPhaseName`, `isBacklogPhaseName`/`IsBacklogPhaseName`, `IsBacklogFeature`, `BacklogFeatures`, `NonBacklogFeatureSet` |
| internal/roadmap/analysis.go | edit — `roadmapFeatureSet` now delegates to `NonBacklogFeatureSet` (Backlog exempt; shared by ValidateAnalysis+Quality) |
| internal/roadmap/readiness.go | edit — `DeriveReadiness` skips Backlog phases |
| internal/roadmap/rawio.go | new — raw read (map/RawMessage), atomic temp-file+rename write, `compactBytes` |
| internal/roadmap/rawrender.go | new — deterministic re-indent render; untouched phases re-indented from raw bytes, dirty phases rebuilt |
| internal/roadmap/rawphase_render.go | new — dirty-phase renderer: features one compact object per line (merge-union friendly) |
| internal/roadmap/rawmutate.go | new — decode/encode phases, feature-name scan, `appendBacklog` (creates Backlog phase last) |
| internal/roadmap/rawmove.go | new — `findInBacklog`, `removeBacklogFeature`, `appendToPhase`, known-phase listing |
| internal/roadmap/defer.go | new — `Defer` orchestration (validate-before-write, stamp deferredAt) |
| internal/roadmap/defer_validate.go | new — slug rule (`// mirrors worktree.ValidateFeatureSlug`), collision, empty-summary checks |
| internal/roadmap/promote.go | new — `Promote` + `LoadBacklogFinding` + `BacklogFinding` decode |
| internal/roadmap/promote_scores.go | new — `ParseScores` CSV (six ints, 1-10, overall >= 9) |
| internal/roadmap/promote_artifacts.go | new — raw-preserving append to analysis/quality JSON + provenance bullets |
| internal/roadmap/artifactio.go | new — `writeArtifact` (features one-per-line), `writeFeatureArray`, `appendLine` |
| internal/ui/render_roadmap.go | edit — skip Backlog in phase loop; append `renderBacklogSection` |
| internal/ui/render_backlog.go | new — Backlog findings block (`○ <slug>  <summary>`, IconPending/StyleMuted) |
| internal/ui/render_promote.go | new — `RenderPromoteEvaluatorContext` panel (renderSystemPanel pattern) |
| cmd/centinela/roadmap_defer.go | new — cobra `roadmap defer`; `--source` auto-resolved via worktree.DetectFeatureFromCwd |
| cmd/centinela/roadmap_promote.go | new — cobra `roadmap promote` (`--phase/--summary/--scores`); evaluator vs scored path |
| cmd/centinela/start_guard.go | edit — refuse a Backlog slug with a "promote it first" error |
| docs/architecture/*-prompt.md (8) | edit — uniform `#### Deferred Findings` section |
| internal/scaffold/assets/docs/architecture/*-prompt.md (8) | edit — byte-identical mirrors |

#### Architecture Compliance
- Boundary checks passed: `internal/roadmap` adds NO import to `internal/worktree` (slug rule
  duplicated with a mirror comment per G2). `cmd/centinela` imports roadmap/ui/worktree (all
  permitted for the cmd layer). `internal/ui` imports `internal/roadmap` (existing edge, unchanged).
  `internal/roadmap` (UNMAPPED, Warn-only) imports nothing new from mapped layers. `internal/gates`
  and `internal/verify` untouched.
- G1 file size: every new/edited source file <= 100 lines. Largest new files: rawrender.go 93,
  artifactio.go 93, rawmutate.go 88, rawio.go 85, rawmove.go 82, promote.go 78, roadmap_promote.go
  77, promote_artifacts.go 73, backlog.go 71, rawphase_render.go 64, roadmap_defer.go 59, defer.go
  55, defer_validate.go 41, render_promote.go 40, promote_scores.go 36, render_backlog.go 23.
- G7: `cmd/` files are flag parsing + wiring only. Business logic (validation, raw I/O, move/append,
  score parsing) lives in `internal/roadmap`; all formatting in `internal/ui`.

#### Type-Safety Notes
- Raw-preserving I/O uses `map[string]json.RawMessage` / `[]json.RawMessage` so unknown fields on
  untouched entries (live analysis `legacyDependsOn`, roadmap `customField`) are never dropped or
  re-keyed — verified in smoke tests.
- `Source` is a pointer (`*Source`, omitempty) so non-Backlog and root-defer entries serialize with
  no `source` key. No `any` leaks into public signatures beyond the unavoidable `json.RawMessage`
  raw-region handling. `ParseScores` returns typed `QualityScores`; validated before any write.

#### Trade-Offs
- Deterministic re-render vs literal byte-preservation: untouched phases are re-indented via
  `json.Compact`+`json.Indent` (preserves key order and every field, normalizes whitespace). The
  writer is deterministic and idempotent, so "byte-identical untouched entries" holds across the
  tool's own writes; field-preservation (the real intent) is guaranteed. A hand-authored file is
  reformatted once on first write.
- Backlog features (and analysis/quality `features`) render one compact object per line so
  concurrent appends are a trivial textual merge-union, per the operator-accepted risk.
- `--source` auto-resolution lives in the cmd layer (calls `worktree.DetectFeatureFromCwd`) so
  `internal/roadmap` keeps no worktree import edge; `Defer` takes an already-resolved `*Source`.

#### Fix Pass (tests step)
The edge-case-tester surfaced four in-scope defects; all fixed before tests were written against them.
1. **Promote partial-write on missing artifacts (High).** `Promote` (`internal/roadmap/promote.go`)
   now mutates roadmap.json only in-memory, then calls `preflightArtifacts()`
   (`internal/roadmap/promote_preflight.go`) to confirm both artifact JSON files exist+parse and both
   `.md` companions exist BEFORE the first `writeRawRoadmap` byte. Any failure leaves all five files
   byte-identical (no half-promoted state).
2. **Duplicate entry on slug collision (Medium).** `appendToPhase` (`internal/roadmap/rawmove.go`)
   scans the target phase and refuses with a clear error when the slug already exists there, instead
   of appending a second entry.
3. **Nondeterministic key order (Medium).** `writeArtifact` (`internal/roadmap/artifactio.go`) and
   `rawDoc.render` (`internal/roadmap/rawrender.go`) now emit map-backed non-feature keys via
   `sortedKeys` (`internal/roadmap/mapkeys.go`) in sorted order — chosen over original-order
   preservation because the source is a `map[string]json.RawMessage` with no retained order; sorting
   is the cheapest deterministic choice. Feature-array order is preserved (slice, append-only).
4. **`--scores ""` silent evaluator path (Low).** `runRoadmapPromote`
   (`cmd/centinela/roadmap_promote.go`) uses `cmd.Flags().Changed("scores")` to distinguish unset
   from explicitly-empty; an explicit `--scores ""` is now a usage error (exit 1).

#### Deferred Findings
- `rawio-reformat-diff-churn` (source: deferred-findings-roadmap-capture/senior-engineer) — the first
  defer/promote re-renders untouched phases of roadmap.json, producing spurious git diff churn. Known
  trade-off (writer is deterministic/idempotent thereafter); deferred to the worktree Backlog via the
  `roadmap defer` command. The two edge-case-tester-deferred bugs
  (`promote-partial-write-on-missing-artifacts`, `promote-duplicate-entry-on-slug-collision`) were
  fixed in-feature and removed from the Backlog.

#### Handoff
- Next role: qa-senior
- Outstanding TODOs:
  - Write per-package colocated `_test.go` (<= 100 lines each) for raw I/O, defer/promote, the
    backlog predicate, and the ui render helpers to meet the 95% coverage gate.
  - Golden-file test: round-trip a roadmap with a `customField`, assert untouched entries unchanged
    and Backlog features emitted one-per-line.
  - Table tests: ParseScores (count/range/threshold), validate exemption matrix (real feature still
    required, Backlog exempt, "Pre-Backlog Work" not exempt), start-guard refusal, promote
    rejections (all zero-write).
  - Decision logged: an emptied Backlog phase is KEPT (not dropped) after promote — harmless
    (exempt everywhere) and avoids extra phase-reshuffle mutation; render shows no section when empty.
