# Plan: Harden Main Release Automation

## Scope
Adjust release workflows so `main` pushes drive semantic version bumps and `v*` tags drive cross-platform release artifacts.

## Work Items
1. Update `.github/workflows/version-bump.yml`.
   - Trigger only on `push` to `main`.
   - Derive bump type from conventional commit messages since latest `v*` tag.
   - Bump `Makefile` `VERSION :=` and create `chore(release)` commit.
   - Create and push `vX.Y.Z` tag.
   - Keep bot-loop prevention guard.
2. Update `.github/workflows/release.yml`.
   - Trigger on pushed `v*` tags.
   - Build binaries for `linux`, `darwin`, `windows` x `amd64`, `arm64`.
   - Generate and upload `SHA256SUMS` with release assets.
3. Update tests for workflow policy.
   - Unit assertions for workflow file existence and naming.
   - Integration assertions for semantic bump logic and 6-target matrix.
   - Acceptance assertions for bump commit/tag behavior and published artifacts.

## Validation
- `go test ./...`
- `go run ./cmd/centinela validate`

## Constraints
- Keep files under 100 lines.
- Do not change validate workflow behavior.
- Maintain conventional commit based bumping as default policy.
