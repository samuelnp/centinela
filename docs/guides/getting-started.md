# Getting Started

> Full setup walkthrough — from `centinela init` to your first shipped feature.

For the quick version, see the [Quickstart in the README](../../README.md#quickstart). This guide covers the complete flow.

## 1. Initialize a project

Run once in your project root:

```bash
centinela init
```

This creates:

| File / Directory | Purpose |
|-----------------|---------|
| `CLAUDE.md` | Framework rules — used by Claude, and by OpenCode via compatibility mode |
| `PROJECT.md.template` | Fill in and rename to `PROJECT.md` |
| `centinela.toml` | Configure validate commands and gate checks |
| `docs/architecture/` | architecture reference documents + edge-case tester prompt |
| `docs/plans/` `specs/` `tests/` | Required empty directories |
| `.claude/settings.json` | Claude hooks wired automatically |
| `opencode.json` + `.opencode/plugins/centinela.js` | OpenCode integration wiring |

Safe to re-run — existing files are never overwritten.

`init` is the bootstrap command. For ongoing upgrades to managed docs and setup
artifacts, use `migrate` (preview by default, apply only with `--apply`).

Use `--local` to write Claude hooks to `.claude/settings.local.json` (useful for personal overrides without committing to the repo):

```bash
centinela init --local
```

Choose integration target explicitly when needed:

```bash
centinela init --agent claude
centinela init --agent opencode
centinela init --agent both   # default
```

> **Harness selection happens here**, not in `centinela.toml`. The `--agent` flag
> (and the resulting `.claude/settings.json` / `opencode.json`) decides which
> harness Centinela governs. `centinela.toml` configures *what* is enforced, not
> *which agent* runs.

## 2. Fill in PROJECT.md

Open your coding agent in the project — if `PROJECT.md` is missing, Centinela will prompt Claude or OpenCode to interview you and write it. Alternatively, rename `PROJECT.md.template` to `PROJECT.md` and complete every section manually. This file tells both you and the agent what the project is, which architecture pattern it follows, and where everything lives.

The **architecture archetype** is chosen here (Hexagonal, Rails-native, N-Tier, ECS, or Modular) — see [Architecture Archetypes](../architecture/architecture-overview.md). It is inferred from your codebase and confirmed in `PROJECT.md`; it is not a `centinela.toml` key.

## 3. Configure centinela.toml

Add your stack's lint and test commands:

```toml
[validate]
commands = [
  "./scripts/check-coverage.sh",
]
```

Default coverage threshold is `95.0%`. Override temporarily with:

```bash
MIN_COVERAGE=96.5 ./scripts/check-coverage.sh
```

Commands run natively via the OS. No shell scripts, no bash dependency — works on Windows, macOS, and Linux.

**For a setup tuned to your situation** — solo prototyping, a disciplined team with CI, offline/local models, a regulated codebase, or a CI fleet — copy a ready-made `centinela.toml` from the [Configuration guide](configuration.md). Every knob is documented in the [Configuration reference](configuration-reference.md).

### Point Centinela at a local model

Declare a local backend with an `[orchestration.local]` block and `centinela init` / `centinela migrate` will wire the OpenCode provider for you — no hand-editing `opencode.json`. The block has four fields:

| Field | Required | Meaning |
|-------|----------|---------|
| `provider` | yes | `ollama` or `openai-compatible` |
| `endpoint` | yes | Base URL of the local server (OpenAI-compatible `/v1` path) |
| `model` | yes | Opaque model id passed through to the runner |
| `api_key_env` | no | Env var name whose value the runner reads as the API key (`openai-compatible` only) |

Both kinds drive an OpenAI-compatible endpoint through the npm `@ai-sdk/openai-compatible` provider; they differ only in whether an API-key reference is written.

Pick **one** of the two forms:

```toml
# Ollama running locally — no API key needed.
[orchestration.local]
provider = "ollama"
endpoint = "http://localhost:11434/v1"
model    = "qwen2.5-coder"
```

```toml
# Any other OpenAI-compatible server (llama.cpp, vLLM, LM Studio, ...).
# Use api_key_env when the server expects an Authorization header.
[orchestration.local]
provider    = "openai-compatible"
endpoint    = "http://localhost:8080/v1"
model       = "my-local-model"
api_key_env = "LOCAL_API_KEY"   # runner reads the value of this env var
```

When a local block is present, `centinela init`/`migrate` add a **managed** provider block to `opencode.json` (`options.baseURL` from `endpoint`; `options.apiKey = "{env:NAME}"` for `openai-compatible` with `api_key_env`; the model under `models`). Centinela owns only its own provider key, never clobbering a user-written or foreign provider, and the wiring is idempotent — re-running rewrites it only on a real change. A config with no `[orchestration.local]` block produces byte-for-byte the same managed output as before.

A declared local `model` with no explicit capability class defaults to the `limited` capability → `strict` profile, as the strictly-lowest precedence tier — so you get maximum scaffolding just by declaring an endpoint. An explicit `--profile` or a global `[workflow] enforcement_profile` still wins. `centinela status` shows the provenance:

```
Profile  strict  (local default: qwen2.5-coder → limited → strict)
```

Centinela validates only the *shape* of the block (known `provider`, all-or-nothing, non-empty `endpoint` + `model`). The `endpoint`, `model`, and `api_key_env` strings are opaque — Centinela never connects to the server, verifies the model exists, or resolves the env var. Availability is the runner's job.

## Example: bootstrap a new project

Use this flow when starting from scratch.

```bash
centinela init
```

Then open your coding agent in the repo and follow setup prompts.

Expected sequence:

1. If `PROJECT.md` is missing, centinela asks the agent to interview you and write it.
2. Once `PROJECT.md` exists, centinela asks the agent to define your roadmap.
3. The agent produces:
   `ROADMAP.md` (human-readable phased plan), `.workflow/roadmap.json` (machine-readable roadmap), `.workflow/roadmap-analysis.md`, `.workflow/roadmap-analysis.json`, `.workflow/roadmap-quality.md`, `.workflow/roadmap-quality.json`, and `docs/features/<feature-slug>.md` for the initial feature set.

If the agent misses one of those files, use `docs/architecture/artifact-templates.md` for the exact setup and per-feature workflow artifact shapes.

Verify roadmap status:

```bash
centinela roadmap
centinela roadmap validate
```

Start implementation from the first roadmap feature:

```bash
centinela start <first-feature-slug>
```

You can also use natural language instead of typing commands directly. The agent
maps your intent to centinela commands under the hood.

Roadmap and status intent examples:

- `Check roadmap`
- `Show roadmap progress`
- `What feature should we build next?`
- `What step are we currently in?`

Start feature intent examples:

- `Implement first feature`
- `Start the next roadmap feature`
- `Begin feature: user-auth`
- `Kick off checkout-flow`

Continue feature intent examples:

- `Continue current feature`
- `Resume work on billing-retries`
- `Complete this step`
- `Move to the next step for onboarding-wizard`

## 4. Start building

```bash
centinela start my-feature
```

For standard product features, follow the same path every time:

1. Write the plan artifacts in `docs/plans/` and `specs/`.
2. Run `centinela complete <feature>` to advance to `code`.
3. Implement the change, then advance to `tests`.
4. Add unit, integration, acceptance, and edge-case coverage, then advance to `validate`.
5. Run `centinela validate`, resolve any gate failures, then advance to `docs`.
6. Run `centinela docs validate` and `centinela docs generate` to publish the project-facing HTML output.

For a complete agent-collaboration example, see [`HOWTO.md`](../../HOWTO.md). It walks through using Centinela to generate a small landing page MVP without skipping the required workflow steps.

Proper use checklist:

- Start or resume a named feature before editing files: `centinela start <feature>` or `centinela status <feature>`.
- Keep all feature work inside the current step; if a write is blocked, create the missing artifact or advance the workflow instead of forcing the edit.
- Treat `complete` prompts as review gates. Approve advancement only after the current step artifacts exist and match the plan.
- Put acceptance tests in `tests/acceptance/` and ensure `[validate].commands` runs them.
- Finish with `centinela validate`, `centinela docs validate`, and `centinela docs generate --out docs/project-docs/index.html`.

## 5. Migrate managed assets

Preview all managed upgrades (docs + setup):

```bash
centinela migrate
```

Preview docs-only upgrades:

```bash
centinela migrate docs
```

Apply full sync:

```bash
centinela migrate --apply
```

Scope setup migration to one integration when needed:

```bash
centinela migrate setup --agent claude
centinela migrate setup --agent opencode
centinela migrate setup --agent both --apply
```

Regenerate the HTML presentation explicitly when needed:

```bash
centinela docs validate
centinela docs generate --out docs/project-docs/index.html --title "Centinela Project Documentation"
```

## Generated project structure

After `centinela init` and filling in `PROJECT.md`:

```
your-project/
  CLAUDE.md                      ← framework rules (auto-loaded by Claude)
  PROJECT.md                     ← project definition
  centinela.toml                 ← validate commands + gate config
  .claude/
    settings.json                ← centinela hooks
  .workflow/
    <feature>.json               ← workflow state per feature
    <feature>-gatekeeper.md      ← gatekeeper conflict report
  docs/
    architecture/                ← reference documents
    plans/                       ← one plan doc per feature
  specs/
    <feature>.feature            ← Gherkin acceptance criteria
  tests/
    unit/
    integration/
    acceptance/
      <feature>.steps.*          ← executable Gherkin step definitions
```

## Included architecture documentation

`centinela init` copies reference documents into `docs/architecture/`, including:

| Document | Contents |
|----------|---------|
| `architecture-overview.md` | All five archetypes compared — when to use each |
| `artifact-templates.md` | Exact setup and per-feature workflow artifact templates |
| `hexagonal.md` · `rails-native.md` · `n-tier.md` · `ecs.md` · `modular.md` | Per-archetype layer rules, forbidden imports, and conventions |
| `dependency-injection.md` | DI container patterns across archetypes |
| `testing-strategy.md` | Unit, integration, and acceptance test structure for all archetypes |
| `gatekeepers.md` | Full gate reference (G1–G11) with per-archetype rules |
| `gatekeeper-prompt.md` | Prompt for the Gatekeeper AI subagent conflict review |
| `documentation-generator-prompt.md` | LLM-first prompt template for polished docs generation with CLI fallback |
| `workflow-enforcement.md` | How the three enforcement layers work |
| `i18n-strategy.md` | Translation key conventions by format |
| `example-feature-walkthrough.md` | End-to-end example of the five-step workflow |
| `new-project-guide.md` | Step-by-step setup for new projects |

---

← Back to the [documentation index](README.md) · [Configure for your use case](configuration.md)
