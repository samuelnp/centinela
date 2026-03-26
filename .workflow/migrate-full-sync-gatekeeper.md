# Gatekeeper Report: migrate-full-sync

**Status:** SAFE

## Scope Reviewed

- Unified migration CLI (`cmd/centinela/migrate.go`, `cmd/centinela/migrate_setup.go`)
- Hook migration guidance (`cmd/centinela/hook_migrate.go`)
- Setup sync planning/apply (`internal/setup/sync*.go`)
- Setup migration rendering (`internal/ui/render_setup_migrate.go`)
- Coverage and regression tests for migrate/setup/workflow helpers

## Gate Checks

- File size gate: PASS (all files under 100 lines)
- Layer boundaries: PASS (command layer orchestrates; setup logic in `internal/setup`)
- Tests: PASS (`go test ./...`)
- Validate: PASS (`centinela validate`)

## Notes

- `centinela migrate` now supports preview-first full sync and `--apply` execution.
- `centinela migrate setup --agent ...` scopes setup migrations by integration.
- Custom unmanaged setup files are flagged as `manual-review` and not overwritten.
