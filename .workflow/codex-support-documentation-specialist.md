# codex-support — documentation-specialist

## What Changed for Users

### New `--agent codex` selector

`centinela init --agent codex` and `centinela migrate setup --agent codex` are now valid. Codex joins Claude, OpenCode, and Aider as a fully first-class harness with all three capabilities: `blocks-writes`, `prompt-context`, and `rules-file`.

### Files Written by `init --agent codex`

- **`.codex/config.toml`** — fully managed by Centinela (carries `# centinela:managed-version=` header). Wires three hook events:
  - `PreToolUse` (`apply_patch` matcher) → `centinela hook prewrite` — blocks writes outside the active workflow step.
  - `PostToolUse` (`apply_patch` matcher) → `centinela hook postwrite` — records the written path for workflow tracking.
  - `UserPromptSubmit` — six hook entries in the same chain order as OpenCode: setup, migrate, autostart, orchestration, plan-advisor, context.
- **`AGENTS.md`** — the shared managed rules file (identical to Claude/OpenCode/Aider).

Running `migrate setup --agent codex` a second time reports "already up to date" — no drift, because the managed-version header is recognized by the existing seam.

### apply_patch Prewrite Blocking

Codex uses `apply_patch` (not `Write`/`Edit`) as its file-write tool. The prewrite hook now:
1. Reads `tool_input.command` (the patch envelope string) when `file_path`/`filePath` are absent.
2. Extracts all target paths from `*** Add File:`, `*** Update File:`, `*** Delete File:`, `*** Move to:` lines.
3. Resolves repo-relative paths against cwd before policy evaluation (relative paths were the key correctness fix).
4. Returns the first non-Allow decision (exit 2) with the blocking path name; allows (exit 0) if all paths pass.

Multi-file patches are handled: each path is evaluated independently; the first blocked path short-circuits.

## Where It Is Documented

- `docs/plans/codex-support.md` — design decisions and gate checklist.
- `specs/codex-support.feature` — Gherkin acceptance criteria.
- `internal/setup/codex_config.go` — TOML body and managed-file seam usage.
- `internal/hookpolicy/applypatch.go` — `ExtractApplyPatchPaths` + `EvaluatePrewriteMulti` public API.
- `docs/project-docs/index.html` — generated HTML docs (this step).

## Residual Note

`UserPromptSubmit` stdin shape under Codex is unverified end-to-end. The exact JSON payload Codex pipes to UserPromptSubmit hook commands is not confirmed. If Codex passes no or empty stdin, the prompt chain degrades gracefully (autostart no-ops, context still emits workflow tags) — it does not error or block. The load-bearing prewrite/postwrite blocking surface is fully verified via dogfood and acceptance tests.

## Outcome

`centinela init --agent codex` is fully operational. The `.codex/config.toml` managed file, apply_patch-aware prewrite blocking (including relative-path resolution), and AGENTS.md are all wired and tested. Coverage gates (≥97%) and gatekeeper SAFE status confirmed by the validation-specialist.
