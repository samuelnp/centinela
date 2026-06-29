### Big-Thinker Report: host-harness-adapters
**Date:** 2026-06-29

#### Problem
Centinela wires each host harness as a hand-maintained parallel branch in
`internal/setup/`: Claude Code (`settings*.go`/`hooks.go`/`statusline.go`, JSON
hooks in `.claude/settings.json`) and OpenCode (`opencode_*.go` — a JS plugin,
`opencode.json`, `AGENTS.md`). They join at four seams — `BuildSyncPlan(agent)`'s
`if useClaude/useOpenCode` ladder, `applyItem()`'s `SyncKind` switch, the
`useClaude()/useOpenCode()` predicates, and the `cmd/centinela` wiring
(`init.go`, `init_agent.go`, `migrate.go`, `migrate_setup.go`) hardcoded to
`claude|opencode|both`. Every new harness multiplies that edit surface and risks
silent parity drift. The maintainer pays a growing tax; users of any
un-integrated harness cannot adopt governance at all. Phase 11 of the roadmap
queues Cursor/Aider/Windsurf/Copilot and `codex-support`/`local-harness-support`
explicitly depend on a shared core landing first.

#### Scope
- **In:** Extract a `HarnessAdapter` interface + registry in `internal/setup/`;
  refactor BOTH Claude and OpenCode onto it with ZERO behavior change (golden-file
  byte-for-byte parity); add exactly ONE new harness, Aider; a tiered capability
  model over the fixed vocabulary `{blocks-writes, prompt-context, rules-file}`
  (Claude/OpenCode = all three; Aider = `prompt-context` + `rules-file`, NO
  `blocks-writes`); drive `BuildSyncPlan`/`applyItem` and `cmd` `--agent`
  validation/dispatch off the registry; `init --agent aider` / `migrate --agent
  aider` write Aider files idempotently and scoped; a capability-parity test.
- **Out:** Cursor, Windsurf, Copilot, Codex (separate roadmap features that plug
  into this contract); the orchestration `Runner` model-routing enum — NO `aider`
  runner key (`orchestration_model_map.go`/`resolve.go` untouched); growing the
  capability vocabulary beyond the three values.

#### Dependencies & Assumptions
- Touches `internal/setup/` (new `adapter.go` registry + per-harness adapters +
  `aider_*.go`; `sync.go`/`sync_types.go`/`sync_hooks.go`/`sync_managed_files.go`
  refactored off the registry) and `cmd/centinela/` (`isValidAgent` + dispatch
  read from the registry).
- Reuses, not rewrites, the existing `plan*`/`Inject*`/`writeManaged*` functions
  and `SyncItem`/`SyncKind`/`SyncPlan` types — adapters are thin wrappers so
  managed output is provably unchanged.
- Layer rules: `internal/setup` = Infrastructure; `cmd/` = thin outer layer (G7),
  so `isValidAgent`/dispatch move INTO the registry. G1 ≤100 lines/file → one
  small file per adapter.
- Aider surface (researched): no pre-write hook (confirms no `blocks-writes`);
  reads a read-only conventions/rules file; always-loads it via repo-root
  `.aider.conf.yml` `read: <file>` (YAML). Recommend reusing `AGENTS.md` as the
  rules surface (matches the brief's shared-surface edge case) and emitting
  `.aider.conf.yml` via the existing managed-marker seam to avoid a new YAML
  dependency.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Refactor regresses Claude/OpenCode managed output | High | Medium | Golden-file byte-for-byte parity test added BEFORE refactor; keep opencode parity acceptance test green; adapters wrap existing `plan*` funcs verbatim |
| Aider rules surface guessed wrong (name/path/format) | Medium | Medium | Surface researched; pin AGENTS.md-reuse vs CONVENTIONS.md + YAML emission in the code step before writing planners |
| `.aider.conf.yml` needs structural YAML merge → new dependency | Medium | Medium | Use managed-marker file seam (no parse/merge); unmanaged pre-existing config → manual-review, never clobbered |
| `cmd/` retains business logic the registry should own (G7) | Medium | Medium | Move `isValidAgent` + dispatch into the registry; gatekeeper checks outer-layer purity |
| Capability vocabulary too narrow for codex/local | Low | Low | Keep enum minimal and additive |
| G1 (>100 lines) from folding branches into adapters | Low | Medium | Separate small file per adapter/registry/Aider planner |
| Conflating harness registry with the `Runner` model-routing enum | Medium | Low | Explicit out-of-scope; zero edits to model-routing keys |
| Adding Aider rewrites other harnesses' files / non-idempotent | Medium | Low | Per-adapter `PlanItems`; existing managed markers; acceptance test asserts scope + idempotent re-apply |

#### Rollout
- **Slice 0:** Golden-file parity guard — snapshot today's `.claude/settings.json`,
  `opencode.json`, plugin, `AGENTS.md` bytes + `BuildSyncPlan` for
  `claude`/`opencode`/`both`. Tripwire before any refactor.
- **Slice 1:** Extract `HarnessAdapter` interface + registry; migrate Claude and
  OpenCode as thin wrappers over existing `plan*` funcs; rewrite `BuildSyncPlan`
  to iterate the registry. Slice 0 stays green.
- **Slice 2:** Add `aiderAdapter` (tiered: `prompt-context` + `rules-file`, no
  `blocks-writes`) with `aider_*.go` planners (rules file + `.aider.conf.yml`
  via managed-marker seam). No prewrite hook.
- **Slice 3:** Move `isValidAgent`/dispatch into the registry; wire `--agent
  aider` through init/migrate; invalid `--agent` lists registered harnesses; add
  the capability-parity test + Aider idempotency/scoping acceptance test.
- **Can wait:** Cursor/Windsurf/Copilot/Codex adapters; any `Runner` key;
  capability-vocabulary growth.

#### Deferred Findings
- none. (Every Scope "Out" item is a deliberate, already-known exclusion —
  Cursor/Windsurf/Copilot/Codex are existing roadmap features and the `Runner`
  enum boundary is a pre-stated constraint — not a new discovery.)

#### Handoff
- Next role: feature-specialist
- Outstanding questions:
  1. Aider rules surface: reuse `AGENTS.md` (recommended) or dedicated
     `CONVENTIONS.md`?
  2. `.aider.conf.yml` emission: managed-marker file (recommended, no YAML dep)
     vs structural YAML merge; and which `read:` entries.
  3. Apply path: keep the central `applyItem` `SyncKind` switch or move apply
     behind `HarnessAdapter` — whichever keeps files ≤100 lines and output
     byte-identical.
