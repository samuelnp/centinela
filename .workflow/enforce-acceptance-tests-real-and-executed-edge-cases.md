# Edge Cases: enforce-acceptance-tests-real-and-executed

## Covered

- Acceptance files with only comments are rejected.
- Acceptance files with placeholder no-op patterns (`if false`, `t.Skip`, `TODO`) are rejected.
- Tests step fails when `validate.commands` lacks acceptance execution.
- Existing acceptance files with executable assertions continue to pass.

## Residual Risks

- Heuristic content detection may require future keyword tuning for uncommon frameworks.
