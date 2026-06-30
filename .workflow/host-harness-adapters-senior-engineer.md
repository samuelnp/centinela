# host-harness-adapters — senior-engineer

## Files Touched

New (`internal/setup`): `adapter.go`, `adapter_registry.go`, `adapter_claude.go`,
`adapter_opencode.go`, `adapter_aider.go`, `aider_config.go`,
`golden_parity_test.go`, `testdata/golden/` fixtures.
Modified (`internal/setup`): `sync.go` (registry-driven `BuildSyncPlan` +
`applyItem` arms), `sync_types.go` (`SyncKindPrewriteHook`, `SyncAiderConfig`),
`sync_hooks.go`, `sync_managed_files.go` (`#` managed-marker prefix), plus the
3 internal tests that referenced the old kind names.
Modified (`cmd/centinela`): `init.go`, `init_agent.go` (registry validation +
`AgentsFor` dispatch + `setupAider`), `migrate.go`, `migrate_setup.go`.

## Architecture Compliance

- Registry + interface live in `internal/setup` (domain). `internal/setup`
  imports neither `cmd/` nor `internal/ui` (verified) — n-tier layering intact.
- G7: `cmd/centinela` is thin — `--agent` validity (`setup.IsValidAgent`),
  resolution (`setup.AgentsFor`), and the harness list for error messages
  (`setup.RegisteredAgents`) all come from the registry, no hardcoded agent
  lists. Per-harness UI rendering stays in `cmd` (setup cannot import `ui`).
- `BuildSyncPlan` iterates `adaptersFor` — no `useClaude`/`useOpenCode`
  if-ladder; `both` is a `composites` map entry, not a special-case branch.
- `applyItem` keeps the central `SyncKind` switch (pinned Q3).
- G1: every new/modified source file <= 100 lines.

## Type-Safety Notes

- `Capability` and `SyncKind` are string enums with named constants; no `any`
  on the new surface beyond the pre-existing JSON merge code I did not touch.
- Unknown agents return a wrapped `ErrUnknownAgent` (errors.Is-compatible),
  never a panic.

## Trade-Offs

- Collapsed `SyncClaudeHooks` + `SyncOpenCodePlug` into one
  `SyncKindPrewriteHook` so the capability-parity invariant (`blocks-writes`
  => a `SyncKindPrewriteHook` item) is directly assertable and Aider is
  provably advisory. `applyItem` disambiguates Claude vs OpenCode by the
  constant plugin path. Emitted bytes are unchanged — proven by the golden test.
- `.aider.conf.yml` via the managed-marker seam (pinned Q2), no YAML dep;
  unmanaged files surface manual-review and are never clobbered.
- `PlanItems()` is argument-less (matches spec scenarios) vs the brief's
  `PlanItems(agent string)` — adapters are single-harness scoped.

## Handoff

-> qa-senior. Write unit/integration/acceptance tests for the 26 scenarios.
Capability-parity test: iterate `setup.RegisteredAdapters()`, assert each
`blocks-writes` adapter's `PlanItems()` contains a `SyncKindPrewriteHook` item
and Aider's does not. Golden fixtures: `internal/setup/testdata/golden/`.
Byte-parity guard `TestGoldenParityClaudeOpenCode` already green.
