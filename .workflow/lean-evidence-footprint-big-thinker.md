# lean-evidence-footprint — big-thinker

## Problem

`.workflow/` is committed (1,419 tracked files). Per feature, ~15–25 files
land in git: a `-<role>.json` + `-<role>.md` pair per role, a `.lock` per
role, plus artifacts. The `.json` is a machine contract read only by
`centinela complete` while a feature is active — inert after merge. The
`.lock` files are never deleted (`lock.go` release never `os.Remove`s),
so 212 zero-byte locks have accumulated. Result: bloated PR diffs and repo,
i.e. a token burner. The `-<role>.md` narratives, by contrast, are wanted
(reviewer + LLM KB).

## Scope

Keep `.md` committed. Stop committing `.workflow/*.json` (except
`roadmap.json`) and `.workflow/*.lock` via `.gitignore`, and retroactively
`git rm --cached` the 747 already-tracked plumbing files. No Go behavior
change.

## Dependencies & Assumptions

- Verified: `centinela validate` (CI) runs gates only; only `centinela
  complete` reads evidence `.json`, locally — `validate.go:30`,
  `orchestration/evidence.go:25`.
- Verified: rehydration reads only `roadmap.json` — `hook_session.go:24`,
  required by `hook_setup.go:88`. It must stay tracked.
- Verified: all other `.workflow/*.json` are feature-scoped and inert
  post-merge. Gitignore does not delete local files, so the local-evidence
  contract for `complete` is preserved.

## Risks

- Low. No code reads committed evidence for a merged feature today. A
  future such reader would break — covered by an acceptance test that runs
  `validate`/`complete` against gitignored-but-local evidence.
- Deliberately NOT adding `os.Remove` to lock release: unlink-after-unlock
  race. Local locks stay (untracked, cosmetic).

## Rollout

Single PR, no migration. History retains untracked files; working `main`
sheds ~747 files.

## Handoff

→ feature-specialist: encode the ignore/keep matrix and retroactive
cleanup as acceptance scenarios; confirm `.md` and `roadmap.json` parity.
