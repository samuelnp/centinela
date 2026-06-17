# audit-baseline-ratchet — feature-specialist

## Behavior Summary

`centinela audit baseline` records a committed baseline of all current
violations (full-repo scan) across participating gates. `centinela audit`
re-scans and partitions violations into new (block, non-zero exit), baselined
(tolerate, exit 0), and resolved (prune on next record). Safe by default:
missing baseline / `severity=warn` / `enabled=false` never block.

## Acceptance Criteria (Gherkin)

`specs/audit-baseline-ratchet.feature` — 21 scenarios, 1:1 mapped to Go
acceptance tests via `// Scenario: <name>`, modeled on
`specs/centinela-insights.feature` (narrative Feature, Background, comment block,
exit-code + determinism rigor).

## UX States

CLI/text + `--json`. States: baseline recorded (confirmation + counts), clean
ratchet (`0 new`, exit 0), blocking ratchet (new violations named, non-zero),
no-baseline (advisory, non-blocking).

## Edge Cases

Record captures all (full scan); no-change tolerates; new blocks + named;
fixing never fails + prune; ratchet only tightens (pruned→reintroduced is new);
fingerprint stable across line-count churn; missing baseline non-blocking;
empty repo empty baseline; newly-enabled gate unbaselined; diff-aware still
full-scans; deterministic byte-identical re-record; config gating (off/warn,
custom path); deleted/renamed file resolves; multiple new violations all named.
(14 recorded in evidence `edgeCases`.)

## Out-of-Scope

A structured per-violation `Finding` refactor of every gate's `Result`
(v1 parses `Details` strings); auto-fix of violations; time-window filtering;
non-mechanical (subjective) gates.

## Handoff

→ senior-engineer. Implement per `docs/plans/audit-baseline-ratchet.md`; baseline
schema, fingerprint scheme, and the `cmd/`-wired gate seam are fixed there.
