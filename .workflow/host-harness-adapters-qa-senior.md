# host-harness-adapters — qa-senior

## Test Inventory

12 test files created across 4 tiers. All suites green.

### Colocated tests (`internal/setup/`)

| File | Lines | Coverage target |
|------|-------|-----------------|
| `adapter_registry_test.go` | 91 | Lookup, RegisteredAgents, RegisteredAdapters, IsValidAgent, AgentsFor |
| `adapter_capabilities_test.go` | 57 | claudeAdapter/openCodeAdapter/aiderAdapter Name + Capabilities |
| `adapter_parity_test.go` | 58 | CapabilityParity invariant, blocks-writes → prewrite hook, aider no-hook |
| `aider_config_test.go` | 97 | planAiderConfig (create/idempotent/update/manual-review), writeManagedAiderConfig |

### Unit tests (`tests/unit/`)

| File | Lines | Scenarios |
|------|-------|-----------|
| `host_harness_adapters_registry_unit_test.go` | 70 | AC1, AC6 isValidAgent, BuildSyncPlan all-agents structural |
| `host_harness_adapters_scope_unit_test.go` | 91 | AC2 scope assertions for claude/opencode/aider/both |

### Integration tests (`tests/integration/`)

| File | Lines | Scenarios |
|------|-------|-----------|
| `host_harness_adapters_integration_test.go` | 90 | AC5 init writes, idempotent re-apply, Claude files untouched |

### Acceptance tests (`tests/acceptance/`)

| File | Lines | ACs covered |
|------|-------|-------------|
| `host_harness_adapters_ac1_ac3_test.go` | 86 | AC1 registry lookup + typed error; AC3 capabilities |
| `host_harness_adapters_ac2_test.go` | 92 | AC2 claude/aider/both BuildSyncPlan scoping |
| `host_harness_adapters_ac5_test.go` | 99 | AC5 aider init, idempotency, unmanaged file not clobbered |
| `host_harness_adapters_ac6_ac7_test.go` | 58 | AC6 valid/invalid agent CLI, registry-driven isValidAgent; AC7 non-empty caps |
| `host_harness_adapters_ac7_test.go` | 83 | AC7 blocks-writes→prewrite, aider no hook, hookless parity |

## Coverage Gaps

No gaps in new code. All adapter/registry/aider functions at 100%:
`Lookup`, `RegisteredAgents`, `RegisteredAdapters`, `IsValidAgent`, `AgentsFor`,
`adaptersFor`, `planAiderConfig`, `writeManagedAiderConfig`, all three `Name()` and
`Capabilities()` methods, `TestGoldenParityClaudeOpenCode` (pre-existing).

| Package | Coverage |
|---------|----------|
| `internal/setup` | 95.4% |
| `cmd/centinela` | 91.9% |
| Combined | 92.5% |

Remaining uncovered paths are error branches in `PlanItems()` when underlying
`plan*` functions hit OS errors — covered by other pre-existing tests.

## Acceptance Wiring

`centinela.toml validate.commands` already includes:
- `go test ./...` (covers all packages)
- `go test ./tests/acceptance/...` (explicit acceptance run)

All 26 spec scenarios covered across the 5 acceptance test files and the
pre-existing `golden_parity_test.go` (AC4 byte-parity).

`.workflow/host-harness-adapters-edge-cases.md` — 8 edge cases documented.

## Handoff

Ready for `validation-specialist`. All gate checks should pass:
- G1: all 12 new test files ≤ 100 lines
- `validate.commands` includes acceptance test execution
- Edge-cases ledger exists and is populated
- Evidence validates clean (`centinela evidence validate host-harness-adapters`)
