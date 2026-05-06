# opencode-native-subagents

## Problem

OpenCode supports native custom subagents, but Centinela currently only injects prompt instructions and validates evidence files. That makes OpenCode orchestration weaker than Claude Code because the main session can simulate specialist roles instead of invoking configured subagents.

## User Stories

- As an OpenCode user, Centinela configures native specialist subagents for plan, code, tests, docs, and UI review.
- As a maintainer, existing OpenCode agent config is preserved while missing Centinela agents are added.
- As a reviewer, generated OpenCode config makes specialist delegation visible and available through the Task tool.

## Acceptance Criteria

- `centinela init` or setup migration adds Centinela subagents to `opencode.json`.
- Existing custom OpenCode agents and existing fields on Centinela agents are preserved.
- The build primary agent is configured to allow Task access to Centinela specialist subagents.
- Tests verify all required Centinela roles are present as `mode: subagent` entries.

## Edge Cases

- Existing user-defined agents remain untouched.
- Existing Centinela agent customizations are not overwritten.
- Missing or empty `agent` config is created safely.
- OpenCode still keeps `AGENTS.md` and `CLAUDE.md` instructions.

## Risks

- OpenCode may still not expose child-session metadata for proof of delegation.
- Overwriting user agent config would be disruptive, so merging must be additive.
