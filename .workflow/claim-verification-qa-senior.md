# QA-Senior Report: claim-verification

**Date:** 2026-05-29
**Handoff to:** validation-specialist

## Test Inventory (by tier)

| Tier | Files | Tests |
|------|-------|-------|
| Unit (colocated) | `internal/verify/result_test.go`, `verify_test.go`, `runner_test.go`, `claim_tests_test.go`, `claim_coverage_test.go`, `claim_stubs_test.go`, `claim_edgecases_test.go`, `fakes_test.go`; `internal/config/verify_config_test.go`; `internal/evidence/coverage_field_test.go`; `internal/ui/render_verify_test.go`; `cmd/centinela/verify_cmd_test.go`, `verify_gate_test.go` | ~37 in verify + config/evidence/ui/cmd helpers |
| Integration | `tests/integration/claim_verification_test.go` | 2 (complete-gate block on fabricated claim; worktree-root resolution) |
| Acceptance | `tests/acceptance/claim_verification_test.go`, `claim_verification_more_test.go` | 6 (fabricated tests blocked, stub blocked, coverage overclaim blocked, unmapped edge WARNs/not blocks, honest evidence green, no-evidence skip) |

All checks use injected `CommandRunner` (fake/scripted) and `EvidenceLoader`
(fake) so no real shell-out or `.workflow` disk dependency leaks into the suite.

## Why colocated unit tests

`./scripts/check-coverage.sh` measures **per-package** coverage with no
`-coverpkg`. A `tests/`-tier test that *calls* `internal/verify` does not count
toward `internal/verify`'s coverage. Every new package
(`internal/verify`, the `Verify*` additions in `internal/config`,
`internal/evidence`, `internal/ui`, and `cmd/centinela`) therefore has
colocated `_test.go` files. Measured: `internal/verify` ≈ 97.4%; total gate
**95.2% ≥ 95.0%**.

## Coverage Gaps

None blocking. The execRunner timeout and start-error paths, the
default-evidence-loader branch, the coverage parse-error branch, and every
claim-check status transition are exercised. The `PriorTestRun`-reuse branch is
covered via `TestCheckTestsPassPriorRun`; full gate-reuse wiring remains a
senior-engineer TODO (perf only, no correctness gap).

## Acceptance Wiring

The acceptance tier is plain Go tests under `tests/acceptance/`, so
`go test ./...` (already in `validate.commands`) executes them — no separate
runner is needed. The scenarios map directly to `specs/claim-verification.feature`
(see the scenario comments above each `TestAcceptance_*`).

## Scaffold mirror sync

`internal/scaffold/assets/docs/architecture/evidence-contract.md` was stale
relative to the senior-engineer's source edit (new `coverage` field + global
rules), failing `TestScaffoldArchitectureMirrorParity` /
`TestScaffoldMirrorParityForUpdatedPrompts`. Synced the mirror to the source so
the parity gate is green. No content authored here — the mirror now tracks the
source verbatim.

## Handoff to validation-specialist

- All four gates green: `go build ./...`, `go vet ./...`, `go test ./...`
  (950 passed), `./scripts/check-coverage.sh` (95.2%).
- Gatekeeper still owes: confirm `internal/verify`'s dependency set against
  PROJECT.md G2 (new package, flagged by senior-engineer).
- Outstanding non-blocking TODO: wire `Deps.PriorTestRun` at the complete gate
  to avoid re-running the suite; document the `verify` knobs in the
  `centinela.toml` reference.
