# Plan: `centinela update` — Self-Update + Passive Startup Notice

References: [docs/features/cli-self-update.md](../features/cli-self-update.md),
ROADMAP.md (Phase 5 · `cli-self-update`),
[.github/workflows/release.yml](../../.github/workflows/release.yml),
[PROJECT.md](../../PROJECT.md) (G2 layer rules, G7 thin-cmd).

## Problem

Centinela ships as a prebuilt GitHub Release binary
(`centinela-v<tag>-<goos>-<goarch>[.exe]` for linux/darwin/windows × amd64/arm64
plus a `SHA256SUMS` asset). A user on an installed release binary has **no
in-tool upgrade path**: they must hand-find the latest release, pick the right
asset, verify it, and swap the binary manually. `centinela doctor` catches local
dev-build drift, but the released-binary user (CI images, teammates' machines)
silently keeps running stale governance — old gates, old enforcement, already-fixed
bugs. We close that gap with a self-update command and a quiet, non-blocking
"update available" notice, without ever auto-installing.

## Verified ground truth (read from the real repo)

- **Asset name (exact):** `centinela-<tag>-<goos>-<goarch>[.exe]` where `<tag>`
  carries the leading `v` — confirmed live: `centinela-v0.40.2-darwin-arm64`,
  `centinela-v0.40.2-windows-arm64.exe`. So
  `assetName(goos, goarch, tag) = "centinela-" + tag + "-" + goos + "-" + goarch + ext`
  with `ext == ".exe"` only on windows.
- **`SHA256SUMS` format:** coreutils `sha256sum *` output —
  `<64-hex><two spaces><filename>` per line, one line per asset (confirmed).
- **Running version:** ldflag-injected `-X main.Version=${TAG#v}`, i.e. **without**
  the `v` (binary reports `0.40.2`); dev builds report the sentinel `"dev"`
  (`cmd/centinela/main.go`). Latest release `tag_name` is `v0.40.2` (with `v`).
  Comparison MUST normalize by stripping a leading `v` on both sides; `"dev"` is
  a sentinel (uncomparable) — see Risks.
- **Repo slug:** module `github.com/samuelnp/centinela` ⇒ API base
  `https://api.github.com/repos/samuelnp/centinela/releases/latest`
  (`tag_name`, `assets[].name`, `assets[].browser_download_url`), unauthenticated,
  read-only.
- **SessionStart hook:** `cmd/centinela/hook_session.go` (`runHookSession`) already
  drains stdin and prints roadmap rehydration. The notice appends here.
- **Coverage gate:** `scripts/check-coverage.sh` enforces total `go test` line
  coverage ≥ `95.0` (no `-coverpkg`), so the new package needs **colocated**
  `_test.go` files; G1 (≤100 lines) applies to those test files too.

## Scope

**In (v1):**

- New **leaf** package `internal/selfupdate` — stdlib only (`net/http`,
  `crypto/sha256`, `encoding/json`, `os`, `path/filepath`, `runtime`, `time`,
  `io`, `fmt`, `errors`). No new third-party deps. No internal imports (computes
  the XDG path from env itself to stay a pure, testable leaf).
- `centinela update`: resolve latest release → select host asset → download →
  verify against `SHA256SUMS` → **atomic** replace (temp file in the SAME dir as
  the resolved target, write, `fsync`, copy existing mode bits, `os.Rename`).
  Prints `old -> new`; no-op `already up to date` when current.
- `centinela update --check`: **read-only** version verdict honoring the TTL
  cache; exit non-zero when behind, zero when current; **zero writes** to the
  binary or any temp file.
- TTL cache (default 24h) at `${XDG_CACHE_HOME:-~/.cache}/centinela/update-check.json`
  holding `{ latestTag, checkedAt }`; within-TTL reads perform **no** network call.
- Throttled, fail-silent startup notice wired into the **existing** `centinela hook
  session`; shows only when running `< latest`; never auto-installs, never blocks.
- HTTP behind an injected interface so all tests drive an `httptest.Server` with a
  temp HOME/XDG dir — no real network, no real GitHub.

**Out (deliberate, pre-known — not new discoveries):**

- Auto-install on startup (notice is passive only).
- Package-manager distribution (brew/apt/scoop).
- Delta/partial updates (full-asset replace only).

## Dependencies & Assumptions

- **Release workflow is the contract.** `release.yml` already publishes the assets
  + `SHA256SUMS` exactly as parsed above; `assetName` must mirror it byte-for-byte.
- **Version normalization.** Binary `Version` is `v`-stripped; release `tag_name`
  is `v`-prefixed. Normalize both before compare. `"dev"` ⇒ notice suppressed and
  `update` prints a clear "development build" message (no swap target version).
- **G2 import graph.** `internal/selfupdate` is a NEW package; the
  `[gates.import_graph]` layer map in `centinela.toml` must add it to the **leaf**
  layer (else the gate surfaces a non-failing "unassigned package" warning).
  Confirm/extend during the code step.
- **G7 thin cmd.** `cmd/centinela/update.go` is wiring only; all logic lives in
  `internal/selfupdate`. The session hook calls one fail-silent notice function.
- **Self-leaf path resolution.** Use `os.Executable()` + `filepath.EvalSymlinks`
  to find the real target before swapping (handles symlinked installs).
- **Scaffold mirror.** No architecture-doc changes expected; if any
  `docs/architecture/*` is touched, mirror into `internal/scaffold/assets/`
  ([[project_scaffold_mirror_partial_parity]]).

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Replacing a running binary corrupts it | High | Low | Verify SHA256 **before** any rename; write temp in SAME dir + `fsync` + copy mode + `os.Rename` (atomic on POSIX); Windows rename-then-delete note |
| Tampered / truncated download installed | High | Low | Mandatory `SHA256SUMS` verify before swap; on mismatch abort + remove temp, binary untouched (AC3) |
| Version-string mismatch (`0.40.2` vs `v0.40.2`, `dev`) | Medium | High | Normalize (strip leading `v`) both sides; treat `"dev"` as uncomparable sentinel — suppress notice, explicit message in `update` |
| Coverage gate (≥95% total, no `-coverpkg`) on a net/http pkg | Medium | High | Inject HTTP `Doer` interface + temp HOME/XDG; exercise download/verify/replace against `httptest.Server` and temp dirs; colocated `_test.go` ≤100 lines (G1) |
| G1 file-size on a net/http-heavy package | Low | Medium | Split into small files: client / release / asset / download / replace / cache / update / notice |
| GitHub unauth rate-limit (60/hr) | Medium | Medium | TTL cache (24h) skips network within window; notice fails silent on 4xx/5xx/offline; explicit error only on `update` |
| Unwritable / symlinked install dir | Medium | Medium | `os.Executable` + `EvalSymlinks`; writability precheck; typed error, temp cleaned, binary untouched (AC5) |
| Windows in-place `.exe` swap can't rename over a running file | Medium | Low | Document rename-then-delete dance; typed error never a corrupt binary; deep Windows runtime test deferred to existing cross-platform gate |
| New package unmapped in import_graph gate | Low | Medium | Add `internal/selfupdate` to the leaf layer in `centinela.toml` during code step |

## Rollout — smallest correct slice first

**Slice 1 — read-only resolution + `centinela update --check` (zero-write, ships value alone).**
Build `internal/selfupdate`: HTTP `Doer` interface + default client (`client.go`),
`ResolveLatest` parsing `releases/latest` (`release.go`), `assetName` + version
normalize/compare (`asset.go`), TTL cache read/write at the XDG path (`cache.go`),
and `Check` orchestration (`update.go`). Wire `cmd/centinela/update.go` with only
the `--check` path: prints the verdict, honors the cache, exits non-zero when
behind, **makes no writes**. Fully testable via `httptest.Server` + temp HOME/XDG.
This is the smallest slice that ships user value (they learn they're behind) with
**zero** risk to the installed binary — covers AC2 and the cache half of AC6/AC7.

**Slice 2 — write path: download + verify + atomic replace + `centinela update`.**
Add `download.go` (fetch asset + `SHA256SUMS`, parse the two-space format, verify
sha256) and `replace.go` (resolve target via `os.Executable`+`EvalSymlinks`,
write temp in same dir, `fsync`, copy mode, `os.Rename`). Extend `update.go` with
`Update`. Flip on the default `centinela update` path. Covers AC1 (happy path +
no-op), AC3 (checksum mismatch fail-safe), AC4 (unsupported platform typed error),
AC5 (permission-denied fail-safe + temp cleanup).

**Slice 3 — passive SessionStart notice.**
Add `notice.go` (cache-throttled, fail-silent, shows only when `< latest`, never
installs) and append it to `runHookSession` in `cmd/centinela/hook_session.go`.
Covers AC6 (throttled, silent-on-failure, never auto-installs).

Slices are dependency-ordered and each is independently shippable behind the same
command; Slice 1 is the only hard prerequisite for Slices 2–3. The qa-senior step
may keep them as one feature (the brief's "single feature, sequenced") or split
follow-ups if sequencing pressure appears.

## Deferred Findings

All "Out" items (auto-install, package-manager distribution, delta updates) are
**deliberate, pre-known** exclusions stated in the brief and ROADMAP — not new
discoveries — so no `roadmap defer` is recorded. The Windows in-place `.exe` swap
is captured here as a residual risk; deep Windows runtime coverage rides on the
existing cross-platform build gate rather than a new roadmap item. **Deferred slugs:
none.**

## Handoff

- **Next role:** feature-specialist — author the Gherkin `.feature` spec at
  `specs/cli-self-update.feature` from AC1–AC7 and the edge cases, and refine the
  per-slice acceptance mapping.
- **Outstanding questions:** (1) keep `internal/selfupdate` a pure leaf computing
  the XDG path from env, or allow the `internal/config` import for path resolution?
  (plan recommends pure leaf for testability). (2) `dev`-build behavior for
  `update` — hard error vs. informational no-op (plan recommends informational).

## Pinned Decisions

_(Recorded by feature-specialist; binding on senior-engineer and qa-senior.)_

### Decision 1 — `internal/selfupdate` is a pure leaf (no `internal/config` import)

**Resolved: pure leaf.**

`internal/selfupdate` MUST NOT import `internal/config` or any other internal
package. It computes the XDG cache path directly from the environment variable
`XDG_CACHE_HOME` (falling back to `~/.cache`), constructing the path
`<cache_home>/centinela/update-check.json` with stdlib `os` and `path/filepath`.

Rationale:

1. **Testability.** Test code sets `HOME` / `XDG_CACHE_HOME` in a temp dir and
   the leaf reads from there naturally. Importing `internal/config` would pull
   in its initialization path and any future dependency it grows, complicating
   test isolation.
2. **G2 import graph.** A leaf that imports another internal package is no longer
   a leaf; the import_graph gate would classify it as a utility or domain layer,
   creating a dependency edge that is harder to reverse later.
3. **Simplicity.** The XDG path computation is two lines of stdlib; the benefit
   of reusing `internal/config` is near-zero, while the coupling cost is real.
4. **`internal/config` is upstream.** Importing it risks future circular imports
   if config ever needs to consult the selfupdate package for anything.

**Code-step task:** Register `internal/selfupdate` as a **leaf** in the
`[gates.import_graph]` layer map in `centinela.toml`. Confirm this during the
code step by running `centinela validate --gate import_graph` after the package
is created.

### Decision 2 — `dev` build prints an informational message and skips the update

**Resolved: informational no-op (not a hard error).**

When `centinela update` is run on a binary whose version string is the sentinel
`"dev"` (i.e. a local development build), the command MUST:

1. Print an informational message to stdout, e.g.:
   `"centinela update: this is a development build — self-update is not available"`
2. Exit with code 0.
3. Make no download, no temp file, and no replacement of the binary.

For the startup notice (`centinela hook session`), a `"dev"` version MUST:

1. Suppress the notice entirely (no network call, no output).
2. Not write or refresh the cache file.

Rationale:

1. **Not an error.** A developer running from source is in a legitimate and
   expected state. Exiting non-zero on `update` would break CI pipelines and
   `centinela doctor` checks that call `update` as a health probe.
2. **Consistent with `doctor`.** `centinela doctor` reports a binary-version
   mismatch as a warning, not an error. The same informational tone is
   appropriate here.
3. **Startup noise.** A dev build updating itself makes no sense and could create
   confusing feedback loops. Suppressing the notice is the safest default.
4. **Operator clarity.** An explicit message is more helpful than silent exit or
   a confusing "already up to date" message that doesn't explain why no version
   number is shown.

The Gherkin scenarios for this decision are the two `dev build` scenarios in
`specs/cli-self-update.feature`.
