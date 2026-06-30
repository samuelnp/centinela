# fix-init-managed-sync-drift — big-thinker

## Problem

Greenfield `centinela init` leaves the migration system reporting pending
updates for `AGENTS.md` and `.opencode/plugins/centinela.js` — the
`SETUP MIGRATION REQUIRED` prompt fires immediately and on every turn.

## Scope

Route init's `setupOpenCode()` through the managed-sync path
(`BuildSyncPlan("opencode")` + `ApplySync`) so it writes the same
`managed-version`-header'd content the migration expects. One-function change.

## Dependencies & Assumptions

- `setupAider()` already uses this path and is consistent — it is the model.
- `openCodeAdapter.PlanItems()` covers opencode.json + plugin + AGENTS.md with
  the header'd `target`. Verified: after init migrate=2; after writing the
  managed versions migrate=0.

## Risks

Low. Proven path; the only prior gap was init bypassing it. Legacy `Ensure*`
writers stay (unit tests reference them), just unused by init.

## Rollout

Bug fix; no migration/config. Acceptance test pins init→migrate idempotency that
was previously untested (which is why the bug shipped).

## Handoff

→ feature-specialist: encode init→migrate-clean, header-present, and
plan-idempotent as acceptance scenarios.
