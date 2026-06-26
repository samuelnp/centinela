# Plan — team-dashboard

A read-only, single-shot board that aggregates current multi-feature /
multi-contributor Centinela state into three panels: **in-flight features**,
**roadmap burn-down**, and **gate health**. It mirrors `centinela insights`
exactly: a pure aggregator package (`internal/teamdashboard`) computes a
serializable `Dashboard` from in-memory inputs, a thin `cmd/` orchestrator
reads the sources from disk, and `internal/ui` renders. No I/O, no git, no
network in the aggregator.

## 1. Architecture — `internal/teamdashboard` (aggregator layer)

A new pure package on the **aggregator** layer, peer to `internal/insights`.
It performs **no I/O** and **does not shell out to git**: the `cmd/` caller
reads every source from disk, derives owners via a git seam, and passes a
plain `Inputs` struct in. `Compute` is deterministic — every list it emits is
sorted by a stable key, never ranged from a map in output order.

### Public API

```go
package teamdashboard

// Inputs is the plain, caller-populated aggregate. The package reads nothing
// from disk: cmd/ fills every field. now is injected for deterministic ages.
type Inputs struct {
	Active  []*workflow.Workflow // workflow.ActiveWorkflows(".workflow")
	Roadmap *roadmap.Roadmap     // roadmap.Load() result, nil when absent/unreadable
	Events  []telemetry.Event    // telemetry.ReadDefault() result
	Owners  map[string]string    // feature -> git-derived owner ("unknown" allowed)
	Now     time.Time            // age reference; cmd/ passes time.Now().UTC()
}

// Dashboard is the pure, serializable board. Field names are a stable --json
// contract; do not rename without bumping consumers.
type Dashboard struct {
	Features []FeatureRow    // one row per active workflow (input order preserved)
	Roadmap  RoadmapBurndown // schedulable-phase counts + overall
	Gates    []GateHealth    // gate-failure tallies, ranked desc/asc
}

type FeatureRow struct {
	Feature   string // wf.Feature
	Step      string // wf.CurrentStep
	StepIndex int    // done-count (0-based position in OrderedSteps)
	StepTotal int    // len(wf.OrderedSteps())
	AgeDays   int    // floor((Now - wf.StartedAt) / 24h); 0 if StartedAt zero/future
	Profile   string // wf.EnforcementProfile, "" -> "default" at render only
	Archetype string // wf.Archetype, "" -> "canonical" at render only
	Worktree  string // wf.WorktreePath ("" when single-checkout)
	Owner     string // Inputs.Owners[feature]; "unknown" when missing
}

type RoadmapBurndown struct {
	Present    bool          // false when Inputs.Roadmap == nil (empty state)
	Planned    int           // from Roadmap.Summary()
	InProgress int           // from Roadmap.Summary()
	Done       int           // from Roadmap.Summary()
	Total      int           // Planned + InProgress + Done (schedulable only)
	Phases     []PhaseStatus // per-schedulable-phase done/total, file order
}

type PhaseStatus struct {
	Name  string
	Done  int
	Total int
}

type GateHealth struct {
	Gate  string // gate name ("<none>" bucket inherited from insights.Gates)
	Fails int    // gate-failure count
}

// Compute turns Inputs into a pure Dashboard. Empty/nil sources each yield an
// honest empty state, never a panic: no Active -> empty Features; nil Roadmap
// -> RoadmapBurndown{Present:false}; no gate-failure events -> empty Gates.
func Compute(in Inputs) Dashboard
```

### Concrete files (each < 100 lines)

