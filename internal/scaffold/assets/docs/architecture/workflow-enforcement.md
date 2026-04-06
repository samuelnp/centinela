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
| tests | File search | Test suite files in `tests/unit/` or `tests/integration/` + acceptance step definitions in `tests/acceptance/` + `.workflow/<feature>-edge-cases.md` |
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
4. Respect `workflow.step_confirmation_mode` for review prompts:
   - `every_step` (default): require explicit confirmation each step.
   - `after_plan`: require confirmation only for plan -> code.
   - `auto`: no review prompt; still run `centinela complete` explicitly.

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
]
```

Commands run natively via the OS — no shell scripts required. This works on Windows, macOS, and Linux.

## Skip Rules

All five steps are mandatory. No step can be skipped — this is enforced by the binary.
Domain/core logic, tests, and validate are especially non-negotiable.

In strict orchestration mode, `plan` evidence from `big-thinker` and
`feature-specialist` must include all `docs/features/*.md` paths in JSON `inputs`
(including the current feature brief).
