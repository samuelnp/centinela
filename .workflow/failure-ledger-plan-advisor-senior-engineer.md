# failure-ledger-plan-advisor — senior-engineer

## Files Touched

| File | Lines | Change |
|------|------:|--------|
| `internal/insights/gates.go` | 20 | Exported `Gates(events, topN)` pass-through over unexported `gates`; `Compute` unchanged |
| `internal/planadvisor/failures.go` | 29 | NEW — `recurringFailures(cfg, topN)` (telemetry-gated, read-only) + `failureTopN(cfg)` |
| `internal/planadvisor/failures_view.go` | 18 | NEW — `failureSummary` renders `gate (×N)` joined |
| `internal/planadvisor/context.go` | 31 | `bundle.Failures []insights.Count` + populated in `buildBundle` |
| `internal/planadvisor/context_summary.go` | 89 | Appends "recurring gate failures" line when non-empty |
| `internal/planadvisor/questions.go` | 65 | One pre-warning question + `worstGate`/`topFailureCount` helpers |
| `internal/config/workflow_config.go` | 30 | `PlanAdvisorFailureTopN` toml knob |
| `internal/config/plan_advisor.go` | 43 | `Default=3`, `Max=5`, `NormalizePlanAdvisorFailureTopN` |
| `internal/config/defaults.go` | 29 | Normalize the knob in `applyDefaults` |

## Architecture Compliance

- **G1 (file size):** every touched file ≤89 lines (≤100 budget).
- **G2 (import-graph):** new edges `planadvisor → insights`, `planadvisor →
  telemetry`. planadvisor is unmapped; insights is `aggregator`; telemetry an
  unmapped leaf. Per the documented matrix policy these are non-failing
  warnings (same kind insights/calibration already rely on). No cycle —
  insights does not import planadvisor. `go build ./...` clean.
- **Reuse (AC-5):** counts come from the single `insights.Gates` counter; no
  duplicated counting, so the advisor and `centinela insights` can never
  disagree.
- **Read-only / quiet-by-default (AC-3/4):** disabled telemetry or a
  missing/empty ledger → `nil` failures → no summary line, no question →
  byte-identical advisor output.

## Type-Safety Notes

Strict Go. No `interface{}`/`any`. `recurringFailures` returns the typed
`[]insights.Count`; nil-guards on `cfg` and empty `Failures` keep the eagerly
built question text panic-free. `go vet ./...` clean; `gofmt` clean.

## Trade-Offs

- Split `failureSummary` into its own `failures_view.go` (plan-authorized) to
  keep `context_summary.go` under the 100-line budget.
- Recurrence floor for the pre-warning question is a hardcoded `>= 2` (not a
  separate config knob) to keep the config surface minimal; the list size is
  the one new knob (`plan_advisor_failure_top_n`, default 3, cap 5).

## Handoff

→ qa-senior. Not user-facing (CLI text advisor, no UI surface) so no
ux-ui-specialist. Tests step must add: unit tests for `recurringFailures`
(telemetry-gated, missing/empty ledger, topN clamp), `failureSummary` format,
and the question helpers; integration over `Directive` (summary line +
threshold question + byte-identical quiet path); acceptance tests mapping the 14
`specs/failure-ledger-plan-advisor.feature` scenarios.
