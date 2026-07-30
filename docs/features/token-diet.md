# Feature Brief: token-diet

**Archetype:** refactor (plan → code → tests → validate; no docs step)
**Source:** `retrospective.md` §WS3 "Token diet", items 1, 3, 4, 6
**Roadmap entry:** Phase 13 "Lighter Centinela" — *"Drop the O(N) plan-input
glob, fix the UX evidence tag matcher, digest-dedup the per-prompt roadmap
summary, replace dated model pins with family aliases"*

## Problem

Centinela taxes every feature and every user prompt with tokens that buy
nothing. Four independent leaks, all measured on this repo today:

1. **The O(N) plan-input glob.** `orchestration.RequiredPlanInputs` globs
   *every* `docs/features/*.md` in the repo and demands the plan role
   enumerate all of them in its evidence `inputs`. This repo has **120**
   briefs. Every plan-step evidence file therefore carries 120 paths that the
   planner never reads, and the cost *grows with repo age* — feature #200 pays
   for the 199 briefs before it. The stated value (cross-feature awareness) is
   already delivered by the adversarial verifier's conflict analysis, which
   reads the diff, not the brief corpus.
2. **The UX tag matcher demands the same fact twice.**
   `orchestration.normalizeUXTag` normalizes case, `_` and spaces but *not*
   the descriptive suffix, so `"mobile-first: renders at 80x24"` does not
   match the required tag `mobile-first`. The ux-ui-specialist must write the
   bare token *and* the descriptive line — eight required tags, sixteen
   entries, pure ritual.
3. **The roadmap summary re-renders on every single user prompt.**
   `cmd/centinela/hook_context.go` unconditionally prints
   `ui.RenderRoadmapSummary(r)` on every `UserPromptSubmit`. The content is
   identical prompt after prompt (`Roadmap: 43/57 done · 1 in-progress`) and
   is already delivered in full by the `SessionStart` rehydration hook.
4. **Dated/version-pinned model IDs.** `orchestration.tierModels` hardcodes
   `claude-opus-4-7`, `claude-sonnet-4-6`, `claude-haiku-4-5-20251001`. A
   dated snapshot goes stale, and every stale ID is printed into the
   orchestration directive on every hook fire and then acted on by the
   orchestrator, which may spawn a retired model.

The user is (a) the orchestrating agent, which pays items 1, 2 and 4 in
context, and (b) the human operator, who pays item 3 in prompt latency and
context budget across a long session.

**Why now:** the retrospective measured a `high-score`-sized feature at ~700k
subagent tokens under strict. Item 1 alone is ~120 paths × every plan
evidence file, and it is the only leak whose cost is *unbounded* in repo age.

## User Stories

- **US1** — As a planner agent, I want to snapshot only the brief and plan I
  actually read, so that my evidence does not grow by one line for every
  feature the repo has ever shipped.
- **US2** — As a ux-ui-specialist agent, I want `"error-state: shows the
  retry affordance"` to satisfy the `error-state` requirement, so that I state
  each UX fact once instead of twice.
- **US3** — As a human operator in a long session, I want the roadmap summary
  injected when it is new or has changed, not on every prompt, so that
  unchanged state stops consuming my context window.
- **US4** — As an operator whose provider retired a model snapshot, I want the
  built-in tier defaults to be undated family aliases I can remap in
  `centinela.toml`, so that a model refresh is a config line, not a Centinela
  release.
- **US5** — As the owner of an in-flight workflow whose plan evidence already
  enumerates all 120 briefs, I want it to keep validating after the upgrade,
  so that the diet never strands work already in progress.

## Acceptance Criteria

### A — Plan snapshot (retrospective WS3.1)

- **AC1** `orchestration.RequiredPlanInputs(f)` returns exactly two entries —
  `docs/features/<f>.md` and `docs/plans/<f>.md`, normalized and sorted — and
  performs **no filesystem glob**, regardless of how many briefs exist.
- **AC2** Plan-role evidence whose `inputs` are exactly those two paths passes
  `centinela evidence validate` with exit 0.
- **AC3** Plan-role evidence missing either path fails with
  `missing feature-doc snapshot inputs`, naming the missing path(s).
- **AC4 (back-compat, superset)** Plan-role evidence that additionally
  enumerates every other `docs/features/*.md` — the shape an in-flight legacy
  workflow already wrote — still passes. The rule is *"must include own brief
  + plan"*, never *"must include only"*.
