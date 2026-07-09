### Gatekeeper Report: roadmap-edit-move
**Date:** 2026-07-01
**Status:** SAFE

#### Analyzed Specs
- specs/roadmap-edit-move.feature (primary)
- specs/roadmap-crud-add-remove.feature (sibling — reused raw-feature helpers)
- specs/roadmap-json-contract.feature, specs/roadmap-doc-sync.feature, specs/roadmap-parallel-readiness.feature (skimmed for overlap)

#### Findings
- none — `rewriteDependents` (internal/roadmap/rawdeps.go) scans every phase via `decodePhase`/`setPhase`, which are dirty-map-layered (rawrender.go `phaseBytes`), so it composes correctly with the later `replaceFeatureAt` call in the same `Edit()` path; it only marks a phase dirty (`changed` flag) when a dependent was actually rewritten, so phases with no dependents round-trip byte-identically. Verified no index-shift risk: `replaceFeatureAt` uses the `featIdx` captured before `rewriteDependents` runs, and `rewriteDependents` never reorders/inserts/removes entries (only replaces values in place), so the captured index stays valid.
- none — the self-anchor no-op guards in `Move` (internal/roadmap/move.go:28-30) and `Reorder` (internal/roadmap/reorder.go:29-31) both fire immediately after the read-only `findFeature` call and strictly before any `removeFeatureAt`/`insertFeatureAt`/`setPhase` mutation — genuinely byte-identical (no write is ever attempted). `editIsNoop` (internal/roadmap/edit.go:33-35, 58-61) fires immediately after `findFeature`, before `Unmarshal`/`applyEditFields`/`applyRename` — also byte-identical. Reorder additionally has a second-layer guard (`sameOrder` before/after comparison in reorder.go:53-55) for the "already in requested position" case; this one mutates the in-memory `doc` first but never calls `finalizeMutation`/`writeRawRoadmap`, so the on-disk file is still never touched — equally safe.
- none — reuse of feature-2's `findFeature`/`removeFeatureAt`/`replaceFeatureAt` is a pure reuse with zero diff against main (confirmed via `git diff main -- internal/roadmap/rawfeature_mutate.go internal/roadmap/rawfeature_find.go`); the only change to that file is the net-new `insertFeatureAt` function. add/remove/promote call sites and behavior are unaffected.
- none — all 12 changed/added source files are ≤100 lines (61, 43, 49, 57, 66, 74, 83, 36, 50, 49, 50, 47); no G1 exception needed.
- none — cross-layer imports follow the n-tier archetype: `cmd/centinela/roadmap_{edit,move,reorder}.go` (outer layer) import only `internal/roadmap` (domain) and `internal/ui` (presentation), matching the existing `roadmap_add.go`/`roadmap_remove.go` pattern. No outer-layer business logic — command files only parse flags, build a request struct, and call into `internal/roadmap`.
- none — i18n: PROJECT.md declares `gates.i18n = false` and "English-only CLI output" for this project; the new success/error strings in the three new command files match the existing hardcoded-English precedent in `roadmap_add.go`/`roadmap_remove.go`. Not a gate violation.
- `go build ./...` and `go test ./internal/roadmap/... ./cmd/centinela/...` both pass (832 tests, 2 packages) — re-verified independently in this session, not just trusted from the tests-step report.

#### Deferred Findings
- none

#### Recommendation
- SAFE — no conflicts with sibling roadmap specs or feature-2's reused raw-feature helpers; no-op guards are correctly placed before mutation (or before the single write); ship as-is.
