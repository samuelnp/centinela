# QA Senior — Tests Validation

- Feature: claude-status-line
- Step: tests
- Role: qa-senior
- Status: done

## Inputs
- specs/claude-status-line.feature
- cmd/centinela/hook_statusline_test.go

## Outputs
- tests/unit/claude_statusline_unit_test.go
- tests/integration/claude_statusline_integration_test.go
- tests/acceptance/claude_statusline_acceptance_test.go
- .workflow/claude-status-line-edge-cases.md

## Edge Cases
- no active workflow
- malformed workflow entries missing feature or step
- tests step without edge-case report

## Handoff
- handoffTo: validate
