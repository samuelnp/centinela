# QA-Senior Report: configurable-subagent-models

## Files Written (line counts)

### Flat test files (tests/ directory — house convention)
| File | Lines | Tests |
|------|-------|-------|
| `tests/unit/configurable_subagent_models_unit_test.go` | 93 | 6 |
| `tests/unit/configurable_subagent_models_resolve_unit_test.go` | 66 | 5 |
| `tests/unit/configurable_subagent_models_config_unit_test.go` | 99 | 8 |
| `tests/integration/configurable_subagent_models_integration_test.go` | 87 | 4 |
| `tests/acceptance/configurable_subagent_models_test.go` | 91 | 3 |
| `tests/acceptance/configurable_subagent_models_config_test.go` | 65 | 4 |

### Package-level test files (required for coverage counting)
| File | Lines | Tests |
|------|-------|-------|
| `internal/orchestration/models_test.go` | 87 | 6 |
| `internal/orchestration/resolve_test.go` | 89 | 8 |
| `internal/config/orchestration_models_test.go` | 100 | 6 |

**Total: 777 lines across 9 files, 50 test functions. All files ≤100 lines.**

## go test Result

```
go test ./... → PASS (all packages except 1 pre-existing failure)
```

Pre-existing failure (NOT introduced by this feature):
- `TestScaffoldMirrorParityForUpdatedPrompts` — `docs/architecture/documentation-generator-prompt.md` was modified before this tests step (shown in git status at session start). Confirmed by running without our test files: same failure.

## Coverage Total

**95.0%** (up from 93.8% before adding internal package tests).

Per-package for new code:
- `internal/orchestration`: 96.2% (was 80.7% before adding `models_test.go` + `resolve_test.go`)
- `internal/config`: 96.4% (was 88.0% before adding `orchestration_models_test.go`)

## Edge Cases Covered

1. Tier normalization — casing (`"Reasoning"` → `"reasoning"`)
2. Tier normalization — whitespace (`" fast "` → `"fast"`)
3. Empty vs absent `[orchestration.models]` table — both valid
4. Unknown role key rejected with precise error naming the key
5. Invalid tier after normalization (`" Genius "`) rejected
6. Missing tier→model mapping no-panic (returns tier name + ok=false)
7. Unknown runner emits tier name as fallback, ok=false
8. Out-of-band roles not annotated in directive
9. Allow-list parity: all `AllowedTiers()` and `AllowedRoleSlugs()` values accepted by `config.Load()`
10. `ModelReference` deduplication (duplicate tiers → single entry)
11. `ModelReference` stable order (reasoning before balanced before fast)
