# Edge Cases: fix-init-managed-sync-drift

## Covered

| # | Edge case | Handling | Test |
|---|-----------|----------|------|
| 1 | Fresh init must leave 0 pending migrations (the regression) | init writes managed-sync target content | `acceptance:TestAccInitLeavesNoPendingMigration` |
| 2 | Managed files carry the managed-version header after init | header'd `target` from the sync plan | `acceptance:TestAccInitLeavesNoPendingMigration`, `integration:TestOpencodeSyncApplyIsIdempotent` |
| 3 | Clean repo plans create for plugin + AGENTS.md | `BuildSyncPlan("opencode")` → SyncCreate | `unit:TestOpencodeSyncPlanCreatesManagedAssets` |
| 4 | Apply is idempotent (init output == migrate expected) | re-plan yields no create/update | `integration:TestOpencodeSyncApplyIsIdempotent` |
| 5 | Pre-existing unmanaged custom content | SyncManualReview → ApplySync skips (no clobber), setupOpenCode returns nil | `cmd:TestSetupOpenCodeAlreadyConfigured` (existing) |
| 6 | Second init is a no-op | managed target already present → no items | `cmd:TestSetupOpenCode...idempotent` (existing) |

## Residual Risks

- Legacy `EnsureAgentsFile`/`EnsureOpenCodePlugin`/`InjectOpenCodeConfig` remain
  (still unit-tested in internal/setup); init no longer calls them. If they are
  ever reintroduced to an init path, the drift returns — the acceptance test
  guards against that.
- `.claude/settings.json` migrations in established projects are real pending
  hooks from newly-merged features, not this drift — out of scope.
