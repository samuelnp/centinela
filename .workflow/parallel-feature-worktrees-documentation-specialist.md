### Documentation-Specialist Report: parallel-feature-worktrees
**Date:** 2026-05-15
**Status:** done
**Handoff:** complete

#### KB Entry Summary

Wrote `docs/project-docs/kb/parallel-feature-worktrees.md` to the exact
end-user KB contract (frontmatter `feature` / `summary` / `audience:
end-user` / `status: done`; required sections `## What it does`,
`## When you'd use it`, `## How it behaves`; optional `## Examples`).
Plain language only — no Given/When/Then, no engineering vocabulary.

The entry explains, for non-technical Centinela users:
- per-feature isolated checkouts under `.worktrees/<feature>/` so
  multiple workflows / agents run concurrently without colliding;
- opt-in via `workflow.use_worktrees` (wizard ON by default; existing
  projects via `centinela migrate`); fully backward compatible when off;
- `centinela merge <feature>` — spec-conflict pre-check, git merge,
  re-validate, Merge Steward on text conflict OR post-merge validate
  failure, no silent resolutions, worktree removed on clean success.

Both acknowledged v1 gaps are stated honestly without over-emphasis:
- Merge Steward agent auto-dispatch is a follow-up — `centinela merge`
  prepares the `.workflow/<feature>-merge-steward.md` hint and exits
  for the operator to act on (closing "Examples" note + the conflict
  bullet in "How it behaves").
- Restarting a feature must be run from the main checkout, not from
  inside an existing worktree (parenthetical in the resume bullet).

#### Spec Coverage

Source spec: `specs/parallel-feature-worktrees.feature` — Feature
"Parallel feature worktrees with merge steward".

| Metric | Value |
|--------|-------|
| Scenarios in spec | 11 |
| `## How it behaves` bullets | 11 (one per scenario, rewritten as user-visible behavior) |

Scenario → bullet mapping (all 11 covered):
1. Start provisions a worktree when use_worktrees is enabled
2. Start runs in the main checkout when use_worktrees is disabled
3. Wizard syncs tool ignore lists for new projects
4. Migrate syncs tool ignore lists for existing projects
5. Clean merge when git applies cleanly and validate passes
6. Text conflict invokes the Merge Steward
7. Semantic conflict after a clean text merge invokes the Steward
8. Spec conflict across in-flight worktrees is detected before merging
9. Merge Steward escalates uncertain resolutions to the user
10. Restarting a feature with an existing worktree resumes in place
11. Merge fails fast when the main working tree is dirty

#### Roadmap Dependencies

`ROADMAP.md` lists only Phase 0 (`docs-migration-managed-docs`). This
feature is not yet a roadmap phase entry; it has no upstream roadmap
dependency and introduces no roadmap blocker. It builds additively on
the orchestration evidence contract (`add-agent-evidence-contract`) and
the managed-docs KB pages (`docs-knowledge-base-pages`), both `done`.

#### Workflow Status Matrix

| Step     | Role                   | Status | Evidence |
|----------|------------------------|--------|----------|
| plan     | big-thinker            | done   | `.workflow/parallel-feature-worktrees-big-thinker.{md,json}` |
| plan     | feature-specialist     | done   | `.workflow/parallel-feature-worktrees-feature-specialist.{md,json}` |
| code     | senior-engineer        | done   | `.workflow/parallel-feature-worktrees-senior-engineer.{md,json}` |
| tests    | qa-senior              | done   | `.workflow/parallel-feature-worktrees-qa-senior.{md,json}` |
| validate | gatekeeper             | SAFE   | `.workflow/parallel-feature-worktrees-gatekeeper.md` |
| validate | validation-specialist  | WARNING (ship) | `.workflow/parallel-feature-worktrees-validation-specialist.{md,json}` |
| docs     | documentation-specialist | done | this report + JSON |

Validation-specialist returned WARNING (ship) with two documented,
tracked follow-ups (stubbed Steward auto-dispatch; second-`start`-from-
inside-worktree edge bug). Both are reflected honestly in the KB entry.

#### Generated Outputs

`centinela docs generate --out docs/project-docs/index.html` exited 0.
Confirmed on disk:
- `docs/project-docs/kb/parallel-feature-worktrees.md` (KB source)
- `docs/project-docs/kb/parallel-feature-worktrees.html` (rendered)
- `docs/project-docs/kb/index.html` (KB index, lists this feature)
- `docs/project-docs/index.html` (main report, references the feature)

#### Handoff

- Next: `complete` — run `centinela complete parallel-feature-worktrees`.
- `centinela docs validate` confirmed inputs valid before authoring.
- No source code edited; generation succeeded with no missing artifacts.
