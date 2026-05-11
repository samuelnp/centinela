# Orchestration Evidence: senior-engineer

- Feature: `extract-agent-shared-blocks`
- Step: `code`
- Outcome:
  - **A** Created `docs/architecture/agent-invocation.md` (46 lines) — single canonical Agent-tool invocation reference plus the `.workflow/<feature>-<role>.{md,json}` artifact contract.
  - **B** Replaced the verbose "How to Invoke" body in nine prompt files with a one-line reference to `agent-invocation.md`: `gatekeeper-prompt.md`, `edge-case-tester-prompt.md`, `production-readiness-prompt.md.template`, `big-thinker-prompt.md`, `feature-specialist-prompt.md`, `senior-engineer-prompt.md`, `qa-senior-prompt.md`, `ux-ui-specialist-prompt.md`, `validation-specialist-prompt.md`. `documentation-generator-prompt.md` had no `## How to Invoke` section and was correctly skipped. `ux-ui-specialist-prompt.md` keeps its "skip if not user-facing" qualifier inline.
  - **C** Removed the duplicate `## Decision Rules` table from `gatekeeper-prompt.md` (the SAFE/WARNING/BLOCKING decisions remain expressed in the Output Format Recommendation block). Gatekeeper went from 81 → 69 lines.
  - **D** Created `docs/architecture/stack-checks-reference.md` (43 lines) with the four-language example matrix; replaced the inline matrix in `production-readiness-prompt.md.template` with a single placeholder + cross-link. Template went from 95 → 90 lines.
  - Mirrored every canonical edit and the two new files into `internal/scaffold/assets/docs/architecture/`. The only diff remaining under `internal/scaffold/assets/docs/architecture/` is the pre-existing `gatekeepers.md` drift, unrelated to this feature.
  - No Go code modified.
- Line-budget impact: gatekeeper −12 lines, production-readiness.template −5 lines, edge-case-tester +1 line (wrap), six promoted prompts unchanged in line count but centralised through a single shared invocation reference.
- Handoff: `qa-senior`
