# Documentation-Specialist Report — capability-calibration

Internal-surface (right-sized) docs step.

## Outputs
- `.workflow/capability-calibration-changelog.md` — one-line feat changelog.
- Regenerated `docs/project-docs/index.html`.

## User-facing note
Telemetry events now carry the driver `model` (stamped at emission from the feature's pinned DriverModel). `centinela calibrate` groups telemetry per model and reports over/under/well-governed with an evidence-backed enforcement-profile recommendation (friction = rework/advances vs the model's capability-class profile; tighten/loosen/keep). Advisory only — it never auto-applies a config change. `--json` for tooling; empty log → clean empty-state exit 0.

Handoff → complete.
