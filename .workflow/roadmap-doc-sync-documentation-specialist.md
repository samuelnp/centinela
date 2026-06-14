# Documentation-Specialist Report — roadmap-doc-sync

Internal-surface (right-sized) docs step.

## Outputs
- `.workflow/roadmap-doc-sync-changelog.md` — one-line feat changelog.
- Regenerated `docs/project-docs/index.html` from current artifacts.

## User-facing note
`ROADMAP.md` is now generated: edit `.workflow/roadmap.json` (descriptions/fixes/notes/intro) then run `centinela roadmap generate`. The `roadmap_drift` gate (warn by default) flags hand-edits to ROADMAP.md; ratchet `[gates.roadmap_drift] severity` to `fail` once adopted.

Handoff → complete.
