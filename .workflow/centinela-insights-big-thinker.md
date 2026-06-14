# Big-Thinker Report — centinela-insights

## Decision
`centinela insights`: read-only analytics over the governance-telemetry JSONL log. Pure `internal/insights` aggregator: `Compute([]telemetry.Event, topN) Report` (imports only internal/telemetry leaf + stdlib). Thin `cmd/centinela/insights.go` (--top default 5, --json à la verdict.go); `internal/ui/render_insights.go` (house style).

## The 4 metrics (precise, computable)
1. most-triggered blocks — count `block` events by reason/fileType, ranked.
2. most-failed gates — count `gate-failure` by Gate, ranked.
3. features with most rework — per-feature #{gate-failure+verify-rejection+complete-rejected}; empty-Feature excluded.
4. mean steps-to-green — (#{complete-rejected}+#{step-advanced})/#{step-advanced}; n/a when 0 advances (guarded).

## Determinism & edges
Sort count desc then key asc; never range a map. Empty/missing log -> clean empty-state, exit 0.

## v1 scope
In: 4 metrics, Compute, command (--top/--json), renderer, empty-log, import-graph mapping. Out: --since/time-window, per-model breakdown (capability-calibration owns), trends.

## Layer
internal/insights joins the existing `aggregator` import_graph layer (allow domain+leaf; telemetry is leaf → fully mapped, no warning) + one-line PROJECT.md G2 prose.

Handoff → feature-specialist.
