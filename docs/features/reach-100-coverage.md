# Feature Brief: Reach 100% Statement Coverage

## Problem
Coverage is currently above 90% but still below 100%, with remaining gaps concentrated in CLI edge paths and hard-to-test execution branches.

## Goal
Increase statement coverage to 100% while keeping behavior unchanged.

## Scope
- Add targeted tests for remaining uncovered branches.
- Introduce minimal testability seams where needed (without changing external behavior).
- Keep file-size and architecture rules intact.

## Acceptance Criteria
- `go test ./... -coverpkg=./... -coverprofile=coverage.out` reports 100.0%.
- Existing workflow and command behaviors remain unchanged.
- All tests pass in `go test ./...`.
