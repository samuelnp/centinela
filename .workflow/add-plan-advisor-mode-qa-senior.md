# QA Senior — Tests Validation

- Feature: add-plan-advisor-mode
- Step: tests
- Role: qa-senior
- Status: done

## Inputs
- specs/add-plan-advisor-mode.feature
- cmd/centinela/hook_plan_advisor.go
- internal/planadvisor/advisor.go
- internal/planadvisor/coverage.go
- internal/planadvisor/questions.go

## Outputs
- tests/unit/add_plan_advisor_mode_unit_test.go
- tests/integration/add_plan_advisor_mode_integration_test.go
- tests/acceptance/add_plan_advisor_mode_test.go
- .workflow/add-plan-advisor-mode-edge-cases.md

## Edge Cases
- advisor mode must stay silent outside the plan step
- advisor mode must avoid repeating already documented topics
- user-facing plan prompts must include UX/mobile-first questions only when missing

## Handoff
- handoffTo: validate
