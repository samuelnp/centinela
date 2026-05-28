### Gatekeeper Report: evidence-cli
**Date:** 2026-05-28
**Status:** WARNING

#### Analyzed Specs (all `specs/*.feature`)

Every `.feature` file was scanned. Specs with direct evidence overlap:
- `specs/evidence-cli.feature` — primary spec for this feature
- `specs/add-agent-evidence-contract.feature` — predecessor; partially superseded (see Finding 1)
- `specs/enforce-actionable-orchestration-evidence.feature` — validator rules; honored
- `specs/enforce-step-subagent-orchestration.feature` — step-gate rules; honored
- `specs/enforce-plan-snapshot-inputs.feature` — snapshot rules; honored
- `specs/refine-ux-specialist-evidence.feature` — UX evidence rules; honored
- `specs/parallel-feature-worktrees.feature` — postwrite scoping; honored
- `specs/merge-steward-auto-dispatch.feature` — merge-steward role present in AllRoles(); honored

All other specs (CI, docs, landing page, roadmap, session-rehydration, etc.) have no interaction with the evidence-cli surface.

#### Findings

**Finding 1 (WARNING) — `add-agent-evidence-contract.feature` Scenario 2 is now stale**

The spec says: _"each prompt should embed a role-specific JSON skeleton with realistic
placeholders"_. The `evidence-cli` feature deliberately removed all embedded JSON skeletons
from every `docs/architecture/*-prompt.md` and replaced them with
`centinela evidence schema <role>` as the single source of truth. The acceptance test
`TestPromptsMandateEvidenceCLI` (via `assertNoEmbeddedSkeleton`) now FAILS if a skeleton
is present.

The spec text and the implemented behaviour are in direct contradiction. No test currently
enforces the spec's Scenario 2 wording; the enforcing test
(`agent_evidence_contract_acceptance_test.go`) documents the inversion inline with
`// Slice 3 (evidence-cli): the embedded JSON skeleton was removed`.

**Risk:** Low. The accepted behaviour is clearly the intent; the spec was not updated to
reflect the decision. A future agent reading the spec without the acceptance tests could
re-introduce the skeleton and then break `assertNoEmbeddedSkeleton`.
**Recommendation:** Update `specs/add-agent-evidence-contract.feature` Scenario 2 to say
_"each prompt must reference `centinela evidence schema <role>` (no embedded skeleton)"_.

**Finding 2 (WARNING) — `mobileFirst: true` set in non-ux-ui-specialist evidence**

Three shipped evidence files for this feature include `"mobileFirst": true`:
- `.workflow/evidence-cli-big-thinker.json`
- `.workflow/evidence-cli-feature-specialist.json`
- `.workflow/evidence-cli-senior-engineer.json`

`evidence-contract.md` Rule 6 states: _"`mobileFirst` is omitted unless the role is
`ux-ui-specialist`"_. The validator in `internal/orchestration/evidence.go` does not
reject this (it only requires it for `ux-ui-specialist`), so `centinela evidence validate`
passes. The contract is violated at the authoring level, not the enforcement level.

**Risk:** Low. Validator passes; surplus field is harmless. However the authoring agents
mis-applied the rule, suggesting the schema init stub or prompt may be nudging all roles
to set `mobileFirst`.
**Recommendation:** Audit `centinela evidence schema` output for big-thinker,
feature-specialist, and senior-engineer to confirm `mobileFirst` is absent from generated
stubs. If stubs emit it, fix `internal/evidence/schema_init.go`.

**Finding 3 (INFORMATIONAL) — `jsonKeyOrder` duplicated in `hookpolicy`**

`internal/hookpolicy/format_evidence_order.go` duplicates `jsonKeyOrder` from
`internal/evidence/schema.go`. The comment acknowledges the duplication and states drift
is caught by `format_evidence_test.go`. Confirmed: the test cross-checks both packages
produce identical output. Acceptable, but worth tracking.

#### Cross-feature Conflict Check

| Check | Result |
|-------|--------|
| `orchestration.ValidateEvidence` API unchanged (no rule duplication) | PASS |
| `internal/orchestration` not modified; only delegated to from new package | PASS |
| Scaffold mirror parity allowlist honest (4 pre-existing entries) | PASS |
| New roles (`gatekeeper`, `production-readiness`) excluded from step gating | PASS |
| PostToolUse hook scoped to active feature via `isActiveFeatureEvidence` | PASS |
| All 9 updated prompts mirror-synced to `internal/scaffold/assets/` | PASS |
| `centinela evidence validate evidence-cli` exits 0 | PASS |

#### Recommendation

**WARNING** — safe to proceed. Two issues should be resolved as follow-up:

1. Amend `specs/add-agent-evidence-contract.feature` Scenario 2 to align with the new
   prompt-mandate: no embedded skeleton, reference `centinela evidence schema <role>`.
2. Audit `centinela evidence schema` stub output for non-UX roles; remove `mobileFirst`
   if it is being emitted; correct the three existing evidence files in a follow-up fix.

Neither issue blocks shipping this feature. No cross-feature regressions detected.
