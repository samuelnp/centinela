# Plan: Automate Semantic Version Releases

## Scope
Implement CI workflows and installer script for automatic semantic releases on `main`.

## Work Items
1. Add `.github/workflows/version-bump.yml`:
   - Trigger: `push` on `main`
   - Determine major/minor/patch from commit messages since last tag
   - Update `Makefile` `VERSION :=`
   - Commit with `[skip ci]` and push `vX.Y.Z` tag
2. Add `.github/workflows/release.yml`:
   - Trigger: `push` tags `v*`
   - Build binaries for linux/darwin/windows (amd64 + arm64 where applicable)
   - Package artifacts + `SHA256SUMS`
   - Publish GitHub Release with generated notes
3. Add `scripts/install.sh`:
   - Resolve latest release
   - Download platform artifact + checksum file
   - Verify checksum and install binary to `${INSTALL_DIR:-$HOME/.local/bin}`
4. Update `README.md` install docs for curl workflow.

## Validation
- `go test ./...`
- `go run ./cmd/centinela validate`

## Constraints
- Bump workflow runs only on `main`.
- Prevent CI loops with `[skip ci]` and actor guard.
