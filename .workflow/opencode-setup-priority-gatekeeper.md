### Gatekeeper Report: opencode-setup-priority
**Date:** 2026-04-24
**Status:** WARNING

#### Analyzed Specs
- `specs/opencode-setup-priority.feature`
- `specs/clarify-roadmap-missing-artifacts.feature`
- `specs/docs-latest-features-getting-started.feature`
- `specs/enforce-acceptance-tests-real-and-executed.feature`
- `specs/g1-justified-file-size-exceptions.feature`
- `specs/simplify-output-prefix-emojis.feature`
- `specs/configurable-step-confirmation-mode.feature`
- `specs/roadmap-quality-overall-threshold.feature`
- `specs/add-personality-feedback.feature`
- `specs/enforce-plan-snapshot-inputs.feature`
- `specs/add-docs-step-workflow.feature`
- `specs/improve-docs-llm-hybrid-ui.feature`
- `specs/claude-status-line.feature`
- `specs/generate-html-project-docs.feature`
- `specs/roadmap-senior-pm-analysis.feature`
- `specs/orchestration-smoke-sim.feature`
- `specs/docs-readme-bootstrap-tutorial.feature`
- `specs/docs-update-migrate-readme.feature`
- `specs/migrate-full-sync.feature`
- `specs/enforce-step-subagent-orchestration.feature`
- `specs/auto-start-feature-intent.feature`
- `specs/fix-release-workflow-run-tag-resolution.feature`
- `specs/fix-release-trigger-after-bump.feature`
- `specs/harden-main-release-automation.feature`
- `specs/reach-100-coverage.feature`
- `specs/refactor-hook-policy-core.feature`
- `specs/improve-centinela-render-ui.feature`
- `specs/opencode-hook-parity.feature`
- `specs/raise-test-coverage-90.feature`
- `specs/fix-setup-hook-template-detection.feature`
- `specs/harden-opencode-plugin-compat.feature`
- `specs/fix-roadmap-write-blocked.feature`
- `specs/fix-setup-next-step.feature`
- `specs/fix-validate-plan-by-name.feature`
- `specs/docs-migration-managed-docs.feature`
- `specs/edge-case-subagent-tests-phase.feature`
- `specs/enforce-coverage-in-validate.feature`
- `specs/bootstrap-phase-zero-workflow.feature`
- `specs/docs-consistency-pass.feature`
- `specs/add-ci-validate-workflow.feature`
- `specs/automate-semver-release.feature`
- `specs/adapt-opencode-support.feature`

#### Findings
- **Affected spec:** `specs/adapt-opencode-support.feature`
  - **Affected scenario:** Existing OpenCode config is preserved
  - **Risk:** If setup priority forced `AGENTS.md` ahead of user-owned instruction entries, it would break the preservation guarantee for existing OpenCode setups.
  - **Suggestion:** Preserve existing non-Centinela instruction order and only guarantee Centinela-managed ordering (`AGENTS.md` before `CLAUDE.md`).
- **Affected spec:** `specs/auto-start-feature-intent.feature`
  - **Affected scenario:** Prompt intent auto-starts workflow
  - **Risk:** If setup priority short-circuited prompt processing, autostart, orchestration, or context hooks could stop running after setup is complete.
  - **Suggestion:** Keep the priority change scoped to setup and migration ordering while continuing downstream prompt hooks.
- **Affected spec:** `specs/docs-consistency-pass.feature`
  - **Affected scenario:** Agent support wording is accurate
  - **Risk:** Claude-only onboarding wording in setup docs would remain inconsistent with the new OpenCode behavior.
  - **Suggestion:** Update root and scaffolded setup docs to use agent-neutral wording in the docs step.

#### Recommendation
- WARNING: Potential conflicts found. Document risks and proceed with caution.
