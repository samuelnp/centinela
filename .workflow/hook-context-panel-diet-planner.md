# hook-context-panel-diet — planner

## Problem

`centinela hook context` runs on every `UserPromptSubmit` and its output is
injected into the model's context. A bare run against this repo today
(single active workflow, plan step) measures only 234 bytes — the brief's
1770/1034/736(41.6%) figures did **not** reproduce as-is, because none of the
per-turn nudge panels happen to be active for this exact workflow right now
(its own brief already exists; its plan artifacts are not yet complete).
Re-measured directly against the brief's actual scenario — the
`FEATURE BRIEF REQUIRED` nudge active, which is the state a workflow spends
most of its plan step in — by calling `RenderContextCapped` +
`RenderReviewReady` + `RenderFeatureBriefNeeded` together: **2203 bytes
total, 1249 bytes of content, 954 bytes (43.3%) trailing whitespace**. This
confirms the brief's mechanism and order of magnitude even though the exact
numbers depend on which nudge panel is active on a given turn.

Root cause, precisely: `remove-panel-borders` (commit `3ab3cef`, already on
`main`) removed the *outer* fixed-width bordered box that used to square
every panel to a terminal column width. It left the *inner*
`lipgloss.JoinVertical(lipgloss.Left, ...)` calls inside each panel-body
constructor untouched — `JoinVertical` still right-pads every line to the
width of that call's longest line. That residual, per-panel padding is the
entire remaining leak; there is no bordered rectangle left to square.

## Scope

In scope: the six padding-bearing render functions `hook_context.go` calls —
`RenderContextCapped`, `RenderReviewReady`, `RenderFeatureBriefNeeded`,
`RenderEdgeCaseReportNeeded`, `RenderDocumentationNeeded`,
`RenderChangelogNeeded` — plus a deterministic size guard. Out of scope:
`hook_prewrite_block.go`, `hook_merge.go`, `hook_migrate.go`, and the
`hook setup` family, which share the identical padding root cause but fire
on rare/gated events, not every turn (see Deferred Findings). Also out of
scope: `hook statusline` and `hook session-start` (verified live: already
carry ~zero trailing whitespace), and any change to CLI-only renderers
(`RenderStatus`, `RenderRoadmap`).

## Dependencies & Assumptions

- Depends on `remove-panel-borders` (merged, `3ab3cef`, 2026-06-29) having
  already removed the outer bordered box — confirmed via
  `git show 3ab3cef -- internal/ui/panel.go`.
- Assumes no non-hook caller exists for the six functions being touched —
  verified by grep (`grep -rn "<Func>(" --include='*.go' . | grep -v
  _test.go`), each has exactly one caller, `hook_context.go`.
- Assumes `lipgloss.JoinVertical(lipgloss.Left, ...)` is the sole source of
  the residual padding — confirmed by finding zero `Width(`/`Border(` calls
  anywhere in non-test `internal/ui`/`cmd/centinela` source.
- Assumes no ANSI/colour handling is needed — verified live under both a
  non-tty pipe and a real PTY (`pty.openpty()` harness): zero `\x1b` bytes
  either way.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|---|---|---|---|