- **AC5** `centinela evidence init <f> planner` pre-fills exactly those two
  inputs, and the pre-filled file validates by construction (init and the
  validator keep sharing one function).
- **AC6** The required set is derived by **construction, not existence**: the
  two paths are required even when the brief or plan is not yet on disk.
- **AC7** Path normalization accepts `./docs/features/<f>.md`, backslash
  separators, and a path prefixed by a longer directory chain, for **both**
  the `docs/features/` and the `docs/plans/` entry.
- **AC8** The three architecture docs that state the old rule
  (`evidence-contract.md`, `planner-prompt.md`, `workflow-enforcement.md`) and
  their `internal/scaffold/assets/` mirrors state the new rule and stay
  byte-identical to their mirrors.

### B — UX tag matcher (WS3.3)

- **AC9** `normalizeUXTag` strips everything from the first `:` onward, so
  `"mobile-first: renders at 80x24"` satisfies the `mobile-first` requirement.
- **AC10** A bare token (`"error-state"`) still satisfies its requirement —
  the change strictly loosens; no previously-passing evidence starts failing.
- **AC11** An edgeCase whose prefix matches no required tag still leaves that
  tag missing, and `validateUXEvidence` still names every missing tag.
- **AC12** Degenerate entries — `":"`, `": text"`, `""`, whitespace only —
  normalize to the empty string, match nothing, and never panic.

### C — Per-prompt roadmap summary (WS3.4)

- **AC13** The first `centinela hook context` of a session renders the roadmap
  summary line exactly as today.
- **AC14** A subsequent call in the same session with an unchanged roadmap
  emits **no** roadmap line; every other part of the hook output (directive,
  active-workflow panel, brief/edge-case/changelog nudges) is unchanged.
- **AC15** After `.workflow/roadmap.json` changes in a way that alters the
  summary, the next call renders the summary again.
- **AC16** A different session id re-renders the summary even when the roadmap
  is unchanged.
- **AC17 (fail-open)** Any failure to read, parse or write the digest state —
  and any stdin payload without a session id — results in the summary being
  **rendered**. The hook never errors, never blocks, and is never quieter than
  correct.
- **AC18** The digest state file is git-ignored: `git status` stays clean
  after a hook fire.

### D — Model aliases (WS3.6)

- **AC19** No built-in tier model ID matches a dated-snapshot suffix
  (`-YYYYMMDD`), for any tier or runner.
- **AC20** The `claude` runner's built-in defaults are bare family aliases —
  `opus`, `sonnet`, `haiku` — containing no version digits at all.
- **AC21** `[orchestration.model_map].<tier>.<runner>` still overrides the
  built-in default (resolution precedence 1→2→3→4 is unchanged), so an
  operator remaps a tier without a Centinela release.
- **AC22** Capability classification is a **superset**: both the new aliases
  and every previously-pinned ID still resolve to their class
  (`opus`/`claude-opus-4-7` → frontier, `sonnet`/… → capable,
  `haiku`/`claude-haiku-4-5-20251001` → limited), so no operator's
  `driver_model` silently loses its default profile.
- **AC23** `ModelReference` and the `hook orchestration` directive print the
  alias, and the rule-4 fallback (a runner with no entry for a tier, e.g.
  `codex`) still prints the **tier name**, never another runner's ID.

## Edge Cases

