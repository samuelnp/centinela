# Edge Cases: cross-platform-build-gate

## Covered

- **Empty targets list -> Skip.** `[gates] build = true` with no/empty `targets`
  no-ops cleanly. `checkBuild` returns `Status: Skip` with "Build gate enabled
  but no targets configured." — never a false Pass and never a panic.
  Covered by `TestCheckBuild_SkipWhenNoTargets`
  (internal/gates/build_test.go).
- **Malformed target dropped at config load.** A target with blank `goos` or
  `goarch` (or whitespace-only) is dropped by `NormalizeBuildGate` so it cannot
  produce a silently-empty cross-compile invocation. Surrounding whitespace is
  trimmed and valid targets are preserved in order. Covered by
  `TestNormalizeBuildGate_DropsBlankTargets` /
  `TestNormalizeBuildGate_EmptyTargetsStayEmpty`
  (internal/config/build_gate_test.go).
- **CGO off / no C toolchain on host.** `buildEnv` always exports
  `CGO_ENABLED=0` alongside `GOOS`/`GOARCH`, so pure-Go cross-compiles succeed
  without a C compiler. Covered by `TestBuildEnv_SetsTargetAndCGO`
  (internal/gates/build_test.go).
- **Unknown / garbage target reported, not paniced.** A failing build (any
  cause, including an unknown GOOS/GOARCH or a command that exits non-zero) is
  folded into a `Fail` Result with a `goos/goarch: <first error line>` Details
  entry — no panic, no stack trace. Failure naming is covered by
  `TestBuildTarget_FailureNamesTarget` and aggregation by
  `TestCheckBuild_FailAggregatesSortedDetails` (Details sorted for stable
  output) and the acceptance `TestBuildGate_OneTargetFails_NamesGOOSGOARCH`.
- **Command argv-parse / no shell injection.** The configurable `command` is
  split with `strings.Fields` and exec'd directly via `exec.Command` — never
  `sh -c` — so a command string cannot reach a shell. An empty/whitespace-only
  command yields a descriptive error instead of executing anything. Covered by
  `TestBuildTarget_EmptyCommand`. Because there is no shell, the broken-target
  acceptance scenario is simulated with an explicit executable script file that
  inspects `$GOOS`/`$GOARCH` (the script is the program, not a shell-conditional
  inside the command field) — `TestBuildGate_OneTargetFails_NamesGOOSGOARCH`,
  `TestRunTargets_AggregatesFailures`.
- **stderr-empty failure (e.g. command not found).** `firstStderrLine` falls
  back to the exec error string when stderr is empty, so the Details entry is
  always non-empty. Covered by `TestFirstStderrLine_FallsBackToRunErr`.
- **Multi-target aggregation / all-pass.** `runTargets` runs targets through a
  bounded GOMAXPROCS worker pool and aggregates every failure; returns none when
  all pass. Covered by `TestRunTargets_AggregatesFailures` /
  `TestRunTargets_AllPass`.
- **Target-list drift caught at validate.** `TestBuildMatrixParity`
  (internal/gates/build_matrix_parity_test.go) parses `[gates.build].targets`
  from centinela.toml and the `strategy.matrix` `{goos, goarch}` lists from
  `.github/workflows/release.yml`, expands the cross-product, and asserts set
  equality — naming any target present in one but not the other. Manually
  verified it fails (and names `windows/arm64`) when a target is removed from
  centinela.toml, and passes once restored.

## Residual Risks

- **Build-cache warm reuse is a performance property, not asserted.** The
  feature spec's "second run completes in well under 1 second per target" is a
  perf characteristic of the real `go build` (no `-a`, output discarded). The
  unit/acceptance tests deliberately use synthetic harmless commands
  (`go version` / temp scripts) for speed and determinism, so they do not
  shell out to six real cross-compiles and do not time the warm path. Mitigation:
  the design (no `-a`, `CGO_ENABLED=0`, output to `io.Discard`) preserves the Go
  build cache; warm-run latency is validated in practice by `centinela validate`
  on this repo, not by the test suite.
- **Real cross-compile of all six targets is not run in the unit suite** (by
  design — kept fast/deterministic). End-to-end real-build behavior is exercised
  by `centinela validate` itself (the gate is enabled in centinela.toml) and by
  the release pipeline; the parity test guarantees the two target lists stay in
  sync.
- **Failure ordering from the parallel pool is non-deterministic;** mitigated by
  `checkBuild` sorting `Details` before returning, asserted by
  `TestCheckBuild_FailAggregatesSortedDetails`.
