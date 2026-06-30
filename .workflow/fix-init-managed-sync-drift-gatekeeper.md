### Gatekeeper Report: fix-init-managed-sync-drift
**Date:** 2026-06-30
**Status:** SAFE

#### Analyzed Specs
- specs/adapt-opencode-support.feature
- specs/add-agent-evidence-contract.feature
- specs/add-ci-validate-workflow.feature
- specs/add-docs-step-workflow.feature
- specs/add-personality-feedback.feature
- specs/add-plan-advisor-mode.feature
- specs/add-ux-ui-specialist-orchestration.feature
- specs/adoption-baseline.feature
- specs/agent-performance-audit.feature
- specs/archetype-inference-project-synthesis.feature
- specs/audit-baseline-ratchet.feature
- specs/auto-start-feature-intent.feature
- specs/automate-semver-release.feature
- specs/bootstrap-phase-zero-workflow.feature
- specs/brownfield-roadmap-generation.feature
- specs/capability-calibration.feature
- specs/centinela-doctor.feature
- specs/centinela-insights.feature
- specs/claim-verification.feature
- specs/clarify-roadmap-missing-artifacts.feature
- specs/claude-status-line.feature
- specs/code-quality-hardening.feature
- specs/completion-delivery-prompt.feature
- specs/configurable-model-routing.feature
- specs/configurable-step-confirmation-mode.feature
- specs/configurable-subagent-models.feature
- specs/cost-governance.feature
- specs/cross-platform-build-gate.feature
- specs/custom-gate-sdk.feature
- specs/deep-codebase-analysis.feature
- specs/deferred-findings-roadmap-capture.feature
- specs/delivery-artifact-generation.feature
- specs/deterministic-artifact-scaffolds.feature
- specs/diff-aware-gatekeeper.feature
- specs/docs-consistency-pass.feature
- specs/docs-knowledge-base-pages.feature
- specs/docs-latest-features-getting-started.feature
- specs/docs-migration-managed-docs.feature
- specs/docs-readme-bootstrap-tutorial.feature
- specs/docs-update-migrate-readme.feature
- specs/edge-case-subagent-tests-phase.feature
- specs/enforce-acceptance-tests-real-and-executed.feature
- specs/enforce-actionable-orchestration-evidence.feature
- specs/enforce-coverage-in-validate.feature
- specs/enforce-plan-snapshot-inputs.feature
- specs/enforce-step-subagent-orchestration.feature
- specs/enforcement-profiles.feature
- specs/enrich-plan-advisor-context.feature
- specs/evidence-cli.feature
- specs/extract-agent-shared-blocks.feature
- specs/failure-ledger-plan-advisor.feature
- specs/fix-init-managed-sync-drift.feature
- specs/fix-release-trigger-after-bump.feature
- specs/fix-release-workflow-run-tag-resolution.feature
- specs/fix-roadmap-write-blocked.feature
- specs/fix-setup-hook-template-detection.feature
- specs/fix-setup-next-step.feature
- specs/fix-status-non-tty.feature
- specs/fix-validate-plan-by-name.feature
- specs/g1-justified-file-size-exceptions.feature
- specs/g2-import-graph-gate.feature
- specs/g2-multi-language-import-graph.feature
- specs/generate-html-project-docs.feature
- specs/governance-telemetry.feature
- specs/governed-project-memory.feature
- specs/harden-main-release-automation.feature
- specs/harden-opencode-plugin-compat.feature
- specs/headless-governance.feature
- specs/host-harness-adapters.feature
- specs/improve-centinela-render-ui.feature
- specs/improve-docs-llm-hybrid-ui.feature
- specs/landing-page.feature
- specs/lean-evidence-footprint.feature
- specs/mcp-governance-server.feature
- specs/merge-steward-auto-dispatch.feature
- specs/migrate-full-sync.feature
- specs/model-capability-profiles.feature
- specs/opencode-force-setup-flow.feature
- specs/opencode-greeting-workflow.feature
- specs/opencode-hook-parity.feature
- specs/opencode-native-subagents.feature
- specs/opencode-setup-priority.feature
- specs/opencode-setup-question-parity.feature
- specs/orchestration-smoke-sim.feature
- specs/parallel-feature-worktrees.feature
- specs/precommit-and-pr-gate.feature
- specs/promote-orchestration-agents.feature
- specs/raise-test-coverage-90.feature
- specs/reach-100-coverage.feature
- specs/readme-centinela-usage.feature
- specs/refactor-hook-policy-core.feature
- specs/refine-ux-specialist-evidence.feature
- specs/right-size-docs-step.feature
- specs/roadmap-checkpoint-prompt.feature
- specs/roadmap-doc-sync.feature
- specs/roadmap-parallel-readiness.feature
- specs/roadmap-quality-overall-threshold.feature
- specs/roadmap-senior-pm-analysis.feature
- specs/security-gate.feature
- specs/session-context-rehydration.feature
- specs/simplify-output-prefix-emojis.feature
- specs/spec-reconstruction.feature
- specs/spec-traceability-gate.feature
- specs/team-dashboard.feature
- specs/workflow-archetypes.feature

#### Scope of change
One-function bug fix: `cmd/centinela/init_agent.go` `setupOpenCode()` now routes
through the managed-sync seam (`BuildSyncPlan("opencode")` + `ApplySync`) instead
of the legacy header-less `EnsureAgentsFile`/`EnsureOpenCodePlugin`, so init writes
the managed-version-header'd content the migration system expects.

#### Checklist
- [x] All source + test files <=100 lines (G1 green).
- [x] No cross-layer violations: cmd wiring only; sync logic stays in internal/setup.
- [x] `centinela validate` green: G1, Cross-Compile, go test ./..., acceptance, coverage 95.0%, fmt; roadmap_drift in sync.
- [x] No business logic in outer layer.
- [x] i18n: init now prints managed-sync action lines (same as the Aider path); no new locale surface.
- [x] Root cause fixed and proven: fresh `init` -> `migrate` reports 0 pending; AGENTS.md/plugin carry the managed-version header. Regression test added (was previously untested, which is why the bug shipped).

#### Findings
- import_graph + spec-traceability warnings are empty-body, non-blocking, pre-existing in diff-aware mode.
- Legacy Ensure* writers retained (still unit-tested in internal/setup); init no longer calls them. The acceptance test guards against the drift returning.
- No data/security risk; pre-existing unmanaged files route to manual-review (no clobber).

#### Recommendation
- SAFE -- minimal, proven fix; mirrors the already-correct Aider path; init->migrate idempotency now pinned by an acceptance test.