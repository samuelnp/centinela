# token-diet — planner

**Date:** 2026-07-29
**Role:** planner — first feature planned under the merged `planner-v1`
contract (one agent, both lenses: strategy, then spec)
**Artifacts:** [docs/features/token-diet.md](../docs/features/token-diet.md) ·
[docs/plans/token-diet.md](../docs/plans/token-diet.md) ·
[specs/token-diet.feature](../specs/token-diet.feature)

## Problem

Centinela taxes every feature and every user prompt with tokens that buy
nothing, and one of those taxes grows without bound. `RequiredPlanInputs`
globs every `docs/features/*.md` and demands the plan role enumerate all of
them in its evidence `inputs` — **120 briefs today**, so every plan-step
evidence file carries 122 paths the planner never read, and feature #200 will
pay for the 199 before it. Alongside it sit three fixed-size leaks:
`normalizeUXTag` does not strip a descriptive suffix, so the ux-ui-specialist
must state each of eight UX facts twice (bare token *and* descriptive line);
`hook_context.go` re-emits an identical roadmap summary on **every**
`UserPromptSubmit` even though `SessionStart` already delivered it in full;
and `tierModels` hardcodes version-pinned and dated model ids
(`claude-haiku-4-5-20251001`) that go stale and are then printed into the
orchestration directive and acted on. The people hurting are the orchestrating
agent (items 1, 2, 4 — paid in context on every feature) and the human
operator (item 3 — paid in context on every prompt of a long session). Why
now: the retrospective measured a `high-score`-sized feature at ~700k subagent
tokens, and item 1 is the only leak whose cost is unbounded in repo age, so
every feature shipped from here makes it worse.

## Scope

**In (v1):**
- Kill the O(N) plan-input glob — `RequiredPlanInputs` returns the feature's
  own brief + own plan, construction-derived, no filesystem glob.
