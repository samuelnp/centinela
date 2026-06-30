### Feature-Specialist Report: codex-support
**Date:** 2026-06-30

#### Behavior Summary
OpenAI Codex becomes a full first-class HarnessAdapter alongside Claude Code and OpenCode. Registering `codexAdapter` in `orderedAgents`/`registry` (and adding a `"codex"` dispatch case in `runHarnessSetup`) lets `centinela init --agent codex` write two managed files: a project-local `.codex/config.toml` carrying the `centinela:managed-version` header and wiring three lifecycle hooks (PreToolUse→prewrite blocking via exit 2, PostToolUse→postwrite, UserPromptSubmit→context-injection chain), plus the shared managed `AGENTS.md`. The managed-version header is recognized by the existing `planManagedFile` seam, so a subsequent `migrate setup --agent codex` sees zero pending drift. Pre-existing unmanaged `.codex/config.toml` files are routed to manual-review by the same seam and never overwritten. The `both` composite (`{claude, opencode}`) is not changed; codex is single-select only for v1.

#### Gherkin Scenarios
All scenarios live in `specs/codex-support.feature`.

- **Codex is a valid --agent selector** — Given the registry, When Lookup("codex"), Then codex adapter returned without error and Name()="codex".
- **both composite is unchanged by codex addition** — When BuildSyncPlan("both"), Then no `.codex/config.toml` item; plan identical to claude+opencode union.
- **Codex adapter declares all three capabilities** — Capabilities() returns {blocks-writes, prompt-context, rules-file}.
- **Codex adapter satisfies prewrite-hook parity invariant** — PlanItems() includes SyncKindPrewriteHook item at path ".codex/config.toml".
- **centinela init --agent codex writes managed .codex/config.toml** — .codex/config.toml created with managed header, PreToolUse/PostToolUse/UserPromptSubmit hooks, plus AGENTS.md; .claude/settings.json untouched.
- **init then migrate setup reports no pending drift** — after init, BuildSyncPlan("codex").HasChanges() is false.
- **centinela init --agent codex is idempotent on re-run** — second init leaves files unchanged, exit 0.
- **Pre-existing unmanaged .codex/config.toml is not clobbered** — manual-review warning surfaced, file not overwritten.
- **Codex managed output matches golden fixture byte-for-byte** — emitted files match testdata/golden/codex/ fixtures exactly.

#### UX States
| State | CLI stdout / behaviour |
|-------|------------------------|
| loading | n/a (synchronous CLI) |
| empty (greenfield) | `created: .codex/config.toml`, `created: AGENTS.md`, exit 0 |
| already managed (idempotent) | "already up to date" or equivalent, exit 0 |
| unmanaged file conflict | `manual-review: .codex/config.toml — file exists without managed marker`, exit 0 (warning, not error) |
| unknown agent flag | `invalid agent "codex"...` — n/a, this scenario is the happy path; unknown string → `invalidAgentError` listing registered harnesses, exit 1 |

#### Out-of-Scope
- Global `~/.codex/config.toml` wiring (project-local `.codex/` only).
- Widening `composites["both"]` to include codex.
- `internal/orchestration` runner enum changes (Codex runner already exists there as a separate concern).
- Editing `agentsContent` or bumping `setupDocVersion`.
- Codex subagent/model routing.

#### Deferred Findings
None. The `agents-md-canonical-surface` roadmap item (Phase 11) already covers the "## OpenCode Integration" heading generalization; no new gaps surfaced.

#### Handoff
Next role: **senior-engineer**
