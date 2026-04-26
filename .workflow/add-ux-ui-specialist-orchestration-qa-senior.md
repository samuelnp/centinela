# QA Senior — Tests Validation

- Feature: add-ux-ui-specialist-orchestration
- Step: tests
- Role: qa-senior
- Status: done

## Inputs
- specs/add-ux-ui-specialist-orchestration.feature
- cmd/centinela/hook_orchestration.go
- internal/config/orchestration.go
- internal/orchestration/policy.go
- internal/orchestration/feature_surface.go
- internal/orchestration/output_ui.go

## Outputs
- tests/unit/add_ux_ui_specialist_orchestration_unit_test.go
- tests/integration/add_ux_ui_specialist_orchestration_integration_test.go
- tests/acceptance/add_ux_ui_specialist_orchestration_test.go
- .workflow/add-ux-ui-specialist-orchestration-edge-cases.md

## Edge Cases
- internal features must not require UX evidence
- user-facing code steps must block completion until UX evidence exists
- UX evidence must fail when outputs do not touch configured UI paths

## Handoff
- handoffTo: validate
