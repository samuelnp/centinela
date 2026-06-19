# g2-multi-language-import-graph — senior-engineer

## Files Touched

**New leaf package `internal/importgraph`** (stdlib + os/exec + internal/golist):
provider.go (interface, Pkg/Graph, Runner, ErrNoProvider, ToolMissingError),
runner.go (default exec runner + onPath), manifests.go (detectKind, walks up to
the nearest manifest), select.go (provider resolution), go_backend.go +
go_scope.go (reference backend wrapping golist + scoping moved out of the gate),
node_backend.go + node_parse.go (dependency-cruiser/madge), python_backend.go +
python_script.go (embedded AST walker), script_backend.go (custom escape hatch),
graph_json.go (shared JSON contract for python + script).

**Gate (`internal/gates`)**: import_graph.go now dispatches through
loadGraph→importgraph.Select; import_graph_load.go is the provider adapter
(loadGraph + toPkgs); import_graph_classify.go maps load errors
(ErrNoProvider/ToolMissing→Warn, else Fail); import_graph_check.go lost the
go-specific scopePackages/stripModulePrefix (now in the leaf);
import_graph_glob.go dropped the unused errEmptyModule.

**Config**: ImportGraphConfig gains `Provider` + `ScriptCommand`, normalized.
**Docs**: centinela.toml (importgraph added to the leaf layer + provider example)
and PROJECT.md G2 prose updated.

## Architecture Compliance

n-tier respected: importgraph is a **leaf** (imports only stdlib, os/exec, and
the golist leaf — leaf→leaf is same-layer/allowed). gates (domain) → importgraph
(leaf) is an allowed domain→leaf edge; no cycle (golist/importgraph never import
gates). Dogfooded `pr-gate`: import_graph = **0 failed** (the pre-existing
non-failing "unmapped packages" Warn), importgraph correctly mapped, no
forbidden edge introduced. All source + test files ≤100 lines (G1 pass).

## Type-Safety Notes

No `interface{}`/`any` in logic; JSON decoded into typed structs (graphJSON,
depcruiseJSON, madge map). Errors are sentinels (`ErrNoProvider`) or typed
(`*ToolMissingError`, matched via errors.As) so the gate classifies precisely
rather than string-matching. The Runner seam is a typed func, injectable for
tests; onPath is a swappable var.

## Trade-Offs

- **Go behind the interface, not special-cased**: the existing scope/load unit
  tests moved into the leaf (go_scope_test.go); the valuable fixture-based
  integration tests in the gate stay and pass unchanged. Behavior is
  byte-identical for Go (verified: 2370 tests pass).
- **detectKind walks up** the tree (not CWD-only) to mirror how Go/Node/Python
  resolve a project root from a subdirectory — also fixed a real bug the gate
  tests caught (gate invoked from a package dir found no go.mod).
- **Tool-missing → Warn always** (not Fail-on-explicit) to keep CI green; the
  message names the remedy.
- **Node ships depcruise + madge**; one external tool, injected Runner keeps it
  unit-testable without the tool installed.

## Handoff

→ qa-senior: the leaf package needs colocated unit tests for the 95% per-package
coverage gate — select/manifests/node_parse/python_script(parse via
decodeGraphJSON)/script_backend/graph_json/runner, all fixture/fake-Runner
driven (no real tools). Add integration tests (real `go list`; `t.Skip`-guarded
node/python) and the acceptance suite for specs/g2-multi-language-import-graph.feature
+ the edge-cases artifact. Go path + dogfood already green.
