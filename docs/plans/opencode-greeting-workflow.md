# Plan: opencode-greeting-workflow

## Scope

Align OpenCode's first-response behavior with Claude Code by making Centinela setup and workflow requirements unavoidable in generated OpenCode instructions, especially for greeting-only prompts.

## Work Items

1. Update generated OpenCode instructions in `internal/setup/opencode_agents.go` to require a Centinela-first response for greetings when setup or workflow guidance is required.
2. Add unit and acceptance coverage proving generated OpenCode assets contain the stronger greeting-first setup rule.
3. Re-run relevant OpenCode tests and the full Go test suite.

## Validation

- `go test ./...`
- `centinela validate`

## Compatibility

- Keep existing OpenCode config and plugin paths unchanged.
- Do not overwrite unmanaged user OpenCode files outside the existing migration/sync flow.
- Preserve setup and migration prompt order before autostart/context hooks.
