# Implementation Plan — failure-ledger plan advisor

> Feature brief: `docs/features/failure-ledger-plan-advisor.md`.
> Spec: `specs/failure-ledger-plan-advisor.feature`.

Phase 7 closes the governance loop: the plan advisor already runs automatically
during the `plan` step and assembles a context `bundle`
(`internal/planadvisor/context.go`). We add one more **read-only** context
source — the top-N recurring `gate-failure` events from the telemetry ledger —
surfaced as (a) a "Recurring gate failures" summary line and (b) one
pre-warning advisor question naming the worst gate(s). Aggregation reuses the
existing insights gate-counter verbatim, so counts can never diverge. When the
ledger is empty/missing/has no gate-failures, or `[telemetry] enabled = false`,
the failure slice is empty and output is **byte-identical to today**.

## Decisions (DECIDED)

1. **Reuse, don't duplicate (AC-5).** `internal/insights/gates.go` currently has
   an unexported `gates(events, topN) []Count`. Export a thin public wrapper in
   that same file:
   ```go
   // Gates ranks gate-failure events by Gate (count desc, key asc), top-N.
   // Exported for reuse by the plan advisor so counts never diverge.
   func Gates(events []telemetry.Event, topN int) []Count { return gates(events, topN) }
   ```
   `Compute` keeps calling lowercase `gates`; planadvisor calls `insights.Gates`.
   `Count` (`{Key string; Count int}`) is already exported. No counting logic is
   copied. The `<none>` bucketing for empty `Gate` (edge case) is inherited from
   `orNone`, and the stable order from `rankTop` (count desc, key asc) — so
   **determinism is inherited, not re-implemented** (AC-1 tie-break, Risk #2).

2. **Read path.** A new `internal/planadvisor/failures.go` owns the read:
   ```go
   func recurringFailures(cfg *config.Config, topN int) []insights.Count {
       if cfg == nil || !cfg.Telemetry.IsEnabled() { return nil }   // AC-4
       events, err := telemetry.ReadDefault()                       // missing ⇒ (nil,nil)
       if err != nil { return nil }                                 // never break planning
       return insights.Gates(events, topN)                          // empty ⇒ nil
   }
   ```
   `telemetry.ReadDefault()` already returns `(nil, nil)` for a missing file and
   skips malformed lines, so no panic on a missing/sparse ledger (edge cases).
   The advisor stays read-only against the ledger — it never calls
   `telemetry.Record`.

3. **Bundle field.** `bundle` (context.go) gains `Failures []insights.Count`,
   populated in `buildBundle` via `recurringFailures(cfg, failureTopN(cfg))`.
   `buildBundle` already receives `cfg`, so no signature change ripples out to
   `Directive`/the hook.

4. **Config knobs (minimal).** Add ONE knob, `plan_advisor_failure_top_n`
   (`[workflow]`), default **3**, normalized by a new
   `NormalizePlanAdvisorFailureTopN` (clamp `<=0` → 3, cap at 5 to keep the
   summary block concise — Risk "noise" + edge case "many distinct gates").
   **No separate recurrence threshold knob:** the pre-warning question fires when
   the top gate's count `>= 2` (a hardcoded "recurred" floor — a single failure
   is not yet a pattern; AC-2). A 2nd config knob for the threshold is **out of
   v1 scope** (deferred; the constant is the right default and avoids knob bloat).
   The question cap reuses the existing `plan_question_limit` via the unchanged
   `selectQuestions(..., limit, ...)` loop (AC-2).

## Import-graph / layer decision (CALL-OUT)

**Verdict: no new layer, no cycle, no new failing edge — safe.**

- `internal/planadvisor` is **currently unmapped** in the G2 `import_graph`
  matrix (`centinela.toml` maps leaf/domain/aggregator/cmd; planadvisor is none
  of these). It already imports the unmapped `internal/memory` and
  `internal/roadmap` plus the `internal/config` leaf. Unmapped→unmapped/leaf
  edges surface only as the **existing non-failing Warn**, per the matrix's
  conservative policy (toml lines 47-55).
