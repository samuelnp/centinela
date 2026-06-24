# adoption-baseline — qa-senior

> Authored directly by the orchestrator after the delegated qa-senior subagent died on a socket
> error having written only stubs (empty evidence, no tests). The bounded remainder — the full test
> suite — was completed here and every gate re-verified independently.

## Test Inventory

| Tier | File | Scenarios / seams |
|------|------|-------------------|
| colocated unit | `internal/audit/adopt_test.go` | fresh adopt records; skip-if-exists no write (byte-unchanged); --force widens |
| colocated unit | `internal/audit/adopt_more_test.go` | byte-identical to Record+Save; clean repo 0 findings; Load-error propagation; `Baseline.Total()` |
| colocated unit | `internal/ui/render_adopt_test.go` | per-gate counts + total + ratchet framing; 0-finding "nothing to ratchet" |
| colocated cmd | `cmd/centinela/adopt_test.go` | happy report; skip exits non-zero + byte-unchanged; --force; config error |
| colocated cmd | `cmd/centinela/adopt_json_test.go` | `--json` adopt verdict; `--json` skip verdict (`per_gate:{}`, non-zero, byte-unchanged) |
| tests/unit | `tests/unit/adoption_baseline_unit_test.go` | record-then-skip lifecycle over a real temp repo |
| tests/integration | `tests/integration/adoption_baseline_integration_test.go` | adopt → ratchet reports 0 new; gate not Fail |
| tests/acceptance | `tests/acceptance/adoption_baseline_test.go` | Scenarios 1, 2, 6, 8 |
| tests/acceptance | `tests/acceptance/adoption_baseline_edge_test.go` | Scenarios 3, 4, 5, 7 |

All 8 spec scenarios carry exact-match `// Scenario:` comments across the two acceptance files
(both carry `// Acceptance: specs/adoption-baseline.feature`). Every new `_test.go` is ≤79 lines.

## Coverage Gaps

None. The audit/ui/cmd statements for adopt are covered by colocated tests (per-package gate, no
`-coverpkg`). The literal `Record`+`Save` byte-identity is asserted in the colocated `internal/audit`
test; the acceptance tier adds a binary-level determinism proxy.

## Acceptance Wiring

`centinela.toml` `[validate].commands` already runs the acceptance tier + coverage — not edited:
`go test ./...`, `go test ./tests/acceptance/...`, `./scripts/check-coverage.sh`, `./scripts/check-fmt.sh`.

## Handoff

Next role: validation-specialist. Full suite green, coverage gate green, all 8 scenarios traced,
`.workflow/adoption-baseline-edge-cases.md` filled, evidence JSON validates.
