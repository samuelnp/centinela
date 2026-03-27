# Centinela

A development workflow enforcer for Claude Code and OpenCode projects. Centinela turns the "plan → code → tests → validate → docs" discipline from a suggestion into a mechanical constraint — enforced by agent integrations that run automatically in coding sessions.

---

## Why Centinela

AI coding agents are fast but undisciplined. Left to their own devices they skip planning, write tests as an afterthought, and ship without validation. Centinela fixes this by:

- **Blocking file writes** in the wrong workflow step via agent integrations
- **Requiring artifacts** before a step can advance — no plan file means no code, no tests means no validate
- **Running gate checks** automatically at the validate step (file size limits, i18n completeness, your test suite)
- **Injecting context** into every agent session so the model always knows which feature is active and which step it is on

The result: every feature ships with a written plan, a Gherkin spec, three test layers, and a passing gate suite — regardless of whether a human or an AI agent wrote it.

---

## Install

**Prerequisites:** Go 1.21+

```bash
go install github.com/samuelnp/centinela@latest
```

Or download a pre-built binary from [Releases](https://github.com/samuelnp/centinela/releases).

For macOS/Linux, you can install the latest release binary directly:

```bash
curl -fsSL https://raw.githubusercontent.com/samuelnp/centinela/main/scripts/install.sh | sh
```

Verify:

```bash
centinela --help
```

> **macOS/Linux:** `go install` places the binary in `~/go/bin`. Ensure that directory is on your PATH:
> ```bash
> export PATH="$HOME/go/bin:$PATH"  # add to ~/.zshrc or ~/.bashrc
> ```

---

## Quick Start

### 1. Initialize a project

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

### 2. Fill in PROJECT.md

Open a Claude session in your project — the setup hook will detect that `PROJECT.md` is missing and automatically prompt Claude to interview you and write it. Alternatively, rename `PROJECT.md.template` to `PROJECT.md` and complete every section manually. This file tells both you and Claude what the project is, which architecture pattern it follows, and where everything lives.

### 3. Configure centinela.toml

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

### Example: Bootstrap a new project

Use this flow when starting from scratch.

```bash
centinela init
```

Then open your coding agent in the repo and follow setup prompts.

Expected sequence:

1. If `PROJECT.md` is missing, centinela asks the agent to interview you and write it.
2. Once `PROJECT.md` exists, centinela asks the agent to define your roadmap.
3. The agent produces:
   - `ROADMAP.md` (human-readable phased plan)
   - `.workflow/roadmap.json` (machine-readable roadmap)
   - `docs/features/<feature-slug>.md` for each Phase 1 feature

Verify roadmap status:

```bash
centinela roadmap
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

### 4. Start building

```bash
centinela start my-feature
```

The hooks take it from here.

### 5. Migrate managed assets

Preview all managed upgrades (docs + setup):

```bash
centinela migrate
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

---

## The Five-Step Workflow

Every feature follows the same five steps in order. No step can be skipped.

```
plan → code → tests → validate → docs
```

| Step | What you produce | What centinela checks before advancing |
|------|-----------------|---------------------------------------|
| **plan** | Plan doc in `docs/plans/` + Gherkin spec in `specs/` | Both files exist on disk |
| **code** | Implementation | Nothing — architecture rules govern this step |
| **tests** | Unit, integration, acceptance + edge-case analysis | Test files exist + `.workflow/<feature>-edge-cases.md` exists |
| **validate** | Gatekeeper conflict report | All gate checks pass + all `centinela.toml` commands exit 0 |
| **docs** | Human-facing project documentation | `.workflow/<feature>-documentation-specialist.md` + `.workflow/<feature>-documentation-specialist.json` + `docs/project-docs/index.html` |

### Workflow commands

```bash
centinela start <feature>       # Start a feature (required before any file writes)
centinela status <feature>      # Show current step and artifact status
centinela status-all            # Show all active features
centinela complete <feature>    # Mark step done and advance
centinela validate              # Run gate checks manually
centinela docs validate         # Validate inputs for project documentation report
centinela docs generate         # Generate HTML docs with Mermaid diagrams
```

---

## How the Hooks Work

`centinela init` registers Claude Code hooks that run automatically:

### PreToolUse — Write / Edit

Before Claude writes or edits any file, centinela checks whether that file belongs to the current workflow step. If you are in the `plan` step and Claude tries to write a source file, the hook blocks it and explains why.

### PostToolUse — Write / Edit

After every file write, centinela appends a compact status tag to the session:

```
↳ my-feature · code · 2/4
```

### UserPromptSubmit

Multiple hooks run at the start of every message:

**Project setup** — if `PROJECT.md` is missing but `PROJECT.md.template` exists, centinela injects a prompt instructing Claude to interview the user and write `PROJECT.md`. The prompt disappears automatically once the file is created.

**Managed migration** — if managed docs or setup assets are outdated, centinela
injects migration guidance and instructs the assistant to ask for approval before
running apply commands.

**Workflow context** — injects a context block showing all active workflows and their current step, so Claude always has accurate state without reading any files.

---

## Gate Checks

Gates are quality checks that must pass before a feature can ship. They run during `centinela validate` and automatically when completing the `validate` step.

### Built-in gates

| Gate | Rule | Config |
|------|------|--------|
| **G1: File Size** | No source file exceeds 100 lines | `[gates] file_size = true` |
| **G11: i18n** | All locale files have identical keys (no missing translations) | `[gates] i18n = true` |

G11 supports two formats natively:

```toml
# JSON locale files (next-intl, i18next, vue-i18n)
[i18n]
format  = "json"
dir     = "src/i18n/messages"
locales = ["en", "es", "fr"]

# GNU gettext .po files (Godot, Qt)
[i18n]
format  = "gettext"
dir     = "i18n"
locales = ["en", "es"]
```

For other formats (Unity CSV, Android XML, iOS `.lproj`), set `format = "none"` and add a custom command to `[validate] commands`.

### Manual gates (code review)

| Gate | Rule |
|------|------|
| **G2: Layer Dependencies** | No imports cross forbidden layer boundaries (archetype-specific) |
| **G3: Type Safety** | Strictest static analysis — no `any`, no untyped variables |
| **G5: Spec First** | Every feature has a `.feature` file before implementation starts |
| **G6: Plan First** | Every feature has a plan document before implementation starts |
| **G7: No Business Logic in Outer Layer** | UI components and adapters contain no domain logic |
| **G8: Single Responsibility** | Each file exports one thing and does one thing |

Full gate documentation: [`docs/architecture/gatekeepers.md`](internal/scaffold/assets/docs/architecture/gatekeepers.md)

---

## Architecture Archetypes

Centinela supports five architecture patterns. You choose one when filling in `PROJECT.md`. The choice determines which layer rules, forbidden imports, and test expectations apply.

| Archetype | Best for |
|-----------|---------|
| **Hexagonal** | Multiple external integrations (APIs, databases), domain logic that must be testable without infrastructure |
| **Rails-native** | Framework-opinionated stacks (Rails, Django, Laravel) — follow the framework conventions |
| **N-Tier** | Classic layered apps: HTTP handlers → services → repositories |
| **ECS** | Games — entities, components, and systems |
| **Modular** | Monorepo-style projects with clear public API contracts between modules |

Each archetype has its own layer dependency rules (G2), outer-layer definition (G7), and test coverage expectations (G4).

See [`docs/architecture/architecture-overview.md`](internal/scaffold/assets/docs/architecture/architecture-overview.md) for the full comparison.

---

## centinela.toml Reference

```toml
# Commands centinela runs during the validate step.
# Executed natively — no shell scripts required.
[validate]
commands = [
  # TypeScript:  "npx tsc --noEmit", "npx vitest run"
  # Python:      "mypy --strict src", "pytest"
  # Go:          "go vet ./...", "go test ./..."
  # Ruby:        "bundle exec rubocop", "bundle exec rspec"
  # Rust:        "cargo check", "cargo test"
]

# Built-in gate toggles
[gates]
file_size = true   # G1: fail if any source file exceeds 100 lines
i18n      = false  # G11: check translation key completeness

# i18n config (required when gates.i18n = true)
[i18n]
format  = "json"              # "json" | "gettext" | "none"
dir     = "src/i18n/messages"
locales = ["en"]
```

---

## Generated Project Structure

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
    architecture/                ← 14 reference documents
    plans/                       ← one plan doc per feature
  specs/
    <feature>.feature            ← Gherkin acceptance criteria
  tests/
    unit/
    integration/
    acceptance/
      <feature>.steps.*          ← Gherkin step definitions
```

---

## Included Architecture Documentation

`centinela init` copies 15 reference documents into `docs/architecture/`:

| Document | Contents |
|----------|---------|
| `architecture-overview.md` | All five archetypes compared — when to use each |
| `hexagonal.md` | Ports-and-adapters layers, dependency rules, forbidden imports |
| `rails-native.md` | MVC conventions, what belongs in models vs services vs views |
| `n-tier.md` | Controller → Service → Repository layer rules |
| `ecs.md` | Entity-Component-System patterns for games |
| `modular.md` | Module boundaries and public API contracts |
| `dependency-injection.md` | DI container patterns across archetypes |
| `testing-strategy.md` | Unit, integration, and acceptance test structure for all archetypes |
| `gatekeepers.md` | Full gate reference (G1–G11) with per-archetype rules |
| `gatekeeper-prompt.md` | Prompt for the Gatekeeper AI subagent conflict review |
| `documentation-generator-prompt.md` | LLM-first prompt template for polished docs generation with CLI fallback |
| `workflow-enforcement.md` | How the three enforcement layers work |
| `i18n-strategy.md` | Translation key conventions by format |
| `example-feature-walkthrough.md` | End-to-end example of the four-step workflow |
| `new-project-guide.md` | Step-by-step setup for new projects |

---

## Build from Source

```bash
git clone https://github.com/samuelnp/centinela
cd centinela
go build -o centinela ./cmd/centinela/
```

Cross-compile for other platforms:

```bash
GOOS=linux   GOARCH=amd64 go build -o centinela-linux-amd64  ./cmd/centinela/
GOOS=darwin  GOARCH=arm64 go build -o centinela-darwin-arm64 ./cmd/centinela/
GOOS=windows GOARCH=amd64 go build -o centinela-windows.exe  ./cmd/centinela/
```

---

## Contributing

Centinela uses its own workflow to develop itself.

```bash
centinela start <feature-name>
# plan → code → tests → validate → docs
centinela complete <feature-name>
```

Conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`.
One feature per branch. Never push failing tests.

---

## License

MIT
