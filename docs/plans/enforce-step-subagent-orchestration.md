# Plan: Enforce Step Subagent Orchestration

## Scope
Add strict orchestration validation for newly started workflows using role evidence artifacts.

## Work Items
1. Add orchestration policy package in `internal/orchestration/`.
2. Add evidence path and JSON validation logic.
3. Version workflow metadata so strict mode applies only to new workflows.
4. Integrate orchestration checks into `workflow.ValidateArtifacts`.
5. Add orchestration directive hook output and wire into prompt hooks.
6. Update OpenCode plugin generation and local plugin parity.
7. Add unit, integration, and acceptance tests.

## Validation
- `go test ./...`
- `go run ./cmd/centinela validate`

## Constraints
- Keep files under 100 lines.
- Keep business logic in `internal/`.
- JSON validation allows unknown fields; `checksum` remains optional.