| Regressing human-readable `centinela status`/`validate`/blocked-write output | High (violates the brief's hard constraint) | Low | Fix scoped to functions with zero non-hook callers; `RenderStatus`/`RenderRoadmap`/`RenderBlocked` untouched; spec scenario asserts byte-identical CLI output |
| Breaking the byte-identity tests `token-diet` was assumed to have added | Medium if they existed | **None found** — exhaustive grep across `cmd/centinela/*_test.go`, `internal/ui/*_test.go`, `tests/acceptance/token_diet_*.go` for `mustEqual`/`assertEqual`/`require.Equal`/exact-length/`.golden` patterns turned up zero matches; all existing hook/panel tests are `strings.Contains`-based | Documented as an honest finding, not assumed away; tests step must still re-run the full suite after code lands |
| Losing a governance signal the model needs (active feature, step, blocking action) | High | Low, if content audit is followed | No content removed this feature, only trailing whitespace; step ladder kept specifically because archetype step sets differ (canonical/hotfix/refactor/spike) |
| `render.go` trips G1 (100-line cap) on the first line added | Medium (blocks the code step) | Medium if implemented carelessly | Plan specifies a same-line wrap of the existing `return` statement (0 net lines, 0 new imports — `trimTrailingWS` lives in the same `ui` package) |
| Size guard flakes across machines/repo age | Medium | Low if built correctly | Guard fixture is fixed in-memory (`workflow.Workflow{Feature:"sample-feature",...}`), never the live `.workflow/` directory or `docs/features/*.md` count |

## Rollout

1. **Slice 1 (padding removal, independently revertible):** add
   `trimTrailingWS` to `internal/ui/panel.go`; wrap the return statements of
   the six padding-bearing functions across `render.go`, `render_review.go`,
   `render_brief.go`. No emitter (`hook_context.go`) changes needed.
2. **Slice 2 (size guard):** `internal/ui/panel_budget_test.go`, tests step —
   asserts zero trailing whitespace (zero-tolerance) and a byte budget of
   1249 measured + ~12% headroom (≤1400 bytes) against the fixed fixture.
3. **Slice 3 (regression audit):** re-run `go test ./internal/ui/...
   ./cmd/centinela/...` after slice 1 lands; the plan-step audit found no
   byte-identity assertions to update, but the tests-step agent must confirm
   this still holds against the actual diff.

Full detail, including the emitter inventory table and the exact G1
boundary numbers, is in `docs/plans/hook-context-panel-diet.md`.

## Behavior Summary

`centinela hook context` continues to emit the ACTIVE WORKFLOWS panel and
any due nudge panels (feature-brief-required, review-required, edge-case,
documentation, changelog) with identical content and identical *leading*
whitespace/structure, but with zero *trailing* whitespace on any line. The
active feature name, its current step, its progress count, and its full
archetype-specific step ladder still render every time. Any due blocking
instruction (e.g. "STOP. Do not advance.") still renders and still names the
feature it applies to. Nothing changes for `centinela status`, `centinela
validate`, or blocked-write refusal output — those keep their current bytes,
trailing whitespace included, unchanged.

## Acceptance Criteria (Gherkin)

See `specs/hook-context-panel-diet.feature` — 11 scenarios covering: no
trailing whitespace on the active-workflows panel and on each nudge panel;
the feature/step/progress/step-ladder governance signal surviving,
including cross-archetype step-ladder disambiguation; a blocking
review-required instruction still naming its feature; `centinela status`
and blocked-write output staying byte-identical; the size guard passing at
budget, failing when padded past budget, and being independent of the live
workflow count.

## UX States

- **No active workflow:** unchanged — `RenderSuccess("No active
  workflows.")` is a single line via `renderSystemLine`, never padded, not
  touched by this feature.
- **Single active workflow, no nudge due:** ACTIVE WORKFLOWS panel only,
  trailing whitespace removed (measured 234B→~210B live today).
- **Single active workflow, one nudge due** (feature-brief-required,
  review-required, edge-case, documentation, or changelog): ACTIVE
  WORKFLOWS panel + one nudge panel, both trailing-whitespace-free (measured
  2203B→1249B for the heaviest combination).
- **Multiple active workflows (up to the cap of 5), several with due
  nudges:** each workflow's section and each due nudge panel individually
  trailing-whitespace-free; nudge panels still self-identify by feature name
  since they render in a flat loop, not nested under their workflow header.
- **>5 active workflows:** unchanged cap-and-"+N more" behavior
  (`workflow.CapActive`), untouched by this feature.
- **Human TTY (`centinela status`, `validate`, blocked-write refusal):**
  byte-for-byte unchanged, trailing whitespace included, in every state.

## Edge Cases

See evidence `edgeCases` (7 entries) — archetype step-order disambiguation,
the `render.go` G1 boundary, the size-guard fixture determinism, the
byte-identity-test honest-negative finding, CLI-only renderers staying
untouched, and the deferred sibling emitters.

## Out-of-Scope

- `hook_prewrite_block.go` (`RenderBlocked`, `RenderBlockedStaleBinary`),
  `hook_merge.go`, `hook_migrate.go`, and the `hook setup` family — same
  root cause, not the measured every-turn leak (see Deferred Findings).
- Any content removal beyond trailing whitespace — audited and found
  load-bearing (archetype step ladder, per-panel feature self-identification,
  the feature-brief checklist for a model that hasn't seen it this session).
- ANSI/colour handling — verified zero escape bytes on this path, live,
  under both a pipe and a real PTY.
- Changing what hooks decide (gate/directive/step semantics) — presentation
  only, per the brief's constraint.

## Deferred Findings

- `hook-panel-diet-remaining-emitters` (new backlog candidate): apply the
  identical `trimTrailingWS` wrap to `RenderBlocked`,
  `RenderBlockedStaleBinary`, `RenderMergeStewardNeeded`,
  `RenderMigrationNeeded`, and the seven `hook setup` panel functions. Same
  fix shape, deferred because none fire on every turn, so none are part of
  this feature's measured leak — folding them in here would be scope
  unjustified by measurement.
- `centinela status`/`RenderStatus` and other CLI-only renderers carry the
  identical residual `JoinVertical` padding (confirmed live) and are, in
  principle, just as dead now that no border remains — but the brief's
  constraint is explicit that this feature changes hook output only, so
  this is noted, not actioned.

## Handoff

Next role: **senior-engineer**. Implement Slice 1 exactly as specified in
`docs/plans/hook-context-panel-diet.md` §2 and §6 — `trimTrailingWS` in
`internal/ui/panel.go`, same-line return wraps in `render.go` (G1: exactly
100 lines today, net-zero-line edit required),
`render_review.go`, `render_brief.go`. No `hook_context.go` changes needed
for Slice 1.

