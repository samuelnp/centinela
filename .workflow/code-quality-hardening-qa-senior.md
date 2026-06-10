# code-quality-hardening — qa-senior

## Test Inventory

| Tier | File | Lines | Covers |
|------|------|-------|--------|
| unit | internal/hookpolicy/format_evidence_parity_test.go | 62 | `TestEvidenceKeyOrderParity` — byte-parity vs `evidence.MarshalJSON`, coverage between mobileFirst/handoffTo |
| unit | internal/workflow/state_load_missing_test.go | 21 | missing state file → "no workflow found" |
| unit | internal/workflow/state_load_corrupt_test.go | 36 | invalid JSON → names path + wraps parse cause, not absence |
| unit | internal/workflow/state_load_unreadable_test.go | 40 | chmod 000 (root-skip) → names path, not absence |
| unit | internal/workflow/active_warn_test.go | 42 | `ActiveWorkflows` emits stderr `workflow warning:` for corrupt state |
| unit | cmd/centinela/start_corrupt_config_test.go | 33 | corrupt toml → error names centinela.toml, no `.workflow/<f>.json` |
| unit | cmd/centinela/hook_context_corrupt_config_test.go | 30 | corrupt toml → nil error (exit 0) + `config warning:` injected |
| integration | tests/integration/code_quality_hardening_test.go | 45 | cross-package: Load missing-vs-corrupt + formatter parity w/ coverage |
| acceptance | tests/acceptance/code_quality_hardening_test.go | 70 | check-fmt.sh fail/pass + centinela.toml validate wiring |

All files ≤100 lines (G1). Colocated unit tests live in the same package as the
code they cover, so they move the per-package 95% coverage gate (no -coverpkg).

## Coverage Gaps

None. The 9 Gherkin scenarios map 1:1 to executable assertions:

1. Hook formatter preserves canonical key order → `TestEvidenceKeyOrderParity` (also `TestEvidenceFormatterParityWithCoverage`)
2. Unformatted Go source fails the format check → `TestUnformattedSourceFailsFormatCheck`
3. Formatted tree passes the format check → `TestFormattedTreePassesFormatCheck`
4. Validate suite gates formatting → `TestValidateSuiteGatesFormatting`
5. Starting a feature with corrupted config fails loudly → `TestRunStartCorruptConfigFailsAndWritesNoState`
6. Prompt hook degrades with a warning on corrupted config → `TestRunHookContextCorruptConfigWarnsAndExitsZero`
7. Loading a missing workflow reports absence → `TestLoadMissingReportsAbsence` (+ `TestLoadDistinguishesMissingFromCorrupt`)
8. Loading a corrupted workflow reports the cause → `TestLoadCorruptReportsPathAndCause`
9. Loading an unreadable workflow is not reported as absence → `TestLoadUnreadableIsNotAbsence`

Plus the senior-engineer's `ActiveWorkflows` change is covered by
`TestActiveWorkflowsWarnsOnCorruptStateFile` (beyond the spec, guards the
silent-drop regression).

## Acceptance Wiring

`centinela.toml` `[validate] commands` (asserted by `TestValidateSuiteGatesFormatting`):

```toml
[validate]
commands = [
  "go test ./...",
  "...acceptance...",
  "./scripts/check-coverage.sh",
  "./scripts/check-fmt.sh"
]
```

Acceptance tests read repo files via `filepath.Join("..","..",...)` and run the
script with `exec.Command("sh", script)` + `FMT_DIRS` override.

## Verification

- `gofmt -l cmd internal tests`: empty (clean)
- `go vet ./...`: no issues
- `go build ./cmd/centinela`: success
- `go test ./...`: 1232 passed in 24 packages
- `./scripts/check-coverage.sh`: coverage gate passed: 95.1% >= 95.0%

## Handoff

- Next role: validation-specialist
