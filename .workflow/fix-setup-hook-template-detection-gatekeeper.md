# Gatekeeper Report: fix-setup-hook-template-detection

**Status:** SAFE

## Scope Reviewed

- `cmd/centinela/hook_setup.go`
- `cmd/centinela/hook_setup_directive_test.go`

## Findings

- Setup hook no longer depends solely on `PROJECT.md.template` presence.
- Roadmap guidance still appears when `PROJECT.md` exists and `ROADMAP.md` is missing.
- Added plain `CENTINELA DIRECTIVE` lines before boxed guidance to improve
  instruction salience for LLMs.

## Validation

- `go test ./...` passes.
- `centinela validate` passes.
