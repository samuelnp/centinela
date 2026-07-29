### Big-Thinker Report: unified-plan-specialist
**Date:** 2026-07-29

## Problem

Every `plan` step spawns two subagents. `big-thinker` (reasoning tier) frames
problem, scope, dependencies, risks and rollout; `feature-specialist` (balanced
tier) then re-reads the same inputs — the brief, the roadmap, every
`docs/features/*.md` — and writes acceptance criteria, UX states and edge cases.
Each writes its own `.md` + `.json` evidence pair. The two lenses are sequential
elaboration, not independent judgment: the spec lens *builds on* the strategy
lens, it never checks it. V2 principle §12.3 says a separate agent is justified
only when independence or parallelism matters; neither applies here. The
2048-rust field test recorded this as failure #7 (fixed choreography regardless
of task size) and #8 (evidence duplication). The orchestrating operator pays ~2x
plan-step context and then reconciles two reports, for zero added assurance.

## Scope

- **In:** `RolePlanner = "planner"` at the reasoning tier replacing both plan
  roles in `RequiredRoles("plan")`; a `PlanContract` field pinned at start
  (`planner-v1`) that state-dates legacy back-compat; one merged prompt doc
  `docs/architecture/planner-prompt.md` carrying both lenses as ordered sections
  and replacing the two legacy prompt docs plus their scaffold mirrors; evidence
  CLI wiring (stub step/handoff, plan-input pre-fill, companion headers,
  contract-aware refusal of retired-role stubs); config override key `planner`
  with legacy-key aliasing and a deprecation notice; plan-advisor lens wording;
  hook directive, statusline, OpenCode agent config, `.gitignore` role pattern.
- **Out:** any change to the code/tests/validate/docs roles (validate belongs to
  the just-merged `adversarial-validate-verifier`); the after-plan human
  confirmation stop, kept exactly as-is because it is the assurance mechanism at
  this step and is not what we are cutting; **the O(N) plan-snapshot input
  glob**, which stays byte-for-byte as it is — `planner` inherits
  `requiresPlanSnapshot` and the identical `RequiredPlanInputs` rule. Halving
  the number of *roles* halves how many times that ~120-entry list is written,
  which is the only win this feature claims; changing the input *content* rule
  is `token-diet`'s scope and entangling the two would couple their acceptance
  criteria.

## Dependencies & Assumptions

