# Feature Brief: Parallel Feature Worktrees

## Problem

Centinela today drives one feature through `plan -> code -> tests ->
validate -> docs` against the active checkout. When an LLM (or a team)
wants to run multiple feature workflows in parallel, the single working
tree is the bottleneck: file writes collide, branches step on each other,
and hooks see ambiguous workflow context. There is also no automated path
for merging finished features back to `main`; the merge step is manual
and conflict-prone when several features in flight touch overlapping
code or contradict each other at the spec level.

## Goal

Provision a per-feature git worktree (`.worktrees/<feature>/`) when a
project opts in, so feature workflows run in fully isolated checkouts.
Add a new `centinela merge <feature>` command and a Merge Steward
subagent that handles the merge back to `main` — attempting smart
conflict resolution from the feature's spec and diff, and escalating to
the user when uncertain.

## Scope

- Opt-in via `workflow.use_worktrees` in PROJECT.md / centinela.toml.
  Wizard defaults it ON for new projects. Existing projects opt in
  through `centinela migrate`. Backward-compatible when the flag is
  false — the single-checkout flow is untouched.
- Worktrees live at `.worktrees/<feature>/` inside the repo. Wizard and
  `centinela migrate` keep `.gitignore`, `.eslintignore`,
  `.prettierignore`, `.dockerignore`, `.rgignore`/`.ripgreprc`,
  `tsconfig.json` "exclude", and framework scan globs (jest/vitest/
  tailwind/vite) in sync so tooling does not double-scan the worktree.
- New `centinela merge <feature>` command. HYBRID merge: git attempts
  the merge first. On text conflict OR if `centinela validate` fails
  on the merged tree, the Merge Steward agent is invoked. Spec/contract
  contradictions across in-flight feature branches are detected BEFORE
  the merge by reading `specs/*.feature` across active worktrees.
- New `merge-steward` orchestration role. Handles three conflict
  classes: git text, semantic (clean merge but validate fails), and
  spec/contract (contradictory Gherkin). Uncertain proposals MUST
  escalate to the user — no silent resolutions.
- Hooks (`hook_prewrite.go`, `hook_postwrite.go`) recognise when cwd is
  inside `.worktrees/<feature>/` and resolve workflow state from the
  worktree's own `.workflow/` directory.

## Edge Cases

- `.workflow/` is per-worktree (each worktree is self-contained). No
  shared state across worktrees.
- Hook cwd is inside `.worktrees/<feature>/` — feature resolution must
  not fall back to the main checkout's roadmap.
- `centinela validate` runs the test suite from the worktree; commands
  in `[validate] commands` must be cwd-relative, not repo-root-relative.
- A clean text merge that breaks the build is a semantic conflict —
  Steward must run after validate, not only on `git merge` failure.
- Two in-flight features modifying the same Gherkin scenario are flagged
  pre-merge; the second merge is blocked until the conflict is resolved.
- Opting back out (`use_worktrees = false`) with active worktrees is a
  warning, not an error — existing worktrees keep working until manually
  removed.

## Out of Scope (v1)

- Per-worktree language toolchain isolation (separate `node_modules` /
  venv / Cargo target). Tracked as v2.
- Nested worktrees, or worktrees outside `.worktrees/`.
- IDE integration helpers (VSCode multi-root, JetBrains attach).
- Automatic `main` sync into long-running worktrees (rebase from the
  agent). Manual for v1.
- Multi-repo or submodule merges.
</content>
</invoke>