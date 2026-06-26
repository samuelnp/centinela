# Plan: lean-evidence-footprint

## Summary

Stop committing machine-only `.workflow/` evidence (`*.json` except
`roadmap.json`, and all `*.lock`), keep the readable `-<role>.md`
narratives, and retroactively untrack the 747 files already in the index.
No Go behavior changes.

## Design decisions (verified)

| Question | Verified answer | Source |
|----------|-----------------|--------|
| Does CI's `centinela validate` read evidence `.json`? | No — gates only | `cmd/centinela/validate.go:30-80` |
| What reads evidence `.json`? | Only `centinela complete` → `ValidateStep` → `ValidateEvidence`, locally | `internal/orchestration/evidence.go:25` |
| Does the rehydration hook read feature evidence? | No — only `roadmap.json` | `cmd/centinela/hook_session.go:24` |
| Is `roadmap.json` required-committed? | Yes | `cmd/centinela/hook_setup.go:88` |
| Are all other `.workflow/*.json` feature-scoped? | Yes (root `<feature>.json` + `<feature>-<role>.json`) | `git ls-files` audit |
| Are `.md` companions read programmatically? | Only during delivery (PR body/changelog); existence-checked otherwise | `internal/evidence/companion.go:32`, `internal/workflow/validate.go:18` |
| Why not delete `.lock` on release? | Unlink-after-unlock race; accumulation is intentional | `internal/evidence/lock.go:49`, `repair_race_test.go` |

## Changes

### 1. `.gitignore`
Append a documented block:
```gitignore
# Per-feature evidence is machine plumbing — generated and validated locally
# by `centinela complete`, never read after a feature merges. Keep the readable
# -<role>.md companions (reviewer + LLM knowledge base) and roadmap.json
# (required project state, read by the rehydration hook) committed.
.workflow/*.json
!.workflow/roadmap.json
.workflow/*.lock
```

### 2. Retroactive untracking
```bash
git rm --cached -q $(git ls-files '.workflow/*.json' | grep -v '^.workflow/roadmap.json$')
git rm --cached -q $(git ls-files '.workflow/*.lock')
```
Removes 535 `.json` + 212 `.lock` = **747 files** from the index; local
copies untouched. Committed `.workflow/` footprint drops ~1,419 → ~672
(the `.md` narratives + `roadmap.json`).

## Acceptance (see specs/lean-evidence-footprint.feature)

- New evidence `.json`/`.lock` written during a workflow is gitignored.
- `roadmap.json` stays tracked.
- `-<role>.md` companions stay tracked.
- `git ls-files '.workflow/*.lock'` and `'.workflow/*.json'` (excl roadmap)
  return empty after cleanup.
- `centinela validate` and `centinela complete` still pass with evidence
  present locally but untracked.

## Test strategy

- **Unit/integration**: a test asserting the `.gitignore` patterns match
  a sample `.workflow/<feature>-<role>.json` / `.lock` and do **not** match
  `roadmap.json`, plus that `.md` is unaffected. Drive via Go using
  `git check-ignore` against a temp repo seeded with the real `.gitignore`
  block, or a pure matcher test over the patterns.
- **Acceptance**: a binary/git-driven test that, in a temp repo with the
  shipped `.gitignore`, creates evidence files and asserts `git status
  --porcelain` ignores the json/lock but tracks the md and roadmap.json.

## Rollout

Single PR. No migration. Existing merged features keep their `.md`
narratives in history; their now-untracked `.json`/`.lock` remain in git
history (recoverable) but leave the working `main`.

## Risks

- **Low**: a future code path that reads committed evidence for a merged
  feature would break — none exists today (verified). Mitigated by the
  acceptance test exercising `validate`/`complete` with untracked evidence.
- Local `.lock` accumulation persists (cosmetic, untracked). Could be
  swept by `centinela doctor` in a later feature if desired.
