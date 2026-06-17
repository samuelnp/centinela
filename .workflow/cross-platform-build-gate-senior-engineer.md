### Senior-Engineer Report: cross-platform-build-gate

**Date:** 2026-05-29

> Implementation was begun by a delegated senior-engineer subagent (config layer + parallel runner) and completed by the orchestrator after the subagent hit a session limit mid-step (missing `buildEnv`, `checkBuild`, gate wiring, toml). Finished per the discipline of not abandoning a mid-step.

#### Files Touched
| Path | Reason | Lines |
|------|--------|-------|
| internal/config/build_gate.go | `BuildTarget{GOOS,GOARCH}` + `BuildGateConfig{Enabled,Command,Targets}` + `NormalizeBuildGate` (default command, drop blank targets) | 41 |
| internal/config/config.go (edit) | `Build BuildGateConfig` field on `GatesConfig`; `NormalizeBuildGate` call in `applyDefaults` | +2 |
| internal/gates/build_runner.go | `buildTarget` (argv-parse, no shell), `firstStderrLine`, bounded-parallel `runTargets` worker pool | 84 |
| internal/gates/build.go | `buildEnv` (GOOS/GOARCH/CGO_ENABLED=0) + `checkBuild` aggregating failures into a sorted `Result.Details` | 47 |
| internal/gates/gates.go (edit) | wire `checkBuild` into `RunWithFilter` behind `cfg.Gates.Build.Enabled` | +3 |
| centinela.toml (edit) | enable `build = true` + `[gates.build]` with all 6 release targets | +12 |

#### Architecture Compliance
- G1: every source file ≤100 lines (max 84). 
- G2: `internal/gates` imports only `internal/config` (+ stdlib `os`, `os/exec`, `runtime`, `sort`, `sync`, `io`, `fmt`, `strings`); `internal/config` imports nothing internal. 
- G7: `cmd/` untouched — the gate appears through the existing `gates.RunWithFilter` path; no business logic added to the outer layer. Existing `RenderGateResult` renders the new `Result` with no `internal/ui` change.

#### Type-Safety Notes
- `BuildTarget` is a typed struct; targets are `[]BuildTarget`, never free-form maps. No `any`/`interface{}`.
- **Security (feature-specialist flag):** the configurable `command` is parsed with `strings.Fields` and exec'd via `exec.Command(argv[0], argv[1:]...)` — never `sh -c` — so a malicious/injected command string cannot reach a shell.

#### Trade-Offs
- Failures collected in parallel are non-deterministic in order, so `checkBuild` sorts `Details` for stable output/tests.
- `command` is injectable (config-driven), which keeps the gate language-neutral AND lets qa-senior unit-test `buildTarget`/`runTargets` with a harmless or synthetic-failing command instead of real cross-compiles.

#### Verification
- Host `go build ./...` succeeds; `go vet ./internal/gates/ ./internal/config/` clean.
- The gate itself was NOT executed (it would cross-compile centinela and fail on the pre-#10 `windows` targets — expected; see sequencing). `centinela validate`/`complete` were NOT run.

#### Handoff
- Next role: qa-senior
- Outstanding TODOs:
  - Unit tests: `NormalizeBuildGate` (default command, blank-target drop), `buildTarget` (argv-parse, synthetic failure naming target, empty command), `runTargets` (multi-failure aggregation), `checkBuild` (Pass all-ok / Fail aggregates+sorted / Skip empty targets).
  - **Anti-drift parity test** `internal/gates/build_matrix_parity_test.go`: parse `centinela.toml` `[gates.build].targets` vs `.github/workflows/release.yml` `strategy.matrix` cross-product; assert set-equality.
  - Acceptance (Gherkin): broken target → validate fails naming GOOS/GOARCH; all targets build → validate passes (use an explicit script for the broken-target simulation, not a shell-conditional, per the argv-parse design).
  - **Sequencing:** rebase onto post-#10 `main` BEFORE the validate step; do not weaken the gate to go green.
