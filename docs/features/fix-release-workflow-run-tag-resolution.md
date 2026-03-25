# Feature Brief: Fix Release Tag Resolution for workflow_run

## Problem
The Release workflow runs after Version Bump via `workflow_run`, but it resolves tags using `workflow_run.head_sha`. That SHA points to the pre-bump commit, so no release tag is found and release publishing is skipped.

## Goal
Resolve the latest release tag from `main` when Release is triggered by `workflow_run`, so CI-created bump tags always publish artifacts and a GitHub Release.

## Scope
- Update `workflow_run` tag lookup in `.github/workflows/release.yml`.
- Keep manual `push` tag behavior unchanged.
- Preserve skip behavior when no `v*` tag exists.

## Acceptance Criteria
- `workflow_run` path fetches tags and resolves latest semver tag merged into `origin/main`.
- Release builds and publishes for `linux/darwin/windows` x `amd64/arm64`.
- Manual `push` on `v*` still works.
- Tests and `centinela validate` pass.
