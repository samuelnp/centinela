# Plan: promote-orchestration-agents

## Scope

Create six new Markdown prompt files for the orchestration roles wired
into `internal/orchestration/policy.go` so every step has structured
specialist guidance equivalent to the existing report-agent prompts.
Doc-only, additive change. No Go code is modified.

## Work Items

1. Author six new prompt files under `docs/architecture/`, modelled on
   `edge-case-tester-prompt.md` (concise) and `gatekeeper-prompt.md`
   (richer):
   - `big-thinker-prompt.md` (plan step)
   - `feature-specialist-prompt.md` (plan step)
   - `senior-engineer-prompt.md` (code step)
   - `qa-senior-prompt.md` (tests step)
   - `ux-ui-specialist-prompt.md` (code step, conditional on
     `surface: user-facing`)
   - `validation-specialist-prompt.md` (validate step)
2. Mirror each new file byte-identically under
   `internal/scaffold/assets/docs/architecture/`.
3. Each file must contain `## Purpose`, `## Prompt Template`, and
   `## Required Artifact` sections; per-file ≤ 70 lines.
4. Update the CLAUDE.md Quick Reference table to list the six new prompts.
5. Add an acceptance test (`tests/acceptance/`) asserting that each new
   prompt file exists, parses as Markdown, and contains the three
   required headings.
6. Write `.workflow/promote-orchestration-agents-edge-cases.md` with
   the qa-senior + edge-case-tester analysis.

## Validation

- `go test ./...` passes.
- `centinela validate` passes (gatekeeper + readiness + tests).
- `diff -r docs/architecture internal/scaffold/assets/docs/architecture`
  shows no drift for the new files.
- `wc -l docs/architecture/{big-thinker,feature-specialist,senior-engineer,qa-senior,ux-ui-specialist,validation-specialist}-prompt.md`
  reports ≤ 70 each.

## Compatibility

- Additive doc-only change. No runtime behaviour shifts.
- `internal/setup/opencode_agent_config.go` remains the runtime source of
  truth for OpenCode subagent registration; the new prompt files are
  human-facing structured guidance the orchestrator (Claude Code) reads
  when delegating.
- `documentation-specialist` already has
  `documentation-generator-prompt.md` so it is **not** in scope here.
- A future follow-up could refactor the Go code to load these prompt
  files at runtime (single source of truth), but that is out of scope.
