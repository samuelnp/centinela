# Plan: OpenCode Hook Parity

## Scope
Bring OpenCode plugin behavior to parity with existing Claude hook flow by invoking the same `centinela hook` commands.

## Work Items
1. Update generated OpenCode plugin content in `internal/setup/opencode_plugin.go`.
2. Keep `prewrite` behavior unchanged and add `postwrite` call after write tools.
3. Add prompt lifecycle handlers that call `context` and `setup` hooks.
4. Append non-empty hook outputs to prompt/session context in OpenCode.
5. Add tests validating generated plugin includes all required hook commands.

## Validation
- `go test ./...`
- Smoke test: `centinela init --agent opencode` and verify generated plugin code.

## Compatibility
- No changes to workflow state model.
- No changes to Claude settings merge behavior.
