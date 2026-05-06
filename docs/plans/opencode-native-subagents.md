# Plan: opencode-native-subagents

## Scope

Generate native OpenCode subagent configuration for Centinela specialist roles while preserving existing `opencode.json` content.

## Work Items

1. Extend OpenCode config merging to add missing Centinela `agent` entries.
2. Add concise prompts and descriptions for `big-thinker`, `feature-specialist`, `senior-engineer`, `qa-senior`, `documentation-specialist`, and `ux-ui-specialist`.
3. Configure the `build` primary agent to allow Task access to those subagents without removing existing permissions.
4. Add unit and acceptance tests covering generation, idempotency, and preservation of existing user agents.

## Validation

- `go test ./...`
- `centinela validate`

## Compatibility

- Preserve existing `opencode.json` keys.
- Do not overwrite existing agent entries.
- Keep OpenCode plugin and instruction generation unchanged except for native agent availability.
