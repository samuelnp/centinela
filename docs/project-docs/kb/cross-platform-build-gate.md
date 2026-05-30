---
feature: cross-platform-build-gate
summary: Centinela cross-compiles your project for every configured release target during validate and fails with the exact GOOS/GOARCH pair that breaks, so platform build errors are caught locally before the release pipeline ever runs.
audience: end-user
status: done
---

## What it does

The cross-platform build gate adds a new built-in check called **G-Build: Cross-Compile** that runs during `centinela validate`. For each platform target you list in `centinela.toml`, it compiles your project binary with the matching `GOOS` and `GOARCH` environment variables set. If any target fails to compile, the gate reports `Fail` and lists each broken `GOOS/GOARCH` pair with the first error line, so you know exactly what broke and where.

The gate is disabled by default and language-neutral: the build command is whatever you put in `[gates.build] command`, so it works for any project that can be cross-compiled via a command-line tool. For Go projects, the default is `go build ./cmd/<binary>` run with `CGO_ENABLED=0` so no C toolchain is needed on the host.

A parity test (`TestBuildMatrixParity`) keeps the target list in `centinela.toml` in sync with the release matrix in `.github/workflows/release.yml`. If someone edits one without the other, `go test ./...` fails during validate, closing the drift gap before it reaches CI.

## When you'd use it

Enable this gate whenever your project ships release binaries for multiple platforms and you want to catch OS-specific compilation errors locally — before the release job discovers them. It is especially useful after changes to OS-specific code paths, build tags, or dependencies that might only exist on certain platforms.

## How it behaves

- When all configured targets compile cleanly, the gate reports `Pass` with a message like "All 6 release targets compile." and `centinela validate` exits 0.
- When one or more targets fail to build, the gate reports `Fail` with the message "These release targets failed to build:" and one detail line per broken `GOOS/GOARCH` (for example, `windows/amd64` and `windows/arm64`). Targets that compile successfully are not listed. `centinela validate` exits non-zero.
- When the gate is disabled (`[gates] build = false`), it does not appear in the gate report at all and validate proceeds normally.
- When the centinela.toml target list drifts from the release.yml matrix, the parity test `TestBuildMatrixParity` fails during `go test ./...` and names the targets that are present in one place but not the other.
- Each cross-compile is run with `CGO_ENABLED=0` so the gate works on any host without a C toolchain.
- When the target list is empty and the gate is enabled, the gate reports `Pass` with the message "No targets configured; skipping cross-compile."
- When the build cache is warm (a second validate run with no source changes), each cross-compile completes in well under one second per target.
- When a target entry is syntactically valid but the platform is unknown to the Go toolchain, the gate reports `Fail` and names the unknown target with a descriptive error rather than panicking.

## Examples

Enable the gate and configure your release targets in `centinela.toml`:

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

Sample gate output when everything compiles:

```
G-Build: Cross-Compile  PASS  All 6 release targets compile.
```

Sample gate output when a platform-specific call breaks Windows builds:

```
G-Build: Cross-Compile  FAIL  These release targets failed to build:
  · windows/amd64  undefined: syscall.Flock
  · windows/arm64  undefined: syscall.Flock
```
