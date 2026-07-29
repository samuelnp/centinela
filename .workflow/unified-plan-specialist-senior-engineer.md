# unified-plan-specialist — senior-engineer

### Senior-Engineer Report: unified-plan-specialist
**Date:** 2026-07-29

## Files Touched

Slices S1–S6 of `docs/plans/unified-plan-specialist.md`. S7 (spec/acceptance
wiring) is the tests step and was deliberately not written.

| Path | Reason |
|------|--------|
| `internal/workflow/state.go` (100) | S1 — `PlanContract` field, state-dated back-compat |
| `internal/workflow/contract.go` (50) | S1 — `PlanContractUnified`, `UsesUnifiedPlanner`, `FeatureUsesUnifiedPlanner` |
| `internal/workflow/order.go` (59) | S1 — `NewWithOrder` pins the contract |
| `internal/orchestration/policy.go` (64) | S2 — `RolePlanner`; `RequiredRoles("plan")` → `[planner]` |
| `internal/orchestration/models.go` (89) | S2 — planner on `TierReasoning`; `AllowedRoleSlugs` |
| `internal/orchestration/output_rules.go` (48) | S2 — plan/spec artifact rule covers planner |
| `internal/orchestration/plan_snapshot.go` (73) | S2 — `requiresPlanSnapshot` gains planner; glob rule unchanged |
| `internal/orchestration/evidence.go` (63) | S2 — `needsEdgeCases` gains planner (it carries the spec lens) |
| `internal/workflow/validate_orchestration.go` (57) | S2/D3 — plan branch INSIDE `RequiredEvidenceRoles` + contract annotation |
| `internal/evidence/roles.go` (60) | S3 — `AllRoles` gains planner |
| `internal/evidence/schema_init.go` (77) | S3 — planner step `plan`, handoff `senior-engineer` |
| `internal/evidence/plan_inputs.go` (16) | S3 — planner inherits the snapshot pre-fill |
| `internal/evidence/companion_skeletons.go` (41) | S3 — planner headers = ordered union of both lenses |
| `internal/evidence/roles_retired.go` (35, new) | S3/D7 — contract-aware retired-role guard |
| `internal/evidence/invalidation_targets.go` (34) | S3 — plan rewind sheds the retired pair (see Divergences) |
| `cmd/centinela/evidence_init.go` (74) | S3 — calls `EnsureRoleAllowed` before writing |
| `internal/config/orchestration_models.go` (75) | S4/D8 — `planner` allowed, legacy keys retained |
| `internal/config/orchestration.go` (88) | S4/D8 — both accessors alias legacy → planner |
| `internal/config/orchestration_plan_alias.go` (63, new) | S4 — `aliasPlanRole`, `LegacyPlanModelRoleKeys/Notice` |
| `internal/doctor/check_config.go` (69) | S4 — deprecation surfaced as a config advisory |
| `cmd/centinela/start.go` (92) + `start_notices.go` (23, new) | S4 — notice once at start; split to stay ≤100 |
| `docs/architecture/planner-prompt.md` (129, new) + scaffold mirror | S5/D5/D6 — both lenses, ordered, within budget |
| `docs/architecture/{big-thinker,feature-specialist}-prompt.md` + both mirrors | S5 — **deleted** in this change |
| `docs/architecture/evidence-contract.md` (+ mirror) | S5 — `### planner` section, legacy note, worked example |
| `docs/architecture/{senior-engineer-prompt,artifact-templates,workflow-enforcement}.md` | S5 — role references re-pointed (first two mirrored) |
| `CLAUDE.md`, `README.md`, `HOWTO.md`, `docs/guides/workflow-and-hooks.md` | S5 — quick-ref row, mermaid node, evidence paths |
| `cmd/centinela/hook_statusline_view.go` (79) | S6 — `isRoleWorkflow` gains `-planner` |
| `internal/planadvisor/questions.go` (70), `advisor.go` (38) | S6/D9 — lens tags `strategy`/`spec`; one-agent header |
| `internal/setup/opencode_agent_config.go` (90) | S6/D10 — emits `planner`, drops the two legacy entries |
| `internal/setup/testdata/golden/opencode/opencode.json`, `opencode.json` | S6 — regenerated via `BuildSyncPlan`/`ApplySync` only |
| `.gitignore` (48) | S6 — `.workflow/*-planner.json` beside the other role patterns |

