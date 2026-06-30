# Implementation Plan: codex-support

> Phase 11 (Ecosystem). Extends the `HarnessAdapter` registry from
> `host-harness-adapters`. Adds OpenAI Codex as a **full first-class** harness
> (all three capabilities), addressable via `--agent codex`.

## Goal

`centinela init --agent codex` and `centinela migrate setup --agent codex` wire
Codex's project-local integration surface to Centinela governance: a
fully-managed `.codex/config.toml` (PreToolUse→prewrite blocking, PostToolUse→
postwrite, UserPromptSubmit→context-injection chain) plus the reused managed
`AGENTS.md`. Greenfield init→migrate must report **no** pending drift.

## Design (lowest-blast-radius, mirrors aider + opencode)

Codex declares `{blocks-writes, prompt-context, rules-file}` — same as Claude
and OpenCode. It satisfies the capability-parity invariant (a `blocks-writes`
adapter MUST emit a `SyncKindPrewriteHook` item) by making the single managed
`.codex/config.toml` that prewrite-hook surface.

### Files to ADD

1. **`internal/setup/adapter_codex.go`** (`codexAdapter`):
   - `Name() string` → `"codex"`.
   - `Capabilities()` → `{CapBlocksWrites, CapPromptContext, CapRulesFile}`.
   - `PlanItems()` → `itemSlice(cfg, agents)` where `cfg, _ = planCodexConfig()`
     and `agents, _ = planAgentsFile()` (the EXISTING shared AGENTS.md seam).
     Mirror `aiderAdapter`/`openCodeAdapter` error handling exactly.

