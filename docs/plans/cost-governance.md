# Plan: cost-governance

## Summary

Add per-feature/per-step token budgets with a **soft** (warn-only) gate, fed by
parsing the host-harness transcript. New `[cost]` config, an `internal/cost`
analytics package over telemetry+config, a `cost-sample` telemetry event, a
capture hook, a `centinela cost` report, and a non-failing validate warning.

## Architecture / layer placement (G2 import_graph)

- `internal/telemetry` (config-only leaf): add `TypeCostSample` + token fields on
  the flat `Event`. Stays a leaf — no new imports.
- `internal/cost` (NEW): joins the **analytics layer** alongside
  `internal/calibration` / insights — read-only over `internal/telemetry` +
  `internal/config` leaves, imported only by `cmd/`. Add it to the existing
  analytics layer `paths` in `centinela.toml [gates.import_graph]`.
- `cmd/centinela`: `cost.go` (command) + `hook_cost.go` (capture). Outer layer.

## Components (each file ≤100 lines)

| File | Responsibility |
|------|----------------|
| `internal/config/cost.go` | `[cost]` schema: `Enabled`, `StepTokenBudget`, `FeatureTokenBudget`, `TierBudgets map[string]int`; Normalize/defaults (0 = off) |
| `internal/telemetry/event.go` | add `TypeCostSample` + `InputTokens`/`OutputTokens` (omitempty) |
| `internal/telemetry/constructors.go` | `CostSample(feature, step, model, in, out)` constructor |
| `internal/cost/transcript.go` | tolerant JSONL reader: sum `usage.{input,output}_tokens`; cursor (byte offset) per feature to count deltas |
| `internal/cost/aggregate.go` | fold cost-sample events → per feature/step totals |
| `internal/cost/budget.go` | compare totals vs config budgets → `Status{Used,Budget,Over,Remaining}` |
| `internal/cost/report.go` | build the report model (rows per feature/step) |
| `internal/ui/render_cost.go` | render report + the over-budget ⚠ line |
| `cmd/centinela/hook_cost.go` | read `transcript_path` from hook stdin JSON; emit a cost-sample for the active feature/step |
| `cmd/centinela/cost.go` | `centinela cost [feature]` command |
| `cmd/centinela/validate.go` | append non-failing ⚠ when active step over budget |

## Data flow

1. Hook fires (Stop / step complete) → `hook_cost.go` reads `transcript_path`,
   parses the delta since the feature's cursor, sums tokens, emits a
   `cost-sample` telemetry line attributed to the active feature/step/model.
2. `internal/cost` aggregates samples and compares to `[cost]` budgets.
3. `centinela cost` renders spend vs budget; `centinela validate` surfaces a ⚠
   for the active step when over budget. Neither blocks.

## Soft-gate contract

Over-budget is **always** non-failing: it appends a warning, exit code unchanged.
The only blocking remains the existing correctness gates. (`cost.enforce` hard
cap is explicitly a future feature — see brief Out-of-scope.)

## Test strategy

- **unit**: config Normalize/defaults; `budget.go` over/under/remaining math;
  `transcript.go` token summing + tolerant parse (malformed line skipped, no
  `transcript_path` → empty, cursor delta correctness).
- **integration**: telemetry round-trip — write cost-sample events, aggregate,
  assert per-step totals + budget status.
- **acceptance**: binary-driven — seed a fake transcript + `[cost]` config in a
  temp repo, run the capture path + `centinela cost`, assert the report shows
  over-budget; assert `centinela validate` stays exit 0 with a ⚠.

## Graceful degradation

No `transcript_path`, unreadable file, or unknown schema → zero samples, no
error (telemetry is non-blocking). `[cost] enabled=false` or all budgets 0 →
feature is a complete no-op (back-compat: zero config = silent).

## Rollout

Single feature branch. Additive config + new command + new event type
(back-compat: old telemetry lines lack token fields → read as 0). No migration.
