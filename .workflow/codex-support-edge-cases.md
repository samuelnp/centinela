# Edge Cases: codex-support

## Covered

- **Relative apply_patch code path blocks** (the load-bearing regression): a
  repo-relative `*** Add File: internal/foo.go` with no active workflow exits 2.
  Guarded at the domain level (`EvaluatePrewriteMulti`), cmd level
  (`runHookPrewrite`), and end-to-end (acceptance, piped stdin to the binary).
  An absolute-only test passes even with the bug present, so a relative case is
  asserted everywhere.
- **Relative apply_patch docs path is allowed**: `*** Add File: docs/notes.md`
  classifies as TypeOther → Allow (exit 0). Confirms the block is path-aware,
  not blanket.
- **Absolute path under cwd still blocks**: backward-compat for Claude/OpenCode
  `file_path`/`filePath` (IsAbs short-circuit) — `EvaluatePrewriteMulti` with an
  absolute code path blocks.
- **Multi-file patch envelope**: `ExtractApplyPatchPaths` returns every touched
  path (Add/Update/Delete/Move, and repeated Add File verbs).
- **First-blocking-path-wins**: a mixed envelope (allowed docs then blocked
  code) returns the first non-Allow decision with `.Path` = the offending path.
- **`*** Move to:` target**: parsed and (being relative) resolved against cwd.
- **No-path no-op**: an envelope with no patch verbs, empty input, or empty
  path string yields nil paths → Allow (exit 0); the hook never breaks the host.
- **Whitespace trimming**: leading/trailing whitespace around the verb and path
  is trimmed; an empty path after the prefix is skipped.
- **Unmanaged `.codex/config.toml` not clobbered**: a hand-written file with no
  managed header → SyncManualReview; init surfaces a manual-review warning and
  leaves the bytes untouched (unit + acceptance).
- **Managed lifecycle**: absent → SyncCreate; managed (header present) →
  SyncUpdate; emitted file begins with `# centinela:managed-version=`.
- **Init idempotency / no drift**: fresh `init --agent codex` then
  `migrate setup --agent codex` reports no `create:`/`update:`.
- **Golden byte-parity**: `.codex/config.toml` + `AGENTS.md` match fixtures
  byte-for-byte via BuildSyncPlan("codex")+ApplySync.
- **Scope isolation**: `init --agent codex` does not create `.claude/settings.json`.

## Residual Risks

- **UserPromptSubmit stdin shape under Codex is unverified end-to-end** (flagged
  by the senior). The prompt chain (autostart/context) reads stdin; Codex's exact
  payload to UserPromptSubmit hooks is unconfirmed. Mitigation: the chain degrades
  gracefully (no error/block) on empty stdin; the load-bearing prewrite/postwrite
  blocking surface IS fully verified here. Deferred to a follow-up once a live
  Codex stdin capture is available.
