# precommit-and-pr-gate — documentation-specialist

Internal-surface (right-sized) docs step.

## KB Pages

No standalone KB page — two new commands on the existing gate surface,
documented via the brief, plan, and regenerated project docs.

## project-docs Entries

- `.workflow/precommit-and-pr-gate-changelog.md` — one-line `feat` changelog.
- Regenerated `docs/project-docs/index.html` (picks up the brief, plan, changelog).

## User-facing note

Two new ways to fire Centinela's mechanical gates:

**Pre-commit** — `centinela precommit` runs the gates against your **staged**
changes and exits non-zero (blocking the commit) on a fail-severity gate. It is
fast: it skips the cross-compile/build gate by default (`[precommit] skip_build
= true`). Install it once with `centinela precommit install`, which writes a
marker-delimited, idempotent `.git/hooks/pre-commit` that preserves any existing
hook (`centinela precommit uninstall` removes only its block).

**PR gate** — `centinela pr-gate` runs the gates against a PR's changed files
and prints a deterministic Markdown verdict (per-gate ✅/❌ + details) with a
pass/fail exit code. The shipped GitHub Actions workflow posts/updates that
verdict as a single PR comment on each run. Configurable via `[pr_gate]`
(`fail_on_warning`).

Custom gates and the audit-baseline ratchet participate in both surfaces exactly
as they do in `centinela validate`. The binary stays network-free — the GitHub
posting lives in CI (`gh pr comment`).

## Outcome

Docs generated and validated. Handoff → complete.
