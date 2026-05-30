# Plan — cross-platform-build-gate (Part B)

> Brief: [docs/features/cross-platform-build-gate.md](../features/cross-platform-build-gate.md)
> Scope (LOCKED): Part B only — a first-class cross-compile **build gate** that
> fails `centinela validate` when any configured release target fails to build,
> reporting the broken `GOOS/GOARCH`. Part A (the `lock.go` portability fix) is
> OUT — already shipped as hotfix PR #10.

## 1. Decision: first-class built-in gate, command-driven

**Recommendation: add a first-class built-in gate `G_BUILD` in `internal/gates/`,
whose *command* and *targets* are config-driven (not hard-coded to Go).**

Rationale:
- The user explicitly wants "a first-class gate, not a hidden CI step." A
  `validate.commands` shell script (like `check-coverage.sh`) is exactly the
  "hidden CI step" they reject: it does not appear in the **Built-in Gates**
  report block, is not toggled via `[gates]`, and gives no structured
  per-target detail. The gate path reuses the existing `Result{Name, Status,
  Message, Details}` contract and `RenderGateResult`, so broken targets render
  as `· windows/amd64` detail lines for free.
- The legitimate objection — "cross-compile is Go-specific, Centinela governs
  non-Go projects too" — is resolved by making the gate **generic**: it is a
  *build gate* configured by `[gates.build]` with a `command` template and a
  `targets` list of `{GOOS, GOARCH}` (or, generically, `{env_key=value}`
  pairs). For a Go project the template is `go build ./cmd/centinela` run once
  per target with `GOOS`/`GOARCH` exported. A non-Go project either leaves the
  gate disabled (default OFF) or supplies its own command template. Nothing is
  hard-coded to Go in the gate logic; Go specifics live only in this repo's
  `centinela.toml`.
- Default is **disabled** (`build = false`), matching how `i18n` ships off.
  This repo's `centinela.toml` opts in. Existing non-Go consumers are
  unaffected — additive, zero behavior change unless enabled.

## 2. Decision: single source of targets (no drift with release.yml)

The release matrix lives in `.github/workflows/release.yml`:
`{goos:[linux,darwin,windows], goarch:[amd64,arm64]}` → 6 targets.

**Recommendation: the gate's `targets` in `centinela.toml` are the source of
truth, and a Go test asserts the release matrix equals that list.** Concretely:
- Gate reads `targets` from `[gates.build]` in `centinela.toml`.
- A parity test (`internal/gates/build_matrix_parity_test.go`, ≤100 lines)
  parses `centinela.toml` `[gates.build].targets` AND parses the `strategy.
  matrix` block of `.github/workflows/release.yml`, expands the cross-product,
  and asserts the two sets are equal. If anyone edits one without the other,
  `go test ./...` (already in `validate.commands`) fails — drift is caught at
  validate, not release.
- Rejected alternative: having the release job *consume* a generated target
  file. It is heavier (codegen + workflow rewrite), touches the release
  pipeline (out of locked scope), and a parity test gives the same anti-drift
  guarantee with far less surface.

## 3. Decision: performance

Six `go build` cross-compiles add seconds. Mitigations, in order:
- **Build-only, no test execution per target** (`go build`, not `go test`) — the
  gate proves *compilability* per platform, which is exactly the hole the brief
  describes. Correctness is already covered by the host `go test ./...`.
- **`CGO_ENABLED=0`** exported for every cross-build (brief edge case: no C
  toolchain on runner; also makes builds deterministic + cache-friendly).
- **Build to `os.DevNull`** (`-o /dev/null` equivalent via `io.Discard`/temp) so
  no artifacts are written.
- **Reuse the Go build cache** (do NOT pass `-a`); second `validate` run is
  near-instant for unchanged code.
- **Parallelism**: run the N target builds concurrently with a bounded
  worker pool (GOMAXPROCS-sized), collecting failures. 6 builds in parallel is
  a few seconds cold, sub-second warm.
- **Honoring "caught at validate, don't water it down":** the gate runs the
  FULL matrix in `validate` by default. We do NOT silently drop to a subset.
  If a future opt-out is wanted, expose `[gates.build] local_subset` as an
  *explicit* config knob (documented as a known weakening) rather than a hidden
  default. v1 ships full-matrix everywhere.

## 4. Layered design (G2/G7)

