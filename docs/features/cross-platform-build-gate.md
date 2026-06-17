# cross-platform-build-gate

> Make Centinela compile for every release target, and add a gate so platform-specific build breaks are caught at `validate` instead of at release.

> **Status note (2026-05-29):** Part A (the `lock.go` portability fix) was fast-tracked as hotfix PR **#10** (`hotfix/windows-lock`) to unblock the v0.5.0 Windows release. This feature now centers on **Part B — the cross-compile validate gate**. Dependency: this branch is off `main` at v0.5.0 and still contains the broken `lock.go`; the gate's own `validate` step will (correctly) fail until PR #10 is merged and this branch is rebased onto the updated `main`. Sequence the code/validate steps accordingly.

## Problem — what pain does this solve? Who is the user?

The v0.5.0 release action fails: `internal/evidence/lock.go` calls `syscall.Flock` / `syscall.LOCK_EX` / `syscall.LOCK_NB` / `syscall.LOCK_UN` / `syscall.EWOULDBLOCK`, which exist only on Unix. The release matrix builds `{linux, darwin, windows} × {amd64, arm64}`, so both **Windows** targets fail to compile. The bug landed in the `evidence-cli` feature (commit `16ca989`) and shipped to `main`.

The deeper pain: **Centinela's own `validate` gate never caught it.** The gate runs `go test ./...` on the host (Linux/macOS) only — where `syscall.Flock` is defined — so a release-breaking compile error was invisible through plan → code → tests → validate → docs and only surfaced in the release pipeline. For a framework whose entire value proposition is "validate before it ships," a host-only build check is a robustness hole.

**Users:** (1) anyone installing Centinela on Windows (currently no working v0.5.0 binary); (2) every Centinela-governed project that builds for multiple platforms — they inherit a gate that can't see cross-platform breaks.

## User Stories

- As a Windows user, I want the Centinela binary to compile and run, so I can install the released artifact.
- As a maintainer, I want `centinela validate` to fail locally when code breaks a release target, so I never discover it in the release job.
- As an agent operator, I want the cross-compile check to be a first-class gate (not a hidden CI step), so the workflow's "validate" actually means "buildable everywhere we ship."

## Acceptance Criteria — concrete, testable (→ Gherkin scenarios)

**Part A — portability fix (unblocks the release):**
1. `GOOS=windows GOARCH=amd64 go build ./cmd/centinela` succeeds; same for `windows/arm64`, `darwin/{amd64,arm64}`, `linux/{amd64,arm64}`.
2. Advisory file locking still works on Unix (existing `lock_test.go` behavior unchanged: exclusive acquire, busy-timeout error naming the file).
3. On Windows the lock provides **real** advisory mutual exclusion (LockFileEx/UnlockFileEx via `golang.org/x/sys/windows`), not a silent no-op.
4. `Lock`'s public signature (`func(feature string, role Role) (func(), error)`) is unchanged; callers don't change.

**Part B — cross-platform build gate:**
5. A new gate (or `validate.commands` step) cross-compiles `./cmd/centinela` for every configured release target and fails `centinela validate` if any target fails to build.
6. The set of targets is the single source of truth shared with (or matching) the release matrix, so they cannot drift.
7. The gate runs in reasonable time (build-only, `go build`, no test execution per target) and reports which `GOOS/GOARCH` failed.

## Edge Cases — invalid input, concurrency, empty state, limits

- **Two concurrent agents** acquiring the same (feature, role) lock on Windows must mutually exclude, matching Unix semantics.
- **Lock file handle leaks:** the release func must unlock AND close on every OS.
- **CGO / build constraints:** cross-compile must work with `CGO_ENABLED=0` (no C toolchain on the runner).
- **Unsupported/extra target** in config: gate reports it clearly rather than panicking.
- **Slow gate:** cross-compiling 6 targets must not balloon `validate` time unacceptably; consider build caching / parallelism.
- **Existing host build** must remain the fast default; the cross-compile is additive.

## Data Model — entities, key fields, relationships

- No new persisted entities. A build-target tuple `{GOOS, GOARCH}` list is the shared config between the gate and the release matrix.
- `internal/evidence` lock split into build-tagged files: `lock.go` (shared API + timeout consts + poll loop), `lock_unix.go` (`//go:build !windows`, flock), `lock_windows.go` (`//go:build windows`, LockFileEx).

## Integration Points — APIs, events, external services

- `internal/evidence/lock*.go` — the fix.
- `internal/gates/` or `internal/config` `validate.commands` — the cross-compile gate.
- `.github/workflows/release.yml` — the existing matrix is the reference target list; ideally the gate and the matrix read one source.
- `golang.org/x/sys/windows` — new (official, x/) dependency for the Windows lock.

## Risks — performance, security, unclear requirements

- **x/sys dependency:** adds `golang.org/x/sys` (likely already an indirect dep); low risk, official package.
- **Gate runtime:** 6 cross-compiles add seconds to every `validate`; mitigate with build cache and/or making the full matrix CI-only while a representative subset (e.g. windows/amd64) runs in `validate`.
- **Windows lock correctness:** LockFileEx semantics differ subtly from flock (mandatory vs advisory, byte-range); must lock a fixed byte range and test it. Hard to test Windows locking on a macOS dev host — rely on CI cross-build + careful implementation.
- **Drift:** if the gate's target list and the release matrix are maintained separately they will diverge; prefer a shared source.

## Decomposition — if large, split into:

- `fix-windows-lock-portability` — Part A only (build-tag split + Windows LockFileEx). Smallest correct slice; unblocks the release immediately.
- `cross-compile-validate-gate` — Part B (the gate that prevents recurrence).

These can ship as one feature (Part A first within the code step, then Part B), or as two sequential features if the release needs unblocking before the gate is designed.
