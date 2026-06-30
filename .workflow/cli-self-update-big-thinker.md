### Big-Thinker Report: cli-self-update
**Date:** 2026-06-30

#### Problem
Centinela ships as a prebuilt GitHub Release binary
(`centinela-v<tag>-<goos>-<goarch>[.exe]` plus a `SHA256SUMS` asset). A user on an
installed release binary has no in-tool upgrade path — they must manually find the
latest release, pick the right asset, verify it, and swap the binary by hand. The
result: people (especially in CI images and on teammates' machines) silently keep
running stale governance — old gates, old enforcement, already-fixed bugs.
`centinela doctor` catches local dev-build drift but does nothing for the
released-binary user. We close the gap with `centinela update` plus a quiet,
non-blocking startup notice that never auto-installs.

#### Scope
- **In:** new **leaf** package `internal/selfupdate` (stdlib only — `net/http`,
  `crypto/sha256`, `encoding/json`, `os`, `path/filepath`, `runtime`, `time`; no
  new deps, no internal imports); `centinela update` (resolve latest → select host
  asset → download → verify `SHA256SUMS` → atomic same-dir temp + fsync + copy mode
  + `os.Rename`); `centinela update --check` (read-only verdict, TTL-cached, zero
  writes, non-zero exit when behind); 24h TTL cache at
  `${XDG_CACHE_HOME:-~/.cache}/centinela/update-check.json`; throttled fail-silent
  notice in the existing `centinela hook session`; HTTP behind an injected
  interface for `httptest.Server` + temp HOME/XDG tests.
- **Out (deliberate, pre-known):** auto-install on startup, package-manager
  distribution (brew/apt), delta/partial updates.

#### Dependencies & Assumptions
- Release workflow (`release.yml`) is the contract. Verified live against v0.40.2:
  asset literal `centinela-<tag>-<goos>-<goarch>[.exe]` with `<tag>` carrying the
  leading `v` (e.g. `centinela-v0.40.2-darwin-arm64`,
  `centinela-v0.40.2-windows-arm64.exe`); `SHA256SUMS` is coreutils format
  `<64-hex><two spaces><filename>` per line.
- Running version is ldflag-injected `${TAG#v}` (no `v`, e.g. `0.40.2`); release
  `tag_name` is `v`-prefixed; dev builds report `"dev"`. Comparison normalizes by
  stripping a leading `v`; `"dev"` is an uncomparable sentinel (notice suppressed).
- Repo slug `samuelnp/centinela` ⇒ API
  `https://api.github.com/repos/samuelnp/centinela/releases/latest`, unauth,
  read-only.
- SessionStart hook is `cmd/centinela/hook_session.go` (`runHookSession`).
- Coverage gate is total `go test` ≥95.0% (no `-coverpkg`) — needs colocated
  `_test.go`; G1 (≤100 lines) applies to those test files.
- `internal/selfupdate` is new; add it to the **leaf** layer in
  `[gates.import_graph]` (`centinela.toml`) or the gate warns "unassigned package".
- G7: `cmd/centinela/update.go` is wiring only; the session hook calls one
  fail-silent notice function.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Replacing running binary corrupts it | High | Low | Verify SHA256 before any rename; temp in same dir + fsync + copy mode + `os.Rename`; Windows rename-then-delete note |
| Tampered / truncated download installed | High | Low | Mandatory `SHA256SUMS` verify before swap; mismatch ⇒ abort + remove temp, binary untouched (AC3) |
| Version-string mismatch (`0.40.2` vs `v0.40.2`, `dev`) | Medium | High | Normalize (strip leading `v`); `"dev"` sentinel suppresses notice, explicit message in `update` |
| Coverage ≥95% total (no `-coverpkg`) on net/http pkg | Medium | High | Inject HTTP `Doer` + temp HOME/XDG; httptest + temp dirs; colocated `_test.go` ≤100 lines |
| G1 file-size on net/http-heavy package | Low | Medium | Split: client / release / asset / download / replace / cache / update / notice |
| GitHub unauth rate-limit (60/hr) | Medium | Medium | TTL cache skips network within window; notice fails silent on 4xx/5xx/offline |
| Unwritable / symlinked install dir | Medium | Medium | `os.Executable` + `EvalSymlinks` + writability precheck; typed error, temp cleaned, binary untouched (AC5) |
| Windows in-place `.exe` swap (running file) | Medium | Low | Rename-then-delete dance documented; typed error never a corrupt binary; deep Windows test rides cross-platform gate |
| New package unmapped in import_graph gate | Low | Medium | Map `internal/selfupdate` to leaf layer in `centinela.toml` during code step |

#### Rollout
- **Slice 1 — read-only resolution + `update --check` (zero-write, ships value alone):**
  `internal/selfupdate` HTTP `Doer` + `ResolveLatest` + `assetName`/version compare
  + TTL cache + `Check`; `cmd/centinela/update.go` `--check` path only. Covers AC2
  and the cache half of AC6/AC7. No risk to the installed binary.
- **Slice 2 — write path:** `download.go` (asset + `SHA256SUMS` parse + verify) and
  `replace.go` (`os.Executable`+`EvalSymlinks`, same-dir temp, fsync, copy mode,
  `os.Rename`); `Update` + default `centinela update`. Covers AC1, AC3, AC4, AC5.
- **Slice 3 — passive notice:** `notice.go` (cache-throttled, fail-silent, shows
  only when `< latest`, never installs) appended to `runHookSession`. Covers AC6.
- Dependency-ordered; Slice 1 is the only hard prerequisite. May stay one feature or
  split follow-ups if qa-senior flags sequencing pressure.

#### Deferred Findings
All "Out" items (auto-install, package-manager distribution, delta updates) are
deliberate pre-known exclusions from the brief/ROADMAP — not new discoveries — so no
`roadmap defer` recorded. Windows in-place `.exe` swap is captured as a residual risk
covered by the existing cross-platform build gate. **Recorded slugs: none.**

#### Handoff
- **Next role:** feature-specialist — author `specs/cli-self-update.feature` from
  AC1–AC7 and the edge cases; refine per-slice acceptance mapping.
- **Outstanding questions:** (1) keep `internal/selfupdate` a pure leaf computing the
  XDG path from env, or import `internal/config` for paths (plan recommends pure
  leaf); (2) `dev`-build `update` behavior — hard error vs informational no-op (plan
  recommends informational).
