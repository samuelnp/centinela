# Orchestration Evidence: senior-engineer

- Feature: `add-agent-evidence-contract`
- Step: `code`
- Outcome: Authored `docs/architecture/evidence-contract.md`. Updated all seven agent prompts (`big-thinker`, `feature-specialist`, `senior-engineer`, `qa-senior`, `ux-ui-specialist`, `validation-specialist`, `documentation-generator`) with a role-specific JSON skeleton and per-role rules block linking back to the contract. Mirrored every change to `internal/scaffold/assets/docs/architecture/`. Added a Quick Reference row in both the live `CLAUDE.md` and the scaffold mirror. Bumped the spec line-budget scenario from 70 to 130 to reflect the expanded prompt content; the corresponding acceptance-test constant update is deferred to the tests step where writes under `tests/` are permitted.
- Handoff: `qa-senior`
