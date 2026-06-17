# Edge Cases: g2-import-graph-gate

## Covered

1. **Unmapped package (matches no layer glob).** Not a violation. Both edge
   endpoints must be mapped for an edge to be evaluated; an edge into or out of
   an unmapped package is ignored. The unmapped package surfaces as a
   non-failing **Warn** listing its import path.
   Covered: `TestCheckEdges_UnmappedAndIgnoredEdges`,
   `TestCheckImportGraph_WarnOnUnmapped`.

2. **Malformed config vs a real violation.** A config error (empty layer name,
   layer with no paths, allow-list referencing an unknown layer, or a blank
   module path) Fails with a message prefixed `import_graph config:` and MUST
   NOT contain the violation arrow `->`, keeping misconfig distinguishable from
   a genuine forbidden edge.
   Covered: `TestBuildMatrix_Errors`, `TestCheckImportGraph_ConfigErrorFails`,
   acceptance `TestAccept_ImportGraph_MalformedConfigDistinctFromViolation`.

3. **Uncompilable / unloadable module -> load error (never a false Pass).**
   When `go list -json ./...` (or `go list -m` for module discovery) exits
   non-zero — e.g. a malformed `go.mod` — the gate Fails with the load
   diagnostic folded in. NOTE: a plain syntax error in a `.go` file does NOT
   make `go list` exit non-zero (it reports a per-package Error field), so the
   load-error path is driven by a broken `go.mod`.
   Covered: `TestRunImportGraph_LoadErrorFails`,
   `TestResolveModule_DiscoveryErrorFails`.

4. **_test.go files cannot smuggle a forbidden cross-layer import.** TestImports
   and XTestImports (external `_test` packages) fold into the
   package-under-test's import set, so a test file inherits the production
   package's layer and a test-only forbidden import is still caught.
   Covered: `TestScopePackages_FoldsTestImportsDropsExternal`.

5. **Intra-layer and self imports always allowed.** Same-layer imports (incl.
   duplicate layer entries that union to one name) are never flagged;
   self-imports are dropped during scoping.
   Covered: `TestAllowed_SameLayerAndAllowList`,
   `TestCheckEdges_AllowedAndForbidden`,
   `TestBuildMatrix_ValidatesAndUnionsDuplicates`.

6. **Stdlib and third-party imports ignored.** Imports outside the module path
   prefix (`fmt`, `os`, `golang.org/x/tools/...`) are stripped before checking.
   Module-prefix matching is segment-boundary safe, so a third-party path that
   shares the module string as a substring (e.g. `module + "x/other"`) is not
   misclassified as in-module.
   Covered: `TestStripModulePrefix`,
   `TestScopePackages_FoldsTestImportsDropsExternal`.

7. **Empty matrix (block present, zero layers) -> Warn, not silent Pass.** An
   enabled block with no layers Warns ("layer matrix is empty") so an
   accidentally-empty config does not quietly disable enforcement. A disabled or
   absent block omits the gate entirely.
   Covered: `TestCheckImportGraph_EmptyMatrixWarns`,
   `TestImportGraph_EmptyMatrixWarns`, `TestImportGraph_DisabledOmitted`.

8. **Diff filter ignored — whole-module load.** A graph edge can be broken/fixed
   by a file outside the diff set, so a diff-scoped load would yield false
   passes. The gate deliberately ignores the `*gitdiff.Set` filter and always
   loads the whole module; a violation outside the diff is still reported.
   Covered by design (signature ignores the filter; documented in source) and
   exercised whole-module via the integration/acceptance fixtures.

## Residual Risks

- **Two exec-error branches remain uncovered (~unreachable).** The JSON-decode
  error branch in `loadPackages` and the `errEmptyModule` return in
  `resolveModule` require `go list` to emit malformed JSON / an empty module
  path *without* a non-zero exit, which the Go toolchain does not do in
  practice. Total coverage stays >=95% with these unhit. Mitigation: any real
  `go list` failure exits non-zero and is already surfaced as Fail, so these
  branches cannot cause a false Pass.
- **Acceptance drives the gate via `gates.RunWithFilter`, not a spawned
  `centinela validate` binary.** This keeps the test hermetic and binary-version
  independent; exit-code mapping is owned by the validate command and covered by
  the validate step.
