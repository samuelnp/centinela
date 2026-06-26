# team-dashboard

## Problem

Centinela tracks each feature's lifecycle in its own `.workflow/<feature>.json`
and (under `use_worktrees`) its own `.worktrees/<feature>` checkout. Today the
only way to see the *whole* picture — which features are in flight, what step
each is stuck on, how old they are, who last touched them, and how the roadmap
is burning down — is to poll each worktree by hand (`centinela status <f>` per
feature) and eyeball `ROADMAP.md`. `governance-telemetry` and `centinela
insights` already aggregate the *event log*, but there is no single read-only
view of *current* multi-feature, multi-contributor state.

## Who / Why

**Who.** A developer or lead (and the orchestrating agent) running several
Centinela features at once — across multiple worktrees and, on a shared repo,
multiple contributors — who needs an at-a-glance status board.

**Why.** Multi-feature/multi-contributor state is invisible without manual
polling, so stalled work, ownership ambiguity, and roadmap drift go unnoticed.
One aggregated board turns "poll every worktree" into "read one screen."

## In Scope

- A new read-only aggregator `internal/teamdashboard` (aggregator layer,
  mirroring `internal/insights`): pure `Compute(...) Dashboard` over inputs the
  caller reads from disk — active workflows, worktrees, roadmap, telemetry.
- A `centinela dashboard` command (thin orchestrator) that reads the sources,
  calls `Compute`, and renders via `internal/ui`. Supports `--json` for the
  serializable `Dashboard` struct.
- **Three panels:**
  1. **In-flight features** — one row per active workflow: feature, current
     step + `X/5` progress, age (from `StartedAt`), profile/archetype, worktree
     path, and a best-effort **owner** (latest committer on the feature branch).
  2. **Roadmap burn-down** — per-schedulable-phase done/in-progress/planned
     counts and an overall `N/M done` line (from `roadmap.Summary()` +
     `FeatureStatus`).
  3. **Gate health** — aggregate gate-failure tallies by gate name, derived
     from `telemetry` `gate-failure` events (reusing the `insights` ranking
     approach).
- Graceful empty/missing handling: no active workflows, no telemetry, or no
  roadmap each render an honest empty-state, never an error.

## Out of Scope

- Live/watch mode or auto-refresh — single-shot snapshot only.
- Cross-repo / multi-tenant fleet aggregation — that is Magallanes' control
  plane, not Centinela.
- A new persisted "owner/assignee" field on workflow state — ownership is
  derived from git, not stored (a real owner model is a separate feature).
- Running gates live for the board (`gates.RunAll`) — gate health comes from the
  persisted telemetry event log, not a fresh per-feature gate run.
- Historical time-series / trends beyond what `centinela insights` already
  shows.

## Acceptance Summary

- `centinela dashboard` prints three panels (in-flight features, roadmap
  burn-down, gate health) from current on-disk state, touching no files.
- Each active workflow row shows feature, current step with `X/5` progress, age,
  and a best-effort git-derived owner; a feature with no commits yet shows an
  honest "unknown" owner rather than failing.
- Roadmap burn-down matches `roadmap.Summary()` (schedulable phases only;
  Backlog/Baseline excluded) and the overall done/total count.
- Gate health aggregates `gate-failure` telemetry by gate name; with telemetry
  absent or empty, the panel shows an empty-state, not an error.
- `--json` emits the full `Dashboard` struct; the render path is pure
  presentation in `internal/ui` and `internal/teamdashboard` performs no I/O
  (G7 / aggregator purity).
