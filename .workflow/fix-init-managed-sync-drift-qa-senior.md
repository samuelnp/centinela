# fix-init-managed-sync-drift — qa-senior

## Test Inventory

- **unit** `tests/unit/fix_init_managed_sync_drift_unit_test.go`:
  `BuildSyncPlan("opencode")` on a clean repo plans `create` for AGENTS.md + the
  plugin.
- **integration** `tests/integration/...`: `ApplySync` writes the managed-version
  header and re-planning is idempotent (no create/update).
- **acceptance** `tests/acceptance/...` (the regression): binary `centinela init`
  then `centinela migrate` reports 0 pending; AGENTS.md carries the header.
- Existing `cmd/centinela` init tests (`TestSetupOpenCode...`) still cover the
  manual-review + idempotent branches of the rewritten `setupOpenCode`.

## Coverage Gaps

None. Total 95.0% ≥ gate. The rewritten `setupOpenCode` is exercised by existing
init tests (create/report + manual-review + no-op paths); the bug-specific
behavior is pinned by the new acceptance test.

## Acceptance Wiring

`specs/fix-init-managed-sync-drift.feature` → `TestAccInitLeavesNoPendingMigration`
(+ unit/integration for plan + idempotency). `centinela.toml` already runs
`go test ./tests/acceptance/...`.

## Handoff

→ validation-specialist: full suite + gates green; produce the gatekeeper report.
