# codex-support — senior-engineer

## Files Touched

ADDED:
- `internal/setup/codex_config.go` (61) — managed `.codex/config.toml` via the existing `planManagedFile`/`writeManaged` seam; `SyncKindPrewriteHook`; `# centinela:managed-version=` header recognized by the seam, so init→migrate reports no drift. Body fit under 100 lines, so no `codex_config_body.go` split was needed.
- `internal/setup/adapter_codex.go` (24) — `codexAdapter`: `Name()="codex"`, `Capabilities()={CapBlocksWrites, CapPromptContext, CapRulesFile}`, `PlanItems()` returns the codex config item + shared `planAgentsFile()` via `itemSlice`.
- `internal/hookpolicy/applypatch.go` (50) — `ExtractApplyPatchPaths` + `EvaluatePrewriteMulti` (the load-bearing fix).

EDITED:
- `internal/setup/adapter.go` (50) — `"codex"` appended to `orderedAgents`; `registry["codex"]=codexAdapter{}`. `composites["both"]` unchanged.
- `internal/setup/sync.go` (82) — `applyItem` `SyncKindPrewriteHook` case: `codexConfigFile`→`writeManagedCodexConfig` ordered before the `pluginFile` branch, then `InjectHooks` fallback.
- `cmd/centinela/init_agent.go` (67) — extracted `applyManagedSetup(agent, label)`; `setupOpenCode`/`setupAider`/`setupCodex` are now one-line callers; `case "codex"` added to `runHarnessSetup`.
- `cmd/centinela/hook_prewrite.go` (93) — added `Command` to `ToolInput`; `prewriteTargets` resolves file_path/filePath else `ExtractApplyPatchPaths(command)`; calls `evalPrewriteMulti`; `blockPrewrite` preserves NeedInit/out-of-step stderr + telemetry, rendering `d.Path`.
- `cmd/centinela/hook_postwrite.go` (96) — added `Command`; `extractPostwritePath` falls back to first apply_patch path so the status tag classifies under Codex.
- `cmd/centinela/init.go`, `cmd/centinela/migrate_setup.go` — `--agent` help strings now list `codex`.
- `internal/hookpolicy/prewrite.go` (74) — added `Path string` to `PrewriteDecision` (set by `EvaluatePrewriteMulti` so the cmd can render which patched file blocked).

TEST MAINTENANCE (existing only):
- `internal/setup/adapter_registry_test.go` — roster `{claude,opencode,aider,codex}`, adapter count 3→4, `AgentsFor("both")` still `{claude,opencode}`.
- `cmd/centinela/hook_prewrite_block_test.go` — stub switched from `evalPrewrite` to `evalPrewriteMulti` (the new cmd seam).

## Architecture Compliance

- All touched files ≤100 lines (G1). Largest: hook_postwrite.go 96.
- G7 (no business logic in outer layer): path extraction + multi-path policy live in `internal/hookpolicy` (domain). `cmd/centinela` stays a thin caller; `applyManagedSetup` is presentation glue only.
- Capability-parity invariant satisfied: codex declares `blocks-writes` and emits a `SyncKindPrewriteHook` item (the `.codex/config.toml`). `adapter_parity_test.go` auto-covers it — green.
- Reused existing seams only (`planManagedFile`, `writeManaged`, `itemSlice`, `planAgentsFile`); no new SyncKind.

## Type-Safety Notes

No dynamic typing. `ExtractApplyPatchPaths` returns `[]string`; `EvaluatePrewriteMulti` returns the existing typed `PrewriteDecision`. New JSON field `Command string` is statically typed on the input structs. `Path string` added to `PrewriteDecision` is additive and zero-valued for existing single-path callers.

## Exact Codex TOML schema used

Nested matcher groups (confirmed Codex hooks schema): `[[hooks.<Event>]]` carries `matcher`, with a nested `[[hooks.<Event>.hooks]]` array of `{type="command", command="..."}`. `command` is a single shell-run string. Canonical file-write tool = `apply_patch`.
- `PreToolUse` matcher `apply_patch` → `centinela hook prewrite`
- `PostToolUse` matcher `apply_patch` → `centinela hook postwrite`
- One `[[hooks.UserPromptSubmit]]` group with 6 `[[hooks.UserPromptSubmit.hooks]]` blocks in the OpenCode plugin's prompt-chain order: setup, migrate, autostart, orchestration, plan-advisor, context.

## The apply_patch fix (and why)

Codex's `apply_patch` sends `tool_input.command` (a patch-envelope STRING), not `file_path`/`filePath`. The old prewrite hook read only those two keys → empty path → returned nil → NEVER blocked under Codex. A declared `blocks-writes` capability that doesn't block is a correctness bug. `ExtractApplyPatchPaths` scans envelope lines `*** Add File:`, `*** Update File:`, `*** Delete File:`, `*** Move to:` and returns every trimmed path (multi-file aware; nil if none). `EvaluatePrewriteMulti` evaluates each path through the unchanged `EvaluatePrewrite` and returns the FIRST non-Allow decision (with `.Path` set), else Allow. Dogfooded with a /tmp binary: Add File, Move to, and a multi-file patch (allowed README.md skipped, `.go` blocked) all exit 2; greenfield init→migrate reports "already up to date".

