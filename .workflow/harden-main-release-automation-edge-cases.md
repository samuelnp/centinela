# Edge-Case Review: harden-main-release-automation

## Scenarios Reviewed

- No prior `v*` tag exists: workflow falls back to commit log from `HEAD`.
- Push actor is `github-actions[bot]`: bump job is skipped to prevent loops.
- `Makefile` lacks `VERSION :=` key: workflow fails fast with explicit error.
- Multiple pushes on `main`: concurrency group serializes version bump runs.
- Release matrix includes `windows/arm64`: artifact naming and extension remain valid.

## Outcome

- Semantic bumping remains conventional-commit driven.
- First-release path is now safe without invalid git ranges.
- Release coverage remains `linux/darwin/windows` x `amd64/arm64` with checksums.
