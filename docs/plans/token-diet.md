# Plan: token-diet

**Brief:** [docs/features/token-diet.md](../features/token-diet.md)
**Spec:** [specs/token-diet.feature](../../specs/token-diet.feature)
**Archetype:** refactor — plan → code → tests → validate (no docs step)
**Plan contract:** `planner-v1` (first feature planned under the merged
planner role: one agent, two lenses)

## 0. What this plan is

Four leak fixes, four slices, no new user-facing surface. Every slice follows
the same shape: **measure the leak → close it → pin it shut with a test that
fails if it reopens**. The plan is written for a code agent that will be given
this file and the spec, and nothing else.

Two facts dominate the sequencing and must be read before touching anything:

- **`internal/config/capability.go` is exactly 100 lines. `cmd/centinela/hook_context.go` is 98.** Both are targets. G1 (≤100 lines, `_test.go` included) trips on the first line added. Each of those slices therefore *opens* with a split, not with the behavior change.
- **Slice A changes the validator that gates this feature's own plan evidence.** The evidence written during *this* plan step lists ~122 inputs (the old rule). After slice A lands, `centinela complete`, `centinela verify` and the MCP verdict tool all re-validate that same file under the new rule. It must still pass. That is precisely AC4 (superset), so this feature's own evidence is the acceptance test. **The rule must be "inputs must *include* the required set", never "must equal".** The existing validator already has this shape (it only reports *missing* entries) — do not "tidy" it into an equality check.

## 1. Measured baseline

Record these numbers before changing anything; the tests assert against them.

| Leak | Measurement command | Today |
|------|--------------------|-------|
| Plan glob | `ls docs/features/*.md \| wc -l` | 120 briefs → 122 required inputs per plan evidence |
| UX tags | `internal/orchestration/evidence_ux.go` `requiredUXTags` | 8 tags, each needing a bare token *and* a descriptive line |
| Per-prompt hook | `centinela hook context < /dev/null \| wc -lc` | 23 lines / 1538 bytes, of which the roadmap line is 1 line / ~56 bytes, re-emitted on **every** prompt |
| Model pins | `internal/orchestration/resolve.go` `tierModels` | 6 ids, 1 dated snapshot (`claude-haiku-4-5-20251001`), 5 version-pinned |

**Honest note on slice C.** `ui.RenderRoadmapSummary` already returns exactly
one line. The retrospective's phrasing ("full render on change, otherwise one
line") describes a state the renderer is already in; the only saving left is
**suppression** on no-change. Slice C therefore renders the current one-line
summary on change / session start, and emits **nothing** otherwise. The saving
is modest per prompt (~56 bytes) but is the only one that scales with session
length; it is included because it also establishes the digest-state seam that
a later feature can reuse to quiet the larger panels (explicitly out of scope
here).

---

## 2. Slice A — kill the O(N) plan snapshot

**Delivers:** AC1–AC8, E1–E7. **Blast radius:** the plan-step evidence gate on
every feature, including this one (R1).

### A1. Shrink the required set

`internal/orchestration/plan_snapshot.go`

- Delete the `filepath.Glob("docs/features/*.md")` fan-out and the `seen` map
  it needed. `RequiredPlanInputs(feature)` returns exactly the two
  construction-derived paths, normalized and sorted:
  `docs/features/<feature>.md`, `docs/plans/<feature>.md`.
- Drop the now-unused `path/filepath` glob usage and the `sort` import only if
  they become unused — a 2-element sort is still cheaper to keep than to
  reason about ordering by hand; keep `sort.Strings` for determinism.
- `validatePlanSnapshotInputs` is **unchanged in shape**: build the provided
  set, report only the *missing* required entries. Superset stays legal (AC4).
- `requiresPlanSnapshot` is unchanged — planner + the two retired legacy roles
  all get the shrunken set (E7).

### A2. Symmetric path normalization

`internal/orchestration/plan_snapshot.go` — `normalizeFeatureDocPath`

