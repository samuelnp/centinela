# unified-plan-specialist — qa-senior

### QA-Senior Report: unified-plan-specialist
**Date:** 2026-07-29

## Test Inventory

S7 (spec/acceptance wiring), handed off by senior-engineer as the sole open
item. 19 new acceptance tests across 11 files, each with a `// Scenario:`
marker matching `specs/unified-plan-specialist.feature` verbatim (spec
traceability requires exact, normalized title match), plus 2 colocated
regression tests filling gaps the senior-engineer's own unit suite left open.

| Tier | File | Scenarios |
|------|------|-----------|
| acceptance | tests/acceptance/unified_plan_specialist_fixtures_test.go | shared binary/fixture helpers (no scenario of its own) |
| acceptance | tests/acceptance/unified_plan_specialist_fresh_start_test.go | Fresh workflow's plan directive names exactly one planner role |
| acceptance | tests/acceptance/unified_plan_specialist_complete_test.go | Planner evidence passes plan complete; forged legacy pair blocked naming planner |
| acceptance | tests/acceptance/unified_plan_specialist_evidence_init_test.go | evidence init refused: pinned workflow; no workflow state at all |
| acceptance | tests/acceptance/unified_plan_specialist_legacy_test.go | Legacy complete pair advances; partial legacy set fails w/ contract annotation; legacy evidence init succeeds |
| acceptance | tests/acceptance/unified_plan_specialist_directive_gate_test.go | Directive/gate agreement outline (pinned + unpinned); guided profile still resolves planner, prints no evidence requirement |
| acceptance | tests/acceptance/unified_plan_specialist_config_test.go | Legacy model key aliases to planner; deprecation notice at doctor/start only, absent from hook |
| acceptance | tests/acceptance/unified_plan_specialist_advisor_test.go | Plan-advisor two-lens header, strategy/spec tags only |
| acceptance | tests/acceptance/unified_plan_specialist_prompt_content_test.go | planner-prompt.md lens order, single headings/CLI block, line budget, legacy docs absent |
| acceptance | tests/acceptance/unified_plan_specialist_scaffold_parity_test.go | Scaffold mirror byte-identical; legacy mirrors absent |
| acceptance | tests/acceptance/unified_plan_specialist_statusline_test.go | Statusline never shows a planner/legacy role sub-workflow as primary |
| unit (cmd/centinela, colocated) | cmd/centinela/hook_statusline_view_more_test.go | `TestPrimaryWorkflowSkipsPlannerRoleWorkflow` — the "-planner" isRoleWorkflow addition had no colocated assertion (only the pre-existing "-big-thinker" suffix was tested) |
| unit (cmd/centinela, colocated) | cmd/centinela/evidence_init_retired_test.go | `TestEvidenceInitRefusesRetiredRoleWithNoWorkflowAtAll` — the D7 "no workflow at all" CLI path had no colocated test; writing it surfaced the bug below |

No `tests/unit` or `tests/integration` package-level additions were needed:
every internal-package behavior this feature touches (contract pinning,
policy/tier resolution, evidence roles, config aliasing, doctor check,
plan-advisor lenses, opencode agent config) already has colocated `_test.go`
coverage from the senior-engineer's own slices (see
`.workflow/unified-plan-specialist-senior-engineer.md` "Colocated unit tests
added"). The only gaps found were at the `cmd/centinela` CLI-surface and
acceptance (spec-scenario) tiers, filled above.

## Bug found and fixed (production code, not test-only)

`cmd/centinela/evidence_init.go` ran `requireKnownFeature` before
`evidence.EnsureRoleAllowed`, so `evidence init <feature-with-no-workflow>
feature-specialist` returned the generic "unknown feature" error instead of
the D7-mandated "role is retired; use planner" message — failing the spec's
"errors with no workflow state at all" scenario outright (confirmed live
before the fix: `Error: unknown feature "ghost-feature" ...`). Fixed by
reordering the two checks (`EnsureRoleAllowed` already treats a
missing/unreadable workflow as refused, so it needs no workflow to exist).
Two pre-existing colocated tests (`TestEvidenceInitRejectsUnknownFeature`,
`TestEvidenceInitListsActiveOnUnknown`) used the now-retired `big-thinker`
slug as an arbitrary role placeholder for the *unrelated* unknown-feature
case; they were updated to use `planner` instead, preserving their original
intent. Full `cmd/centinela` suite (534 tests) re-verified green after both
changes. Full detail in `.workflow/unified-plan-specialist-edge-cases.md`.

## Coverage Gaps

None remaining against the 19 spec scenarios — all 19 have a passing,
uniquely-mapped acceptance test (verified via `go test
./tests/acceptance/ -run TestUPS_ -v`: 19/19 pass). The three roadmap-deferred
findings from the plan step (`codex-claude-role-agent-registry`,
`managed-agent-retirement-sweep`, `prompt-doc-budget-ratchet`) remain
out-of-scope by design; none re-deferred here (see Deferred Findings below).

## Acceptance Wiring

`validate.commands` in `centinela.toml` was NOT modified (not this role's
task, and it already covers the acceptance tier):

```toml
[validate]
commands = [
  "go test ./...",
  "./scripts/check-coverage.sh",
  "./scripts/check-fmt.sh"
]
```

`go test ./...` includes `tests/acceptance/...` (confirmed: `go test
./tests/acceptance/... ` → 765 passed, all under the module's `./...` glob).
A dedicated `go test ./tests/acceptance/...` line was deliberately removed in
an earlier feature to avoid a redundant ~200s re-run within
`centinela complete`'s claim-verification wall-clock budget — see the
comment directly above `commands` in `centinela.toml`.

## Verification run

- `go build ./...` — clean.
- `go test ./... -run xxxNONE` — clean (no test-compile breakage).
- `go test ./tests/acceptance/ -run TestUPS_ -v` — 19 passed (all 19 spec
  scenarios).
- `go test ./tests/acceptance/...` — 765 passed.
- `go test ./...` — 3621 passed, 45 packages.
- `./scripts/check-fmt.sh` — clean.
- `centinela evidence validate unified-plan-specialist` — ok.
- `centinela validate` (full gate suite incl. spec-traceability) — run
  separately; see gatekeeper handoff for the recorded gate line.

## Deferred Findings

None new. The three findings this feature surfaced at the plan step
(`codex-claude-role-agent-registry`, `managed-agent-retirement-sweep`,
`prompt-doc-budget-ratchet`) were already recorded and re-confirmed by
senior-engineer; duplicating them here would be roadmap noise.

## Handoff

- Next role: validation-specialist (this workflow predates `adversarial-v1`
  is false — this feature's own workflow is itself LEGACY on the plan
  contract per `.workflow/unified-plan-specialist.json`, planned before
  unified-plan-specialist shipped; its validate step is pinned
  `validateContract: adversarial-v1`, so the actual next subagent is the
  gatekeeper/adversarial verifier per the repo's current validate-step
  contract — `handoffTo` is kept as the evidence-schema default
  `validation-specialist` per `centinela evidence schema qa-senior`, which
  the validator does not gate against the workflow's validate contract).
- Edge-case report: `.workflow/unified-plan-specialist-edge-cases.md`
  (produced via `centinela artifact new unified-plan-specialist edge-cases`,
  filled with every negative path exercised, file:test references, and the
  bug-found-and-fixed note).
