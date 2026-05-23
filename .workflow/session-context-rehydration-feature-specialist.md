### Feature-Specialist Report: session-context-rehydration
**Date:** 2026-05-23

#### Behavior Summary

This feature delivers two observable behaviors that share one goal — useful,
non-noisy context after `/clear`. Half A fixes the per-prompt active-workflows
panel emitted by `centinela hook context` (UserPromptSubmit): it now lists ONLY
genuine non-done workflow-state files (feature name equals the file base name
AND `currentStep` is a real, non-empty, non-`done` step), deduplicated by
feature name, sorted by file mtime descending, capped to ~5 entries with a
`+N more` hint when more exist — so evidence JSONs (`<feature>-<role>.json`),
ad-hoc roadmap JSONs, and DONE workflows no longer leak in or duplicate. Half B
adds a new `SessionStart` hook (`centinela hook session`) that fires once per
session entry (sources `startup`, `clear`, `compact`, `resume`) and injects a
`CENTINELA DIRECTIVE: session rehydration` payload: the full roadmap with
per-feature status, the next feature to plan (the first incomplete feature
across ALL phases in declared order via `roadmap.FirstIncomplete`), and
read-on-demand pointers (file paths to `PROJECT.md` and
`docs/features/<next>.md` — never inlined). When the roadmap is complete it
emits a graceful "roadmap complete" state with no next feature; when the
roadmap is missing or invalid it exits zero and emits no payload.

#### Gherkin Scenarios

All scenarios live in `specs/session-context-rehydration.feature`.

Half A — active-workflows panel:
- **An evidence JSON in .workflow/ is not rendered as an active workflow** —
  Given a real `alpha.json` and an evidence `alpha-qa-senior.json` (no
  `currentStep`), When `hook context` runs, Then only `alpha` is listed and the
  evidence-derived entry is not. (Acceptance test of record.)
- **A done workflow is excluded while a genuine non-done workflow is shown** —
  `done` feature suppressed, real `tests`-step feature shown.
- **Ad-hoc roadmap JSON files are not treated as active workflows** —
  `roadmap.json` / `roadmap-quality.json` excluded; real `delta.json` shown.
- **Duplicate feature entries are deduplicated to a single panel row** —
  a feature with several evidence JSONs appears exactly once.
- **More active workflows than the cap show only the most-recently-touched plus
  a "+N more" hint** — 7 features → 5 most-recent shown (mtime desc) + `+2 more`.
- **At-or-below the cap shows no "+N more" hint** — 3 features → 3 listed, no
  indicator.

Half B — SessionStart rehydration:
- **SessionStart injects the rehydration payload on each supported source**
  (Scenario Outline over `startup|clear|compact|resume`) — directive line, full
  roadmap with per-feature status, named next feature, pointer paths
  `PROJECT.md` + `docs/features/<next>.md`, no inlined file content, exit zero.
- **Next feature is the first incomplete across all phases, not just Phase 0** —
  all Phase 0 done → picks first incomplete Phase 1 feature.
- **Every roadmap feature done yields a graceful roadmap-complete state with no
  next feature** — roadmap-complete indicated, no next name, no `<next>.md`
  pointer, no crash.
- **Missing roadmap is handled gracefully without crashing** — exit zero, no
  payload.
- **Invalid roadmap json is handled gracefully without crashing** — exit zero,
  no payload.

#### UX States

This feature has no GUI; the "surface" is injected hook stdout (LLM-readable
text). Loading is n/a (synchronous CLI invocation).

| State    | Trigger | Surface |
|----------|---------|---------|
| loading  | n/a (synchronous one-shot hook invocation) | n/a |
| empty    | No active non-done workflows (panel) / every roadmap feature done (session) | Panel: "No active workflows." success line. Session: graceful "roadmap complete" state, no next feature, no `<next>.md` pointer |
| error    | ROADMAP.md / `.workflow/roadmap.json` missing or malformed (session) | No `session rehydration` payload emitted; command exits zero without crashing (graceful no-op) |
| success  | Active non-done workflows present (panel) / valid roadmap with a first-incomplete feature (session) | Panel: deduped, mtime-sorted, capped list with `+N more`. Session: `CENTINELA DIRECTIVE: session rehydration` + full roadmap + next-feature line + `PROJECT.md` and `docs/features/<next>.md` pointers |

#### Out-of-Scope

- OpenCode SessionStart parity — OpenCode has no SessionStart-equivalent event
  (only `tui.prompt.append`); only the panel fix benefits OpenCode.
- Inlining `PROJECT.md` / feature-brief content — pointers (paths) only.
- Cross-session-entry suppression — the hook fires once per SessionStart event;
  no session-id marker is added to suppress repeats within one run.
- Suppressing the per-prompt context hook's roadmap summary when SessionStart
  already injected one — the minor overlap is explicitly accepted for v1.
- Pointers beyond `PROJECT.md` and `docs/features/<next>.md`.
- Garbage-collecting / relocating the evidence JSONs out of `.workflow/`.
- Changing the `RenderContext` signature or the other UserPromptSubmit
  directives (setup/migrate/autostart/orchestration/plan-advisor).

#### Handoff

- Next role: senior-engineer
- Open clarifications for the senior-engineer to decide (code-step details, not
  re-litigated here):
  1. Exact `.claude/settings.json` SessionStart matcher syntax — confirm a
     single combined `startup|clear|compact|resume` matcher is honored by Claude
     Code; fall back to one HookGroup per source if not. The spec asserts the
     per-source BEHAVIOR (via the Scenario Outline), not the matcher string.
  2. Exact wording/format of the rehydration banner and the pointer block, and
     the precise `+N more` phrasing in `RenderContextCapped`. The spec asserts
     substrings (`CENTINELA DIRECTIVE: session rehydration`, `+2 more`, pointer
     paths) — keep those literals stable so the acceptance assertions hold.
  3. The cap value is fixed at ~5 in the spec; if the constant changes, update
     the "+N more" arithmetic in the acceptance scenarios accordingly.
