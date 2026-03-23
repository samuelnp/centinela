# Plan: Add CI Validate Workflow

## Scope
Introduce a GitHub Actions workflow to run repository validation checks in CI.

## Work Items
1. Add `.github/workflows/validate.yml` with:
   - triggers: `push`, `pull_request`
   - Go setup and module cache
   - test execution (`go test ./...`)
   - validate execution (`go run ./cmd/centinela validate`)
2. Keep runtime fast and deterministic.
3. Document CI behavior in README.

## Validation
- `go test ./...`
- `go run ./cmd/centinela validate`
