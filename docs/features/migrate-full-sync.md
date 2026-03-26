# migrate-full-sync

## Problem

Centinela can migrate managed docs, but integration assets (hooks, OpenCode plugin,
and setup files) still depend on rerunning `centinela init`, which is coarse-grained
and does not provide a clear migration preview.

## User Stories

- As a maintainer, I want one migration command that upgrades all managed assets.
- As a user, I want preview mode before any files are modified.
- As a user, I want apply mode to create or update only Centinela-managed artifacts.

## Acceptance Criteria

- `centinela migrate` shows a unified preview for docs + setup assets.
- `centinela migrate --apply` performs full sync for docs + setup assets.
- `centinela migrate setup` supports `--agent claude|opencode|both`.
- Claude/OpenCode setup files are updated without clobbering unrelated user config.
- Hook migration prompt includes setup migration needs, not only docs.

## Edge Cases

- Missing setup files are shown as create actions and created on apply.
- Existing managed files with older signatures are updated to current versions.
- User-customized unmanaged files are reported for manual review.
