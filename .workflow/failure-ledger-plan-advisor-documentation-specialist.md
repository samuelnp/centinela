# failure-ledger-plan-advisor — documentation-specialist

Internal-surface (right-sized) docs step.

## KB Pages

No new knowledge-base page — this extends the existing plan-advisor surface
rather than adding a new command. Covered by the regenerated project docs.

## project-docs Entries

- `.workflow/failure-ledger-plan-advisor-changelog.md` — one-line `feat`
  changelog entry.
- Regenerated `docs/project-docs/index.html` (picks up the new feature brief,
  plan, and changelog).

## User-facing note

During the `plan` step the plan advisor now reads the governance-telemetry
ledger (`.workflow/telemetry/events.jsonl`) and surfaces the repo's most
recurring gate failures: a "recurring gate failures" context line listing the
top-N gates as `gate (×N)` (count desc, then name asc), and — when a gate has
recurred at or above the threshold (3) — one pre-warning planning question
naming the worst gate. The list size is configurable via `[workflow]
plan_advisor_failure_top_n` (default 3, cap 5). It reuses the same counter as
`centinela insights`, so counts never diverge, and is read-only and quiet by
default: a missing/empty ledger, no gate-failure events, or `[telemetry]
enabled = false` leaves advisor output byte-identical to before.

## Outcome

Docs generated and validated. Handoff → complete.
