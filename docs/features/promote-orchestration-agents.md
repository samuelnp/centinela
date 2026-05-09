# promote-orchestration-agents

## Problem

The agent-performance-audit feature documented an asymmetry: four
report-producing subagents (gatekeeper, production-readiness,
edge-case-tester, documentation-generator) have full Markdown prompt files
under `docs/architecture/`, but the six orchestration roles wired into
`internal/orchestration/policy.go` (big-thinker, feature-specialist,
senior-engineer, ux-ui-specialist, qa-senior, validation-specialist) live
only as ~150-char strings embedded in `internal/setup/opencode_agent_config.go`.
Plan, code, tests, and validate steps therefore run on terse one-liners,
while only docs-step report agents enjoy structured guidance.

## User Stories

- As a Centinela maintainer, I want every orchestration role to have a
  full prompt file with the same structure as `gatekeeper-prompt.md` so
  prompt evolution and review are uniform.
- As a Centinela user, I want consistent specialist quality across all
  workflow steps, not just docs/validate reports.
- As a new-project bootstrap consumer, I want the scaffold to ship the
  same six prompt files so my project starts with full guidance.

## Acceptance Criteria

- Six new prompt files exist under `docs/architecture/`:
  `big-thinker-prompt.md`, `feature-specialist-prompt.md`,
  `senior-engineer-prompt.md`, `qa-senior-prompt.md`,
  `ux-ui-specialist-prompt.md`, `validation-specialist-prompt.md`.
- Each new file contains `## Purpose`, `## Prompt Template`, and
  `## Required Artifact` sections.
- Each new file is mirrored byte-identically under
  `internal/scaffold/assets/docs/architecture/`.
- Each new file is ≤ 70 lines.
- The CLAUDE.md Quick Reference table lists each new prompt.
- No changes to `internal/setup/opencode_agent_config.go`,
  `internal/orchestration/policy.go`, or
  `cmd/centinela/hook_orchestration.go` (additive doc-only change).
- Acceptance test asserts the six new files exist with the required
  section headings.

## Edge Cases

- `ux-ui-specialist` runs only when a feature declares
  `surface: user-facing` in its brief; the prompt should call this out so
  invocation is correct.
- `validation-specialist` orchestrates the existing gatekeeper and
  production-readiness subagents; its prompt must cross-link to them
  without restating their content.
- The scaffold mirror tree is byte-identical to `docs/architecture/` for
  the existing four prompts; the new six must preserve that invariant.
- Each role's required artifact path is `.workflow/<feature>-<role>.md`
  (and `.json` companion where the directive system expects evidence) —
  matches the existing evidence file convention emitted by
  `cmd/centinela/hook_orchestration.go:42`.

## Risks

- Adding too much prose per prompt would grow per-invocation context cost
  rather than reducing it; length budget of ≤ 70 lines per file mitigates.
- Drift between `docs/architecture/` and the scaffold mirror; the
  validate step diff check enforces parity.
- Restating CLAUDE.md or PROJECT.md content inside the new prompts would
  inflate context with no benefit; prompts should reference, not copy.
