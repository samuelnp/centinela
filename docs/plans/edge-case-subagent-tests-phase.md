# Plan: Edge-Case Subagent for Tests Phase

## Scope
Add mandatory edge-case analysis artifact and guidance so tests phase covers hard paths, not just happy paths.

## Work Items
1. Add edge-case subagent prompt doc:
   - `docs/architecture/edge-case-tester-prompt.md`
   - scaffold copy under `internal/scaffold/assets/docs/architecture/`
2. Enforce tests-step artifact:
   - require `.workflow/<feature>-edge-cases.md` in workflow validation.
3. Add tests-step reminder in hook context when artifact is missing.
4. Update docs (`README.md`, `CLAUDE.md`, workflow/testing docs) with the new requirement.
5. Add unit/integration/acceptance tests for enforcement and reminders.

## Validation
- `go test ./...`
- `go run ./cmd/centinela validate`

## Output Artifact Format
`### Edge-Case Report: <feature>`
- Risk matrix (impact × likelihood)
- Missing scenarios
- Added/proposed tests (unit/integration/acceptance)
- Residual risks and mitigations
