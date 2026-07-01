# `centinela.toml` Reference

> Every configuration key, its type, default, allowed values, and meaning.

`centinela.toml` lives in your project root and is read on every `centinela` command. Edit it directly; changes take effect on the next command. Run [`centinela doctor`](#validating-your-config) to check it for errors.

Looking for a ready-made setup instead of individual keys? See the [Configuration guide](configuration.md) for copy-paste recipes by use case.

**What is *not* in `centinela.toml`:**

- **Harness selection** (Claude / OpenCode / Codex) is decided by `centinela init --agent` and the resulting `.claude/settings.json` / `opencode.json` — not a config key.
- **Architecture archetype** (Hexagonal, Rails, N-Tier, ECS, Modular) is inferred from your codebase and confirmed in `PROJECT.md` — not a config key.
- **Enforcement profile** (`strict` / `guided` / `outcome`) scales process strictness; it does *not* select a different config file.

---

## `[workflow]`

Workflow step validation and confirmation behavior.

| Key | Type | Default | Allowed values | Description |
|-----|------|---------|----------------|-------------|
| `step_confirmation_mode` | string | `every_step` | `every_step`, `after_plan`, `auto` | When Centinela pauses for manual review before advancing a step |
| `plan_advisor_mode` | string | `missing_info` | `off`, `always`, `missing_info` | Adaptive planning prompts during the plan step |
| `plan_question_limit` | int | `4` | `1..4` | Max advisor questions per round |
| `plan_advisor_failure_top_n` | int | `3` | `1..5` | Max recurring-failure list size shown to the advisor |
| `enforcement_profile` | string | `strict` | `strict`, `guided`, `outcome` | Global profile that scales how strictly the 5 steps are enforced |
| `use_worktrees` | bool | `false` | `true`, `false` | Run each feature in its own git worktree under `.worktrees/<feature>/` |
| `test_suffixes` | []string | `[]` (any file in `tests/`) | file extensions | Suffixes identifying unit/integration tests |
| `acceptance_suffix` | string | `""` (any file in `tests/acceptance/`) | file extension | Suffix for acceptance-test step definitions |
| `code_dirs` | []string | `/src/`, `/app/`, `/cmd/`, `/internal/`, `/pkg/`, `/lib/`, `/backend/`, `/frontend/` | path segments | Paths that classify a file as "code" for step enforcement |

## `[validate]`

Commands and scoping for the validate step.

| Key | Type | Default | Allowed values | Description |
|-----|------|---------|----------------|-------------|
| `commands` | []string | `[]` | native argv commands | Commands run sequentially during validate (lint, type-check, tests); each must exit 0 |
| `diff_mode` | string | `auto` | `auto`, `always`, `off` | Whether file-walking gates scope to changed files |
| `diff_base` | string | `main` | any git ref | Merge-base ref that defines the change set |

Commands run natively (no shell) — cross-platform on Windows, macOS, Linux. See [Diff-aware mode](gates.md#diff-aware-mode).

## `[gates]`

Built-in gate toggles.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `file_size` | bool | `true` | G1: fail if a source file exceeds 100 lines (exceptions up to 130) |
| `i18n` | bool | `false` | G11: check translation-key completeness across locales |
| `production_readiness` | bool | `false` | G12: run the production-readiness subagent check |

> If both `file_size` and `i18n` are left unset, `file_size` defaults to `true`.

### `[[gates.file_size_exceptions]]`

Array of tables. Each entry lifts one file's line limit (max 130 exceptions total).

| Key | Type | Required | Allowed values | Description |
|-----|------|----------|----------------|-------------|
| `path` | string | yes | non-empty | File path relative to project root |
| `kind` | string | yes | `configuration`, `domain_atomic` | Category of exception |
| `reason` | string | yes | non-empty | Justification (audited) |
| `max_lines` | int | yes | `101..130` | Max lines allowed for this file |

### `[gates.build]` — G-Build: Cross-Compile

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `false` | Enable the cross-compile gate |
| `command` | string | `go build ./cmd/centinela` | Build command; run once per target with `GOOS`/`GOARCH`/`CGO_ENABLED=0` set |
| `targets` | list of `{goos, goarch}` | `[]` | Cross-compile targets; each needs both `goos` and `goarch` |

### `[gates.import_graph]` — G2: Layer Boundaries

| Key | Type | Default | Allowed values | Description |
|-----|------|---------|----------------|-------------|
| `enabled` | bool | `false` | `true`, `false` | Enable import-graph layer enforcement |
| `provider` | string | `""` (auto-detect) | `go`, `node`, `python`, `script`, `""` | Language backend; empty auto-selects by manifest |
| `module` | string | `""` (auto via `go list -m`) | go module | Module name override (Go provider) |
| `script_command` | []string | `[]` | argv | Custom graph emitter (required when `provider = "script"`) |
| `layers` | list of `{name, paths, allow}` | `[]` | — | Layer rules: `name`, `paths` (globs), `allow` (allowed layer names) |

### `[gates.security]` — G-Secrets / G-Vuln (opt-in)

| Key | Type | Default | Allowed values | Description |
|-----|------|---------|----------------|-------------|
| `enabled` | bool | `false` | `true`, `false` | Enable security scanning |
| `secrets.tool` | string | `gitleaks` (when enabled) | `gitleaks` | Secret scanner (hard-fail) |
| `secrets.allowlist` | []string | `[]` | patterns | Secret patterns to ignore |
| `vuln.tools` | []string | `govulncheck`, `osv-scanner` (when enabled) | `govulncheck`, `osv-scanner` | Dependency vuln scanners (warn-only) |

### `[gates.spec_traceability]`

| Key | Type | Default | Allowed values | Description |
|-----|------|---------|----------------|-------------|
| `enabled` | bool | `false` | `true`, `false` | Enable Gherkin→acceptance-test traceability check |
| `spec_dir` | string | `specs` | dir path | Directory of `.feature` files |
| `test_dir` | string | `tests/acceptance` | dir path | Directory of acceptance step definitions |
| `severity` | string | `fail` | `fail`, `warn` | Whether uncovered scenarios block or warn |

### `[gates.roadmap_drift]`

| Key | Type | Default | Allowed values | Description |
|-----|------|---------|----------------|-------------|
| `enabled` | bool | `false` | `true`, `false` | Fail/warn when `ROADMAP.md` drifts from `.workflow/roadmap.json` |
| `severity` | string | `warn` | `fail`, `warn` | Block merge or warn |

### `[gates.audit_baseline]`

| Key | Type | Default | Allowed values | Description |
|-----|------|---------|----------------|-------------|
| `enabled` | bool | `false` | `true`, `false` | Ratchet: fail only on *new* violations vs a committed baseline |
| `severity` | string | `warn` | `fail`, `warn` | Block or warn on new violations |
| `baseline_path` | string | `.workflow/audit-baseline.json` | file path | Committed baseline snapshot |
| `target_gates` | []string | `[]` (all) | gate names | Restrict which gates participate |

### `[[gates.custom]]`

Array of tables — run any command as a gate.

| Key | Type | Default | Allowed values | Description |
|-----|------|---------|----------------|-------------|
| `name` | string | required | unique, not a built-in name | Gate name in the report |
| `command` | string | required | argv | Command; exit 0 = pass |
| `enabled` | bool | — | `true`, `false` | Whether the gate runs |
| `severity` | string | `fail` | `fail`, `warn` | Block or warn on failure |
| `output` | string | `blob` | `blob`, `lines` | Output rendering |
| `timeout_seconds` | int | `60` | positive | Kill after N seconds |
| `diff_aware` | bool | `false` | `true`, `false` | Respect changed-files-only mode |

Built-in names you cannot shadow: `G1: File Size`, `G11: i18n`, `G-Build: Cross-Compile`, `import_graph`, `G-Secrets: Secret Scan`, `G-Vuln: Dependency Audit`, `spec-traceability-gate`, `roadmap_drift`, `audit_baseline`.

## `[i18n]`

Required when `gates.i18n = true`.

| Key | Type | Allowed values | Description |
|-----|------|----------------|-------------|
| `format` | string | `json`, `gettext`, `none` | `json` = next-intl/i18next/vue-i18n; `gettext` = `.po`; `none` = defer to a custom command |
| `dir` | string | dir path | Directory holding locale files |
| `locales` | []string | locale codes | Expected locales, e.g. `["en", "es", "fr"]` |

## `[verify]`

Claim verification (drives `centinela verify` and the complete-gate check).

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `verify_timeout` | int | `60` | Seconds before a test command is killed during verification |
| `coverage_tolerance` | float | `0.001` | Max gap allowed between a claimed and measured coverage figure (0.001 = 0.1%) |

## `[orchestration]`

AI model routing across subagent roles and runners.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `ui_paths` | []string | `src/ui`, `src/components`, `app/views`, `web`, `styles`, `internal/ui` | Paths identifying UI code |
| `models` | table | `{}` | Per-role tier or per-runner model override (see below) |
| `model_map` | table | `{}` | Tier→runner→model remap |
| `capabilities` | table | `{}` | Model id → capability class override (`frontier`/`capable`/`limited`) |
| `capability_profiles` | table | `{}` | Capability class → enforcement profile remap |
| `driver_model` | string | `""` | Model id keying the workflow's default profile |

### `[orchestration.models.<role>]`

Roles: `big-thinker`, `feature-specialist`, `senior-engineer`, `ux-ui-specialist`, `qa-senior`, `documentation-specialist`, `validation-specialist`, `merge-steward`, `gatekeeper`, `edge-case-tester`.

Two forms:

```toml
[orchestration.models]
big-thinker = "reasoning"          # string form → tier for all runners (reasoning|balanced|fast)

[orchestration.models.senior-engineer]   # table form → per-runner model override
claude   = "claude-opus-4-7"
opencode = "anthropic/claude-opus-4-7"
```

Runners: `claude`, `opencode`, `codex`.

### `[orchestration.model_map.<tier>.<runner>]`

```toml
[orchestration.model_map.balanced]
claude   = "claude-sonnet-4-6"
opencode = "anthropic/claude-sonnet-4-6"
```

### `[orchestration.local]`

Point Centinela at a local model backend. All-or-nothing: if `provider` is set, `endpoint` and `model` are required.

| Key | Type | Default | Allowed values | Description |
|-----|------|---------|----------------|-------------|
| `provider` | string | `""` (disabled) | `ollama`, `openai-compatible` | Local provider |
| `endpoint` | string | `""` | URL | API endpoint (OpenAI-compatible `/v1`) |
| `model` | string | `""` | model name | Opaque model id |
| `api_key_env` | string | `""` | env var name | Env var holding the API key (`openai-compatible` only) |

An unmapped local `model` defaults to the `limited` capability class → `strict` profile. See [Getting Started](getting-started.md#point-centinela-at-a-local-model).

### Capability classes & enforcement profiles

| Class | Built-in models | Default profile |
|-------|-----------------|-----------------|
| `frontier` | `claude-opus-4-7` | `outcome` |
| `capable` | `claude-sonnet-4-6` | `guided` |
| `limited` | `claude-haiku-4-5` | `strict` |

Override a model's class with `[orchestration.capabilities]`, or a class's profile with `[orchestration.capability_profiles]`.

## `[memory]`

Governed project-memory ledger (recall injected into the plan step).

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `true` (opt-out) | Enable the memory ledger |
| `recall_max_entries` | int | `10` | Max entries recalled into the plan step |
| `recall_max_bytes` | int | `4096` | Max total bytes recalled |

## `[telemetry]`

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `true` (opt-out) | Append-only governance event log (JSONL) |

## `[headless]`

Non-interactive mode for CI / fleets.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `false` | Force headless (non-interactive) mode |
| `detect_ci` | bool | `false` | Auto-enable headless when `CI` env is `1`/`true` |

Precedence: `CENTINELA_HEADLESS` env → `enabled` → (`detect_ci` and `CI`).

## `[cost]`

Soft cost-governance gate (never blocks `complete`).

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `false` | Enable cost governance |
| `step_token_budget` | int | `0` (off) | Per-step token budget |
| `feature_token_budget` | int | `0` (off) | Per-feature token budget |
| `tier_budgets` | table | `{}` | Per-model budgets, keyed by model id |

## `[precommit]` and `[pr_gate]`

Advisory surfaces for git-hook and PR-gate documentation (the commands run when invoked regardless of `enabled`).

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `precommit.enabled` | bool | `false` | Surface pre-commit hook install guidance |
| `precommit.skip_build` | bool | `true` | Whether `centinela precommit` skips the heavy build gate |
| `pr_gate.enabled` | bool | `false` | Surface PR-gate documentation |
| `pr_gate.fail_on_warning` | bool | `false` | Escalate warn-severity gates to a non-zero exit in CI |

## Environment variables

| Variable | Effect |
|----------|--------|
| `CENTINELA_HEADLESS` | `1`/`true` forces headless mode (overrides config + CI detection) |
| `CENTINELA_MODEL` | Driver-model override (precedence: `--model` flag > env > config) |
| `CI` | `1`/`true` marks a CI environment (used by headless detection and diff-aware default) |

## Validating your config

```bash
centinela doctor          # checks config syntax + reports errors
centinela doctor --fix    # applies safe, idempotent repairs
```

---

← Back to the [documentation index](README.md) · [Configuration guide (recipes)](configuration.md)