- Adding edges `planadvisor → internal/insights` (aggregator) and
  `planadvisor → internal/telemetry` (leaf) keeps planadvisor unmapped, so both
  new edges are the same non-failing-Warn kind. They do **not** trip the gate.
- **No import cycle:** `insights` imports only `internal/telemetry`; `telemetry`
  imports only `internal/config`. Neither imports `planadvisor` (verified:
  planadvisor is imported only by `cmd/centinela/hook_plan_advisor.go`; insights
  only by `cmd/` and `internal/ui`). The graph `planadvisor → insights →
  telemetry → config` is acyclic.
- **Why this over alternatives:** lifting the gate-counter into a new leaf
  package both insights and planadvisor import would touch insights' internals,
  re-route `Compute`, and add a package for a 3-line function — more churn, no
  layering benefit, since exporting `insights.Gates` already guarantees a single
  source of truth (AC-5). Passing pre-aggregated counts from the hook layer
  would force the hook to read telemetry and thread a new arg through
  `Directive`/`buildBundle` — leaking an aggregation concern into `cmd/`. The
  chosen in-bundle read keeps the advisor self-contained and matches how it
  already reads roadmap/memory directly.
- **No toml change required** (planadvisor stays unmapped). PROJECT.md G2 prose
  gets a one-line note: *the plan advisor may read `internal/insights`
  (aggregator) and `internal/telemetry` (leaf) read-only.* If a future change
  maps planadvisor, it would join the `cmd`-adjacent consumers allowing
  `aggregator`+`leaf`; out of scope now.

## v1 scope

**In:** `insights.Gates` export; `planadvisor/failures.go` read (telemetry-gated,
read-only); `bundle.Failures`; one "Recurring gate failures" summary line; one
pre-warning question (lens `feature-specialist`) gated on top-count `>= 2` and
the existing question cap; `plan_advisor_failure_top_n` knob (default 3, cap 5);
byte-identical empty-state behaviour.
**Out (deferred):** a configurable recurrence threshold knob; time-window /
`--since` filtering of the ledger; per-step or per-feature failure attribution;
surfacing block/rework metrics into the advisor (only gate-failures in v1);
rendering counts anywhere but the plan-advisor directive.

## Step 2 — code

New / edited source files (each ≤100 lines):

| File | Change | Budget |
|------|--------|--------|
| `internal/insights/gates.go` | add exported `Gates` wrapper over `gates` (4 lines + doc) | ~22 |
| `internal/planadvisor/failures.go` | NEW. `recurringFailures(cfg, topN) []insights.Count` (telemetry-gated read) + `failureTopN(cfg) int` (calls `config.NormalizePlanAdvisorFailureTopN`) | ~35 |
| `internal/planadvisor/context.go` | add `Failures []insights.Count` to `bundle`; set it in `buildBundle` | ~30 |
| `internal/planadvisor/context_summary.go` | add a "Recurring gate failures" line builder appended in `contextLines` (renders `gate (×N)` joined, top-N) | ~98 (currently 86) |
| `internal/planadvisor/questions.go` | add ONE candidate: `{"feature-specialist", "The ledger shows recurring gate failures (worst: <gate> ×N). What plan choices prevent that gate from biting again?", topFailureCount(b) >= 2}` ; helper `topFailureCount(b)`/`worstGate(b)` | ~95 (currently 50) |
| `internal/config/workflow_config.go` | add `PlanAdvisorFailureTopN int \`toml:"plan_advisor_failure_top_n"\`` | ~31 |
| `internal/config/plan_advisor.go` | add `DefaultPlanAdvisorFailureTopN = 3`, `MaxPlanAdvisorFailureTopN = 5`, `NormalizePlanAdvisorFailureTopN(n) int` | ~40 |
| `internal/config/defaults.go` | normalize `PlanAdvisorFailureTopN` in `applyDefaults` | +1 line |

