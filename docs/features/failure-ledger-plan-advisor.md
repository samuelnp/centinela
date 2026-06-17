# Feature Brief — failure-ledger plan advisor

> Phase 7: Instrument the Loop. Feed recurring gate failures from the
> governance-telemetry ledger into the plan advisor so the next feature is
> pre-warned about the failure modes that actually bite this repo.

## Problem

`centinela insights` already aggregates the telemetry ledger
(`.workflow/telemetry/events.jsonl`) into ranked metrics — including the
most-failed gates — but that signal is **passive**: a maintainer must run the
command and read it. The plan advisor (`internal/planadvisor/`), which runs
automatically during the `plan` step, never consults the ledger. So a feature
whose repo has failed `g1-file-size` 8 times in the last month starts planning
with zero awareness of that recurring failure mode. The loop records its own
pain but does not feed it forward into the one place — planning — where it is
cheapest to prevent.

## User Stories

- As an engineer starting a feature, I want the plan advisor to tell me which
  gates have recently bitten this repo, so I plan around them up front instead
  of rediscovering them at the `validate` gate.
- As a governance owner, I want the advisor to convert recurring ledger
  failures into pointed planning questions, so prevention is built into the
  plan artifact, not bolted on after a rejection.
- As a maintainer of a clean repo (empty/sparse ledger), I want zero noise —
  the advisor must behave exactly as today when there is nothing to warn about.

## Acceptance Criteria

1. When the ledger contains `gate-failure` events, the plan advisor's context
   summary includes a "Recurring gate failures" line listing the top-N gates by
   failure count (count shown, deterministic order: count desc, then gate name
   asc).
2. When at least one gate has recurred at/above a threshold, the advisor emits
   a pre-warning question naming that gate (subject to the existing
   `plan_question_limit` cap and lens tagging).
3. When the ledger is missing, empty, or has no `gate-failure` events, advisor
   output is byte-identical to current behaviour (no new line, no new question).
4. The feature respects `[telemetry] enabled = false` — a disabled ledger
   produces no failure context.
5. Aggregation reuses the existing insights gate-counting logic (no duplicated
   counting), so insights and the advisor never disagree on counts.
6. All new source files ≤100 lines; no cross-layer import violations; advisor
   stays read-only against the ledger.

## Edge Cases

- Missing ledger file → empty failure list, no panic (telemetry read is already
  lenient and returns `(nil, nil)`).
- Ledger present but only `block` / `step-advanced` events (no `gate-failure`)
  → no recurring-failure output.
- `gate-failure` events with an empty `Gate` field → rendered/handled exactly
  as insights handles them (`<none>` bucket) without crashing.
- Ties in failure count → stable alphabetical order so output is reproducible.
- Telemetry disabled in config → advisor must not read or surface ledger data.
- Very long gate names / many distinct gates → truncated to top-N; N bounded so
  the summary block stays concise.
- Headless mode → unchanged (advisor hook already exits silently when headless).

## Data Model

No new persisted schema. Read-only consumption of the existing
`telemetry.Event` (`internal/telemetry/event.go`), filtered to
`TypeGateFailure`, aggregated into the existing `insights.Count{Key,N}` shape.
The advisor context bundle (`internal/planadvisor/context.go`) gains one new
field: a small slice of `{Gate string, Count int}` (top-N recurring failures).

## Integration Points

- **Read**: `telemetry.ReadDefault()` (or feature-scoped read) for the event
  slice; `[telemetry] enabled` gate via `config`.
- **Aggregate**: existing insights gate-counting (exported/reused) →
  `[]Count`.
- **Surface**: `internal/planadvisor/context.go` (bundle), `context_summary.go`
  (new summary line), `questions.go` (new pre-warning question candidate).
- **Config**: optional new `[workflow] plan_advisor_failure_top_n` /
  threshold knobs, defaulted; reuses `plan_question_limit` for question cap.

## Risks

- **Import-graph**: `planadvisor` depending on `insights` must not create a
  cycle and must respect the archetype's layer rules — to be confirmed in the
  plan step (may instead lift the shared gate-counter into a lower package).
- **Determinism**: counts must be stably ordered; the plan validates this is
  inherited from insights' existing stable-sort.
- **Noise**: over-eager questions could regress the "clean repo = quiet"
  guarantee; thresholds + the existing question cap mitigate.