| File | Responsibility |
|------|----------------|
| `internal/teamdashboard/dashboard.go` | Package doc + all type declarations (`Inputs`, `Dashboard`, `FeatureRow`, `RoadmapBurndown`, `PhaseStatus`, `GateHealth`). Types only. |
| `internal/teamdashboard/compute.go` | `Compute(in Inputs) Dashboard` — the single entry point; calls `features`, `burndown`, `gatehealth`. |
| `internal/teamdashboard/features.go` | `features(active []*workflow.Workflow, owners map[string]string, now time.Time) []FeatureRow` + `ageDays` + `ownerOf` (defaults missing to `"unknown"`). Uses `wf.OrderedSteps()` for total and the done-count pattern from `ui.wfDoneCount` for `StepIndex`. |
| `internal/teamdashboard/burndown.go` | `burndown(r *roadmap.Roadmap) RoadmapBurndown` — nil → `{Present:false}`; else `r.Summary()` + per-schedulable-phase `PhaseStatus` (skips Backlog/Baseline via the roadmap predicate; counts `FeatureStatus=="done"`). |
| `internal/teamdashboard/gatehealth.go` | `gatehealth(events []telemetry.Event, topN int) []GateHealth` — delegates to `insights.Gates(events, topN)` and maps `[]insights.Count` → `[]GateHealth` (aggregator→aggregator import, allowed). topN is a package const (e.g. 10) so failure counts never diverge from `insights`. |

Imports: `internal/workflow`, `internal/roadmap`, `internal/telemetry`,
`internal/insights` (all read-only) + stdlib (`time`, `sort`). No `cmd/`, no
`internal/ui`, no `os/exec`.

## 2. The three panels — data, sourcing, degradation

### Panel 1 — In-flight features
One `FeatureRow` per `workflow.ActiveWorkflows(".workflow")` entry (already
deduped by feature, sorted mtime-desc — we preserve that order).

| Datum | Source |
|-------|--------|
| Feature | `wf.Feature` |
| Step + `X/5` | `wf.CurrentStep`; `StepIndex`=done-count, `StepTotal`=`len(wf.OrderedSteps())` |
| Age (days) | `floor((Now - wf.StartedAt)/24h)`; zero/future `StartedAt` → 0 |
| Profile / Archetype | `wf.EnforcementProfile` / `wf.Archetype` (renderer fills "default"/"canonical" when empty) |
| Worktree | `wf.WorktreePath` ("" → renderer shows "—") |
| Owner | `Inputs.Owners[feature]`, `"unknown"` when absent |

**Degradation:** no active workflows → empty `Features`; renderer prints
"no active features — run `centinela start <feature>`".

### Panel 2 — Roadmap burn-down
`roadmap.Load()` in `cmd/`; aggregator uses `r.Summary()` for the overall
planned/in-progress/done (schedulable phases only — Backlog/Baseline excluded
by the existing `isNonSchedulablePhase` predicate) and walks `r.Phases` for
per-phase `Done/Total` (`Total` = schedulable feature count in that phase,
`Done` = features with `FeatureStatus=="done"`). Overall `Total` =
Planned+InProgress+Done. Renderer shows the `N/M done` line.

**Degradation:** `roadmap.Load()` error (missing/unreadable) → `cmd/` passes
`Roadmap: nil` → `RoadmapBurndown{Present:false}`; renderer prints "no roadmap
— run `centinela roadmap …`". An empty roadmap (zero schedulable features)
renders `0/0 done`, not an error.

### Panel 3 — Gate health
`telemetry.ReadDefault()` in `cmd/`; aggregator calls `insights.Gates(events,
topN)` and maps to `[]GateHealth`. Only `Type=="gate-failure"` events count;
empty `Gate` buckets under `<none>` (inherited from `insights`).

**Degradation:** telemetry missing/empty, or present-but-no-gate-failures →
empty `Gates`; renderer prints "no gate failures recorded". A telemetry read
**error** is surfaced by `cmd/` (return err) only if `ReadDefault` itself
errors hard; a missing log returns empty, matching `insights`.

## 3. Owner derivation — git seam in `cmd/` (best-effort, never in aggregator)

Git stays entirely out of `internal/teamdashboard`. In `cmd/centinela/`:

