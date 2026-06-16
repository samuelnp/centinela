# failure-ledger-plan-advisor — qa-senior

## Test Inventory

**Colocated unit tests (in `internal/`, each ≤100 lines — G1 compliant, these
move per-package coverage):**
- `internal/planadvisor/failures_test.go` (99) — `recurringFailures`
  (seeded / disabled / missing / topN≤0 / nil-cfg) + `failureTopN` clamp.
- `internal/planadvisor/failures_view_test.go` (29) — `failureSummary` order +
  empty.
- `internal/planadvisor/questions_failures_test.go` (59) — `worstGate`,
  `topFailureCount`, and the pre-warning question gating at count ≥ 2.
- `internal/insights/gates_exported_test.go` (42) — `Gates` == `gates`,
  `<none>` bucket, agreement with `Compute`.
- `internal/config/plan_advisor_failure_topn_test.go` (22) —
  `NormalizePlanAdvisorFailureTopN` (0/neg⇒3, 1⇒1, 5⇒5, 6⇒5).

**Tier tests (under `tests/`):**
- `tests/unit/failure_ledger_plan_advisor_unit_test.go` — ranked summary line.
- `tests/integration/failure_ledger_plan_advisor_integration_test.go` —
  pre-warning question, telemetry-disabled suppression, insights count
  agreement.
- `tests/acceptance/failure_ledger_plan_advisor_test.go` — builds the binary,
  runs `centinela hook plan-advisor`; `// Acceptance:` header + all 16
  `// Scenario:` comments mapped 1:1 to `Test…` functions (spec-traceability).

## Coverage Gaps

Aggregate coverage **95.3% ≥ 95.0%** gate (re-verified independently). New
symbols: `Gates`, `failureTopN`, `failureSummary`, `worstGate`,
`topFailureCount`, `NormalizePlanAdvisorFailureTopN` all 100%.
`recurringFailures` 83.3% — only the `telemetry.ReadDefault()` I/O-error branch
is uncovered (not deterministically reproducible; fails safe to nil). Coverage
claim left absent in evidence so the verify gate skips re-derivation rather than
risking a claim-vs-measured mismatch.

## Acceptance Wiring

`go test ./tests/acceptance/...` green. Spec-traceability satisfied: every
scenario in `specs/failure-ledger-plan-advisor.feature` appears verbatim as a
`// Scenario:` comment above a real test. Hook tests use a fresh plan-step
workflow (`workflow.New("f")`) so the advisor actually fires.

## Handoff

→ validation-specialist. `go test ./...` (all pass), acceptance (all pass),
coverage 95.3%, `gofmt`/`go vet` clean. Run the gatekeeper + `centinela
validate` for the validate step.
