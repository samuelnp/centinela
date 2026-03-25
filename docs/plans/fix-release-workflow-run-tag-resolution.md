# Plan: Fix Release Tag Resolution for workflow_run

## Scope
Make Release workflow reliably discover the new tag created by Version Bump.

## Work Items
1. Update `.github/workflows/release.yml` tag resolution.
   - Keep `push` path using `GITHUB_REF_NAME`.
   - For `workflow_run`, fetch tags and `origin/main`.
   - Resolve latest `v*` using `git tag --merged origin/main --list 'v*' --sort=-version:refname`.
2. Keep graceful skip when no tag is found.
3. Update integration and acceptance tests for new tag lookup commands.

## Validation
- `go test ./...`
- `go run ./cmd/centinela validate`

## Constraints
- Keep all touched files under 100 lines.
- Preserve existing release artifact and checksum flow.