- Keep the snapshot rule **include, never only** — a superset validates, so
  legacy in-flight evidence (and this feature's own) survives.
- Make path normalization symmetric across `docs/features/` and `docs/plans/`.
- `normalizeUXTag` cuts at the first `:` so a descriptive entry matches.
- Digest-dedup the per-prompt roadmap summary, keyed on host session id plus a
  projection hash, failing open on every uncertainty.
- Replace the built-in `claude`-runner pins with bare family aliases
  (`opus`/`sonnet`/`haiku`); keep the capability id→class map a strict
  superset so no operator's `driver_model` loses its default profile.
- Update the three architecture docs that state the old snapshot rule and
  re-mirror them into `internal/scaffold/assets/`.

**Out (v1):**
- **WS3.2 evidence-lite** (one per-step record under `guided`, `inputs`
  auto-derived from `git diff --name-only`) — a change to the evidence
  *contract shape*, not a leak fix; the retrospective itself flags it as
  needing coordination with the in-flight `lean-evidence-footprint` feature.
  Pre-agreed exclusion, not a new finding.
- **WS3.5 scaffold slimming** (ship only the chosen archetype's guide out of
  the ~276KB asset tree) — a distribution-size concern with its own
  back-compat question (what happens on archetype change after `init`).
  Pre-agreed exclusion, not a new finding.
- WS2 default-profile / greenfield-cascade flips; any step-gating change.
- The *other* per-prompt hook panels — this one **is** a new discovery; see
  Deferred Findings.

## Dependencies & Assumptions

- Builds directly on `unified-plan-specialist` (`planner-v1`): the shrunken
  snapshot rule must apply to `RolePlanner` **and** the two retired legacy
  roles, so an unpinned in-flight workflow is treated identically.
- `internal/evidence.PlanInputs` and the validator already share
  `orchestration.RequiredPlanInputs`. That sharing is what makes AC5 ("a
  pre-filled init validates by construction") free — preserve it.
- `internal/roadmap` is an already-mapped domain package, so the new digest
  seam needs **no** `import_graph` layer entry and **no** PROJECT.md G2 edit.
- Model-resolution precedence (role override → tier map → built-in → tier
  name) already exists and already gives operators a no-release override path.
  Slice D changes the *defaults*, not the mechanism.
- Assumption, holding for the Claude Code runner: bare `opus`/`sonnet`/`haiku`
  are valid model references. **Not** assumed for opencode — that column keeps
  a provider-qualified id rather than an invented "latest" form.
- Assumption: the `UserPromptSubmit` payload carries `session_id`. Every code
  path treats its absence as "render", so the assumption being wrong costs
  nothing but the saving.
- `internal/orchestration` cannot import `internal/config` (config is the leaf
  and imports nothing internal), so the "every tier model has a capability
  class" assertion must live in `internal/config` or `tests/unit/`.
- Historical `.workflow/*-<role>.md` files mentioning retired model ids are
  frozen evidence records and must not be rewritten by the slice-D sweep.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **Self-referential validator** — slice A rewrites the very rule that gates *this* feature's plan evidence, written under the old 122-input rule and re-validated by `complete`, `verify` and the MCP verdict tool after the change lands | High | High | Exactly the AC4 superset case, so this feature's own evidence is its own regression test. Keep `validatePlanSnapshotInputs` reporting only *missing* entries — never "tidy" it into a set-equality check. Re-run `centinela evidence validate token-diet` after the code step and again at validate |
| **G1 overflow** — `internal/config/capability.go` is at **exactly 100** lines, `cmd/centinela/hook_context.go` at 98; the gate trips on the first line added, and diff-aware `validate` hides it until CI's full scan | Medium | Certain | Each of those slices *opens* with a split (`capability_models.go`, `hook_context_roadmap.go`) before any behavior change; per-file budgets tabulated in the plan |
| **Capability-map orphaning** — changing `tierModels` values without adding alias keys makes `DefaultProfileForModel` return `ok=false`, silently disengaging the capability tier and changing an operator's default enforcement profile with **no error** | High | Medium | AC22 makes the map a strict superset (aliases added, every legacy pin retained); a table-driven test asserts both sets classify |
| **Model-id fan-out** — 5 Go test files, 4 `.feature` files and a `centinela.toml` comment hardcode the pins | Medium | Certain | Enumerated file-by-file in the plan's slice-D inventory; prefer asserting the tier→alias *property* over re-pinning a new literal |
| **Over-quieting the hook** — suppression could hide roadmap state from an agent joining mid-session | Low | Medium | `SessionStart` rehydration untouched and remains the authoritative full render; a new session id always re-renders; `centinela roadmap` is always on demand; fail-open everywhere |
| **Scaffold mirror drift** — three edited architecture docs must be re-mirrored, and the parity test's allowlist already excludes `workflow-enforcement.md`, so drift there passes silently | Medium | Medium | Slice A ends with an explicit three-way `diff`; the pre-existing allowlist gap is already tracked as `mirror-workflow-enforcement-doc` |
| **Untracked digest file dirties the tree** — a new `.workflow/` file outside the allow-list blocks merges and the roadmap-drift gate | Medium | Medium | AC18: the gitignore line lands in the same commit as the writer; the acceptance test asserts `git status --porcelain` is clean after a hook fire |
| **Weakened UX gate** — after AC9, `"error-state: n/a"` satisfies the requirement | Low | High | Accepted and intentional; the retrospective names the double-entry rule as ritual. The gate still fails when the *tag* is absent (AC11) |
| **Parallel-merge semantic conflict** — sibling Phase 13 features touch the same hook and config surfaces | Medium | Medium | Rebase before validate, re-run the full gate on the merged tree, re-append any new `docs/features/*.md` to plan evidence after the rebase |

## Rollout

Four slices, none depending on another, sequenced by blast radius and by which
must be proven earliest:

- **Step 1 — Slice A (plan snapshot).** Biggest and only unbounded win, and the
  one entangled with this feature's own evidence (R1) — prove it while the
  workflow is still early. `RequiredPlanInputs` → 2 entries, symmetric path
  normalization, 3 docs + 3 scaffold mirrors.
- **Step 2 — Slice B (UX tag).** Smallest diff, zero coupling, one commit.
- **Step 3 — Slice C (hook digest).** New persisted state and a new stdin
  parse; needs the `cmd` split and the gitignore line in the same commit.
  Domain logic lands in `internal/roadmap` (G7); `cmd` stays thin wiring.
- **Step 4 — Slice D (model aliases).** Widest textual fan-out (10 files) and
  the silent capability risk; last, so the full suite runs against an
  otherwise-settled tree.

After every slice: `go test ./... -run xxxNONE` (surfaces test-compile breaks
that `go build` hides), then the full `go test ./...`.

**Can wait:** quieting the remaining hook panels (deferred below); rewriting
the four foreign specs to assert model *properties* rather than literals.

## Behavior Summary

Nothing user-facing changes shape. A plan role's evidence now needs to list
only the feature's own brief and its own plan, and extra inputs are still
accepted, so evidence written under the old rule keeps validating — the rule
became *include*, not *only*. A ux-ui-specialist's descriptive edge case
(`"mobile-first: renders at 80x24"`) now satisfies the `mobile-first`
requirement because the matcher cuts at the first colon; bare tokens keep
working, and a genuinely absent tag is still reported missing. The roadmap
summary is injected on the first prompt of a session and thereafter only when
the roadmap's projection actually changed, with every uncertainty — no session
id, corrupt state, unwritable `.workflow/` — resolving to "render". The
built-in `claude`-tier defaults become `opus`/`sonnet`/`haiku`, printed as such
in the orchestration directive, remappable per tier in `centinela.toml`, with
every previously-pinned id still classified so no operator's default
enforcement profile shifts underneath them.

## Acceptance Criteria (Gherkin)

Full spec: [specs/token-diet.feature](../specs/token-diet.feature) — 30
scenarios across four sections; every acceptance criterion carries at least one
happy and one negative path. Anchors:

- **Happy, A:** *Given* a repository containing 120 files under
  `docs/features/`, *When* the required plan inputs are computed for
  `token-diet`, *Then* the set contains exactly `docs/features/token-diet.md`
  and `docs/plans/token-diet.md`. Paired with the no-glob invariant: the same
  set is returned from a repository with **zero** briefs.
- **Negative, A:** *Given* evidence listing the brief but not the plan, *Then*
  validation fails with `missing feature-doc snapshot inputs` naming
  `docs/plans/token-diet.md`.
- **Back-compat, A:** *Given* an in-flight workflow whose evidence lists all
  120 briefs plus its own brief and plan, *Then* validation passes — include,
  not only.
- **Happy, B:** *Given* edge cases that are all descriptive
  (`"mobile-first: renders at 80x24"` …), *Then* UX validation passes.
  **Negative, B:** *Given* seven tags covered descriptively and `error-state`
  omitted, *Then* validation fails naming only `error-state`.
- **Happy, C:** *Given* no digest state and session `s-1`, *Then* the summary
  prints. **Quiet, C:** *Given* state recording `s-1` and the current digest,
  *Then* no summary line prints and every other panel is unchanged. A six-row
  fail-open outline covers absent/corrupt/unreadable state and
  empty/non-JSON/session-less payloads — all render, all exit 0.
- **Happy, D:** *Given* the built-in table, *Then* no id ends in a `-YYYYMMDD`
  suffix and the claude column is `opus`/`sonnet`/`haiku` with no digits.
  **Negative, D:** *When* resolving for runner `codex`, *Then* the value is the
  tier name `reasoning` with not-ok — never another runner's id. A nine-row
  outline pins the capability superset (aliases **and** legacy pins).

## UX States

n/a — no UI surface. The only rendered output that changes is CLI/hook text:

| State | Trigger | Surface |
|-------|---------|---------|
| loading | n/a — every path is synchronous file I/O | — |
| empty | roadmap absent or invalid; no active workflow | no roadmap line, no digest write (unchanged from today) |
| quiet (new) | same session, unchanged roadmap projection | roadmap summary suppressed; all other hook panels unchanged |
| error | corrupt/unreadable/unwritable digest state, unparseable stdin | fails **open**: summary renders, hook exits 0, nothing surfaced to the host session |
| success | first prompt of a session, or the roadmap changed | current one-line `ROADMAP  Roadmap: N/M done · K in-progress` |

## Edge Cases

30 enumerated in the JSON companion's `edgeCases` and in the brief's table.
Grouped:

- **A (E1–E7):** 120 briefs vs. own-two-only; legacy 120-path superset;
  brief-present/plan-missing; empty inputs; brief not yet on disk (required by
  construction, not existence); `./`-prefixed, backslash and long-absolute
  paths on **both** the features and plans side; retired legacy roles get the
  same shrunken set.
- **B (E8–E14):** descriptive suffix matches; two colons cut at the first;
  trailing colon with empty description; degenerate `":"` / `": text"` / `""` /
  whitespace normalize to empty, match nothing, never panic; bare + descriptive
  duplicate dedups; non-UX role never consults the matcher.
- **C (E15–E24, E28):** first fire renders; unchanged projection suppresses;
  status change re-renders; new session re-renders; corrupt/unreadable state,
  unwritable `.workflow/`, empty/non-JSON/session-less stdin all render and
  exit 0; `roadmap.Load()` failure writes nothing; concurrent fires cost at
  most an extra render (temp+rename); two worktrees keep independent state;
  whitespace-only reformat of `roadmap.json` does **not** re-render.
- **D (E25–E27):** a retired pin in an operator's config keeps its class and
  profile; `codex` falls back to the tier name; `model_map` remap beats the
  alias.
- **Meta (E29, E30):** this feature's own 122-input plan evidence must still
  validate after slice A; `capability.go` at exactly 100 lines and
  `hook_context.go` at 98 mean the G1 gate trips on the first added line.

## Out-of-Scope

- WS3.2 evidence-lite (per-step records under `guided`, auto-derived
  `inputs`/`outputs`) — pre-agreed exclusion; coordinate with
  `lean-evidence-footprint`.
- WS3.5 scaffold slimming (archetype-specific asset shipping) — pre-agreed
  exclusion.
- WS2 default-profile flip, greenfield-cascade slimming, self-graded
  quality-threshold removal.
- Trimming the remaining per-prompt hook panels — **new finding, deferred**.
- Rewriting the four foreign specs that pin model ids to assert properties
  instead of literals; slice D updates the literals where a property assertion
  does not fit the scenario's point.
- Any change to what `SessionStart` rehydration prints.

## Deferred Findings

- **`hook-context-panel-diet`** — *new discovery, recorded.* Measuring the
  per-prompt hook before planning slice C showed it emits **23 lines / 1538
  bytes**, of which the roadmap summary this feature quiets is **1 line /
  ~56 bytes (~4%)**. The `ACTIVE WORKFLOWS` panel and the
  brief/edge-case/changelog nudge panels are the remaining ~96% and re-render
  unchanged on every prompt. The roadmap entry scopes this feature to the
  roadmap summary only, so extending the same session+digest seam to those
  panels is deferred. This also corrects the retrospective's framing: the
  renderer is *already* one line, so the only saving available here is
  suppression — the seam, not the bytes, is slice C's real deliverable.
  `centinela roadmap defer hook-context-panel-diet --source token-diet/planner`

Checked and **not** deferred, because already tracked or already in scope:
`mirror-workflow-enforcement-doc` (existing backlog item — the parity-allowlist
gap this feature's doc edits brush against); the `docs/plans/`-side path
normalization asymmetry (pulled **into** scope as slice A2 rather than
deferred, because with only two required entries a normalization miss is half
the rule).

## Handoff

- **Next role:** `senior-engineer`.
- **Outstanding questions / clarifications:**
  1. **Opencode alias form.** The plan deliberately keeps
     `anthropic/claude-<family>-<version>` for the opencode column rather than
     inventing a `-latest` form. If the engineer can *verify* an undated
     opencode reference against the provider, upgrading that column is in the
     spirit of AC19 — but do not guess.
  2. **Foreign spec edits.** Slice D touches four `.feature` files owned by
     other features. Where a scenario's point is "an override wins", rewrite it
     generically; where it genuinely pins an id, update the literal. Flag any
     scenario where neither reads cleanly rather than forcing it.
  3. **Do not weaken these three invariants** while refactoring: the snapshot
     rule stays a superset check (R1); the capability map stays a superset
     (R3); slice C fails open on every uncertainty.
