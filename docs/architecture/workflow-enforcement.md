<!-- centinela:doc-version=1 template=docs/architecture/workflow-enforcement.md -->
# Workflow Enforcement System

## Problem
AI agents (and humans) skip steps when given a list. Instructions alone are
not enough. We need mechanical enforcement.

## Three Enforcement Layers

### Layer 1: Workflow Tracker (`.workflow/`)

Every feature gets a state file at `.workflow/<feature-name>.json`:

```json
{
  "feature": "example-feature",
  "startedAt": "2026-03-05T17:00:00Z",
  "currentStep": "code",
  "steps": {
    "plan": {
      "status": "done",
      "artifact": "docs/plans/example-feature.md",
      "completedAt": "2026-03-05T17:00:00Z"
    },
    "code": {
      "status": "in-progress",
      "artifact": null,
      "completedAt": null
    },
    "tests": { "status": "pending" },
    "validate": { "status": "pending" }
  }
}
```

Rules:
- A step can only move to "in-progress" if the previous step is "done".
- A step is "done" only when its required artifact exists on disk.
- The workflow file is the source of truth for current progress.

### Layer 2: Required Artifacts Per Step (ENFORCED by `centinela complete`)

The `complete` command validates artifacts exist on disk BEFORE advancing.
If validation fails, the step stays in-progress and the command exits with an error.

| Step | Validation | What it checks |
|------|-----------|----------------|
| plan | File search | A plan file in `docs/plans/` + a `.feature` file exists in `specs/` |
| code | None | Architecture rules govern this step |
| tests | File search | Test suite files in `tests/unit/` or `tests/integration/` + executable acceptance artifacts in `tests/acceptance/` + `validate.commands` includes acceptance execution + `.workflow/<feature>-edge-cases.md` |
| validate | Gate checks + commands | All built-in gates pass + all `centinela.toml` validate commands exit 0 |

> Note: The exact file extensions and paths checked are project-specific. See PROJECT.md → Folder Structure for the authoritative paths.

### Layer 3: The `centinela` Binary

The `centinela` CLI enforces the workflow with two mechanisms:

1. **Pre-write hook**: Blocks file writes in the wrong workflow step. Runs automatically via agent integrations.
2. **`centinela complete`**: Prevents advancing past a step without its required artifact or passing gates.

## Workflow Commands

The AI agent must use these commands to manage workflow:

```bash
# Start a new feature workflow
centinela start <feature-name>

# Mark current step as done (validates artifact exists, runs gates on validate step)
centinela complete <feature-name>

# Show current status
centinela status <feature-name>

# Show status of all active features
centinela status-all

# Run gate checks and validate commands manually
centinela validate
```

## How the AI Agent Must Behave

BEFORE starting any feature:
1. Run `centinela start <feature-name>`
2. This creates the `.workflow` JSON and sets step to "plan"

BEFORE writing any code file:
1. Check current step in workflow
2. If the target file's layer doesn't match current step, STOP
3. Complete the current step first, then advance

AFTER each step:
1. Run `centinela complete <feature-name>`
2. This validates the artifact exists and advances to next step
3. Output the current workflow status to the user

## Validate Step

When completing the `validate` step, centinela automatically runs:
1. **Built-in gates** — G1 (file size), G11 (i18n if configured in `centinela.toml`)
2. **User commands** — all entries in `centinela.toml → [validate] commands`

Configure your stack's lint/type-check/test commands in `centinela.toml`:

```toml
[validate]
commands = [
  "npx tsc --noEmit",
  "npx vitest run",
  "npx cucumber-js"
]
```

Commands run natively via the OS — no shell scripts required. This works on Windows, macOS, and Linux.

## Skip Rules

All five steps are mandatory. No step can be skipped — this is enforced by the binary.
Domain/core logic, tests, and validate are especially non-negotiable.

## Preserved Custom Sections

## Layer 1: Workflow State (`.workflow/`)

Each feature has `.workflow/<feature>.json` with:

- `currentStep` in `plan|code|tests|validate|docs|done`
- per-step status (`pending|in-progress|done`)
- completion timestamps

This file is the source of truth for progress.


## Layer 2: Artifact Validation (`centinela complete`)

Before advancing, Centinela validates required artifacts:

| Step | Required artifacts |
|------|--------------------|
| plan | `docs/features/<feature>.md`, `docs/plans/<feature>.md`, and at least one `specs/*.feature` |
| code | none (architecture rules apply during implementation) |
| tests | test files in `tests/unit` or `tests/integration`, executable acceptance files in `tests/acceptance`, at least one acceptance execution command in `[validate] commands`, and `.workflow/<feature>-edge-cases.md` |
| validate | gatekeeper report at `.workflow/<feature>-gatekeeper.md` and `centinela validate` pass |
| docs | `.workflow/<feature>-documentation-specialist.md`, `.workflow/<feature>-documentation-specialist.json`, and `docs/project-docs/index.html` |

In strict orchestration mode, `plan` evidence from `big-thinker` and
`feature-specialist` must include a full snapshot of `docs/features/*.md` paths in
their JSON `inputs` list (including the current feature brief).

If validation fails, the step remains in progress.


## Layer 3: Hook Enforcement

Centinela hooks enforce write discipline and context:

- `centinela hook prewrite` blocks out-of-step writes.
- `centinela hook postwrite` emits compact workflow tags.
- `centinela hook setup` injects setup guidance when required files are missing.
- `centinela hook context` injects active workflow context.

These hooks are wired by `centinela init` for Claude and OpenCode integrations.


## Required Agent Behavior

1. Start every feature with `centinela start <feature>`.
2. Do work only for the active step.
3. After producing artifacts, run `centinela complete <feature>`.
4. Respect `workflow.step_confirmation_mode` for review prompts:
   - `every_step` (default): require explicit user confirmation for each step.
   - `after_plan`: require confirmation only for plan -> code transition.
   - `auto`: no review prompt; still run `centinela complete <feature>` explicitly.
