# Plan: Fix Release Trigger After Version Bump

## Scope
Make release automation run for tags created by CI and for manual tag pushes.

## Work Items
1. Update `.github/workflows/release.yml` triggers.
   - Keep `push` tags `v*` for manual tagging.
   - Add `workflow_run` for successful `Version Bump` completion.
2. Add trigger-aware tag resolution in release workflow.
   - Use `GITHUB_REF_NAME` on tag pushes.
   - Use `git tag --points-at <head_sha>` for `workflow_run` path.
3. Keep build and publish behavior unchanged.
   - Build 6 binaries and publish `SHA256SUMS`.
4. Update release workflow tests.
   - Unit/integration/acceptance checks include `workflow_run` and tag resolution.

## Validation
- `go test ./...`
- `go run ./cmd/centinela validate`

## Constraints
- File size limit under 100 lines per file.
- No business logic in CLI command layer.
