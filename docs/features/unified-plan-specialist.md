# Feature Brief — unified-plan-specialist

> Phase 13 "Lighter Centinela", from the centinela-v2 port-back analysis
> (2026-07-28). V2 principle §12.3: a separate agent is justified only when
> independence or parallelism matters. The two plan lenses are sequential
> elaboration, not independent judgments — they belong in ONE context.

## Problem — what pain, who
Every plan step spawns TWO subagents — big-thinker (reasoning tier) and
feature-specialist (balanced tier) — each a full context that re-reads the
same brief/roadmap inputs, each producing its own `.md`+`.json` evidence pair.
The 2048-rust field test flagged this as failure #7 (fixed choreography
regardless of task size) and #8 (evidence duplication): double subagent spend
and double evidence for lenses that are not independent — the spec lens
*builds on* the strategy lens and benefits from sharing its context. The
orchestrating user pays ~2x plan-step tokens for zero added assurance.

## Scope (this feature ONLY)
- **In:** one `planner` role replacing big-thinker + feature-specialist for
  the plan step; a merged prompt with both lenses as ordered sections; one
  evidence pair `.workflow/<feature>-planner.{md,json}`; legacy-evidence
  acceptance for in-flight workflows; plan-advisor and hook-directive wording.
- **Out:** any change to code/tests/validate/docs roles (adversarial-validate-
  verifier is a separate feature); the after-plan confirmation stop (kept
  as-is); plan-snapshot input *content* rules (token-diet owns the O(N) glob).

## User Stories
- As an orchestrating agent, when the plan step starts I delegate to ONE
  planner subagent and get one plan whose strategy and spec sections agree,
  instead of coordinating two agents and reconciling their outputs.
- As an operator, my plan step costs roughly half the subagent tokens and
  produces half the evidence files, with the same after-plan review stop.
- As a maintainer, an in-flight feature started before this change still
  completes its plan step with the old two-role evidence.

## Acceptance Criteria (→ Gherkin)
1. `RequiredRoles("plan")` returns `[planner]`; the plan-step hook directive
   names one role and lists `.workflow/<feature>-planner.{md,json}` as the
   only required plan evidence.
2. The planner prompt doc (`docs/architecture/planner-prompt.md`) contains
   both lenses as ordered sections — strategy/scope/risks first, then
   acceptance criteria/edge cases/spec — and replaces big-thinker-prompt.md +
   feature-specialist-prompt.md in the CLAUDE.md quick-reference and the
   scaffold asset mirror (parity test updated in the same change).
3. The planner resolves to the reasoning tier in `internal/orchestration/
   models.go` and is remappable via the existing per-role model override
   config (`internal/config/orchestration_models.go`).
4. Evidence validation accepts EITHER the new planner pair OR the complete
   legacy set (all four big-thinker + feature-specialist files); a partial
   legacy set (e.g. big-thinker only) still fails with a message naming both
   accepted sets.
5. Plan-advisor output (`internal/planadvisor`) phrases its questions as two
   lenses for one agent ("strategy lens… spec lens…"), no longer instructing
   delegation to two agents.
6. OpenCode/Codex agent-config generation (`internal/setup/
   opencode_agent_config.go` and codex equivalent) emits the planner agent and
   stops emitting the two legacy plan agents; managed-sync (`BuildSyncPlan`/
   `ApplySync`) migrates existing projects without clobbering user config.
7. `centinela validate`, the full test suite, and the scaffold-parity test
   pass; statusline/status views show `planner` for plan-step delegation.

## Edge Cases
- In-flight workflow with legacy evidence already written → plan `complete`
  passes on the legacy set (criterion 4); a NEW feature must not pass with
  legacy-named files authored fresh (accept legacy only when the workflow
  predates the migration — decided by workflow state, or documented as
  "either set" if state-dating is not tracked; plan step must pick one and
  test it).
- Per-role model override config keyed `big-thinker`/`feature-specialist` →
  accepted with a deprecation warning and mapped to `planner`; unknown-role
  error must not brick existing centinela.toml files.
- `evidence init`/`artifact new` stubs for the plan step generate planner
  stubs; requesting a legacy role stub errors with a pointer to planner.
- Brownfield analysis surfaces referencing big-thinker (`internal/brownmap`)
  keep working — references updated, no orphaned role strings.
- Guided/outcome profiles (no subagent evidence required) → directive still
  names the single planner; no evidence requirement appears.
- Scaffold mirror drift: docs/architecture prompt files and
  `internal/scaffold/assets` mirror must change in the same commit (known
  partial-parity trap).

## Data Model
No persisted schema change. `Role` constant: `RolePlanner = "planner"`
replaces `RoleBigThinker`/`RoleFeatureSpecial` in `RequiredRoles`; legacy
constants retained only where evidence back-compat needs them. Evidence JSON
`role` field value: `planner`.

## Integration Points
- `internal/orchestration/{policy,models,output_rules,plan_snapshot}.go`
- `internal/planadvisor/{advisor,questions}.go`
- `internal/config/orchestration_models.go` (override keys + deprecation)
- `internal/setup` agent-config emitters + managed sync
- `cmd/centinela/hook_statusline_view.go`, hook directives
- docs: CLAUDE.md table, planner-prompt.md, scaffold assets mirror
- Consumers: evidence validate, `centinela verify` claim checks, delivery
  composer inputs listing plan artifacts.

## Risks
- **Transition hole** (High): a lax "either set" rule could let a fresh
  feature pass with one hand-authored legacy file. Mitigation: complete-set
  semantics per criterion 4 with explicit tests for partial sets.
- **Managed-sync clobber** (Med): agent-config regeneration must go through
  BuildSyncPlan/ApplySync, not legacy writers (known drift bug class).
- **Prompt-quality regression** (Med): merged prompt could dilute the
  strategy lens; mitigate by keeping both sections' full content and the
  after-plan human review unchanged.
- **Coverage** (Low): colocated `_test.go` (≤100 lines each), ≥97% on touched
  packages.

## Decomposition
Single feature; no split. Independent of adversarial-validate-verifier.
