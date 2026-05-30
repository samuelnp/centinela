### Big-Thinker Report: cross-platform-build-gate

**Date:** 2026-05-29

#### Problem

Centinela's `validate` gate runs `go test ./...` on the host only (Linux/macOS),
so a release-breaking compile error invisible on the host (the `syscall.Flock`
Unix-only call in `lock.go`) passed plan → code → tests → validate → docs and
only surfaced in the release matrix `{linux,darwin,windows}×{amd64,arm64}`,
where both Windows targets fail to compile. For a framework whose value
proposition is "validate before it ships," a host-only build check is a
robustness hole. Part A (the lock fix) shipped as hotfix PR #10. **This feature
is Part B: a first-class gate that cross-compiles every release target during
`validate` and fails when any target breaks, so the gap cannot recur.**

#### Scope

- **In:** A first-class built-in **build gate** in `internal/gates/`, toggled via
  `[gates] build`, appearing in the Built-in Gates report. Generic + config-
  driven (command template + `{GOOS,GOARCH}` targets) so it is not hard-coded to
  Go and non-Go projects are unaffected (default OFF). Build-only (no per-target
  tests), `CGO_ENABLED=0`, parallel, build-cache-friendly. Reports which
  `GOOS/GOARCH` failed. An anti-drift parity test vs `release.yml`. Enable it in
  this repo's `centinela.toml` for all 6 targets.
- **Out:** Part A / any `lock.go` change (done in #10). Rewriting `release.yml`
  to consume a generated target file. A local "representative subset" default
  (full matrix runs at validate; any subset would be an explicit, documented
  opt-out, not v1). New persisted entities.

#### Dependencies & Assumptions

- Builds on existing gate infra: `gates.Result`/`RunWithFilter`/`AllPassed`,
  `ui.RenderGateResult`, `config.GatesConfig`/`applyDefaults` — all reused, not
  rebuilt.
- Assumes a Go toolchain on the validate host (true for this repo); the gate's
  command is config-supplied, so the assumption is per-project, not baked in.
- **Self-dependency:** this branch is off `main` at v0.5.0 and still contains the
  broken pre-#10 `lock.go`. The gate will (correctly) fail its own validate on
  the two windows targets until #10 merges and this branch rebases. Sequence the
  validate step AFTER the rebase.
- `release.yml` matrix is the reference target list (6 targets). Anti-drift is
  enforced by a Go test, not by sharing a file with the workflow.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Target list drifts from release.yml matrix | High | Medium | Parity test parses both `centinela.toml` targets and `release.yml` matrix cross-product and asserts equality; runs under existing `go test ./...` in validate |
| 6 cross-compiles slow every `validate` | Medium | High | Build-only (no tests), `CGO_ENABLED=0`, output discarded, reuse Go build cache, bounded-parallel worker pool; warm runs sub-second |
| Gate perceived as Go-specific; breaks non-Go consumers | Medium | Medium | Gate is generic (command template + targets), default OFF; only this repo's toml opts in; non-Go projects unaffected |
| Gate fails its own validate before #10 rebase | Medium | High (expected) | Documented sequencing: implement in code step, rebase onto post-#10 main BEFORE validate step; do not weaken gate to go green |
| `cmd/` business-logic creep (G7) | Low | Low | All logic in `internal/gates` + `internal/config`; `cmd/` already calls RunWithFilter — zero new cmd code |
| Source file >100 lines (G1) | Low | Medium | Split into build.go (orchestrate) + build_runner.go (exec/parallel) + config/build_gate.go; each ≤90 |
| Cross-compile env leaks into host build | Low | Low | Set GOOS/GOARCH/CGO_ENABLED only on the per-target exec.Cmd env, never the process env |

#### Rollout

- Step 1: Config plumbing — `BuildGateConfig` + `GatesConfig.Build` + defaults + unit tests (no behavior).
- Step 2: Gate logic single-target serial; wire into `RunWithFilter`; enable in `centinela.toml`. Gate now visible in report; fails windows (expected pre-#10).
- Step 3: Full 6-target matrix + bounded parallelism + `CGO_ENABLED=0` + discard output + build-cache reuse.
- Step 4: Anti-drift parity test (`build_matrix_parity_test.go`): toml targets == release.yml matrix.
- Step 5: Rebase onto post-#10 `main`; confirm `centinela validate` green; then proceed to docs.

#### Handoff

- Next role: feature-specialist
- Outstanding questions:
  1. Config shape for targets — keyed `{ goos, goarch }` table array (proposed) vs generic `env`-pair list for true language-neutrality? Proposal favors the typed `{goos,goarch}` form for ergonomics; confirm.
  2. Gate name/label string and exact Pass/Fail message wording for the report.
  3. Parallelism bound — GOMAXPROCS vs a fixed small cap (e.g. 4) to avoid memory spikes on small CI runners?
  4. Should a future explicit `local_subset` opt-out be specified now or deferred? (Plan defers; v1 is full-matrix.)
  5. Acceptance test mechanics: simulate a broken target via a throwaway unbuildable command vs a temp package with a bad build constraint?
