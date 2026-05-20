# Plan: Roadmap Definition Checkpoint Prompt

## Problem

After every roadmap-definition artifact exists, `centinela hook setup` returns
silently and Claude has no explicit instruction to ask the user whether to keep
iterating on the roadmap or start the first Phase 0 feature. The user wants
unlimited time to iterate; the prompt must therefore be one-shot until either
(a) the user picks "iterate" and a marker file is written, or (b) the user
picks "start" and a workflow file is created by `centinela start`. If the user
later edits roadmap artifacts, the marker becomes stale and the prompt must
re-fire.

## Solution

1. New package `internal/roadmapcheckpoint/` (n-tier supporting domain) with:
   - `RequiredArtifacts() []string` — canonical list of roadmap-defining files
     (`ROADMAP.md`, `.workflow/roadmap.json`,
     `.workflow/roadmap-analysis.{md,json}`,
     `.workflow/roadmap-quality.{md,json}`).
   - `Decide(now time.Time, fs FS) Decision` — pure decision function:
     returns one of `DecisionEmit`, `DecisionSuppressed`, `DecisionStale`.
     Inputs: marker existence + timestamp, mtimes of required artifacts,
     bootstrap completeness, first incomplete Phase 0 feature.
   - `WriteMarker(path string, now time.Time)` and `Marker` struct
     (`{"choice":"iterate","at":"<RFC3339>"}`) at
     `.workflow/roadmap-checkpoint.json`.
   - `FirstIncompleteBootstrap(r *roadmap.Roadmap) (string, bool)` — uses
     `roadmap.BootstrapFeatures` + `roadmap.FeatureStatus`.
2. New UI renderer `internal/ui/render_roadmap_checkpoint.go` exposing
   `RenderRoadmapCheckpoint(featureName string) string` — system panel with the
   two-option prompt + recovery hints. No business logic; only formatting.
3. Wire into `cmd/centinela/hook_setup.go` after the existing production-readiness
   check: load roadmap, call `roadmapcheckpoint.Decide`, and on
   `DecisionEmit | DecisionStale`:
   - Print `CENTINELA DIRECTIVE: roadmap checkpoint ...` line.
   - Print `ui.RenderRoadmapCheckpoint(first)`.
   - Return nil (do not block).
4. Suppression rules (handled in `Decide`):
   - If bootstrap is already complete (no incomplete Phase 0 feature) →
     `DecisionSuppressed` (never emit).
   - If marker exists and its `at` >= max(mtime of required artifacts) →
     `DecisionSuppressed`.
   - If marker exists but `at` < max(mtime) → `DecisionStale` (emit).
   - If marker absent → `DecisionEmit`.
   - If a workflow file for the first incomplete Phase 0 feature already
     exists (`.workflow/<feature>.json`) → `DecisionSuppressed` (user already
     started).

## Files Changed / Added

- `internal/roadmapcheckpoint/checkpoint.go` — package types + `Decide`,
  `Marker`, decision constants.
- `internal/roadmapcheckpoint/artifacts.go` — `RequiredArtifacts`,
  mtime helpers.
- `internal/roadmapcheckpoint/firstfeature.go` —
  `FirstIncompleteBootstrap(r *roadmap.Roadmap) (string, bool)`.
- `internal/ui/render_roadmap_checkpoint.go` — `RenderRoadmapCheckpoint`.
- `cmd/centinela/hook_setup.go` — append checkpoint branch after the
  production-readiness check; remain a thin orchestrator.
- `tests/unit/roadmapcheckpoint/decide_test.go` — emit / suppress / stale.
- `tests/unit/roadmapcheckpoint/firstfeature_test.go` — picks first
  non-done Phase 0 feature; respects status.
- `tests/integration/hook_setup_checkpoint_test.go` — drives `runHookSetup`
  with the full artifact set + marker variants and asserts panel/directive
  output.
- `tests/acceptance/roadmap_checkpoint_prompt_test.go` — Gherkin-backed end-to-end.
- `specs/roadmap-checkpoint-prompt.feature` — acceptance scenarios.

## Rollout Sequence

1. Smallest correct slice — emit + suppress on marker presence (no freshness):
   land `Decide` covering emit / suppress paths, marker writer, directive,
   renderer. Tests for those two outcomes only.
2. Layer in freshness check (mtime comparison). Add `DecisionStale` path and
   tests for "marker present but stale".
3. Layer in suppression when the first Phase 0 workflow file already exists
   (natural "start" path).
4. Polish renderer copy and ensure mtime-granularity edge case is covered by a
   documented test ("re-firing on no-op save is acceptable").

## Risks & Mitigations

- Directive spam if marker logic is wrong → unit tests own every decision
  path; integration test asserts a second `runHookSetup` with unchanged disk
  state stays silent.
- Bootstrap already complete on adoption → `Decide` returns
  `DecisionSuppressed` whenever `FirstIncompleteBootstrap` returns false.
- First Phase 0 feature missing from roadmap → renderer falls back to
  guidance text; no directive emitted (treated as already-complete).
- mtime granularity / no-op saves bumping mtime → accepted trade-off, called
  out in tests and feature brief.
- Cross-layer leak — keep `cmd/centinela/hook_setup.go` ≤100 lines by
  delegating all decision logic to `internal/roadmapcheckpoint`.
