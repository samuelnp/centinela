# Workflow Enforcement System

Centinela uses mechanical enforcement so agents cannot silently skip process.

## Layer 1: Workflow State (`.workflow/`)

Each feature has `.workflow/<feature>.json` with:

- `currentStep` in `plan|code|tests|validate|done`
- per-step status (`pending|in-progress|done`)
- completion timestamps

This file is the source of truth for progress.

## Layer 2: Artifact Validation (`centinela complete`)

Before advancing, Centinela validates required artifacts:

| Step | Required artifacts |
|------|--------------------|
| plan | `docs/features/<feature>.md`, `docs/plans/<feature>.md`, and at least one `specs/*.feature` |
| code | none (architecture rules apply during implementation) |
| tests | test files in `tests/unit` or `tests/integration`, acceptance files in `tests/acceptance`, and `.workflow/<feature>-edge-cases.md` |
| validate | gatekeeper report at `.workflow/<feature>-gatekeeper.md` and `centinela validate` pass |

If validation fails, the step remains in progress.

## Layer 3: Hook Enforcement

Centinela hooks enforce write discipline and context:

- `centinela hook prewrite` blocks out-of-step writes.
- `centinela hook postwrite` emits compact workflow tags.
- `centinela hook setup` injects setup guidance when required files are missing.
- `centinela hook context` injects active workflow context.

These hooks are wired by `centinela init` for Claude and OpenCode integrations.

## Workflow Commands

```bash
centinela start <feature>
centinela complete <feature>
centinela status <feature>
centinela status-all
centinela validate
```

## Required Agent Behavior

1. Start every feature with `centinela start <feature>`.
2. Do work only for the active step.
3. After producing artifacts, run `centinela complete <feature>`.
4. Do not auto-advance without explicit user confirmation.

## Validate Step

`centinela validate` runs:

1. Built-in gates (for example G1 file-size, G11 i18n if enabled).
2. All commands in `centinela.toml` `[validate] commands`.

Example:

```toml
[validate]
commands = [
  "go test ./..."
]
```
