# Feature Brief — `centinela insights`

> Plan: `docs/plans/centinela-insights.md`.
> Spec: `specs/centinela-insights.feature`.

`centinela insights` is a **read-only analytics** command over the governance
telemetry log (`.workflow/telemetry/events.jsonl`, shipped by
`governance-telemetry`). It computes and renders four evidence metrics:
**most-triggered blocks**, **most-failed gates**, **features with the most
rework**, and **mean steps-to-green**. No mutation, no telemetry writes.

## Problem

The maintainer of a Centinela-governed repo decides what to harden next —
which prewrite blocks are noisiest, which validate gates fail most, which
features churned hardest, how many `complete` attempts a step typically takes —
**by anecdote and memory**, not evidence. The telemetry log already records
every block, gate failure, verify rejection, complete-rejection, and
step-advance, but nothing reads it back. So roadmap prioritization (and rule
tuning) is guesswork. `centinela insights` turns the existing append-only event
stream into a deterministic, sectioned report (and a `--json` payload for
tooling), so prioritization is driven by counts, not gut feel.

## User Stories

- As a **repo maintainer**, I run `centinela insights` and immediately see the
  top-N most-triggered prewrite blocks (by reason + file type) so I know which
  guardrails fire most and may need ergonomics work.
- As a **governance owner**, I see the most-failed validate gates ranked by
  count so I can prioritize gate fixes or developer guidance.
- As a **roadmap planner**, I see features ranked by rework (governance friction
  before green) so I invest hardening effort where churn is highest.
- As a **process analyst**, I see the mean steps-to-green (complete attempts per
  advanced step) so I can quantify workflow friction over time.
- As a **tooling author**, I run `centinela insights --json` and get a stable,
  machine-readable `Report` to feed dashboards or CI annotations.
- As **any user with no telemetry yet**, I run the command and get a clean
  "no telemetry yet" report with exit 0, never an error.

## Acceptance Criteria

> Concrete, testable → Gherkin in `specs/centinela-insights.feature`.

1. **Most-triggered blocks** — Given a log with N `block` events, when I run
   `centinela insights`, then the Blocks section lists block buckets keyed by
   `(reason, fileType)` (step shown when present), ranked by count descending,
   with the top default = 5 (configurable via `--top`).
2. **Most-failed gates** — Given `gate-failure` events, the Gates section lists
   gates keyed by `Gate`, ranked by count descending, top-N.
3. **Features with the most rework** — Given mixed events, the Rework section
   ranks features by **rework score = count of `gate-failure` +
   `verify-rejection` + `complete-rejected` events attributed to that feature**,
   descending, top-N.
4. **Mean steps-to-green** — Given `step-advanced` and `complete-rejected`
   events, the report prints **mean attempts-to-green =
   (complete-rejected count + step-advanced count) / step-advanced count**,
   rounded to 2 decimals; with zero advances it prints `n/a` (no panic).
5. **Empty / missing log** — Given no telemetry file (or zero events), the
   command prints a clean "no telemetry yet" report and exits 0.
6. **`--json`** — Given `--json`, the command emits the structured `Report` as
   indented JSON to stdout (no styled prose, no ANSI) and exits 0.
7. **Determinism** — Given the same log, two runs produce byte-identical output
   (human and `--json`); ties break by count desc, then key asc; no map is
   ranged in rendered order.

## Edge Cases

- **Missing log file** — `telemetry.Read` returns `(nil, nil)`; report is the
  empty-state report, exit 0.
- **Empty / whitespace-only log** — zero events ⇒ empty-state report.
- **Malformed JSONL lines** — already skipped by the lenient reader; insights
  operates only on parsed events (no extra handling needed, but tested).
- **Feature with zero advances** — excluded from the mean denominator; if NO
  feature ever advanced, mean = `n/a`.
- **Division by zero** — guarded in the mean computation (denominator 0 ⇒ `n/a`).
- **Ties** — equal counts break deterministically by key ascending (so output
  order is stable across runs and platforms).
- **Events missing optional fields** — a `block` with empty `step`/`fileType`
  buckets under `(reason, "")`; a `gate-failure` with empty `Gate` buckets under
  `""` (rendered as `<none>`); a feature-less event is excluded from rework.
- **Very large logs** — single O(N) pass over events, O(buckets log buckets)
  sort; no quadratic behavior; the lenient reader caps line size at 1 MiB.
- **`--json` shape stability** — field names/structure are part of the contract;
  golden-tested so tooling does not break silently.
- **Non-TTY output** — lipgloss styles auto-strip ANSI, so piped/redirected
  output is plain and parseable (matches doctor/gates renderers).
