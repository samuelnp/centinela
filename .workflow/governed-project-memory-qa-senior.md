### QA-Senior Report: governed-project-memory
**Date:** 2026-05-30

#### Test Inventory
| Tier        | File | Scenarios |
|-------------|------|-----------|
| unit        | `internal/memory/entry_test.go`, `serialize_test.go`, `dedupe_test.go`, `parse_test.go`, `parse_helpers_test.go`, `capture_test.go`, `capture_more_test.go`, `index_test.go`, `recall_test.go`, `rank_test.go`, `rank_extra_test.go` | SC-01..05, SC-07..12 (content-hash, parsers, dedupe, ranking, caps, config gating) |
| unit        | `internal/config/memory_test.go` | config defaults + disabled (SC-12) |
| unit        | `internal/planadvisor/memory_test.go` | recall surfaced in plan-advisor bundle (SC-08) |
| unit        | `internal/ui/render_memory_test.go` | MEMORY render block |
| integration | `tests/integration/governed_project_memory_integration_test.go` | capture→entry, idempotence (SC-01/05) |
| integration | `tests/integration/governed_project_memory_recall_test.go` | plan-advisor recall path, concurrent capture (SC-08/10, SC-13) |
| acceptance  | `tests/acceptance/governed_project_memory_test.go` (SC-01/02), `_b_test.go` (SC-03/04/05), `_c_test.go` (SC-06/07/08) | all 13 Gherkin scenarios mapped |

#### Coverage Gaps
- None blocking. Full suite: **975 tests pass**. Coverage gate: **95.1% ≥ 95.0%**.
- Per-package on new code: `internal/memory` 94.7%, `internal/config` 96.8%,
  `internal/planadvisor` 94.7%, `internal/ui` 97.0%.
- Colocated `_test.go` files were used (not just `tests/`-tier) because Go
  measures coverage per package — `tests/` files do not move the gate for
  `internal/...`/`cmd/...` code.

#### Acceptance Wiring
`centinela.toml` `[validate].commands` runs `go test ./...`, which includes
`tests/acceptance/` (the acceptance package is a normal Go test package):
```toml
[validate]
commands = [
  "go test ./...",
  "./scripts/check-coverage.sh"
]
```

#### Recovery note
The qa-senior run was interrupted (session limit) mid-step. The orchestrator
finished the step: fixed unused-import build errors in the acceptance files,
split four test files that exceeded the 100-line G1 limit
(`parse_test.go`→`parse_helpers_test.go`, `capture_test.go`→`capture_more_test.go`,
acceptance `_b`→`_c`, integration →`_recall`), removed dead `buildBin` helper,
authored the edge-cases report, and produced this evidence. No implementation
logic was changed to make tests pass.

#### Handoff
- Next role: validation-specialist
- Edge-case report: `.workflow/governed-project-memory-edge-cases.md`
