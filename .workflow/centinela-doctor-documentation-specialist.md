# Documentation-Specialist Report — centinela-doctor

Internal-surface (right-sized) docs step.

## Outputs
- `.workflow/centinela-doctor-changelog.md` — one-line feat changelog.
- Regenerated `docs/project-docs/index.html`.

## User-facing note
`centinela doctor` is read-only (exit 1 only on ERROR); `centinela doctor --fix` applies only safe idempotent repairs. Destructive cleanups (abandoned worktrees, stale `.workflow` state) are reported with the exact command, never auto-applied. Run it anytime — no active workflow required; from a worktree it resolves and checks the canonical repo.

Handoff → complete.
