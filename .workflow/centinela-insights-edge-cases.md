# Edge Cases: centinela-insights

## Covered

Each edge case below is exercised by a named test (unit colocated in
`internal/insights`/`internal/ui`/`cmd/centinela`, or acceptance in
`tests/acceptance/centinela_insights_*`).

- **Missing telemetry log** → clean "no telemetry yet", exit 0, no stack trace.
  `TestInsightsMissingLogEmptyState`, `cmd TestInsightsMissingLog`,
  `TestComputeEmpty`.
- **Empty log file** (zero bytes) → empty-state. `TestInsightsEmptyLogEmptyState`.
- **Whitespace/blank-only lines** → treated as empty (reader skips empty lines).
  `TestInsightsWhitespaceLogEmptyState`.
- **Malformed JSONL line** mixed with valid events → garbage skipped, valid
  events still aggregated. `TestInsightsMalformedLinesSkipped`,
  integration `TestInsightsPipelineAndJSONRoundTrip` (garbage line dropped).
- **Empty `fileType` on a block** → buckets as `<reason> · <none>`, no collapse.
  `TestBlocksEmptyFileTypeBucketsAsNone`, `TestInsightsBlocksEmptyFileType`.
- **Empty `Gate` on a gate-failure** → buckets as `<none>`.
  `TestGatesEmptyGateBucketsAsNone`, `TestInsightsGatesEmptyGateNone`.
- **Empty `Feature`** on a friction event → excluded (no anonymous bucket).
  `TestReworkExcludesEmptyFeature`, `TestInsightsReworkExcludesEmptyFeature`.
- **`step-advanced` not counted as rework** → excluded from friction score.
  `TestReworkExcludesStepAdvanced`, `TestInsightsReworkExcludesStepAdvanced`.
- **Zero step-advanced events** (division-by-zero guard) → `HasValue=false`,
  renders `n/a`, no panic. `TestStepsToGreenZeroAdvances`,
  `TestRenderInsightsStepsNA`, `TestInsightsStepsZeroAdvances`.
- **Single advance / single rejection means** (1.00 / 2.00 / 1.50 boundaries).
  `TestStepsToGreen*`, `TestInsightsSteps*`.
- **Tie-breaking** (equal counts) is stable by key ascending across all three
  ranked sections. `Test*TieBreak*`, `TestInsightsTiesBrokenByKeyAsc`.
- **`--top N` truncation**, including `N` larger than available buckets (no
  padding) and the default of 5. `Test*TopN*`, `TestInsightsBlocksDefaultTopFive`,
  `TestInsightsBlocksTopLargerThanBuckets`, `cmd TestInsightsTopTruncates`.
- **Determinism** — same log → byte-identical human and `--json` output.
  `TestComputeDeterministic`, `TestInsightsHumanByteIdentical`,
  `TestInsightsJSONStableTwoRuns`, `cmd TestInsightsJSONValidAndStable`.
- **`--json` contract** — valid indented JSON, exactly the 7 top-level Report
  fields, zero-count report on empty log, no ANSI. `TestInsightsJSON*`,
  `TestInsightsJSONEmptyLogZeroCount`.
- **Non-TTY / piped output** — no ANSI escapes, plain printable text.
  `TestInsightsPipedNoANSI`, `TestRenderInsightsNoANSI`.
- **Span min/max** over RFC3339 timestamps, ignoring empty timestamps; `(none)`
  when no timestamps. `TestSpanMinMax`, `TestSpanEmpty`, `TestInsightsSpanRange`,
  `TestRenderInsightsNoSpan`.
- **Single-type logs** — only blocks / only gates render that section non-empty
  and the others as `(no events)`. `TestInsightsOnlyBlocks`, `TestInsightsOnlyGates`.
- **One-of-each-type log** reported without crash; total event count surfaced.
  `TestInsightsSingleOfEachType`, `TestInsightsTotalEventCount`.

## Residual Risks

- **Oversized JSONL lines** (>1 MiB) are dropped by the telemetry reader's
  scanner buffer cap (owned by `internal/telemetry`, not this feature). Out of
  scope here; mitigated upstream where events are written one compact line each.
- **Non-UTC / non-RFC3339 timestamps** would sort lexically and could misorder
  the span; mitigated because the only writer (`telemetry.Record`) emits RFC3339
  UTC. No defensive parse is added (would duplicate the leaf's contract).
- **`--since` / time-window filtering** is deferred (plan v1 scope-out); current
  span covers the whole log. Not a correctness risk, a feature gap.
