# spec-conflict-false-positives — qa-senior

## Test Inventory

The senior-engineer's colocated unit + acceptance suite already covers every
false-positive class and every true-positive class at the function level
(`worktree.DetectSpecConflicts` / `worktree.FormatSpecConflicts` called
directly, or `runMerge` called in-process). Auditing that suite against the
rewritten `.feature` scenarios found exactly one gap: nothing drove the real,
compiled `centinela` binary end-to-end through `centinela merge <feature>` for
either the false-positive or the true-positive spec-conflict path. Added two
binary-driven acceptance tests to close it; wrote no other new tests to avoid
duplicating already-thorough coverage.

| Tier | File | Scenarios |
|------|------|-----------|
| acceptance (NEW) | `tests/acceptance/spec_conflict_binary_test.go` (92 lines) | Real `centinela` binary, real git worktrees: identical/companion specs across worktrees never block a real merge; two genuinely diverging worktrees block the real merge once, output bounded |
| acceptance (NEW, helpers) | `tests/acceptance/spec_conflict_binary_helper_test.go` (60 lines) | `seedSpecRepo`, `addWorktreeBranch`, `mainHeadSHA` — reuses `buildCentinela`/`runBin` (merge_steward_auto_dispatch_helper_test.go) and `mustWrite`/`runGit`/`commit` (diff_aware_gatekeeper_acceptance_test.go) |
| acceptance (existing) | `tests/acceptance/parallel_feature_worktrees_test.go` | Spec conflict detected pre-merge (function-driven); identical/superseding specs do not block (function-driven) |
| integration (existing) | `cmd/centinela/merge_test.go` | `TestRunMerge_SpecConflict_Blocks` — `runMerge` in-process against a real git repo |
| unit (existing) | `internal/worktree/spec_conflicts*_test.go` (6 files) | Full false-positive/true-positive/dedup/cap/malformed-input matrix — see `.workflow/spec-conflict-false-positives-edge-cases.md` for the complete scenario-to-test map |
| unit (existing) | `internal/worktree/coverage_merge_helpers_test.go` | `TestIndexByKey_FirstRecordWins`, `TestReadSpecsFrom_UnreadableEntrySkipped` |

## Coverage Gaps

None outstanding against the `.feature` spec. Both spec-conflict scenarios
("Spec conflict across in-flight worktrees is detected before merging" and
"Superseding and identical specs never block a merge") now have an executing
`// Scenario:` marker at both the function-driven acceptance tier and the new
binary-driven tier. The two genuinely unaddressed gaps (scenario deletion,
When/And/second-Then divergence) were already identified and deferred by the
senior-engineer to the Backlog — see Residual Risks in the edge-cases report.

## Acceptance Wiring

`validate.commands` in `centinela.toml` was NOT modified (per hotfix
constraints). It already runs the acceptance tier:

```toml
[validate]
commands = [
  "go test ./... -coverprofile=coverage.out",
  "COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh",
  "./scripts/check-fmt.sh"
]
```

`go test ./...` includes `tests/acceptance/...`, so both new binary-driven
tests execute on every `centinela validate` run.

## Deferred Findings

No genuinely new gaps found. Every residual risk uncovered during this audit
(worktree-only detection scope, cross-file semantic contradictions) is either
already captured as a Backlog item by the senior-engineer
(`spec-conflict-scenario-deletion-detection`,
`spec-conflict-deep-gherkin-diff`) or is a deliberate, already-documented
trade-off in `.workflow/spec-conflict-false-positives-senior-engineer.md`. No
new `centinela roadmap defer` entries were recorded.

## Handoff

- Next role: validation-specialist
- Edge-case report: `.workflow/spec-conflict-false-positives-edge-cases.md`
