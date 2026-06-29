### Feature-Specialist Report: host-harness-adapters
**Date:** 2026-06-29

#### Behavior Summary

`host-harness-adapters` replaces Centinela's parallel per-harness branching with
a typed `HarnessAdapter` interface backed by a registry. After this feature, every
host harness (Claude, OpenCode, Aider — and future additions) is expressed as one
small registered struct that declares its capabilities and returns a list of
`SyncItem`s; the sync planner, apply switch, and CLI dispatch all read from the
registry instead of hardcoded predicates. Concretely: `BuildSyncPlan(agent)` iterates
the registry to compose a plan, `applyItem()` keeps its central `SyncKind` switch
(thin and byte-identical), and `cmd/centinela` calls registry validation rather than
a hardcoded `isValidAgent` list. Aider is introduced at the tier it genuinely supports:
`prompt-context` + `rules-file` (no `blocks-writes`); it writes `AGENTS.md` (shared
with OpenCode) and `.aider.conf.yml` (a managed-marker file with `read: AGENTS.md`)
idempotently and in scope. Existing Claude and OpenCode behavior is unchanged —
golden-file byte-for-byte parity tests lock this before the refactor begins.

#### Gherkin Scenarios

Full spec: `specs/host-harness-adapters.feature`

**AC1 — Registry lookup + typed error**
- `Registry resolves a known agent name to its adapter` — claude/opencode/aider all resolve.
- `Registry returns a typed error for an unknown agent` — ErrUnknownAgent, lists registered names, no panic.

**AC2 — BuildSyncPlan driven by the registry**
- `BuildSyncPlan for "claude"` — produces only Claude items; no opencode/aider items.
- `BuildSyncPlan for "opencode"` — opencode.json + plugin + AGENTS.md; no claude/.aider items.
- `BuildSyncPlan for "aider"` — AGENTS.md + .aider.conf.yml; no claude/opencode items.
- `BuildSyncPlan for "both"` — union of claude + opencode; no aider items; result equals sum of individual plans.
- `BuildSyncPlan contains no per-harness if-ladder` — registry iteration only.

**AC3 — Capabilities per adapter**
- `Claude adapter declares all three capabilities` — blocks-writes, prompt-context, rules-file.
- `OpenCode adapter declares all three capabilities` — blocks-writes, prompt-context, rules-file.
- `Aider adapter declares prompt-context and rules-file but not blocks-writes`.

**AC4 — Golden-file byte parity**
- `Claude managed output is byte-for-byte identical after refactor`.
- `OpenCode managed output is byte-for-byte identical after refactor`.

**AC5 — Aider init/migrate idempotent and scoped**
- `centinela init --agent aider writes Aider managed files` — AGENTS.md + .aider.conf.yml; claude/opencode untouched.
- `centinela init --agent aider is idempotent on re-run` — no change on second run.
- `centinela migrate --agent aider is idempotent and scoped` — doesn't touch .claude/settings.json.
- `Pre-existing unmanaged .aider.conf.yml is not clobbered` — manual-review warning surfaced.

**AC6 — --agent validation**
- `--agent with a known value is accepted`.
- `--agent with an unknown value lists registered harnesses` — output contains claude, opencode, aider.
- `isValidAgent is resolved by the registry, not a hardcoded list`.

**AC7 — Capability-parity invariant**
- `Every registered adapter declares a non-empty capability set`.
- `Any adapter claiming blocks-writes wires a prewrite hook` — PlanItems includes SyncKindPrewriteHook.
- `Aider does not wire a prewrite hook`.

**Edge case scenarios**
- `AGENTS.md shared surface` — OpenCode + Aider share the file; managed region appears exactly once.
- `Partial existing install` — adding Aider leaves .claude/settings.json untouched.
- `Hook-less harness cannot claim blocks-writes` — capability-parity test enforces the invariant.
- `both selector composes adapters without a special-case branch` — registry definition owns the composition.

#### UX States

| State   | Trigger                                      | Surface                                     |
|---------|----------------------------------------------|---------------------------------------------|
| n/a     | No user-facing UI surface in this feature    | CLI only (init/migrate output messages)     |
| error   | Unknown --agent value                        | CLI stderr: lists registered harness names  |
| warning | Pre-existing unmanaged .aider.conf.yml found | CLI stderr: manual-review warning           |
| success | init/migrate --agent aider completes         | CLI stdout: files written/already up-to-date|

#### Pinned Decisions

**Q1 — Aider rules surface: AGENTS.md (not CONVENTIONS.md)**
Reuse `AGENTS.md` as the shared rules surface. The brief's shared-surface edge case
("OpenCode and Aider both read AGENTS.md") makes this the only clean option — a
separate `CONVENTIONS.md` would require two managed files with identical content and
create an idempotency hazard. Aider's `read:` key accepts any filename.

**Q2 — .aider.conf.yml: managed-marker seam; read: AGENTS.md only**
Emit via the existing `planManagedFile` seam (header comment + minimal content block).
Managed region contains exactly `read: AGENTS.md` — no structural YAML parse/merge,
no YAML library dependency, no inclusion of CLAUDE.md/PROJECT.md (Claude-specific,
not guaranteed present). Pre-existing unmanaged file → manual-review warning, no clobber.

**Q3 — Apply path: keep central applyItem SyncKind switch**
Keep `applyItem()` as the central dispatch hub. Moving apply behind `HarnessAdapter`
would duplicate apply logic (both Claude and OpenCode adapters would reimplement
`applyManagedFile` etc.). The switch stays ≤100 lines with Aider's two new SyncKind
arms; byte-parity is simpler to guarantee from one location; G7 is preserved since
apply lives entirely in `internal/setup/`.

#### Out-of-Scope

- Cursor, Windsurf, Copilot, Codex adapters (separate roadmap features that plug into
  this contract once it lands).
- The orchestration `Runner` model-routing enum (`orchestration_model_map.go`,
  `resolve.go`) — no `aider` runner key added here.
- Expanding the capability vocabulary beyond `{blocks-writes, prompt-context, rules-file}`.
- Structural YAML merge/parse of `.aider.conf.yml`.
- Any UI surface beyond the CLI init/migrate output messages.

#### Deferred Findings

none. All out-of-scope items are pre-agreed exclusions from the brief and big-thinker plan.

#### Handoff
- Next role: senior-engineer
- Open clarifications: none — all three open questions pinned above.
