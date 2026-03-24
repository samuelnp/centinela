# Gatekeeper Report: docs-migration-managed-docs

**Status:** SAFE

## Scope Reviewed

- Managed-doc migration engine (`internal/migration/`)
- CLI + hook integration (`cmd/centinela/migrate*`, `cmd/centinela/hook_migrate.go`)
- Setup integration parity (`internal/setup/hooks.go`, `internal/setup/opencode_plugin.go`)
- Scaffold template version headers (`internal/scaffold/assets/*.md`)

## Gate Checks

- File size gate: PASS (all files under 100 lines)
- Layer boundaries: PASS (no business logic in `cmd/`, logic in `internal/`)
- Tests: PASS (`go test ./...`)
- Validate: PASS (`centinela validate`)

## Notes

- Migration is preview-first and apply-only on explicit command.
- Hook output asks for user approval before applying changes.
- Keep blocks and custom sections are preserved during migration.