Colocated unit tests added (all ≤100 lines): `internal/workflow/{contract_plan,
validate_orchestration_plan, validate_orchestration_plan_gate,
plan_evidence_helper}_test.go`, `internal/orchestration/policy_planner_test.go`,
`internal/evidence/{roles_retired, schema_init_planner}_test.go`,
`internal/config/{orchestration_plan_alias, orchestration_deprecations}_test.go`,
`internal/doctor/check_config_planner_test.go`,
`internal/planadvisor/questions_lens_test.go`,
`internal/setup/opencode_agent_planner_test.go`,
`cmd/centinela/{evidence_init_retired, hook_orchestration_plan,
start_notices}_test.go`. Existing fixtures re-pointed across
`tests/{unit,integration,acceptance}` and the five prompt-list acceptance tests.

## Architecture Compliance

- Boundary checks passed. `internal/config` still imports nothing internal (the
  alias helper is a pure map function over `RoleModelValue`). No import cycle
  appeared: `internal/evidence` already imported `internal/workflow`, so
  `EnsureRoleAllowed` lives in `internal/evidence` and the `cmd/` fallback the
  plan reserved was not needed. `cmd/centinela` only wires; the D7 predicate,
  the config alias and the contract resolver are all domain-side.
- **D3 honored — the PR #83 CRITICAL class is structurally excluded.** The plan
  branch lives inside `workflow.RequiredEvidenceRoles`, the single resolver
  already used by all four consumers (complete gate, hook directive, `verify`,
  `complete_verify`). No new call site uses `orchestration.RequiredRoles*`
  directly for an operator-facing statement. Dogfooded: for the same feature the
  directive and the gate name the same set under both contract states.
- G1 file size: every file under `internal/` and `cmd/` is ≤100 lines including
  `_test.go` (full scan). `start.go` reached 102 and was split into
  `start_notices.go`. The prompt doc is 129 lines against the ≤130 budget.
- G7: no business logic moved into the outer layer.
- Scaffold mirrors byte-identical (`cmp`) for `planner-prompt.md`,
  `evidence-contract.md`, `senior-engineer-prompt.md`, `artifact-templates.md`;
  the two retired prompts' mirrors were deleted in the same change.
  `workflow-enforcement.md` is on `mirrorParityAllowlist` — no mirror created.

## Type-Safety Notes

- `PlanContract` is a typed struct field with `omitempty`; absence (never a
  sentinel value and never a clock read) is the legacy signal.
- `aliasPlanRole[T any](models map[string]RoleModelValue, out map[string]T)` is
  generic over the two accessor value shapes rather than `any` + casts, so both
  call sites stay fully type-checked.
- `retiredPlanRoles` is a `map[Role]bool` keyed by the domain `Role` type, not
  by loose strings.
- No `interface{}`/`any` introduced anywhere in this change.

## Trade-Offs

- **D6 fallback NOT taken.** The merged prompt lands at 129 lines against the
  ≤130 budget with both lenses' analysis items, the CLI mandate, the Deferred
  Findings block and the union output headers all intact. No per-file budget
  exception and no aggregate-ratchet assertion were added. The savings came from
  deduplicating the shared How-to-Invoke / authoring-rules / Required-Artifact
  blocks (215 → 129 lines) and from compressing prose — never lens content.
- Deprecation surfaced as a detail line on the existing `config` doctor check
  rather than a new registry row: it is exactly what `configCheck` documents
  itself as reporting (centinela.toml advisories), and a new row would shift the
  registry's fixed check order and counts for unrelated tests.
- `artifact new <feature> big-thinker` needed no guard — `ParseKind` already
  rejects role slugs, since plan roles are not artifact kinds. The D7 guard is
  applied at `evidence init` only.
