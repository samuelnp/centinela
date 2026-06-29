# Plan: host-harness-adapters

> Big-Thinker plan (plan step). Scope is **locked** by the maintainer — this plan
> frames and sequences it; it does not relitigate the boundaries.

## Problem

Centinela wires each host harness as a hand-maintained parallel branch in
`internal/setup/`. Claude Code lives in `settings*.go` / `hooks.go` /
`statusline.go` (JSON hooks in `.claude/settings.json`); OpenCode lives in the
`opencode_*.go` cluster (a JS plugin, `opencode.json`, `AGENTS.md`). The two
branches are stitched together at four seams: `BuildSyncPlan(agent)` (an
`if useClaude / if useOpenCode` ladder), `applyItem()` (a `switch` over the
`SyncKind` enum), the `useClaude()/useOpenCode()` predicates, and the
`cmd/centinela` wiring (`init.go`, `init_agent.go`, `migrate.go`,
`migrate_setup.go`) whose `isValidAgent` and dispatch are hardcoded to
`claude|opencode|both`. Every new harness multiplies that edit surface and
risks one harness silently drifting out of parity. The roadmap (Phase 11)
queues Cursor, Aider, Windsurf, Copilot, and `codex-support` /
`local-harness-support` explicitly depend on a shared core existing first.

The hurt: the maintainer pays a growing per-harness tax, and users of any
not-yet-integrated harness cannot adopt governance at all.

## Scope

### In (v1 — locked)
- Extract a **`HarnessAdapter` interface + registry** in `internal/setup/`.
- Refactor **both existing harnesses** (Claude, OpenCode) onto the interface
  with **zero behavior change**, guarded by a **golden-file byte-for-byte
  parity test** added BEFORE the refactor.
- Add **exactly one** new harness: **Aider**.
- A **tiered capability model**: each adapter declares capabilities from the
  fixed vocabulary `{blocks-writes, prompt-context, rules-file}`. Claude and
  OpenCode declare all three; **Aider declares `prompt-context` + `rules-file`
  only — NO `blocks-writes`** (Aider has no blocking prewrite hook).
- Drive `BuildSyncPlan(agent)` / `applyItem()` and the `cmd` `--agent`
  validation/dispatch off the registry (no per-harness `if`/`switch` ladder).
- `centinela init --agent aider` and `centinela migrate --agent aider` write
  Aider's managed files idempotently and scoped (no touching Claude/OpenCode
  files); invalid `--agent` lists the registered harnesses.
- A **capability-parity test**: every registered adapter declares a non-empty
  capability set, and any adapter claiming `blocks-writes` wires a prewrite
  hook.

### Out (deliberate, not new discoveries)
- **Cursor, Windsurf, Copilot, Codex** — separate roadmap features that plug
  into this contract (`codex-support` and the rest). No code for them here.
- The orchestration **`Runner` enum** (model routing in
  `internal/orchestration/resolve.go` +
  `internal/config/orchestration_model_map.go`). **Do NOT add an `aider`
  runner key** — harness integration and model routing are separate registries
  and must not be conflated.
- Expanding the capability vocabulary beyond the three values (codex /
  local-harness can extend it additively later).

## Dependencies & Assumptions

- **Internal modules touched:** `internal/setup/` (new `adapter.go` registry +
  per-harness adapter files + new `aider_*.go`; `sync.go` / `sync_types.go` /
  `sync_hooks.go` / `sync_managed_files.go` refactored to drive off the
  registry) and `cmd/centinela/` (`init.go`, `init_agent.go`, `migrate.go`,
  `migrate_setup.go` read `isValidAgent`/dispatch from the registry).
- **Reuse, do not rewrite:** the existing `plan*`/`apply*`/`build*` functions
  (`planHooksSettings`, `planOpenCodeConfig`, `planPluginFile`,
  `planAgentsFile`, `InjectHooks`, `InjectOpenCodeConfig`, `writeManaged*`) and
  the `SyncItem`/`SyncKind`/`SyncPlan` types stay; adapters are **thin wrappers**
  over them so managed output is provably unchanged.
- **Layer rules (PROJECT.md):** `internal/setup` is Infrastructure; `cmd/` is
  the thin outer layer (G7) — `isValidAgent` and dispatch decisions move INTO
  `internal/setup`'s registry, `cmd/` only calls it. No new cross-layer edges.
- **G1:** ≤100 lines/file — the registry, each adapter, and each Aider planner
  are separate small files.
- **Aider integration surface (researched, pin in code-step):**
  - Aider has **no pre-write hook** → confirms no `blocks-writes`.
  - **rules-file:** Aider reads a read-only conventions/rules file. Centinela
    already emits `AGENTS.md`; the brief's "AGENTS.md shared surface" edge case
    indicates Aider should **reuse `AGENTS.md`** as its rules surface rather
    than emit a second file (idiomatic Aider name is `CONVENTIONS.md` — the
    feature-specialist must pin: reuse `AGENTS.md` [recommended, single surface,
    matches the edge case] vs. a dedicated `CONVENTIONS.md`).
  - **prompt-context:** Aider always-loads the rules file via the repo-root
    config `.aider.conf.yml` key `read: <file>` (single string or list). This
    is YAML.
