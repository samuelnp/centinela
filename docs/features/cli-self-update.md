# Feature Brief: cli-self-update

## Problem

Centinela is distributed as a prebuilt release binary (GitHub Releases publish
`centinela-v<tag>-<goos>-<goarch>[.exe]` for linux/darwin/windows × amd64/arm64,
plus a `SHA256SUMS` asset). A user who installed a release binary has **no
in-tool path to upgrade** — they must manually find the latest release,
download the right asset, verify it, and replace the binary by hand. The result:
people silently keep running stale governance (old gates, old enforcement, known
bugs already fixed upstream). `centinela doctor` can detect local dev-build
drift, but there is nothing for the released-binary user.

**Who is hurting:** any user/operator running an installed Centinela release
binary — especially in CI images and on teammates' machines where rebuilding
from source is not the workflow.

## User Stories

- As a **Centinela user**, I want to run `centinela update` and have it fetch,
  verify, and install the latest release for my platform, so I stay current
  without manual download steps.
- As a **cautious user**, I want `centinela update --check` to tell me whether
  I'm behind **without changing anything**, so I can decide when to upgrade.
- As a **user**, I want a quiet "update available" notice at session start
  (never a silent auto-install, never a blocking prompt), so I learn about new
  versions without it getting in my way or hammering the network.
- As a **security-conscious operator**, I want the downloaded binary verified
  against the release `SHA256SUMS` and the swap to be atomic, so a corrupted or
  tampered download can never replace my working binary.

## Acceptance Criteria

(Concrete and testable — these become the Gherkin scenarios.)

1. **Update happy path.** On an outdated binary, `centinela update` resolves the
   latest release, downloads the asset matching the host `GOOS/GOARCH`, the
   SHA256 matches the entry in `SHA256SUMS`, the running binary is replaced
   atomically, and it prints `old -> new` versions. On the latest version it is
   a no-op printing `already up to date`.
2. **`--check` is read-only.** `centinela update --check` performs the version
   check (honoring the TTL cache), prints the availability verdict, exits
   non-zero when a newer version exists and zero when current, and makes
   **zero writes** to the binary or any temp file.
3. **Checksum mismatch fails safe.** A `SHA256SUMS` mismatch aborts **without
   touching** the installed binary and removes the temp file.
4. **Unsupported platform.** No matching asset for the host `GOOS/GOARCH`
   returns a clear typed error with no partial write.
5. **Permission denied fails safe.** When the target binary's directory is not
   writable, `centinela update` returns a clear typed error, leaves the
   installed binary untouched, and cleans up any temp file it created.
6. **Startup notice is throttled & silent-on-failure.** The SessionStart notice
   is cache-throttled (a second start within the TTL performs **no** network
   call), fails silent when offline/rate-limited (no error, no block), and
   appears only when running `< latest`. It never auto-installs.
7. **Deterministic tests.** GitHub API, asset download, `SHA256SUMS`, and the
   cache file are exercised against an `httptest.Server` and a temp HOME/XDG
   dir — no real network, no real GitHub.

## Edge Cases

- **Already current** → `already up to date`, exit 0, no download.
- **Checksum mismatch / truncated download** → abort, binary untouched, temp
  removed (AC3).
- **Unsupported / missing asset** for host platform → typed error (AC4).
- **Unwritable install dir** (e.g. system path, read-only FS) → typed error,
  untouched binary, temp cleaned (AC5).
- **Offline / GitHub rate-limited / API 4xx-5xx** → `update` returns a clear
  error; the **startup notice** instead fails silent (never blocks a session).
- **Cache within TTL** → no network call (AC6); **cache stale/missing/corrupt**
  → one network check, cache rewritten.
- **Windows `.exe`** → asset name carries `.exe`; replacing a running `.exe` on
  Windows may require the rename-then-delete dance (note for plan).
- **Running binary not on a writable path / symlinked** → resolve the real
  target path (`os.Executable` + EvalSymlinks) before swapping.

## Data Model

No persisted domain entities. Types:

- `Release` — latest tag + asset list (name → download URL), parsed from the
  GitHub Releases API response.
- `UpdateCheck` (cache) — `{ latestTag string, checkedAt int64 }`, persisted as
  JSON at `${XDG_CACHE_HOME:-~/.cache}/centinela/update-check.json`.
- `assetName(goos, goarch, tag) string` — pure mapping to
  `centinela-v<tag>-<goos>-<goarch>[.exe]`.

## Integration Points

- **New leaf package `internal/selfupdate`** (`net/http`, `crypto/sha256`,
  `os`, `encoding/json` only) — release resolution, asset download, checksum
  verify, atomic replace, cache read/write. HTTP behind an injected interface
  so tests drive an `httptest.Server`.
- **`cmd/centinela`** — new `update` command (`--check` flag); wires the
  startup notice into the **existing** SessionStart hook (`centinela hook
  session`). cmd stays thin (G7): all logic in `internal/selfupdate`.
- **External:** GitHub Releases API + release assets (`SHA256SUMS` + per-platform
  binaries) — read-only, unauthenticated.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Replacing a running binary mid-execution corrupts it | High | Low | Atomic write-temp-in-same-dir + `os.Rename`; verify SHA256 before rename; Windows rename-then-delete note |
| Tampered/corrupted download installed | High | Low | Mandatory SHA256SUMS verification before swap; abort + cleanup on mismatch (AC3) |
| Startup notice adds latency or network on every session | Medium | Medium | TTL cache (24h), no network within TTL, fail-silent (AC6) |
| GitHub API rate-limit (unauthenticated) | Medium | Medium | Cache throttling; fail-silent on the notice; clear error only on explicit `update` |
| `os/exec`-free networking pulls a new dependency | Low | Low | Use stdlib `net/http` + `crypto/sha256` only — no new deps |
| File-size G1 on the new package | Low | Medium | Split selfupdate into small files (resolve / download / verify / replace / cache) |

## Decomposition

Single feature, sequenced in the plan:

1. `internal/selfupdate` core: release resolution + `assetName` mapping +
   download + SHA256 verify + atomic replace (behind an HTTP interface).
2. TTL cache (read/write at the XDG path) + `--check`.
3. `cmd/centinela update` command + the SessionStart startup notice.

No sub-features split out. Out of scope: auto-install on startup, package-manager
distribution (brew/apt), and delta/partial updates.