Today the function anchors only on `docs/features/`, so a `docs/plans/` input
written as an absolute or deeply-prefixed path never normalizes and can never
match. With 122 required entries a normalization miss was one failure among
many; with **two**, it is half the rule. Generalize the anchor to the first
occurrence of either `docs/features/` or `docs/plans/`, keeping the existing
backslash→slash, `./`-strip and `filepath.Clean` behavior (AC7, E6).

This change only ever makes previously-rejected inputs match — it cannot break
existing evidence.

### A3. Documentation + scaffold mirror

Update the three docs that state the old rule, then re-mirror:

| Doc | Line stating the rule |
|-----|----------------------|
| `docs/architecture/evidence-contract.md` | ~56 (per-role rules) and ~190 (glossary/summary) |
| `docs/architecture/planner-prompt.md` | ~117 (role rules block) |
| `docs/architecture/workflow-enforcement.md` | ~150 |

New wording, used verbatim in all three:

> `inputs` MUST include the feature's own brief `docs/features/<feature>.md`
> and its plan `docs/plans/<feature>.md`. Additional inputs are allowed — the
> rule is *include*, not *only* — so evidence written before this rule shrank
> still validates.

Then copy each edited file over its `internal/scaffold/assets/docs/architecture/`
twin and verify byte identity (`diff` must be silent). All three twins exist
today (R7).

### A4. Tests (slice A)

| File | New/changed | Budget |
|------|-------------|--------|
| `internal/orchestration/plan_snapshot_test.go` | colocated unit: exact-2 return, no-glob invariance (create a temp brief, assert the set does not grow), superset accept, each-missing reject, construction-not-existence, normalization matrix | ≤100 |
| `internal/evidence/plan_inputs_test.go` | colocated: prefill == `RequiredPlanInputs`, len 2 | ≤60 |
| `tests/unit/token_diet_plan_snapshot_test.go` | tier-level: legacy 122-path evidence still validates | ≤100 |
| `tests/acceptance/token_diet_plan_evidence_test.go` | drives the built binary: `evidence init` → `evidence validate` exit 0 with the 2-entry prefill | ≤100 |

Key assertion for the no-glob invariant: call `RequiredPlanInputs` from a
`t.Chdir`-ed temp dir containing 0 briefs and one containing 50, and assert
**identical** results. That is the test that fails if anyone reintroduces a
glob.

---

## 3. Slice B — UX tag matcher

**Delivers:** AC9–AC12, E8–E14. **Blast radius:** the ux-ui-specialist gate
only; strictly loosening.

`internal/orchestration/evidence_ux.go` — `normalizeUXTag`

Order of operations matters and is part of the contract:

1. trim, lowercase
2. cut at the **first** `:` (keep the head; a leading `:` yields `""`)
3. trim again (handles `"MOBILE_FIRST : Renders"`, E13)
4. `_` → `-`, space → `-`

Guard: an empty result is inserted into the tag set like any other value and
simply matches no required tag — no special-casing, no panic (E11). Do **not**
add an early `continue` that skips empty entries; keeping the map insert
uniform is what makes E11 a one-line proof.

`validateUXEvidence` and `missingUXTags` are otherwise untouched.

### B tests

| File | New/changed | Budget |
|------|-------------|--------|
| `internal/orchestration/evidence_ux_test.go` | table-driven `normalizeUXTag` cases E8–E13 + a full 8-tag descriptive-only evidence set passing + a missing-tag set still failing and naming it | ≤100 |
| `tests/unit/token_diet_ux_tag_test.go` | tier-level: `validateUXEvidence` accepts an all-descriptive edgeCases list; non-ux role untouched (E14) | ≤80 |

**Regression guard (AC10):** keep the existing bare-token test cases. The
whole point is that the change only adds matches.

---

## 4. Slice C — digest-dedup the per-prompt roadmap summary

**Delivers:** AC13–AC18, E15–E24. **Blast radius:** every user prompt.
**Opens with a split** (R2).

### C1. Domain: digest + decision (G7 — no logic in `cmd/`)

New `internal/roadmap/summary_digest.go` (target ≤80 lines):

