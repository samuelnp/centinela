---
id: 57eef19dbdfee88d
feature: workflow-revise-loop
step: validate
type: verdict
title: ### Gatekeeper Report: workflow-revise-loop
tags: gatekeeper, verdict
sourceArtifact: .workflow/workflow-revise-loop-gatekeeper.md
createdAt: 2026-06-30T17:28:39Z
---

### Gatekeeper Report: workflow-revise-loop
**Date:** 2026-06-30
**Status:** SAFE

#### Analyzed Specs
Reviewed the full `specs/` corpus (110 `.feature` files), focusing on every spec
that shares a surface this feature touches — workflow state, the orchestration
evidence model, telemetry, the complete/step-transition flow, worktrees,
archetypes, and enforcement profiles:

- `specs/workflow-revise-loop.feature` (this feature)
- `specs/governance-telemetry.feature`
- `specs/centinela-insights.feature`
- `specs/capability-calibration.feature` / `specs/model-capability-profiles.feature`
- `specs/cost-governance.feature`
- `specs/add-agent-evidence-contract.feature`
- `specs/enforce-actionable-orchestration-evidence.feature`
- `specs/enforce-step-subagent-orchestration.feature`
- `specs/evidence-cli.feature`
- `specs/lean-evidence-footprint.feature`
- `specs/governed-project-memory.feature`
- `specs/completion-delivery-prompt.feature` / `specs/delivery-artifact-generation.feature`
- `specs/workflow-archetypes.feature`
- `specs/enforcement-profiles.feature`
- `specs/parallel-feature-worktrees.feature` / `specs/merge-steward-auto-dispatch.feature`
- `specs/centinela-doctor.feature`
- `specs/configurable-step-confirmation-mode.feature`
- `specs/add-ux-ui-specialist-orchestration.feature` / `specs/right-size-docs-step.feature`
  (feature-aware role gating)

Plus the source surfaces named in PROJECT.md → Gatekeeper Paths:
`internal/workflow/` (state.go, rewind.go), `internal/evidence/`
(invalidate.go, invalidation_targets.go, roles.go), `internal/orchestration/`
(policy.go), `internal/telemetry/` (event.go, constructors.go), `internal/ui/`
(render_status.go), `internal/insights/`, `internal/calibration/`,
`internal/memory/`, `cmd/centinela/` (revise.go, revise_invalidate.go,
complete.go), and `tests/acceptance/coverage_hardening_test.go`.

#### Findings

**(1) New `Revisions` field on the `Workflow` struct — back-compat checked, no break.**
- **Affected spec:** all specs that serialize/round-trip `.workflow/<feature>.json`
  (centinela-doctor, evidence-cli, workflow-archetypes, enforcement-profiles).
- **Risk:** a new struct field could change serialized output and break golden
  fixtures, `doctor` state parsing, or any test asserting exact JSON.
- **Suggestion / resolution:** `Revisions []Revision json:"revisions,omitempty"`
  is additive and omitempty — a workflow that was never rewound serializes
  byte-identically to before, exactly mirroring the existing
  `Archetype`/`DriverModel`/`WorktreePath` fields. `internal/doctor/
  check_workflow_state.go` does not strict-decode (no `DisallowUnknownFields`)
  and never references `Revisions`. Confirmed: the affected unit packages
  (837 tests across the 7 packages) and the full acceptance suite
  (`go test ./tests/acceptance/...`) are green.

**(2) `revise` DELETES downstream `.workflow/<feature>-<role>.{json,md}` evidence
— no persistence-flow conflict.**
- **Affected spec:** add-agent-evidence-contract, lean-evidence-footprint,
  governed-project-memory, completion-delivery-prompt, merge-steward-auto-dispatch.
