### Gatekeeper Report: parallel-feature-worktrees
**Date:** 2026-05-15
**Status:** SAFE

#### Analyzed Specs

All `specs/*.feature` files were reviewed. Those with potential overlap to the
domains this feature touches (`internal/workflow/`, `internal/orchestration/`,
`internal/config/`, scaffold docs) were analyzed in depth:

- parallel-feature-worktrees.feature (the new spec under review)
- add-agent-evidence-contract.feature
- enforce-actionable-orchestration-evidence.feature
- enforce-step-subagent-orchestration.feature
- enforce-actionable-orchestration-evidence.feature
- promote-orchestration-agents.feature
- extract-agent-shared-blocks.feature
- diff-aware-gatekeeper.feature
- configurable-step-confirmation-mode.feature
- enforce-coverage-in-validate.feature
- raise-test-coverage-90.feature / reach-100-coverage.feature
- add-docs-step-workflow.feature
- The remaining 45 specs were scanned for shared-entity / DTO / workflow-state
  collisions; none touch `Workflow` JSON shape, the `Role` enum, the evidence
  validator, or `WorkflowConfig` in a way the new feature regresses.

#### Findings

**1. `Workflow` DTO shape change (`internal/workflow/state.go`)**
- **Affected spec:** every spec that exercises `centinela start` / `status` /
  `complete` (e.g. enforce-step-subagent-orchestration, add-docs-step-workflow,
  configurable-step-confirmation-mode).
- **Risk:** added field `WorktreePath`; legacy `.workflow/<feature>.json`
  written before this feature lacked the field.
- **Assessment:** NON-BLOCKING. The field is `json:"worktreePath,omitempty"`.
  `encoding/json` ignores missing keys on decode, so existing workflow JSON
  still decodes (zero value `""` = single-checkout flow, which `render_status`
  and `start.go` already treat as "no worktree"). No existing scenario asserts
  the absence of this key. Verified the full `go test ./...` suite is green,
  including `internal/workflow` state round-trip tests.

**2. `merge-steward` Role added to the orchestration enum (`policy.go`)**
- **Affected spec:** add-agent-evidence-contract.feature ("A canonical
  evidence contract document exists"; "Scaffold mirrors stay in sync"),
  enforce-actionable-orchestration-evidence.feature, promote-orchestration-agents.feature.
- **Risk:** a new `Role` constant could leak into `RequiredRoles(step)` and
  gate an existing workflow step, or break the evidence contract role list.
- **Assessment:** NON-BLOCKING. `RoleMergeSteward` is intentionally absent from
  `RequiredRoles` and `RequiredRolesForFeature` (confirmed by reading
  `policy.go` / `validate.go`); it only acquires meaning when
  `.workflow/<feature>-merge-steward.json` is written out-of-band by
  `centinela merge`. `add-agent-evidence-contract`'s acceptance test
  (`agent_evidence_contract_acceptance_test.go`) asserts the 7 in-workflow
  roles are *present* in `evidence-contract.md` — it neither requires nor
  forbids merge-steward, so the additive doc entry does not regress it (suite
  green).
- **Note (informational, not a conflict):** `promote-orchestration-agents.feature`
  scenario "Runtime configuration is unchanged" asserts
  `internal/orchestration/policy.go should be unchanged`. That spec describes a
  *completed historical* feature's acceptance criteria (a point-in-time
  guarantee for that change set), not a live regression invariant — there is no
  executing test enforcing policy.go immutability, and the full suite passes.
  Flagged only so it is on the record; it does not block this feature.

**3. Scaffold-mirror parity for the two docs this feature touched**
- **Affected spec:** add-agent-evidence-contract.feature ("Scaffold mirrors
  stay in sync with the docs prompts").
- **Risk:** editing `docs/architecture/evidence-contract.md` and adding
  `docs/architecture/merge-steward-prompt.md` without mirroring into
  `internal/scaffold/assets/docs/architecture/` would drift the scaffold and
  fail `TestScaffoldMirrorParityForUpdatedPrompts`.
- **Assessment:** RESOLVED / SAFE. Both feature-touched files were verified
  byte-identical to their scaffold mirrors:
  `evidence-contract.md` ✓ identical, `merge-steward-prompt.md` ✓ identical and
  present in the mirror tree. The `add-agent-evidence-contract` parity test
  passes.

**4. Spec-conflict pre-check (`internal/worktree/spec_conflicts.go`)**
- **Affected spec:** parallel-feature-worktrees.feature scenario "Spec conflict
  across in-flight worktrees is detected before merging".
- **Risk:** the detector reads every `specs/*.feature` from main plus each
  worktree; a false positive could block legitimate merges across the existing
  60 specs.
- **Assessment:** NON-BLOCKING. `collectScenarios` tags each record by owner
  and `scenariosConflicts` only flags same-Given/different-Then across
  *different* owners (verified via `worktree_spec_conflicts_test.go` cases:
  same-owner self-comparison is silent, same Given+Then is silent). It does not
  mutate any spec. No effect on existing scenarios.

#### Recommendation

- **SAFE: No conflicts detected. Proceed.** The two DTO/enum extensions
  (`WorktreePath`, `RoleMergeSteward`) are strictly additive and backward
  compatible (`omitempty` decode-safe; role excluded from `RequiredRoles`).
  Scaffold parity for the two feature-touched docs is clean. The unrelated
  full-tree `diff -r` drift (gatekeepers.md, new-project-guide.md,
  testing-strategy.md, workflow-enforcement.md, production-readiness-prompt.md)
  pre-dates this feature and is outside its scope — the scaffold-parity tests
  are file-scoped and all pass. The full `go test ./...` suite is green.
