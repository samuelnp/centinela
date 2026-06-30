# Feature Brief: host-harness-adapters

## Problem

Centinela integrates with host harnesses (the AI coding tools that drive the
agent loop) by hardcoding each one as a parallel branch in `internal/setup/`:
Claude Code lives in `settings.go`/`hooks.go`/`statusline.go` (JSON hooks),
OpenCode lives in the `opencode_*.go` cluster (a JS plugin + `opencode.json` +
`AGENTS.md`). Adding a harness today means new `SyncKind` constants, new
`plan*`/`apply*` branches, a new `applyItem()` switch arm, and new
`use<Harness>()` predicates threaded through `cmd/centinela/init*.go` and
`migrate*.go`. With two harnesses this is tolerable; the roadmap requires
Cursor, Aider, Windsurf, Copilot, and Codex — at which point the per-harness
branching becomes a maintenance tax and the parity guarantee (every harness
gets the same governance) erodes silently.

**Who is hurting:** the Centinela maintainer (every new harness multiplies the
edit surface and the risk of one harness drifting out of parity) and, downstream,
users of any harness Centinela has not yet integrated — they cannot adopt
governance at all.

**Why now:** `codex-support` and `local-harness-support` both depend on this
feature. Building the shared adapter core first means those land as small,
bounded additions instead of repeating the full integration each time.

## User Stories

- As a **Centinela maintainer**, I want to add a new host harness by
  implementing one `HarnessAdapter` and registering it, so that I do not touch
  the sync planner, the apply switch, or the init/migrate commands.
- As a **maintainer**, I want existing Claude and OpenCode behavior to be
  identical after the refactor (parity tests green, byte-for-byte managed
  output), so that generalization carries zero regression.
- As an **Aider user**, I want `centinela init --agent aider` to wire Centinela
  governance into my project, so that I get the same workflow enforcement that
  Claude/OpenCode users already have, at the depth Aider's integration surface
  allows.
- As a **maintainer**, I want each adapter to **declare its capabilities**
  (blocks-writes, prompt-context, rules-file), so that a harness with no
  blocking hook degrades to advisory governance explicitly rather than
  pretending to enforce.

## Acceptance Criteria

(Concrete and testable — these become the Gherkin scenarios.)

1. A `HarnessAdapter` interface exists with a registry; `claude`, `opencode`,
   and `aider` are registered. Looking up an unknown agent returns a typed
   error, not a panic.
2. `BuildSyncPlan(agent)` is driven by the registry (no per-harness `if`
   ladder): for `claude` it plans the Claude settings file; for `opencode` the
   config + plugin + AGENTS.md; for `aider` its rules file + config; for `both`
   it composes Claude + OpenCode exactly as before.
3. Each adapter exposes a `Capabilities()` set drawn from a fixed vocabulary
   (`blocks-writes`, `prompt-context`, `rules-file`). Claude and OpenCode
   declare `blocks-writes` + `prompt-context` + `rules-file`; Aider declares the
   tier it actually supports (`prompt-context` + `rules-file`, no
   `blocks-writes`).
4. The existing Claude + OpenCode parity acceptance tests pass unchanged, and
   the managed files they emit are byte-for-byte identical to pre-refactor
   output (a golden-file assertion).
5. `centinela init --agent aider` and `centinela migrate --agent aider` write
   Aider's managed files idempotently (re-running makes no further change) and
   are scoped: they do not touch Claude or OpenCode files.
6. `--agent` validation accepts the new value(s); an invalid `--agent` lists the
   registered harnesses.
7. A capability-parity test asserts every registered adapter declares a
   non-empty capability set and that any adapter claiming `blocks-writes` wires a
   prewrite hook.

## Edge Cases

- **Unknown / misspelled agent** → typed error listing registered harnesses.
- **`both`** (and any future multi-harness selector) → composition of adapters,
  not a special case in the planner.
- **Idempotent re-apply** → second `init`/`migrate` for the same harness is a
  no-op (managed-region markers respected, like today's sync).
- **Hook-less harness** (Aider/Cursor/Copilot) → must NOT register a
  `blocks-writes` capability or emit a prewrite hook it cannot honor; governance
  is advisory (rules-file + prompt-context) and that degradation is explicit.
- **Partial existing install** (user already has Claude wired, adds Aider) →
  adding one harness leaves the others’ managed files untouched.
- **AGENTS.md shared surface** → OpenCode and Aider both read AGENTS.md;
  emitting it once must not double-write or conflict.
- **File-size G1** → adapter code stays ≤100 lines/file; the registry and each
  adapter are separate small files.

## Data Model

No persisted entities. In-memory types only:

- `HarnessAdapter` (interface): `Name() string`, `Capabilities() []Capability`,
  `PlanItems(agent string) ([]SyncItem, error)` (reusing the existing
  `SyncItem`/`SyncKind`), and validation of its own managed files.
- `Capability` (string enum): `blocks-writes`, `prompt-context`, `rules-file`.
- `registry map[string]HarnessAdapter` keyed by agent name, with an ordered
  lookup for `both`/composition.
- Concrete adapters: `claudeAdapter`, `openCodeAdapter`, `aiderAdapter` —
  thin wrappers over the existing `plan*` functions (Claude/OpenCode) and new
  Aider planners.

## Integration Points

- `internal/setup/` — new `adapter.go` (interface + registry), per-harness
  adapter files, new `aider_*.go` files; `sync.go`/`sync_types.go` refactored to
  drive off the registry.
- `cmd/centinela/init.go`, `init_agent.go`, `migrate.go`, `migrate_setup.go` —
  `--agent` validation and dispatch read from the registry instead of hardcoded
  predicates.
- `internal/orchestration/resolve.go` + `internal/config/orchestration_model_map.go`
  — note only: the model-routing `Runner` enum is **separate** from harness
  integration; adding an `aider` *runner key* for model routing is out of scope
  here (Aider's enforcement integration does not require it). Flagged so the
  two registries are not conflated.
- Aider's actual integration surface (rules file path/format, config) — must be
  researched and pinned in the plan step before coding.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Refactor regresses Claude/OpenCode managed output | High | Medium | Golden-file byte-for-byte parity test added BEFORE refactor; existing parity acceptance tests must stay green |
| Aider's integration surface is weaker/different than assumed | Medium | Medium | Research and pin Aider's rules-file + config contract in the plan; ship only the tier Aider genuinely supports |
| Capability vocabulary is wrong/incomplete (codex/local need more) | Medium | Low | Keep the enum minimal and additive; codex-support/local-harness-support can extend it |
| File-size G1 violations from splitting branches into adapters | Low | Medium | One small file per adapter; registry separate from adapters |
| Conflating harness registry with the orchestration Runner enum | Medium | Low | Explicit out-of-scope note; no edits to model-routing keys in this feature |

## Decomposition

Single feature, sequenced internally (see plan rollout):

1. Extract `HarnessAdapter` interface + registry; migrate Claude + OpenCode
   onto it with a golden-file parity guard (no behavior change).
2. Add the Aider adapter (tiered: `rules-file` + `prompt-context`).
3. Wire `--agent aider` through init/migrate via the registry; capability-parity
   test.

Remaining harnesses (Cursor, Windsurf, Copilot) and Codex are **out of scope**
here — they are separate roadmap features (`codex-support` and the rest) that
plug into the adapter contract this feature establishes.
