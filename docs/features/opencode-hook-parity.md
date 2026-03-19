# Feature Brief: OpenCode Hook Parity

## Problem
OpenCode integration currently enforces only prewrite blocking. It does not yet mirror the Claude integration for postwrite tags, workflow context injection, or setup/roadmap prompts.

## Users
- Teams using OpenCode as their primary coding agent.
- Teams using Claude and OpenCode in mixed workflows.

## Goals
- Mirror the existing hook behavior in OpenCode sessions.
- Keep workflow decision logic centralized in existing `centinela hook` commands.
- Preserve backward compatibility for Claude-only projects.

## Acceptance Criteria
- OpenCode plugin calls `centinela hook postwrite` after write/edit/patch operations.
- OpenCode plugin calls `centinela hook context` and `centinela hook setup` on prompt submit.
- OpenCode plugin appends hook outputs into prompt context so the agent sees the same guidance.
- Existing prewrite enforcement keeps working.

## Risks
- Plugin event shape differences across OpenCode versions.
- Duplicate context injection if hooks are invoked multiple times per prompt.

## Decomposition
- Extend `.opencode/plugins/centinela.js` generated content.
- Keep payload format aligned with existing hook command parsers.
- Add tests for generated plugin content and coverage for new behavior.
