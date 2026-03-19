# Feature Brief: Adapt Centinela for OpenCode

## Problem
Centinela currently enforces workflow rules through Claude-specific hooks.
Projects using OpenCode cannot get the same write blocking, context reminders, and setup guidance.

## Users
- Teams that use OpenCode as their primary coding agent.
- Teams that run Claude Code and OpenCode in parallel.

## Goals
- Provide OpenCode integration with parity to current Claude behavior.
- Keep existing Claude integration working without migration friction.
- Avoid duplicating workflow decisions across separate implementations.

## Non-Goals
- Replacing Claude support.
- Changing the core workflow model (`plan -> code -> tests -> validate`).
- Adding new workflow steps.

## Acceptance Criteria
- `centinela init` can set up OpenCode artifacts for a project.
- OpenCode blocks out-of-step file writes using current workflow state.
- OpenCode surfaces setup/context prompts equivalent to existing hooks.
- Existing Claude setup remains supported and backward compatible.

## Risks
- Drift between Claude and OpenCode enforcement logic.
- Config merge errors when project already has OpenCode files.
- Incomplete docs leading to inconsistent project setup.

## Decomposition
- Add target-aware init flow (`claude`, `opencode`, `both`).
- Introduce shared hook decision logic in internal package.
- Add OpenCode assets (config + plugin).
- Add tests for setup, merge behavior, and enforcement parity.
