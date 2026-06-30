# Plan: fix-init-managed-sync-drift

## Summary

`centinela init` leaves a greenfield OpenCode project permanently reporting 2
pending migrations because `setupOpenCode()` writes AGENTS.md + the plugin via
the legacy headerless writers, while the migration expects the
`managed-version`-header'd content. Fix: route `setupOpenCode()` through the same
managed-sync path `setupAider()` already uses.

## Change (cmd/centinela/init_agent.go)

Replace the body of `setupOpenCode()` with the sync-plan pattern:
```go
func setupOpenCode() error {
	plan, err := setup.BuildSyncPlan("opencode")
	if err != nil { return err }
	for _, it := range plan.Items {
		if it.Action == setup.SyncManualReview {
			fmt.Println(ui.StyleYellow.Render("⚠ manual-review " + it.Path + " (" + it.Reason + ")"))
		}
	}
	if err := setup.ApplySync(plan); err != nil {
		return fmt.Errorf("failed to write OpenCode assets: %w", err)
	}
	for _, it := range plan.Items {
		if it.Action != setup.SyncManualReview {
			fmt.Println(ui.RenderSuccess(string(it.Action) + " " + it.Path))
		}
	}
	return nil
}
```
`openCodeAdapter.PlanItems()` already plans opencode.json (planOpenCodeConfig) +
plugin (planPluginFile) + AGENTS.md (planAgentsFile) with the header'd `target`,
so the separate `InjectOpenCodeConfig` + `EnsureAgentsFile` + `EnsureOpenCodePlugin`
calls are no longer needed in init.

## Why it works (verified)

In a fresh repo: after `init` migrate=2; after writing the managed (header'd)
versions migrate=0 and AGENTS.md starts with the managed-version marker. The new
init path writes exactly those header'd versions.

## Test strategy

- **unit**: `BuildSyncPlan("opencode")` on a clean dir yields create actions for
  plugin + AGENTS.md with the managed-version header in the target content.
- **integration**: apply the opencode plan, then re-plan → no items need update
  (idempotent: init's output == migrate's expected).
- **acceptance** (binary, the real bug): in a temp repo, `centinela init` then
  `centinela migrate` reports 0 pending; AGENTS.md + plugin carry the header.

## Rollout

Bug fix, no migration/config. Legacy `Ensure*` writers retained (used by tests),
just unused by init.

## Risks

Low. Mirrors the proven `setupAider` path; acceptance test guards the
init→migrate idempotency that was previously untested.
