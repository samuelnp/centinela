# Plan: Parallel Feature Worktrees

1. Add `internal/worktree/` package — `Create`, `Remove`, path
   helpers, and "are we inside a worktree" detection. All files <=100
   lines with unit tests against a tmp git repo fixture.
2. Add `Workflow.UseWorktrees bool` (toml `use_worktrees`) to
   `internal/config/`. Default `false`; wizard writes `true` for new
   projects.
3. Wire `centinela start <feature>`: when the flag is on, provision
   `.worktrees/<feature>/`, branch, and chdir into it before writing
   `.workflow/<feature>.json`.
4. Sync ignore lists from the wizard and `centinela migrate`:
   `.gitignore`, `.eslintignore`, `.prettierignore`, `.dockerignore`,
   `.rgignore`/`.ripgreprc`, `tsconfig.json` "exclude", and framework
   scan globs declared in PROJECT.md. Idempotent.
5. Update `hook_prewrite.go` / `hook_postwrite.go` to resolve the
   feature from cwd when cwd is inside `.worktrees/<feature>/`, and
   read `.workflow/` from that worktree.
6. Add `centinela merge <feature>` (thin cmd) + merge logic in
   `internal/worktree/merger.go`. Run spec-conflict pre-check, then
   `git merge --no-ff`, then `centinela validate` against the merged
   tree. Invoke the Merge Steward on text conflict OR validate failure.
7. Author `docs/architecture/merge-steward-prompt.md` and add the
   `merge-steward` role to the orchestration evidence validator
   (`internal/orchestration/`) plus an entry in
   `docs/architecture/evidence-contract.md`.
8. Add `internal/worktree/spec_conflicts.go` to detect contradictory
   `specs/*.feature` content across active worktrees before merging.
9. Surface worktree status in `centinela status` and the roadmap UI.
10. Tests — unit for `internal/worktree/`; integration for hook +
    worktree resolution; acceptance scenarios for parallel features,
    clean merge, text conflict (stubbed Steward), semantic conflict
    (validate fails post-merge), and spec contradiction.
11. Documentation updates (docs step): README workflow section,
    `workflow-enforcement.md`, `centinela.toml` reference, and
    evidence-contract entry for `merge-steward`.
</content>
</invoke>