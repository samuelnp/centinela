# Big-Thinker Report: parallel-feature-worktrees

**Date:** 2026-05-14

## Problem

Centinela today drives one feature at a time through the 5-step
workflow against the project's single working tree. As soon as a user
(or an agent) wants two features in flight, they collide: file writes
conflict, the active branch is shared, and hooks see ambiguous workflow
context because there is only one `.workflow/` directory and one cwd.
Merging finished work back to `main` is fully manual today — there is
no command, no agent help, and no detection of cross-feature spec or
semantic conflicts before the developer hits them. The opportunity is
to lean on `git worktree` for cheap isolation, and to add a small,
opinionated merge agent that uses the feature's own spec and diff to
attempt smart resolution while always escalating uncertainty to the
user. Centinela's "enforce best practices for LLM-driven workflows"
positioning makes parallel agents the natural next mode.

## Scope

**In:**
- Per-feature worktree at `.worktrees/<feature>/`, opt-in via
  `workflow.use_worktrees`. Wizard defaults ON for new projects;
  `centinela migrate` adds the flag (defaulted off) and patches
  ignore lists for existing projects.
- Provisioning/cleanup wired into `centinela start` and the merge
  command. State (`.workflow/`) is per-worktree — self-contained.
- New `centinela merge <feature>` (HYBRID): spec-conflict pre-check
  across active worktrees, then `git merge --no-ff`, then
  `centinela validate` on the merged tree. Steward is invoked on
  text conflict OR validate failure.
- New `merge-steward` orchestration role with its own prompt and
  evidence contract entry. Handles git text, semantic, and spec
  conflicts. Uncertain proposals escalate; no silent resolutions.
- Hook awareness: `hook_prewrite.go` / `hook_postwrite.go` resolve
  the feature from cwd when cwd is inside `.worktrees/<feature>/`.
- Ignore-list sync: `.gitignore`, `.eslintignore`,
  `.prettierignore`, `.dockerignore`, `.rgignore`/`.ripgreprc`,
  `tsconfig.json` "exclude", and framework scan globs (jest/vitest/
  tailwind/vite). The wizard and `centinela migrate` own this.

**Out (v1):**
- Per-language toolchain isolation (separate `node_modules`/venv/
  cargo target dirs per worktree). Tracked as v2.
- Nested worktrees and worktrees outside `.worktrees/`.
- IDE integration helpers (VSCode multi-root, JetBrains attach).
- Automatic main-branch sync into long-running worktrees (agent-driven
  rebase). Manual for v1.
- Multi-repo and submodule merges.

## Dependencies & Assumptions

- Targets are projects whose root is a git repo. We add a no-op /
  warning path for non-git directories but worktree mode is off there.
- `internal/orchestration/` already validates per-role JSON. Adding
  `merge-steward` is a known extension surface — the contract doc
  enumerates roles explicitly, and the validator reads from a typed
  `Role` set in `evidence.go`.
- The hook entry points (`hook_prewrite.go`) already compute cwd via
  `os.Getwd()`. Resolving the feature from a `.worktrees/<feature>/`
  prefix is a small addition; the rest of the hook pipeline is cwd-
  agnostic because `.workflow/` paths are relative.
- A new `internal/worktree/` package fits the n-tier rules: it is
  domain logic (workflow concern), consumable from `cmd/centinela/`
  and from `internal/workflow/` if needed. Flag this for the
  senior-engineer to confirm boundary (e.g. should the spec-conflict
  detector live in `internal/worktree/` or in a new `internal/specs/`
  package).
- Acceptance tests already shell out to `git` via the project's tmp
  repo fixture helpers (used by `gitdiff` tests), so worktree tests
  have a precedent to copy.
- Phase 2+3 (Merge Steward) depends on agent invocation via the
  existing `Agent` tool — same pattern as gatekeeper / production
  readiness. Steward output must follow the evidence contract.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Hooks (PreToolUse) misclassify cwd inside `.worktrees/X` and either block legitimate writes or apply the wrong feature's policy. | High | Medium | Add explicit worktree-prefix detection in `evalPrewrite`; integration test that writes a file inside a worktree and asserts the resolved feature. |
