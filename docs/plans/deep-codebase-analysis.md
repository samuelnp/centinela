# Implementation Plan — deep-codebase-analysis

> Feature brief: `docs/features/deep-codebase-analysis.md`.
> Spec: `specs/deep-codebase-analysis.feature`.

Phase 9 (Brownfield Onboarding) is blocked at the root until Centinela can read
an existing repo. This feature adds a **mechanical, deterministic, no-LLM** scan:
a new `centinela analyze` command that walks the repo, detects languages /
manifests / locales / layout / dependency graph, assembles a typed `Inventory`,
writes it to the well-known path `.workflow/analysis.json`, and prints a concise
human summary. All logic lives in a new `internal/analyze/` package; `cmd/` only
wires. The Go import graph **reuses the proven `go list -json` loader** behind a
shared seam so we do not duplicate or fork that mechanism. v1 is the **smallest
correct slice**: capture the truth, freeze a `schemaVersion`-tagged contract,
defer everything non-essential to the roadmap.

## Decisions (DECIDED)

1. **Output contract = `Inventory` JSON at `.workflow/analysis.json`,
   `schemaVersion: 1`.** This is the stable interface
   `archetype-inference-project-synthesis`, `spec-reconstruction`, and
   `adoption-baseline` bind to. The plan freezes the field set (Data Model
   below); changes after v1 bump `schemaVersion`. A `--out <path>` flag overrides
   the default for callers that want it elsewhere; default is committed (like the
   roadmap artifacts).

2. **Detection is data-driven, not bespoke parsers.** Two lookup tables —
   `extension → language` and `manifest filename → {Kind, extractor}` — keep the
   logic small, testable, and ≤100 lines per file. Adding a language/framework is
   a table edit, not new control flow.

3. **Analyze is diagnostic, never a gate — it does not hard-fail.** Any
   sub-detector that errors (`go list` on uncompilable code, malformed
   `package.json`, unreadable file) degrades to a best-effort/empty result with a
   recorded reason; the overall command still emits a valid inventory and exits 0
   (AC-3/7, edge cases). Only an un-writable output path is a real error.

4. **Read-only scan with a fixed skip set.** The walker skips `vendor/`,
   `node_modules/`, `.git/`, `.workflow/`, `dist/`, `build/` and gitignored paths
   so counts reflect source, not dependencies (AC-5). It never mutates source;
   the only write is the single output JSON (AC-6).

5. **Determinism everywhere.** Every list is sorted (languages by count desc then
   name; locales, packages, graph edges lexicographically); no map is ranged in
   output order; `Save` uses `json.MarshalIndent(_, "", "  ")` + trailing `\n`.
   A second run on an unchanged repo is byte-identical (AC-3).

## Go import-graph reuse (CALL-OUT — avoid fork/duplication, avoid a cycle)

The Go package graph must reuse, not reimplement, the loader proven in
`internal/gates/import_graph_load.go` (`loadPackages()` → `[]goListPkg` via
`go list -json ./...`, `loadModulePath()` via `go list -m`, `runGo()`). Two
options; the plan **DECIDES option (a)**:

- **(a) Extract a tiny shared leaf `internal/golist/`** (`Packages() []Pkg`,
  `ModulePath() string`, `Pkg{ImportPath, Imports, TestImports, XTestImports}`).
  Both `internal/gates` (import_graph) and `internal/analyze` import this leaf.
  This removes the only reason `analyze` would need to import `gates`, so
  `analyze` stays a clean **leaf-ish domain** importing only `internal/config` (if
  needed) + `internal/golist` + stdlib — **no aggregator edge, no cycle.**
  `import_graph_load.go` is refactored to delegate to `internal/golist` (behavior
  preserved; its existing tests still pass).
- (b) Re-invoke `go list` directly from `analyze` (copy ~40 lines). Rejected:
  duplicates the streamed-JSON decode + stderr-surfacing logic that already has
  tests, and drifts over time.

