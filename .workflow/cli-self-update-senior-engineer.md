# cli-self-update — senior-engineer

## Files Touched

New pure-leaf package `internal/selfupdate` (stdlib only, all files ≤100 lines):

| File | Lines | Responsibility |
|------|------:|----------------|
| `doer.go` | 58 | package doc, `Doer` HTTP seam, consts, `Updater` struct, `New` |
| `version.go` | 24 | dev sentinel, `normalize` (strip leading v), `behind`, `availableMsg` |
| `asset.go` | 18 | `AssetName(goos,goarch,tag)` — tag keeps leading v, `.exe` on windows |
| `errors.go` | 38 | typed `Error{Kind,Msg,Err}` + `Kind` constants |
| `release.go` | 53 | `Release` parse + `resolveLatest` (GET releases/latest) |
| `cache.go` | 62 | XDG TTL cache at `${XDG_CACHE_HOME:-~/.cache}/centinela/update-check.json` |
| `download.go` | 50 | `fetchBytes`, `sumFor` (coreutils 2-space SHA256SUMS), `verify` |
| `replace.go` | 71 | `targetPath` (Executable+EvalSymlinks), atomic `replaceBinary`, `writeTemp` |
| `install.go` | 36 | download → verify → atomic replace orchestration |
| `update.go` | 63 | `Check` (read-only), `Update`, `latestTag` (cache-aware) |
| `notice.go` | 19 | fail-silent `Notice` for SessionStart |

`cmd/centinela`: new thin `update.go` (55, `--check` flag + `newSelfUpdater`
seam); `hook_session.go` (48, `emitUpdateNotice` appended); `centinela.toml`
registers `internal/selfupdate/**` in the `leaf` layer.

## Architecture Compliance

- **G1** all source + test files ≤100 lines (validated by the gate).
- **G7** cmd is wiring only — every decision lives in `internal/selfupdate`.
- **G2** `internal/selfupdate` is a pure leaf: imports only stdlib (verified via
  `go list` — no dotted import paths), computes the XDG path from env directly
  (no `internal/config`), and is mapped to the `leaf` layer. import_graph gate
  passes with no failing edge.
- Typed errors only; no panics on bad input. HTTP behind an injected `Doer`;
  binary-replace target, symlink resolver, and fsync/chmod behind seams so all
  paths are testable offline (httptest.Server + temp HOME/XDG).

## Type-Safety Notes

Strict typing throughout: `Kind` is a string enum for error classification;
`Updater` holds explicit injectable fields (no globals). No `interface{}` except
the JSON-encode map in test helpers. Checksum compared as lowercased hex.

## Trade-Offs

- Version comparison is normalize-then-equality (release tag is authoritative
  "latest"), per the pinned decision — no semver ordering is required by the
  scenarios. Documented in `version.go`.
- `Update` resolves fresh each run (needs the asset list) and does **not** write
  the cache, keeping the "already up to date" path strictly write-free; the
  cache is owned by the read-only `Check`/`Notice` throttle.
- Asset is downloaded fully into memory and verified before any disk write, so
  the checksum-mismatch / missing-asset / permission paths never create a temp
  file (satisfies "no partial write" without extra cleanup logic).

## Verification

- `go build ./...` + `go vet` clean. Tests green.
- Coverage: `internal/selfupdate` **98.2%**, `cmd/centinela` 91.2%
  (`update.go`/`hook_session.go` ~100%), repo total **95.1%** (gate ≥95.0).

## Handoff

→ **qa-senior** (tests step): author the full Gherkin-mapped suite under
`tests/`, the acceptance tier, and `.workflow/cli-self-update-edge-cases.md`.
The `newSelfUpdater` / `Target` / `osExecutable` / `evalSymlinks` / `syncFn` /
`chmodFn` / `writeTempFn` seams make every error branch injectable.
