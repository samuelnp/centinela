# Gatekeeper Report: auto-start-feature-intent

**Status:** SAFE

## Scope Reviewed

- `internal/hookpolicy/prewrite.go`
- `cmd/centinela/hook_autostart.go`
- `cmd/centinela/hook_context.go`
- `internal/autostart/intent.go`
- `internal/setup/hooks.go`
- `internal/setup/opencode_plugin.go`

## Gate Checks

- File size gate: PASS (all touched files under 100 lines)
- Layer boundaries: PASS (logic in `internal/`, thin `cmd/` orchestration)
- Workflow enforcement: PASS (`done` workflows no longer permit new writes)
- Auto-start behavior: PASS (new-feature intent auto-starts only when no active workflow)

## Validation

- `go test ./...` passes.
- `centinela validate` passes.
