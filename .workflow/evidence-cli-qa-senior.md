# QA-Senior Report: evidence-cli
**Date:** 2026-05-28

## Test Inventory

| Tier        | File                                                                        | Scenarios covered |
|-------------|-----------------------------------------------------------------------------|-------------------|
| unit        | internal/evidence/io_test.go                                                | S6 (atomic write / repair orphan), S7 (concurrent serialize) |
| unit        | internal/evidence/io_more_test.go                                           | S6 (atomic write round-trip, no temp remains) |
| unit        | internal/evidence/io_errors_test.go                                         | S6 error paths (mkdir fail, rename fail) |
| unit        | internal/evidence/lock_test.go                                              | S7 (concurrent appends serialize, timeout hint) |
| unit        | internal/evidence/appender_test.go                                          | S3 (dedup), S4 (read returns typed value) |
| unit        | internal/evidence/setter_branches_test.go                                   | S12 (extra slot), S2 (scalar fields) |
| unit        | internal/evidence/schema_test.go + schema_more_test.go                     | S1 (skeleton valid), S11 (unknown fields preserved) |
| unit        | internal/evidence/validate_test.go + validate_extra_test.go                | S5 (validate exits non-zero with hint) |
| unit        | internal/evidence/repair_more_test.go                                       | S6 (repair idempotent on missing) |
| unit        | internal/evidence/artifact_test.go + artifact_more_test.go                 | S10 (artifact new unknown kind), S13 (artifact templates) |
| unit        | internal/evidence/companion_test.go + companion_more_test.go               | S1 (companion written on init) |
| unit        | internal/evidence/roles_test.go                                             | S8 (unknown role exits non-zero), S9 (not-found hint) |
| unit        | internal/evidence/repair_race_test.go (NEW)                                 | EC1 (repair deletes live tmp — mtime gap pinned), EC2 (lock files not cleaned) |
| unit        | internal/evidence/extra_collision_test.go (NEW)                             | EC3 (extra.feature collision round-trip behavior pinned) |
| unit        | internal/hookpolicy/format_evidence_test.go                                 | S14 (postwrite reformats), S15 (scoped to active feature) |
| unit        | internal/hookpolicy/format_evidence_more_test.go                            | S14/S15 (non-object JSON, dir filter, sort) |
| unit        | cmd/centinela/hook_postwrite_reformat_test.go                               | S14 (hook end-to-end reformat), S15 (ignores other feature) |
| unit        | cmd/centinela/hook_postwrite_extract_test.go                                | S14 (path extraction branches) |
| unit        | cmd/centinela/hook_postwrite_glob_test.go                                   | S14 (glob skips non-workflow JSON) |
| unit        | cmd/centinela/hook_postwrite_noworktree_test.go (NEW)                       | EC4 (no-op outside worktree pinned) |
| unit        | cmd/centinela/evidence_init_test.go                                         | S1 (init drops skeleton), S8 (unknown feature) |
| unit        | cmd/centinela/evidence_set_test.go                                          | S2 (set writes atomically), S5 (no temp remains) |
| unit        | cmd/centinela/evidence_append_test.go                                       | S3 (append deduplicates) |
| unit        | cmd/centinela/evidence_read_test.go                                         | S4 (read field), S9 (not-found hint) |
| unit        | cmd/centinela/evidence_validate_test.go                                     | S5 (validate exits non-zero with hint) |
| unit        | cmd/centinela/evidence_repair_test.go                                       | S6 (repair removes orphan) |
| unit        | cmd/centinela/evidence_schema_test.go                                       | schema skeleton output |
| unit        | cmd/centinela/artifact_test.go                                              | S10, S13 (artifact new CLI) |
| integration | tests/integration/enforce_actionable_orchestration_evidence_integration_test.go | S5, S8 multi-package |
| acceptance  | tests/acceptance/agent_evidence_contract_acceptance_test.go                 | S16 (prompts reference evidence-contract), S17 (scaffold mirror) |
| acceptance  | tests/acceptance/prompts_mandate_cli_acceptance_test.go                     | S16 (no python3/heredoc, mandates CLI) |
| acceptance  | tests/acceptance/scaffold_arch_parity_acceptance_test.go                    | S17 (scaffold mirror parity) |

## Coverage Gaps

All 17 feature scenarios have executable assertions as mapped above. The following scenarios were partially or fully deferred with rationale:

- **S6 mtime guard** (race condition Repair vs concurrent append): pinned as known limitation in `repair_race_test.go`. No mtime guard exists; fixing it is deferred to a follow-up since it requires a non-trivial atomic check + mtime comparison that changes the Repair contract.
- **EC2 lock accumulation cleanup**: pinned in `repair_race_test.go`. The fix (extending Repair to glob `*.lock`) is a small addition deferred to avoid scope creep here.
- **EC3 extra key clobber on round-trip**: pinned as a known limitation in `extra_collision_test.go`. The correct fix is to make `setExtra` reject reserved keys. Deferred; test documents the regression boundary.
- **false-positive in prompts_mandate_cli**: the proximity-window logic and `forbiddingPrefixes` list already handle "Do NOT use python3 -c" and similar negative-example text without tripping the test.

## Acceptance Wiring

```toml
[validate]
commands = [
  "go test ./...",
  "./scripts/check-coverage.sh"
]
```

`go test ./...` recurses into `tests/acceptance/` (package `acceptance_test`) and runs all three tiers. Confirmed green: 859 tests pass, all gates pass.

## Handoff

- Next role: validation-specialist
- Edge-case report: `.workflow/evidence-cli-edge-cases.md` (produced by edge-case-tester; 121 lines, 16 risks)
- New test files added:
  - `internal/evidence/repair_race_test.go` (65 lines) — EC1 + EC2
  - `internal/evidence/extra_collision_test.go` (72 lines) — EC3
  - `cmd/centinela/hook_postwrite_noworktree_test.go` (49 lines) — EC4
- All tests green: `go test ./...` 859 passed, 0 failed
- `centinela validate` all gates passed
- `centinela complete` NOT invoked — awaiting orchestrator confirmation
