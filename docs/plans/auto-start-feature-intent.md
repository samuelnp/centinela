# Plan: Auto-start Workflow from New Feature Intent

## Scope
Add intent-based workflow auto-start and tighten prewrite enforcement when no workflow is active.

## Work Items
1. Update prewrite policy to ignore `done` workflows for write authorization.
2. Add domain logic to detect new-feature intent and derive a feature slug.
3. Add `centinela hook autostart` command for prompt-time auto-start.
4. Wire autostart hook into setup hook registration and OpenCode plugin output.
5. Add directive in context hook when no active workflow exists.
6. Add unit, integration, and acceptance coverage.

## Validation
- `go test ./...`
- `go run ./cmd/centinela validate`

## Constraints
- Keep each touched file under 100 lines.
- Keep business decisions in `internal/`, not `cmd/`.
