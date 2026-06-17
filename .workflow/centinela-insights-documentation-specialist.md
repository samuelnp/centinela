# Documentation-Specialist Report — centinela-insights

Internal-surface (right-sized) docs step.

## Outputs
- `.workflow/centinela-insights-changelog.md` — one-line feat changelog.
- Regenerated `docs/project-docs/index.html`.

## User-facing note
`centinela insights` is read-only analytics over `.workflow/telemetry/events.jsonl`: most-triggered blocks, most-failed gates, features with most rework, mean steps-to-green. `--top N` (default 5) bounds each section; `--json` emits the structured Report for tooling. Empty/missing log → clean empty-state, exit 0.

Handoff → complete.
