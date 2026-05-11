### Gatekeeper Report: extract-agent-shared-blocks
**Date:** 2026-05-11
**Status:** SAFE

#### Analyzed Specs
- specs/extract-agent-shared-blocks.feature (new)
- All existing `specs/*.feature` reviewed for entity / port / use-case / DTO conflicts.

#### Findings

No conflicts detected.

This feature is purely additive-plus-edits, doc-only:
- No domain entity in `internal/workflow/`, `internal/gates/`, `internal/orchestration/` is modified.
- No existing port or use-case interface changes.
- No DTO shape changes; `.workflow/<feature>-<role>.json` schema unchanged.
- No state-machine modifications.
- New files (`agent-invocation.md`, `stack-checks-reference.md`) are net-new content under `docs/architecture/` and the scaffold mirror only.
- Edited prompts (gatekeeper, edge-case-tester, production-readiness.template, six promoted prompts) keep their required headings (`## Purpose`, `## Prompt Template`, `## Required Artifact`) — existing acceptance tests `TestPromoteOrchestrationAgents_RequiredSections` and `TestEdgeCaseSubagentPrompt_DocIncludesRequiredSections` remain green.
- `documentation-generator-prompt.md` is unchanged.
- `internal/migration/header.go` is unchanged; doc-version comments preserved on all affected files.

#### Recommendation

SAFE: No conflicts detected. Proceed with implementation.
