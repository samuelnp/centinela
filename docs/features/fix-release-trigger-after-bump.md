# Feature Brief: Fix Release Trigger After Version Bump

## Problem
The release workflow only triggers on `push` of `v*` tags. Tags created by the version-bump workflow use `GITHUB_TOKEN`, so GitHub does not trigger a second workflow from that push.

## Goal
Ensure releases always run after automated version bumps, while still supporting direct manual `v*` tag pushes.

## Scope
- Update release workflow triggers to support post-bump execution.
- Resolve tag name reliably for both trigger paths.
- Keep existing cross-platform artifact build and checksum publishing.

## Acceptance Criteria
- Successful Version Bump workflow run triggers release publishing.
- Manual `v*` tag push still triggers release publishing.
- Release builds binaries for `linux/darwin/windows` x `amd64/arm64`.
- Release assets include `SHA256SUMS` and generated notes.
