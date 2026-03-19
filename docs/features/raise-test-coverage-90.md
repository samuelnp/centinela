# Feature Brief: Raise Test Coverage Above 90%

## Problem
Current global coverage is around 10% when measured with `go test ./... -coverpkg=./...`.
Most core packages still have zero direct statement coverage.

## Goal
Increase measurable project coverage to above 90% with deterministic tests.

## Scope
- Add missing unit/integration/acceptance tests for core internal packages.
- Focus on pure and deterministic logic first.
- Add stable coverage command for validation.

## Non-Goals
- Large production refactors unrelated to testability.
- Changing workflow semantics.

## Acceptance Criteria
- Coverage command reports >90% statements.
- Existing test suite remains green.
- New tests cover previously untested branches in key packages.
