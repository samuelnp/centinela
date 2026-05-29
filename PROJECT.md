# Centinela — Project Definition

> This file is the single source of truth for everything project-specific.
> CLAUDE.md defines how to build. This file defines what to build.

## Elevator Pitch

Centinela is a best-practices and engineering workflow enforcement tool that forces Claude to always plan, write code, write tests, validate, and document — in order — without skipping steps.

## Problem Statement

**Who uses it:** Developers using Claude Code to autonomously generate code.

**Pain it solves:** Claude tends to skip planning, skip writing tests, or jump straight to code without validation and documentation. Centinela enforces the plan → code → tests → validate → docs cycle via Claude hooks, blocking out-of-order file writes and requiring explicit step completion.

## Architecture Choice

**Archetype:** n-tier

**Pattern:** N-Tier Layered Architecture

**Why:** Centinela is a CLI tool with a clear, strict dependency stack: commands (cmd/) orchestrate business logic (internal/workflow, internal/gates), which reads configuration (internal/config). Each layer only depends on layers below it. No circular dependencies. Presentation (internal/ui) is a pure rendering concern. This maps naturally to n-tier.

**Reference:** [architecture-overview.md](docs/architecture/architecture-overview.md)

**G2 rule (layer boundaries):** `cmd/` may import `internal/*`. `internal/workflow` and `internal/gates` may import `internal/config` only. `internal/ui` may import `internal/workflow`, `internal/gates`, `internal/roadmap` (read-only, for rendering types). `internal/config` imports nothing internal. `internal/verify` may import `internal/config`, `internal/evidence`, `internal/orchestration`, and `internal/worktree` (read-only); it must not import `cmd/` or `internal/ui`.

**G7 rule (outer layer):** `cmd/centinela/` is the outer layer. Commands must be thin orchestrators — no business logic, no validation rules, no file classification. All decisions belong in `internal/`.

## Tech Stack

| Concern | Technology |
|---------|------------|
| Framework | Cobra (CLI) |
| Language | Go |
| Styling | Charmbracelet Lipgloss |
| Persistence | JSON files in `.workflow/` |
| Unit/Integration Tests | `go test` (stdlib) |
| Acceptance Tests | `go test` (stdlib, in tests/acceptance/) |
| i18n | None (English-only CLI output) |
| External APIs | Claude Code hooks (stdin/stdout JSON) |

## Folder Structure

```
cmd/centinela/        # CLI entry points: Cobra commands and hooks (outer layer)
internal/
  config/             # TOML config loading — leaf layer, no internal imports
  workflow/           # Core domain: step state, file classification, artifact validation
  gates/              # Quality gate enforcement (G1 file size, G11 i18n)
  ui/                 # Terminal rendering: styles, boxes, status — pure presentation
  roadmap/            # Feature/phase tracking, derives status from workflow state
  setup/              # Hook wiring: injects centinela commands into .claude/settings.json
  scaffold/           # Template extraction: embeds and writes CLAUDE.md, PROJECT.md.template, docs/
docs/
  architecture/       # Architecture guides (archetype docs, gatekeeper prompt, etc.)
  plans/              # Per-feature plan documents
specs/                # Gherkin .feature files (acceptance criteria)
tests/
  unit/               # Unit tests mirroring internal/ packages
  integration/        # Integration tests (multi-package interactions)
  acceptance/         # Acceptance step definitions
.workflow/            # Runtime state: per-feature JSON + gatekeeper reports
```

## Domain Language

| Entity | What it represents |
|--------|--------------------|
| Workflow | A single feature's 5-step lifecycle (plan → code → tests → validate → docs) |
| Step | One of the four phases; has a status (pending / in-progress / done) |
| Feature | The named unit of work being tracked (maps to a `.workflow/<feature>.json` file) |
| Gate | A built-in automated check (G1 file size, G11 i18n) run at the validate step |
| Config | The centinela.toml settings for the host project |
| Scaffold | The embedded templates centinela copies into a new project on `init` |

## Layer Mapping

| Abstract Layer | Concrete Path |
|---------------|---------------|
| Outer (CLI) | `cmd/centinela/` |
| Application / Orchestration | `cmd/centinela/` (thin wiring only) |
| Domain / Business Logic | `internal/workflow/`, `internal/gates/` |
| Supporting Domain | `internal/roadmap/` |
| Presentation | `internal/ui/` |
| Infrastructure | `internal/setup/`, `internal/scaffold/` |
| Configuration (leaf) | `internal/config/` |

## Integration Points

| Service | Purpose | Protocol / Auth |
|---------|---------|-----------------|
| Claude Code hooks | Intercept file writes (PreToolUse) and inject workflow tags (PostToolUse) | stdin/stdout JSON |
| `.claude/settings.json` | Register centinela commands as hooks | Local file read/write |

## Project-Specific Rules

1. `cmd/` commands must not contain business logic. If a decision belongs to the domain (e.g. "is this file allowed in this step?"), it lives in `internal/workflow/` or `internal/gates/`.
2. `internal/ui/` must not mutate state. It renders; it does not decide.
3. All user-visible text is English only. No i18n gate required.
4. Test file suffix for Go: `_test.go` (stdlib convention). The `[workflow] test_suffixes` in `centinela.toml` should be set to `["_test.go"]` for centinela's own tests.

## Naming Conventions

| Layer | Pattern | Example |
|-------|---------|---------|
| Domain type | PascalCase noun | `Workflow`, `StepState`, `FileType` |
| Domain function | PascalCase verb phrase | `ClassifyFile`, `ValidateArtifacts` |
| Gate | PascalCase noun | `FileSizeGate` (in file_size.go) |
| Config struct | PascalCase + Config suffix | `WorkflowConfig`, `GatesConfig` |
| CLI command file | snake_case verb | `hook_prewrite.go`, `complete.go` |
| UI render function | Render + PascalCase noun | `RenderContext`, `RenderBlocked` |
| Test file | mirrors file under test + `_test.go` | `classify_test.go` |
| Spec | kebab-case + `.feature` | `workflow-steps.feature` |

## Locales

| Code | Language |
|------|----------|
| `en` | English (only) |

i18n gate: **disabled** (`gates.i18n = false` in centinela.toml).

## Gatekeeper Paths

| What | Path |
|------|------|
| Feature specs | `specs/` |
| Domain — workflow logic | `internal/workflow/` |
| Domain — gates | `internal/gates/` |
| Configuration | `internal/config/` |
| Presentation | `internal/ui/` |
| CLI commands (outer layer) | `cmd/centinela/` |
| Supporting domain | `internal/roadmap/` |
| Infrastructure | `internal/setup/`, `internal/scaffold/` |