| `.workflow/` state contract becomes ambiguous (shared vs per-worktree). | High | Medium | Decided: per-worktree, fully self-contained. Document explicitly in the brief and in `workflow-enforcement.md`. Roadmap aggregation in the main checkout is read-only across worktrees. |
| `centinela validate` runs from a worktree and a user `[validate] command` assumes repo-root cwd. | Medium | Medium | Document that commands are run from cwd of the worktree (which is the project root for the feature branch). Smoke-test go test, npm test patterns. |
| Steward silently resolves a conflict the user would have wanted to review. | High | Low (by design) | Prompt enforces "uncertain -> escalate"; evidence JSON requires explicit `confidence` + the resolution diff so we can audit. Acceptance test for the escalation path. |
| Two worktrees both edit shared scaffold assets (e.g. `docs/architecture/*`) and produce non-overlapping text but contradictory intent. | Medium | Medium | Spec-conflict pre-check is the v1 net; senior-engineer should add a configurable "shared-asset" allow-list later. Out of scope for v1 if it bloats the surface — listed as a v2 candidate. |
| Tooling double-scans `.worktrees/` (lint/typecheck/test runners). | Medium | High without mitigation | Wizard and `centinela migrate` write the ignore-list entries listed in Scope. Idempotent. |
| Disk/branch sprawl from forgotten worktrees. | Low | Medium | `centinela status` lists active worktrees with last-touched time; `centinela merge` removes the worktree on success; add `centinela worktree prune` later if needed (v2). |
| Existing CI workflows expect `git checkout main` semantics, break under worktrees. | Low | Low | Worktrees are local-only; CI continues to clone normally. No CI change required. |
| Backward incompatibility for projects already on Centinela. | High | Low | Flag defaults off; `centinela migrate` is opt-in. No code path runs differently until the flag is on. |
| Hook injection of `[workflow: <feature>]` tag picks the wrong feature when cwd briefly leaves the worktree (e.g. agent runs `cd ..`). | Medium | Low | PostToolUse resolves feature from the workflow state, not cwd, so this is naturally safe; document and add a regression test. |

## Rollout

The smallest correct slice is **worktree provisioning + the config
flag + hook awareness + ignore-list sync**, with manual merge still
performed by the user. This is independently shippable, immediately
useful for parallel workflows, and has no dependency on the Steward
agent. Steward and conflict detection follow once that foundation is
proven.

- **Phase 1 — Isolation (ship first)**: `internal/worktree/` package,
  `use_worktrees` flag, `centinela start` provisioning, hook cwd
  resolution, ignore-list sync in wizard + migrate, `centinela status`
  surfacing. Manual merge. Unit + integration + acceptance tests for
  parallel feature start. **This is the minimum viable feature.**
- **Phase 2 — Merge Steward for text conflicts**: `centinela merge`
  command, `merge-steward` role + prompt + evidence-contract entry,
  validator extension, acceptance scenarios for clean merge and
  text-conflict-with-Steward.
- **Phase 3 — Semantic + spec conflicts**: post-merge `validate`
  trigger for Steward, `internal/worktree/spec_conflicts.go` pre-check,
  acceptance scenarios for both. These two extensions reuse the Phase
  2 plumbing; splitting them out de-risks the agent prompt by letting
  Phase 2 prove the integration on a narrower input class first.

## Handoff

- **Next role:** feature-specialist.
- **Outstanding questions for the senior-engineer (flag in the plan):**
  - Should `internal/worktree/` own the spec-conflict detector, or
    does that warrant a new `internal/specs/` package? My instinct is
    worktree for v1; specs only if a second reader emerges.
  - Confirm `cmd/centinela/merge.go` stays a thin orchestrator; all
    decision logic should live in `internal/worktree/merger.go`.
  - Validator extension: should `merge-steward` be a new step
    (`merge`) or run outside the 5-step workflow? Recommendation:
    outside — it operates on the result of `docs`, not as a step.
- **Outstanding questions for the user (none blocking):**
  - Default branch name for `git worktree add -b`. Recommend: feature
    slug (same as the workflow feature name).
  - Confidence threshold for Steward auto-apply vs escalate. Recommend:
    require a unanimous "high confidence" signal from the prompt; any
    other state escalates. Tune in Phase 2.
</content>
</invoke>