| # | Case | Expected |
|---|------|----------|
| E1 | 120 briefs on disk, planner lists only its own brief + plan | validates (AC1/AC2) |
| E2 | Legacy in-flight evidence lists all 120 + own brief + plan | validates — superset allowed (AC4) |
| E3 | `inputs` lists the brief but not the plan | fails, names `docs/plans/<f>.md` |
| E4 | `inputs` empty | fails, names both required paths |
| E5 | Brief file does not exist on disk yet | still required (construction-based, AC6) |
| E6 | `inputs` entry is `./docs/plans/<f>.md` or uses `\` separators | normalizes and matches (AC7) |
| E7 | Legacy roles `big-thinker` / `feature-specialist` on an unpinned workflow | get the **same** shrunken required set; a legacy workflow's already-written 120-path evidence still passes via E2 |
| E8 | `"mobile-first: renders at 80x24"` | matches `mobile-first` (AC9) |
| E9 | `"loading-state: spinner: with a 3s timeout"` (two colons) | cuts at the **first** colon → `loading-state` |
| E10 | `"error-state:"` (trailing colon, no description) | → `error-state`, matches |
| E11 | `":"` / `": text"` / `""` / `"   "` | → empty tag, matches nothing, no panic (AC12) |
| E12 | Both `"empty-state"` and `"empty-state: …"` present | one tag, still passes (dedup by map key) |
| E13 | `"MOBILE_FIRST : Renders"` (case, underscore, space before colon) | → `mobile-first`, matches |
| E14 | Role is not `ux-ui-specialist` | matcher not consulted at all (unchanged early return) |
| E15 | First hook fire of a session, digest state absent | renders (AC13) |
| E16 | Same session, roadmap byte-identical | suppressed (AC14) |
| E17 | Same session, a feature moved planned → done | renders (AC15) |
| E18 | New session id, roadmap unchanged | renders (AC16) |
| E19 | Digest state file corrupt / truncated / unreadable | renders; state rewritten (AC17) |
| E20 | `.workflow/` not writable | renders; no error surfaced to the host |
| E21 | stdin empty, non-JSON, or JSON without `session_id` | renders (AC17) |
| E22 | Two hook processes fire concurrently | at worst one extra render; never a partial-write corrupt state (atomic temp+rename) |
| E23 | `roadmap.Load()` fails | no roadmap line and no digest write — unchanged from today |
| E24 | Two worktrees of the same repo | independent digest state (per-worktree `.workflow/`) |
| E25 | Operator's `centinela.toml` still names a retired pinned ID | keeps resolving via precedence 1/2 and keeps its capability class (AC22) |
| E26 | `codex` runner, any tier | rule-4 fallback prints the tier name (AC23) |
| E27 | A tier is remapped by `[orchestration.model_map]` | override wins over the alias default (AC21) |

## Data Model

This is a refactor: three of the four slices change pure functions and
constant tables and introduce no persisted state. Slice C introduces one new
runtime record.

**Roadmap summary digest state** — `.workflow/.roadmap-digest` (git-ignored,
per worktree):

| Field | Type | Meaning |
|-------|------|---------|
| `sessionId` | string | Host session id from the hook stdin payload; a change forces a render |
| `digest` | string | Stable hash of the summary-relevant roadmap projection (counts + per-phase feature names and statuses) |

Written atomically (temp file + rename). Absent, corrupt or unwritable state
is not an error — it degrades to "render" (AC17).

**Changed constant tables (no schema change):**
`orchestration.tierModels` (tier → runner → model id) values become family
aliases; `config.builtinModelCapability` (model id → capability class) gains
alias keys while retaining every existing key.

## Integration Points

| Surface | Interaction | Change |
|---------|-------------|--------|
| `centinela evidence init/validate` | shares `RequiredPlanInputs` with the validator | pre-fill shrinks from ~122 to 2 entries |
| `centinela complete` (plan gate) | runs the same evidence validation | accepts both the shrunken and the legacy superset shape |
| `centinela verify` / `verdict` / MCP verdict tool | re-validate plan evidence | inherit the relaxed rule; no per-surface change |
| Claude Code `UserPromptSubmit` hook | stdin JSON payload | now parsed (best-effort) for `session_id`; previously drained and discarded |
| Claude Code `SessionStart` hook | full roadmap rehydration | unchanged — remains the authoritative full render |
| `centinela hook orchestration` | prints per-role model ids + `ModelReference` | prints aliases instead of pinned ids |
| `[orchestration.model_map]`, `[orchestration.models]`, `[orchestration.capabilities]` | model resolution precedence 1→4 | unchanged mechanism; documented as the refresh path |
| `driver_model` → capability class → default profile | `config.DefaultProfileForModel` | superset map keeps every existing id classified |
| `internal/scaffold/assets/docs/architecture/` | byte-identical doc mirror | 3 docs must be re-mirrored |
| `centinela.toml` | quota comment referencing `claude-opus-4-7` | updated to the alias |

## Risks

| # | Risk | Impact | Likelihood | Mitigation |
|---|------|--------|------------|------------|
| R1 | **Self-referential validator.** Slice A changes the very validator that gates *this* feature's own plan evidence. Evidence written now (under the old 120-path rule) is re-validated by `complete` and by claim verification *after* the change lands. | High | High | This is exactly the AC4 superset case — so this feature's own evidence becomes its own regression test. Re-run `centinela evidence validate token-diet` after the code step and again at validate. Never make the rule an equality check. |
| R2 | **G1 overflow.** `internal/config/capability.go` is at **exactly 100** lines and `cmd/centinela/hook_context.go` at **98**. Either slice C or slice D adds a line and trips the 100-line gate. | Medium | Certain | Split both *first*, as the opening move of their slice: extract the capability table to `capability_models.go`, extract the roadmap-render decision to `hook_context_roadmap.go`. Budgets are in the plan. |
| R3 | **Capability-map orphaning.** Changing `tierModels` values without adding alias keys to `builtinModelCapability` silently returns `ok=false` from `DefaultProfileForModel`, disengaging the capability tier and *changing an operator's default enforcement profile* with no error. | High | Medium | AC22 makes the map a strict superset; a table-driven test asserts every `tierModels` value **and** every legacy pinned id classifies. |
| R4 | **Model-id fan-out.** 5 Go test files, 4 `.feature` files and `centinela.toml` hardcode the pinned ids. | Medium | Certain | Enumerated in the plan's slice-D inventory. Prefer asserting the *property* (undated, tier-mapped) over a literal id where the spec allows; update literals otherwise. |
| R5 | **Over-quieting the hook.** Suppressing the summary could hide roadmap state from an agent that joins mid-session. | Low | Medium | `SessionStart` rehydration is untouched and remains the full render; a new session id always re-renders; `centinela roadmap` is always available on demand; fail-open on every uncertainty (AC17). |
| R6 | **Directive output churn.** The orchestration directive's printed model id changes, and hook-output assertions are spread across `cmd/centinela/*_test.go`. | Low | Certain | Slice D runs the full suite before the slice is called done; the directive is dogfooded by observing `centinela hook orchestration` from a locally built binary. |
| R7 | **Scaffold mirror drift.** Editing `docs/architecture/*.md` without re-mirroring breaks the byte-identity parity test — and the parity test only covers a subset of docs, so drift can pass silently. | Medium | Medium | Slice A ends by copying all three edited docs into `internal/scaffold/assets/` and diffing them. |
| R8 | **Untracked digest file dirties the tree.** A new `.workflow/` file that is not git-ignored blocks merges and the roadmap-drift gate. | Medium | Medium | AC18: add the ignore pattern in the same commit as the writer; assert `git status --porcelain` is clean in the acceptance test. |
| R9 | **Weakened UX gate.** After AC9 an entry like `"error-state: n/a"` satisfies the requirement. | Low | High | Accepted and intentional — the retrospective names the double-entry requirement as ritual. Documented in the plan; the gate still fails when the *tag* is absent. |
| R10 | **Parallel-merge semantic conflict.** Other Phase 13 features touch the same hook and config surfaces. | Medium | Medium | Rebase before validate and re-run the full gate on the merged tree, per the repo's standing parallel-merge lesson. |

## Decomposition

Four independently testable slices, shipped as one refactor feature because
each is small and they share the "measure the leak, close it, prove it stays
closed" test pattern. Sequenced smallest-blast-radius-last:

| Slice | Scope | Depends on |
|-------|-------|------------|
| **A** | Plan-snapshot: `RequiredPlanInputs` + path normalization + 3 docs + scaffold mirrors | — |
| **B** | `normalizeUXTag` colon strip | — |
| **C** | Hook roadmap digest-dedup: `internal/roadmap` digest + `cmd` split + gitignore | — |
| **D** | Model family aliases: `tierModels` + capability superset + test/spec/toml fan-out | — |

No slice depends on another; they may land in any order. Slice A is first
because it is the largest token win and because R1 makes it the one that must
be proven against this feature's own evidence.

## Explicitly out of scope

- **WS3.2 "evidence lite"** — one per-step evidence record under `guided` and
  auto-derived `inputs`/`outputs` from `git diff --name-only`. This is a
  change to the *evidence contract shape*, not a leak fix, and the
  retrospective itself flags it as needing coordination with the in-flight
  `lean-evidence-footprint` feature. Separate concern, separate feature.
- **WS3.5 "scaffold slimming"** — shipping only the chosen archetype's guide
  out of `internal/scaffold/assets/docs/architecture/` (~276KB today). This is
  a *distribution-size* concern with its own back-compat question (what
  happens on archetype change after init), not a per-prompt or per-evidence
  token leak. Separate concern, separate feature.
- Changing the enforcement-profile default, the greenfield cascade, or any
  step-gating behavior (WS2).
- Trimming the *other* per-prompt hook panels (active-workflow panel,
  brief/edge-case nudges). Only the roadmap summary is in scope.