```go
// gitOwner returns the latest committer name on the feature's branch, or
// "unknown" on any error / no commits. Overridable in tests.
var gitOwner = func(repoRoot, feature string) string { /* exec git log -1 */ }
```

Implementation: best-effort `git log -1 --format=%an <branch>` (branch = the
feature slug; falls back gracefully). Any non-zero exit, empty output, or
error → `"unknown"`. `cmd/` builds `owners := map[string]string{}` over the
active features and passes it into `Inputs.Owners`. The overridable
package-level `var` is the test seam (set in integration tests to avoid real
git), mirroring existing `cmd/` seams. The owner column is **advisory** — a
flaky/empty git never fails the command.

## 4. cmd wiring + renderer

### `cmd/centinela/dashboard.go` (thin orchestrator, G7-clean)
- New `dashboardCmd` (cobra), `--json` bool flag, registered in `init()`.
- `runDashboard`:
  1. `active := workflow.ActiveWorkflows(workflow.WorkflowDir)`
  2. `rm, _ := roadmap.Load()` (ignore error → nil for empty state)
  3. `events, err := telemetry.ReadDefault()` (hard error → return err)
  4. `owners` via `gitOwner` seam over active features
  5. `dash := teamdashboard.Compute(teamdashboard.Inputs{Active: active, Roadmap: rm, Events: events, Owners: owners, Now: time.Now().UTC()})`
  6. `--json` → `json.MarshalIndent(dash)`; else `ui.RenderDashboard(dash)`
- No business logic: every decision is in the aggregator; cmd only reads,
  derives owners, and routes output (mirrors `insights.go`).

### `internal/ui/render_dashboard.go` (pure presentation)
- `RenderDashboard(d teamdashboard.Dashboard) string` — three Lipgloss panels
  joined vertically, house style (`StyleBold`/`StyleMuted`, `renderSystemPanel`
  or section helpers from `render_insights.go`).
- Each panel renders its own empty state (see §2). ANSI auto-strips on non-TTY
  (Lipgloss default, as `RenderInsights` already relies on) so piped/`--json`-
  adjacent output is plain and parseable.
- Imports `internal/teamdashboard` read-only (its `Dashboard` type), exactly
  as `render_insights.go` imports `internal/insights`. Split into a helper
  file (`render_dashboard_panels.go`) if it exceeds 100 lines.

## 5. PROJECT.md G2 paragraph + centinela.toml import_graph

