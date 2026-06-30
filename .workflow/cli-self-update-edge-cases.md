# Edge Cases: cli-self-update

Feature: `centinela update` self-update and passive startup notice.

## Identified Edge Cases

### EC-1: Version string normalization
- **Risk**: The ldflag-injected version omits the `v` prefix ("0.40.2") but the GitHub release tag always carries it ("v0.40.2"). A naive string comparison would treat them as different, triggering a spurious update loop.
- **Mitigation**: `normalize()` strips a single leading `v` before comparing. Both `0.40.2` and `v0.40.2` normalize to `"0.40.2"`.
- **Tests**: `TestCliSelfUpdate_VersionNormalizationEqual`, `TestNormalizeAndBehind`, `TestCliSelfUpdate_AssetNameWithLeadingV`.

### EC-2: Dev build sentinel
- **Risk**: A dev build (`Version = "dev"`) has no comparable release. Any comparison would produce garbage; a network call would be wasted.
- **Mitigation**: `isDev()` short-circuits both `Update()` and `Notice()` before any network or comparison logic. No HTTP call is made.
- **Tests**: `TestCliSelfUpdate_DevBuildSkipsUpdate`, `TestCliSelfUpdate_DevBuildSuppressesNotice`, `TestUpdateDevBuild`.

### EC-3: Symlinked binary resolution
- **Risk**: Users often install via symlink (`/usr/local/bin/centinela → /opt/centinela/bin/centinela`). Replacing the symlink path instead of the real file would break the installation.
- **Mitigation**: `targetPath()` calls `os.Executable()` then `filepath.EvalSymlinks()` to resolve the canonical path. Both functions are injected seams (`osExecutable`, `evalSymlinks`) so test coverage reaches these branches without a real binary.
- **Tests**: `TestCliSelfUpdate_SymlinkResolution`, `TestTargetPathResolves`, `TestTargetPathExecutableError`, `TestTargetPathSymlinkError`.

### EC-4: Checksum mismatch / tampered download
- **Risk**: A MITM or corrupted CDN asset could replace the binary with malicious content. Replacing without verification is a supply-chain risk.
- **Mitigation**: The asset is downloaded into memory first. SHA256 is verified against the release's SHA256SUMS before `replaceBinary` is called. On mismatch the temp file is never written.
- **Tests**: `TestCliSelfUpdate_ChecksumMismatch`, `TestUpdateChecksumMismatch`.

### EC-5: Missing asset / unsupported platform
- **Risk**: A release may not include an asset for a rare GOOS/GOARCH (e.g., plan9/mips). Attempting to download a non-existent URL would produce a confusing HTTP error.
- **Mitigation**: `install()` first probes `rel.assetURL(name)` and returns a `KindPlatform` typed error if absent. No download is attempted.
- **Tests**: `TestCliSelfUpdate_MissingAsset`, `TestUpdateMissingAsset`.

### EC-6: Unwritable install directory
- **Risk**: If the binary is installed system-wide (e.g., `/usr/local/bin`) and the user runs without sudo, the temp-file creation will fail. The error must be clear and the original binary must be untouched.
- **Mitigation**: `replaceBinary` wraps `os.CreateTemp` failures as `KindPermission`. The binary is never touched before the temp write succeeds.
- **Tests**: `TestCliSelfUpdate_PermissionDenied`, `TestUpdatePermissionDenied`.

### EC-7: Stale or corrupt cache
- **Risk**: A cache file older than the TTL must not be served; a corrupt/empty cache file must not panic or block the session.
- **Mitigation**: `readCache()` checks age against TTL and returns `("", false)` on any parse error. A nil/missing file is silently skipped. The session hook is never blocked.
- **Tests**: `TestCliSelfUpdate_StaleCacheRefreshes`, `TestCliSelfUpdate_CorruptCacheRefreshes`, `TestNoticeCorruptCacheRefreshes`.

