# docs-migration-managed-docs

## Problem

Centinela scaffold docs can become stale across versions, but there is no guided
upgrade path, so users keep outdated rules and prompts.

## User Stories

- As a maintainer, I want Centinela to detect outdated managed docs.
- As a user, I want a preview of proposed changes before files are modified.
- As a user, I want migration to preserve my custom sections and keep blocks.

## Acceptance Criteria

- Versioned headers are added to managed markdown templates.
- `centinela migrate docs` shows a non-destructive migration plan.
- `centinela migrate docs --apply` updates outdated files to latest template.
- Hook context asks for user approval before applying migration.

## Edge Cases

- Files without headers are treated as legacy and planned for migration.
- Missing managed files are shown as create actions.
- Keep blocks and non-template sections are preserved on apply.