### PROJECT.md (append to the G2 rule, mirroring the `insights` allowance)
> `internal/teamdashboard` also joins the **aggregator** layer: a read-only
> team-status aggregator for `centinela dashboard` that may import
> `internal/workflow` (domain), `internal/roadmap` (domain), the
> `internal/telemetry` leaf, and `internal/insights` (aggregator) read-only
> plus stdlib; it must not import `cmd/` or `internal/ui` and is itself
> imported only by `cmd/` (its `Dashboard` type by `internal/ui` for
> rendering). Its `workflow`/`roadmap` edges are domain (allowed); its
> `teamdashboard → insights` edge is aggregator→aggregator (allowed via the
> aggregator layer's `allow: ["domain","leaf","aggregator"]`);
> `workflow`/`roadmap`/`telemetry`/`insights` never import `teamdashboard`, so
> there is no cycle.

Also add `| Aggregator — team dashboard | internal/teamdashboard/ |` to the
Gatekeeper Paths table.

### centinela.toml
Add `"internal/teamdashboard/**"` to the existing aggregator layer `paths`
(its `allow` already includes `domain`/`leaf`/`aggregator`, so no `allow`
change). Append a comment mirroring the `insights`/`brownmap` notes describing
the edges. No new layer block needed; the existing `cmd` layer already allows
aggregator imports.

## 6. Rollout — smallest correct slice first

- **Slice 1 (burn-down + in-flight):** `dashboard.go` types, `compute.go`,
  `features.go`, `burndown.go`; `cmd/centinela/dashboard.go` (owners stubbed to
  `"unknown"` for now), `render_dashboard.go` with panels 1 & 2 + empty states;
  the full G2 mapping (PROJECT.md paragraph + toml `paths`) lands in this slice
  so the import_graph gate is green from the start; `--json` emits the partial
  `Dashboard` (Gates empty).
- **Slice 2 (gate health + owner git seam):** `gatehealth.go` +
  `insights.Gates` wiring + panel 3; replace the owner stub with the real
  `gitOwner` seam in `cmd/`. No G2 change (insights edge already covered by the
  aggregator allow). Adds gate-health and real owners to `--json`.

Each slice is independently shippable, leaves the gate green, and the
aggregator stays pure throughout.

## 7. Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| No persisted owner field (git-derived gap) | Med | High | Best-effort `gitOwner` seam in `cmd/`; `"unknown"` fallback; owner is advisory, never fails the command; documented Out-of-Scope (real owner model = separate feature). |
| `gitOwner` flakiness (detached HEAD, no commits, no git) | Low | Med | Any error/empty → `"unknown"`; seam is overridable in tests so integration tests never touch real git; git fully excluded from the pure aggregator. |
| import_graph mapping wrong → gate fails | High | Low | Land the PROJECT.md paragraph + toml `paths` in Slice 1; only the `teamdashboard → insights` edge is new (aggregator→aggregator, already allowed); dry-run `centinela validate` before completing. |
| File-size G1 (>100 lines) | Med | Med | Split per §1 (types / compute / features / burndown / gatehealth); split renderer into `render_dashboard_panels.go` if needed; test files also ≤100 lines (per-file G1 applies to `_test.go`). |
| Empty-state correctness (no wf / no telemetry / no roadmap) | Med | Med | Each panel owns its empty state; `nil Roadmap` → `Present:false`; missing telemetry → empty `Gates` (matches `insights`); explicit unit tests per empty source. |
| Coupling to `insights.Gates` (ranking divergence) | Low | Low | Reuse `insights.Gates` verbatim (the prompt-mandated dedup); pin `topN` as a package const; a unit test asserts `GateHealth` mirrors `insights.Gates` output. |

## 8. Test strategy

- **Unit (colocated, 95% per-package coverage gate):**
  `internal/teamdashboard/compute_test.go`, `features_test.go`,
  `burndown_test.go`, `gatehealth_test.go` — drive `Compute` with in-memory
  `Inputs` (no disk, no git). Cover: age math (zero/future/normal `StartedAt`),
  step index/total, owner `"unknown"` fallback, nil-Roadmap empty state,
  schedulable-phase filtering vs `Summary()`, gate mapping vs `insights.Gates`,
  and the three empty-source states. Per-package coverage means these
  colocated tests (each ≤100 lines) move the gate — `tests/` tier files do
  not.
- **cmd integration (`tests/integration/`):** drive `runDashboard` with the
  `gitOwner` seam overridden (deterministic owners, no real git) and a
  temp-`.workflow`/temp-telemetry fixture; assert `--json` shape and text
  output, including empty states.
- **Acceptance (`tests/acceptance/`, local only — NO network):** build the
  `centinela` binary, run `centinela dashboard` and `centinela dashboard
  --json` inside a temp repo seeded with workflow JSONs, `roadmap.json`, and a
  telemetry log; assert the three panels / JSON keys and the honest empty
  states. No `git push`, no network — any git is a local temp repo (avoids the
  acceptance-hang failure mode). Wire the acceptance invocation into
  `validate.commands` (already runs `go test ./tests/acceptance/...`).
- Add a `.feature` spec in `specs/team-dashboard.feature` covering the panels,
  owner fallback, roadmap match, and gate-health empty state (acceptance
  scenarios map to it for the spec-traceability gate).
