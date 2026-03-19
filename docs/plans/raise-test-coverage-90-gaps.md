# Coverage Gap Report (toward >90%)

Baseline command:

```bash
go test ./... -coverpkg=./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Current total is far below target. Top missing statement groups:

1. `cmd/centinela` (~1064 missed statements)
2. `internal/gates` (~480)
3. `internal/workflow` (~473)
4. `internal/ui` (~378)
5. `internal/setup` (~251)
6. `internal/roadmap` (~104)
7. `internal/scaffold` (~76)
8. `internal/config` (~52)

## Missing tests to add first (highest impact)

- `tests/unit/config_load_test.go`
  - `Load()` default path when `centinela.toml` is missing.
  - Parse errors and default-application behavior.

- `tests/unit/gates_runall_test.go`
  - `RunAll()` combinations for enabled/disabled gates.
  - `AllPassed()` pass/fail behavior.

- `tests/unit/gates_filesize_test.go`
  - source file detection, ignore dir behavior, line counting.
  - failure formatting and `itoa` coverage.

- `tests/unit/gates_i18n_test.go`
  - JSON key parity pass/fail.
  - gettext missing/translated entry behavior.

- `tests/unit/workflow_state_steps_test.go`
  - `New()`, `Save()`, `Load()`, `StepNumber()`, `Complete()` transitions.

- `tests/unit/workflow_validate_tests_test.go`
  - `validateTests()` happy path and missing artifacts cases.

- `tests/unit/roadmap_test.go`
  - `Load()`, `Save()`, `FeatureStatus()`, `Summary()`.

- `tests/unit/scaffold_extract_test.go`
  - `Extract()` creates files and preserves existing files.

- `tests/unit/setup_hooks_merge_test.go`
  - `InjectHooks()` merge behavior and idempotency.

- `tests/unit/ui_render_test.go`
  - `RenderBlocked`, `RenderContext`, `RenderGateResult`, `RenderRoadmap*`.

## Then add command-layer integration tests

- `tests/integration/cmd_init_validate_status_test.go`
  - Run `go run ./cmd/centinela ...` in temp dirs.
  - Cover `init`, `start`, `status`, `validate` command paths.

- `tests/integration/cmd_hook_context_postwrite_test.go`
  - Cover non-blocking hook command paths with fixture workflows.

## Notes

- Achieving >90% requires heavy coverage in `cmd/centinela` and `internal/gates`.
- Most current tests run through `tests/*` packages; add direct package-level tests to increase statement hit rate.