2. **`internal/setup/codex_config.go`** — fully-managed `.codex/config.toml`
   authored through the EXISTING `planManagedFile(...)` seam (exact mirror of
   `aider_config.go`):
   ```go
   const codexConfigFile = ".codex/config.toml"
   const codexConfigBody = `...TOML body (see sketch)...`
   const codexConfigHeader = "# centinela:managed-version=" + setupDocVersion +
       " template=.codex/config.toml"

   func planCodexConfig() (*SyncItem, error) {
       return planManagedFile(codexConfigFile,
           codexConfigHeader+"\n"+codexConfigBody, codexConfigBody,
           SyncKindPrewriteHook)
   }
   func writeManagedCodexConfig(path string) error {
       return writeManaged(path, codexConfigHeader+"\n"+codexConfigBody)
   }
   ```
   - Kind = `SyncKindPrewriteHook` (NOT a new kind — reuse, per parity invariant).
   - The `# centinela:managed-version=` prefix is ALREADY recognized by
     `planManagedFile` (sync_managed_files.go:34), so migrate idempotency works
     with zero change to the seam. This is the load-bearing reason init→migrate
     reports no drift.

   **TOML body sketch** (final hook-table schema to be confirmed against Codex
   docs during the code step — see Risks):
   ```toml
   [[hooks.PreToolUse]]
   matcher = "Write|Edit|Patch"
   command = ["centinela", "hook", "prewrite"]

   [[hooks.PostToolUse]]
   matcher = "Write|Edit|Patch"
   command = ["centinela", "hook", "postwrite"]

   [[hooks.UserPromptSubmit]]
   command = ["centinela", "hook", "setup"]
   # ...one table per chain stage: migrate, autostart, orchestration,
   #    plan-advisor, context, merge (same order/commands as hooks.go)
   ```
   Keep the file ≤100 lines; if the literal body exceeds that, split the body
   constant into a second `codex_config_body.go` file (a `const` literal is the
   only content, like `opencode_plugin.go`'s `pluginContent`).

### Files to EDIT

3. **`internal/setup/sync.go`** — `applyItem`, `SyncKindPrewriteHook` case:
   add `if it.Path == codexConfigFile { return writeManagedCodexConfig(it.Path) }`
   BEFORE the existing `pluginFile` check / `InjectHooks` fallback (mirror the
   `pluginFile` branch). Order: pluginFile → codexConfigFile → InjectHooks.

4. **`internal/setup/adapter.go`** — add `"codex"` to `orderedAgents` and
   `registry["codex"] = codexAdapter{}`. Leave `composites["both"]` =
   `{claude, opencode}` UNCHANGED (codex is single-select only for v1).

5. **`cmd/centinela/init_agent.go`** — add `case "codex": return setupCodex()`
   to `runHarnessSetup`. To stay ≤100 lines AND DRY (G7), **extract a shared
   `applyManagedSetup(agent, label string) error`** helper that does the
   BuildSyncPlan → manual-review warnings → ApplySync → success-render flow, and
   collapse `setupOpenCode`/`setupAider`/`setupCodex` to one-line callers.
   (init_agent.go is 84 lines today; a third ~19-line copy would breach 100, so
   this extraction is REQUIRED, not cosmetic.)

6. **`cmd/centinela/init.go`** (line 31) and **`cmd/centinela/migrate_setup.go`**
   (line 24) — update the `--agent` flag help strings to include `codex`
   (`"claude, opencode, aider, codex, or both"`). `invalidAgentError` already
   reads `RegisteredAgents()`, so the error message auto-updates.

### Files to LEAVE UNTOUCHED (explicit non-edits)

- `internal/setup/opencode_agents.go` `agentsContent` — do NOT change. Editing
  it bumps `setupDocVersion`, ripples every golden fixture, and forces every
  existing project to re-migrate. The "## OpenCode Integration" heading is a
  known cosmetic gap owned by Phase-11 `agents-md-canonical-surface`.
- `internal/orchestration` runner enum — Codex already exists there as a
  separate model-routing concern; not part of this feature.

## Test Plan

- **Golden parity** (`internal/setup/golden_parity_test.go`): add
  `"codex": {".codex/config.toml", "AGENTS.md"}` to the `cases` map and add
  fixtures under `internal/setup/testdata/golden/codex/.codex/config.toml` and
  `.../codex/AGENTS.md`. (Fixtures live under `internal/setup/testdata/golden/`,
  not repo-root.)
- **Capability parity**: auto-covers codex once registered
  (`adapter_parity_test.go` iterates `RegisteredAdapters()` and asserts every
  `blocks-writes` adapter emits a `SyncKindPrewriteHook` item).
- **Colocated unit `_test.go`** (each ≤100 lines, incl. test files per G1):
  - `adapter_codex_test.go`: `codexAdapter.Name/Capabilities` (mirror
    `adapter_capabilities_test.go`) + `PlanItems()` returns the config item with
    `Kind==SyncKindPrewriteHook` and the AGENTS.md item.
  - `codex_config_test.go`: `planCodexConfig` create/update/manual-review +
    managed-version header present (mirror `aider_config_test.go`).
  - applyItem codex branch: assert `writeManagedCodexConfig` is reached for
    `.codex/config.toml` (extend `sync_test.go`/`sync_more_test.go`).
  These hold coverage ≥95% (aim ≥97%); `tests/` tier files do NOT move the
  per-package gate, so coverage must be colocated in `internal/setup`.
- **Acceptance** (`tests/acceptance/`): add a codex init→migrate idempotency
  test mirroring `host_harness_adapters_ac2_opencode_test.go` /
  `fix_init_managed_sync_drift_test.go`: BuildSyncPlan("codex") + ApplySync, then
  a second BuildSyncPlan("codex") reports `!HasChanges()` (no pending drift).
- `centinela validate` must include the acceptance execution in `validate.commands`.
- `.workflow/codex-support-edge-cases.md` authored in the tests step.

## Gate Checklist (must all pass before validate ships)

- [ ] All new/edited source AND `_test.go` files ≤100 lines (G1).
- [ ] No cross-layer import violations: `internal/setup` = Infrastructure;
      `cmd/centinela` = thin outer layer (G7) — `applyManagedSetup` is
      presentation glue only, no business logic.
- [ ] `centinela validate` passes (fmt + lint + import_graph + full suite).
- [ ] Coverage ≥95% (aim ≥97%) on `internal/setup` and `cmd/centinela`.
- [ ] Greenfield `init --agent codex` then `migrate setup --agent codex` →
      "already up to date" (no pending drift; managed-version header present).
- [ ] Golden + capability parity green for all three harnesses byte-for-byte.
- [ ] No hardcoded user-facing strings outside i18n where applicable.
- [ ] Gatekeeper report SAFE/WARNING.

## Rollout (smallest correct slice first)

1. `codex_config.go` + `adapter_codex.go` + register in `adapter.go` + the
   `applyItem` branch — the pure `internal/setup` core (parity tests turn green).
2. `cmd/` wiring: extract `applyManagedSetup`, add `setupCodex`, dispatch case,
   flag help strings.
3. Golden fixtures + colocated unit tests + acceptance idempotency test.

Each slice keeps existing claude/opencode/aider behavior byte-for-byte unchanged.
