# Implementation Plan — `centinela insights`

> Feature brief: `docs/features/centinela-insights.md`.
> Spec: `specs/centinela-insights.feature`.

A new `internal/insights/` package takes `[]telemetry.Event` and computes a
pure, deterministic `Report`. `cmd/centinela/insights.go` is a THIN orchestrator
that reads events via `telemetry.ReadDefault()`, calls `insights.Compute`, and
renders (human via `internal/ui`, or JSON via `--json`). No reimplementation of
telemetry reading; no mutation; no telemetry writes.

## Metric definitions (DECIDED — implemented near-verbatim)

1. **Most-triggered blocks** — count `block` events bucketed by display key
   `"<reason> · <fileType>"` (each empty field rendered `<none>`; `step` shown
   in the human detail line when present). Rank by count desc, key asc; top-N.
2. **Most-failed gates** — count `gate-failure` events bucketed by `Gate`
   (empty ⇒ `<none>`). Rank count desc, key asc; top-N.
3. **Features with the most rework** — `rework score(feature) = #{gate-failure}
   + #{verify-rejection} + #{complete-rejected}` events whose `Feature` equals
   that feature (events with empty `Feature` are excluded). This is "governance
   friction before green." Rank desc, feature asc; top-N. (Chosen over
   "repeated failures on the same step" because it needs no windowing, is a
   single pass, and is unambiguous.)
4. **Mean steps-to-green** — across the whole repo:
   `mean = (#{complete-rejected} + #{step-advanced}) / #{step-advanced}`.
   "Green" = a `step-advanced` event (a step that completed and advanced). Each
   advance took `1 + (its rejections)` attempts; summed and averaged, this is
   total complete-attempts per successful advance. Denominator 0 (no advances)
   ⇒ `HasValue=false`, renderer prints `n/a` (no division, no panic).

## Determinism contract

Every ranked section sorts a `[]Count` slice (count desc, then `Key` asc); no
map is ranged in rendered/JSON order. `SpanStart`/`SpanEnd` are min/max of
event timestamps by string compare (RFC3339 UTC sorts lexically). Empty log ⇒
empty-state `Report`, exit 0. Golden tests assert byte-identical output.

## v1 scope

**In:** the 4 metrics; `internal/insights` package; `Compute`; the command with
`--top` (default 5) and `--json`; `internal/ui/render_insights.go`; empty/missing
log handling; import-graph layer mapping.
**Out (deferred):** `--since` / time-window filtering (cheap to add later, not
needed for prioritization v1); per-model breakdowns (owned by
`capability-calibration`); trend lines / sparklines (presentation, not core);
config-driven metric weights. `--json` IS in (cheap, enables tooling, recommended).

## Layer / import-graph decision (CALL-OUT)

`internal/insights` imports `internal/telemetry` (a config-only **leaf**) +
stdlib only; it must NOT import `cmd/` or `internal/ui`. It is itself imported
only by `cmd/`. This is the **same shape as `internal/doctor`** (an aggregator
read-only over domains), but lighter — it reads exactly one leaf. **Decision:**
join the existing `aggregator` layer rather than adding a new one. In
`centinela.toml`, extend `[[gates.import_graph.layers]] name="aggregator"` to
`paths = ["internal/doctor/**", "internal/insights/**"]` (allow already =
`["domain", "leaf"]`; telemetry is a leaf, so insights' only edge is satisfied —
**no warning, fully mapped**). Update PROJECT.md G2 prose to note insights joins
the aggregator layer and may import the `telemetry` leaf read-only. This avoids
a matrix rewrite and is cleaner than doctor (no unmapped edges).

## Step 1 — plan (this step)

Artifacts: this plan, `docs/features/centinela-insights.md`, and
`specs/centinela-insights.feature` (one scenario per metric + empty-log + --json
+ determinism). Big-thinker + feature-specialist evidence pairs in `.workflow/`.

## Step 2 — code

New files (each ≤100 lines):

| File | Responsibility | Budget |
|------|----------------|--------|
| `internal/insights/report.go` | `Report`, `Count`, `StepsStat` types; `rankTop(map[string]int, n) []Count` (stable sort: count desc, key asc, slice n) | ~70 |
| `internal/insights/blocks.go` | `blocks(events) []Count` via `--top` (bucket `reason · fileType`) | ~40 |
| `internal/insights/gates.go` | `gates(events) []Count` (bucket `Gate`) | ~35 |
| `internal/insights/rework.go` | `rework(events) []Count` (per-feature friction sum) | ~45 |
| `internal/insights/steps.go` | `stepsToGreen(events) StepsStat` + `span(events) (start,end)` | ~55 |
| `internal/insights/compute.go` | `Compute([]telemetry.Event, topN int) Report` — single pass + reducers | ~60 |
| `cmd/centinela/insights.go` | thin cobra command: flags `--top` (def 5), `--json`; `telemetry.ReadDefault`; `Compute`; render or `json.MarshalIndent` | ~70 |
| `internal/ui/render_insights.go` | `RenderInsights(insights.Report) string` — sectioned house-style report (header span+count, 4 sections, empty-state line) | ~95 |

Config: edit `centinela.toml` aggregator layer `paths` to add
`internal/insights/**`. Edit PROJECT.md G2 prose. Mirror any architecture-doc
changes into `internal/scaffold/assets/` if touched (none expected here).
`cmd/centinela/insights.go` registers `insightsCmd` on `rootCmd` in its `init()`.

## Step 3 — tests

Per-package **colocated** `_test.go` (the 95% per-package coverage gate is NOT
moved by `tests/` tier files — no `-coverpkg`), each ≤100 lines:

- `internal/insights/blocks_test.go`, `gates_test.go`, `rework_test.go`,
  `steps_test.go` — table-driven: empty, single, ties, missing-field buckets,
  top-N truncation; `steps_test.go` covers zero-advance ⇒ `n/a`.
- `internal/insights/compute_test.go` — end-to-end on a fixture event slice;
  determinism (run twice, assert equal); span min/max; report shape.
- `internal/insights/report_test.go` — `rankTop` tie-break + truncation.
- `internal/ui/render_insights_test.go` — golden human output + empty-state +
  non-TTY (ANSI-stripped) assertions.
- **Integration:** `tests/integration/insights_test.go` — write a temp
  `events.jsonl`, call `telemetry.Read(dir)` + `insights.Compute`, assert report.
- **Acceptance:** `tests/acceptance/insights_*.go` (executable) — one per
  Gherkin scenario: build/run the binary against a fixture log, assert human +
  `--json` + empty-log + exit codes. Add the acceptance runner to
  `validate.commands` in `centinela.toml`.
- `.workflow/centinela-insights-edge-cases.md` — enumerate the brief's edge
  cases with the test that covers each.

Note: `go test ./...` runs ~75s; verify_timeout=240 gives margin.

## Step 4 — validate

Gatekeeper report `.workflow/centinela-insights-gatekeeper.md`; `centinela
validate` green (lint + types + full suite + import-graph gate now clean for
insights). Production-readiness if gate enabled.

## Step 5 — docs

Documentation-specialist `.md` + `.json`; regenerate `docs/project-docs/
index.html`; changelog artifact `.workflow/centinela-insights-changelog.md`
(create early). Document the four metric formulas and the `--json` schema.
