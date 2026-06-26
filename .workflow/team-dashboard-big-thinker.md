# team-dashboard — big-thinker

## Problem

Multi-feature / multi-contributor Centinela state is invisible without manual
polling. Today you must run `centinela status <f>` per feature across every
`.worktrees/<feature>` and eyeball `ROADMAP.md` to learn which features are in
flight, what step each is stuck on, how old they are, who last touched them,
and how the roadmap is burning down. `centinela insights` aggregates the
*event log* but there is no single read-only view of *current* state. We add a
`centinela dashboard` board with three panels (in-flight features, roadmap
burn-down, gate health) built as a pure aggregator mirroring `internal/insights`.

## Scope

**In:**
- New pure aggregator `internal/teamdashboard` with `Compute(in Inputs) Dashboard`
  (no I/O, no git, no `internal/ui`/`cmd` imports). Files: `dashboard.go`
  (types), `compute.go`, `features.go`, `burndown.go`, `gatehealth.go` (each
  <100 lines).
- `cmd/centinela/dashboard.go` thin orchestrator: reads active workflows,
  `roadmap.Load()`, `telemetry.ReadDefault()`, derives owners via a git seam,
  calls `Compute`, renders or emits `--json`.
- `internal/ui/render_dashboard.go` pure presentation: three Lipgloss panels,
  per-panel empty states, ANSI auto-strip on non-TTY.
- Git owner seam in `cmd/` (overridable var, best-effort, `"unknown"` fallback).
- PROJECT.md G2 paragraph + `centinela.toml` aggregator `paths` addition.

**Out:**
- Live/watch mode (single-shot snapshot only).
- Cross-repo / multi-tenant fleet aggregation (that is Magallanes).
- A persisted owner/assignee field on workflow state (ownership is git-derived;
  a real owner model is a separate feature).
- Running gates live for the board — gate health comes from the telemetry log,
  not a fresh `gates.RunAll`.
- Historical time-series / trends beyond `centinela insights`.

## Dependencies & Assumptions

Verified real APIs (read, not assumed):
- `workflow.ActiveWorkflows(".workflow") []*Workflow` (deduped, mtime-desc),
  `CapActive`, `Workflow{Feature, StartedAt time.Time, CurrentStep, Steps,
  StepOrder, EnforcementProfile, Archetype, WorktreePath, DriverModel}`,
  `wf.OrderedSteps()`; done-count pattern from `ui.wfDoneCount`.
- `roadmap.Load() (*Roadmap, error)`, `(*Roadmap).Summary() (planned,
  inProgress, done int)` (schedulable phases only; Backlog/Baseline excluded by
  `isNonSchedulablePhase`), `FeatureStatus(name) string`, `Roadmap.Phases`.
- `telemetry.ReadDefault() ([]Event, error)`, `Event{Type, Gate, …}`,
  `TypeGateFailure == "gate-failure"`; `insights.Gates(events, topN) []Count`
  reused verbatim for gate ranking (aggregator→aggregator import allowed).
- `worktree.Dir`/`Path`/`Exists` available, but the in-flight panel reads the
  worktree path straight off `wf.WorktreePath` (no scan needed).

Assumptions:
- **GAP — no owner field anywhere.** Owner is git-derived (`git log -1
  --format=%an <branch>`) via an overridable `cmd/` seam; no commits / any error
  → `"unknown"`. Git is kept entirely out of the pure aggregator.
- A missing roadmap/telemetry is an empty state, never an error: `Roadmap: nil`
  → `RoadmapBurndown{Present:false}`; missing telemetry log → empty `Gates`
  (matches `insights`). Only a hard `ReadDefault` error propagates.
- The aggregator layer's `allow` already includes `domain`/`leaf`/`aggregator`,
  so the only G2 change is adding `internal/teamdashboard/**` to `paths`.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| No persisted owner field (git-derived gap) | Med | High | Best-effort `gitOwner` seam in `cmd/`; `"unknown"` fallback; advisory column never fails the command; real owner model is Out-of-Scope. |
| `gitOwner` flakiness (no commits / detached HEAD / no git) | Low | Med | Any error/empty → `"unknown"`; overridable seam so tests never touch real git; git excluded from the aggregator. |
| import_graph mapping wrong → gate fails | High | Low | Land PROJECT.md paragraph + toml `paths` in Slice 1; only new edge `teamdashboard → insights` is aggregator→aggregator (already allowed); dry-run `centinela validate`. |
| File-size G1 (>100 lines incl. `_test.go`) | Med | Med | Split per types/compute/features/burndown/gatehealth; split renderer if needed; keep test files ≤100 lines. |
| Empty-state correctness (no wf / telemetry / roadmap) | Med | Med | Each panel owns its empty state; explicit unit test per empty source. |
| Coupling to `insights.Gates` divergence | Low | Low | Reuse `insights.Gates` verbatim; pin `topN` const; unit test asserts `GateHealth` mirrors it. |

## Rollout

- **Slice 1 — burn-down + in-flight:** aggregator types + `compute.go` +
  `features.go` + `burndown.go`; `cmd/centinela/dashboard.go` (owner stubbed
  `"unknown"`); `render_dashboard.go` panels 1 & 2 + empty states. Full G2
  mapping (PROJECT.md + toml `paths`) lands here so the gate is green from the
  start. `--json` emits the partial `Dashboard`.
- **Slice 2 — gate health + owner git seam:** `gatehealth.go` + `insights.Gates`
  wiring + panel 3; replace owner stub with the real `gitOwner` seam. No G2
  change (insights edge already covered). Adds gate-health + real owners to
  `--json`.

Each slice is independently shippable and leaves the import_graph gate green;
the aggregator stays pure throughout.

## Deferred Findings

none

## Handoff

→ **feature-specialist.** Plan doc: `docs/plans/team-dashboard.md`. Build
Slice 1 first (in-flight + burn-down panels + G2 mapping), then Slice 2 (gate
health + git owner seam). Keep `internal/teamdashboard` pure (no I/O, no git,
no `internal/ui`). Owner derivation lives behind an overridable `cmd/` var.
Tests: colocated `internal/teamdashboard/*_test.go` (≤100 lines each) for the
95% per-package gate; cmd integration with the git seam overridden; acceptance
driving `centinela dashboard` + `--json` in a temp repo (local only, no
network); add `specs/team-dashboard.feature`.
