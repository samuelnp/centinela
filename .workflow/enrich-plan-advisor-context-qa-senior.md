# QA Senior — Tests Validation

- Feature: enrich-plan-advisor-context
- Step: tests
- Role: qa-senior
- Status: done

## Inputs
- specs/enrich-plan-advisor-context.feature
- internal/planadvisor/advisor.go
- internal/planadvisor/context.go
- internal/planadvisor/context_summary.go
- internal/planadvisor/roadmap_context.go

## Outputs
- tests/unit/enrich_plan_advisor_context_unit_test.go
- tests/integration/enrich_plan_advisor_context_integration_test.go
- tests/acceptance/enrich_plan_advisor_context_test.go
- .workflow/enrich-plan-advisor-context-edge-cases.md

## Edge Cases
- roadmap dependencies must outrank same-phase siblings
- related edge-case lessons must shape questions without prompt bloat
- missing roadmap artifacts must fall back to local feature context

## Handoff
- handoffTo: validate
