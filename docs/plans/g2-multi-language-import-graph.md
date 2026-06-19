# Plan: g2-multi-language-import-graph

> Make the G2 import-graph gate language-aware. Reconciles the big-thinker
> architecture with the feature-specialist file plan. **Decision:**
> `internal/importgraph` is a **leaf** package (stdlib + `os/exec` + the
> `internal/golist` leaf only) with its own minimal manifest detector — this
> avoids a domain→domain edge to `internal/analyze` and keeps `internal/gates`
> importing leaves only. No cycle.

## Architecture

A new leaf package `internal/importgraph` owns the provider abstraction, the
backends, provider selection, and an injectable `Runner` exec seam. Every
backend returns an already-scoped, module-relative `Graph{Module, Pkgs[]Pkg}`
so the gate's existing `checkEdges`/matrix logic runs unchanged on any
provider's output.

```go
type Pkg struct { Path string; Imports []string } // module-relative
type Graph struct { Module string; Pkgs []Pkg }
type GraphProvider interface { Name() string; Load(root string) (Graph, error) }
type Runner func(name string, args ...string) ([]byte, error) // injectable
var ErrNoProvider error  // → gate self-skips with WARN
var ErrToolMissing error // → gate WARNs (CI without the tool stays green)
```

**Selection.** `Select(root, provider string, scriptCmd []string, run Runner)`:
explicit `provider` wins; else `detectKind(root)` maps a manifest to a backend;
`script` is never auto-selected; no match → `ErrNoProvider`. node/python missing
their PATH tool → `ErrToolMissing`.

**Gate wiring.** `checkImportGraph` resolves a provider, calls `Load`, and
classifies errors: `ErrNoProvider`/`ErrToolMissing` → **Warn** (non-failing);
any other load error → **Fail** (preserves the uncompilable-Go behavior). On
success it runs the unchanged `checkEdges` against the layer matrix.

**Custom-script contract.** The configured argv runs at the project root and
must emit the shared JSON: `{"module":"…","pkgs":[{"path":"…","imports":["…"]}]}`.
Non-zero exit → Fail; valid empty graph → Pass.

## Source files (each ≤100 lines, n-tier respected)

### config (leaf — stdlib only)
1. `internal/config/import_graph.go` (modify) — add `Provider string` and
   `ScriptCommand []string`; extend `NormalizeImportGraph` (trim + lowercase
   provider, trim script argv). Unset fields → empty (backward compatible).
2. `internal/config/import_graph_provider.go` (new) — `normalizeProvider` +
   script-argv trim helper, if (1) would exceed 100 lines.

### internal/importgraph (NEW leaf — stdlib, os/exec, internal/golist)
3. `provider.go` — `Pkg`, `Graph`, `GraphProvider`, `Runner`, sentinels.
4. `runner.go` — default `exec.Command` runner + `lookPath` helper.
5. `manifests.go` — `detectKind(root)` file-stat detector (go > node > python).
6. `select.go` — `Select(...)` provider resolution + tool-missing detection.
7. `go_backend.go` — `goProvider` wrapping `golist`, reference backend.
8. `go_scope.go` — `scopeGoPkgs` (module-prefix strip, fold test imports, drop
   stdlib/3rd-party, dedupe) — moved out of the gate.
9. `node_backend.go` — `nodeProvider{run}`: shell depcruise/madge, build Graph.
10. `node_parse.go` — pure `parseNodeGraph`/`parseMadge([]byte) ([]Pkg, error)`.
11. `python_backend.go` — `pythonProvider{run}`: run embedded AST walker.
12. `python_script.go` — embedded `astScript` const + `parsePythonGraph`.
13. `script_backend.go` — `scriptProvider{cmd, run}`: run argv, decode JSON.
14. `graph_json.go` — shared JSON DTO + `decodeGraphJSON` (python + script).

### gates (domain — imports leaves only; now imports internal/importgraph)
15. `internal/gates/import_graph_load.go` (modify) — `loadGraph(cfg)` calls
    `importgraph.Select` + `provider.Load`; `toPkgs` adapter to local `pkg`.
16. `internal/gates/import_graph.go` (modify) — `checkImportGraph` calls
    `loadGraph`, delegates error mapping to `classifyLoadError`, runs `checkEdges`.
17. `internal/gates/import_graph_classify.go` (new) — `classifyLoadError`.
18. `internal/gates/import_graph_check.go` (modify) — drop `scopePackages`
    (moved to leaf); `checkEdges` consumes adapted `pkg` (byte-identical tests).

### config/docs
19. `centinela.toml` (modify) — add `internal/importgraph/**` to the `leaf`
    layer paths + a commented `provider`/`script_command` example.
20. `PROJECT.md` (modify) — G2 prose: language-aware gate + new leaf package.

## Test plan (coverage is per-package → colocated `_test.go`, each ≤100)

- **Unit:** `config/import_graph_test.go` (normalize); `importgraph` package —
  `select_test`, `manifests_test`, `go_scope_test`, `node_parse_test`,
  `python_script_test`, `script_backend_test`, `node_backend_test`,
  `python_backend_test` (all fixture/fake-`Runner` driven, no real tools);
  `gates/import_graph_classify_test`; existing `gates/import_graph_*_test`
  adapted + a forced-`script`-provider forbidden-edge Fail.
- **Integration (`tests/integration/`):** `importgraph_go_test` (real `go list`
  fixture module, always-on CI); `importgraph_python_test` /
  `importgraph_node_test` (`t.Skip` unless the tool is present).
- **Acceptance (`tests/acceptance/`):** `g2_multi_language_import_graph_test.go`
  (+ `_steps.go` helpers) driving `centinela validate` over temp fixtures: Go
  Pass, no-manifest WARN+exit0, script-provider enforced, script forbidden-edge
  Fail. Carries `// Acceptance:` / `// Scenario:` traceability comments.

External tools may be absent on CI: all shell-out backends take an injectable
`Runner`; only the Go path uses the real toolchain (always present).

## Spec

`specs/g2-multi-language-import-graph.feature` — 7 scenarios (Go enforced, Node
enforced, Python enforced, no-manifest skip/WARN, custom-script enforced,
tool-missing WARN, empty-matrix WARN).

## Risks

- Shelling out to external parsers (depcruise/python) → mitigated by the
  injectable `Runner` + skip-guarded integration tests.
- Per-package coverage gate (95%) → each new file gets colocated unit tests
  ≤100 lines; pure parse functions carry the bulk of coverage.
- ≤100-line rule → aggressive file splits already encoded above (14 source +
  ~12 test files).
