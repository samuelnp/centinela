# Plan: Refactor Hook Policy Core

## Scope
Extract prewrite hook policy from CLI command handler into shared internal logic while keeping behavior backward compatible.

## Work Items
1. Introduce shared policy package:
- Add decision function that evaluates file path, workflows, and current step.
- Return structured decision (`allow` or `block`) with context for rendering.

2. Adapt CLI hook:
- Keep stdin parsing and process exit concerns in `cmd/centinela/hook_prewrite.go`.
- Delegate classification and workflow checks to shared policy package.

3. Add tests:
- Unit tests for policy cases (no workflow, wrong step, allowed file types, roadmap).
- Regression coverage for existing blocked/allowed behavior.

## Validation
- `go test ./...`
- Spot check with `centinela hook prewrite` input payloads.

## Compatibility
- No changes to `centinela init`, workflow state format, or gate logic.
