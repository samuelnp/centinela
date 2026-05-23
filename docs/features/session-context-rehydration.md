# Feature: Session Context Rehydration

## Problem

After a `/clear` (or compact/resume/startup), Claude loses the conversational
context and struggles to rediscover project state when asked to "plan the next
feature". Two compounding causes:

1. **No SessionStart injection.** Centinela registers ZERO `SessionStart`
   hooks — every context injection rides on `UserPromptSubmit`
   (`centinela hook context`, `setup`, etc.). So immediately after `/clear`,
   nothing is injected until the user's first prompt, and the model has no
   roadmap, no "next feature", and no pointers to read.

2. **The per-prompt active-workflows panel is noisy.** When the
   `UserPromptSubmit` context hook does fire, `runHookContext` renders
   `RenderContext(loadActiveWorkflows())`. `loadActiveWorkflows` globs ALL
   `.workflow/*.json`. That directory holds 263 JSON files — only 5 are real
   active workflow-state files; **197 are per-role evidence files**
   (`<feature>-big-thinker.json`, `<feature>-qa-senior.json`, …) that have no
   `currentStep` field, so `workflow.Load` unmarshals them into a zero-value
   `Workflow` with `CurrentStep == ""` (which is `!= "done"`) and they leak
   through as "active". Files that carry a `feature` field render the same
   feature multiple times (e.g. `add-agent-evidence-contract` 6×). The panel
   balloons to ~200 KB of duplicated/DONE noise that drowns the tiny roadmap
   summary — the real reason the model "struggles".

## Outcome

After `/clear` (and on startup/compact/resume), Claude receives a single,
compact, LLM-readable bootstrap and the per-prompt panel stays small and
truthful:

- A **SessionStart hook** (`centinela hook session`) injects ONCE per session
  entry: the full roadmap with per-feature status, the **next feature to plan**
  (first incomplete feature across ALL phases in declared order), and
  **pointers** (file paths to `PROJECT.md` and `docs/features/<next>.md` for the
  model to read on demand — content is NOT inlined).
- The **active-workflows panel** shows ONLY genuinely active workflow-state
  files, deduplicated by feature name, capped to the ~5 most-recently-touched,
  with a `+N more` hint when more exist. Evidence JSONs and DONE workflows
  never appear.

## Scope

- **New SessionStart hook** `centinela hook session` (thin orchestrator in
  `cmd/centinela`), wired into BOTH `internal/setup/hooks.go` and
  `.claude/settings.json` under a `SessionStart` matcher covering all four
  sources: `startup`, `clear`, `compact`, `resume`.
- **New rehydration payload renderer** in `internal/ui` that composes the full
  roadmap (reusing `RenderRoadmap`), the "next feature to plan" line, and the
  read-on-demand pointers (`PROJECT.md`, `docs/features/<next>.md`).
- **Generalize** `roadmapcheckpoint.FirstIncompleteBootstrap` (Phase-0-only)
  into a roadmap-domain helper that returns the first incomplete feature across
  ALL phases in declared order. Reuse `roadmap.FeatureStatus`; do not duplicate
  status logic. The existing Phase-0 helper is refactored to delegate, not
  re-implement.
- **Fix the active-workflows panel root cause** in the input-building path
  (`loadActiveWorkflows`): only treat real workflow-state files as workflows
  (exclude evidence JSONs, exclude files with empty `currentStep`), dedupe by
  feature name, sort by recency (workflow file mtime), cap to ~5, surface a
  `+N more` count. Dedupe/cap is domain logic, not raw `cmd/` code.
- Unit, integration, and acceptance tests for both halves
  (panel correctness + SessionStart payload + first-incomplete-across-phases).

## Out of Scope

- OpenCode parity for SessionStart — the OpenCode plugin has no
  SessionStart-equivalent event (only `tui.prompt.append`). The panel fix still
  benefits OpenCode because both runtimes call `centinela hook context`.
- Inlining `PROJECT.md` / feature-brief content into the payload — we emit
  pointers only, to keep the injection small.
- Persisting or de-duplicating across multiple session entries within one run
  (the hook fires once per SessionStart event; we do not add a session-id
  marker to suppress repeats).
- Garbage-collecting or relocating the 197 evidence JSONs out of `.workflow/`
  (the panel fix makes them harmless; cleanup is a separate concern).
- Changing the `RenderContext` signature or the other `UserPromptSubmit`
  directives (setup/migrate/autostart/orchestration/plan-advisor).
