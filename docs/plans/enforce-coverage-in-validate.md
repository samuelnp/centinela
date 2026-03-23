# Plan: Enforce Coverage in Validate

## Scope
Wire a coverage threshold command into the repository's validate flow and document how to use it locally and in CI.

## Work Items
1. Add a coverage script under `scripts/` that:
   - runs package coverage profile generation
   - extracts total coverage
   - compares against threshold (initially 95%)
   - exits non-zero when below threshold
2. Add `centinela.toml` `[validate].commands` entry for the new script.
3. Add convenience `make` target for coverage checks.
4. Update README/docs with coverage gate usage.

## Validation
- `go test ./...`
- `centinela validate` passes at current coverage.
- Simulated lower threshold failure path validated in tests.
