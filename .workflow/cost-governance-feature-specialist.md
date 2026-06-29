# cost-governance — feature-specialist

## Behavior Summary

A capture hook reads the harness `transcript_path`, sums new token usage since
the feature's cursor, and records a `cost-sample` telemetry event attributed to
the active feature/step/model. `centinela cost` reports used/budget/remaining
per feature/step; `centinela validate` appends a non-failing ⚠ when the active
step is over budget. Nothing about cost ever blocks `centinela complete`.

## Acceptance Criteria (Gherkin)

See `specs/cost-governance.feature` — capture, no-double-count (cursor),
over-budget-warns-never-blocks (validate exits 0), report shows used/budget/
remaining, disabled/zero = silent no-op, missing/malformed transcript degrades.

## UX States

- **Within budget**: `centinela cost` row shows used < budget, remaining > 0; no
  warning anywhere.
- **Over budget**: row flagged (⚠/over marker); validate appends one ⚠ line.
- **No budget / disabled**: no cost rows, no warnings (silent).
- **No spend data**: report shows 0 / "no samples yet", not an error.

## Edge Cases

- Repeated capture must not double-count — cursor per feature counts only delta.
- Over-budget is always non-failing — validate exits 0 with a ⚠.
- Malformed transcript line skipped, not fatal; missing transcript_path → no-op.
- Old telemetry lines lack token fields → read as 0 (schema back-compat).

## Out-of-Scope

- Hard/blocking enforcement (`cost.enforce` is a future feature).
- Dollar pricing (report tokens/units, not currency).
- Non-Claude transcript formats beyond graceful degradation.

## Handoff

→ senior-engineer: implement config + `internal/cost` (analytics layer over
telemetry+config) + telemetry cost-sample + capture hook + `centinela cost` +
validate ⚠. Keep each file ≤100 lines; add `internal/cost` to the import_graph
analytics layer paths.
