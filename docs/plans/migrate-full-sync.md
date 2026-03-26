# Plan: migrate-full-sync

## Scope

Implement full-sync migration so users can preview and apply docs + setup upgrades
without rerunning `centinela init`.

## Tasks

1. Extend migration domain model to support setup artifacts and manual-review actions.
2. Add setup migration planning/apply logic for Claude hooks, OpenCode config,
   plugin file, and AGENTS.md.
3. Add `centinela migrate` top-level execution with `--apply` for unified full sync.
4. Add `centinela migrate setup` with `--apply` and `--agent` scoping.
5. Extend migration hook output to include setup migration warnings.
6. Update migration UI and README documentation.
7. Add unit, integration, and acceptance tests for preview/apply/idempotency and
   unmanaged-file manual-review behavior.

## Non-Goals

- Migrating arbitrary user-authored files outside Centinela-managed artifacts.
- Interactive confirmation prompts inside CLI commands.