- `SummaryDigest(r *Roadmap) string` — pure. Hash a *stable projection* of the
  summary-relevant state: the `planned/inProgress/done` counts plus each
  phase's name and each feature's name+status, in declaration order. Do **not**
  hash the raw file bytes — reformatting churn (a known repo hazard) would
  cause spurious re-renders. `sha256` truncated to 16 hex chars is plenty.
- `type SummaryState struct { SessionID, Digest string }`
- `LoadSummaryState(path string) SummaryState` — never returns an error;
  absent / corrupt / unreadable ⇒ zero value (E19, E20).
- `SaveSummaryState(path string, s SummaryState) error` — temp file +
  `os.Rename` in the same directory (E22). Caller ignores the error.
- `ShouldRenderSummary(prev SummaryState, sessionID, digest string) bool` —
  pure: `true` when `sessionID == ""` (fail-open, no session signal),
  `prev.SessionID != sessionID`, or `prev.Digest != digest`.
- `SummaryStatePath() string` — `.workflow/.roadmap-digest`.

`internal/roadmap` is already a mapped domain package, so no `import_graph`
layer change and no `PROJECT.md` G2 edit is needed. Keep it stdlib-only.

### C2. Outer layer: split, then wire

- Extract the roadmap block from `cmd/centinela/hook_context.go` into new
  `cmd/centinela/hook_context_roadmap.go` (target ≤50 lines), mirroring the
  existing `hook_context_review_mode.go` pattern. It holds a thin
  `emitRoadmapSummary(sessionID string)` that: loads the roadmap, computes the
  digest, loads prior state, asks `roadmap.ShouldRenderSummary`, prints
  `ui.RenderRoadmapSummary(r)` when true, and saves state. No decisions of its
  own — every branch is a domain call (G7).
- `runHookContext` changes from `io.ReadAll(os.Stdin)`-and-discard to
  read-then-best-effort-`json.Unmarshal` into
  `struct{ SessionID string \`json:"session_id"\` }`. Any read/parse failure ⇒
  empty session id ⇒ render (E21, AC17). Reuse the `hook_prewrite.go` stdin
  pattern; do **not** add a dependency between the two hook files.
- `roadmap.Load()` failure keeps its current behavior: no line, no state write
  (E23).

### C3. Gitignore (AC18, R8)

Add `.workflow/.roadmap-digest` to `.gitignore` **in the same commit as the
writer**. The existing `.workflow/` block is an explicit allow-list of durable
state with negations — add the ignore line inside that block and re-read the
block before editing; a stray pattern there has previously blocked greenfield
`start`.

### C tests

| File | New/changed | Budget |
|------|-------------|--------|
| `internal/roadmap/summary_digest_test.go` | pure: digest stability across reformat, digest change on status change, `ShouldRenderSummary` truth table (E15–E18), corrupt/absent load (E19), atomic save round-trip | ≤100 |
| `internal/roadmap/summary_digest_io_test.go` | unwritable dir (E20), concurrent save (E22) | ≤80 |
| `cmd/centinela/hook_context_roadmap_test.go` | colocated: first call prints the line, second identical call does not, other output unchanged; empty/garbage stdin ⇒ prints (E21) | ≤100 |
| `tests/acceptance/token_diet_hook_quiet_test.go` | built binary, two invocations with the same `session_id` in stdin: line present then absent; then a mutated roadmap ⇒ present; `git status --porcelain` clean (AC18) | ≤100 |

**Isolation warning:** these tests write `.workflow/.roadmap-digest`. Run them
in a `t.TempDir()` + `t.Chdir()` so they never touch the real repo state and
never leak across tests.

---

## 5. Slice D — family aliases instead of dated pins

**Delivers:** AC19–AC23, E25–E27. **Blast radius:** every printed model id and
the capability→profile default chain (R3). **Opens with a split** (R2).

### D1. The table

`internal/orchestration/resolve.go` — `tierModels`

