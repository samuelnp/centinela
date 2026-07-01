# roadmap-json-contract — qa-senior

## Test Inventory

| Tier | File | Lines | Covers |
|------|------|-------|--------|
| colocated (internal/roadmap) | `internal/roadmap/view_test.go` | 75 | BuildView status/readiness/blockedBy mapping, declared order, counts, dependsOn always non-nil |
| colocated (internal/roadmap) | `internal/roadmap/view_edge_test.go` | 61 | byte-stable marshal, empty & nil roadmap → exact contract bytes, Backlog/Baseline exclusion |
| colocated (cmd/centinela) | `cmd/centinela/roadmap_json_test.go` | 79 | `roadmap --json` shape/counts, `ready --json` == view ready-set |
| colocated (cmd/centinela) | `cmd/centinela/roadmap_json_edge_test.go` | 99 | `ready --json` empty `[]`, `show --json` verbatim (no derived fields), `show` text == `roadmap` text, missing-file no-partial-JSON for all 3 surfaces |
| unit | `tests/unit/roadmap_json_contract_unit_test.go` | 49 | BuildView projection (done/ready/blocked), Backlog exclusion, counts |
| integration | `tests/integration/roadmap_json_contract_integration_test.go` | 49 | Load-from-disk → BuildView excludes non-schedulable + byte-stable |
| acceptance | `tests/acceptance/roadmap_json_contract_helpers_test.go` | 85 | shared build-once binary + project/seed/run helpers (no installed binary) |
| acceptance | `tests/acceptance/roadmap_json_contract_test.go` | 96 | full `roadmap --json` contract, `ready --json` == readiness:ready set, determinism |
| acceptance | `tests/acceptance/roadmap_json_contract_edge_test.go` | 64 | `show/list --json` verbatim + alias, `show` text, empty & zero-feature phases |
| acceptance | `tests/acceptance/roadmap_json_contract_error_test.go` | 57 | missing-file (3 json + text), malformed JSON, dependency cycle |

All files ≤100 lines (G1 satisfied). Acceptance drives a binary built from `./cmd/centinela` into a temp dir — no installed-binary or network dependence.

## Coverage Gaps

- Global coverage gate: **97.4% ≥ 95.0%** (PASS, ~2.4% above floor).
- Per-package: `internal/roadmap` **97.5%**, `cmd/centinela` **96.3%** (up from 95.2% baseline).
- New code: `internal/roadmap/view.go` **100%** (all funcs). New cmd handlers fully covered except the unreachable `json.MarshalIndent` error returns (`runRoadmap` 90.9%, `runRoadmapReady` 92.9%, `runRoadmapShow` covered on both branches) — defensive checks with no forceable failure path.

## Acceptance Wiring

`centinela.toml` already runs `go test ./tests/acceptance/...` (unchanged). Spec-traceability comments map scenarios 1:1, e.g.:

```go
// Acceptance: specs/roadmap-json-contract.feature
// Scenario: roadmap --json emits ordered phases and features with counts
func TestRoadmapViewJSONFullContract(t *testing.T) { ... }
```

24 of 28 spec scenarios carry matching `// Scenario:` comments under an `// Acceptance:` header (gate severity is `warn`). The 4 not annotated are the two text-mode "byte-for-byte unchanged from today" regression scenarios (no stored golden to compare against) and their internal-outline variants; behavior is still exercised (text output asserted JSON-free and `show`==`roadmap`).

## Handoff

- **Next role:** validation-specialist
- `go test ./...` all pass; `go vet ./...` clean; `gofmt -l` clean on all new files.
- Coverage gate passes at 97.4%. No production code modified. No deferred findings.
