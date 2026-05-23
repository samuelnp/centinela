# Plan: Session Context Rehydration

## Problem

Post-`/clear` the model has no project context until the first prompt, and when
the `UserPromptSubmit` context hook finally fires, the active-workflows panel is
~200 KB of noise. Two root causes, one shared goal (useful, not noisy,
post-`/clear` context):

1. **No SessionStart hook.** `.claude/settings.json` registers only
   PreToolUse / PostToolUse / UserPromptSubmit. Nothing injects roadmap state on
   session entry.
2. **Panel noise — confirmed root cause.** `loadActiveWorkflows`
   (`cmd/centinela/hook_workflows.go`) globs every `.workflow/*.json` (263
   files). Only 5 are real workflow-state files; 197 are per-role evidence
   JSONs (`<feature>-<role>.json`) with no `currentStep`. `workflow.Load`
   unmarshals them to a zero-value `Workflow` (`CurrentStep == ""`), which is
   `!= "done"`, so they pass the active filter. Files carrying a `feature` field
   render the same feature N× (e.g. 6× for `add-agent-evidence-contract`),
   producing the duplication and ~200 KB balloon.

## Solution

### Slice 1 — Panel dedupe / cap / evidence-leak fix (real bug fix)

Domain logic lives in `internal/workflow`; `cmd/` stays a thin orchestrator.

- **`internal/workflow/active.go` — new** `ActiveWorkflows(dir string) []*Workflow`
  (domain helper). Walks `<dir>/*.json` but:
  - Skips files whose base name is a `<feature>-<role>` evidence file. The
    canonical guard is: only accept a file when the parsed `wf.Feature` equals
    the file's base name (`<feature>.json`) AND `wf.CurrentStep` is a real,
    non-empty, non-`done` step. This rejects both evidence JSONs (whose feature
    field, when present, never equals their own filename, and which have empty
    `currentStep`) and ad-hoc roadmap JSONs (`roadmap.json`,
    `roadmap-quality.json`).
  - Dedupes by `wf.Feature` (keep the most-recently-touched instance).
  - Sorts surviving workflows by workflow-file mtime descending (recency).
  - Returns the full ordered slice; capping is the caller/presentation's job so
    the domain stays cap-agnostic. Provide a tiny companion
    `CapActive(wfs []*Workflow, max int) (shown []*Workflow, more int)` for the
    cap + `+N more` count (pure, testable).
- **`cmd/centinela/hook_workflows.go` — modify** `loadActiveWorkflows` to
  delegate to `workflow.ActiveWorkflows(workflow.WorkflowDir)` and keep ONLY the
  worktree-scoping filter (`worktree.DetectFeatureFromCwd`) it already owns. No
  classification logic remains in `cmd/`.
- **`internal/ui/render.go` — modify** the `RenderContext` call site path: keep
  `RenderContext(wfs []*workflow.Workflow)` signature unchanged (so existing
  `improve_centinela_render_ui_integration_test.go` is untouched) but render a
  trailing `+N more active` muted line when the caller passes the `more` count.
  Cleanest: add `RenderContextCapped(wfs, more int)` that wraps the existing
  body and appends the hint; `RenderContext` delegates with `more = 0`. The
  cap value (~5) is chosen in `cmd/` (outer wiring), passed down.
- `cmd/centinela/hook_context.go` calls `workflow.CapActive(wfs, 5)` and
  `ui.RenderContextCapped(shown, more)`.

### Slice 2 — SessionStart rehydration hook + payload

- **`internal/roadmap/firstincomplete.go` — new**
  `FirstIncomplete(r *Roadmap) (string, bool)` — walks ALL phases in declared
  order and returns the first feature whose `FeatureStatus != "done"`. This is
  the generalization of `roadmapcheckpoint.FirstIncompleteBootstrap`.
- **`internal/roadmapcheckpoint/firstfeature.go` — modify**
  `FirstIncompleteBootstrap` to delegate to a phase-filtered reuse of the same
  walk (it keeps its Phase-0-only contract by scanning only
  `BootstrapFeatures`, but the per-feature "first not-done" predicate is shared,
  not re-implemented). No behavior change for the checkpoint feature.