**Layer placement (load-bearing):**
`internal/golist` is a new **leaf** (imports stdlib + `os/exec` only) → add
`internal/golist/**` to the leaf layer `paths` in `centinela.toml`
(`allow = []`). `internal/analyze` is a **domain** package importing only leaves
(`internal/golist`, optionally `internal/config`) → it fits the existing
**domain** layer (`allow = ["leaf"]`); add `internal/analyze/**` to the domain
layer `paths`. `cmd/centinela/analyze.go` (outer, allows domain+leaf+aggregator)
wires it. **No edge to `internal/gates`, no cycle.** If reusing `internal/ui` for
the summary, `analyze` returns the typed `Inventory` and `cmd/` calls
`ui.Render…`; `analyze` itself does **not** import `ui`.

**toml change:** add `internal/golist/**` to the leaf layer and
`internal/analyze/**` to the domain layer (two `paths` edits), with a comment.
**Mirror into `internal/scaffold/assets`** if the import-graph matrix is part of
the scaffolded `centinela.toml` template — verify by hand during code (the
scaffold parity test only covers 8 arch docs, so toml drift is not auto-caught).

## v1 scope

**In:**
- `internal/analyze/` package: directory walker (skip set, read-only),
  extension→language counting + `primaryLanguage`, manifest detection
  (`go.mod`, `package.json`+scripts, `Gemfile`, `Cargo.toml`,
  `pyproject.toml`/`requirements.txt`, `Makefile`) with build/test/framework
  extraction, i18n locale detection, bounded (depth-limited) package/dir layout,
  dependency graph (**Go: real `go list` package edges**; other ecosystems:
  declared manifest dependency names).
- `Inventory` schema with `schemaVersion`, deterministic sorted JSON `Save`/load.
- `internal/golist/` leaf extracted from the existing import_graph loader; the
  import_graph gate refactored to delegate (behavior + tests preserved).
- `centinela analyze` command: scan → write `.workflow/analysis.json` → print
  concise stdout summary; `--out` override; best-effort/exit-0 semantics.
- import-graph matrix mapping for the two new packages (+ scaffold mirror check).

**Out (deferred — see Deferred Findings):**
- **LLM inference of archetype / specs / baseline** — that is the explicit job of
  the downstream Phase 9 features (`archetype-inference-project-synthesis`,
  `spec-reconstruction`, `adoption-baseline`); this feature is the deterministic
  substrate only. (Deliberate exclusion — already-known, not deferred-new.)
- **Source-level import graphs for non-Go languages** (parsing JS/TS/Ruby/Rust
  imports) — v1 records declared manifest deps only. **NEW → defer.**
- **Broad framework fingerprinting** (detecting Rails/Next/Django/etc. by
  directory and dependency heuristics beyond manifest scripts). **NEW → defer.**
- **Incremental / cached re-analysis** (only re-scan changed dirs). **NEW → defer.**
- **Metrics enrichment** (LOC, complexity, churn, test-coverage inference). **NEW → defer.**

## Step 2 — code

New / edited source files (each ≤100 lines):