- **Assumption — no new YAML dependency:** `.aider.conf.yml` is emitted via the
  existing `planManagedFile` managed-marker seam (header comment + a minimal
  `read:` block; absent→create, managed→update, unmanaged→manual-review),
  matching the plugin/AGENTS.md pattern. This avoids a structural YAML
  merge/dependency in v1. (A YAML lib appears only transitively in `go.sum`;
  adding a direct YAML merge is out of scope and a risk to avoid.)

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Refactor regresses Claude/OpenCode managed output | High | Medium | Add a golden-file byte-for-byte parity test capturing today's `.claude/settings.json`, `opencode.json`, plugin, `AGENTS.md` output **BEFORE** refactoring; keep `tests/acceptance/opencode_hook_parity_test.go` green; adapters wrap existing `plan*` funcs verbatim |
| Aider rules surface guessed wrong (file name/path/format) | Medium | Medium | Surface researched (`.aider.conf.yml` `read:` + conventions file); pin AGENTS.md-reuse vs CONVENTIONS.md and the YAML emission strategy in the code step before writing planners |
| `.aider.conf.yml` YAML needs structural merge → new dependency | Medium | Medium | Use the managed-marker file seam (no parse/merge); unmanaged pre-existing config → `manual-review`, never clobbered |
| `cmd/` keeps business logic when registry should own it (G7) | Medium | Medium | Move `isValidAgent` + agent→adapter dispatch into the registry; `cmd/` only renders; gatekeeper checks outer-layer purity |
| Capability vocabulary too narrow for codex/local | Low | Low | Keep the enum minimal and additive; downstream features extend it |
| G1 (>100 lines) from folding branches into adapters | Low | Medium | One small file per adapter + separate registry file + separate Aider planners |
| Conflating harness registry with the `Runner` model-routing enum | Medium | Low | Explicit out-of-scope note; zero edits to `orchestration_model_map.go` / `resolve.go` |
| Idempotency/scoping break: adding Aider rewrites others' files | Medium | Low | Per-adapter `PlanItems`; selecting `aider` plans only Aider items; idempotent re-apply via existing managed markers; acceptance test asserts scope |

## Rollout — smallest correct slices

**Slice 0 — Parity guard (safety net, no production code change).**
Add a golden-file test that snapshots the bytes of every managed file Claude +
OpenCode emit today (`.claude/settings.json`, `opencode.json`,
`.opencode/plugins/centinela.js`, `AGENTS.md`) and asserts the `BuildSyncPlan`
results for `claude`/`opencode`/`both`. This locks behavior before any
refactor and is the regression tripwire for Slices 1–2.

**Slice 1 — Extract interface + registry; migrate Claude & OpenCode (zero
behavior change).** Add `adapter.go` (`HarnessAdapter` interface:
`Name()`, `Capabilities()`, `PlanItems(agent)`; a `Capability` string enum; a
registry with ordered lookup for `both`). Add `claudeAdapter` and
`openCodeAdapter` as thin wrappers over the existing `plan*` functions. Rewrite
`BuildSyncPlan` to iterate the registry instead of the `if` ladder; keep
`applyItem`'s `SyncKind` switch (or fold per-adapter apply behind the
interface — whichever keeps files ≤100 lines and output byte-identical).
Slice 0 must stay green.

**Slice 2 — Add the Aider adapter (tiered).** Research-pinned: `aiderAdapter`
declares `prompt-context` + `rules-file` (no `blocks-writes`), with new
`aider_*.go` planners for its rules file (AGENTS.md reuse or CONVENTIONS.md, per
pin) and `.aider.conf.yml` (`read:` entry) via the managed-marker seam. Register
it. No prewrite hook emitted.

**Slice 3 — Wire `--agent aider` through the CLI + capability-parity test.**
Move `isValidAgent` and agent→adapter dispatch into the registry; `init.go` /
`init_agent.go` / `migrate.go` / `migrate_setup.go` call the registry; invalid
`--agent` lists registered harnesses. Add the capability-parity test (every
adapter has a non-empty capability set; `blocks-writes` ⇒ a prewrite hook is
wired). Idempotency + scoping acceptance test for `aider`.

**Can wait (out of scope):** Cursor / Windsurf / Copilot / Codex adapters; any
`Runner` model-routing key; capability-vocabulary growth.

