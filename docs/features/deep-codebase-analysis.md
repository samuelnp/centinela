# Feature Brief — deep-codebase-analysis

> Phase 9: Brownfield Onboarding. `centinela analyze` scans an existing repo and
> produces a machine-readable **Inventory** — language(s), framework, build/test
> setup, i18n locales, module/package layout, and the import/dependency graph.
> It is the **foundation** every other brownfield feature reads from:
> `archetype-inference-project-synthesis`, `spec-reconstruction`, and
> `adoption-baseline` all `dependsOn` it. On a mature repo the truth already
> lives in the code and must be **reverse-engineered, not interviewed** — and
> this feature is the deterministic, no-LLM scan that captures that truth.

## Problem

Centinela is greenfield-first: `centinela start` interviews the user and
generates a roadmap from a blank slate. A team with a mature codebase has the
opposite situation — the architecture, stack, locales, and dependency structure
*already exist in the code*, but Centinela has no way to read them. Today a
brownfield adopter would have to hand-author `PROJECT.md`, guess an archetype,
and manually describe their layout, which is exactly the error-prone, drifts-
from-reality work the framework exists to eliminate. Every downstream brownfield
feature (infer the archetype, reconstruct specs, baseline existing violations)
needs a single, trustworthy, machine-readable description of "what this repo
actually is" before it can run. **Why now:** Phase 9 is blocked at the root —
nothing in brownfield onboarding can proceed until this inventory exists, and it
is the lowest-risk slice because it is purely mechanical (no LLM, read-only).

## How it works (mechanism)

1. **Scan** — `centinela analyze` walks the repo from the project root
   (respecting `.gitignore` / skipping `vendor/`, `node_modules/`, `.git/`,
   `.workflow/`) and collects deterministic facts:
   - **Languages** — count source files by extension, map extensions to language
     names, and derive a **primary language** (highest source-file count).
   - **Manifests / build & test setup** — detect known manifest files
     (`go.mod`, `package.json` + its `scripts`, `Gemfile`, `Cargo.toml`,
     `pyproject.toml`/`requirements.txt`, `Makefile`) and extract the
     framework / build / test signals each exposes (e.g. `package.json` scripts
     `build`/`test`, `go.mod` module path).
   - **i18n locales** — detect locale files/dirs (`locales/`, `i18n/`, `*.po`,
     `config/locales/*.yml`, etc.) and list the locale codes found.
   - **Module/package layout** — a bounded directory map (top-level + one or two
     levels) of the source tree.
   - **Import/dependency graph** — for **Go**, reuse the existing
     `import_graph` approach (`go list -json ./...` → package edges). For other
     languages v1 records the **declared dependency manifest** (names from
     `package.json`/`Cargo.toml`/etc.) and explicitly **defers** a parsed
     source-level import graph.
2. **Emit** — write the `Inventory` as JSON to a **stable well-known path**
   (`.workflow/analysis.json`) so downstream brownfield features consume a fixed
   contract, and print a **concise human summary** to stdout (primary language,
   detected framework/build/test, locale count, package count, graph edge count).
3. **Re-runnable & deterministic** — running again on an unchanged repo yields a
   byte-identical JSON (sorted keys/lists) so it diffs cleanly and is safe to
   commit.

## Key decisions to resolve in the plan

- **Inventory schema (the central contract).** The `Inventory` struct is the
  stable interface three downstream features bind to. The plan must fix field
  names, nesting, and a top-level `schemaVersion`, and decide which fields are
  required vs best-effort/optional, because changing it later breaks consumers.
- **Detection model = data-driven tables, not bespoke parsers.** Extension→language
  and manifest→framework/build/test detection should be small lookup tables so
  adding a language/framework is data, not code (keeps files ≤100 lines and the
  logic testable).
- **Import-graph reuse vs duplication.** The Go graph should reuse the proven
  `go list -json` loading already in `internal/gates/import_graph_load.go`. The
  plan must decide whether to **extract** that loader to a shared seam or
  **re-invoke** `go list` from `internal/analyze` (respecting layering — `analyze`
  must not create a forbidden edge to `gates`).
- **Output path & overwrite policy.** `.workflow/analysis.json` is the proposed
  well-known path. Decide overwrite (replace) vs `--out` override, and whether
  the file is git-committed (recommended: yes, like the roadmap artifacts).
- **Layer placement.** New `internal/analyze/` domain/aggregator package + thin
  `cmd/centinela/analyze.go`. The plan must confirm `internal/analyze/**` is
  mapped in the `import_graph` layers (it is **not** today) and pick the layer
  (leaf-ish domain importing only `config` + stdlib + `go list` subprocess, OR
  aggregator if it reuses `internal/gates`).

## Acceptance Criteria

