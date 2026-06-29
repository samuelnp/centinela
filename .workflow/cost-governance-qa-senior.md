# cost-governance — qa-senior

## Test Inventory

**Colocated (coverage):**
- `internal/cost/`: `transcript_test.go` (sum, skip garbage, missing file, delta,
  truncation), `cursor_test.go` (round-trip, path-scoped offset, malformed),
  `report_test.go` (Fold, budget math, Build/ActiveStatus/Empty), `coverage_test.go`
  (feature-branch + step-absent + sort).
- `internal/config/cost_test.go` (Normalize clamps, IsActive matrix).
- `internal/ui/render_cost_test.go` (empty, rows, OVER marker, warning line).
- `internal/telemetry/cost_sample_test.go` (records + zero no-op).
- `cmd/centinela/`: `cost_test.go` (active resolver, report, warning), `cost_more_test.go`
  (--json, not-over, nil/none), `hook_cost_test.go` (capture+attribute, no-transcript,
  disabled).

**Tier (workflow):** `tests/unit` (soft-gate math), `tests/integration` (telemetry
round-trip → aggregate), `tests/acceptance` (binary-driven capture → cost OVER →
validate exit 0 ⚠; missing-transcript no-op). All test files ≤100 lines.

## Coverage Gaps

Total **95.1% ≥ 95.0%** gate. Uncovered: defensive I/O error branches
(`SaveCursor` MkdirAll/Marshal, `SumFrom` open/seek failures) — non-deterministic
to trigger; all return safe zero/no-op values. No logic gaps.

## Acceptance Wiring

`specs/cost-governance.feature` scenarios map to `TestAccCostCaptureReportAndSoftGate`
(capture + report + soft gate) and `TestAccCostMissingTranscriptNoOp` (graceful
degradation). `centinela.toml` already runs `go test ./tests/acceptance/...`.

Also repaired stale `lean-evidence-footprint` gitignore tests (unit + integration)
that asserted the pre-f138f90 broad `*.json` rule — they now assert the role-suffix
design (root state + roadmap kept, role evidence + locks ignored).

## Handoff

→ validation-specialist: full suite + gates green; produce the gatekeeper report.
Expect no coverage delta and no file-size violations (all source + tests ≤100).
