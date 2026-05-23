# Documentation-Specialist Report: roadmap-checkpoint-prompt

**Date:** 2026-05-23
**Step:** docs (final)
**Role:** documentation-specialist

## Knowledge Base entry

Wrote `docs/project-docs/kb/roadmap-checkpoint-prompt.md` (audience: end-user,
status: done) in plain language, no Given/When/Then or engineering jargon.

- **Summary:** Once the roadmap is fully defined, Centinela asks you once
  whether to keep refining the roadmap or start building the first feature,
  instead of leaving Claude to guess.
- **Sections:** the three required sections — `## What it does`,
  `## When you'd use it`, `## How it behaves` (one bullet per spec scenario) —
  plus an optional `## Examples` showing the `centinela roadmap iterate` and
  `centinela start <feature>` commands and the rendered checkpoint panel.

## Roadmap dependencies

The checkpoint only fires after the full roadmap-definition artifact set
exists and is valid: `PROJECT.md`, `ROADMAP.md`, a valid `.workflow/roadmap.json`,
plus the senior-PM analysis (`roadmap-analysis.{md,json}`) and quality
(`roadmap-quality.{md,json}`) reports. It depends on the existing roadmap
bootstrap chain to have completed: missing `ROADMAP.md` or invalid
`roadmap.json` take precedence and the checkpoint stays silent until those are
resolved. The first incomplete Phase 0 bootstrap feature is resolved
dynamically from the roadmap; once it is started or all Phase 0 features are
done, the checkpoint goes quiet.

## Workflow status matrix

| Step      | Status | Evidence |
|-----------|--------|----------|
| plan      | done   | docs/features/, docs/plans/, specs/roadmap-checkpoint-prompt.feature |
| code      | done   | senior-engineer report (osfs.go, render_roadmap_checkpoint.go, hook_setup.go, roadmap_iterate.go) |
| tests     | done   | qa-senior report + edge-cases; unit + integration + acceptance (12 scenarios + anti-spam regression) |
| validate  | done   | gatekeeper + `centinela validate` (lint + types + full suite, coverage gate ≥ 95%) |
| docs      | done   | this report + KB entry + generated HTML (`centinela docs generate` exit 0) |

## Major specs and scenario counts

`specs/roadmap-checkpoint-prompt.feature` — **12 scenarios**:

1. Happy path emits the checkpoint directive when no marker exists
2. Suppressed when the marker is fresh against all roadmap artifacts
3. Stale marker re-fires when `ROADMAP.md` was modified after the marker
4. Stale marker re-fires when any roadmap supporting artifact was modified after the marker
5. Suppressed when bootstrap is already complete
6. Suppressed when no Phase 0 bootstrap features exist in the roadmap
7. Suppressed when the workflow file for the first Phase 0 feature already exists
8. Precedence — missing roadmap-defining artifact lets the existing setup directives fire instead
9. Precedence — invalid `roadmap.json` yields the roadmap-json directive, not the checkpoint
10. Multiple Phase 0 features, only the first is done — picks the second as target
11. Malformed marker JSON is treated as missing and re-emits without crashing
12. Marker `at` field unparseable as RFC 3339 is treated as stale and re-emits

Each scenario is reflected as one user-visible behavior bullet in the KB
`## How it behaves` section.

## Generated outputs (all confirmed on disk)

- `docs/project-docs/kb/roadmap-checkpoint-prompt.md`
- `docs/project-docs/kb/roadmap-checkpoint-prompt.html`
- `docs/project-docs/kb/index.html`
- `docs/project-docs/index.html`

`centinela docs validate` exited 0; `centinela docs generate` exited 0. The
main `index.html` links the Knowledge Base and references this feature; the KB
index lists it. No production source was edited.
