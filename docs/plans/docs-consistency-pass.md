# Plan: Docs Consistency Pass

## Scope
Update documentation and scaffolded documentation for command consistency and current agent support wording.

## Work Items
1. Locate stale references to legacy script workflow commands.
2. Replace with current CLI commands:
- `centinela start <feature>`
- `centinela complete <feature>`
- `centinela status <feature>`
- `centinela status-all`
- `centinela validate`
3. Update wording where docs imply Claude-only behavior but support is now dual-agent.
4. Apply same updates to scaffold assets under `internal/scaffold/assets/docs/architecture/`.

## Validation
- Run `go test ./...`.
- Search for remaining `scripts/centinela-workflow.sh` references.

## Compatibility
- No code behavior changes expected; documentation-only feature.
