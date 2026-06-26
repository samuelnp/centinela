# lean-evidence-footprint — senior-engineer

## Files Touched

- `.gitignore` — added a documented block:
  `.workflow/*.json` + `!.workflow/roadmap.json` + `.workflow/*.lock`.
- Index only (no working-tree deletion): `git rm --cached` of 752
  already-tracked plumbing files (`*.json` excl `roadmap.json`, all `*.lock`).

No Go source changed — this is repository-configuration only.

## Architecture Compliance

N/A to layer rules (no application code). The change is git metadata. The
governing constraint is the evidence contract: machine `.json` is read only
locally by `centinela complete`; gitignore does not remove local files, so
the contract is intact. `roadmap.json` (required by `hook_setup.go`) is
preserved by the `!` negation.

## Type-Safety Notes

N/A — no code. Verified behavior empirically with `git check-ignore`:
`demo-big-thinker.json`, `demo.json`, `demo-big-thinker.lock` → ignored;
`roadmap.json`, `demo-big-thinker.md` → tracked.

## Trade-Offs

- Chose gitignore + `git rm --cached` over adding `os.Remove` to the lock
  release closure (`internal/evidence/lock.go:49`) — the latter introduces
  an unlink-after-unlock race. Local 0-byte locks remain (untracked,
  cosmetic).
- Kept the per-feature root `<feature>.json` ignored too (machine-only),
  rather than maintaining a brittle per-role allowlist; `roadmap.json` is
  the single explicit exception.

## Handoff

→ qa-senior: add unit (gitignore contains patterns), integration
(`git check-ignore` matrix in a temp repo), and acceptance (`git status`
ignores json/lock, tracks md + roadmap.json) tests; record edge cases.
