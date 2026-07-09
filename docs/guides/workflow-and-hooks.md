# The Five-Step Workflow & How the Hooks Work

> The enforced delivery pipeline and the agent hooks that make it automatic.

## The standard five-step workflow

Most delivery features follow the same five steps in order. Phase 0 bootstrap work can use a shorter roadmap-defined step order, but Centinela still enforces that order and its required artifacts.

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

In strict orchestration mode, specialist evidence must be actionable:

- `big-thinker` and `feature-specialist` outputs must point to real `docs/plans/...` or `specs/...` artifacts.
- `senior-engineer` outputs must include at least one real non-evidence implementation file.
- `ux-ui-specialist` is required during `code` only for features whose brief declares `surface: user-facing`; its outputs must include at least one real UI file, `mobileFirst: true`, and the required UX review tags.
- `qa-senior` outputs must include at least one real test file and `.workflow/<feature>-edge-cases.md`.

UI path enforcement is configurable:

```toml
[orchestration]
ui_paths = ["src/ui", "src/components", "app/views", "web", "styles"]
```

Step confirmation prompts are configurable in `centinela.toml`:

```toml
[workflow]
step_confirmation_mode = "every_step" # every_step | after_plan | auto
plan_advisor_mode = "missing_info"   # missing_info | always | off
plan_question_limit = 4               # capped at 4 questions per round
```

See the [configuration reference](configuration-reference.md) for every workflow knob.

### Workflow commands

```bash
centinela start <feature>       # Start a feature (required before any file writes)
centinela status <feature>      # Show current step and artifact status
centinela status-all            # Show all active features
centinela complete <feature>    # Mark step done and advance
centinela verify <feature>      # Independently re-derive ground truth for evidence claims
centinela roadmap               # Show roadmap phase and feature progress
centinela roadmap validate      # Validate roadmap analysis and quality artifacts
centinela validate              # Run gate checks manually
centinela migrate               # Preview full managed docs + setup migration
centinela migrate docs          # Preview managed docs migration only
centinela migrate setup         # Preview setup migration only
centinela docs validate         # Validate inputs for project documentation report
centinela docs generate         # Generate HTML docs with Mermaid diagrams
```

## How the hooks work

`centinela init` wires Claude and OpenCode integrations. Under the hood those integrations call `centinela hook ...` commands to keep workflow enforcement, setup guidance, and session context in sync.

### PreToolUse — Write / Edit

Before Claude writes or edits any file, centinela checks whether that file belongs to the current workflow step. If you are in the `plan` step and Claude tries to write a source file, the hook blocks it and explains why.

### PostToolUse — Write / Edit

After every file write, centinela appends a compact status tag to the session:

```
↳ my-feature · code · 2/5
```

For Claude, Centinela can also render a compact status line so the current feature, step, and risk state stay visible outside the main response flow.

### Prompt Advisor — Plan Step

During the `plan` step, Centinela also injects a plan-advisor directive. By default it runs in
`missing_info` mode, inspects `docs/features/<feature>.md`, `docs/plans/<feature>.md`, and
`specs/<feature>.feature`, then enriches that with roadmap dependencies, same-phase sibling
features, roadmap quality notes, and related edge-case lessons. It asks up to 4 missing
high-value questions through `big-thinker` and `feature-specialist` lenses. Dependency context is
preferred before sibling context, and user-facing features receive UX/mobile-first questions only
when those topics are still missing.

### UserPromptSubmit

Multiple hooks run at the start of every message:

**Project setup** — if `PROJECT.md` is missing but `PROJECT.md.template` exists, centinela injects a prompt instructing the agent to interview the user and write `PROJECT.md`. Once `PROJECT.md` exists, Centinela can also require `ROADMAP.md`, roadmap analysis artifacts, roadmap quality artifacts, and production-readiness setup before feature work continues.

**Managed migration** — if managed docs or setup assets are outdated, centinela
injects migration guidance and instructs the assistant to ask for approval before
running apply commands.

**Workflow context** — injects a context block showing all active workflows and their current step, so Claude always has accurate state without reading any files.

**Autostart + orchestration** — when no workflow is active, Centinela can auto-start a feature from prompt intent. In strict orchestration mode it also tells the agent which specialist evidence files are required before `centinela complete` can advance the step.

---

← Back to the [documentation index](README.md) · [Quality gates](gates.md) · [MCP governance](mcp.md)
