---
feature: parallel-feature-worktrees
summary: Run several Centinela feature workflows at the same time, each in its own isolated checkout, and merge them back to main with conflict detection and a Merge Steward.
audience: end-user
status: done
---

## What it does
When you opt in, every feature you start gets its own private working copy of the project under `.worktrees/<feature>/`, with its own branch and its own workflow state. That means two (or more) features — or two agents — can run the full plan → code → tests → validate → docs cycle at the same time without overwriting each other's files or fighting over the branch. A new `centinela merge <feature>` command brings a finished feature back to `main`: it checks for clashes first, lets git do the merge, re-runs validation, and pulls in a Merge Steward when a conflict needs judgement. The whole thing is opt-in, so projects that don't turn it on keep working exactly as before.

## When you'd use it
Reach for this whenever you want more than one feature in flight at once — for example, running parallel agents on independent features, or working on a second feature while a long one validates. You enable it by setting `workflow.use_worktrees = true` (new projects created by the wizard get it on by default; existing projects opt in with `centinela migrate`). After that, `centinela start` isolates each feature automatically and `centinela merge` is how you land it.

## How it behaves
- With worktrees enabled, `centinela start alpha` creates an isolated checkout at `.worktrees/alpha/`, checks out an `alpha` branch inside it, and stores that feature's workflow state under `.worktrees/alpha/.workflow/`.
- With worktrees disabled, nothing changes: no `.worktrees/` directory is created and the feature's state stays at the repo root, exactly like before this feature existed.
- When the wizard sets up a new project with worktrees on, it adds a `.worktrees/` entry to `.gitignore`, `.eslintignore`, `.prettierignore`, `.dockerignore`, and `.rgignore`, plus a `.worktrees` exclude in `tsconfig.json`, so linters, formatters and search tools don't scan the worktrees twice. Re-running the wizard never duplicates those entries.
- Running `centinela migrate` on an existing project and opting in applies the same ignore-list updates, and the command is safe to run again with no further changes.
- `centinela merge gamma` on a clean main, when git merges with no conflict and validation passes, lands the feature, leaves the Merge Steward out of it, and removes the `.worktrees/gamma/` checkout once the merge succeeds.
- If merging produces a text conflict, the merge stops and points you at the Merge Steward: a `.workflow/<feature>-merge-steward.md` hint is prepared with the conflicted files and the feature's spec so the conflict can be resolved deliberately rather than guessed at.
- If git merges cleanly but validation then fails on the merged tree, that "silent" breakage is treated as a conflict too — the merge stops and surfaces the failing validation output and the spec for the Steward, instead of shipping a broken main.
- Before any merge runs, Centinela compares the Gherkin specs across all in-flight worktrees; if two features assert different outcomes for the same starting situation, the merge is blocked, the message names both feature files and the clashing scenario, and nothing is committed to main.
- When a proposed conflict resolution isn't high-confidence, the merge exits without touching main and shows the proposed change and the reason confidence was low, so a person makes the call — resolutions are never applied silently.
- Starting a feature that already has a worktree resumes in place: the existing checkout and its workflow state are reused, the branch is not recreated, and no error is reported. (Run `centinela start` from the main checkout, not from inside an existing worktree.)
- `centinela merge` refuses to run when the main checkout has uncommitted changes: it exits with an error before touching git, tells you to commit or stash first, and leaves the feature's worktree untouched.

## Examples
Turn it on for an existing project:

    centinela migrate --apply   # opt into workflow.use_worktrees and sync ignore files

Or set it directly in `centinela.toml`:

    [workflow]
    use_worktrees = true

Start two features in parallel — each lands in its own isolated checkout:

    centinela start checkout-redesign
    centinela start search-filters
    # .worktrees/checkout-redesign/  and  .worktrees/search-filters/

Merge a finished feature back to main:

    centinela merge checkout-redesign
    # clean: merges, re-validates, removes .worktrees/checkout-redesign/

If a conflict needs a human eye, the command stops and points you at the Steward hint:

    centinela merge search-filters
    # Merge needs review — Merge Steward required.
    # See .workflow/search-filters-merge-steward.md for the conflicted
    # paths and the feature spec, then resolve and re-run.

Note: in this version, `centinela merge` prepares the Merge Steward hint and exits so you can act on it; automatic hand-off to the Steward agent is a planned follow-up.
