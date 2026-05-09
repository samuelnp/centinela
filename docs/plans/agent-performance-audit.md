# Plan: agent-performance-audit

## Scope

Create a documentation audit of Centinela's configured OpenCode agents, prompt efficiency, and workflow step coverage.

## Work Items

1. Review generated OpenCode agents in `internal/setup/opencode_agent_config.go`.
2. Review orchestration step role mapping in `internal/orchestration/policy.go`.
3. Write an audit document with findings, context-trimming opportunities, and step coverage gaps.
4. Add the missing native OpenCode `validation-specialist` agent and map it to the validate step.
5. Add tests that assert the audit covers all configured agents and workflow steps.

## Validation

- `go test ./...`
- `centinela validate`

## Compatibility

- This audit does not change generated agent prompts.
- Any agent prompt changes should be handled by a follow-up feature after review.
