# lean-evidence-footprint — qa-senior

## Test Inventory

| Tier | File | Asserts |
|------|------|---------|
| unit | `tests/unit/lean_evidence_footprint_unit_test.go` | shipped `.gitignore` contains the 3 patterns; negation line follows the json ignore |
| integration | `tests/integration/lean_evidence_footprint_integration_test.go` | `git check-ignore` over a temp repo seeded with the real `.gitignore`: json/lock ignored, roadmap.json + md not |
| acceptance | `tests/acceptance/lean_evidence_footprint_test.go` | `git status`/`diff --cached` ignore matrix; retroactive `git rm --cached` untracks yet keeps the local file |

All three tiers read the **repo's own** `.gitignore`, so removing the
patterns fails the suite. All files ≤100 lines (G1 for test files).

## Coverage Gaps

None relevant: the feature adds no `internal/`/`cmd/` source, so the
per-package 95% gate is unaffected. The behavior under test is git's, driven
through real `git` subprocesses rather than mocks.

## Acceptance Wiring

`specs/lean-evidence-footprint.feature` scenarios map to
`TestAccEvidenceIgnoreMatrix` (ignore/keep matrix) and
`TestAccRetroactiveUntrack` (retroactive untrack). `centinela.toml`
`validate.commands` already runs `go test ./tests/acceptance/...`.

## Handoff

→ validation-specialist: run full suite + gates; produce the gatekeeper
report. Expect no coverage delta and no file-size violations.
