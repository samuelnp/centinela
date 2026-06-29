### Gatekeeper Report: cost-governance
**Date:** 2026-06-29
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
- specs/improve-centinela-render-ui.feature
- specs/improve-docs-llm-hybrid-ui.feature
- specs/landing-page.feature
- specs/lean-evidence-footprint.feature
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
New cost-governance soft gate: `[cost]` config, `internal/cost` aggregator
(transcript/cursor/aggregate/budget/report), a `cost-sample` telemetry event,
a Stop-hook capture, `centinela cost`, and a non-failing validate ⚠. Also fixed
stale lean-evidence-footprint gitignore tests left red by f138f90.

#### Checklist
- [x] All source + test files ≤100 lines (G1 gate green; manually verified the 11 new source + 10 new test files).
- [x] No cross-layer import violations: `internal/cost` joins the aggregator layer, importing the `telemetry`+`config` leaves only; imported solely by `cmd/` (Report type by `internal/ui`). `import_graph` warning has an empty body (benign, pre-existing).
- [x] `centinela validate` passes: G1, Cross-Compile (6 targets), `go test ./...`, acceptance suite, coverage (95.1% ≥ 95.0%), fmt — all green. roadmap_drift regenerated.
- [x] No business logic in the outer layer (cmd/ only wires; logic lives in `internal/cost`).
- [x] i18n: CLI tool, house-style strings consistent with siblings (insights/calibration); no new locale surface.
- [x] Soft-gate contract honored: over-budget NEVER changes the exit code (asserted by `TestAccCostCaptureReportAndSoftGate`).
- [x] Back-compat: zero config = silent no-op; old telemetry lines lack token fields → read as 0.

#### Findings
- `spec-traceability-gate` and `import_graph` warnings are emitted with empty
  bodies (non-blocking, pre-existing in diff-aware mode). The two acceptance
  scenarios carry `// Acceptance:` + `// Scenario:` markers; remaining feature
  scenarios are covered at the unit/integration tiers.
- No correctness/security/data risk: every failure mode (no transcript, parse
  error, disabled, no active feature) is a silent no-op; telemetry is
  append-only and non-blocking; the gate cannot block `complete`.

#### Recommendation
- **SAFE** — additive soft gate, all gates green, soft-never-blocks contract
  verified end-to-end, and the f138f90 test regression repaired along the way.
