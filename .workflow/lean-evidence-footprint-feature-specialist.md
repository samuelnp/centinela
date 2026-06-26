# lean-evidence-footprint — feature-specialist

## Behavior Summary

After this change, git ignores `.workflow/*.json` and `.workflow/*.lock`,
with a single exception for `.workflow/roadmap.json`. The readable
`-<role>.md` companions stay tracked. Files already committed in those
ignored classes are removed from the index (kept on disk).

## Acceptance Criteria (Gherkin)

See `specs/lean-evidence-footprint.feature`. Key scenarios:
- `demo-big-thinker.json`, `demo.json`, `demo-big-thinker.lock` → ignored.
- `roadmap.json` → NOT ignored, stays tracked.
- `demo-big-thinker.md` → NOT ignored.
- After cleanup: `git ls-files '.workflow/*.lock'` empty; `*.json`
  excludes everything but `roadmap.json`; local files still present.
- `centinela validate` passes and `centinela complete` still reads local
  (gitignored) evidence.

## UX States

N/A — no user-facing surface. This is repository/CLI plumbing. The only
observable "UX" is a smaller `git status` and PR diff.

## Edge Cases

- Per-feature root `<feature>.json` is also ignored (machine-only); only
  `roadmap.json` is kept.
- `-<role>.md` companions must remain tracked — the explicit value to keep.
- CI's `centinela validate` runs gates only and never reads evidence
  `.json`, so untracking is safe on feature branches.
- Gitignore never deletes local files, so `centinela complete`'s local
  evidence read is unaffected.

## Out-of-Scope

- Modifying lock semantics (`os.Remove` on release) — race risk.
- Collapsing the `.json`/`.md` pair into one file — separate feature.

## Handoff

→ senior-engineer: implement the `.gitignore` block + retroactive
`git rm --cached`; keep changes Go-behavior-neutral.
