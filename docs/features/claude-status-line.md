# Feature Brief: Claude status line for Centinela

## Problem
Claude users can see hook panels, but they do not get a compact always-visible workflow signal. This makes it easy to miss the current step, blockers, and next required action.

## Users
- Developers using Claude Code with Centinela enabled.
- Teams that enforce plan -> code -> tests -> validate flow.

## Goals
- Add a Centinela-powered Claude `statusLine` with high-signal workflow tokens.
- Keep output compact, deterministic, and cheap to render.
- Reuse existing workflow/gate rules instead of duplicating policy in shell scripts.

## Non-Goals
- Replacing existing hook panels.
- Changing workflow semantics or gate definitions.
- Adding OpenCode-specific UI in this feature.

## Acceptance Criteria
- Claude settings include a managed `statusLine` command for Centinela.
- Status line displays feature, step, progress, next action, blocker code, and risk.
- Output handles no-workflow state gracefully.
- Tokens are stable and parse-friendly (for both humans and Claude).

## Edge Cases
- Multiple active workflows.
- Missing or invalid workflow files.
- Missing roadmap/setup artifacts.
- Validate step warning states from production readiness report.

## Risks
- Slow status line updates if shell/git checks are expensive.
- Drift between status output and workflow validation logic.
- Truncation in narrow terminals if line grows too long.
