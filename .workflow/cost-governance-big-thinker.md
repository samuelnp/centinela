# cost-governance — big-thinker

## Problem

Centinela governs correctness but is blind to cost: a runaway agent can burn a
large token/compute budget on one step with no visibility. `internal/telemetry`
records the driver `Model` but no token counts, so there is no spend or budget
concept today.

## Scope

Per-feature/per-step token budgets surfaced against a **soft** (warn-only) gate,
fed by parsing the host-harness transcript. New `[cost]` config,
`internal/cost` analytics package, a `cost-sample` telemetry event, a capture
hook, a `centinela cost` report, and a non-failing validate ⚠. v1 reports
tokens/units (no currency) and never blocks.

## Dependencies & Assumptions

- Builds on existing `internal/telemetry` (append-only JSONL, non-blocking) and
  the analytics-layer pattern (`internal/calibration`/insights read telemetry).
- Assumes the host harness passes `transcript_path` to hooks and the transcript
  is JSONL with `usage.{input,output}_tokens`. For other harnesses / local
  models the same path carries a compute unit; absence degrades to no-op.
- team-dashboard (Phase 10 predecessor) is done; no blocking deps.

## Risks

- Harness coupling of the transcript schema → isolate in one tolerant reader;
  parse failure is a no-op, never an error.
- Double-counting across repeated hook fires → per-feature byte cursor.

## Rollout

Single branch, additive. Zero config = silent no-op; old telemetry lines read as
0 tokens. No migration.

## Handoff

→ feature-specialist: encode capture/no-double-count/soft-gate/degradation as
acceptance scenarios; confirm the report shows used/budget/remaining per row.
