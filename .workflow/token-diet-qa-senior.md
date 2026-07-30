# token-diet — qa-senior

## Test Inventory

18 new binary/direct-call acceptance files under `tests/acceptance/` cover all
35 Gherkin scenarios (`specs/token-diet.feature`), each ≤100 lines. No new
`tests/unit/` file was added — every "assembled behavior" the plan proposed
for the unit tier (legacy-122-path superset, all-descriptive UX pass, non-UX
role bypass, `ResolveModel` precedence 1→4 with aliases, `model_map` override)
is already covered colocated (`internal/orchestration/{plan_snapshot,
evidence_ux_tag,evidence_ux,resolve,model_routing}_test.go`), so adding a
duplicate tier-level file would only pin the same assertion twice.

| Tier | File | Scenarios |
|------|------|-----------|
| acceptance | `token_diet_plan_snapshot_test.go` (77) | 3 (does-not-grow, empty-repo-identical, construction-not-existence) |
| acceptance | `token_diet_plan_evidence_test.go` (80) | 4 (exact-2 validates, legacy-superset validates, missing-own-plan rejected, empty-inputs rejected) |
| acceptance | `token_diet_plan_init_test.go` (72) | 2 (init prefill, retired-role same shrunken set) |
| acceptance | `token_diet_plan_normalize_test.go` (32) | 1 outline (8-row path normalization) |
| acceptance | `token_diet_plan_docs_test.go` (42) | 1 (docs + scaffold mirror) |
| acceptance | `token_diet_ux_tag_test.go` (69) | 2 (descriptive satisfies, bare still works) |
| acceptance | `token_diet_ux_role_test.go` (47) | 2 (genuinely-absent still missing, non-UX role unaffected) |
| acceptance | `token_diet_ux_normalize_test.go` (51) | 2 outlines (6-row colon-cut, 3-row degenerate) |
| acceptance | `token_diet_hook_fixtures_test.go` (62) | shared fixtures only |
| acceptance | `token_diet_hook_render_test.go` (69) | 2 (first-renders, unchanged-not-re-rendered) |
| acceptance | `token_diet_hook_render2_test.go` (61) | 3 (changed-re-renders, new-session-re-renders, reformat-no-re-render) |
| acceptance | `token_diet_hook_failopen_test.go` (59) | 1 outline (6-row absent/corrupt/unreadable state x payload) |
| acceptance | `token_diet_hook_noroadmap_test.go` (50) | 2 (unwritable dir, missing/invalid roadmap) |
| acceptance | `token_diet_hook_git_test.go` (76) | 2 (clean tree, independent worktrees) |
| acceptance | `token_diet_directive_test.go` (95) | 5 (no dated snapshot, alias outline, model_map override, codex fallback, directive prints alias) |
| acceptance | `token_diet_capability_test.go` (56) | 2 (9-row classification outline, retired-pin default profile) |
| acceptance | `token_diet_scope_test.go` (46) | 1 (evidence shape + scaffold size unchanged) |
| acceptance | `token_diet_fixtures_test.go` (96) | shared fixtures only |

Total: 1,140 new lines across 18 files; 35/35 scenarios carry an executing
`// Scenario:` marker under `// Acceptance: specs/token-diet.feature`.

## Coverage Gaps

None. Every scenario in `specs/token-diet.feature` (35, including 6 Scenario
Outlines counted by title) has an executing acceptance-tier assertion. Two
structural CLI-surface limitations were hit while writing tests — not
coverage gaps, since the underlying behavior IS exercised via the exported
function the CLI itself calls — documented in
`.workflow/token-diet-edge-cases.md` under Residual Risks:
1. `evidence validate <feature>` always passes nil `uiPaths`, so a
   ux-ui-specialist evidence file can only be observed passing by calling
   `orchestration.ValidateEvidence` directly with uiPaths (same function the
   CLI delegates to).
2. A wholly-empty `inputs` list always trips the generic "incomplete evidence
   fields" branch before the plan-snapshot check, at the full-evidence level;
   the per-path message is asserted directly at the colocated unit tier.

## Acceptance Wiring

`centinela.toml` — unmodified, confirmed already correct:

```toml
[validate]
commands = [
  "go test ./...",
  "./scripts/check-coverage.sh",
  "./scripts/check-fmt.sh"
]
```

`go test ./...` includes `tests/acceptance/...` (no separate acceptance
invocation is configured or needed — the existing comment in
`centinela.toml` explains this was a deliberate wall-clock optimization from
an earlier feature). All 35 new tests run under `go test ./...` and under
`rtk go test ./tests/acceptance/ -run TestTD_` (35 passed).

## Deferred Findings

None. No new gap was discovered while writing this suite — every scenario in
the spec was reachable and green against the code senior-engineer landed.

## Handoff

- Next role: validation-specialist
- Edge-case report: `.workflow/token-diet-edge-cases.md` (produced directly
  in this step; 12 covered items + 3 residual-risk notes, each with a
  `file:test` reference)
