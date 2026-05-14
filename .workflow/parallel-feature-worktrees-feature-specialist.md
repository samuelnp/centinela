### Feature-Specialist Report: parallel-feature-worktrees
**Date:** 2026-05-14

#### Behavior Summary

When a project opts into `workflow.use_worktrees`, `centinela start <feature>`
provisions an isolated git worktree at `.worktrees/<feature>/` with its own
branch and its own `.workflow/` state, so multiple feature workflows can run
concurrently without colliding. The wizard (new projects) and `centinela
migrate` (existing projects) keep tool ignore lists (`.gitignore`,
`.eslintignore`, `.prettierignore`, `.dockerignore`, `.rgignore`, and
`tsconfig.json` `exclude`) in sync so linters, formatters, type-checkers, and
search tools do not double-scan worktrees. A new `centinela merge <feature>`
command runs out-of-band (it is NOT a sixth workflow step): it first performs
a spec-conflict pre-check by reading `specs/*.feature` across every active
worktree; then it performs `git merge --no-ff`; then it runs `centinela
validate` against the merged tree. The Merge Steward agent is invoked on any
text conflict OR on post-merge validate failure (semantic conflict). The
Steward must escalate to the user — with the proposed diff and a reason —
whenever its confidence is not high; no silent resolutions are permitted.
On full success the worktree is removed; on any failure or escalation the
worktree is left untouched so the user can inspect or resume.

#### Gherkin Scenarios

All scenarios live in [`specs/parallel-feature-worktrees.feature`](../specs/parallel-feature-worktrees.feature):

- **Start provisions a worktree when use_worktrees is enabled** — `centinela
  start alpha` creates `.worktrees/alpha/` with a matching branch and a
  per-worktree `.workflow/` directory.
- **Start runs in the main checkout when use_worktrees is disabled** —
  back-compat: no `.worktrees/` directory, single-checkout flow unchanged.
- **Wizard syncs tool ignore lists for new projects** — `.gitignore`,
  `.eslintignore`, `.prettierignore`, `.dockerignore`, `.rgignore`, and
  `tsconfig.json` `exclude` all gain a `.worktrees/` entry; idempotent.
- **Migrate syncs tool ignore lists for existing projects** — `centinela
  migrate` is the back-compat path for the same ignore-list sync; safe to
  re-run.
- **Clean merge when git applies cleanly and validate passes** — Steward is
  NOT invoked; the worktree is removed after success.
- **Text conflict invokes the Merge Steward** — Steward is invoked with the
  conflicted paths and the feature spec; evidence written.
- **Semantic conflict after a clean text merge invokes the Steward** — clean
  git merge but `centinela validate` fails -> Steward triggered with the
  failing validate output and the feature spec.
- **Spec conflict across in-flight worktrees is detected before merging** —
  the pre-check reads `specs/*.feature` from every active worktree and
  blocks the merge when two scenarios assert different observable outcomes
  for the same Given context; output names both files and the conflicting
  scenario; no commits land on main.
- **Merge Steward escalates uncertain resolutions** — non-high-confidence
  resolutions exit without touching main and surface the proposed diff plus
  the reason; evidence JSON records the escalation and the proposed diff.
- **Restarting a feature with an existing worktree resumes in place** —
  decided contract: reuse the existing worktree, do not error, preserve
  existing `.workflow/` state.
- **Merge fails fast when the main working tree is dirty** — non-zero exit
  before invoking git merge; the worktree is not touched.

#### UX States

This feature has no UI surface; all states are CLI/terminal output. The
table below lists the observable terminal state per phase.

| State    | Trigger                                                                 | Surface                                                                                              |
|----------|-------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------|
| loading  | `centinela start` provisioning a worktree, or `centinela merge` running | Terminal prints the worktree path being created and the current merge phase (pre-check / merge / validate / steward). |
| empty    | n/a (no list-style UI affected)                                          | n/a                                                                                                  |
| error    | Dirty main on merge, spec conflict pre-check fails, Steward escalates    | Non-zero exit; actionable terminal message naming the offending file(s) and the recommended next step; worktree left intact for inspection. |
| success  | Clean merge + validate pass, or Steward applies a high-confidence fix   | Terminal prints the merged commit summary, the validate summary, and confirms the worktree at `.worktrees/<feature>/` was removed. |

#### Out-of-Scope

- Per-language toolchain isolation (separate `node_modules`, venv, Cargo
  target). Tracked as v2.
- Nested worktrees and worktrees outside `.worktrees/`.
- IDE integration helpers (VSCode multi-root, JetBrains attach).
- Automatic `main`-into-worktree rebase from the agent — manual for v1.
- Multi-repo or submodule merges.
- A new sixth workflow step. `centinela merge` runs out-of-band on the
  result of `docs`, not as part of the 5-step sequence.

#### Handoff

- **Next role:** senior-engineer.
- **Open clarifications (forwarded from big-thinker):**
  - Should `internal/worktree/` own the spec-conflict detector, or does it
    warrant a new `internal/specs/` package? feature-specialist preference:
    keep in `internal/worktree/` for v1; promote only if a second reader
    emerges.
  - Confirm `cmd/centinela/merge.go` stays a thin orchestrator and all
    decision logic lives in `internal/worktree/merger.go`.
  - Confirm `merge-steward` runs outside the 5-step workflow (the spec
    treats it as out-of-band on `centinela merge`).
  - Tune the Steward "high confidence" threshold during phase 2; the spec
    pins only the observable behavior ("escalate when not high-confidence").