- Legacy model keys are aliased, never rewritten or removed from the returned
  map, so every existing consumer and its tests keep resolving them unchanged.
  An explicit `planner` entry in *either* value form suppresses the alias, so an
  aliased legacy override table cannot shadow an explicit planner tier.
- Alias precedence `big-thinker` > `feature-specialist`, per D8: the strategy
  lens set the reasoning tier the planner inherits.

## Divergences from the plan

1. `internal/evidence/invalidation_targets.go` — the plan said it "needs **no**
   change; verify this in a test rather than by inspection". Verification found
   a real hazard the file's own `validate` branch already guards against: a
   legacy (unpinned) workflow rewinding past `plan` would keep its stale legacy
   pair, which its legacy gate would immediately re-satisfy on the next pass.
   Added a `case "plan"` that also sheds `big-thinker` + `feature-specialist`,
   mirroring the existing `validation-specialist` line. One line of policy,
   asserted in `invalidation_targets_test.go`; a no-op for pinned workflows
   because those files never exist.
2. `tests/acceptance/planner_prompt_test.go` (listed under S5) was **not**
   written — the code-step brief explicitly excludes writing `tests/acceptance`
   files. Handed to qa-senior. The existing acceptance suite already covers
   `planner-prompt.md` for required section headings, byte-identical mirror, the
   ≤130 budget, the CLI mandate, `evidence schema planner`, the snapshot rule
   and Deferred Findings; what remains is the lens-ordering assertion (strategy
   index < spec index), the union output headers, the legacy-pointer line, and
   the "legacy prompt files no longer exist" check.

## Verification run

- `go build ./...` — clean.
- `go vet ./...` — clean.
- `go test ./... -run xxxNONE` — clean (no test-compile breakage).
- `go test ./...` — **3598 passed, 0 failed, across 45 packages**.
- `./scripts/check-fmt.sh` — clean.
- `./scripts/check-coverage.sh` — **97.1% ≥ 95.0%**.
- Per-package coverage on touched packages: workflow 97.3, orchestration 97.8,
  evidence 96.8, config 98.3, setup 97.3, doctor 96.0, cmd/centinela 96.5,
  planadvisor 94.7. The planadvisor figure is pre-existing: the two files this
  feature touched (`advisor.go`, `questions.go`) are at 100%; the residue is in
  `context_summary.go`, `failures.go` and `roadmap_context.go`, untouched here.
- Dogfood (`/tmp/centinela-planner` built from `./cmd/centinela`, throwaway temp
  git repo): `start` pins `"planContract": "planner-v1"`; the plan directive
  names exactly `planner` at the reasoning tier and lists only
  `.workflow/<f>-planner.{md,json}`; `evidence init <pinned> big-thinker` is
  refused naming `planner`; `evidence init <unpinned> big-thinker` and
  `feature-specialist` both succeed; the unpinned directive names the legacy
  pair and its `complete` failure carries the contract annotation without
  offering planner; a pinned workflow with a hand-authored legacy pair is
  blocked naming planner; a legacy model-override key prints the deprecation
  notice at `start` and at `doctor` but never in the hook output.

## Deferred Findings

No new out-of-scope discoveries. The three findings this feature surfaced
(`codex-claude-role-agent-registry`, `managed-agent-retirement-sweep`,
`prompt-doc-budget-ratchet`) were already deferred at the plan step and each was
re-confirmed during implementation: the Codex adapter emits no per-role
registry, `mergeOpenCodeAgents` still only adds (which is what makes the
no-clobber guarantee hold and what leaves retired agents resident in existing
`opencode.json` files), and the prompt budget remains per-file. None re-deferred
— duplicating an existing roadmap entry would be noise. Recorded slugs: none.

## Handoff

- Next role: qa-senior
- Open items: write the S7 acceptance suite
  `tests/acceptance/unified_plan_specialist_test.go` from
  `specs/unified-plan-specialist.feature`, driving the binary in a temp repo
  with a **local bare origin** (never a network remote — a real push hangs
  `go test`); add `tests/acceptance/planner_prompt_test.go` per divergence 2;
  and cover directive-vs-gate agreement end-to-end through the binary for both
  the pinned and the unpinned contract.
