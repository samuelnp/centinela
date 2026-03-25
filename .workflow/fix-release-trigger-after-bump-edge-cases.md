# Edge-Case Review: fix-release-trigger-after-bump

## Scenarios Reviewed

- Tag created by CI with `GITHUB_TOKEN`: release now runs via `workflow_run`.
- Manual `v*` tag push: release still runs from `push` trigger.
- `workflow_run` success without tag on `head_sha`: release skips gracefully.
- `workflow_run` failure conclusion: release job does not run.
- Windows arm64 artifact path: naming and `.exe` extension stay valid.

## Outcome

- Release automation is now resilient for both manual and CI tag creation paths.
- Skip behavior prevents false failures when no release tag is present.
