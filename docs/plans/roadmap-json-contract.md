# Plan — roadmap-json-contract

> Feature 1 of 4. Brief: [docs/features/roadmap-json-contract.md](../features/roadmap-json-contract.md).
> Umbrella design: [docs/plans/roadmap-editing-suite-design.md](roadmap-editing-suite-design.md).

## Goal

Add a deterministic, machine-readable `--json` contract to the read-only roadmap
commands so Magallanes can render a Plan page by shelling out. No mutation, no
schema change to persisted `roadmap.json`.

## Deliverables

### New: `internal/roadmap/view_types.go` (~40 lines)
Derived view types (never persisted), JSON-tagged, ordered-slice only:
```go
type FeatureView struct {
    Name      string   `json:"name"`
    Phase     string   `json:"phase"`
    Status    string   `json:"status"`      // planned|in-progress|done
    Readiness string   `json:"readiness"`   // ready|blocked  (done/in-progress carry status)
    DependsOn []string `json:"dependsOn"`
    BlockedBy []string `json:"blockedBy,omitempty"`
}
type PhaseView struct {
    Name     string        `json:"name"`
    Features []FeatureView `json:"features"`
}
type StatusCounts struct {
    Planned    int `json:"planned"`
    InProgress int `json:"inProgress"`
    Done       int `json:"done"`
}
type RoadmapView struct {
    Phases []PhaseView  `json:"phases"`
    Counts StatusCounts `json:"counts"`
}
```

### New: `internal/roadmap/view.go` (~55 lines)
```go
func BuildView(r *Roadmap) RoadmapView
```
Iterates `r.Phases` → `phase.Features` in order. Per feature: `Status =
FeatureStatus(f.Name)`; readiness/blockedBy from the existing
`classifyFeature`/`DeriveReadiness` machinery in `readiness.go` (reuse — do not
re-derive). `counts` tallies schedulable features by status exactly as
`Summary()` scopes them (Backlog/Baseline excluded from counts). Byte-stable:
no map ranging.

> If `view.go` + `view_types.go` risk the 100-line cap, keep them as two files
> (types vs builder), which is already the split above.

### Changed: `internal/roadmap/readiness.go`
Add `json:"..."` tags to `FeatureReadiness` fields (currently untagged) so
readiness data can serialize; keep the `-`-tagged fields internal where they
must stay unexported from the contract. Do not change behavior.

### Changed: `cmd/centinela/roadmap.go`
Add `var roadmapJSON bool` + `Flags().BoolVar(&roadmapJSON,"json",false,…)`. In
`runRoadmap`: if `roadmapJSON`, `json.MarshalIndent(roadmap.BuildView(r),"","  ")`
to stdout; else the existing `ui.RenderRoadmap` path unchanged.

### Changed: `cmd/centinela/roadmap_ready.go`
Add `var readyJSON bool`. If set, `json.MarshalIndent(roadmap.ReadySet(r))`
(array of names); else existing `ui.RenderReadyList`.

### New: `cmd/centinela/roadmap_show.go` (~45 lines)
`roadmap show` (alias `list`): text → `ui.RenderRoadmap(r)`; `--json` →
`json.MarshalIndent(r)` (the persisted typed `Roadmap` verbatim). Registered via
`init()` + `roadmapCmd.AddCommand`, `Aliases:[]string{"list"}`, mirroring
`roadmap_ready.go`.

## Reuse (do not reimplement)
- `roadmap.Load()` — `internal/roadmap/roadmap.go:16` (loads + validates).
- `FeatureStatus(name)` — `roadmap.go:45`.
- `DeriveReadiness`/`classifyFeature`/`ReadySet` — `readiness.go`.
- `Summary()` scoping (schedulable set) — `roadmap.go:60`, `backlog.go`.
- `--json` flag + `json.MarshalIndent` pattern — `cmd/centinela/dashboard.go`, `verdict.go`.

## Constraints
- Every source + `_test.go` file ≤ 100 lines.
- Commands stay thin: no view logic in `cmd/` beyond the one-line marshal.
- Deterministic/byte-stable JSON (ordered slices only).
- No mutation; no change to persisted `roadmap.json` schema.

## Tests (colocated for coverage — aim ≥97%)
- `internal/roadmap/view_test.go` — `BuildView` maps status/readiness/blockedBy;
  counts correct; empty roadmap → empty phases + zero counts; byte-stable marshal
  (compare exact bytes across two builds).
- `cmd/centinela/roadmap_json_test.go` — `roadmap --json` / `ready --json` /
  `show --json` flag parsing and shape; missing-file error path.
- Extend/verify no regression in existing `roadmap` text tests.

## Verification (end-to-end)
1. `go test ./...` green; `scripts/check-coverage.sh` ≥95% (target ≥97%).
2. Build a local binary (`go build -o /tmp/centinela-dev ./cmd/centinela`) and run
   `/tmp/centinela-dev roadmap --json`, `roadmap ready --json`, `roadmap show
   --json` against `.workflow/roadmap.json` — assert valid JSON and stable bytes
   on repeated runs (`diff <(… ) <(…)`).
3. `centinela validate` passes in the worktree.
