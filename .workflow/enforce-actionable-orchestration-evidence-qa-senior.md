# QA Senior — Tests Validation

- Feature: enforce-actionable-orchestration-evidence
- Step: tests
- Role: qa-senior
- Status: done

## Inputs
- specs/enforce-actionable-orchestration-evidence.feature
- internal/orchestration/evidence.go
- internal/orchestration/output_rules.go

## Outputs
- tests/unit/enforce_actionable_orchestration_evidence_unit_test.go
- tests/integration/enforce_actionable_orchestration_evidence_integration_test.go
- tests/acceptance/enforce_actionable_orchestration_evidence_test.go
- .workflow/enforce-actionable-orchestration-evidence-edge-cases.md

## Edge Cases
- summary-only outputs must fail even when role metadata is otherwise valid
- code-step evidence must not pass with `.workflow/`-only outputs
- docs-step evidence must continue to bypass actionable-output enforcement

## Handoff
- handoffTo: validate
