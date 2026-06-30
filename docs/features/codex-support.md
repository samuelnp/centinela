# Feature Brief: codex-support

> Phase 11 (Ecosystem). Depends on: `host-harness-adapters`.

## What

First-class support for **OpenAI Codex** as a host harness alongside Claude Code
and OpenCode. Codex plugs into the existing `HarnessAdapter` interface +
registry (`internal/setup`) so `centinela init --agent codex` and
`centinela migrate setup --agent codex` wire Codex's integration surface to
Centinela's governance: prewrite enforcement (blocking out-of-order writes),
postwrite status tags, and prompt-context injection, plus the Codex-native
rules file (`AGENTS.md`).

## Why

Codex has a large and growing user base. Today Codex users are locked out of
Centinela because enforcement is wired only for Claude Code / OpenCode.
Dual-support with Claude Code materially widens who can adopt Centinela.

## Integration-surface findings (from OpenAI Codex docs, June 2026)

Codex is a **full first-class harness** — its hook system mirrors Claude Code's:

- **Lifecycle hooks** (`PreToolUse`, `PostToolUse`, `UserPromptSubmit`,
  `SessionStart`, `Stop`, …) are declared as inline `[hooks]` tables in
  `config.toml` **or** in a `hooks.json` file. Command hooks are supported
  (prompt/agent handlers are parsed but skipped).
- **Project-local discovery**: Codex loads repo-local hooks from
  `<repo>/.codex/hooks.json` or `<repo>/.codex/config.toml` (only when the
  project `.codex/` layer is trusted). Layers coexist — project hooks do not
  replace global ones.
- **Hook I/O**: hooks receive a JSON object on **stdin** (`tool_name`,
  `tool_input`, `hook_event_name`, `cwd`, `session_id`, …). A `PreToolUse`
  hook **blocks** via exit code `2` (reason on stderr) or by emitting a JSON
  `permissionDecision: "deny"`. `UserPromptSubmit` injects context via
  `additionalContext` JSON or plain stdout text.
- **Rules file**: Codex natively reads `AGENTS.md` (checked per directory:
  `AGENTS.override.md`, `AGENTS.md`, `TEAM_GUIDE.md`, `.agents.md`) — the same
  surface OpenCode and Aider already reuse.

These map cleanly onto the existing `centinela hook prewrite|postwrite|setup|
migrate|autostart|orchestration|plan-advisor|context` commands, which already
speak stdin-JSON / exit-2 blocking — exactly Codex's contract.

## Capabilities (tiered model from host-harness-adapters)

Codex declares **all three**, like Claude Code and OpenCode:

- `blocks-writes` — `PreToolUse` command hook → `centinela hook prewrite`
  (exit 2 blocks). Satisfies the parity invariant: a `blocks-writes` adapter
  MUST emit a `prewrite-hook` SyncItem.
- `prompt-context` — `UserPromptSubmit` hook injects setup/migrate/autostart/
  orchestration/plan-advisor/context output.
- `rules-file` — reuse the managed `AGENTS.md` surface.

## Acceptance (high level)

1. `codexAdapter` registered in `orderedAgents`; `Lookup("codex")` resolves it;
   `centinela init --agent codex` and `centinela migrate setup --agent codex`
   both succeed end-to-end.
2. Init writes Codex's managed hook surface (a fully-managed file carrying the
   `centinela:managed-version` header) wiring PreToolUse→prewrite,
   PostToolUse→postwrite, UserPromptSubmit→context injection, **plus**
   `AGENTS.md`.
3. Capability + golden parity tests extended to cover `codex` so the three
   harnesses stay byte-for-byte in lockstep through `init` and `migrate`;
   greenfield init→migrate reports **no** pending drift (managed header
   present).
4. All gates green: ≤100-line files, import-graph layers, ≥95% coverage
   (aim ≥97%), fmt.

## Out of scope (v1)

- Global `~/.codex/config.toml` wiring (project-local `.codex/` only).
- Codex subagent/model routing (`internal/orchestration` runner enum is
  separate — see host-harness-adapters).
- The `both` composite stays `{claude, opencode}`; Codex is addressable via
  `--agent codex` (a wider composite can come later if wanted).
