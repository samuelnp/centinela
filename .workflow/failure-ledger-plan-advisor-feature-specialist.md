# failure-ledger-plan-advisor — feature-specialist

## Behavior Summary

During the plan step the advisor reads the ledger and surfaces the top-N
recurring gate failures as a context-summary line plus one threshold-gated
pre-warning question. Quiet by default: missing/empty/disabled ledger or no
gate-failure events leaves output byte-identical to today.

## Acceptance Criteria (Gherkin)

`specs/failure-ledger-plan-advisor.feature` — 14 scenarios, 1:1 mapped to future
Go acceptance tests via `// Scenario: <name>`, modeled on
`specs/centinela-insights.feature` (narrative Feature, Background, comment block,
exit-code/determinism rigor).

## UX States

CLI/text advisor output only. States: quiet (no failure line/question), summary
present (ranked `gate (×N)` list), and pre-warning question present. No graphical
UI surface.

## Edge Cases

Missing ledger; empty ledger; only block/step-advanced events; telemetry
disabled; ranked count desc then name asc; ties broken alphabetically; empty
`Gate` → `<none>`; top-N cap; threshold gating of the question;
`plan_question_limit` cap; plan-step-only/headless unchanged; read-only (never
writes ledger); counts agree with `centinela insights`. (12 recorded in
evidence `edgeCases`.)

## Out-of-Scope

No new ledger fields, no per-feature attribution of failures (repo-wide only),
no changes to `centinela insights` output, no new graphical surface.

## Handoff

→ senior-engineer. Implement per `docs/plans/failure-ledger-plan-advisor.md`;
config knob names + defaults fixed there.
