# spec-traceability-gate — qa-senior

**Date:** 2026-06-11

## Summary

Authored the unit, integration, and acceptance test suites for the
spec-traceability gate. All 10 Gherkin scenarios in
`specs/spec-traceability-gate.feature` map to genuine acceptance tests, closing
the dogfood: the gate now reports its own branch spec as **covered** (Pass).
Per-package coverage for the new `internal/gates` / `internal/config` code is
restored via colocated tests so the 95% gate passes (95.2% total). Every new
test file is <=100 lines (G1).

## Test Inventory

### Unit (colocated — drives per-package coverage)
| File | Lines | Covers |
|------|-------|--------|
| `internal/gates/spec_traceability_parse_test.go` | 93 | `parseScenarios` (Scenario + Outline, slug, diff include/exclude, missing-dir, dir-is-file error, non-feature + malformed tolerated), `scanScenarios` (unreadable), `normalizeScenario` |
| `internal/gates/spec_traceability_match_test.go` | 83 | `coveredScenarios` (header annotation, `spec/` typo, normalize, comment-without-header ignored, non-.go skipped, missing-dir empty, dir-is-file error), `uncovered` partition |
| `internal/gates/spec_traceability_entry_test.go` | 82 | `checkSpecTraceability` Pass/Fail+Details/Warn/Skip + parse-error + coverage-error Fail; `reportTraceability` empty-gaps Pass |
| `internal/config/spec_traceability_test.go` | 44 | `NormalizeSpecTraceability` defaults + trim; `validateSpecTraceability` reject-unknown / accept warn+fail / disabled no-op |

### Integration
| File | Lines | Covers |
|------|-------|--------|
| `tests/integration/spec_traceability_gate_integration_test.go` | 65 | `gates.RunWithFilter` end-to-end over a temp specs+acceptance tree: covered -> Pass, uncovered -> Fail naming the gap |

### Acceptance (dogfood closure — carries the `// Acceptance:` + 10 `// Scenario:` comments)
| File | Lines | Covers |
|------|-------|--------|
| `tests/acceptance/spec_traceability_gate_test.go` | 63 | Scenarios 1-5 (header + comments) |
| `tests/acceptance/spec_traceability_gate_more_test.go` | 64 | Scenarios 6-10 (header + comments) |
| `tests/acceptance/spec_traceability_helpers_test.go` | 58 | shared fixture/run helpers (no `// Scenario:` comments) |

## Coverage Gaps (10 scenario -> test mapping — must be none)

| # | Scenario | Acceptance test |
|---|----------|-----------------|
| 1 | A scenario with a matching acceptance test passes the gate | `TestSTG_MatchingTestPasses` |
| 2 | A scenario with no acceptance test fails the gate | `TestSTG_UncoveredFails` |
| 3 | Matching normalizes trailing period, spacing, and letter case | `TestSTG_NormalizationMatches` |
| 4 | An acceptance header with a trailing annotation still matches its spec | `TestSTG_HeaderAnnotationMatches` |
| 5 | A Scenario Outline counts as one covered scenario | `TestSTG_ScenarioOutlineCountsOnce` |
| 6 | Warn severity reports gaps without failing | `TestSTG_WarnSeverityDoesNotFail` |
| 7 | Diff-aware scope gates only changed spec files | `TestSTG_DiffAwareScopesChangedSpecs` |
| 8 | No spec files in scope skips the gate | `TestSTG_NoSpecsSkips` |
| 9 | An unknown severity value is rejected at config load | `TestSTG_UnknownSeverityRejected` |
| 10 | The gate is registered and enabled for Centinela in warn mode | `TestSTG_CentinelaEnablesWarn` |

No gaps: every scenario name (after normalization) equals a `// Scenario:`
comment above a real, asserting test func.

## Acceptance Wiring (from centinela.toml)

```toml
[gates.spec_traceability]
enabled  = true
severity = "warn"
```

`TestSTG_CentinelaEnablesWarn` loads the repo's own `centinela.toml`
(`../..` from the acceptance package) and asserts `Enabled == true` and
`Severity == "warn"`.

## Verification

- `gofmt -l cmd internal tests` -> empty (clean).
- `go vet ./...` -> No issues found.
- `go test ./...` -> 1269 passed, 0 failed (24 packages).
- `./scripts/check-coverage.sh` -> `coverage gate passed: 95.2% >= 95.0%`.
- **Dogfood proof** (`/tmp/cent-stg2 validate`):
  `✓ spec-traceability-gate  All 10 scenarios have acceptance coverage.`
  and overall `All gates passed.`

## Handoff -> validation-specialist

- All three test tiers exist and pass; coverage gate green; dogfood closed.
- The gate's own spec is now self-covering, so diff-aware validate is clean.
- import_graph emits a pre-existing, unrelated WARN (test-fixture packages match
  no layer) — not introduced by this feature.
- **Next role:** validation-specialist (gatekeeper report + final `centinela
  validate` run).
