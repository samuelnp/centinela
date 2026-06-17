# failure-ledger-plan-advisor — big-thinker

## Problem

The telemetry ledger records every gate failure, but that signal is passive —
the plan advisor (which runs automatically during the plan step) never reads it,
so each feature starts planning blind to the failure modes that recently bit
this repo.

## Scope

Add a read-only context source to `internal/planadvisor` that surfaces the
top-N recurring gate failures from `.workflow/telemetry/events.jsonl` as (a) a
"Recurring gate failures" line in the context summary and (b) one pre-warning
question. Reuse the existing insights gate-counter; no new persisted schema.

## Dependencies & Assumptions

- Builds on completed `governance-telemetry` (ledger) and `centinela-insights`
  (gate aggregation in `internal/insights/gates.go`).
- Assumes the telemetry read stays lenient (missing → `(nil,nil)`) and that
  `[telemetry] enabled` gates the read.

## Risks

- **Import-graph (G2):** planadvisor → insights/telemetry. Verified SAFE —
  planadvisor is unmapped, insights is `aggregator`, telemetry an unmapped leaf;
  unmapped-source edges are the existing non-failing warning (centinela.toml
  lines 73–78). No cycle (insights does not import planadvisor).
- **Determinism:** inherited from `rankTop` (count desc, key asc) + `<none>`
  bucketing — not re-implemented.
- **Noise:** over-eager questions could break the "clean repo = quiet"
  guarantee; mitigated by a recurrence floor (`>= 2`) and `plan_question_limit`.

## Rollout

Pure additive context. Empty/missing/disabled ledger → byte-identical to current
advisor output, so existing repos see no change until failures actually recur.

## Handoff

→ feature-specialist. Plan authored at
`docs/plans/failure-ledger-plan-advisor.md` with per-step file/signature budgets.
Reuse decision: export `insights.Gates(events, topN) []Count` over the existing
unexported `gates`.
