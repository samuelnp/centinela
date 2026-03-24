# Plan: docs-migration-managed-docs

## Scope

Implement version-aware migration for Centinela-managed markdown docs with
preview and apply modes, plus hook-based user confirmation guidance.

## Tasks

1. Add migration domain package with:
   - managed doc manifest and target versions
   - first-line header parse/write
   - detection, planning, and apply logic
2. Add `centinela migrate docs` command with `--apply` support.
3. Add `centinela hook migrate` and wire it into UserPromptSubmit hooks.
4. Add UI render for migration-needed prompt and plan summary.
5. Add version headers to scaffold-managed markdown assets.
6. Add tests (unit, integration, acceptance) for detection, plan/apply, and hook UX.

## Non-Goals

- Migrating user-authored feature docs in `docs/features/`.
- Interactive in-terminal confirmation prompts.
