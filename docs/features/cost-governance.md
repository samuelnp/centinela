# Feature: cost-governance

## Problem

A runaway agent can burn a large token/compute budget on a single step with
**no visibility** until the bill (or the wall-clock) lands. Centinela governs
*correctness* (gates, evidence) but says nothing about *cost*. Today
`internal/telemetry` records the driver `Model` per event but **no token
counts**, so there is no per-feature/per-step notion of spend or budget.

## Goal

Surface per-feature and per-step token spend against configurable budgets, with
a **soft gate**: when a step exceeds its budget, Centinela shows a ⚠ warning in
the status line and the `centinela validate` report and via a new `centinela
cost` report — but it **never blocks** `centinela complete`. Pure visibility,
matching the "no visibility into a runaway step" problem it fixes.

## Spend source (decided)

**Parse the host-harness transcript.** Claude Code passes `transcript_path` to
hooks; the transcript is JSONL where assistant messages carry
`usage.{input_tokens,output_tokens,...}`. A hook reads the transcript, sums the
token usage, attributes it to the **active feature/step** (from workflow state),
and records it as a new telemetry event. For local models the same machinery
carries a compute/wall-clock unit instead of spend — the runaway-step problem is
identical (per the roadmap definition).

## Scope (v1)

1. `[cost]` config block: `enabled`, `step_token_budget`, `feature_token_budget`,
   optional `[cost.tier_budgets]` per model-tier (reasoning/balanced/...).
2. `internal/cost` (leaf over telemetry+config): aggregate recorded cost samples
   per feature/step; compare to budget → over/remaining; report struct.
3. Telemetry: a `cost-sample` event type carrying input/output token counts,
   attributed to feature/step/model (extends the existing flat `Event`).
4. Capture hook: reads `transcript_path`, computes the token delta since the
   last sample (cursor per feature), emits the cost-sample event.
5. `centinela cost [feature]`: render spend vs budget per feature/step,
   highlighting over-budget rows.
6. Soft-gate surfacing: `centinela validate` appends a non-failing ⚠ when the
   active feature/step is over budget.

## Out of scope (v1)

- Hard/blocking enforcement (deliberately a soft gate; a `cost.enforce` hard cap
  is a possible follow-up).
- Dollar pricing tables (we report tokens/units, not currency).
- Non-Claude harness transcript formats (degrade gracefully: no `transcript_path`
  or unknown format → no samples, feature still works, just no spend data).

## Risks

- **Harness coupling.** Transcript schema is Claude-Code-specific. Mitigate by
  isolating parsing in one file with a tolerant reader; absence/parse-failure is
  a no-op, never an error (telemetry is already non-blocking by design).
- **Attribution drift.** A cursor per feature avoids double-counting across
  repeated hook fires; a missing cursor counts from 0 once.
