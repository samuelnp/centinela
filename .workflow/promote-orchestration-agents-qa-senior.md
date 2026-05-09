# Orchestration Evidence: qa-senior

- Feature: `promote-orchestration-agents`
- Step: `tests`
- Outcome: Added `tests/acceptance/promote_orchestration_agents_acceptance_test.go` with four assertions covering the four observable acceptance scenarios — existence of all six promoted prompt files, presence of the three required section headings (`## Purpose`, `## Prompt Template`, `## Required Artifact`), byte-identity between canonical and scaffold-mirror copies, and the per-file line budget of 70. Wrote `.workflow/promote-orchestration-agents-edge-cases.md` capturing the higher-level risks (drift, future role/prompt mismatch, restatement). The "runtime configuration unchanged" scenario is enforced by the absence of changes to `internal/setup/opencode_agent_config.go`, `internal/orchestration/policy.go`, and `cmd/centinela/hook_orchestration.go`; not a runtime assertion.
- Handoff: `validation-specialist`
