### Gatekeeper Report: session-context-rehydration
**Date:** 2026-05-23
**Status:** SAFE

#### Analyzed Specs
Reviewed all 64 `.feature` files in `specs/` against the new feature surface
(NEW `internal/workflow/active.go`, `internal/roadmap/firstincomplete.go`,
`internal/ui/render_session.go`, `cmd/centinela/hook_session.go`; MODIFIED
`cmd/centinela/hook_workflows.go`, `cmd/centinela/hook_context.go`,
`internal/ui/render.go`, `internal/roadmapcheckpoint/firstfeature.go`,
`internal/setup/{hooks.go,settings_build.go}`, `.claude/settings.json`).
Specs with a plausible contractual overlap, examined in depth:

- `improve-centinela-render-ui.feature` — `RenderContext` / active-workflows panel
- `claude-status-line.feature` — statusline + hook-wiring ("existing hooks should remain configured")
- `roadmap-checkpoint-prompt.feature` — `FirstIncompleteBootstrap` Phase-0 contract
- `opencode-hook-parity.feature` — hook command set parity
- `auto-start-feature-intent.feature`, `add-plan-advisor-mode.feature`,
  `configurable-step-confirmation-mode.feature`, `edge-case-subagent-tests-phase.feature`,
  `fix-roadmap-write-blocked.feature`, `refactor-hook-policy-core.feature` — all
  reference `hook context` / active workflows / hook wiring.

The remaining specs operate on unrelated surfaces (release automation, docs
generation, opencode setup flow, coverage/evidence gates, worktrees, merge
steward) and share no entity, port, DTO, or rendered-literal contract with this
feature.

#### Findings

- **Affected spec:** `improve-centinela-render-ui.feature`
  - **Affected scenario:** "Context output separates status from action-required notices"
  - **Risk:** A changed `RenderContext` signature or a renamed/removed
    "ACTIVE WORKFLOWS" branded panel would break the integration test
    (`tests/integration/improve_centinela_render_ui_integration_test.go`), which
    calls `ui.RenderContext([]*workflow.Workflow{wf})` and asserts the panel
    contains `"ACTIVE WORKFLOWS"` and the system brand.
  - **Suggestion / Resolution:** No change required. `RenderContext(wfs)`
    signature is UNCHANGED; it now delegates to `RenderContextCapped(wfs, 0)`.
    With `more == 0` the body is byte-identical to the prior panel (no `+N more`
    line, same `renderSystemPanel("HOOK", "ACTIVE WORKFLOWS", …)`). The dedupe /
    evidence-leak / cap logic lives upstream in `internal/workflow` and `cmd/`,
    not in the renderer, so this scenario stays green. Verified.

- **Affected spec:** `edge-case-subagent-tests-phase.feature` (and the
  `edge_case_context_integration_test.go` regression it backs)
  - **Affected scenario:** edge-case-report reminder during the `tests` step
  - **Risk:** The new `ActiveWorkflows` filter could drop a legitimately-saved
    workflow, suppressing the per-step reminders that drive `hook context`.
  - **Suggestion / Resolution:** No change required. The regression test saves a
    single real `f.json` whose `Feature == "f"` and `CurrentStep == "tests"`.
    `ActiveWorkflows` accepts a file ONLY when `wf.Feature == <basename>` and
    `CurrentStep` is non-empty and non-`done` — this fixture passes the guard, so
    the workflow still surfaces and the `Edge-case report missing` reminder still
    fires. The reminder loops in `hook_context.go` iterate the full `wfs` slice
    (not the capped `shown`), so capping never suppresses a reminder. Verified.

- **Affected spec:** `roadmap-checkpoint-prompt.feature`
  - **Affected scenario:** all 12 scenarios (Phase-0-only first-incomplete contract)
  - **Risk:** Refactoring `FirstIncompleteBootstrap` to share the not-done
    predicate could leak Phase-1+ features into the Phase-0-only checkpoint
    target, breaking "Next feature is the first incomplete across all phases" /
    "Multiple Phase 0 features, only the first is done".
  - **Suggestion / Resolution:** No change required. `FirstIncompleteBootstrap`
    still iterates `roadmap.BootstrapFeatures(r)` only (Phase-0 scoped) and merely
    delegates the per-feature predicate to the new shared `roadmap.FirstNotDone`.
    The set of candidate features is unchanged; only the duplicated "status !=
    done" check was extracted. The new `roadmap.FirstIncomplete` (all-phase walk)
    is a SEPARATE function used exclusively by the new `hook session` path and
    does not touch the checkpoint code path. Phase-0-only contract preserved.
    Verified by reading `internal/roadmapcheckpoint/firstfeature.go` and
    `internal/roadmap/firstincomplete.go`.

- **Affected spec:** `claude-status-line.feature`
  - **Affected scenario:** "Claude setup wires statusLine command" → "existing
    hooks should remain configured"
  - **Risk:** Threading a new `SessionStart` key through `buildHookSettings`
    could drop or reshape the existing `PreToolUse` / `PostToolUse` /
    `UserPromptSubmit` blocks.
  - **Suggestion / Resolution:** No change required. `settings_build.go`
    unmarshals, merges, and re-marshals the three existing keys exactly as before
    and ADDS `SessionStart` alongside them — additive only. The package test
    `TestMergeHooksAndIdempotency` already asserts `pre==2, post==2, prompt==7,
    session==1` and that a second merge is a no-op (idempotent), and
    `TestInjectHooksCreatesAndPreserves` confirms `statusLine` and existing hooks
    survive. Both are green in the full suite. Verified.

- **Affected spec:** `opencode-hook-parity.feature`
  - **Affected scenario:** "Plugin invokes setup and context on prompt submit"
  - **Risk:** The spec enumerates required hook commands; an exclusive-set
    assertion would conflict with adding `hook session`.
  - **Suggestion / Resolution:** No change required. The spec asserts that
    specific commands (`prewrite`, `postwrite`, `setup`, `context`) ARE invoked;
    it makes no closed-set / count assertion. `hook session` is a new Claude
    `SessionStart` hook and does not alter the OpenCode plugin's prompt/edit
    lifecycle. No conflict. Verified.

#### Recommendation
- SAFE: No conflicts detected. The two load-bearing invariants that other specs
  depend on are both preserved by construction: (1) `RenderContext`'s signature
  and its "ACTIVE WORKFLOWS" branded output are unchanged (the dedupe/cap/
  evidence-leak fix lives upstream in `internal/workflow` + `cmd/`, never in the
  renderer), and (2) `FirstIncompleteBootstrap` keeps its Phase-0-only scope by
  iterating `BootstrapFeatures` and only sharing the per-feature `FirstNotDone`
  predicate — the all-phase `FirstIncomplete` is a separate function on the new
  `hook session` path. The `SessionStart` wiring is purely additive to
  `.claude/settings.json` and `buildHookSettings`. Proceed.
