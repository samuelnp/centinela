# Edge Cases: unified-plan-specialist

## Covered

- A pinned (`planner-v1`) workflow's plan gate rejects a hand-authored legacy
  `big-thinker` + `feature-specialist` pair and names only `planner` — never
  offers the legacy pair as an alternative.
  `tests/acceptance/unified_plan_specialist_complete_test.go:TestUPS_ForgedLegacyPairBlockedNamingPlanner`
- `evidence init <feature> big-thinker` is refused on a pinned workflow,
  naming the role "retired" and pointing at `planner`; no stub is written.
  `tests/acceptance/unified_plan_specialist_evidence_init_test.go:TestUPS_EvidenceInitRefusesRetiredRoleOnPinnedWorkflow`
- `evidence init <feature> feature-specialist` is refused when NO workflow
  state exists at all (not just when pinned) — `EnsureRoleAllowed` treats a
  missing/unreadable workflow as refused, and the CLI must reach that check
  before the generic "unknown feature" short-circuit.
  `tests/acceptance/unified_plan_specialist_evidence_init_test.go:TestUPS_EvidenceInitRefusesRetiredRoleWithNoWorkflow`
  `cmd/centinela/evidence_init_retired_test.go:TestEvidenceInitRefusesRetiredRoleWithNoWorkflowAtAll`
- A legacy (unpinned) workflow with only ONE of the two required legacy files
  still fails, naming both required roles plus a one-line contract annotation
  that this workflow predates `planner-v1` — and never offers `planner` as an
  alternative this workflow could satisfy.
  `tests/acceptance/unified_plan_specialist_legacy_test.go:TestUPS_LegacyPartialPairFailsWithContractAnnotation`
