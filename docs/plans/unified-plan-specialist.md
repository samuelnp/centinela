# Plan — unified-plan-specialist

> Phase 13 "Lighter Centinela". Brief: `docs/features/unified-plan-specialist.md`.
> Precedent this plan deliberately mirrors: `docs/plans/adversarial-validate-verifier.md`
> (merged as PR #83) — same two problems, same solved shape.

## 1. Problem framing

The `plan` step spawns two subagents. `big-thinker` (reasoning tier) frames
problem/scope/dependencies/risks/rollout; `feature-specialist` (balanced tier)
then re-reads the *same* inputs — every `docs/features/*.md`, the brief, the
roadmap — and writes acceptance criteria, UX states and edge cases. Each
produces its own `.md` + `.json` evidence pair.

The two lenses are **sequential elaboration, not independent judgment**. The
spec lens *builds on* the strategy lens; it does not check it. V2 principle
§12.3 says a separate agent is justified only when independence or parallelism
matters — neither does here. The 2048-rust field test recorded this as failure
#7 (fixed choreography regardless of task size) and #8 (evidence duplication).
The operator pays ~2x plan-step context for zero added assurance, and gets two
reports that a human then has to reconcile.

One `planner` role, one context, one evidence pair, both lenses as ordered
sections of one prompt. The after-plan human confirmation stop is unchanged —
that is the assurance mechanism at this step, and it is not what we are cutting.

## 2. Scope

**In**
- `RolePlanner = "planner"` replacing `RoleBigThinker` + `RoleFeatureSpecial`
  in `RequiredRoles("plan")`, at the reasoning tier.
- `PlanContract` pinned at start (`planner-v1`); empty ⇒ legacy two-role set.
- One merged prompt doc `docs/architecture/planner-prompt.md` carrying both
  lenses in order, replacing the two legacy prompt docs and their mirrors.
- Evidence CLI wiring: stub step/handoff, plan-input pre-fill, companion
  headers, `AllRoles`, contract-aware legacy-role stub refusal.
- Config override key `planner` + legacy-key aliasing with a deprecation notice.
- Plan-advisor lens wording; hook directive; statusline; OpenCode agent config.

**Out**
- Any change to code/tests/validate/docs roles. `adversarial-validate-verifier`
  owns validate; this plan touches `RequiredEvidenceRoles` only to add a `plan`
  branch alongside the existing `validate` branch.
- The after-plan confirmation stop (`workflow.step_confirmation_mode`) — no change.
- **The O(N) plan-snapshot input glob stays exactly as-is.** `RequiredPlanInputs`
  still globs every `docs/features/*.md` (~120 today) and `planner` inherits the
  identical rule via `requiresPlanSnapshot`. Halving the *number of roles* halves
  how many times that list is written, which is the only win this feature claims.
  Changing the *content* rule is `token-diet`'s scope; doing it here would
  entangle two features' acceptance criteria. Explicit non-goal.
- Renaming or deleting `RoleBigThinker` / `RoleFeatureSpecial` constants. They
  survive for legacy evidence, config keys, `brownmap` roadmap provenance
  (`Source.Role: "big-thinker"` on ~2 call sites — historical records, left alone).

## 3. Locked design decisions

**D1 — the role slug is `planner`; the constant is `RolePlanner`.** Unlike the
gatekeeper precedent (where the slug was load-bearing across ten files and only
the *stance* changed), here both legacy slugs are being retired as *required*
roles, so no existing artifact path is preserved by keeping either name. Evidence
lands at `.workflow/<feature>-planner.{md,json}`. `RoleBigThinker` and
`RoleFeatureSpecial` remain declared constants — retired from `RequiredRoles`,
still accepted by the evidence CLI and config for legacy workflows.

**D2 — back-compat is state-dated via a pinned contract field, NOT file-presence
"either set".** Add to `Workflow`:

```go
PlanContract string `json:"planContract,omitempty"`
```

`NewWithOrder` pins `PlanContractUnified = "planner-v1"`. Empty ⇒ the workflow
predates the migration ⇒ the plan step requires the **complete legacy pair set**
(`big-thinker` + `feature-specialist`, exactly today's behavior).

This supersedes the brief's Acceptance Criterion 4 phrasing ("accepts EITHER the
new planner pair OR the complete legacy set"), which predates the PR #83
precedent. An either-set rule is a transition hole: a *fresh* feature could dodge
the new format by hand-authoring two legacy-named evidence pairs, and nothing in
the file system distinguishes "written in flight last week" from "forged today".
State-dating closes it structurally. Clocks are never consulted — the pin is
written once at `centinela start` and read from `.workflow/<feature>.json`.

Consequence, stated so the tests assert it: a fresh `planner-v1` workflow that
writes only legacy evidence **fails**; a legacy workflow that writes only
`big-thinker` evidence **still fails** (partial set), and its error names the
legacy pair. AC4's "message naming both accepted sets" is replaced by a message
naming *this workflow's* required set plus a one-line contract annotation
explaining why (see D3) — naming a set the workflow can never satisfy is worse
than naming none.

**D3 — one contract-aware resolver, `workflow.RequiredEvidenceRoles`, for every
consumer.** The single most important structural constraint in this plan. In
PR #83 a divergence between the hook directive (policy-blind) and the `complete`
gate (contract-aware) was a CRITICAL verifier finding. That resolver already
exists and already has three callers that must not be bypassed:

| Caller | File |
|--------|------|
| `complete` gate | `internal/workflow/validate_orchestration.go` → `validateOrchestration` |
| hook directive | `cmd/centinela/hook_orchestration.go:44` |
| claim verification | `cmd/centinela/complete_verify.go:22`, `cmd/centinela/verify.go:40` |

All four already route through `RequiredEvidenceRoles`, so adding a `plan`
branch **inside it** propagates to every surface at once. No new call site may
use `orchestration.RequiredRolesForFeature` or `orchestration.RequiredRoles`
directly for an operator-facing statement. The plan branch mirrors the validate
branch byte-for-byte in shape:

```go
if step == "plan" && !featureUsesUnifiedPlanner(feature) {
    return []orchestration.Role{orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial}
}
```

Error annotation (the D2 "why" line) is applied by `validateOrchestration`
wrapping `ValidateRoles`' error, not by forking the validator.

**D4 — tier is `reasoning`.** `defaultTierForRole[RolePlanner] = TierReasoning`
(AC3). The strategy lens was already reasoning-tier; the merged context now also
carries the spec lens, so downgrading would be a quality regression on the half
that used to be reasoning-tier. Net spend still drops: one reasoning context
replaces one reasoning + one balanced context.

**D5 — the merged prompt keeps both lenses' full content, in strategy→spec
order.** `docs/architecture/planner-prompt.md` is a new file; the two legacy
prompt docs and their scaffold mirrors are **deleted** in the same commit (AC2:
"replaces … the scaffold asset mirror, parity test updated in the same change").
Section order is locked:

1. `## Purpose` — one agent, two lenses, why they share a context.
2. `## How to Invoke` → `agent-invocation.md` (shared block, one copy).
3. `## Prompt Template`
   - Authoring rules (CLI mandate; **one** copy instead of two — the single
     largest source of the merge's line savings).
   - Inputs to read.
   - `Required analysis — Lens 1: strategy` — problem framing, scope
     boundaries, dependencies & assumptions, risks (impact/likelihood),
     rollout sequence. Verbatim content of the big-thinker's five items.
   - `Required analysis — Lens 2: spec` — behavior summary, Gherkin scenarios
     (happy + ≥1 negative), UX states, out-of-scope. Verbatim content of the
     feature-specialist's four items.
   - `Output format` — one report with the union of both header sets, in the
     same order: Problem, Scope, Dependencies & Assumptions, Risks, Rollout,
     Behavior Summary, Gherkin Scenarios, UX States, Out-of-Scope,
     `#### Deferred Findings` (with the `centinela roadmap defer …
     --source <feature>/planner` line — test-coupled), `#### Handoff`
     (Next role: `senior-engineer`).
4. `## Required Artifact` — `.workflow/<feature>-planner.{md,json}` + the
   validator rules (plan-snapshot inputs, `outputs` under `docs/plans/` or
   `specs/`, non-empty `edgeCases`, RFC 3339, `handoffTo: senior-engineer`).
5. A single **legacy pointer** line: "A workflow pinned before `planner-v1`
   delegates two roles; use Lens 1 for `big-thinker` and Lens 2 for
   `feature-specialist`." This is why deleting the legacy prompt docs is safe —
   the merged doc is a strict superset of both.

**D6 — the prompt-doc line budget.** `promote_orchestration_agents_acceptance_test.go`
enforces ≤130 lines per promoted prompt. The two sources are 107 + 108 = 215
lines; deduplicating the shared authoring-rules / how-to-invoke / required-artifact
blocks removes ~45 lines, so the target is **≤130 lines and that is the primary
requirement**. If, and only if, hitting 130 would require dropping content the ACs
name (either lens' analysis items, the deferred-findings block, the CLI mandate),
the code step must NOT dilute the lenses. The sanctioned fallback is to make the
budget express the invariant it actually protects: an explicit per-file budget map
entry `"planner-prompt.md": 175` **plus** a new assertion that the total line
count of all plan-step prompt docs is strictly less than 215 (the pre-merge sum).
That is a ratchet, not a hole. Choosing the fallback requires recording the
reason in the code-step evidence.

**D7 — legacy-role evidence stubs are refused contract-awarely, not
unconditionally.** The brief's edge case says `evidence init <f> big-thinker`
should error pointing at planner. Erroring unconditionally would brick the
in-flight legacy workflows D2 exists to protect. Rule: `evidence init` /
`artifact new` for `big-thinker` or `feature-specialist` succeeds when the
feature's workflow has an empty `PlanContract`, and errors with
`role %q is retired; use "planner"` when the workflow is pinned to `planner-v1`
(or does not exist). Same predicate as D2 — one source of truth.

**D8 — config keys: alias, never brick.** `allowedModelRoles` gains `"planner"`
and **keeps** `"big-thinker"` / `"feature-specialist"` (removing them would make
`config.Load` fail on existing `centinela.toml` files — an unacceptable
regression, and `tests/unit/configurable_subagent_models_config_unit_test.go`
depends on them resolving). Aliasing happens in the config leaf accessors: if no
explicit `planner` entry exists and a legacy key is present, its value is also
published under `planner`; precedence `big-thinker` > `feature-specialist` (the
strategy lens set the tier that planner inherits). Legacy keys stay in the
returned map so existing consumers/tests are untouched.

Deprecation visibility: a new pure `config.LegacyPlanModelRoleKeys(cfg) []string`
is surfaced by `centinela doctor` (a check) and once at `centinela start` — **not**
by `hook orchestration`, which runs on every prompt and would turn a one-time
migration notice into permanent noise.

**D9 — plan-advisor lens tags become `strategy` / `spec`.** `question.Lens`
values change from the role slugs to lens names, and `advisor.go`'s header line
becomes "One planner agent, two lenses: strategy first, then spec." Rendering
(`- [strategy] …`) is unchanged in shape, so only the literals move.

**D10 — OpenCode agent config: add `planner`, drop the two legacy entries from
the emitted set.** Managed sync only *adds* missing keys (`mergeOpenCodeAgents`
skips existing names), so an existing `opencode.json` keeps its legacy agent
definitions untouched — no clobber, per AC6, and no user config is rewritten.
All emitter changes go through `BuildSyncPlan`/`ApplySync`; no `Ensure*` writer
is touched (known perpetual-drift regression class). Codex has **no** per-role
agent registry (`adapter_codex.go` emits only `.codex/config.toml` + `AGENTS.md`,
neither of which enumerates roles), so "the codex equivalent" in the brief is a
no-op here — recorded as a deferred finding rather than silently ignored.

## 4. Slices

Smallest-correct-first. Every slice compiles, `go test ./...` green, independently
revertible. All new/edited files ≤100 lines including `_test.go` (G1 applies to
test files); target ≥97% per-package coverage with colocated tests.

### Slice 1 — the pinned contract (no behavior change)

| File | Change | Lines |
|------|--------|-------|
| `internal/workflow/state.go` | `PlanContract string \`json:"planContract,omitempty"\`` + doc comment mirroring `ValidateContract` | +5 |
| `internal/workflow/contract.go` | `PlanContractUnified = "planner-v1"`, `(*Workflow).UsesUnifiedPlanner()`, `featureUsesUnifiedPlanner(feature)` | +18 (file 24→42) |
| `internal/workflow/order.go` | `PlanContract: PlanContractUnified` in `NewWithOrder` | +1 |

Nothing reads the field yet, so this slice is pure state addition — a workflow
started here behaves identically.

**Tests:** `internal/workflow/contract_plan_test.go` (new, ~45 lines): pinned ⇒
true; empty ⇒ false; nil workflow ⇒ false; unreadable/missing state ⇒ false
(fail-legacy, never invent a gate). Extend `order_test.go` by one assertion that
`NewWithOrder` pins the field.

### Slice 2 — role, policy, tier, evidence rules, contract-aware resolution

This slice flips behavior; the resolver branch must land in the *same* commit as
the policy change or every in-flight workflow breaks.

| File | Change | Lines |
|------|--------|-------|
| `internal/orchestration/policy.go` | `RolePlanner Role = "planner"` with a comment marking the two legacy constants retired-but-retained; `RequiredRoles("plan")` → `[]Role{RolePlanner}` | +8 |
| `internal/orchestration/models.go` | `RolePlanner: TierReasoning`; `AllowedRoleSlugs()` gains `"planner"` (legacy slugs retained) | +3 |
| `internal/orchestration/output_rules.go` | `case RoleBigThinker, RoleFeatureSpecial, RolePlanner:` (plan/spec artifact rule) | ±1 |
| `internal/orchestration/plan_snapshot.go` | `requiresPlanSnapshot` gains `RolePlanner`; comment updated — **glob rule unchanged** | ±3 |
| `internal/orchestration/evidence.go` | `needsEdgeCases` gains `RolePlanner` (it now carries the spec lens) | ±1 |
| `internal/workflow/validate_orchestration.go` | plan branch per D3; wrap the error with the contract annotation per D2 | +12 (file 39→51) |

**Tests:** `internal/orchestration/policy_test.go` — plan ⇒ `[planner]`, no legacy
slug. `models_test.go` — planner ⇒ reasoning; legacy tiers unchanged (the existing
`tests/unit/configurable_subagent_models_unit_test.go` assertions on
`RoleBigThinker → reasoning` / `RoleFeatureSpecial → balanced` stay true and need
no edit). New `internal/workflow/validate_orchestration_plan_test.go` (~70 lines,
may need splitting into two files to stay ≤100): pinned workflow demands the
planner pair and rejects a hand-authored legacy pair; legacy workflow accepts the
complete legacy pair and rejects a partial one; legacy workflow is not retro-gated
on planner evidence.

**Known test-fixture fallout to fix in this slice:**
`tests/unit/enforce_actionable_orchestration_evidence_unit_test.go` writes legacy
plan evidence and calls the plan validator — update it to planner evidence (or
route through `ValidateRoles` with an explicit legacy set) or it goes red.

### Slice 3 — evidence CLI wiring

| File | Change | Lines |
|------|--------|-------|
| `internal/evidence/roles.go` | `AllRoles()` gains `RolePlanner` (first in the plan position); legacy slugs retained | +1 |
| `internal/evidence/schema_init.go` | `stepForRole`: planner ⇒ `"plan"`; `handoffForRole`: planner ⇒ `senior-engineer` | +4 |
| `internal/evidence/plan_inputs.go` | `PlanInputs` gains `RolePlanner` — same `RequiredPlanInputs` call, so a pre-filled `init` validates by construction | ±1 |
| `internal/evidence/companion_skeletons.go` | planner headers = the ordered union: `Problem`, `Scope`, `Dependencies & Assumptions`, `Risks`, `Rollout`, `Behavior Summary`, `Acceptance Criteria (Gherkin)`, `UX States`, `Edge Cases`, `Out-of-Scope`, `Handoff` | +1 |
| `internal/evidence/roles_retired.go` (new, ~35) | `EnsureRoleAllowed(feature, role) error` implementing D7 — legacy plan roles allowed only for an unpinned workflow | +35 |
| `cmd/centinela/evidence_init.go` (+ `artifact new` entry point) | call `EnsureRoleAllowed` before writing | +4 |

`internal/evidence/invalidation_targets.go` needs **no** change: plan roles arrive
via `RequiredRolesForFeature`. Verify this in a test rather than by inspection —
if it hard-codes a plan role, invalidation would miss the planner pair.

**Import-direction check:** `internal/evidence` must not import
`internal/workflow` if that would create a cycle (`workflow` → `orchestration`,
`evidence` → `orchestration`). Confirm the direction before implementing; if a
cycle appears, `EnsureRoleAllowed` moves to `cmd/centinela` as a thin guard with
the predicate exported from `internal/workflow`.

**Tests:** colocated `roles_retired_test.go` (~60), `schema_init_planner_test.go`
(~40), plus a `cmd/centinela` test that `evidence init <legacy-workflow> big-thinker`
succeeds and `evidence init <pinned> big-thinker` errors naming `planner`.

### Slice 4 — config keys, aliasing, deprecation notice

| File | Change | Lines |
|------|--------|-------|
| `internal/config/orchestration_models.go` | `allowedModelRoles["planner"] = true` (legacy keys retained) | +1 |
| `internal/config/orchestration.go` | alias legacy → `planner` in `OrchestrationModelTiers` / `OrchestrationModelOverrides` per D8; extract a shared `aliasPlanRole` helper to keep both accessors small | +18 |
| `internal/config/orchestration_deprecations.go` (new, ~25) | `LegacyPlanModelRoleKeys(cfg) []string`, sorted, nil-safe | +25 |
| `cmd/centinela/doctor_*.go` | one check row: deprecated `[orchestration.models]` keys → suggest `planner` | +12 |
| `cmd/centinela/start.go` | print the same notice once at start | +5 |
| `tests/unit/configurable_model_routing_parity_unit_test.go` | the config/orchestration allow-list parity test must see `planner` on both sides | ±2 |

**Tests:** colocated `orchestration_deprecations_test.go`; alias precedence test
(explicit `planner` wins; `big-thinker` beats `feature-specialist`; override-table
form aliases too); a test that a `centinela.toml` with only legacy keys still
loads without error.

### Slice 5 — prompts, contract docs, scaffold mirror

- **New** `docs/architecture/planner-prompt.md` per D5/D6, with the
  `<!-- centinela:doc-version=1 template=docs/architecture/planner-prompt.md -->`
  header, `## Purpose` / `## Prompt Template` / `## Required Artifact` sections
  (asserted by `promote_orchestration_agents_acceptance_test.go`), the
  `agent-invocation.md` reference, the CLI-mandate authoring block, and the
  `#### Deferred Findings` block with the `--source <feature>/planner` line.
- **Delete** `docs/architecture/big-thinker-prompt.md`,
  `docs/architecture/feature-specialist-prompt.md`, and **both** scaffold mirrors.
- **Add** `internal/scaffold/assets/docs/architecture/planner-prompt.md`,
  byte-identical, same commit.
- `docs/architecture/evidence-contract.md` — replace the two `### big-thinker` /
  `### feature-specialist` sections with `### planner (step: plan)`; keep a short
  "legacy (pre `planner-v1`) workflows" note listing the retired pair; update the
  `"role"` enumeration line and the worked example to `planner`.
- `CLAUDE.md` — the two "Plan agent —" quick-reference rows collapse into
  `| Plan agent — planner | [planner-prompt.md](docs/architecture/planner-prompt.md) |`.
- `README.md` — the mermaid orchestration diagram node `BT → FS` becomes a single
  `PL["planner<br/>reasoning · opus-4-7"]`.
- `HOWTO.md`, `docs/guides/workflow-and-hooks.md`, `docs/guides/configuration.md`,
  `docs/guides/configuration-reference.md` — evidence-path and role-key examples.
- `docs/architecture/workflow-enforcement.md` + `artifact-templates.md` — plan-row
  role names. (`workflow-enforcement.md` is on `mirrorParityAllowlist`, i.e. **not
  mirrored** — do not create a mirror for it.)
- **Not touched:** `docs/project-docs/kb/*.md` — historical per-feature knowledge
  pages describing what shipped at the time; rewriting them would falsify history.

**Test-list edits required in the same commit** (each currently names the two
legacy files and will fail on a missing file):

| Test | Edit |
|------|------|
| `tests/acceptance/deferred_findings_prompt_parity_test.go` | swap the two entries for `planner-prompt.md` |
| `tests/acceptance/extract_agent_shared_blocks_acceptance_test.go` | same |
| `tests/acceptance/promote_orchestration_agents_acceptance_test.go` | same (+ D6 budget decision) |
| `tests/acceptance/agent_evidence_contract_acceptance_test.go` | `rolePrompts` map, the role-name list at ~line 40, and the mirror-parity list at ~line 140 |
| `tests/acceptance/prompts_mandate_cli_acceptance_test.go` | `promptRoles`: `planner` replaces the two |
| `tests/acceptance/scaffold_arch_parity_acceptance_test.go` | no edit — it globs `docs/architecture/*.md`, so deleting sources and adding `planner-prompt.md` + mirror is self-maintaining |

**New test:** `tests/acceptance/planner_prompt_test.go` (~70 lines) — asserts both
lens headings present and in order (strategy index < spec index), the union output
headers, `handoffTo: senior-engineer`, the legacy-pointer line, and that neither
legacy prompt file exists any more.

### Slice 6 — surfaces: statusline, plan-advisor, agent config, gitignore

| File | Change | Lines |
|------|--------|-------|
| `cmd/centinela/hook_statusline_view.go` | `isRoleWorkflow` gains `"-planner"`; legacy suffixes retained | +1 |
| `internal/planadvisor/questions.go` | lens literals → `"strategy"` / `"spec"` per D9 | ±13 |
| `internal/planadvisor/advisor.go` | header line → "One planner agent, two lenses: strategy first, then spec." | ±1 |
| `internal/setup/opencode_agent_config.go` | add `"planner"` (description + prompt naming both lenses); remove the `big-thinker` / `feature-specialist` map entries | ±8 |
| `internal/setup/testdata/golden/opencode/opencode.json` | regenerate | ±10 |
| `opencode.json` (repo's own) | regenerate via the managed path | ±10 |
| `.gitignore` | add `.workflow/*-planner.json` beside the other role patterns | +1 |
| `tests/unit/lean_evidence_footprint_unit_test.go`, `tests/integration/lean_evidence_footprint_integration_test.go` | add the planner pattern | ±3 |
| `tests/unit/opencode_config_test.go` | expected agent-name list | ±2 |
| `tests/integration/add_plan_advisor_mode_integration_test.go`, `tests/integration/failure_ledger_plan_advisor_integration_test.go` | lens-tag assertions → `[strategy]` / `[spec]` | ±4 |
| `tests/integration/configurable_subagent_models_integration_test.go` | plan-role annotation expectations | ±6 |

`hook_statusline_rules.go`'s plan branch keys on artifact existence
(`docs/plans/<feature>.md`), not on roles — no change needed, verified by reading it.

### Slice 7 — spec + acceptance wiring (tests step; listed for completeness)

`specs/unified-plan-specialist.feature` drives
`tests/acceptance/unified_plan_specialist_test.go` — binary-driven end-to-end in a
temp repo with a **local bare origin** (never a network remote: a real push hangs
`go test` for hours and times out claim verification):

1. `centinela start` ⇒ state pins `planContract: planner-v1`.
2. `hook orchestration` on the plan step names exactly one role and lists only
   `.workflow/<f>-planner.{md,json}`.
3. `complete` blocks with only legacy evidence; passes with the planner pair.
4. A state file with `planContract` removed (simulating in-flight) ⇒ `complete`
   passes on the complete legacy pair set, fails on a partial one, and the
   directive names the two legacy roles — i.e. directive and gate agree (D3).
5. Guided profile (no subagent evidence) ⇒ directive still names `planner`, no
   evidence requirement printed.

## 5. Test strategy

- **Unit (colocated, ≤100 lines each, drives the 97% target):**
  `internal/workflow/contract_plan_test.go`, `validate_orchestration_plan_test.go`,
  `internal/orchestration/{policy,models}_test.go` additions,
  `internal/evidence/{roles_retired,schema_init_planner}_test.go`,
  `internal/config/orchestration_deprecations_test.go`,
  `internal/planadvisor/questions_test.go`, `internal/setup/opencode_agent_config_test.go`.
  Coverage is per-package with no `-coverpkg`, so `tests/` tier files do not move
  the gate — every package touched needs its own colocated additions.
- **Integration:** contract-aware directive vs gate agreement (the PR #83 CRITICAL
  class) exercised through `RequiredEvidenceRoles` from both a pinned and an
  unpinned workflow; config alias end-to-end from TOML to resolved model name.
- **Acceptance:** `planner_prompt_test.go`, `unified_plan_specialist_test.go`,
  plus the five existing prompt-list tests re-pointed.
- **Ratchet discipline:** aim ≥97% on touched packages (gate is 95%) so a parallel
  merge cannot tip main red.

## 6. Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Directive/gate divergence at the plan step (the PR #83 CRITICAL) | High | Medium | Single branch inside `RequiredEvidenceRoles`; no new direct `RequiredRoles` caller; acceptance step 4 asserts both surfaces agree for a legacy workflow |
| Transition hole — fresh feature passes on forged legacy files | High | Low (with D2) | State-dated `PlanContract`; explicit test that a pinned workflow rejects a hand-authored legacy pair |
| In-flight legacy workflow bricked (evidence init refused, gate flipped) | High | Low | D7 contract-aware stub refusal; legacy branch keeps the exact pre-change behavior; acceptance step 4 |
| Prompt-quality regression — merged prompt dilutes the strategy lens | Medium | Medium | D5 keeps both lenses verbatim; D6 forbids trimming lens content to hit a line budget; after-plan human confirmation unchanged |
| Scaffold mirror drift (partial-parity trap) | Medium | Medium | Delete/add sources and mirrors in one commit; the glob-based parity test catches an unmirrored `planner-prompt.md` |
| Managed-sync clobber / perpetual pending drift on agent config | Medium | Low | `BuildSyncPlan`/`ApplySync` only; `mergeOpenCodeAgents` never overwrites existing keys; golden testdata regenerated |
| `centinela.toml` with legacy model keys fails to load | High | Low | D8 keeps legacy keys allowed; explicit test |
| Textual conflict with in-flight `merge-truthful-delivery` (PR #84) | Low | Medium | Overlap is `CLAUDE.md` (different table rows) and `internal/setup/opencode_agent_config.go` (different map entries). Rebase before validate; re-run the full suite on the merged tree, and re-append any new `docs/features/*.md` to this feature's plan-evidence inputs after the rebase |
| Import cycle from `internal/evidence` → `internal/workflow` | Medium | Medium | Slice 3 checks direction first; fallback is a `cmd/centinela` guard over an exported predicate |
| Coverage dip below the 95% gate on `internal/config` / `internal/evidence` | Medium | Low | Colocated tests in the same slice as the code; target 97% |

## 7. Rollout

1. Slice 1 (state pin, inert) → 2 (policy + resolver, the behavior flip) →
   3 (evidence CLI) → 4 (config) → 5 (prompts + mirrors) → 6 (surfaces) →
   7 (spec/acceptance, tests step).
2. Slices 1–2 are the only ones that can break an existing workflow; land them
   together and run the full suite before continuing.
3. Rebase onto main before the validate step (PR #84 may land first); re-append
   any new `docs/features/*.md` to the plan-step evidence `inputs` and re-run
   `centinela evidence validate` — a rebase invalidates plan-evidence snapshots.
4. No migration command and no data migration: existing `.workflow/<feature>.json`
   files simply lack `planContract` and are therefore legacy by construction.
5. Reverting is a straight `git revert` of slices 2–6; slice 1's inert field can
   stay.

## 8. Deferred (recorded on the roadmap)

Genuinely new discoveries, out of scope here, captured via
`centinela roadmap defer --source unified-plan-specialist/big-thinker`:

- `codex-claude-role-agent-registry` — only OpenCode has a per-role agent
  registry; Codex and the Claude harness emit no role definitions, so AC6's
  "codex equivalent" has nothing to write to. Role definitions should be a
  harness-adapter capability rather than an OpenCode-only file.
- `managed-agent-retirement-sweep` — `mergeOpenCodeAgents` only adds; a retired
  agent lingers in every existing `opencode.json` forever. Managed sync needs a
  marked-managed removal path that still never touches user-authored entries.
- `prompt-doc-budget-ratchet` — the per-file 130-line prompt budget does not
  express the invariant it protects (total prompt surface) and penalizes
  legitimate consolidation; convert it to an aggregate ratchet.

Not deferred because already owned elsewhere: the O(N) `RequiredPlanInputs` glob
(`token-diet`), validate-step role changes (`adversarial-validate-verifier`).