## Open questions for the feature-specialist
1. Aider rules surface: reuse `AGENTS.md` (recommended — single surface, matches
   the brief's shared-surface edge case) or emit a dedicated `CONVENTIONS.md`?
2. `.aider.conf.yml` emission: managed-marker file (recommended, no YAML dep) vs.
   structural YAML merge — and exactly which `read:` entries (AGENTS.md only, or
   also CLAUDE.md / PROJECT.md)?
3. Apply path: keep the central `applyItem` `SyncKind` switch, or move apply
   behind `HarnessAdapter` — pick the one that stays ≤100 lines/file and keeps
   golden output byte-identical.

## Pinned Decisions

### Q1 — Aider rules surface: reuse AGENTS.md

**Decision:** Reuse `AGENTS.md` as Aider's rules surface. Do NOT emit a separate
`CONVENTIONS.md`.

**Rationale:**
- The feature brief's "AGENTS.md shared surface" edge case is explicit: "OpenCode
  and Aider both read AGENTS.md; emitting it once must not double-write or
  conflict." This edge case is only solvable cleanly if both adapters target the
  same file.
- Aider's `read:` key in `.aider.conf.yml` can point to any filename; there is
  no technical requirement for it to be named `CONVENTIONS.md`. `AGENTS.md` is
  Aider-compatible.
- A dedicated `CONVENTIONS.md` would mean two managed files with identical
  governance content — a duplication hazard and an idempotency problem whenever
  the centinela-managed content drifts between them.
- Using a single file keeps the shared-surface idempotency test simple: one
  managed region, one file, two adapters that both declare they consume it.

**Spec impact:** Scenarios for AC5 and the AGENTS.md shared-surface edge case
reflect `AGENTS.md` as the single rules file for both OpenCode and Aider.

---

### Q2 — .aider.conf.yml emission: managed-marker file seam; read: AGENTS.md only

**Decision:** Emit `.aider.conf.yml` using the existing `planManagedFile`
managed-marker seam (header comment + minimal content block). The managed region
contains exactly one `read:` entry pointing to `AGENTS.md`. Do not read-in
`CLAUDE.md` or `PROJECT.md` here.

**Rationale:**
- The managed-marker seam is already used for the OpenCode plugin and `AGENTS.md`;
  it handles absent→create, managed→update, and unmanaged→manual-review without
  parsing YAML. Adding a structural YAML merge would introduce a direct YAML
  dependency in v1 and a new risk vector (conflicting keys, ordering, comments
  stripped). The plan and big-thinker both flag this as a risk to avoid.
- `read: AGENTS.md` is the minimal correct entry: Aider loads it at session
  start and the file carries the Centinela governance prompt. `CLAUDE.md` and
  `PROJECT.md` are Claude-specific conventions not guaranteed to be present in
  all Aider projects; including them would widen scope and risk errors on
  projects without those files.
- If a pre-existing unmanaged `.aider.conf.yml` already exists, the seam
  surfaces a `manual-review` warning and does not clobber it — same behavior
  as the OpenCode plugin today.

**Exact content of managed region:**
```yaml
# centinela:managed-start
read: AGENTS.md
# centinela:managed-end
```

**Spec impact:** AC5 scenarios assert `.aider.conf.yml` contains a managed
region with `read: AGENTS.md` and that a pre-existing unmanaged file is not
overwritten.

---

### Q3 — Apply path: keep the central applyItem SyncKind switch

**Decision:** Keep the central `applyItem()` `SyncKind` switch in
`internal/setup/`. Do NOT move apply behind the `HarnessAdapter` interface.

**Rationale:**
- The `SyncKind` enum drives dispatch for `applyItem()`; `SyncItem` already
  carries the `Kind` discriminator. Keeping a central switch means the apply
  logic has one location per kind — exactly one place to audit for parity.
- Moving apply behind `HarnessAdapter` would force each adapter to implement
  apply methods for every `SyncKind` it produces, duplicating logic (e.g., both
  Claude and OpenCode adapters would need to implement `applyManagedFile` — the
  exact duplication the interface was meant to eliminate).
- G1 (≤100 lines/file) is satisfied by the existing structure: `applyItem()` is
  the dispatch hub, and the actual apply functions (`InjectHooks`,
  `writeManaged*`, etc.) live in dedicated small files. Aider adds at most two
  new `SyncKind` values (one for `AGENTS.md` update, one for `.aider.conf.yml`),
  each handled by a one- or two-line case arm. The switch stays well within 100
  lines.
- G7 (cmd/ thin outer layer): the apply path lives entirely in `internal/setup/`,
  so `cmd/` only calls `applyItem()` — no business logic leaks outward.
- Golden-file byte parity is simpler to guarantee with a shared apply path:
  the same `InjectHooks` / `writeManaged*` functions are called for Claude and
  OpenCode whether the planner is registry-driven or not.

**Spec impact:** AC4 golden-file scenarios validate byte parity, implicitly
confirming that the central apply path is not regressed by the refactor.
