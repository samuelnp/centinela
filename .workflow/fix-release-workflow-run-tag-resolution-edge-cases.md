# Edge-Case Review: fix-release-workflow-run-tag-resolution

## Scenarios Reviewed

- `workflow_run` uses pre-bump `head_sha`: no longer used for tag lookup.
- Latest release tag exists on `origin/main`: resolved via merged semver tag query.
- No `v*` tags on `origin/main`: workflow sets `skip=true` and exits cleanly.
- Manual `push` to `refs/tags/v*`: still resolves tag from `GITHUB_REF_NAME`.

## Outcome

- Release workflow now resolves CI-created bump tags reliably.
- Graceful skip behavior is preserved for no-tag repositories.
