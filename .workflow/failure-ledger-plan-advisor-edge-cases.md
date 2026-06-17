# Edge Cases: failure-ledger-plan-advisor

## Covered

- **Missing ledger file** → `recurringFailures` returns empty; no summary line,
  no question; advisor output byte-identical to pre-feature (acceptance +
  colocated `failures_test.go`).
- **Empty ledger file** (exists, zero events) → no recurring-failure output.
- **Only block / step-advanced events** (no `gate-failure`) → no output;
  `insights.Gates` counts only `TypeGateFailure`.
- **Telemetry disabled** (`cfg.Telemetry.Enabled = &false`) → ledger never read;
  nil failures (colocated + integration + acceptance).
- **Ranking** — gate failures ordered count desc, then gate name asc
  (`gates_exported_test.go` proves agreement with `insights.Compute`).
- **Ties in count** broken alphabetically for reproducible output.
- **Empty `Gate` field** → bucketed under `<none>` without crashing.
- **Top-N cap** — only N gates listed when more distinct gates failed; N clamped
  to [1,5] (default 3) by `NormalizePlanAdvisorFailureTopN`.
- **Threshold gating** — pre-warning question only when worst count `>= 2`;
  below threshold (count 1) stays quiet.
- **Question cap** — pre-warning flows through `selectQuestions`' `limit` loop so
  `plan_question_limit` bounds total questions.
- **Plan-step-only / headless** — advisor hook acts only in the plan step and
  exits silently when headless (unchanged).
- **Read-only** — advisor never writes the ledger (acceptance asserts ledger
  bytes unchanged); two runs on the same ledger are byte-identical.
- **Counts agree with `centinela insights`** on the same ledger (shared counter).

## Residual Risks

- `recurringFailures` covers all branches except the `telemetry.ReadDefault()`
  I/O-error path (per-function 83.3%), which is not deterministically
  reproducible in a unit test; on any read error the function fails safe
  (returns nil → quiet advisor). Aggregate coverage gate (95.3% ≥ 95%) is met.
- No per-feature attribution of failures (repo-wide aggregation only) — by
  design per the brief; revisit if planners want feature-scoped signal.