- The hook directive and the `complete` gate are driven through the SAME
  resolver (`workflow.RequiredEvidenceRoles`) for both a pinned and an
  unpinned workflow, so neither surface can name a role the other gate
  refuses (the PR #83 CRITICAL divergence class).
  `tests/acceptance/unified_plan_specialist_directive_gate_test.go:TestUPS_DirectiveAndGateAgree`
- A guided/outcome (non-strict) workflow: the resolver still names the single
  `planner` role (so a human reviewer or downstream tool asking "what does
  this workflow need" gets a correct answer), while the per-prompt hook
  prints NO evidence requirement for it at all — a non-strict workflow is
  never told to produce evidence it isn't gated on.
  `tests/acceptance/unified_plan_specialist_directive_gate_test.go:TestUPS_GuidedProfileStillNamesPlannerNoEvidencePrinted`
- A `centinela.toml` with only legacy `[orchestration.models]` keys
  (`big-thinker`, `feature-specialist`) still loads and resolves the
  `planner` role via aliasing; the deprecation notice prints at `doctor` and
  once at `start`, and is verified ABSENT from `hook orchestration` output
  (which runs on every prompt) so it never becomes permanent noise.
  `tests/acceptance/unified_plan_specialist_config_test.go:TestUPS_LegacyModelKeyAliasesToPlanner`
  `tests/acceptance/unified_plan_specialist_config_test.go:TestUPS_DeprecationNoticeAtDoctorAndStart_NeverOnHook`
- The plan-advisor never tags a question `[big-thinker]`/`[feature-specialist]`
  and never instructs delegating to two separate agents — both lenses render
  under one agent's header.
  `tests/acceptance/unified_plan_specialist_advisor_test.go:TestUPS_PlanAdvisorTwoLensesOneAgent`
- The merged `planner-prompt.md` never duplicates the CLI-authoring-rules
  block or any of `## Purpose`/`## Prompt Template`/`## Required Artifact`
  (a straight concatenation of the two legacy docs would have duplicated all
  four), stays within the 130-line budget, and the two legacy prompt docs
  (source AND scaffold mirror) are confirmed absent, not just "not asserted".
  `tests/acceptance/unified_plan_specialist_prompt_content_test.go:TestUPS_PlannerPromptBothLensesOrdered`
  `tests/acceptance/unified_plan_specialist_prompt_content_test.go:TestUPS_PlannerPromptLineBudget`
  `tests/acceptance/unified_plan_specialist_prompt_content_test.go:TestUPS_LegacyPromptDocsAbsent`
  `tests/acceptance/unified_plan_specialist_scaffold_parity_test.go:TestUPS_PlannerPromptScaffoldMirrorByteIdentical`
- A role sub-workflow accidentally saved under `<feature>-planner` (or the
  legacy `-big-thinker`/`-feature-specialist` suffixes) is never mistaken for
  the primary feature workflow on the statusline surface.
  `tests/acceptance/unified_plan_specialist_statusline_test.go:TestUPS_StatuslineExcludesPlannerRoleWorkflow`
  `tests/acceptance/unified_plan_specialist_statusline_test.go:TestUPS_StatuslineExcludesLegacyRoleWorkflows`
  `cmd/centinela/hook_statusline_view_more_test.go:TestPrimaryWorkflowSkipsPlannerRoleWorkflow`
- A fresh `start` always pins `planContract: planner-v1` and its directive
  names exactly one role, at the reasoning tier, listing only the planner
  evidence pair — never the legacy names.
  `tests/acceptance/unified_plan_specialist_fresh_start_test.go:TestUPS_FreshStartPinsPlannerAndDirectiveNamesOnlyPlanner`
- The complete legacy pair (both files, chained handoffs) still advances an
  unpinned workflow's plan step exactly as before the migration, and a legacy
  workflow can still author a fresh `big-thinker` evidence stub.
  `tests/acceptance/unified_plan_specialist_legacy_test.go:TestUPS_LegacyCompletePairAdvancesPlan`
  `tests/acceptance/unified_plan_specialist_legacy_test.go:TestUPS_EvidenceInitSucceedsOnUnpinnedWorkflow`
- A complete, correctly-shaped planner evidence pair advances the plan step
  for a pinned workflow.
  `tests/acceptance/unified_plan_specialist_complete_test.go:TestUPS_PlannerEvidencePassesPlanComplete`

## Bug found and fixed during this step

- `cmd/centinela/evidence_init.go` checked `requireKnownFeature` BEFORE
  `evidence.EnsureRoleAllowed`, so `evidence init ghost-feature
  feature-specialist` (no workflow state at all) returned the generic
  "unknown feature" message instead of the D7 "role is retired; use planner"
  message the spec requires. Fixed by reordering: `EnsureRoleAllowed` already
  treats a missing/unreadable workflow as refused, so it runs first; the two
  pre-existing colocated tests that used the retired role name `big-thinker`
  as a placeholder for the generic unknown-feature case
  (`TestEvidenceInitRejectsUnknownFeature`,
  `TestEvidenceInitListsActiveOnUnknown`) were updated to use the non-retired
  `planner` role instead, since they were never testing D7 in the first
  place. Full `cmd/centinela` suite re-verified green after the change.

## Residual Risks

- `codex-claude-role-agent-registry`, `managed-agent-retirement-sweep`, and
  `prompt-doc-budget-ratchet` were already deferred at the plan step
  (`docs/plans/unified-plan-specialist.md` §8) and re-confirmed by
  senior-engineer during implementation. Not re-deferred here — see Deferred
  Findings in the qa-senior report for why duplicating them would be noise.
- The statusline protocol has no per-role field (confirmed by reading
  `cmd/centinela/hook_statusline_view.go` and `internal/ui/render_status.go`
  end to end) — "the statusline names planner" is therefore verified as "a
  planner/legacy role sub-workflow is never shown as the primary feature",
  which is the actual guarantee `isRoleWorkflow`'s new `-planner` suffix
  provides. If a future feature adds a per-role statusline field, this
  interpretation should be revisited.
