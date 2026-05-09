# Plan: extract-agent-shared-blocks

## Scope

Reduce per-invocation context cost across the ten prompt files in
`docs/architecture/` by extracting shared boilerplate. Doc-only,
additive-plus-edits change. No Go code is modified.

## Work Items

1. Create `docs/architecture/agent-invocation.md` (≤ 30 lines): canonical
   Agent-tool invocation pattern, FEATURE_NAME placeholder convention,
   `.workflow/<feature>-<role>.{md,json}` artifact contract.
2. Update every prompt file's `## How to Invoke` section (where present)
   to a single-line reference to `agent-invocation.md`, replacing the
   repeated 2-4 line invocation paragraph.
   Affected: `gatekeeper-prompt.md`, `edge-case-tester-prompt.md`,
   `production-readiness-prompt.md.template`, `big-thinker-prompt.md`,
   `feature-specialist-prompt.md`, `senior-engineer-prompt.md`,
   `qa-senior-prompt.md`, `ux-ui-specialist-prompt.md`,
   `validation-specialist-prompt.md`.
   `documentation-generator-prompt.md` has no `## How to Invoke` section
   and is skipped.
3. Cut the `Decision Rules` table in `gatekeeper-prompt.md` (lines
   75-81). Equivalent SAFE / WARNING / BLOCKING decisions remain in the
   Output Format Recommendation block.
4. Create `docs/architecture/stack-checks-reference.md` (≤ 35 lines):
   the four-language example matrix moved out of the template.
5. Edit `docs/architecture/production-readiness-prompt.md.template` to
   keep a single-stack placeholder line that points at
   `stack-checks-reference.md`.
6. Mirror every edit and the two new files into
   `internal/scaffold/assets/docs/architecture/`.
7. Add acceptance test
   `tests/acceptance/extract_agent_shared_blocks_acceptance_test.go`
   that asserts: both new files exist; each affected prompt references
   `agent-invocation.md`; the gatekeeper duplicate table is gone;
   scaffold mirror parity for the affected files.
8. Write `.workflow/extract-agent-shared-blocks-edge-cases.md`.

## Validation

- `go test ./...` passes (existing tests for prompts must continue to
  pass; new acceptance test must pass).
- `centinela validate` passes (all gates).
- `diff -r docs/architecture internal/scaffold/assets/docs/architecture`
  shows no drift on the affected files.
- Token-savings spot-check via `wc -l docs/architecture/*-prompt.md*`
  comparing against pre-feature baseline (recorded in plan-step
  evidence).

## Compatibility

- Additive-plus-edit doc-only change. No runtime behaviour shift.
- `internal/setup/opencode_agent_config.go`,
  `internal/orchestration/policy.go`,
  `cmd/centinela/hook_orchestration.go`, and
  `internal/migration/header.go` are unchanged.
- `<!-- centinela:doc-version=… -->` headers are preserved; the
  manifest-based refactor that would let us drop them is tracked as a
  separate future feature.
- Existing acceptance tests
  (`TestEdgeCaseSubagentPrompt_DocIncludesRequiredSections`,
  `TestPromoteOrchestrationAgents_*`) continue to pass without
  modification.