## Trade-Offs / deviations from plan

- Plan's TOML sketch used `matcher = "Write|Edit|Patch"` and array-form `command`; replaced with the confirmed Codex schema (matcher `apply_patch`, nested `[[hooks.<Event>.hooks]]`, string `command`) per the code-step instruction. This is the deliberate schema-confirmation the plan flagged as the top risk.
- Added a `Path` field to `PrewriteDecision` so the multi-evaluator reports the offending path to the thin cmd layer (keeps rendering in cmd, policy in domain). Minimal, additive, no behavior change for single-path callers.
- Body fit ≤100 lines, so the optional `codex_config_body.go` split was not needed.
- Backward compatibility 100% preserved: when file_path/filePath is present, `prewriteTargets` returns `[that path]` (verified: still exits 2). No-path input is a no-op (exit 0).

## Relative-path bug + fix (coordinator-flagged, load-bearing)

A second correctness bug was found after the initial fix: Codex apply_patch envelopes carry REPO-RELATIVE paths (`*** Add File: internal/foo.go`), but `EvaluatePrewriteMulti` passed them straight to `EvaluatePrewrite`. `isInsideWorkspace` then ran `filepath.Rel(absCwd, "internal/foo.go")`, which errors on a relative second arg → returns false → Allow → NEVER BLOCKED. So `blocks-writes` was non-functional for real Codex traffic (the earlier dogfood only used absolute paths, which masked it).

Fix in `internal/hookpolicy/applypatch.go` `EvaluatePrewriteMulti`: resolve each relative path against cwd before evaluating, but report the original path for rendering. Added `path/filepath` import. File now 58 lines (≤100). Claude/OpenCode send absolute `file_path` → `filepath.IsAbs` true → unchanged. `*** Move to:` targets are relative too, so they are covered.

```
abs := path
if cwd != "" && !filepath.IsAbs(abs) {
    abs = filepath.Join(cwd, abs)
}
d := EvaluatePrewrite(abs, cwd, cfg, wfs)
if !d.Allow {
    d.Path = path // original (relative) path for rendering
    return d
}
```

Re-dogfooded (temp dir with project's centinela.toml + git init + NO workflow, /tmp binary):
- RELATIVE apply_patch code path (`internal/foo.go`) → exit 2 (was 0 before fix)
- RELATIVE apply_patch docs path (`docs/notes.md`) → exit 0 (allowed)
- legacy ABSOLUTE `file_path` code → exit 2 (backward compat intact)

CRITICAL for qa-senior: the regression test MUST use a RELATIVE apply_patch path that blocks. An absolute-only test passes even with the bug present and would have masked this. Add a relative-path case to both the `EvaluatePrewriteMulti` unit test and the cmd-level prewrite test.

## Residual limitation (flag for qa-senior)

UserPromptSubmit stdin shape under Codex is unverified end-to-end. `centinela hook autostart`/`context` read stdin (the prompt JSON); Codex runs `command` through a shell and the exact stdin payload it pipes to UserPromptSubmit hooks isn't confirmed here. If Codex passes no/empty stdin, the prompt chain degrades gracefully (autostart no-ops, context still emits workflow tags) — it does not error or block. The prewrite/postwrite blocking path (the load-bearing surface) IS verified.

## Handoff — qa-senior

Author the new test suite (do not duplicate the existing-test maintenance above). Cover:
- `ExtractApplyPatchPaths`: Add/Update/Delete File, `Move to:`, multi-file (multiple paths), none→nil, leading/trailing whitespace, non-patch text.
- `EvaluatePrewriteMulti`: first-blocking-wins, all-allowed→Allow, empty→Allow, `.Path` set to the blocking path.
- `cmd` prewrite/postwrite: apply_patch command JSON resolves paths; backward-compat file_path/filePath; no-path no-op.
- Golden fixtures `internal/setup/testdata/golden/codex/{.codex/config.toml,AGENTS.md}` + add `"codex"` to `golden_parity_test.go` cases.
- Colocated unit tests `adapter_codex_test.go` + `codex_config_test.go` (create/update/manual-review, managed-version header) — keep coverage ≥97% per-package; each `_test.go` ≤100 lines.
- Acceptance: codex init→migrate idempotency (`BuildSyncPlan("codex")` + `ApplySync`, second plan `!HasChanges()`); unmanaged `.codex/config.toml` → manual-review (not clobbered). Wire acceptance execution into `validate.commands`; author `.workflow/codex-support-edge-cases.md`.
