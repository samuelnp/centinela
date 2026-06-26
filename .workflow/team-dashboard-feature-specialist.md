# team-dashboard — feature-specialist

## Behavior Summary

`centinela dashboard` is a single-shot, read-only board that aggregates current
multi-feature state into three terminal panels without touching any files.

The command reads three on-disk sources — active workflow JSONs, `roadmap.json`,
and the telemetry event log — passes them as a plain `Inputs` struct into the
pure `internal/teamdashboard.Compute` aggregator, and routes the resulting
`Dashboard` either to `internal/ui.RenderDashboard` (human panels) or
`json.MarshalIndent` (`--json` flag). No source is mandatory: each absent or
unreadable source yields an honest empty-state panel and the command still exits
zero.

**Panel 1 — In-flight features.** One row per `workflow.ActiveWorkflows` entry
(mtime-descending, already deduped). Each row shows: feature name, current step,
`StepIndex/StepTotal` progress (e.g. `2/5`), age in floor-days from `StartedAt`,
enforcement profile (`"default"` when blank), archetype (`"canonical"` when
blank), worktree path (`"—"` when blank), and a git-derived owner. Owner comes
from `git log -1 --format=%an <branch>` via an overridable `cmd/` seam; any
error or empty output resolves to `"unknown"` without aborting the row or the
command.

**Panel 2 — Roadmap burn-down.** Per-schedulable-phase done/total counts plus an
overall `N/M done` line, sourced from `roadmap.Summary()`. Non-schedulable
phases (`Backlog`, `Baseline`) are excluded by the existing predicate. A nil
roadmap (absent or unreadable file) renders `RoadmapBurndown{Present:false}` and
the panel shows an empty-state message.

**Panel 3 — Gate health.** Gate-failure tallies ranked by count descending,
derived by delegating to `insights.Gates(events, topN)` and mapping to
`[]GateHealth`. Only `type=="gate-failure"` events count; an empty `Gate` field
buckets under `<none>`. Missing or empty telemetry yields an empty panel, not an
error.

The `internal/teamdashboard` package is a pure aggregator: no I/O, no git, no
`cmd/` or `internal/ui` imports. All disk reads and git calls live in
`cmd/centinela/dashboard.go`. The `--json` contract (`Dashboard` field names) is
stable and must not be renamed without bumping consumers.

## Acceptance Criteria (Gherkin)

Spec: `specs/team-dashboard.feature`

Scenario groups covered:

- **Happy path / full board** — three panels rendered, exit 0, no files written.
- **In-flight row content** — step, X/5 progress, age math (floor-days, zero/future
  StartedAt clamp), blank profile/archetype/worktree defaults, mtime-desc ordering.
- **Owner fallback** — git-derived name present; no commits → `"unknown"`; git
  unavailable → `"unknown"`; per-feature error does not abort other rows.
- **Roadmap burn-down** — per-phase counts, overall done/total line, schedulable-only
  filtering (Backlog/Baseline excluded), zero-feature roadmap shows `0/0 done`.
- **Gate health** — ranked by count desc, `<none>` bucket for empty gate name, ranking
  matches `insights.Gates`, non-gate-failure events excluded.
- **Empty/degraded states** — no workflows, no telemetry, no roadmap, all three absent
  simultaneously — each an honest empty panel, exit 0.
- **`--json`** — stable field names, correct shape including `Roadmap.Present:false`,
  no ANSI, deterministic across two runs.
- **Aggregator purity** — `Compute` makes no I/O or git calls; deterministic.
- **Non-TTY** — piped output strips ANSI for both human and `--json` modes.

## UX States

This is a CLI tool. Graphical states are n/a.

| Surface | State | Rendering |
|---------|-------|-----------|
| Terminal — in-flight panel | Active features present | Table rows: feature, step, X/5, age, profile, archetype, worktree, owner |
| Terminal — in-flight panel | No active workflows | `"no active features — run centinela start <feature>"` |
| Terminal — roadmap panel | Roadmap present | Per-phase Done/Total rows + overall `N/M done` line |
| Terminal — roadmap panel | Roadmap absent/unreadable | `"no roadmap — run centinela roadmap …"` |
| Terminal — roadmap panel | Roadmap present, zero schedulable features | `"0/0 done"` |
| Terminal — gate health panel | Gate failures present | Ranked list: gate name, fail count |
| Terminal — gate health panel | No gate failures / no telemetry | `"no gate failures recorded"` |
| `--json` output | Any state | Indented JSON, `Dashboard` struct, no ANSI |
| Piped / non-TTY | Any state | All ANSI stripped by Lipgloss default |
| Graphical states | — | n/a (CLI only) |

## Edge Cases

1. Owner derivation failure for one feature must not abort the remaining in-flight rows
   or change the exit code — failed owner resolves to `"unknown"`, dashboard continues.
2. Zero `StartedAt` / future `StartedAt` must clamp age to `0d` — no negative values,
   no panic.
3. Backlog and Baseline phases must be excluded from the burn-down panel even if their
   feature count is non-zero.
4. All three sources absent simultaneously must still render three valid empty-state
   panels and exit 0.
5. Gate-failure event with empty `Gate` field must bucket under `<none>` — not panic
   or be silently dropped.
6. Non-gate-failure event types (`block`, `step-advanced`, `verify-rejection`) must be
   excluded from gate health counts.
7. Roadmap present but zero schedulable features must show `"0/0 done"` — not an error
   or crash.
8. `Compute` must be called twice with the same `Inputs` and return byte-identical
   `Dashboard` values (determinism — no map iteration in output order).

## Out-of-Scope

- Live/watch mode or auto-refresh (single-shot snapshot only).
- Cross-repo or multi-tenant fleet aggregation (Magallanes scope).
- A persisted owner/assignee field on workflow state (real owner model is a separate
  feature — ownership here is git-derived and advisory).
- Running gates live (`gates.RunAll`) for the board — gate health derives from the
  telemetry event log only.
- Historical time-series or trends beyond `centinela insights`.

No new out-of-scope discoveries were found requiring a roadmap defer.

## Handoff

→ **senior-engineer** (code step).

Build in two slices as described in `docs/plans/team-dashboard.md`:

**Slice 1** — `internal/teamdashboard/dashboard.go` (types), `compute.go`,
`features.go`, `burndown.go`; `cmd/centinela/dashboard.go` (owners stubbed
`"unknown"`); `internal/ui/render_dashboard.go` panels 1 & 2 + empty states;
full G2 mapping (PROJECT.md paragraph + `centinela.toml` `paths`) so the
import_graph gate is green from the start. `--json` emits partial `Dashboard`
(Gates empty).

**Slice 2** — `internal/teamdashboard/gatehealth.go` + `insights.Gates` wiring
+ panel 3; replace owner stub with real `gitOwner` seam in `cmd/`. No G2 change
(insights edge already covered by aggregator allow). Adds gate-health and real
owners to `--json`.

Key constraints for the engineer:

- `internal/teamdashboard` must not import `cmd/`, `internal/ui`, `os/exec`, or
  any I/O package. Git stays entirely in `cmd/`.
- All source files and test files ≤ 100 lines (G1). Split renderer into
  `render_dashboard_panels.go` if it exceeds 100 lines.
- Test files colocated in `internal/teamdashboard/` move the 95% per-package
  coverage gate — `tests/` tier files do not count for per-package.
- Acceptance test must drive the real binary from a temp repo with local git only
  (no network push — avoids the acceptance-hang failure mode).
- `gitOwner` seam (`var gitOwner = func(...)`) must be overridable in integration
  tests so no real git calls occur during CI.