All decisions in `internal/`; `cmd/` stays thin wiring (it already calls
`gates.RunWithFilter`, so the gate appears with zero `cmd/` business logic).

| File | Layer | Role | Lines (budget) |
|------|-------|------|------|
| `internal/config/build_gate.go` | config (leaf) | `BuildGateConfig` struct (`Enabled`, `Command`, `Targets []BuildTarget{GOOS,GOARCH}`), `NormalizeBuildGate` defaults | ≤60 |
| `internal/config/config.go` (edit) | config | add `Build BuildGateConfig` field to `GatesConfig`; call normalize in `applyDefaults` | +5 |
| `internal/gates/build.go` | domain | `checkBuild(cfg) Result` — orchestrates target loop, aggregates failures into `Result.Details` | ≤70 |
| `internal/gates/build_runner.go` | domain | `buildTarget(cmd string, t BuildTarget) error` — exec `go build` with `GOOS/GOARCH/CGO_ENABLED=0` env, capture stderr; bounded parallel runner | ≤90 |
| `internal/gates/gates.go` (edit) | domain | add `if cfg.Gates.Build.Enabled { results = append(results, checkBuild(cfg)) }` in `RunWithFilter` | +3 |

No new `cmd/` file. No `internal/ui` change (existing `RenderGateResult`
handles the new `Result`). G2: `internal/gates` imports only `internal/config`
(+ stdlib `os/exec`). G7: `cmd/` untouched beyond what already exists.

Gate name string: `"G-Build: Cross-Compile"` (PassMessage e.g. "All 6 release
targets compile."; FailMessage "These release targets failed to build:" with
each `goos/goarch` + first error line as a `Details` entry).

### centinela.toml addition (this repo)
```toml
[gates]
file_size = true
build = true

[gates.build]
command = "go build ./cmd/centinela"
targets = [
  { goos = "linux",   goarch = "amd64" },
  { goos = "linux",   goarch = "arm64" },
  { goos = "darwin",  goarch = "amd64" },
  { goos = "darwin",  goarch = "arm64" },
  { goos = "windows", goarch = "amd64" },
  { goos = "windows", goarch = "arm64" },
]
```

## 5. Self-dependency / sequencing

This branch is off `main` at v0.5.0 and still contains the broken `lock.go`
(pre-#10). The gate, once wired and enabled, will **correctly fail its own
`validate`** (both `windows/*` targets fail to compile) until PR #10 merges and
this branch rebases. This is the gate working as designed — proof of value.

Sequence: implement gate (code step) → **rebase onto post-#10 `main`** BEFORE
the validate step → then `centinela validate` goes green. Do not attempt to
pass validate before the rebase; do not weaken the gate to get green.

## 6. Rollout (smallest correct slice first)

- **Slice 1 — config plumbing.** `BuildGateConfig` + `GatesConfig.Build` field +
  `NormalizeBuildGate` defaults + config unit tests. No behavior yet. Smallest
  correct, fully testable in isolation.
- **Slice 2 — gate logic (single target, serial).** `checkBuild` + `buildTarget`
  building one target serially; wire into `RunWithFilter`; enable in
  `centinela.toml`. Proves the gate appears in the report and fails on a broken
  target. (On this pre-#10 branch it will fail windows — expected.)
- **Slice 3 — full matrix + parallelism + perf.** Bounded worker pool,
  `CGO_ENABLED=0`, build-to-discard, full 6-target list.
- **Slice 4 — anti-drift parity test.** `build_matrix_parity_test.go` comparing
  `centinela.toml` targets vs `release.yml` matrix.
- **Slice 5 — rebase onto post-#10 main**, confirm `validate` green, docs.

Each slice keeps every source file ≤100 lines and leaves the tree
test-passing (except the deliberate, documented windows failure on the
pre-rebase branch).

## 7. Tests (preview for qa-senior)

- Unit: `NormalizeBuildGate` defaults, empty/extra target handling.
- Unit: `buildTarget` returns error naming the target on a synthetic
  unbuildable command; success on a trivial one.
- Unit: `checkBuild` aggregates multiple failures into `Details`; Pass when all
  succeed; Skip/Pass when disabled.
- Parity test: toml targets == release.yml matrix cross-product.
- Acceptance (Gherkin): "a release target fails to build → validate fails
  naming GOOS/GOARCH" and "all targets build → validate passes."
