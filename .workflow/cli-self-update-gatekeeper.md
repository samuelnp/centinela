### Gatekeeper Report: cli-self-update
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
- specs/cli-self-update.feature
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

#### Findings

**Finding 1 — hook_session.go: emitUpdateNotice() ordering vs empty-output tests**
- **Affected spec:** specs/session-context-rehydration.feature
- **Affected scenario:** "Missing roadmap is handled gracefully without crashing" and "Invalid roadmap json is handled gracefully without crashing"
- **Risk:** Both scenarios assert the session hook produces NO output when the roadmap is absent or invalid. With `emitUpdateNotice()` now running FIRST in `runHookSession`, any notice text emitted would invalidate that assertion. However, this risk is mitigated at the source: the test suite runs with `Version = "dev"` (the package-level default in main.go), and `Updater.isDev()` short-circuits `Notice()` with an immediate `return ""` — no network call, no output. The risk is real only if `Version` is overridden to a semver string in a test context without also injecting a fake `newSelfUpdater`. Current tests do not do this. Status: CONTAINED by dev-sentinel, not BLOCKING.
- **Suggestion:** None required for current state. If a future test in hook_session_test.go overrides `Version` to a real semver, it must also override `newSelfUpdater` to inject a fake Updater (the seam already exists in update.go). Document this constraint in a code comment on `emitUpdateNotice`.

**Finding 2 — hook_session.go: emitUpdateNotice() is synchronous (latency path)**
- **Affected spec:** specs/session-context-rehydration.feature (all SessionStart scenarios)
- **Affected scenario:** All "SessionStart injects the rehydration payload" scenarios
- **Risk:** `emitUpdateNotice()` is synchronous and runs before `roadmap.Load()`. On a cold cache with a slow GitHub API, this adds latency to every session startup. The TTL cache (24h default) and dev-sentinel make this a cold-path-only concern, but it is architecturally real: a 2–5 second GitHub timeout would delay every session hook after a cache expiry. The existing specs say nothing about latency bounds, so this does not fail a spec — but it is a UX risk.
- **Suggestion:** Consider running the notice check with a short explicit HTTP timeout (e.g. 3s) in the production Updater. The `Doer` interface already allows this. Not blocking.

**Finding 3 — G2 / centinela.toml: leaf registration is accurate**
- **Affected spec:** specs/g2-import-graph-gate.feature
- **Risk (confirmed absent):** Verified all non-test files in `internal/selfupdate/` import only stdlib packages (encoding/json, os, path/filepath, time, crypto/sha256, encoding/hex, io, net/http, strings, fmt, runtime). No internal imports exist. The `[[gates.import_graph.layers]]` registration of `internal/selfupdate/**` as `allow = []` is correct and does not loosen any existing layer's rules. The g2-import-graph-gate spec is unaffected.

**Finding 4 — Command name collision: none detected**
- **Affected spec:** none
- **Risk (confirmed absent):** Searched all `cmd/centinela/*.go` files. No existing `rootCmd.AddCommand` registers a command named "update". The `--check` flag on `updateCmd` uses `BoolVar` into the package-level `updateCheck` var — no collision with any other flag or var in the package. The Cobra `Use: "update"` string is unique.

**Finding 5 — Version variable: shared correctly, not mutated**
- **Affected spec:** none
- **Risk (confirmed absent):** `Version` is declared once in `main.go` and consumed by `evidence_init.go`, `update.go`, `hook_session.go`, and `evidence_schema.go`. The selfupdate package receives it as a parameter (never reads it as a global), so no aliasing or mutation risk exists. The ldflag injection path is unchanged.

**Finding 6 — exitMain(1) usage in runUpdate**
- **Affected spec:** specs/cli-self-update.feature (Scenario: --check reports a newer version)
- **Risk (confirmed correct):** `exitMain(1)` is the established project pattern for non-error exits that need a non-zero code (also used in main.go). The `update_paths_test.go` correctly captures this via `withExitCapture`. The Cobra `SilenceErrors: true` + `SilenceUsage: true` prevents double-printing.

**Finding 7 — G1 file-size compliance**
- All production files in `internal/selfupdate/`: largest is `replace.go` at 71 lines. All pass the 100-line limit.
- `cmd/centinela/update.go`: 55 lines. `cmd/centinela/hook_session.go`: 48 lines. Both pass.

**Finding 8 — Spec traceability: verified 25/25**
- The feature spec has exactly 25 scenarios (`grep -c "^  Scenario" specs/cli-self-update.feature` = 25).
- Acceptance test files in `tests/acceptance/cli_self_update_*.go` carry exactly 25 `// Acceptance: specs/cli-self-update.feature` traceability comments (verified by grep count).
- Spot-checked: AC1 (UpdateInstallsNewerRelease), AC2 (UpdateNoOp, CheckBehindReturnsMsg, CheckCurrentReturnsUpToDate, CheckHonorsTTLCache), AC3 (checksum mismatch), AC5 (permission denied), AC6 (5 notice scenarios), AC7 (infra scenarios), and all edge cases map 1:1 to named test functions with matching scenario names in comments. Traceability claim is genuine.

#### Deferred Findings
- none

#### Recommendation
SAFE: No conflicts detected. The `emitUpdateNotice()` insertion into `runHookSession` is strictly additive and fail-silent; the dev-sentinel (Version = "dev") guarantees zero output and zero network calls in all test contexts without injected fakes. `internal/selfupdate` is a genuine leaf with no internal imports; its centinela.toml registration is accurate. The `update` command name is unique. G1, G7, and G2 rules are all met. Spec traceability is verified 25/25. Proceed with implementation.