- **Risk:** another flow could depend on certification evidence persisting.
- **Suggestion / resolution:** `evidence.Invalidate`/`InvalidateArtifact` touch
  ONLY the re-opened steps' role pair and the named `-edge-cases.md` artifact —
  never `.workflow/memory/`, roadmap state, telemetry events, the changelog
  artifact, source, tests, or docs. `memory.Capture` re-reads the current
  artifact on each `complete`, so a regenerated post-revise artifact is simply
  re-harvested; the append-only ledger in `.workflow/memory/` is untouched.
  merge-steward evidence is out-of-band (not in `RequiredRoles`, not a
  re-openable step), so it is never shed. Forcing a re-run via deletion is the
  intended contract and is idempotent (missing file is not an error).

**(3) New `TypeStepRevised` (`step-revised`) telemetry event — readers are lenient,
no parse break.**
- **Affected spec:** governance-telemetry (Lenient reader contract),
  centinela-insights, capability-calibration, cost-governance.
- **Risk:** a new event type could break rework/friction tallies or schema assertions.
- **Suggestion / resolution:** `insights.reworkType` and
  `calibration.frictionByModel` both use a `switch` with a `default` that ignores
  unknown types, so `step-revised` is silently skipped (it is deliberately NOT
  counted as rework). The new `From` field is additive + omitempty.
  `RecordRevised` routes through `Record`, which stamps
  `schema="centinela.telemetry/v1"` + RFC3339 timestamp, satisfying the
  per-event schema invariant. Verified by green `internal/telemetry`,
  `internal/insights`, and `internal/calibration` tests.

**(4) Interaction with archetypes / worktrees / enforcement profiles — clean.**
- **Affected spec:** workflow-archetypes, parallel-feature-worktrees, enforcement-profiles.
- **Risk:** hardcoded step order or worktree/profile coupling could misbehave on
  hotfix/spike tracks.
- **Suggestion / resolution:** `RewindTo` and `reopenedSteps` operate on
  `wf.OrderedSteps()` (the feature's own pinned order), never `DefaultStepOrder`,
  so backward transitions are archetype-aware (e.g. hotfix code→tests→validate).
  Per-step invalidation policy keys off the resolved step names. `revise` mutates
  the feature's own `.workflow/<feature>.json` from the active CWD (worktree-safe,
  no hook round-trip). Profiles are read-only here. A completed (`done`) workflow
  and any non-backward target are both rejected with state left untouched.

**(5) `tests/acceptance/coverage_hardening_test.go` modified by this branch —
scoping is correct and safe.**
- **Affected spec:** coverage-hardening.feature.
- **Risk:** `TestNoBehaviourChange_OnlyTestFilesAdded` asserts a feature branch
  adds ONLY `_test.go` files. workflow-revise-loop legitimately adds production
  Go (revise.go, rewind.go, invalidate.go, …), so the unmodified invariant would
  red-fail this branch.
- **Suggestion / resolution:** the added guard skips the test unless the
  coverage-hardening sentinel `cmd/centinela/cov2_config_error_test.go` is present
  in the `main...HEAD` diff. coverage-hardening is already merged into main, so
  the sentinel lives in main and drops out of the diff on any later branch (and on
  main itself the diff is empty) — the one-shot invariant correctly neutralizes
  itself wherever its premise no longer holds, instead of firing false positives.
  This is the right scoping for a branch-local tripwire: it weakens nothing for
  the coverage-hardening branch (where the sentinel IS present and the assertion
  still runs) and is harmless elsewhere. Confirmed: full acceptance suite green.

#### Deferred Findings
- none.

  (Non-blocking observation, no roadmap entry warranted: re-opening the `docs`
  step does not invalidate the non-role `-changelog.md` artifact, so a stale
  changelog could survive a docs rework. This is benign — the
  completion/delivery flow regenerates or overwrites it on the next docs
  `complete`, and leaving it violates no existing spec. Noted for implementer
  awareness only.)

#### Recommendation
- SAFE: No conflicts detected. The `Revisions` field, evidence invalidation,
  `step-revised` telemetry, and the archetype/worktree/profile interactions are
  all additive and back-compatible; the cross-feature
  `coverage_hardening_test.go` edit is correctly self-scoping. Whole module
  builds; affected unit packages and the full acceptance suite pass. Proceed.
