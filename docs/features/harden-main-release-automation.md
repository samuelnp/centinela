# Feature Brief: Harden Main Release Automation

## Problem
Release automation exists, but we need it aligned to a single policy: pushes to `main` must bump `Makefile` version semantically and create release artifacts for all supported OS and architecture targets.

## Goal
Ensure every eligible push to `main` produces a semantic version bump, a `vX.Y.Z` tag, and a GitHub Release containing checksummed binaries for Linux, macOS, and Windows on `amd64` and `arm64`.

## Scope
- Keep semantic bumping based on conventional commits (`major`/`minor`/`patch`).
- Update `Makefile` `VERSION :=` from workflow logic.
- Build and publish 6 release artifacts from `v*` tags.
- Preserve validate pipeline behavior.

## Acceptance Criteria
- Push to `main` runs version workflow and updates `Makefile` version.
- Version workflow commits release bump and pushes `vX.Y.Z` tag.
- Tag workflow builds binaries for `linux/darwin/windows` x `amd64/arm64`.
- Release workflow publishes binaries plus `SHA256SUMS` to GitHub Releases.
- Workflow loop prevention remains in place for bot-originated commits.
