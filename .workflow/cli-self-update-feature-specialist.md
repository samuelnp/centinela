### Feature-Specialist Report: cli-self-update
**Date:** 2026-06-30

#### Behavior Summary

`centinela update` adds a self-upgrade path for users running installed Centinela release binaries. On invocation it resolves the latest GitHub Release for the repo, selects the asset matching the host `GOOS/GOARCH` (using the verified asset-name format `centinela-<v-prefixed-tag>-<goos>-<goarch>[.exe]`), downloads it, verifies its SHA256 against the `SHA256SUMS` release asset, and atomically replaces the running binary by writing a temp file in the same directory and calling `os.Rename`. The `--check` flag is a zero-write read-only variant: it prints an availability verdict, honors a 24-hour TTL cache, and exits non-zero when a newer version exists. The TTL cache (JSON at `${XDG_CACHE_HOME:-~/.cache}/centinela/update-check.json`) is also used by a passive notice wired into the existing `centinela hook session`: the notice appears only when the running version is below the latest, never when already current or when the binary reports the `"dev"` sentinel, and it fails silently on any network or API error so it never blocks a session. All network calls are behind an injected `Doer` interface, allowing the full update flow to be exercised against an `httptest.Server` with a temp HOME/XDG dir — no real network, no real GitHub.

#### Acceptance Criteria (Gherkin)

See `specs/cli-self-update.feature` for the full set. Scenario titles and coverage mapping:

**AC1 — Update happy path**
- `Update installs a newer release and prints old and new versions` — download, SHA256 verify, atomic replace, version output
- `Update is a no-op when already on the latest version` — no-download branch, "already up to date" output, exit 0

**AC2 — --check is read-only**
- `--check reports a newer version and exits non-zero with zero writes`
- `--check reports already current and exits zero with zero writes`
- `--check honors the TTL cache and makes no network call within TTL`

**AC3 — Checksum mismatch fails safe**
- `Checksum mismatch aborts without touching the installed binary` — abort, temp removed, binary byte-identical

**AC4 — Unsupported platform**
- `Missing asset for the host platform returns a typed error with no partial write`

**AC5 — Permission denied fails safe**
- `Unwritable install directory returns a typed error and leaves binary untouched`

**AC6 — Startup notice throttled and fail-silent**
- `Startup notice appears when running an older version and cache is stale`
- `Startup notice is suppressed when the cache is within the TTL`
- `Startup notice is suppressed when already on the latest version`
- `Startup notice fails silently when the GitHub API is unreachable`
- `Startup notice never auto-installs`

**AC7 — Deterministic tests**
- `All network calls target the httptest.Server and not the real GitHub API`

**Big-thinker handoff — version-string normalization**
- `Version comparison strips leading v from the release tag`
- `Asset name is constructed with the leading v from the tag`
- `Asset name for Windows carries the .exe suffix`

**Big-thinker handoff — dev build sentinel**
- `dev build prints an informational message and skips the update`
- `dev build suppresses the startup notice`

**Additional edge cases**
- `Update resolves a symlinked binary to its real path before replacing`
- `Explicit update returns a clear error when GitHub API is unreachable`
- `Stale cache older than the TTL triggers a fresh network check`
- `Corrupt or empty cache file triggers a fresh network check without panic`
- `GitHub API 429 during startup notice fails silently`
- `GitHub API 403 during explicit update returns a clear typed error`

#### UX States

| State   | Trigger                                           | Surface                          |
|---------|---------------------------------------------------|----------------------------------|
| loading | `centinela update` resolving / downloading        | stdout progress line (optional)  |
| empty   | already up to date (`update` or `--check`)        | "already up to date" to stdout   |
| error   | checksum mismatch, unsupported platform, no write | error line to stderr, non-zero exit |
| success | update installed                                  | "0.37.0 -> 0.40.2" to stdout, exit 0 |
| notice  | session start, running < latest, cache stale      | one-line hint appended to hook session output |

#### Edge Cases

- Already current: "already up to date", exit 0, no download (AC1 no-op scenario).
- Checksum mismatch / truncated download: abort, binary untouched, temp removed (AC3).
- Unsupported / missing asset for host platform: typed error, no partial write (AC4).
- Unwritable install dir: typed error, untouched binary, temp cleaned (AC5).
- Offline / GitHub rate-limited / API 4xx-5xx: `update` returns clear error; startup notice fails silently (AC6).
- Cache within TTL: no network call (AC2 TTL scenario, AC6 TTL suppression).
- Cache stale/missing/corrupt: one network check, cache rewritten (stale + corrupt scenarios).
- dev build: informational message, exit 0, no download, no temp file; startup notice suppressed entirely.
- Version-string normalization: tag `v0.40.2` vs ldflag `0.40.2` — strip leading v on both sides.
- Symlinked binary: resolve real target via `os.Executable` + `EvalSymlinks` before swap.
- Windows .exe: asset name carries `.exe`; rename-then-delete dance documented in plan.
- GitHub API 429 during session start: fails silently, no error printed.
- GitHub API 403 during explicit update: clear typed error, non-zero exit.

#### Out-of-Scope

- Auto-install on startup (the notice is passive only — never installs).
- Package-manager distribution (brew, apt, scoop).
- Delta or partial updates (full-asset replace only).
- Authenticated GitHub API (unauthenticated read-only only).
- Rollback of the previous binary after a successful update.
- Windows in-place `.exe` swap runtime coverage (residual risk; rides the existing cross-platform build gate).

#### Pinned Decisions

See `docs/plans/cli-self-update.md` — Pinned Decisions for full rationale.

1. **`internal/selfupdate` is a pure leaf.** No import of `internal/config`. XDG path computed from `os.Getenv("XDG_CACHE_HOME")` with stdlib `path/filepath` fallback. Code-step task: register it as `leaf` in `[gates.import_graph]` in `centinela.toml`.
2. **`dev` build: informational no-op (exit 0, no download, no temp file).** Startup notice suppressed entirely for `dev` builds (no network call).

#### Deferred Findings

No new out-of-scope discoveries. All "Out" items were pre-agreed exclusions stated in the brief and ROADMAP. Windows in-place `.exe` swap is a residual risk captured in the plan, riding the existing cross-platform build gate. **Recorded slugs: none.**

#### Handoff

- **Next role:** senior-engineer
- **Open clarifications:** none — both open questions from the big-thinker have been pinned in docs/plans/cli-self-update.md.
- **Code-step task reminder:** Add `internal/selfupdate` to the `leaf` layer in `[gates.import_graph]` in `centinela.toml` and verify with `centinela validate --gate import_graph` after the package is created.
