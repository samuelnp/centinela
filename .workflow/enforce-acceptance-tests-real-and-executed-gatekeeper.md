### Gatekeeper Report: enforce-acceptance-tests-real-and-executed
**Date:** 2026-04-07
**Status:** WARNING

#### Analyzed Specs
- Reviewed all feature specs under `specs/*.feature` (39 total), with primary focus on:
- `specs/enforce-acceptance-tests-real-and-executed.feature`
- `specs/edge-case-subagent-tests-phase.feature`
- `specs/bootstrap-phase-zero-workflow.feature`
- `specs/enforce-step-subagent-orchestration.feature`
- `specs/orchestration-smoke-sim.feature`
- `specs/enforce-coverage-in-validate.feature`
- `specs/add-ci-validate-workflow.feature`

#### Findings
- **Affected spec:** `specs/edge-case-subagent-tests-phase.feature`
- **Affected scenario:** `Tests phase passes with edge-case report`
- **Risk:** Scenario implies tests step passes when test artifacts plus edge-case report exist, but `internal/workflow/validate_tests.go` now also requires executable acceptance artifacts and acceptance execution command presence.
- **Suggestion:** Update scenario wording to include executable acceptance artifacts and validate-command acceptance execution.

- **Affected spec:** `specs/bootstrap-phase-zero-workflow.feature`
- **Affected scenario:** `Non-bootstrap tests step ignores placeholder files`
- **Risk:** Partial overlap with new comment-only and no-op acceptance checks can drift across specs.
- **Suggestion:** Harmonize wording around one canonical "real + executable acceptance artifacts" rule.

- **Affected spec:** `specs/enforce-acceptance-tests-real-and-executed.feature`
- **Affected scenario:** `Tests step fails when acceptance execution command is missing`
- **Risk:** Command detection is heuristic in `internal/workflow/validate_tests_acceptance_commands.go` and can allow broad commands that may not always represent explicit acceptance runtime.
- **Suggestion:** Consider stricter command mapping by acceptance runtime or explicit command key.

- **Affected spec:** `specs/enforce-step-subagent-orchestration.feature`
- **Affected scenario:** `Tests step requires QA evidence pair`
- **Risk:** Validation order reports acceptance/command errors before orchestration evidence errors in multi-failure states.
- **Suggestion:** Aggregate tests-step failures or document precedence clearly.

#### Recommendation
- Proceed as **WARNING** (not blocking). Align related tests-step specs with expanded acceptance requirements and harden acceptance-command detection semantics.
