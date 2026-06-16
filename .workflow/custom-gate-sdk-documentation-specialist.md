# custom-gate-sdk — documentation-specialist

Internal-surface (right-sized) docs step.

## KB Pages

No standalone KB page — this extends the existing gate/validate surface via a
new config section, documented through the brief, plan, and regenerated project
docs.

## project-docs Entries

- `.workflow/custom-gate-sdk-changelog.md` — one-line `feat` changelog.
- Regenerated `docs/project-docs/index.html` (picks up the brief, plan, and
  changelog).

## User-facing note

Teams can now define their own mechanical gates without forking Centinela. Add
`[[gates.custom]]` entries to `centinela.toml`:

```toml
[[gates.custom]]
name = "no-console-log"
command = "! grep -rn 'console.log' src/"
severity = "fail"          # fail (blocks) | warn (reported, non-blocking)
output = "lines"           # blob (default) | lines (each stdout line = a violation)
# enabled = true, timeout_seconds = 60, diff_aware = false
```

During `centinela validate` each enabled custom gate runs its command (shell,
with a per-gate timeout): exit 0 → pass; non-zero → fail or warn per `severity`,
with command output surfaced as gate details. Custom gates are first-class —
they emit `gate-failure` telemetry and are **baseline-able by the audit ratchet**
(`output = "lines"` fingerprints each line individually, so `centinela audit`
tolerates pre-existing custom violations and blocks new ones). The built-in gates
are reference implementations of the same `Result` contract. With `diff_aware =
true`, the changed-file list is passed via `CENTINELA_CHANGED_FILES`. Trust
model: commands run from checked-in config (same as `[validate] commands`).

## Outcome

Docs generated and validated. Handoff → complete.