| File | Change | Budget |
|------|--------|--------|
| `internal/analyze/inventory.go` | NEW. `Inventory`, `LanguageStat`, `Manifest`, `DependencyGraph`, `Edge` structs + `schemaVersion` const; `Save(path, Inventory)` (deterministic sorted JSON + trailing `\n`) | ~90 |
| `internal/analyze/walk.go` | NEW. read-only directory walker with the skip set + gitignore awareness; returns file list / per-extension counts; bounded depth for layout | ~95 |
| `internal/analyze/languages.go` | NEW. `extensionLanguage` table + `detectLanguages(counts) ([]LanguageStat, primary string)` (sorted desc, alpha tiebreak) | ~80 |
| `internal/analyze/manifests.go` | NEW. `manifestTable` (filename→Kind+extractor) + `detectManifests(root) []Manifest`; per-ecosystem extractors (go.mod module, package.json scripts/deps, Cargo/Gemfile/py) | ~95 |
| `internal/analyze/locales.go` | NEW. `detectLocales(root) []string` over known locale dirs/patterns | ~55 |
| `internal/analyze/graph.go` | NEW. `buildGraph(root) DependencyGraph` — Go via `golist.Packages()` (module-internal edges only), best-effort empty on error; non-Go falls back to declared manifest deps | ~90 |
| `internal/analyze/analyze.go` | NEW. `Analyze(root) (Inventory, error)` — orchestrate walk → detect → assemble; best-effort sub-detectors (Decision #3) | ~80 |
| `internal/golist/golist.go` | NEW (extracted). `Pkg` struct; `Packages() ([]Pkg, error)`; `ModulePath() (string, error)`; `runGo` | ~90 |
| `internal/gates/import_graph_load.go` | EDIT. delegate to `internal/golist` (decode/loaders moved out); keep gate behavior + its tests green | net −X |
| `cmd/centinela/analyze.go` | NEW. `analyzeCmd`: `--out` flag → `analyze.Analyze(root)` → `analyze.Save` → render summary; thin wiring only (G7) | ~85 |
| `cmd/centinela/root.go` | EDIT. register `analyzeCmd` | +1 line |
| `internal/ui/render_analyze.go` *(optional)* | NEW if summary formatting exceeds the cmd budget: `RenderInventorySummary(Inventory) string` | ~55 |
| `centinela.toml` | EDIT. add `internal/golist/**` to leaf layer, `internal/analyze/**` to domain layer (with comment) | +2 paths |
| `internal/scaffold/assets/centinela.toml` *(if matrix scaffolded)* | EDIT. mirror the two layer additions | match |

**Well-known path constant** (`analyze.DefaultOutPath = ".workflow/analysis.json"`)
lives in `internal/analyze`, not `cmd/`, so downstream features import the same
constant rather than re-stringing the path.

### Inventory schema (`.workflow/analysis.json`)

```go
const SchemaVersion = 1

type Inventory struct {
    SchemaVersion   int             `json:"schemaVersion"`
    PrimaryLanguage string          `json:"primaryLanguage"`
    Languages       []LanguageStat  `json:"languages"`   // sorted: count desc, name asc
    Manifests       []Manifest      `json:"manifests"`   // sorted by Path
    Locales         []string        `json:"locales"`     // sorted
    Packages        []string        `json:"packages"`    // sorted, depth-bounded
    Graph           DependencyGraph `json:"graph"`
}

type LanguageStat struct {
    Name      string `json:"name"`
    FileCount int    `json:"fileCount"`
}

type Manifest struct {
    Kind      string   `json:"kind"`      // go-mod | npm | cargo | gem | python | make
    Path      string   `json:"path"`
    Framework string   `json:"framework,omitempty"`
    Build     string   `json:"build,omitempty"`
    Test      string   `json:"test,omitempty"`
    Deps      []string `json:"deps,omitempty"` // declared dep names, sorted
}

type DependencyGraph struct {
    Kind   string `json:"kind"`   // "go-packages" | "declared-deps" | "none"
    Module string `json:"module,omitempty"`
    Edges  []Edge `json:"edges"`  // sorted by From,To
    Note   string `json:"note,omitempty"` // e.g. "go list failed: …" (best-effort)
}

type Edge struct {
    From string `json:"from"`
    To   string `json:"to"`
}
```

- **Determinism (AC-3):** all slices sorted, no maps in serialized form,
  `MarshalIndent` + trailing newline ⇒ byte-stable re-runs, clean git diffs.
- **Stability:** `SchemaVersion` is the downstream-contract guard (AC-8).

## Step 3 — tests

Colocated per-package `_test.go` (95% per-package coverage gate is NOT moved by
`tests/` tier files — add coverage next to the code). Each ≤100 lines (G1 applies
to `_test.go` too):

- `internal/analyze/languages_test.go` — extension→language mapping; sorting
  (count desc, alpha tiebreak); empty input ⇒ empty + `primary == ""` (edge).
- `internal/analyze/manifests_test.go` — each manifest kind detected from a
  `t.TempDir()` fixture; `package.json` scripts→build/test extracted; malformed
  `package.json` ⇒ detected-but-unparsable, scan continues (edge).
- `internal/analyze/locales_test.go` — locale dirs/patterns detected; no-i18n ⇒
  empty list (edge).
- `internal/analyze/walk_test.go` — skip set excludes `vendor/`/`node_modules/`/
  `.git/`/`.workflow/`; depth bound caps layout; unreadable/symlink skipped (edge).
- `internal/analyze/graph_test.go` — Go graph built from a fixture module; `go
  list` failure ⇒ best-effort empty with `Note` set, no panic (edge); non-Go ⇒
  `declared-deps`.
- `internal/analyze/inventory_test.go` — `Save` is sorted + byte-stable across two
  saves of the same `Inventory` (AC-3); `schemaVersion` present (AC-8).
- `internal/analyze/analyze_test.go` — `Analyze` over a fixture polyglot repo
  assembles all sections; a failing sub-detector still yields a valid inventory,
  no error, (AC-7); empty repo ⇒ valid empty inventory (edge).
- `internal/golist/golist_test.go` — `Packages`/`ModulePath` decode the streamed
  `go list` JSON; `go list` error is surfaced (mirrors the moved coverage).
- `internal/gates/import_graph_load_test.go` — kept green after delegating to
  `internal/golist` (no behavior change).

**Integration:** `tests/integration/analyze_test.go` — in a `t.TempDir()` mini Go
module with a `package.json` and a `locales/` dir: run `Analyze` (or the built
binary), assert `.workflow/analysis.json` is written with Go primary, the module
path, npm manifest scripts, locales, a non-empty Go edge list; re-run and assert
**byte-identical** output (AC-3/4); confirm no source file was mutated (AC-6).

**Acceptance:** `tests/acceptance/analyze_*` (executable, one per Gherkin
scenario) — run the real `centinela analyze` binary against a fixture repo and
assert exit 0 + the summary lines + the presence/shape of `.workflow/analysis.json`
for: scan-writes-inventory, deterministic-rerun, no-manifest-still-valid,
skips-vendor-deps, read-only. Register the acceptance runner in
`validate.commands` in `centinela.toml`.

`.workflow/deep-codebase-analysis-edge-cases.md` — map every brief edge case
(empty repo, polyglot, `go list` fails, malformed package.json, multi-ecosystem,
no-i18n, huge-repo depth bound, re-run overwrite, symlink/unreadable) to the test
covering it.

Note: `go test ./...` runtime — keep `[verify] verify_timeout` margin in mind;
the new packages are small and fast.

## Step 4 — validate

Gatekeeper report `.workflow/deep-codebase-analysis-gatekeeper.md`; `centinela
validate` green (lint + types + full suite). **Confirm the G2 import-graph gate
output:** the new `internal/golist/**` (leaf) and `internal/analyze/**` (domain)
mappings produce **zero new failing edges** (`analyze → golist` is leaf-allowed;
`gates → golist` is leaf-allowed; `analyze` does NOT import `gates`/`ui`/`cmd`).
Confirm every touched source file ≤100 lines (including `_test.go`). **Dogfood
`centinela analyze` from a `/tmp` binary built from `./cmd/centinela`** on this
repo before relying on the installed binary, and eyeball the produced
`.workflow/analysis.json` against AC-4 (Go primary, module path, edges).
Production-readiness subagent if the gate is enabled.

## Step 5 — docs

Documentation-specialist `.md` + `.json`; regenerate `docs/project-docs/index.html`;
changelog artifact `.workflow/deep-codebase-analysis-changelog.md` (create early
via `evidence artifact new` so completion doesn't fail). Document: the `centinela
analyze` command + `--out` flag + exit-0/best-effort semantics; the **`Inventory`
schema** at `.workflow/analysis.json` with `schemaVersion` (this is the
downstream contract — document it as a stable interface); the detection tables
(supported languages/manifests/locales) and how to extend them; the Go import-
graph reuse via `internal/golist`. Add a PROJECT.md G2 note that
`internal/golist` is a leaf and `internal/analyze` is a domain package importing
only leaves, and mirror the `centinela.toml` import-graph change into
`internal/scaffold/assets` if the matrix is scaffolded.
