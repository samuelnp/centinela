# Orchestration Evidence: documentation-specialist

- Feature: `extract-agent-shared-blocks`
- Step: `docs`
- Outcome: Validated documentation inputs (`centinela docs validate` clean), regenerated `docs/project-docs/index.html` so the project documentation reflects the shared `agent-invocation.md` reference and the moved-out `stack-checks-reference.md` content. No CLAUDE.md changes needed: the Quick Reference table still points at individual prompts (which now in turn point at the shared invocation file, so a reader following any prompt link reaches the shared content in one extra hop).
- Highlights:
  - Two new shared reference files at `docs/architecture/agent-invocation.md` (46 lines) and `docs/architecture/stack-checks-reference.md` (43 lines).
  - Nine prompt files now reference `agent-invocation.md` from their `## How to Invoke` section.
  - `gatekeeper-prompt.md` slimmed from 81 → 69 lines (duplicate Decision Rules table removed; equivalent decisions remain in the Output Format Recommendation block).
  - `production-readiness-prompt.md.template` slimmed from 95 → 90 lines (four-language stack matrix moved to the shared reference).
  - Per-invocation context win is concentrated on the gatekeeper prompt (~100 tokens per validate-step run); a smaller win on production-readiness.
  - All eleven scaffold mirrors stay byte-identical with their canonicals.
- Handoff: complete
