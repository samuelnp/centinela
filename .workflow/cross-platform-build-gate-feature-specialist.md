### Feature-Specialist Report: cross-platform-build-gate
**Date:** 2026-05-29

#### Behavior Summary

The cross-platform build gate is a first-class, config-driven built-in gate
(`G-Build: Cross-Compile`) that lives in `internal/gates/`. When enabled via
`[gates] build = true` in `centinela.toml`, it cross-compiles the configured
command (e.g. `go build ./cmd/centinela`) for every `{GOOS, GOARCH}` pair in
`[gates.build].targets`. It runs in a bounded-parallel worker pool with
`CGO_ENABLED=0` on every target, discards the output binary, and reports a
structured `Result` via the existing `RenderGateResult` path. If every target
compiles, it emits a one-line pass message. If any target fails, it emits a
`Fail` status with the broken `GOOS/GOARCH` and the first compiler error line
as individual `Details` entries — so operators know exactly which platform
broke. A parity test (`build_matrix_parity_test.go`) cross-checks the
`[gates.build].targets` list against the `release.yml` strategy matrix; drift
causes `go test ./...` (already in `validate.commands`) to fail, preventing
silent divergence. The gate is default OFF; non-Go projects are unaffected.

---

#### Design Decisions (pinned from big-thinker open questions)

| Question | Decision |
|----------|----------|
| Config shape | Typed `{goos, goarch}` table array under `[gates.build]`. Ergonomic, IDEable. Generic `env` pairs deferred — not v1. |
| Gate name/label | `G-Build: Cross-Compile` |
| Pass message | `"All 6 release targets compile."` (count interpolated at runtime) |
| Fail message | `"These release targets failed to build:"` + one `Details` entry per broken target (`goos/goarch: <first error line>`) |
| Parallelism bound | Worker pool capped at `runtime.GOMAXPROCS(0)`; implementation detail, behavior is deterministic regardless |
| local\_subset opt-out | Deferred — v1 runs full matrix everywhere |
| Broken-target simulation | A synthetic command that exits non-zero for a chosen `$GOOS/$GOARCH` pair (shell conditional wrapping the real `go build`), injected via the configurable `command` field. No throwaway package needed. |

---

#### Gherkin Scenarios

Full spec at `specs/cross-platform-build-gate.feature` — 9 scenarios:

1. **All targets compile — gate passes and validate proceeds** (happy path)
2. **One target fails to build — gate fails naming the broken GOOS/GOARCH** (negative)
3. **Gate disabled — build check is skipped** (negative)
4. **Target list drifts from release.yml matrix — parity test fails** (negative)
5. **CGO disabled — cross-compile succeeds without a C toolchain** (edge)
6. **Unknown or garbage target reported cleanly without panic** (edge)
7. **Empty targets list — gate no-ops with a clear message** (edge)
8. **Build cache reused on second run — gate is fast** (edge)
9. **Acceptance — simulate broken target via unbuildable command** (acceptance mechanic)

---

#### UX States

| State | Trigger | Surface |
|-------|---------|---------|
| Pass | All configured targets compiled successfully | `centinela validate` gate-results block: `G-Build: Cross-Compile  PASS  All 6 release targets compile.` |
| Fail | One or more targets failed to compile | Gate-results block with `FAIL` status; `Details` lines list each `goos/goarch: <error>` |
| Skip | `[gates] build = false` (default) | Gate does not appear in output at all |
| No-op pass | `targets = []` (empty list) | `PASS  No targets configured; skipping cross-compile.` |

---

#### Out-of-Scope

- Part A — `lock.go` portability fix (done as hotfix PR #10)
- Making `release.yml` consume a generated targets file (out of locked scope)
- A `local_subset` config knob to run fewer targets locally (deferred; v1 is full-matrix everywhere)
- Per-target test execution (`go test`); this gate proves compilability only
- Non-Go language cross-compile support beyond what the generic command template provides today
- Any changes to `cmd/` business logic (zero new cmd code)

---

#### Handoff

- **Next role:** senior-engineer
- **Open clarifications for senior-engineer:**
  1. The `command` field is a bare shell string (e.g. `"go build ./cmd/centinela"`). Confirm whether the runner should execute it via `exec.Command("sh", "-c", cmd)` or parse it into argv directly. Parsing argv is safer for injection; shell invocation is needed for the acceptance simulation mechanic. Recommendation: argv parse for production command; acceptance test overrides with an explicit script path.
  2. Should `buildTarget` capture only `stderr` (where the Go compiler writes errors) or both `stdout+stderr`? Recommendation: `stderr` only, single line included in `Details`.
  3. The `Details` entry format is proposed as `"goos/goarch: <first error line>"`. Confirm whether the colon-space separator is consistent with existing gate `Details` formatting in `RenderGateResult`.
