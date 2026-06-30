# Feature: fix-init-managed-sync-drift

## Problem (bug)

In a greenfield project, `centinela init` immediately leaves the migration
system reporting pending updates — `centinela migrate` flags `AGENTS.md` and
`.opencode/plugins/centinela.js` as needing migration the moment after init
wrote them. The `SETUP MIGRATION REQUIRED` prompt then fires on every turn.

## Root cause

The migration manifest's `planManagedFile` (internal/setup/sync_managed_files.go)
treats a managed file as up-to-date only when its content equals
`target = header + "\n" + content`, where `header` is the
`centinela:managed-version=N` marker. Its `if s == legacy → SyncUpdate` branch
flags any file that holds the **headerless** content as needing an update.

`init`'s `setupOpenCode()` (cmd/centinela/init_agent.go) writes those two files
via the **legacy** writers `EnsureAgentsFile` / `EnsureOpenCodePlugin`, which
emit the headerless `legacy` content. So a freshly-init'd OpenCode project is
permanently "2 migrations pending". `setupAider()` already avoids this by using
the managed-sync path (`BuildSyncPlan` + `ApplySync`); `setupOpenCode()` is the
only setup path still on the legacy writers.

## Fix

Make `setupOpenCode()` use the managed-sync path, mirroring `setupAider()`:
`plan := BuildSyncPlan("opencode")` → warn on manual-review → `ApplySync(plan)` →
report actions. The opencode adapter's `PlanItems()` already covers
opencode.json + plugin + AGENTS.md with the correct header'd content, so init
and migrate finally agree → 0 pending after init.

## Scope / out of scope

- In scope: `setupOpenCode()` only. Legacy `EnsureAgentsFile`/`EnsureOpenCodePlugin`
  stay (still referenced by unit tests) but init no longer calls them.
- Out of scope: the `.claude/settings.json` migration seen in established projects
  — that is a *real* pending hook from a newly-merged feature, not this drift bug.

## Acceptance

- After `centinela init` in a fresh repo, `centinela migrate` reports **0**
  pending managed assets (AGENTS.md + plugin carry the managed-version header).
- AGENTS.md begins with `<!-- centinela:managed-version=…`, plugin with
  `// centinela:managed-version=…`.
- Existing setup tests stay green.

## Risks

Low. The managed-sync path is already proven by `setupAider` and by
`migrate --apply`. A new acceptance test pins init→migrate idempotency.