- **Single event of each type** — a log with exactly one block, one gate-failure,
  one verify-rejection, one complete-rejected, and one step-advanced event must
  produce counts of 1 per bucket and steps-to-green = 2.00 (1 rejection + 1
  advance / 1 advance); no crash or empty-section panic.
- **Three-way tie for same count** — e.g. three gate buckets each with count 2 and
  keys "z-gate", "m-gate", "a-gate" must always render in key-ascending order
  ("a-gate", "m-gate", "z-gate") regardless of map iteration order.
- **--top N exceeds available buckets** — when N > number of distinct keys in a
  section (e.g. `--top 10` with only 2 block buckets), all available entries are
  returned without padding or error.
- **Log with only step-advanced events** — rework and gates and blocks sections are
  empty (no crash); steps-to-green mean = 1.00 (advances with no rejections).
- **Log with only block events** — gates and rework sections are empty; steps-to-green
  = `n/a` (no advances); command exits 0.
- **step-advanced events excluded from rework score** — `step-advanced` must not
  increment any feature's rework tally; only gate-failure, verify-rejection, and
  complete-rejected count.
- **SpanStart/SpanEnd in human output** — the human report must display the
  earliest and latest event timestamps so users can understand the log coverage
  window; both are "" for an empty log and omitted from empty-state output.
- **Total EventCount in human output** — the human report must show the total
  number of parsed events considered, giving a sense of log volume.

## Data Model

A new `internal/insights` package owns a pure, serializable `Report`:

```go
type Report struct {
    EventCount int          // total parsed events considered
    SpanStart  string       // earliest event timestamp (RFC3339), "" if none
    SpanEnd    string       // latest event timestamp,           "" if none
    Blocks     []Count      // most-triggered blocks (ranked, top-N)
    Gates      []Count      // most-failed gates    (ranked, top-N)
    Rework     []Count      // features by rework score (ranked, top-N)
    StepsToGreen StepsStat  // mean attempts-to-green
}

// Count is a generic ranked bucket: a display key + its tally.
type Count struct {
    Key   string // "<reason> · <fileType>" / gate name / feature name
    Count int
}

// StepsStat is the steps-to-green metric, computed safely.
type StepsStat struct {
    Advances   int     // # of step-advanced events (the denominator / "green"s)
    Rejections int     // # of complete-rejected events
    Mean       float64 // (Rejections + Advances) / Advances ; 0 when Advances==0
    HasValue   bool    // false when Advances==0 (renderer prints "n/a")
}
```

`Compute([]telemetry.Event, topN int) Report` is the single entry point: pure,
deterministic, stdlib-only beyond importing `internal/telemetry`.

## Integration Points

- **Telemetry log** — consumed read-only via `telemetry.ReadDefault()` (default
  dir) in the thin command; `insights.Compute` takes the slice (no I/O), so it
  is trivially unit-testable.
- **`centinela.toml`** — none required for behavior. One config touch: add an
  `insights` entry to `[gates.import_graph]` (see plan) for layer cleanliness.
- **`--json` consumers** — dashboards / CI annotations read the `Report` JSON.
- **`internal/ui`** — new `render_insights.go` renders the human report in
  house style (glyphs/StyleMuted/StyleBold from `render_doctor.go` /
  `render_gates.go`).

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Metric definitions are misleading (rework / steps-to-green chosen wrong) | Medium | Medium | Definitions stated explicitly here + in `--help`/docs; conservative, defensible formulas; `--json` exposes raw counts so consumers can recompute. |
| Non-deterministic output (map iteration) | Medium | Medium | All ranking sorts a slice; tie-break count desc then key asc; golden tests assert byte-stability. |
| ≤100-line file budget exceeded | Low | Medium | One compute file per metric + a types file + an aggregator; renderer split if needed. |
| `--json` shape drift breaks tooling | Low | Low | Golden JSON test; shape documented as a contract. |
| Telemetry schema evolves | Low | Low | Insights reads only stable fields; unknown fields ignored by the lenient reader. |

## Decomposition

Cohesive, per-metric file units, each ≤100 lines:

- `internal/insights/report.go` — `Report`, `Count`, `StepsStat` types + `topN`
  helper / stable sort utility.
- `internal/insights/blocks.go` — most-triggered blocks.
- `internal/insights/gates.go` — most-failed gates.
- `internal/insights/rework.go` — features by rework score.
- `internal/insights/steps.go` — mean steps-to-green + span.
- `internal/insights/compute.go` — `Compute` orchestrator (single O(N) pass +
  calls per-metric reducers).
- `cmd/centinela/insights.go` — thin command (`--top`, `--json`).
- `internal/ui/render_insights.go` — human renderer.
