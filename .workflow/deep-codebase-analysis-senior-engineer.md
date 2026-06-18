### Senior-Engineer Report: deep-codebase-analysis
**Date:** 2026-06-18

Implements `centinela analyze`: a deterministic, read-only, no-LLM repo scan that
walks the project root, detects languages / manifests / locales / package layout
/ dependency graph, assembles a `schemaVersion`-tagged `Inventory`, writes it to
`.workflow/analysis.json` (overridable via `--out`), and prints a concise
summary. The Go package graph reuses the proven `go list -json` loader via a new
shared `internal/golist` leaf (Decision option (a) in the plan) so the loader is
not forked; the import_graph gate now delegates to that same leaf with behavior
and its tests preserved.

## Files Touched
| Path | Reason |
|------|--------|
| `internal/golist/golist.go` (85) | NEW leaf. Extracted `go list -json` loader: `Pkg`, `Packages()`, `ModulePath()`, internal `runGo`/`firstStderrLine`. Imports stdlib + os/exec only. |
| `internal/gates/import_graph_load.go` (22) | EDIT. Refactored to delegate to `internal/golist` (`goListPkg = golist.Pkg` alias; `loadPackages`/`loadModulePath` thin wrappers). Gate behavior + tests unchanged. |
| `internal/analyze/inventory.go` (80) | NEW. `Inventory`/`LanguageStat`/`Manifest`/`DependencyGraph`/`Edge` structs; `SchemaVersion=1`; `DefaultOutPath`; deterministic `Save` (MarshalIndent + trailing `\n`, marshals in-memory first so a write failure leaves no partial file; creates the parent dir). |
| `internal/analyze/languages.go` (57) | NEW. `extensionLanguage` table + `detectLanguages` (count desc, name asc tiebreak, primary). |
| `internal/analyze/walk.go` (72) | NEW. Read-only walker: skip set (vendor/node_modules/.git/.workflow/dist/build), gitignore-aware, symlink-skipping, depth-bounded layout, per-extension counts; unreadable root is the sole hard error. |
| `internal/analyze/gitignore.go` (50) | NEW. Minimal dependency-free root `.gitignore` matcher (path/name/dir-prefix). |
| `internal/analyze/manifests.go` (52) | NEW. `manifestTable` (filename→kind+extractor) + `detectManifests` (top-level, best-effort, sorted by path). |
| `internal/analyze/extract_npm.go` (72) | NEW. `package.json` extractor: build/test scripts, sorted dep names, framework fingerprint; invalid JSON ⇒ detected-but-unparsable. |
| `internal/analyze/extract_misc.go` (68) | NEW. go.mod / Makefile / Cargo / Gemfile / pyproject / requirements extractors (best-effort line/TOML scans). |
| `internal/analyze/extract_helpers.go` (62) | NEW. `tomlSection`, `firstQuoted`, `asSet`, `splitReqName` helpers. |
| `internal/analyze/locales.go` (65) | NEW. `detectLocales` over known i18n dirs (root + one level), locale-code regex, sorted unique. |
| `internal/analyze/graph.go` (74) | NEW. `buildGraph`: Go via `golist` module-internal edges (best-effort empty + `Note` on toolchain error); non-Go ⇒ declared-deps; else `none`. |
| `internal/analyze/graph_helpers.go` (42) | NEW. `hasManifest`, `declaredEdges`, `sortedEdges` (non-nil so JSON emits `[]`). |
| `internal/analyze/analyze.go` (25) | NEW. `Analyze(root)` orchestrator: walk → detect → assemble; sub-detectors degrade best-effort; only unreadable root errors. |
| `cmd/centinela/analyze.go` (47) | NEW outer/thin. `analyzeCmd` + `--out` flag; calls `Analyze` → `Save` → `ui.RenderInventorySummary`; hard errors ⇒ non-zero exit + clear stderr, no partial artifact. Self-registers via `init()`. |
| `internal/ui/render_analyze.go` (54) | NEW. `RenderInventorySummary(Inventory) string` — keeps cmd thin; analyze never imports ui. |
| `centinela.toml` (EDIT) | Added `internal/golist/**` to the leaf layer and `internal/analyze/**` to the domain layer, each with a comment. |

All 16 Go source files are ≤100 lines (largest: golist.go at 85).

## Architecture Compliance
- Boundary checks passed: the `import_graph` gate reports ⚠ (Warn, "packages
  match no configured layer" — the pre-existing non-failing kind for unmapped
  domains), **not** ✗ Fail. Zero forbidden edges introduced. `internal/golist`
  (leaf) imports only stdlib+os/exec; `internal/analyze` (domain) imports only
  `internal/golist` (leaf) + stdlib — no edge to gates/ui/cmd, no cycle.
  `gates → golist` and `analyze → golist` are both leaf-allowed.
- G1 file size: each touched file ≤100 lines (verified by the G1 gate: "All
  files under 100 lines").
- G7 outer-layer rule: `cmd/centinela/analyze.go` is pure wiring (parse flag →
  call domain → render → print); summary formatting lives in `internal/ui`; all
  detection logic lives in `internal/analyze`. Scaffold mirror checked:
  `internal/scaffold/assets/centinela.toml` does NOT carry the import_graph
  matrix, so no mirror edit was needed.

## Type-Safety Notes
- No `interface{}`/`any` anywhere. `package.json` is decoded into a concrete
  `packageJSON` struct, not a generic map, in the output path.
- `goListPkg = golist.Pkg` is a type alias so existing gate tests that construct
  `[]goListPkg` literals compile unchanged against the shared leaf type.
- `Save` marshals the full payload in memory before writing, so a write error
  yields no partial/corrupt JSON.

## Trade-Offs
- The manifest TOML/Gemfile/requirements extractors are deliberately minimal
  line/section scanners (no full TOML/Ruby parser) to stay within the ≤100-line
  budget and the "table edit, not new control flow" decision. They recover
  declared dep names best-effort; deeper parsing is out of v1 scope.
- The go-mod manifest records the declared module path in its `Build` field (the
  frozen `Manifest` struct has no dedicated module field; the graph also carries
  `Module`). This makes the summary's build line show the module path when only
  go.mod is present; acceptable given the frozen v1 contract.
- The gitignore matcher honors plain/dir patterns only (no negation/nested
  ignore files) — sufficient for the skip-intent (AC-5).

## Deferred Findings
- none (the big-thinker already deferred non-go-source-import-graphs,
  brownfield-framework-fingerprinting, incremental-codebase-analysis, and
  codebase-metrics-enrichment; no new TODOs introduced).

## Handoff
- Next role: qa-senior
- Outstanding TODOs: add colocated `_test.go` coverage for `internal/analyze`
  and `internal/golist` (per-package 95% gate; currently both show "[no test
  files]"), plus the integration + acceptance tiers per the plan's Step 3. The
  implementation is dogfood-verified: Go primary, module path recorded, 133 Go
  edges on this repo, npm/locale/framework detection on a polyglot fixture,
  byte-identical re-run, and exit-0 best-effort vs exit-1 hard-error semantics
  all confirmed. NOTE: the pre-existing `tests/integration`
  `TestReleaseWorkflowAndInstallerContainExpectedFlow` failure is unrelated to
  this feature (it also fails on `main`) — about the release-workflow YAML.
