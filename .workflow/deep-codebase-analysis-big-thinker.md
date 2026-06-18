### Big-Thinker Report: deep-codebase-analysis
**Date:** 2026-06-17

#### Problem
Centinela is greenfield-first: `centinela start` interviews the user and generates
a roadmap from a blank slate. A team with a mature codebase has the opposite
situation — architecture, stack, locales, and dependency structure already exist
*in the code*, but Centinela has no way to read them. Without a deterministic
scan, a brownfield adopter must hand-author `PROJECT.md`, guess an archetype, and
manually describe their layout — exactly the error-prone, drifts-from-reality
work the framework exists to eliminate. `centinela analyze` is the mechanical,
read-only, no-LLM scan that captures "what this repo actually is" as a
machine-readable `Inventory`. It is the **root of Phase 9**: three downstream
features (`archetype-inference-project-synthesis`, `spec-reconstruction`,
`adoption-baseline`) all `dependsOn` it, and it is the lowest-risk slice because
it is purely deterministic and never mutates source.

#### Scope
- **In:** a new `internal/analyze/` domain package (read-only directory walker
  with a skip set; extension→language counting + `primaryLanguage`; manifest
  detection for `go.mod`, `package.json`+scripts, `Gemfile`, `Cargo.toml`,
  `pyproject.toml`/`requirements.txt`, `Makefile` with build/test/framework
  extraction; i18n locale detection; depth-bounded package/dir layout; dependency
  graph — **Go via real `go list -json` package edges**, other ecosystems via
  declared manifest deps); a `schemaVersion`-tagged `Inventory` written
  deterministically (sorted, byte-stable) to the well-known path
  `.workflow/analysis.json`; a concise human summary to stdout; a thin
  `cmd/centinela/analyze.go`; a new `internal/golist/` **leaf** extracted from the
  existing import_graph loader so the Go graph is reused, not forked; the
  import-graph matrix mapping for both new packages (+ scaffold-mirror check).
- **Out:** LLM inference of archetype/specs/baseline (the explicit job of the
  downstream Phase 9 features — deliberate exclusion, not a new defer);
  source-level import graphs for non-Go languages (NEW → deferred);
  broad framework fingerprinting beyond manifest scripts (NEW → deferred);
  incremental/cached re-analysis (NEW → deferred); metrics enrichment — LOC,
  complexity, churn, coverage inference (NEW → deferred).

#### Dependencies & Assumptions
- **Reuses** the proven `go list -json` loader in
  `internal/gates/import_graph_load.go` (`loadPackages`, `loadModulePath`,
  `runGo`). The plan **extracts** it to a new shared leaf `internal/golist/` so
  both `gates` and `analyze` import the leaf — this avoids an `analyze → gates`
  edge entirely (no cycle, no aggregator promotion needed).
- `internal/analyze/**` is **not currently mapped** in the `import_graph` layers;
  the plan maps `internal/golist/**` → leaf (`allow = []`) and
  `internal/analyze/**` → domain (`allow = ["leaf"]`) in `centinela.toml`, with a
  scaffold-assets mirror check.
- No new `config` struct is strictly required for v1 (the output path is a
  package constant `analyze.DefaultOutPath`, with an optional `--out` flag).
- Assumes `analyze` is **diagnostic, not a gate**: sub-detector failures degrade
  to best-effort/empty with a recorded reason and exit 0; only an un-writable
  output path is a hard error.
- **Downstream consumers** bind to the `.workflow/analysis.json` `Inventory`
  contract; `schemaVersion` is the change-detection guard.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Schema instability — 3 features bind to the `Inventory` JSON; a later rename breaks them | High | Medium | Freeze the field set in the plan; `schemaVersion: 1`; keep v1 minimal; document it as a stable interface |
| Import-graph regression — `internal/analyze`/`internal/golist` unmapped or wrong edge | Medium | Medium | Map both layers explicitly (leaf + domain) in `centinela.toml`; extract `golist` so no `analyze → gates` edge exists; verify G2 output green; mirror into scaffold assets |
| Refactoring import_graph_load to delegate to `golist` regresses the import_graph gate | Medium | Low | Preserve behavior; keep the gate's existing tests green; move (not rewrite) the decode/loaders |
| Detection breadth scope creep ("detect every framework/ecosystem") | Medium | High | Ship a small data-driven table; defer broader fingerprinting + non-Go source graphs to the roadmap |
| `go list` cost/failure on a foreign or uncompilable repo | Low | Medium | Best-effort: record an empty Go graph with a `Note`; never abort the scan; exit 0 |
| Non-deterministic output (map iteration) causing noisy git diffs | Low | Low | Sort every list; no map ranged in output; `MarshalIndent` + trailing newline; byte-stable re-run test (AC-3) |

#### Rollout
- **Slice 1 (smallest correct):** `Inventory` schema + deterministic `Save`,
  extract `internal/golist` and refactor import_graph to delegate (gate tests
  stay green), the directory walker (skip set), language counting +
  `primaryLanguage`, and the `centinela analyze` command writing
  `.workflow/analysis.json` + stdout summary. This alone satisfies AC-1/2/3/4/5/6/8
  for Go repos and unblocks downstream consumers with a frozen contract.
- **Slice 2:** manifest detection (build/test/framework extraction) + i18n
  locales + bounded layout + non-Go declared-deps graph fallback, all data-driven
  tables (AC-7, polyglot/edge coverage).
- **Slice 3:** acceptance tests wired into `validate.commands`, gatekeeper +
  G2 verification (zero new failing edges), scaffold-assets mirror, docs +
  `Inventory` contract documentation.

#### Deferred Findings
Recorded to the Backlog phase via `centinela roadmap defer … --source
deep-codebase-analysis/big-thinker`:
- `non-go-source-import-graphs` — parse source-level import graphs for non-Go
  languages; v1 records declared manifest deps only.
- `brownfield-framework-fingerprinting` — detect frameworks via directory +
  dependency heuristics beyond manifest scripts.
- `incremental-codebase-analysis` — incremental/cached re-analysis of only
  changed directories.
- `codebase-metrics-enrichment` — LOC / complexity / churn / coverage inference
  on top of the inventory.

(The LLM-inference "Out" items are *not* deferred: they are the deliberate,
already-known scope of the downstream Phase 9 features, not new discoveries.)

#### Handoff
- Next role: feature-specialist
- Outstanding questions:
  1. Confirm the well-known output path `.workflow/analysis.json` (vs a
     top-level or `docs/`-adjacent path) and whether it is git-committed.
  2. Confirm the `internal/golist` extraction (option a) over re-invoking
     `go list` inside `analyze` (option b) — affects layering and the
     import_graph refactor surface.
  3. Confirm the v1 manifest set and the minimal build/test/framework signals
     to extract per ecosystem.
