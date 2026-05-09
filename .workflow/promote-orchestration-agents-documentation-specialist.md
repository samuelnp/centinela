# Orchestration Evidence: documentation-specialist

- Feature: `promote-orchestration-agents`
- Step: `docs`
- Outcome: Validated documentation inputs (`centinela docs validate` clean), regenerated `docs/project-docs/index.html` so the project documentation reflects the six new orchestration prompt files. Updated CLAUDE.md Quick Reference table during the code step (six new rows pointing at each promoted prompt). No source-code changes — documentation-only.
- Highlights:
  - Six new prompt files documented under `docs/architecture/`: `big-thinker-prompt.md`, `feature-specialist-prompt.md`, `senior-engineer-prompt.md`, `qa-senior-prompt.md`, `ux-ui-specialist-prompt.md`, `validation-specialist-prompt.md`.
  - Mirrored byte-identically under `internal/scaffold/assets/docs/architecture/` for new-project bootstrap.
  - Each prompt declares `## Purpose`, `## Prompt Template`, `## Required Artifact` and respects the 70-line per-file budget.
  - Acceptance test `tests/acceptance/promote_orchestration_agents_acceptance_test.go` enforces existence, sections, mirror-identity, and budget.
  - Production-readiness prompt at `docs/architecture/production-readiness-prompt.md` rendered from the template earlier in this session (Centinela project values: Go, Cobra, n-tier, Claude Code + OpenCode hooks).
- Handoff: closeout. Run `centinela complete promote-orchestration-agents` to mark the workflow done.
