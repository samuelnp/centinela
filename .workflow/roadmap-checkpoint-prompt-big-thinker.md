# Big-Thinker Report: roadmap-checkpoint-prompt

**Date:** 2026-05-20

## Problem

A Centinela project initiator finishes the roadmap-definition flow — `PROJECT.md`,
`ROADMAP.md`, `.workflow/roadmap.json`, `.workflow/roadmap-analysis.{md,json}`,
`.workflow/roadmap-quality.{md,json}`, and
`docs/architecture/production-readiness-prompt.md` all exist — and then sees no
explicit handoff. `centinela hook setup` returns silently, so Claude often
either dives into the first Phase 0 feature unprompted or answers an unrelated
user message. We need an explicit, one-shot directive that asks the user
whether to keep iterating on the roadmap or start implementing the first
incomplete Phase 0 bootstrap feature. The directive must respect "I want to
keep iterating" (suppress until artifacts change), and it must re-fire only
when the user edits a roadmap-defining artifact after picking iterate.

## Scope

- **In**
  - New `internal/roadmapcheckpoint/` package owning emit/suppress/stale
    decision, marker shape, and first-incomplete Phase 0 lookup.
  - New directive branch in `cmd/centinela/hook_setup.go` after the
    production-readiness check.
  - New UI renderer `internal/ui/render_roadmap_checkpoint.go`.
  - Marker file: `.workflow/roadmap-checkpoint.json` with
    `{"choice":"iterate","at":"<RFC3339>"}`.
  - Freshness check via mtime comparison against required artifacts.
  - Unit + integration + acceptance tests covering emit, suppress, and stale.
- **Out**
  - Automated re-running of senior-PM analysis / quality scoring on
    ROADMAP.md edits.
  - Multi-phase checkpoints beyond Phase 0.
  - Cross-project checkpoint history or telemetry.
  - Parsing the user's natural-language reply — the orchestrator (Claude)
    writes the marker or invokes `centinela start`.

## Dependencies & Assumptions

- Depends on `internal/roadmap.BootstrapFeatures`, `roadmap.FeatureStatus`,
  `roadmap.Load`.
- Hosts the new directive inside `centinela hook setup` (already wired in
  `.claude/settings.json` UserPromptSubmit). No new hook registration needed.
- Marker timestamp is RFC 3339; comparison uses `time.Time.Equal` /
  `Before`. mtime is read via `os.Stat`.
- Mtime granularity is acceptable: a no-op editor "save without modification"
  may bump mtime and re-fire the prompt. Documented trade-off; preferable to
  silent drift.
- `cmd/centinela/hook_setup.go` stays a thin orchestrator (n-tier outer
  layer); all decision logic lives in `internal/`.
- Project archetype is n-tier per PROJECT.md; UI is presentation-only and
  must not mutate state.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Directive spam if marker logic regresses | High | Medium | Unit tests for every `Decide` path; integration test asserts second `runHookSetup` with unchanged disk is silent. |
| Existing project where bootstrap is already complete sees the prompt | Medium | Medium | `Decide` returns `DecisionSuppressed` when `FirstIncompleteBootstrap` returns false. |
| First Phase 0 feature absent from roadmap | Medium | Low | `FirstIncompleteBootstrap` returns `("", false)`; checkpoint suppresses, renderer never asked for a name. |
| mtime granularity / no-op saves re-fire prompt | Low | Medium | Accept trade-off; document in feature brief and edge-case file. |
| Race when two concurrent `hook setup` invocations both see no marker | Low | Low | Marker write is single `os.WriteFile`; both writes converge to same `{choice:"iterate"}` content; idempotent. |
| `cmd/centinela/hook_setup.go` exceeds 100 lines | Medium | Medium | Push all decision + path logic into `internal/roadmapcheckpoint`; extract a small helper or split the file if it grows. |
| Cross-layer leak (logic in `cmd/`) | High | Low | Code review against G7; only directive printing + delegation in `cmd/`. |
| User picks "start" but workflow file isn't created yet (edge in orchestrator) | Low | Low | Marker absence + workflow absence keeps prompt active; not destructive. |

## Rollout

1. **Step 1 — emit + suppress slice.** Land `internal/roadmapcheckpoint` with
   `Decide` covering only "marker absent → emit" and "marker present →
   suppress". Wire directive + renderer. Unit + integration tests for those
   two paths. Skip freshness for now.
2. **Step 2 — freshness check.** Add mtime comparison against required
   artifacts; introduce `DecisionStale`. Tests for "marker present but stale".
3. **Step 3 — natural suppression by workflow presence.** Suppress when
   `.workflow/<first-phase-0-feature>.json` already exists; covers the
   "start" branch automatically.
4. **Step 4 — polish.** Renderer copy review; document mtime-granularity
   trade-off; ensure file size stays under 100 lines per source file.

## Handoff

- Next role: `feature-specialist`
- Outstanding questions:
  1. Should the marker also be written for "start" choice (for symmetry /
     audit), or is workflow-file presence the canonical signal?
     Recommendation: workflow file is canonical; do not write a "start" marker.
  2. Should the directive include a hint to run
     `centinela start <feature>` verbatim, so Claude has zero-effort wiring?
     Recommendation: yes — include the literal command in the panel body.
