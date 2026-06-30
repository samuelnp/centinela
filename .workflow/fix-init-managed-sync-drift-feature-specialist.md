# fix-init-managed-sync-drift — feature-specialist

## Behavior Summary

After the fix, `centinela init` writes AGENTS.md and the OpenCode plugin in their
managed form (with the `centinela:managed-version=N` header) via the same
sync-plan path `setupAider` uses. A subsequent `centinela migrate` finds nothing
to migrate.

## Acceptance Criteria (Gherkin)

See `specs/fix-init-managed-sync-drift.feature`: fresh init → migrate reports 0
pending; AGENTS.md + plugin begin with the managed-version marker; the opencode
sync plan is idempotent (no create/update on re-plan).

## UX States

- **init**: prints create/update actions per managed asset (like the Aider path).
- **migrate after init**: "Managed docs and setup assets are already up to date."

## Edge Cases

- After init, migrate = 0 pending (init output == migrate expected).
- AGENTS.md starts `<!-- centinela:managed-version=…`; plugin starts
  `// centinela:managed-version=…`.
- Re-running BuildSyncPlan("opencode") after init → no create/update items.
- `.claude/settings.json` pending in established projects is a real new hook, out
  of scope.

## Out-of-Scope

- Removing the legacy `Ensure*` writers (unit tests still use them).

## Handoff

→ senior-engineer: rewrite `setupOpenCode()` to the BuildSyncPlan/ApplySync
pattern; keep file ≤100 lines.
