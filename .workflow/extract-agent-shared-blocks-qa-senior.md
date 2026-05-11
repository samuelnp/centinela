# Orchestration Evidence: qa-senior

- Feature: `extract-agent-shared-blocks`
- Step: `tests`
- Outcome: Added `tests/acceptance/extract_agent_shared_blocks_acceptance_test.go` with six sub-tests covering shared-file existence, the contract content of `agent-invocation.md`, per-prompt references to the shared file, removal of the gatekeeper Decision Rules duplicate, the production-readiness template's stack-matrix migration, and scaffold-mirror parity for every file touched by the feature. Existing acceptance tests `TestPromoteOrchestrationAgents_*` and `TestEdgeCaseSubagentPrompt_DocIncludesRequiredSections` continue to pass — the Tier 2 cuts deliberately avoided changing the heading set or the output-format strings those tests assert against. Wrote `.workflow/extract-agent-shared-blocks-edge-cases.md` capturing higher-level risks (re-introduction of boilerplate, accidental deletion of shared files, mirror drift).
- Handoff: `validation-specialist`