- **Precedent, not just prior art:** `adversarial-validate-verifier` (PR #83)
  solved the same two problems one step later in the workflow. This plan reuses
  its two structural answers in shape: (a) role replacement with back-compat via
  a **state-dated contract field pinned at start** (`internal/workflow/contract.go`,
  its D4) rather than a file-presence "either set"; (b) a **single
  contract-aware resolver**, `workflow.RequiredEvidenceRoles`, feeding the hook
  directive, the `complete` gate and claim verification alike. A divergence
  between directive and gate was a CRITICAL verifier finding last time; the plan
  step must not reintroduce it.
- The brief's Acceptance Criterion 4 ("accepts EITHER the planner pair OR the
  complete legacy set") predates that precedent and is **superseded** by the
  pinned-contract rule, which closes a hole an either-set rule leaves open: a
  fresh feature could otherwise dodge the new format by hand-authoring two
  legacy-named pairs, and the filesystem cannot distinguish that from real
  in-flight evidence. The plan records the deviation explicitly (D2).
- `RequiredEvidenceRoles` already has all four consumers wired
  (`internal/workflow/validate_orchestration.go`, `cmd/centinela/hook_orchestration.go:44`,
  `cmd/centinela/complete_verify.go:22`, `cmd/centinela/verify.go:40`), so one
  branch propagates to every surface at once.
- Codex has **no** per-role agent registry — `adapter_codex.go` emits only
  `.codex/config.toml` and `AGENTS.md`, neither of which enumerates roles. The
  brief's "codex equivalent" of the OpenCode emitter does not exist; recorded as
  a deferred finding rather than silently skipped.
- The legacy prompt docs are named by five acceptance test lists and are
  parity-mirrored; deleting them is a same-commit multi-file edit.
- Assumes `internal/evidence` can read the workflow pin without an import cycle;
  the plan carries a fallback (a `cmd/centinela` guard over an exported
  predicate) if it cannot.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Hook directive and `complete` gate name different plan roles (the PR #83 CRITICAL class) | High | Medium | One branch inside `RequiredEvidenceRoles`; no new direct `RequiredRoles`/`RequiredRolesForFeature` caller for operator-facing text; acceptance scenario asserts both surfaces agree for a legacy workflow |
| Transition hole: a fresh feature passes on forged legacy files | High | Low | State-dated `PlanContract` pinned at start; test that a pinned workflow rejects a hand-authored legacy pair |
| In-flight legacy workflow bricked — evidence init refused or gate flipped mid-feature | High | Low | Legacy branch preserves today's behavior exactly; retired-role stubs refused *contract-awarely*, so an unpinned workflow can still author `big-thinker` evidence |
| Merged prompt dilutes the strategy lens | Medium | Medium | Both lenses kept verbatim as ordered sections; line-budget pressure must not be relieved by trimming lens content; after-plan human review unchanged |
| Scaffold mirror partial-parity trap | Medium | Medium | Source deletions, the new prompt and its mirror land in one commit; the glob-based parity test catches an unmirrored addition |
| `centinela.toml` with legacy `[orchestration.models]` keys stops loading | High | Low | Legacy keys stay in the allow-list and are aliased to `planner`, never rejected |
| Managed-sync clobber / perpetual pending drift on agent config | Medium | Low | `BuildSyncPlan`/`ApplySync` only; the merge helper never overwrites existing keys; golden testdata regenerated |
| Textual conflict with in-flight `merge-truthful-delivery` (PR #84) | Low | Medium | Overlap is `CLAUDE.md` (different rows) and `opencode_agent_config.go` (different map entries); rebase before validate and re-append new `docs/features/*.md` to the plan evidence inputs afterwards |
| Coverage dip below the 95% gate on touched packages | Medium | Low | Colocated `_test.go` (≤100 lines) in the same slice as each change; target ≥97% |

## Rollout

- Step 1: pin `PlanContract` in `state.go` / `order.go` / `contract.go` — inert,
  nothing reads it yet.
- Step 2: flip policy (`RequiredRoles("plan") → [planner]`), tier, output /
  snapshot / edge-case rules **and** the contract-aware plan branch in the same
  commit; steps 1–2 are the only ones that can break an existing workflow.
- Step 3: evidence CLI wiring (stub step/handoff, plan-input pre-fill, companion
  headers, retired-role refusal). Step 4: config key, aliasing, deprecation notice.
- Step 5: the merged prompt doc, deletion of both legacy docs and their mirrors,
  the five test-list edits, `CLAUDE.md` / `README.md` / guides.
- Step 6: statusline, plan-advisor lens wording, OpenCode agent config, gitignore.
- Step 7 (tests step): Gherkin spec + a binary-driven acceptance test using a
  **local bare origin**, never a network remote.
- No data migration: existing workflow state simply lacks `planContract` and is
  legacy by construction. Revert is a straight revert of steps 2–6.

## Deferred Findings

Recorded with `centinela roadmap defer … --source unified-plan-specialist/big-thinker`:

- `codex-claude-role-agent-registry` — role definitions exist only for OpenCode;
  make them a harness-adapter capability so Codex and the Claude harness get them.
- `managed-agent-retirement-sweep` — managed sync only adds agent keys, so a
  retired managed agent lingers in every existing `opencode.json` forever.
- `prompt-doc-budget-ratchet` — the per-file 130-line prompt budget does not
  express the invariant it protects (total prompt surface) and penalizes
  legitimate consolidation.

Not deferred, already owned elsewhere: the O(N) `RequiredPlanInputs` glob
(`token-diet`) and validate-step role changes (`adversarial-validate-verifier`).

## Handoff

- Next role: feature-specialist
- Outstanding questions:
  1. D6 — can `planner-prompt.md` hold both lenses' full content within the
     130-line promoted-prompt budget after de-duplicating the shared blocks? The
     plan forbids trimming lens content to fit and specifies a documented
     ratcheted fallback; the spec should state which outcome is acceptable.
  2. D7 — confirm the acceptance criteria for retired-role stubs: succeed for an
     unpinned workflow, error naming `planner` for a pinned one and for a feature
     with no workflow state at all.
  3. Whether the Gherkin spec should pin the "guided/outcome profile still names
     one planner, no evidence requirement printed" scenario as a hard criterion.
