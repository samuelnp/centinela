# Feature Brief: Enforce Coverage in Validate

## Problem
Coverage improved significantly, but nothing prevents regressions over time.
`centinela validate` currently checks built-in gates and configured commands, but there is no stable project-level coverage threshold.

## Goal
Add a repeatable coverage threshold check to the standard validation flow.

## Scope
- Define a project coverage command that runs in `centinela validate`.
- Fail validation when coverage drops below an agreed threshold.
- Document the command for local and CI usage.

## Acceptance Criteria
- `centinela validate` runs a coverage command.
- Validation fails when coverage is below threshold and passes when above.
- Documentation explains how to run and update the threshold.
