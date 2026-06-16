# custom-gate-sdk — qa-senior

## Test Inventory

**Colocated (in-package, each ≤100 lines — move coverage):**
- `internal/config/custom_gate_test.go` (98) — `NormalizeCustomGates` defaults/
  trim; `validateCustomGates` valid + 6 indexed-error cases (empty name/command,
  duplicate, built-in `import_graph` collision, bad severity, bad output);
  defaults/validateConfig wiring.
- `internal/gates/custom_command_test.go` (97) — `customResult` Pass/Fail/Warn/
  timeout/empty-generic; `lines` split + cap overflow; `blob` 4 KiB cap.
- `internal/gates/custom_command_exec_test.go` (78) — `runCustom` true/false/
  exit-3, `sleep`+50ms→timedOut, `CENTINELA_CHANGED_FILES` set/unset.
- `internal/gates/custom_command_run_test.go` (83) — `customGates` over a config
  (true/false → 2 Results), disabled skipped, empty→empty, diff-aware path.
- `internal/audit/participation_custom_test.go` (49) — **regression guard** for
  the bug fix: custom-gate Names participate; `target_gates` still honored.

**Tier:** `tests/unit/custom_gate_sdk_unit_test.go` (RunWithFilter),
`tests/integration/custom_gate_sdk_integration_test.go` (AC-4 cross-feature:
per-line baseline → tolerate → new blocks), `tests/acceptance/custom_gate_sdk_test.go`
(`// Acceptance:` header + all 19 `// Scenario:` titles verbatim, built-binary
harness).

## Coverage Gaps

Aggregate **95.1% ≥ 95.0%** (re-verified independently). New symbols at/near
100% (`NormalizeCustomGates`, `validateCustomGates`, `customResult`,
`customDetails`, `blob/lineDetails`, `participatingGates`); `runCustom` ~95% (a
launch-error branch). Windows shell paths skipped per `runtime.GOOS`; unix path
fully exercised. Coverage claim left absent in evidence (verify gate skips
re-derivation).

## Acceptance Wiring

`go test ./tests/acceptance/...` green. Spec-traceability satisfied — all 19
scenarios in `specs/custom-gate-sdk.feature` appear verbatim as `// Scenario:`
comments over real tests. Harness builds the binary and runs `centinela
validate`/`audit` in temp repos with `[[gates.custom]]`.

## Handoff

→ validation-specialist. `go test ./...` (all pass), acceptance (pass), coverage
95.1%, gofmt/vet clean. No implementation file modified; no gate lowered.
