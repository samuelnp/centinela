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

### Layer 2: Required Artifacts Per Step (ENFORCED by `workflow.sh complete`)

The `complete` command validates artifacts exist on disk BEFORE advancing.
If validation fails, the step stays in-progress and the script exits with error.

| Step | Validation | What it checks |
|------|-----------|----------------|
| plan | File search | A plan file in `docs/plans/` mentions the feature name + a `.feature` file exists in `specs/` |
| code | None | Architecture rules govern this step |
| tests | File search | Test suite files exist in `tests/unit/` or `tests/integration/` + acceptance step definitions exist in `tests/acceptance/` |
| validate | Test run | Project's full test suite exits with code 0 |

> Note: The exact file extensions and paths checked are project-specific. See PROJECT.md → Folder Structure for the authoritative paths.

### Layer 3: Enforcement Script

`scripts/centinela-workflow.sh` — TWO enforcement mechanisms:

1. **`check`**: Prevents writing to a layer before its step is active.
   Returns exit code 1 if blocked.
2. **`complete`**: Prevents advancing past a step without its artifact.
   Returns exit code 1 if artifact missing.

## Workflow Commands

The AI agent must use these commands to manage workflow:

```bash
# Start a new feature workflow
scripts/centinela-workflow.sh start <feature-name>

# Mark current step as done (validates artifact exists)
scripts/centinela-workflow.sh complete <feature-name>

# Check if a write to a layer is allowed
scripts/centinela-workflow.sh check <feature-name> <layer>

# Show current status
scripts/centinela-workflow.sh status <feature-name>

# Show status of all active features
scripts/centinela-workflow.sh status-all
```

## How the AI Agent Must Behave

BEFORE starting any feature:
1. Run `scripts/centinela-workflow.sh start <feature-name>`
2. This creates the .workflow JSON and sets step to "plan"

BEFORE writing any code file:
1. Check current step in workflow
2. If the target file's layer doesn't match current step, STOP
3. Complete the current step first, then advance

AFTER each step:
1. Run `scripts/centinela-workflow.sh complete <feature-name>`
2. This validates the artifact exists and advances to next step
3. Output the current workflow status to the user

## Skip Rules

Some features may not need all steps (e.g., a backend-only feature
doesn't need a UI layer). In that case:
- Mark skipped steps explicitly: `scripts/centinela-workflow.sh skip <feature> <step>`
- Skipping requires a reason logged in the workflow JSON
- Domain/core logic, tests, and validate can NEVER be skipped
