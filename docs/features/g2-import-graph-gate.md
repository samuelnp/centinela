# Feature: g2-import-graph-gate

## Problem — what pain does this solve? Who is the user?

**User:** A developer (or AI agent) working inside a Centinela-governed Go
project, plus the maintainer who defined the architecture's layer rules.

**Pain:** The G2 layer-boundary rule lives only as prose in `PROJECT.md`
("`cmd/` may import `internal/*`; `internal/workflow` and `internal/gates`
may import `internal/config` only; …"). Nothing mechanically enforces it.
An agent can silently introduce a forbidden cross-layer import — e.g.
`internal/config` importing `internal/ui`, inverting the dependency stack —
and no gate catches it. The gatekeeper subagent reviews prose, not the
actual import graph, so violations ship. This turns a stated architectural
invariant into a "requested, not enforced" rule, exactly the failure mode
Centinela exists to eliminate.

## User Stories

- As a maintainer, I want the G2 layer matrix expressed as machine-readable
  config so a forbidden import fails `centinela validate` instead of relying
  on human review.
- As an agent, I want immediate, specific feedback ("`internal/config`
  imported `internal/ui` — config is a leaf layer") so I can fix the
  violation in the same loop rather than at PR review.
- As a maintainer, I want the gate to follow the same enable/skip/diff-aware
  conventions as the existing gates (file_size, i18n, build) so it composes
  cleanly with `centinela validate`.

## Acceptance Criteria — concrete, testable (→ Gherkin scenarios)

1. Given a `[gates.import_graph]` config defining layers (path globs) and
   each layer's allowed internal imports, when `centinela validate` runs,
   then the gate parses the Go import graph (`go/packages`) and reports a
   `Result` named `import_graph`.
2. Given a package imports another package its layer is **not** permitted to
   import, when the gate runs, then the gate returns `Fail` and lists each
   violating edge as `<importer-pkg> → <forbidden-pkg> (<importer-layer>
   may not import <imported-layer>)`.
3. Given all imports respect the matrix, when the gate runs, then it returns
   `Pass`.
4. Given the gate is disabled in config (or no `[gates.import_graph]` block),
   when validate runs, then the gate is omitted (no `Result`), exactly like
   the other optional gates.
5. Given a package path matches no configured layer, when the gate runs,
   then it is reported as `Warn` (unmapped package) rather than silently
   passing — the matrix must stay exhaustive.

## Edge Cases — invalid input, concurrency, empty state, limits

- **Standard-library / third-party imports**: ignored — only imports within
  the module's own path are evaluated against the matrix.
- **Test files** (`_test.go`): in-scope (a test importing across layers
  still couples them); same-package and `_test` external test packages map
  to the layer of the package under test.
- **Unmapped package**: `Warn`, surfaced explicitly (criterion 5).
- **Malformed config** (layer with no globs, glob matching nothing,
  unknown layer referenced in an allow-list): gate returns `Fail` with a
  config-error message, distinct from an import violation.
- **Self-imports / intra-layer imports**: always allowed (a layer may
  import within itself).
- **Empty matrix** (block present but no layers): treated as disabled with a
  `Warn` so it can't masquerade as a passing gate.
- **Build failures in `go/packages`** (uncompilable code): gate returns
  `Fail` with the load error, not a false `Pass`.

## Data Model

- `ImportGraphConfig` (in `internal/config`): `Enabled bool`, `Module
  string` (module path prefix, defaulted from `go.mod`), `Layers []Layer`.
- `Layer`: `Name string`, `Paths []string` (globs relative to module root),
  `Allow []string` (names of layers this layer may import).
- Derived at runtime: package→layer map; allow-set per layer.

## Integration Points

- `internal/gates`: new `checkImportGraph(cfg, filter)` wired into
  `RunWithFilter`, gated by `cfg.Gates.ImportGraph.Enabled`. Returns the
  shared `gates.Result`.
- `internal/config`: extend the TOML schema with `[gates.import_graph]`.
- `go/packages`: load the module's packages + import edges.
- `go.mod`: read the module path to scope "internal" imports.
- `centinela.toml`: dogfood by adding the block encoding this repo's own G2
  matrix from `PROJECT.md`.

## Risks — performance, security, unclear requirements

- **Performance**: `go/packages` load is slower than a file scan; mitigate
  by loading once (`NeedImports|NeedName`) and reusing. Diff-aware filtering
  is harder for a graph (a removed import elsewhere can fix/break an edge),
  so v1 runs a whole-module load and ignores the file filter for this gate.
- **Matrix drift**: the config matrix and the PROJECT.md prose can diverge.
  Acceptable for v1; a future drift-check can reconcile them.
- **Archetype generality**: matrix is generic (layer = globs + allow-list),
  but only the Go parser ships in v1. TS/Python parsers are explicitly out.

## Decomposition

Single, cohesive feature; no split required. Out of scope (future work):
TypeScript (ts-morph/madge) and Python (AST) parsers, pre-write hook
enforcement, and a PROJECT.md↔config drift checker (custom-gate-sdk).