| Tier | claude | opencode |
|------|--------|----------|
| reasoning | `opus` | `anthropic/claude-opus-4-7` |
| balanced | `sonnet` | `anthropic/claude-sonnet-4-6` |
| fast | `haiku` | `anthropic/claude-haiku-4-5` |

Decision: the `claude` column becomes the bare CLI family aliases, which the
Claude Code runner resolves to the current family member — this is the column
that carried the only dated snapshot and the one printed into the directive.
The `opencode` column keeps a provider-qualified id because a bare alias is
not a valid opencode model reference; the only change there is dropping the
`-20251001` date, which that column never carried. Do **not** invent an
opencode "latest" form that has not been verified against the provider.

`codex` stays empty — the rule-4 fallback (tier name, `ok=false`) is a
deliberate contract, not a gap (AC23, E26).

Keep the existing comment that this table is the single edit point for a model
refresh, and extend it to name `[orchestration.model_map]` as the
no-release override path (AC21). Add a commented `[orchestration.model_map]`
example to `centinela.toml` next to the existing orchestration block.

### D2. Capability superset (R3 — the silent one)

`internal/config/capability.go` is at **exactly 100 lines**. First move:
extract `builtinModelCapability` (and, if needed for budget,
`defaultProfileForClass`) into new `internal/config/capability_models.go`
(target ≤45 lines). `capability.go` drops to ~75.

Then **add** — never replace — the alias keys:

```
"opus" → frontier, "sonnet" → capable, "haiku" → limited
```

Every existing key (`claude-opus-4-7`, `anthropic/claude-opus-4-7`,
`claude-sonnet-4-6`, `anthropic/claude-sonnet-4-6`,
`claude-haiku-4-5-20251001`, `anthropic/claude-haiku-4-5`) **stays**, so an
operator whose `driver_model` names a retired pin keeps its class and keeps
its default enforcement profile (AC22, E25).

### D3. Fan-out inventory (R4)

Every file below hardcodes a pinned id. Sweep with
`grep -rn "claude-opus-4-7\|claude-sonnet-4-6\|claude-haiku-4-5"` and confirm
the sweep is empty of *stale* expectations before calling the slice done
(`.workflow/` historical evidence files are frozen records — **leave them**).

| File | Action |
|------|--------|
| `cmd/centinela/hook_orchestration_plan_test.go` | update expected id |
| `cmd/centinela/hook_orchestration_routing_test.go` | update expected id |
| `cmd/centinela/calibrate_test.go` | update expected id |
| `cmd/centinela/telemetry_model_test.go` | update expected id |
| `tests/unit/configurable_subagent_models_unit_test.go` | update expected id |
| `specs/configurable-subagent-models.feature` | prefer asserting the tier→alias mapping; update literals otherwise |
| `specs/configurable-model-routing.feature` | same |
| `specs/model-capability-profiles.feature` | same — and add the superset guarantee |
| `specs/capability-calibration.feature` | same |
| `centinela.toml` (~line 174 quota comment) | update the commented example id |

Editing four existing `.feature` files is intentional, not spec drift: those
specs froze the *pins* this feature exists to remove. Where a scenario's point
is "a configured override wins", rewrite it to name the alias generically
rather than re-pinning a new literal.

### D tests

| File | New/changed | Budget |
|------|-------------|--------|
| `internal/orchestration/resolve_test.go` | extend: **no built-in id matches `-\d{8}$`** (AC19), claude column has no digits (AC20), `ModelReference` renders aliases, codex fallback renders the tier name | ≤100 |
| `internal/config/capability_models_test.go` | table-driven: every `tierModels` value classifies, every legacy pinned id still classifies (AC22) | ≤100 |
| `tests/unit/token_diet_model_alias_test.go` | tier-level: `ResolveModel` precedence 1→4 intact with aliases; `model_map` override wins (AC21, E27) | ≤100 |
| `tests/acceptance/token_diet_directive_test.go` | built binary: `hook orchestration` prints the alias and no dated id | ≤80 |

