### QA-Senior Report: session-context-rehydration
**Date:** 2026-05-23

> Note: the test suite below was authored by the QA-Senior subagent pass; that
> pass was interrupted (session limit) before writing this evidence manifest.
> The manifest was reconstructed from the on-disk artifacts after independently
> verifying every claim: `go vet ./...` clean, `go test ./...` green,
> `CI=true centinela validate` → G1 pass + coverage 95.1% ≥ 95.0%, gofmt clean,
> and all internal/ & cmd/ test files ≤100 lines (G1).

#### Test Inventory
| Tier | File | Lines | Covers |
|------|------|-------|--------|
| unit (consumer) | tests/unit/session_context_active_unit_test.go | 104 | ActiveWorkflows: reject evidence/ad-hoc/done, dedupe, mtime sort |
| unit (consumer) | tests/unit/session_context_capactive_unit_test.go | 51 | CapActive: above/at-below cap, max<=0 no-cap |
| unit (consumer) | tests/unit/session_context_firstincomplete_unit_test.go | 81 | FirstIncomplete cross-phase/all-done/nil; FirstNotDone predicate |
| unit (consumer) | tests/unit/session_context_render_unit_test.go | 74 | RenderSessionRehydration payload/pointers/complete; RenderContextCapped +N more |
| unit (in-pkg) | internal/workflow/active_test.go | 71 | ActiveWorkflows/CapActive internal coverage |
| unit (in-pkg) | internal/roadmap/firstincomplete_test.go | 62 | FirstIncomplete/FirstNotDone internal coverage |
| unit (in-pkg) | internal/ui/render_session_test.go | 60 | RenderSessionRehydration internal coverage |
| integration | cmd/centinela/hook_session_test.go | 86 | runHookSession: valid/all-done/missing/invalid roadmap |
| integration | cmd/centinela/hook_context_panel_test.go | 87 | panel: evidence-JSON not active, +N more past cap |
| integration | cmd/centinela/hook_workflows_test.go | 51 | loadActiveWorkflows delegation/scoping |
| acceptance | tests/acceptance/session_context_rehydration_test.go | 272 | 10 funcs mapping all 11 .feature scenarios |

#### Coverage Gaps
- None against the spec: all 11 scenarios have executable assertions at the
  acceptance tier (the SessionStart Scenario Outline over startup|clear|compact|resume
  is asserted as source-independent in one function).
- Coverage gate restored the right way (real tests, gate untouched):
  **95.1% >= 95.0%** under `CI=true centinela validate` (full scan).

#### Acceptance Wiring
- `centinela.toml` `validate.commands` runs `go test ./...`, which compiles and
  executes `tests/acceptance/` — no change needed.

#### Handoff
- Next role: validation-specialist
- Edge-case report: `.workflow/session-context-rehydration-edge-cases.md`
- Note for validate step: keep the stable spec literals intact
  (`CENTINELA DIRECTIVE: session rehydration`, `+N more`, pointer paths
  `PROJECT.md` / `docs/features/<next>.md`, cap = 5).
