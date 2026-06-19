# Edge Cases: g2-multi-language-import-graph

## Covered

- **No recognized manifest** → `ErrNoProvider` → gate self-skips with a
  non-failing Warn (the non-Go hard-fail fix). Tests: `select_test`
  (`TestSelect_NoProvider`), `import_graph_classify_test`
  (`TestCheckImportGraph_NoManifestSelfSkips`), acceptance
  (`TestAccG2_NoManifestSkips`).
- **Multiple manifests** (e.g. go.mod + package.json) → deterministic precedence
  go > node > python; explicit `provider` overrides. Test: `manifests_test`
  (`TestDetectKind`).
- **Manifest in an ancestor dir** (gate/test invoked from a subdirectory) →
  detection walks up. Test: `manifests_test` (`TestDetectKind_WalksUp`).
- **Explicit `provider="script"` with empty `script_command`** → config error,
  not a silent skip. Test: `select_test` (`TestSelect_Script`).
- **Unknown provider name** → error. Test: `select_test`
  (`TestSelect_UnknownProvider`).
- **External tool missing** (depcruise/madge/python3 absent) → `ToolMissingError`
  → Warn, exit non-failing (CI green). Tests: `node_backend_test`
  (`TestNodeProvider_ToolMissing`), `python_backend_test`
  (`TestPythonProvider_ToolMissing`), acceptance (`TestAccG2_NodeToolMissingWarns`).
- **External tool runs but errors / non-zero exit** → load error → Fail (never a
  false Pass). Tests: `node_backend_test` (`TestNodeProvider_RunError`),
  `script_graph_test` (`TestScriptProvider_RunError`).
- **Malformed tool output** (bad JSON from depcruise/madge/script) → parse error
  → Fail. Tests: `node_parse_test` (`*_Malformed`), `script_graph_test`
  (`TestScriptProvider_BadJSON`).
- **Custom-script valid empty output** → empty graph → Pass. Test:
  `script_graph_test` (`TestDecodeGraphJSON_EmptyValid`).
- **node_modules / core modules / self / duplicate edges** dropped in the node
  parsers. Tests: `node_parse_test`.
- **Empty layer matrix** → Warn before any provider is selected. Tests:
  `import_graph_test`, acceptance (`TestAccG2_EmptyMatrixWarns`).
- **Go-repo parity** (provider unset → auto-select go, byte-identical) → the
  full pre-existing gate suite stays green; dogfood `pr-gate` = 0 failed.
- **Broken go.mod** → the go backend Fails with the load diagnostic (no false
  Pass on unloadable code). Tests: `go_backend_test` (`*BrokenModule`),
  `import_graph_load_test` (`TestImportGraph_DiscoveryErrorFails`).
- **Forbidden edge enforced per provider** — go (real `go list`), python (real
  AST walker), script (JSON contract). Tests: integration
  (`TestImportGraph_GoProviderRealModule`, `*PythonProviderRealWalker`),
  acceptance (`TestAccG2_GoEnforced`, `TestAccG2_PythonEnforced`,
  `TestAccG2_ScriptEnforced`).
- **A directory named like a manifest** (e.g. a `go.mod/` dir) is not treated as
  a manifest. Test: `manifests_test` (`TestHasFile_DirIsNotAManifest`).

## Residual Risks

- **checkEdges ignores edges into unknown packages**: an import is only
  evaluated when its target is itself a reported package. Providers must emit
  every intra-project package (the go/python backends do; a custom script must
  too) — documented in the JSON contract.
- **Monorepo / nested manifests**: detection resolves the nearest single
  manifest (top-level for a project); per-subtree provider selection is
  deferred to the `non-go-source-import-graphs` backlog item.
- **Node tool output variance**: depcruise/madge versions differ in path
  shapes; the `TestAccG2_NodeEnforced` smoke test is skip-guarded and lenient.
  Pure parsers are pinned to committed fixtures for deterministic coverage.
- **Remaining built-in languages** (Java/Rust/Ruby/PHP/C#/Elixir/Kotlin) are
  served by the custom-script provider until promoted from the backlog.