**Cross-package caution:** `internal/orchestration` cannot import
`internal/config` (config is the leaf and imports nothing internal; the
dependency runs the other way). The "every `tierModels` value has a capability
class" assertion therefore lives in `internal/config` (which may reference the
ids as literals) or in `tests/unit/` — **not** in `internal/orchestration`.
Exporting a `TierModelIDs()` accessor from `orchestration` for the config test
to consume is the clean option if a literal list is unacceptable.

---

## 6. File-level line budgets

G1 is ≤100 lines for **every** source file including `_test.go`, in both
`internal/` and `cmd/`. Diff-aware `validate` hides pre-existing overflow; CI's
full scan does not.

| File | Now | After | Budget | Note |
|------|-----|-------|--------|------|
| `internal/orchestration/plan_snapshot.go` | 73 | ~62 | ≤75 | glob removal −12, anchor +4 |
| `internal/evidence/plan_inputs.go` | 16 | 16 | ≤30 | comment refresh only |
| `internal/orchestration/evidence_ux.go` | 51 | ~57 | ≤70 | +colon cut + comment |
| `internal/orchestration/resolve.go` | 74 | ~76 | ≤85 | table values + comment |
| `internal/config/capability.go` | **100** | ~75 | ≤100 | **must split first** |
| `internal/config/capability_models.go` | — | ~45 | ≤60 | new: the id→class table |
| `cmd/centinela/hook_context.go` | **98** | ~96 | ≤100 | roadmap block moves out, stdin parse moves in |
| `cmd/centinela/hook_context_roadmap.go` | — | ~50 | ≤60 | new: thin wiring only |
| `internal/roadmap/summary_digest.go` | — | ~80 | ≤95 | new: digest + state + decision |
| every new/changed `_test.go` | — | — | ≤100 | split by concern, not by line count |

Coverage is measured **per package** — `tests/unit/` and `tests/acceptance/`
files do not move the 95% gate. Every behavior change above therefore needs a
**colocated** `_test.go` in its own package; the tier-level files are
additional, not substitutes. Aim ≥97% so a parallel merge cannot tip main red.

## 7. Rollout sequence

1. **A** — biggest, unbounded win; also the one entangled with this feature's
   own evidence (R1), so prove it first while the workflow is still early.
2. **B** — smallest diff, zero coupling; lands in one commit.
3. **C** — new persisted state and a new stdin parse; needs the `cmd` split
   and the gitignore line in the same commit.
4. **D** — widest textual fan-out (10 files) and the silent capability risk;
   last, so the full suite runs against an otherwise-settled tree.

After every slice: `go test ./... -run xxxNONE` (surfaces test-compile breaks
that `go build` hides) then the full `go test ./...`.

## 8. Validation checklist

- [ ] `centinela evidence validate token-diet` exit 0 — **before and after**
      slice A (R1: this feature's own 122-input evidence must still pass)
- [ ] `centinela validate` passes (lint, gates, full suite)
- [ ] `grep -rn "docs/features/\*\.md" docs/architecture/` reflects the new
      rule in all three docs
- [ ] `diff docs/architecture/<doc> internal/scaffold/assets/docs/architecture/<doc>`
      silent for all three (R7)
- [ ] no built-in model id matches `-[0-9]{8}$`
- [ ] `git status --porcelain` clean after a hook fire (AC18)
- [ ] every touched file ≤100 lines (full scan, not diff-aware)
- [ ] rebase onto `main` and re-run the full gate on the merged tree (R10)

## 9. Handoff notes for senior-engineer

- Do not "simplify" `validatePlanSnapshotInputs` into a set-equality check.
  Superset acceptance is a load-bearing back-compat guarantee (AC4/R1).
- Do not delete the legacy pinned ids from `builtinModelCapability`. It is a
  superset by design (AC22/R3).
- Do not put the render/suppress decision in `cmd/` (G7). It belongs in
  `internal/roadmap`.
- Fail-open is the rule for slice C everywhere: when in doubt, render.
- Historical `.workflow/*-{big-thinker,senior-engineer,…}.md` files that
  mention the old model ids are frozen evidence records — do not rewrite them.
