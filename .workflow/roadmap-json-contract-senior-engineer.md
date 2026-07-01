# roadmap-json-contract — senior-engineer

## Files Touched

New:
- `internal/roadmap/view_types.go` (37 lines) — derived, JSON-tagged view types (`FeatureView`, `PhaseView`, `StatusCounts`, `RoadmapView`). Never persisted.
- `internal/roadmap/view.go` (70 lines) — `BuildView(r *Roadmap) RoadmapView` + helpers (`readinessIndex`, `buildFeatureView`, `tally`).
- `cmd/centinela/roadmap_show.go` (43 lines) — new `roadmap show` (alias `list`): text via `ui.RenderRoadmap`, `--json` dumps persisted `Roadmap` verbatim.

Changed:
- `cmd/centinela/roadmap.go` (42 lines) — added `--json` flag; emits `BuildView(r)` as indented JSON, else unchanged text path. Flag var named `roadmapViewJSON` to avoid colliding with the existing `roadmapJSON(...)` test helper.
- `cmd/centinela/roadmap_ready.go` (46 lines) — added `--json` flag; emits `ReadySet(r)` as a JSON array (nil coalesced to `[]`), else unchanged text path.

Not touched: `readiness.go` (BuildView reads `FeatureReadiness` Go fields directly; no JSON tags needed — the plan's optional tag change was unnecessary, kept the change minimal).

## Architecture Compliance

- All 5 source files ≤100 lines (37/70/42/46/43). ≤100-line rule satisfied with margin.
- Layer rules respected: all view/projection logic lives in `internal/roadmap`; `cmd/` files are thin (flag + one-line marshal + existing text branch), matching the `--json` pattern in `dashboard.go`/`verdict.go`.
- Read-only, no mutation: `BuildView` and the show/ready paths only read; persisted `roadmap.json` schema untouched.
- Scoping consistency: `BuildView` and counts skip non-schedulable phases via the shared `isNonSchedulablePhase`, matching `Summary()`/`DeriveReadiness`/`ReadySet`. `roadmap show --json` intentionally dumps the persisted `Roadmap` verbatim (includes Backlog/Baseline) — verified live: show=14 phases (incl. Backlog), view=13.
- Byte-stability: ordered-slice iteration only, no map ranging; readiness indexed into a map but iteration for output is over `r.Phases`→`Features`. Verified identical bytes across two runs for all three surfaces.
- Determinism/i18n: reused existing sibling error helper `roadmapCommandError`; no new hardcoded user-facing strings beyond cobra flag `Short`/usage strings, consistent with sibling roadmap commands.

## Type-Safety Notes

- Strictly typed throughout; no `interface{}`/`any` introduced. `json.MarshalIndent` (whose param is `any`) is called inline per command exactly as `dashboard.go` does — no `any`-typed wrapper added.
- `DependsOn` declared without `omitempty` and nil-coalesced to `[]string{}` so it always serializes as `[]` (persisted input contract). `Readiness` and `BlockedBy` use `omitempty`; readiness/blockedBy only populated for planned `ready`/`blocked` rows.

## Trade-Offs

- `buildFeatureView` calls `FeatureStatus(f.Name)` while readiness also came from `DeriveReadiness` (which internally calls `FeatureStatus`) — a second workflow-state read per feature. Chosen to follow the plan's explicit `Status = FeatureStatus(f.Name)` and keep status/readiness independent dimensions; both reads are deterministic against the same on-disk state, so no consistency risk. Read-only CLI, negligible cost.
- nil→`[]` coalescing for `ready --json` lives in the command (one line) rather than changing `ReadySet`'s nil-return contract, to avoid touching shared code used by the text path.

## Verification

- `go build ./...` OK; `go vet ./...` clean (No issues found).
- Dev binary smoke test (`/tmp/centinela-dev`): `roadmap --json`, `roadmap ready --json`, `roadmap show --json` all valid JSON; `list` alias byte-identical to `show`; all three surfaces byte-stable across two runs; `ready --json` set == `readiness:"ready"` set from `roadmap --json`; done rows omit readiness; blockedBy only on blocked rows; dependsOn always present; plain `roadmap` and `roadmap show` text byte-identical to each other and stable; missing-file path exits non-zero with no partial stdout JSON.

## Deferred Findings

None. No `centinela roadmap defer` calls made.

## Handoff

- **Next role:** qa-senior
- Cover: `BuildView` mapping (status/readiness/blockedBy/dependsOn), counts scoping (Backlog/Baseline excluded), empty-roadmap `{"phases":[],"counts":{0,0,0}}`, byte-stable marshal, nil-coalescing for `dependsOn` and `ready --json`, the show/view scoping asymmetry, and the three missing/malformed error paths. Tests are colocated `_test.go` (≤100 lines each) for the 95%/target-97% coverage gate; no test files were written in this code step.
