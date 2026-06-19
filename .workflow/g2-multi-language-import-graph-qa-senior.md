# g2-multi-language-import-graph — qa-senior

## Test Inventory

**Unit (colocated, package-internal):** internal/importgraph — go_scope_test,
node_parse_test (fixture-driven depcruise + madge parsing, malformed cases),
node_backend_test + python_backend_test (fake Runner + swapped onPath: tool
present / fallback / tool-missing / runner-error / Name), script_graph_test
(script Load, nonzero exit, bad JSON, empty-valid graph, ToolMissingError
message), select_test (every branch: explicit go/node/python/script, empty
script_command, auto-detect, ErrNoProvider, unknown), manifests_test
(precedence, walk-up, dir-not-file), runner_test (execRunner success/nonzero/
missing-binary, firstStderrLine, onPath). internal/gates —
import_graph_classify_test (ErrNoProvider/ToolMissing → Warn, other → Fail; plus
a full-gate no-manifest self-skip). Existing gate suite adapted and green.

**Integration (tests/integration):** importgraph_integration_test — go backend
end-to-end via the real `go list` toolchain (always on CI); python AST walker
against a real package, `t.Skip`-guarded on python3.

**Acceptance (tests/acceptance):** g2_multi_language_import_graph_test (+ _steps)
drives the exported `gates.RunAll` over temp fixtures; carries
`// Acceptance: specs/g2-multi-language-import-graph.feature` + a `// Scenario:`
comment per the 7 spec scenarios. Go/script/no-manifest/empty-matrix run
unconditionally; node/python/tool-missing are skip-guarded on tool presence.

## Coverage Gaps

Per-package coverage: **importgraph 98.1%**, config 98.4%, gates 94.5% (the
import_graph files are ~fully covered; the package figure is dominated by other
gate files unchanged here). Aggregate `check-coverage.sh` gate: **95.1% ≥ 95.0%
PASS**. The only intentionally-uncovered line is the `module == ""` guard in the
go backend (go-list always returns a module or errors first).

## Acceptance Wiring

`centinela.toml` `validate.commands` already runs `go test ./tests/acceptance/...`,
so the new acceptance suite executes in the validate step. Spec-traceability gate
(warn mode) now sees a `// Scenario:` mapping for every scenario in the feature
file.

## Handoff

→ validation-specialist: full suite green (2370+ pre-existing + new), coverage
gate 95.1%, gofmt clean, dogfood `pr-gate` import_graph = 0 failed. Ready for the
gatekeeper + validate gate run.
