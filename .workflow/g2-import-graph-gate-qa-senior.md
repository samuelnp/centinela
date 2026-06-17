### QA-Senior Report: g2-import-graph-gate

**Date:** 2026-06-05

Test suite for the G2 import-graph gate across all three tiers plus colocated
coverage tests. Total module coverage lifted from 91.9% to **95.3%** (gate
threshold 95.0%). All 1097 tests pass. Every new `_test.go` is <=100 lines (G1).

#### Test Inventory

| Tier | File | Lines | What it covers |
|------|------|------:|----------------|
| unit (colocated) | internal/config/import_graph_test.go | 75 | NormalizeImportGraph trim/preserve-empty, TOML decode of `[gates.import_graph]`, applyDefaults module normalization |
| unit (colocated) | internal/config/defaults_test.go | 39 | applyDefaults branches: FileSize default flip, I18n-on case, import-graph layer normalization |
| unit (colocated) | internal/gates/import_graph_test.go | 69 | checkImportGraph orchestrator: empty-matrix Warn, config-error Fail, resolveModule, Pass + Warn(unmapped) on the real module |
| unit (colocated) | internal/gates/import_graph_matrix_test.go | 79 | buildMatrix validation + duplicate-union, layerFor, globMatch (path.Match + `/**` + bare `**`) |
| unit (colocated) | internal/gates/import_graph_check_test.go | 75 | checkEdges: allowed/forbidden, arrow message format, unmapped + ignored edges, sorted+deduped |
| unit (colocated) | internal/gates/import_graph_scope_test.go | 43 | stripModulePrefix boundary cases, scopePackages (fold test imports, drop stdlib/self/third-party) |
| unit (colocated) | internal/gates/import_graph_glob_test.go | 53 | trimDoubleStar, hasPrefixDir, allowed() same-layer/allow-list, errEmptyModule |
| unit (colocated) | internal/gates/import_graph_load_test.go | 99 | real `go list` against fixture modules: forbidden-edge Fail, broken-go.mod load-error Fail, resolveModule discovery error, loadModulePath |
| unit (tier) | tests/unit/g2_import_graph_gate_unit_test.go | 63 | gate result semantics via gates.RunWithFilter: disabled-omitted, empty-Warn, malformed-config Fail (no arrow) |
| integration | tests/integration/g2_import_graph_gate_integration_test.go | 81 | end-to-end real `go list` against on-disk fixture: clean Pass, forbidden-edge Fail with arrow |
| acceptance | tests/acceptance/g2_import_graph_gate_test.go | 95 | EXECUTABLE: clean Pass; forbidden edge Fail + edge string asserted; malformed config Fail distinct from a violation. Mapped to `.feature` scenarios in comments |

#### Coverage Gaps

Final total coverage **95.3% >= 95.0%**. Two near-unreachable branches remain
uncovered and are documented in the edge-cases residual risks:
- `loadPackages` JSON-decode error branch (requires `go list` to emit malformed
  JSON with a zero exit — not produced by the toolchain).
- `resolveModule` `errEmptyModule` return (requires `go list -m` to print an
  empty module path with a zero exit).
Neither can cause a false Pass: any real `go list` failure exits non-zero and is
already surfaced as Fail.

#### Acceptance Wiring

`centinela.toml` validate.commands now includes a command matching
`./tests/acceptance/...` (required by the tests-step gate; `go test ./...` does
not qualify):

```toml
[validate]
commands = [
  "go test ./...",
  "go test ./tests/acceptance/...",
  "./scripts/check-coverage.sh"
]
```

#### Verification

- `go test ./...` -> 1097 passed.
- `./scripts/check-coverage.sh` -> "coverage gate passed: 95.3% >= 95.0%".
- All 11 new `_test.go` files <=100 lines.
- `centinela evidence validate g2-import-graph-gate` -> "evidence ok".

#### Handoff to validation-specialist

Implementation is fully tested across unit/integration/acceptance with the
coverage gate green. validate.commands runs the acceptance tier explicitly. Edge
cases documented in `.workflow/g2-import-graph-gate-edge-cases.md`. Ready for the
validate step: run the gatekeeper report and `centinela validate`.
