### Big-Thinker Report: session-context-rehydration
**Date:** 2026-05-23

#### Problem

Developers using Claude Code with Centinela hit a wall after `/clear`: the model
loses context and cannot reliably rediscover project state when asked to "plan
the next feature". Two compounding causes share one goal — make post-`/clear`
context useful, not noisy. (1) Centinela registers ZERO `SessionStart` hooks;
all injection rides on `UserPromptSubmit`, so nothing is injected on session
entry. (2) When the per-prompt context hook does fire, `loadActiveWorkflows`
globs ALL 263 `.workflow/*.json` files; 197 of them are per-role evidence JSONs
(`<feature>-<role>.json`) that have no `currentStep`, so `workflow.Load`
produces zero-value workflows (`CurrentStep == ""`, i.e. `!= "done"`) that leak
through as "active" and render the same feature up to 6×, ballooning the panel
to ~200 KB and drowning the tiny roadmap summary.

#### Scope

- **In:**
  - New `SessionStart` hook `centinela hook session` injecting ONCE per session
    entry: full roadmap with per-feature status, the next feature to plan
    (first incomplete across ALL phases), and read-on-demand pointers
    (`PROJECT.md`, `docs/features/<next>.md`) — paths only, not inlined.
  - Wire SessionStart in `internal/setup/hooks.go` + `internal/setup/
    settings_build.go` + `.claude/settings.json` for sources
    `startup|clear|compact|resume`.
  - Generalize `roadmapcheckpoint.FirstIncompleteBootstrap` into
    `roadmap.FirstIncomplete` (all phases); refactor the Phase-0 helper to
    delegate the shared predicate.
  - Fix the active-workflows panel root cause: only real workflow-state files,
    dedupe by feature, sort by recency (file mtime), cap to ~5 with `+N more`.
- **Out:** OpenCode SessionStart parity (no equivalent event); inlining file
  content; cross-session-entry suppression markers; GC of the 197 evidence
  JSONs; changing `RenderContext` signature or the other UserPromptSubmit
  directives.

#### Dependencies & Assumptions

- Reuses `roadmap.Load`, `roadmap.FeatureStatus`, `roadmap.RenderRoadmap`,
  `roadmap.Roadmap/Phase/Feature`.
- Reuses `workflow.Load`, `workflow.WorkflowDir`, `workflow.Workflow`
  (`Feature`, `CurrentStep`) and `worktree.DetectFeatureFromCwd` for scoping.
- Builds on the existing checkpoint helper (`internal/roadmapcheckpoint/
  firstfeature.go`) — generalized, not duplicated.
- Assumes Claude Code honors a combined `startup|clear|compact|resume`
  SessionStart matcher (feature-specialist to confirm; fallback: one group per
  source).
- Assumes the workflow-state filename convention `<feature>.json` and that
  evidence files are always `<feature>-<role>.json` — the load-bearing guard is
  "parsed `wf.Feature` equals the file base name AND `currentStep` is a real
  non-empty non-done step".

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Regress hook-output specs/integration tests | High | Low | Keep `RenderContext` signature; add `RenderContextCapped`; `ActiveWorkflows` accepts real `<feature>.json`; re-run full integration suite in slice 1 |
| `ActiveWorkflows` filter too strict/loose | High | Medium | Unit-test the exact real-vs-evidence classification on fixtures (real, `-qa-senior.json`, bare `roadmap.json`, done); base-name equality guard covered directly |
| SessionStart matcher/source coverage unsupported | Medium | Low | Confirm matcher syntax vs current contract; integration-assert settings block; fallback to per-source groups |
| Double injection (SessionStart + first prompt) | Low | Medium | SessionStart = full roadmap + pointers; per-prompt keeps compact summary; flag optional per-prompt suppression to feature-specialist |
| Cross-layer leak (G2/G7) | Medium | Low | Classification in `internal/workflow`, first-incomplete in `internal/roadmap`; `cmd/` stays thin |
| G1 file-size on `hooks.go`/`settings_build.go` | Low | Medium | Keep additions minimal; split `hooks_session.go` if `hooks.go` would exceed 100 lines |

#### Rollout

- Step 1 — **Panel dedupe/cap/evidence-leak fix** (smallest correct slice, real
  bug, independently shippable, biggest user relief). Add
  `internal/workflow/active.go` (`ActiveWorkflows`, `CapActive`); delegate
  `loadActiveWorkflows`; add `ui.RenderContextCapped`. Full unit + integration
  + acceptance coverage of leak/dedupe/recency/cap.
- Step 2 — **SessionStart hook + rehydration payload**. Add
  `roadmap.FirstIncomplete` (refactor checkpoint helper to delegate),
  `ui.RenderSessionRehydration`, `cmd/centinela/hook_session.go`, and settings
  wiring (`hooks.go` + `settings_build.go` + `.claude/settings.json`). Own
  unit/integration/acceptance coverage.

Order rationale: Slice 1 has zero dependency on Slice 2 and is the verified
bug; doing it first de-risks Slice 2 by ensuring the adjacent panel is already
small. The shared `FirstIncomplete` helper is the only checkpoint-feature touch,
isolated in Slice 2.

#### Handoff

- Next role: feature-specialist (writes `specs/session-context-rehydration.feature`
  and refines acceptance criteria).
- Outstanding questions for the feature-specialist:
  1. Confirm Claude Code SessionStart matcher syntax — combined
     `startup|clear|compact|resume` vs one group per source.
  2. Decide whether the per-prompt context hook should suppress its full roadmap
     once SessionStart already injected it this session (currently out of scope;
     minor overlap accepted).
  3. Confirm the exact `RenderSessionRehydration` copy and pointer format
     (which paths beyond `PROJECT.md` and `docs/features/<next>.md`, if any).
