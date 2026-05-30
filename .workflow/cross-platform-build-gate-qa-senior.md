# cross-platform-build-gate — qa-senior

**Date:** 2026-05-29
**Step:** tests (3/5) · **Handoff:** validation-specialist

## Test Inventory by Tier

| Tier | File | Lines | Covers |
|------|------|------:|--------|
| Unit (config) | `internal/config/build_gate_test.go` | 58 | `NormalizeBuildGate`: default command on blank/whitespace; explicit command preserved; blank/whitespace targets dropped + trimmed; empty stays empty |
| Unit (gates) | `internal/gates/build_runner_test.go` | 89 | `buildTarget` success / failure-names-target / empty-command; `firstStderrLine` fallback; `runTargets` aggregate-failures / all-pass |
| Unit (gates) | `internal/gates/build_test.go` | 82 | `buildEnv` (GOOS/GOARCH/CGO_ENABLED=0); `checkBuild` Skip (empty), Pass (all ok), Fail (aggregated + sorted Details) |
| Unit (parity) | `internal/gates/build_matrix_parity_test.go` | 98 | `TestBuildMatrixParity`: centinela.toml `[gates.build].targets` set == release.yml `strategy.matrix` cross-product |
| Integration | `tests/integration/cross_platform_build_gate_integration_test.go` | 41 | `RunWithFilter` includes G-Build when enabled, excludes when disabled |
| Acceptance | `tests/acceptance/cross_platform_build_gate_test.go` | 81 | all 6 targets build -> Pass ("All 6 release targets compile."); one target fails -> Fail naming the broken GOOS/GOARCH only |

Max test file: **98 lines** (≤100, G1 satisfied). All tables are table-driven where it adds value; strict typing, no `any`.

## Test Design Notes

- **Synthetic commands, not real cross-compiles.** Unit/acceptance tests use
  `go version` for the happy path and small temp executable scripts for
  failures, keeping the suite fast and deterministic (no shelling out to six
  real `go build` invocations).
- **No shell injection in tests.** `buildTarget` argv-parses
  (`strings.Fields` + `exec.Command`, never `sh -c`). The broken-target
  acceptance scenario is therefore simulated with an **explicit executable
  script file** that inspects `$GOOS`/`$GOARCH` and `exit 1`s for one pair —
  the script *is* the program in `command`, not a shell-conditional embedded in
  the command string.
- **Parity parsing approach:** TOML side parsed with `github.com/BurntSushi/toml`
  (already a dep) into the typed `config.Config`. The `release.yml`
  `strategy.matrix` `{goos:[...], goarch:[...]}` lists are extracted **dep-free**
  with a targeted regex over the single `matrix:` line (no YAML parser added to
  go.mod). Cross-product expanded, compared as sets, drift named in both
  directions.

## Coverage

- `./scripts/check-coverage.sh` -> **coverage gate passed: 95.2% >= 95.0%**
  (`go test ./... -coverprofile`, NO `-coverpkg` — total project coverage).
- New code covered by **colocated** `_test.go` so it counts toward the total:
  `NormalizeBuildGate` 100%, `buildEnv` 100%, `checkBuild` 100%,
  `buildTarget` 100%, `firstStderrLine` 100%, `runTargets` 95.2%
  (only the GOMAXPROCS<1 guard branch is unhittable on a real host).

## Acceptance Wiring

The acceptance tests are Go tests (`package acceptance_test`, snake_case file)
under `tests/acceptance/`, run by `go test ./...`, which is already in
`centinela.toml` `[validate].commands`. They exercise the public
`gates.RunWithFilter` path (same entry point `centinela validate` uses), not
internal helpers, so they validate the gate end-to-end through its real wiring.

## Anti-Drift Verification

Manually confirmed `TestBuildMatrixParity` is not a no-op: removing
`windows/arm64` from centinela.toml made it FAIL with
`in release.yml not centinela.toml=[windows/arm64]`; restoring made it PASS.

## Quality Gates (all green)

- `go build ./...` — Success
- `go vet ./...` — No issues
- `go test ./...` — 971 passed across 23 packages
- `./scripts/check-coverage.sh` — 95.2% >= 95.0%

## Edge Cases

Full enumeration in `.workflow/cross-platform-build-gate-edge-cases.md`
(empty targets -> Skip; malformed target dropped; CGO off / no toolchain;
unknown garbage target reported not paniced; argv-parse / no shell injection;
stderr-empty fallback; build-cache warm reuse residual risk; target-list drift
caught by parity test).

## Handoff to validation-specialist

Tests + edge-cases artifact + evidence are complete and green. Outstanding for
validate step: gatekeeper report, full `centinela validate` run, production
readiness (if enabled). Note: `centinela evidence validate` currently reports 2
PRE-EXISTING issues in the big-thinker and feature-specialist evidence
(stale `docs/features/claim-verification.md` snapshot reference) — outside
qa-senior scope; flagging for the validation-specialist to resolve.

> NOTE: `centinela complete` was deliberately NOT run — the human advances steps.
