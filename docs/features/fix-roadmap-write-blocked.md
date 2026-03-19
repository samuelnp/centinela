---
feature: fix-roadmap-write-blocked
type: fix
---

# Fix: Roadmap Phase Writes Blocked by Prewrite Hook

During the roadmap conversation, Claude tries to write `docs/features/<slug>.md` feature briefs
before any feature workflow exists. The prewrite hook blocks this because it classifies
`docs/features/` as `TypePlan` and requires an active workflow in the plan step.

The fix introduces `TypeRoadmap` for roadmap-phase artifacts (`docs/features/`, `ROADMAP.md`,
`roadmap.json`) which are always allowed regardless of workflow state.
