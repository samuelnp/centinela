# fix-init-managed-sync-drift — senior-engineer

## Files Touched

- `cmd/centinela/init_agent.go`: `setupOpenCode()` rewritten to use the
  managed-sync seam — `setup.BuildSyncPlan("opencode")` → warn on manual-review
  → `setup.ApplySync(plan)` → report actions — mirroring `setupAider()`. It no
  longer calls the legacy `InjectOpenCodeConfig` / `EnsureOpenCodePlugin` /
  `EnsureAgentsFile` (those wrote header-less content the migration flagged).

## Architecture Compliance

cmd-layer wiring only; the sync logic lives in `internal/setup`. The opencode
adapter's `PlanItems()` already plans opencode.json + plugin + AGENTS.md with the
header'd `target`, so init and the migration manifest now produce identical
content. Legacy `Ensure*` writers stay (referenced by `internal/setup` unit
tests); init simply stops using them.

## Type-Safety Notes

No type changes. Verified end-to-end: in a fresh repo `centinela init` prints
`create opencode.json / .opencode/plugins/centinela.js / AGENTS.md`, then
`centinela migrate` reports 0 pending ("already up to date"), and AGENTS.md
begins with `<!-- centinela:managed-version=1 template=AGENTS.md -->`.

## Trade-Offs

Kept the legacy writers rather than deleting them (still unit-tested); the fix is
the minimal one-function reroute. The init UX now matches the Aider path
(action-per-asset lines) instead of the old bespoke messages.

## Handoff

→ qa-senior: unit (opencode plan targets carry the managed-version header),
integration (apply then re-plan is idempotent — no create/update), acceptance
(binary init→migrate reports 0 pending). Keep file ≤100 lines.
