### Documentation-Specialist Report: session-context-rehydration
**Date:** 2026-05-23
**Step:** docs (final) — handoff: complete

#### KB entry summary
Wrote `docs/project-docs/kb/session-context-rehydration.md` (audience: end-user,
status: done) covering both halves of the feature in plain language:

- **Half A — active-workflows panel fix.** The per-prompt panel now lists only
  genuine in-progress workflows: internal per-role evidence JSONs and ad-hoc
  bookkeeping files (`roadmap.json`, `roadmap-quality.json`) are no longer
  treated as active, done workflows are hidden, entries are deduplicated by
  feature, capped to the 5 most-recently-touched (newest first), with a
  `+N more` hint for the rest. This kills the prior ~200 KB / 6×-duplicated
  balloon.
- **Half B — SessionStart rehydration.** On startup / clear / compact / resume,
  Centinela injects a one-time `CENTINELA DIRECTIVE: session rehydration`
  payload: full roadmap with per-feature status, the next feature to plan
  (first not-done across ALL phases in declared order), and pointer PATHS to
  `PROJECT.md` and `docs/features/<next>.md` (paths only, never inlined).
  Graceful no-crash exit-0 on roadmap-complete, missing roadmap, and malformed
  roadmap.

The entry has all three REQUIRED sections (What it does / When you'd use it /
How it behaves — one bullet per spec scenario, no Given/When/Then) plus an
Examples section. House style matched against existing
`merge-steward-auto-dispatch.md` and `parallel-feature-worktrees.md`.

#### Roadmap dependencies
- `ROADMAP.md` / `.workflow/roadmap.json` currently declare only Phase 0
  (`docs-migration-managed-docs`); the SessionStart "next feature" walk
  generalizes the Phase-0-only `roadmapcheckpoint.FirstIncompleteBootstrap`
  via the shared `roadmap.FirstNotDone` predicate, so cross-phase ordering
  works once later phases are declared (covered by the cross-phase spec
  scenario and tests).
- The active-workflows panel fix is the prerequisite that keeps the
  SessionStart payload's roadmap summary readable (panel sits beside it on the
  first prompt); shipped as Slice 1 first, per the plan's rollout order.
- Related KB pages cross-linked by the generated index:
  `parallel-feature-worktrees`, `merge-steward-auto-dispatch`,
  `roadmap-checkpoint-prompt`, `add-agent-evidence-contract`,
  `docs-knowledge-base-pages`.

#### Workflow status matrix
| Step      | Status      |
|-----------|-------------|
| plan      | done        |
| code      | done        |
| tests     | done        |
| validate  | done        |
| docs      | in-progress (this step; complete NOT run) |

#### Major specs + scenario count
`specs/session-context-rehydration.feature` — **11 scenarios** total:
- Half A (panel, 6): evidence-JSON not active; done excluded / non-done shown;
  ad-hoc roadmap JSONs not active; duplicates dedupe to one row; over-cap shows
  5 + `+N more`; at-or-below cap shows no hint.
- Half B (SessionStart, 5): rehydration payload on each source (1 Scenario
  Outline over startup|clear|compact|resume); next feature is first incomplete
  across all phases; all-done roadmap-complete (no next, no next pointer);
  missing roadmap silent no-crash; invalid roadmap silent no-crash.

#### Generation verification
- `centinela docs validate` → exit 0 (inputs valid).
- `centinela docs generate --out docs/project-docs/index.html` → exit 0.
- Files on disk:
  - `docs/project-docs/kb/session-context-rehydration.md`
  - `docs/project-docs/kb/session-context-rehydration.html`
  - `docs/project-docs/kb/index.html`
  - `docs/project-docs/index.html`
- Feature referenced in both the KB index and the main docs report.

#### Handoff
- Next: complete (no production source edited; `centinela complete` NOT run).