- **`internal/ui/render_session.go` — new**
  `RenderSessionRehydration(r *roadmap.Roadmap, next string, hasNext bool) string`
  — composes: a one-line "session rehydration" banner, the full
  `RenderRoadmap(r)` body, a "Next feature to plan: <next>" line (or "roadmap
  complete" when `!hasNext`), and a POINTERS block listing the file paths
  `PROJECT.md` and `docs/features/<next>.md` for the model to read on demand
  (paths only — never inlined content). Pure presentation, no state mutation.
- **`cmd/centinela/hook_session.go` — new** thin orchestrator. Registers
  `centinela hook session`. Drains stdin, loads roadmap (`roadmap.Load`); if the
  roadmap is absent it returns nil silently (setup hook already covers the
  roadmap-missing case). Otherwise computes `roadmap.FirstIncomplete(r)`, prints
  a `CENTINELA DIRECTIVE: session rehydration ...` line and
  `ui.RenderSessionRehydration(...)`, returns nil. Never blocks.
- **`internal/setup/hooks.go` — modify** `mergeHooks` to also receive a
  `session *[]HookGroup` and call `ensureGroup(session, "startup|clear|compact|resume",
  cmdSession, "Rehydrating session context...")` with a new
  `cmdSession = "centinela hook session"` const. (SessionStart matchers are
  source-scoped strings; a single `startup|clear|compact|resume` matcher covers
  all four.)
- **`internal/setup/settings_build.go` — modify** `buildHookSettings` to
  unmarshal/marshal the `SessionStart` key alongside the existing three, passing
  a `session` slice through `mergeHooks`.
- **`.claude/settings.json` — modify** to add a `SessionStart` block:
  `{"matcher":"startup|clear|compact|resume","hooks":[{"type":"command",
  "command":"centinela hook session","statusMessage":"Rehydrating session
  context..."}]}`.

## Files Changed / Added

Added:
- `internal/workflow/active.go` — `ActiveWorkflows`, `CapActive`.
- `internal/roadmap/firstincomplete.go` — `FirstIncomplete` (all phases).
- `internal/ui/render_session.go` — `RenderSessionRehydration`.
- `cmd/centinela/hook_session.go` — `hook session` command (thin).
- `tests/unit/workflow/active_test.go` — evidence-leak rejection, dedupe,
  recency sort, cap + `+N more`.
- `tests/unit/roadmap/firstincomplete_test.go` — first not-done across phases;
  empty/all-done cases.
- `tests/unit/ui/render_session_test.go` — payload shape + pointers, no inlined
  content, "complete" branch.
- `tests/integration/hook_session_integration_test.go` — drives `centinela
  hook session` with a roadmap + workflow fixtures; asserts directive + roadmap
  + pointers and that evidence JSONs do not appear.
- `tests/acceptance/session_context_rehydration_test.go` — Gherkin-backed e2e
  (feature-specialist writes the `.feature`).
- `specs/session-context-rehydration.feature` — written by feature-specialist.

Modified:
- `cmd/centinela/hook_workflows.go` — delegate to `workflow.ActiveWorkflows`.
- `cmd/centinela/hook_context.go` — apply cap, call `RenderContextCapped`.
- `internal/ui/render.go` — add `RenderContextCapped` (+`RenderContext`
  delegates).
- `internal/roadmapcheckpoint/firstfeature.go` — delegate the per-feature
  predicate; no behavior change.
- `internal/setup/hooks.go` — `cmdSession` const + SessionStart wiring in
  `mergeHooks`.
- `internal/setup/settings_build.go` — thread `SessionStart` through.
- `.claude/settings.json` — add SessionStart block.

## Rollout Sequence

1. **Slice 1 — panel fix first.** It is the smallest correct slice, a real bug
   fix, independently shippable, and the most user-visible relief (kills the
   ~200 KB balloon and the 6× duplication). It touches only the existing
   `UserPromptSubmit` path and adds `internal/workflow/active.go`. Land with
   full unit + integration + acceptance coverage of the evidence-leak,
   dedupe, recency, and cap behaviors.
2. **Slice 2 — SessionStart hook + rehydration payload.** Builds on a clean
   panel. Add `roadmap.FirstIncomplete` (and refactor the checkpoint helper to
   delegate), the `RenderSessionRehydration` renderer, the `hook session`
   command, and the settings wiring (both `hooks.go`/`settings_build.go` and
   `.claude/settings.json`). Land with its own unit/integration/acceptance
   coverage.

Rationale for this order: Slice 1 has zero dependency on Slice 2, is the
verified bug, and de-risks Slice 2 by guaranteeing the panel it sits beside is
already small. Slice 2's domain helper (`FirstIncomplete`) is the only piece
the checkpoint feature touches, so doing it second keeps the refactor isolated.

## Risks & Mitigations

- **Regression to existing hook-output specs/integration tests.**
  `improve_centinela_render_ui_integration_test.go` calls `RenderContext` with
  an explicit slice and `edge_case_context_integration_test.go` drives `hook
  context` with a single saved workflow whose `currentStep` is real — both stay
  green because (a) `RenderContext` keeps its signature and (b) the new
  `ActiveWorkflows` accepts a real `<feature>.json` with a non-empty step.
  Mitigation: keep `RenderContext` unchanged (add `RenderContextCapped`
  alongside), re-run the full integration suite in slice 1.
- **`ActiveWorkflows` filter too strict / too loose.** If the "feature == base
  name AND non-empty currentStep" guard is wrong, real workflows could vanish or
  evidence could still leak. Mitigation: unit tests assert the exact 5 real vs
  197 evidence classification on representative fixtures (one real, one
  `-qa-senior.json`, one bare `roadmap.json`, one done). The base-name equality
  guard is the load-bearing rule and is covered directly.
- **SessionStart matcher / source coverage.** If Claude Code does not honor a
  combined `startup|clear|compact|resume` matcher, some sources won't fire.
  Mitigation: feature-specialist confirms the matcher syntax against the
  current Claude Code hooks contract; integration test asserts the wired
  settings block shape; fall back to one group per source if the combined
  matcher is unsupported.
- **Double injection / noise on session entry.** SessionStart fires then the
  first `UserPromptSubmit` also injects roadmap summary + panel. Mitigation:
  SessionStart emits the full roadmap + pointers (rehydration); the per-prompt
  path keeps its compact summary. Accepted minor overlap; flagged for the
  feature-specialist to decide whether to suppress the per-prompt full roadmap
  (out of scope here).
- **Cross-layer leak (G2/G7).** Classification or first-incomplete logic
  drifting into `cmd/`. Mitigation: all domain logic in `internal/workflow` /
  `internal/roadmap`; `cmd/centinela/hook_session.go` and the modified
  `hook_workflows.go` remain ≤100-line thin orchestrators.
- **G1 file-size on settings wiring.** `settings_build.go` and `hooks.go` are
  near their budget. Mitigation: keep the SessionStart additions minimal; if
  `hooks.go` would exceed 100 lines, split the session wiring into a small
  `hooks_session.go`.