1. `centinela analyze` run in a repo writes `.workflow/analysis.json` containing:
   detected languages with per-language file counts, a `primaryLanguage`,
   detected manifests with their build/test signals, i18n locales, a bounded
   package/directory layout, and a dependency graph (Go: package edges;
   other: declared manifest deps).
2. The command prints a concise human summary to stdout (primary language,
   framework/build/test, locale count, package count, graph edges).
3. The JSON is **deterministic**: a second run on an unchanged repo produces a
   byte-identical file (sorted keys and lists; no map iteration order).
4. Run on **this** Go repo, the inventory correctly reports Go as the primary
   language, the `go.mod` module path, the `Makefile`/`go test` build-test
   signal, and a non-empty Go package import graph.
5. The scan **skips** `vendor/`, `node_modules/`, `.git/`, `.workflow/`, and
   gitignored paths so counts reflect real source, not dependencies.
6. The scan is **read-only** — it never mutates source files (only writes the
   single output JSON).
7. A repo with **no recognized manifest** still produces a valid inventory
   (languages + layout populated; manifests empty; graph empty/best-effort) and
   exits 0 — analysis never hard-fails on an unfamiliar repo.
8. `Inventory` carries a `schemaVersion` so downstream consumers can detect
   format changes.
9. All new source files ≤100 lines; no cross-layer import violations.

## Edge Cases

- **Empty repo / only docs** → valid inventory: languages empty or doc-only,
  `primaryLanguage` empty, manifests empty, exit 0.
- **Polyglot repo** (Go + JS + Ruby) → all languages counted; `primaryLanguage`
  is the highest count; ties broken deterministically (alphabetical).
- **`go list` fails** (uncompilable Go, no toolchain) → Go graph is recorded as
  best-effort empty with a noted reason; the rest of the inventory still emits;
  exit 0 (analyze is diagnostic, not a gate).
- **`package.json` present but malformed JSON** → manifest recorded as detected-
  but-unparsable; scan continues.
- **Multiple manifests of different ecosystems** → all detected; `primaryLanguage`
  decides the headline, but every manifest is listed.
- **No i18n at all** → locales empty list (not an error); summary shows 0.
- **Huge repo** → directory map is depth-bounded so output stays small/bounded.
- **Re-run after the output already exists** → overwritten deterministically;
  no spurious diff if nothing changed.
- **Symlinks / unreadable files** → skipped without aborting the scan.

## Data Model

New committed artifact `.workflow/analysis.json` — a `schemaVersion`-tagged
`Inventory`:

```go
type Inventory struct {
    SchemaVersion   int               // format version, currently 1
    PrimaryLanguage string            // highest source-file count, "" if none
    Languages       []LanguageStat    // {Name, FileCount}, sorted desc then name
    Manifests       []Manifest        // {Kind, Path, Build, Test, Framework, Deps}
    Locales         []string          // detected locale codes, sorted
    Packages        []string          // bounded module/package layout, sorted
    Graph           DependencyGraph   // {Module, Edges []Edge{From,To}} (Go) / declared deps
}
```

Detection tables live as data (extension→language, manifest filename→kind +
extractor) so adding ecosystems is a table edit. No new `config` struct is
strictly required for v1 (the well-known output path can be a constant), but the
plan may add an optional `--out` flag.

## Integration Points

- **Reuse `go list -json`**: the loader pattern in
  `internal/gates/import_graph_load.go` (`loadPackages`, `loadModulePath`,
  `runGo`) is the proven Go-graph mechanism — extract/share or re-invoke per the
  plan's layering call.
- **Logic in `internal/analyze/`**: scan, detect, build Inventory, marshal —
  kept out of `cmd/` (G7) and a leaf/aggregator in the import-graph matrix.
- **Command**: new `cmd/centinela/analyze.go` — thin wiring (flags → call →
  render → write file).
- **Render**: reuse `internal/ui` summary styling for the stdout summary
  (presentation only; `analyze` returns the typed `Inventory`).
- **Downstream consumers**: `archetype-inference-project-synthesis`,
  `spec-reconstruction`, `adoption-baseline` read `.workflow/analysis.json`.

## Risks

- **Schema instability** — three features bind to the `Inventory` JSON; a later
  field rename breaks them. Mitigated by `schemaVersion` + freezing the contract
  in the plan and keeping v1 minimal.
- **Layering / import-graph regression** — `internal/analyze/**` is not yet in
  the `import_graph` layers, so it would surface as an *unmapped* (warn) package
  and any reuse of `internal/gates` could create a forbidden edge. The plan must
  map it explicitly (leaf domain or aggregator) and mirror the toml change into
  `internal/scaffold/assets` if the matrix is scaffolded.
- **Detection breadth scope creep** — "detect every framework" is unbounded. v1
  ships a small, well-tested table and **defers** broader ecosystem coverage and
  non-Go source-level import graphs to the roadmap.
- **`go list` cost/failure on foreign repos** — must be best-effort and never
  abort the whole scan (analyze is diagnostic, not a gate).
