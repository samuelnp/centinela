# governance-telemetry — qa-senior

## Test Inventory
- **Colocated unit tests** (`internal/telemetry/record_test.go`): `Record` enabled/disabled/nil-cfg, schema+timestamp stamping, append, and the swallowed-I/O-error path.
- **Acceptance tests** (`tests/acceptance/governance_telemetry_*_test.go`, 5 files): all **17** `.feature` scenarios mapped 1:1 via exact `// Scenario:` markers — block (out-of-step / need-init), gate-failure (single + per-gate), verify-rejection, complete-rejected (gates/verify), step-advanced, disabled no-op, default-on, schema+timestamp, append-only ordering, two-sequential-intact, I/O-error-non-fatal, lenient Read, missing-log, and rework derivation.

## Coverage
- `./scripts/check-coverage.sh` → **coverage gate passed: 95.2% >= 95.0%**. The acceptance suite exercises every `internal/telemetry` constructor + `Read`/`ReadDefault` + `config.IsEnabled` (both arms).

## Scenario → test mapping
17/17 scenarios covered; the `comm -23` spec-vs-tests title diff is empty.

## Fix applied
`TestRecord_IOErrorIsSwallowed` asserted `os.IsNotExist` after the I/O-error path, but a file at `.workflow` yields `ENOTDIR` on stat — corrected to assert the events file simply isn't readable (any stat error), confirming the write was swallowed.

## Notes
- All test files ≤100 lines (G1); `governance_telemetry_read_test.go` was split (rework moved to its own file) to stay under budget.
- `go test ./...` green; `gofmt` clean.

Handoff: validation-specialist.
