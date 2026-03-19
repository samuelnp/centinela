# Plan: Raise Coverage Above 90%

## Baseline
- Run coverage with `go test ./... -coverpkg=./... -coverprofile=coverage.out`.
- Identify top untested packages and functions.

## Implementation Steps
1. Add tests for `internal/config`, `internal/gates`, `internal/roadmap`, `internal/scaffold`, and `internal/workflow/state`.
2. Expand tests for `internal/workflow/validate*` and UI render helpers.
3. Add command-path tests where practical (pure helpers first).
4. Re-run coverage and iterate until >90%.

## Validation
- `go test ./...`
- `go test ./... -coverpkg=./... -coverprofile=coverage.out`
- `go tool cover -func=coverage.out`
