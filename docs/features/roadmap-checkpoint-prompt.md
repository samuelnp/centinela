# Feature: Roadmap Definition Checkpoint Prompt

## Problem

Once Centinela's project bootstrap has produced every roadmap-definition artifact
(`PROJECT.md`, `ROADMAP.md`, valid `.workflow/roadmap.json`,
`.workflow/roadmap-analysis.{md,json}`, `.workflow/roadmap-quality.{md,json}`, and
`docs/architecture/production-readiness-prompt.md`), `centinela hook setup` falls
silent. The user is left without an explicit handoff and Claude defaults to its
next behavioural instinct — sometimes jumping into the first Phase 0 feature,
sometimes just answering the user's message — instead of explicitly asking
whether the user wants to keep iterating on the roadmap or start implementing.

## Outcome

Emit one new directive from `centinela hook setup` after the existing
roadmap-definition gates pass, telling Claude to ask:

> "Roadmap definition iteration complete. Continue iterating on the roadmap,
> or start implementing the first incomplete Phase 0 bootstrap feature
> `<feature-name>`?"

Make the directive idempotent: once the user picks "iterate", a marker file
suppresses the prompt; if the user later edits any roadmap-defining artifact,
the marker becomes stale and the prompt re-fires. Picking "start" naturally
suppresses the prompt because the workflow file for the first Phase 0 feature
will exist.

## Scope

- New decision module under `internal/roadmapcheckpoint/` covering:
  emit-decision, marker file shape, freshness check (mtime vs. checkpoint).
- New directive branch in `cmd/centinela/hook_setup.go` (thin orchestration).
- New UI panel renderer `internal/ui/render_roadmap_checkpoint.go`.
- Dynamic lookup of the first incomplete Phase 0 feature via
  `roadmap.BootstrapFeatures` + `roadmap.FeatureStatus`.
- Unit, integration, and acceptance tests for emit/suppress/re-fire.

## Out of Scope

- Persisting checkpoint history across projects.
- Automatically re-running senior-PM analysis or quality scoring when ROADMAP.md
  changes — user still drives those steps.
- Multi-phase checkpoints (only Phase 0 is gated here).
- Programmatically detecting user "iterate" vs "start" answers — the
  orchestrator (Claude) writes the marker and runs `centinela start`.