### EC-8: GitHub API rate limiting (HTTP 429 / 403)
- **Risk**: CI pipelines and anonymous users frequently hit GitHub rate limits. A 429 during the passive startup notice must not error the session. A 403 during explicit `update` must produce a clear typed error.
- **Mitigation**: `Notice()` is fail-silent — any non-200 status returns `""`. `Update()` returns a `KindAPI` typed error on any non-200 that includes the status code in the message.
- **Tests**: `TestCliSelfUpdate_RateLimit429SilentOnNotice`, `TestCliSelfUpdate_RateLimit403ExplicitError`, `TestUpdateAPIError`.

### EC-9: Offline / network unreachable
- **Risk**: The user may run `centinela update` in an air-gapped environment or with a firewall blocking GitHub. The error must be a typed `KindNetwork` error, not a raw Go error dump.
- **Mitigation**: Transport failures are wrapped with `KindNetwork` kind. The session notice (`Notice()`) is fail-silent. Explicit `update` returns the typed error for the CLI to print.
- **Tests**: `TestCliSelfUpdate_ExplicitUpdateNetworkError`, `TestCliSelfUpdate_StartupNoticeFailsSilent`.

### EC-10: Atomic binary replacement safety
- **Risk**: If the machine crashes or the process is killed mid-replace, a partial write could leave an unusable binary or a stale temp file on disk.
- **Mitigation**: `replaceBinary` writes to a `.centinela-update-*` temp file in the same directory (same filesystem, avoids EXDEV), fsyncs, copies mode bits, then renames atomically. On any failure the temp file is removed before returning.
- **Tests**: `TestReplaceBinaryAtomic`, `TestReplaceBinaryWriteTempErrorCleansUp`, `TestReplaceBinaryRenameError`, `TestWriteTempSyncError`, `TestWriteTempChmodError`.

### EC-11: --check read-only guarantee
- **Risk**: `--check` must be safe to run in automation without risk of modifying the binary, even if behind.
- **Mitigation**: `Check()` never calls `install()`. It reads (and possibly writes) only the cache file.
- **Tests**: `TestCliSelfUpdate_CheckBehindReturnsMsg`, `TestCliSelfUpdate_CheckCurrentReturnsUpToDate`, `TestRunUpdateCheckBehindExits1`, `TestRunUpdateCheckCurrentExits0`.

### EC-12: Cache path derivation
- **Risk**: Different systems use different XDG_CACHE_HOME configurations. The path must fall back gracefully to `~/.cache` when XDG is not set.
- **Mitigation**: `cachePath()` checks `XDG_CACHE_HOME` first and falls back to `$HOME/.cache/centinela/update-check.json`.
- **Tests**: `TestCachePathHomeFallback`, `TestCacheRoundTripAndStale`.

### EC-13: SHA256SUMS missing from release
- **Risk**: A malformed release might have the binary asset but not the SHA256SUMS file. Without sums there is no way to verify the download.
- **Mitigation**: `install()` checks for SHA256SUMS with the same `assetURL` probe and returns `KindPlatform` if absent.
- **Tests**: `TestInstallMissingSumsAsset`.

### EC-14: Asset SHA256SUMS entry missing for platform
- **Risk**: The SHA256SUMS file exists but has no entry for this platform's asset name. Rather than accepting a nil checksum match, the code must reject the download.
- **Mitigation**: `sumFor()` returns `("", false)` when no matching line is found. `install()` returns `KindChecksum` when ok is false.
- **Tests**: `TestInstallChecksumNoEntry`.

## Test Infrastructure Notes

- All tests use `httptest.Server` — no DNS lookup for `api.github.com` or `github.com` ever occurs.
- `HOME` and `XDG_CACHE_HOME` are isolated via `t.Setenv` in every test.
- The replace target (`Target` func field) is a temp file, never the test binary itself.
- All file operations are bounded to `t.TempDir()` and cleaned up automatically.
