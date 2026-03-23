# Plan: Reach 100% Statement Coverage

## Baseline
- Re-run coverage profile and list remaining uncovered functions.

## Work Items
1. Cover low/zero CLI wrappers (`main`, `runValidate`, `runStatus`, model update branches).
2. Add subprocess helper tests for `os.Exit` and hook-blocking paths.
3. Add tests for remaining branch conditions in `workflow`, `setup`, `roadmap`, and UI helpers.
4. Add minimal test seams only where direct execution is not test-safe.

## Validation
- `go test ./...`
- `go test ./... -coverpkg=./... -coverprofile=coverage.out`
- `go tool cover -func=coverage.out`
