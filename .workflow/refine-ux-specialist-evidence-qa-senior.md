# QA Senior — Tests Validation

- Feature: refine-ux-specialist-evidence
- Step: tests
- Role: qa-senior
- Status: done

## Inputs
- specs/refine-ux-specialist-evidence.feature
- internal/orchestration/evidence.go
- internal/orchestration/evidence_ux.go

## Outputs
- tests/unit/refine_ux_specialist_evidence_unit_test.go
- tests/integration/refine_ux_specialist_evidence_integration_test.go
- tests/acceptance/refine_ux_specialist_evidence_test.go
- .workflow/refine-ux-specialist-evidence-edge-cases.md

## Edge Cases
- ux-ui-specialist evidence must fail when mobileFirst is missing or false
- ux-ui-specialist evidence must fail when required UX tags are missing
- non-UX roles must keep their previous evidence contract

## Handoff
- handoffTo: validate
