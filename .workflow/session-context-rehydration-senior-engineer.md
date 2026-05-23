### Senior-Engineer Report: session-context-rehydration
**Date:** 2026-05-23

#### Files Touched
| Path | Reason |
|------|--------|
| internal/workflow/active.go (new) | Domain helper `ActiveWorkflows` (evidence-leak rejection, dedupe, recency sort) + pure `CapActive`. |
| cmd/centinela/hook_workflows.go (mod) | `loadActiveWorkflows` now delegates to `workflow.ActiveWorkflows`; keeps ONLY the worktree-scoping filter. No classification in cmd/. |
| internal/ui/render.go (mod) | Added `RenderContextCapped(wfs, more)`; `RenderContext` delegates with more=0 (signature unchanged for regression safety). |
| cmd/centinela/hook_context.go (mod) | Applies cap=5 (`workflow.CapActive`) in the outer layer, calls `RenderContextCapped`. |
| internal/roadmap/firstincomplete.go (new) | `FirstIncomplete` (first not-done across ALL phases) + shared `FirstNotDone` predicate. |
| internal/roadmapcheckpoint/firstfeature.go (mod) | `FirstIncompleteBootstrap` reuses `roadmap.FirstNotDone` over BootstrapFeatures only — Phase-0 contract unchanged. |
| internal/ui/render_session.go (new) | `RenderSessionRehydration` — banner + full roadmap + next-feature/complete line + pointer PATHS (PROJECT.md, docs/features/<next>.md). |
| cmd/centinela/hook_session.go (new) | Thin orchestrator for `centinela hook session`; drains stdin, loads roadmap, silent no-op on absent/invalid roadmap, else prints directive + payload. |
| internal/setup/hooks.go (mod) | `cmdSession` const + `sessionMatcher` + SessionStart wiring in `mergeHooks(pre,post,prompt,session)`. |
| internal/setup/settings_build.go (mod) | Threads `SessionStart` raw key through unmarshal/mergeHooks/marshal. |
| .claude/settings.json (mod) | Added SessionStart block (combined matcher, statusMessage). |
| internal/setup/hooks_test.go (mod) | Mechanical update to the new 4-arg `mergeHooks` signature + session group-size assertion (existing test; not a feature test). |

#### Architecture Compliance
- Boundary checks passed: classification logic lives in `internal/workflow`; first-incomplete logic in `internal/roadmap`; all rendering in `internal/ui`; `cmd/` stays thin (only the worktree filter + cap constant + stdin drain + Println). No domain logic leaked into cmd/ (G2/G7).
- G1 file size: every touched `.go` file ≤ 100 lines (render.go is exactly 100). No G1 exception needed; no `hooks_session.go` split required (hooks.go = 73 lines).
- G7 outer-layer rule: no business logic moved into cmd/; the cap value (5) is a presentation/wiring choice and is the only constant chosen in the outer layer, as the plan dictates.
- Spec-asserted literals kept STABLE: `CENTINELA DIRECTIVE: session rehydration`, `+N more active` (contains `+N more`), pointer paths `PROJECT.md` and `docs/features/<next>.md`, cap = 5.

#### Type-Safety Notes
- No `any`/`interface{}` introduced. `ActiveWorkflows` uses a local typed `tracked` struct (`*Workflow`, `int64` mtime) for the dedupe map rather than an untyped map value.
- `CapActive` returns named typed results `(shown []*Workflow, more int)`; `FirstIncomplete`/`FirstNotDone` return `(string, bool)` — explicit presence flags, no sentinel strings.
- SessionStart wiring reuses the existing strongly-typed `HookGroup`/`HookCmd` structs and `json.RawMessage` map (no dynamic decoding shortcuts).

#### Trade-Offs
- **Filter-in-place in `loadActiveWorkflows`** (`active[:0]`) reuses the backing array — safe because writes never outpace reads; avoids a second allocation. Alternative (fresh slice) rejected as needless churn on the hot per-prompt path.
- **`FirstNotDone` exported** (rather than duplicating the predicate or passing a closure) so both `roadmap.FirstIncomplete` and `roadmapcheckpoint.FirstIncompleteBootstrap` share one not-done rule without a circular import (checkpoint already imports roadmap).
- **Combined SessionStart matcher** `startup|clear|compact|resume` chosen over one-group-per-source: verified against the current Claude Code hooks docs (matchers are regex over the source string; docs show `startup|resume` as a SessionStart example). Single group is honored, so the fallback was not needed.
- **Reminder loops in hook_context still iterate full `wfs`** (not the capped `shown`) so edge-case/docs/review reminders fire for every active workflow, not just the visible 5 — preserves prior behavior.

#### Handoff
- Next role: qa-senior
- Outstanding TODOs (coverage dip is expected — new code is untested by design this step; `go test ./...` itself is green, no regressions):
  - `internal/workflow/active.go`: cover `ActiveWorkflows` (evidence-JSON rejection via feature==basename guard, ad-hoc roadmap.json rejection, done-exclusion, dedupe-by-feature keeping most-recent, mtime-desc sort) and `CapActive` (cap arithmetic, more=0 at/below cap, more=N above).
  - `internal/roadmap/firstincomplete.go`: `FirstIncomplete` cross-phase walk (first not-done across all phases incl. all-Phase-0-done -> Phase 1), nil/empty/all-done -> (\"\", false); `FirstNotDone`.
  - `internal/ui/render_session.go`: `RenderSessionRehydration` payload shape, pointer paths present, NO inlined file content, roadmap-complete branch (no next, no <next>.md pointer).
  - `cmd/centinela/hook_session.go`: `runHookSession` integration (directive line on valid roadmap; silent exit-0 on missing AND malformed roadmap).
  - Partial: `loadActiveWorkflows` worktree branch; `RenderContextCapped` more>0 branch.
  - Open decision for qa-senior: none blocking. The combined-matcher choice is settled (verified); integration test should assert the wired SessionStart settings block shape per the plan.
