# Feature Brief: Automate Semantic Version Releases

## Problem
Releases are currently manual. Version bumps, tags, and compiled artifacts can drift or be skipped, making installation via release assets unreliable.

## Goal
Automate semantic version bumping on pushes to `main`, build cross-platform binaries on tag creation, and publish artifacts to GitHub Releases for curl-based installation.

## Scope
- Add workflow to bump `Makefile` version based on conventional commits.
- Add workflow to build and publish release artifacts from `v*` tags.
- Add `scripts/install.sh` for release-based installation.
- Update docs with installer usage.

## Acceptance Criteria
- Push to `main` triggers semantic bump commit and `vX.Y.Z` tag.
- Tag workflow builds artifacts for Linux/macOS/Windows and uploads checksums.
- `scripts/install.sh` installs latest release binary and verifies checksum.
- Existing validate pipeline remains intact.