Notes:
- **Summary line format (AC-1):** `- recurring gate failures: g1-file-size (×8), import-graph (×3)` — gate name then `(×count)`, ordered by the inherited rank, truncated to top-N. Empty `Failures` ⇒ no line appended (AC-3 byte-identical).
- **Question text (AC-2):** names the single worst gate + its count; tagged
  `feature-specialist`; flows through the existing `selectQuestions` cap loop so
  `plan_question_limit` bounds it. `Ask` is `topFailureCount(b) >= 2`, so a
  single isolated failure (count 1) stays quiet (Risk "noise").
- If `questions.go` or `context_summary.go` approach the 100-line budget after
  edits, split the new helpers into a tiny `failures_view.go` in planadvisor
  (pure formatting, no I/O) — keep each file ≤100.

## Step 3 — tests

Colocated per-package `_test.go` (the 95% per-package coverage gate is NOT moved
by `tests/` tier files — add coverage next to the code), each ≤100 lines:

- `internal/insights/gates_test.go` — add a case asserting `Gates` ==
  `gates` for the same input (proves the wrapper is a pure pass-through; AC-5),
  plus the existing `<none>`/tie/top-N cases stay green.
- `internal/planadvisor/failures_test.go` — table-driven over a temp
  `events.jsonl` written under a `t.TempDir()` `.workflow/telemetry/` (chdir or
  inject): (a) gate-failures present ⇒ ranked `[]Count`; (b) `enabled=false` ⇒
  `nil` (AC-4); (c) missing file ⇒ `nil` (edge case); (d) only `block` events ⇒
  `nil` (edge case); (e) empty `Gate` ⇒ `<none>` bucket (edge case); (f) ties ⇒
  alphabetical (AC-1).
- `internal/planadvisor/context_summary_test.go` — assert the summary line is
  present + exactly formatted when `Failures` non-empty, and **absent** when
  empty (AC-3 byte-identical guard).
- `internal/planadvisor/questions_test.go` — assert the pre-warning question
  fires at count `>= 2`, names the worst gate, is silent at count `1` and on
  empty `Failures`, and respects `plan_question_limit` (AC-2).
- `internal/config/plan_advisor_test.go` — `NormalizePlanAdvisorFailureTopN`
  clamps `<=0`→3 and caps `>5`→5; defaults applied.

**Integration:** `tests/integration/planadvisor_failures_test.go` — write a real
temp `events.jsonl`, build a `*config.Config` (telemetry on, then off), call
`planadvisor.Directive(feature, cfg)`, assert the directive string contains the
recurring-failures line + question when on, and is unchanged when off / ledger
empty.

**Acceptance:** `tests/acceptance/planadvisor_failures_*.go` (executable, one per
Gherkin scenario) — run the `centinela hook plan-advisor` path (or `Directive`
via the binary) against a fixture ledger and assert: (1) summary line +
question present with failures; (2) byte-identical output with empty/missing
ledger; (3) no output growth with `[telemetry] enabled=false`. Register the
acceptance runner in `validate.commands` in `centinela.toml`.

`.workflow/failure-ledger-plan-advisor-edge-cases.md` — enumerate the brief's
edge cases (missing file, block-only, empty `Gate`, ties, telemetry disabled,
many gates → top-N, headless unchanged) mapped to the test covering each.

Note: `go test ./...` runs ~75s; `verify_timeout=240` gives margin.

## Step 4 — validate

Gatekeeper report `.workflow/failure-ledger-plan-advisor-gatekeeper.md`;
`centinela validate` green (lint + types + full suite). Confirm the G2
import-graph gate output is unchanged in **failing** edges (only the existing
non-failing Warn set may grow by the new planadvisor→insights/telemetry edges).
Confirm every touched source file ≤100 lines. Production-readiness subagent if
the gate is enabled.

## Step 5 — docs

Documentation-specialist `.md` + `.json`; regenerate
`docs/project-docs/index.html`; changelog artifact
`.workflow/failure-ledger-plan-advisor-changelog.md` (create early). Document:
the new "Recurring gate failures" advisor signal, the `>= 2` recurrence floor,
the `plan_advisor_failure_top_n` knob (default 3, cap 5), and the read-only /
telemetry-gated guarantee. Add a PROJECT.md G2 one-line note that the plan
advisor reads `internal/insights` + `internal/telemetry` read-only.
