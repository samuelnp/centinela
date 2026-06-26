# Senior-Engineer Report — team-dashboard

## Summary
Implemented `centinela dashboard`: a read-only, single-shot board aggregating
multi-feature Centinela state into three panels (in-flight features, roadmap
burn-down, gate health). Mirrors `centinela insights` exactly: a pure
`internal/teamdashboard` aggregator computes a serializable `Dashboard` from
in-memory `Inputs`; a thin `cmd/centinela/dashboard.go` reads sources off disk
and routes output; `internal/ui` renders. Both Slice 1 (in-flight + burn-down +
G2 mapping) and Slice 2 (gate health + real owner git seam) are complete and
build-green.

## Files Touched
Created:
- `internal/teamdashboard/dashboard.go` (73) — package doc + all types (`Inputs`, `Dashboard`, `FeatureRow`, `RoadmapBurndown`, `PhaseStatus`, `GateHealth`).
- `internal/teamdashboard/compute.go` (17) — `Compute(in Inputs) Dashboard` + `gateTopN` const.
- `internal/teamdashboard/features.go` (65) — `features` + `doneCount` + `ageDays` + `ownerOf`.
- `internal/teamdashboard/burndown.go` (36) — `burndown` over schedulable phases.
- `internal/teamdashboard/gatehealth.go` (21) — `gatehealth` mapping `insights.Gates` → `[]GateHealth`.
- `cmd/centinela/dashboard.go` (63) — `dashboardCmd` cobra command + `--json`, thin `runDashboard`.
- `cmd/centinela/dashboard_owner.go` (42) — overridable `gitOwner` seam + `dashboardOwners`.
- `internal/ui/render_dashboard.go` (29) — `RenderDashboard` + header.
- `internal/ui/render_dashboard_panels.go` (79) — the three panel renderers + empty states.

Modified:
- `PROJECT.md` — appended the `teamdashboard` G2 aggregator paragraph; added the Gatekeeper Paths row; added the Folder Structure entry.
- `centinela.toml` — added `internal/teamdashboard/**` to the aggregator layer `paths` with a comment describing the edges.

## Architecture Compliance
- Aggregator (`internal/teamdashboard`) is PURE: no I/O, no git, no `os/exec`,
  no `cmd/`/`internal/ui` imports. Imports only `workflow`+`roadmap` (domain),
  `telemetry` (leaf), `insights` (aggregator), and stdlib (`time`).
- `teamdashboard → insights` is the only new edge: aggregator→aggregator,
  already allowed by the aggregator layer's `allow: ["domain","leaf","aggregator"]`.
- `cmd/` stays thin (G7): reads sources, derives owners via the seam, calls
  `Compute`, routes to `--json` or `RenderDashboard`. No business logic.
- `internal/ui` imports `internal/teamdashboard` read-only for the `Dashboard`
  type only (mirrors `render_insights.go`).
- `centinela validate`: import_graph reports NO forbidden cross-layer edge for
  `teamdashboard` (only the pre-existing generic "no configured layer" warning,
  non-failing). All files ≤100 lines (G1 ✓).

## Type-Safety Notes
- `Compute` is deterministic: features preserve `ActiveWorkflows` order, phases
  follow roadmap file order, gates inherit `insights.Gates` ranking — no map is
  ranged in output order.
- `--json` field names are the stable contract (`Features`, `Roadmap`, `Gates`
  and their sub-fields), verified against the spec.
- `ageDays` guards zero/future `StartedAt` → 0 (no negative age). `ownerOf`
  defaults missing/empty → "unknown". `burndown(nil)` → `{Present:false}`.
- `gitOwner` is a package-level `var` so tests override it without touching real
  git; any error/empty → "unknown" (advisory column never fails the command).

## Trade-Offs
- `gatehealth.go` (a Slice-2 file) was written in Slice 1 so `Compute` is
  complete and the package builds green in one pass; it adds no new import edge
  beyond the already-mapped `insights` edge.
- Renderer uses house `StyleBold`/`StyleMuted` section style (like
  `RenderInsights`), not Lipgloss boxes, for compact plain-on-pipe output.

## Deferred Findings
None.

## Handoff
Handoff to **qa-senior**. Tests step must add:
- Colocated unit tests (`internal/teamdashboard/*_test.go`, each ≤100 lines) for
  per-package 95% coverage: age math (zero/future/normal), step index/total,
  owner fallback, nil-Roadmap empty state, schedulable-phase filtering, gate
  mapping vs `insights.Gates`, three empty-source states.
- cmd integration tests (`tests/integration/`) with the `gitOwner` seam
  overridden + temp `.workflow`/telemetry fixtures.
- Acceptance tests (`tests/acceptance/`, local only, NO network) building the
  binary against a temp repo; wire into `validate.commands`.
- `.workflow/team-dashboard-edge-cases.md`.
