# Plan: g2-import-graph-gate

> Feature brief: [docs/features/g2-import-graph-gate.md](../features/g2-import-graph-gate.md)
> Spec: `specs/g2-import-graph-gate.feature`

## Goal

Turn the prose G2 layer-boundary rule in `PROJECT.md` into a mechanical gate
that parses the Go import graph and fails `centinela validate` on any import
that violates a configurable per-layer allow matrix.

## Scope (v1, per scoping decisions)

- **Language:** Go only (`go/packages`).
- **Matrix source:** structured config — `[gates.import_graph]` in
  `centinela.toml`.
- **Enforcement:** `centinela validate` gate only (no pre-write hook).

Out: TypeScript/Python parsers, pre-write enforcement, PROJECT.md↔config
drift checker.

## Design

### Config (`internal/config`)

```toml
[gates.import_graph]
enabled = true
module = "github.com/samuelnp/centinela"   # optional; default from go.mod

[[gates.import_graph.layers]]
name  = "config"
paths = ["internal/config/**"]
allow = []                                   # leaf layer

[[gates.import_graph.layers]]
name  = "domain"
paths = ["internal/workflow/**", "internal/gates/**"]
allow = ["config"]
# … remaining layers encode the PROJECT.md G2 matrix
```

New types: `ImportGraphConfig{Enabled, Module, Layers}`, `Layer{Name,
Paths, Allow}`. Add to the existing `Gates` struct.

### Gate (`internal/gates`)

- `import_graph.go` — `checkImportGraph(cfg, filter) Result`: orchestrates
  load → classify → check; returns a `gates.Result{Name:"import_graph"}`.
- `import_graph_load.go` — thin wrapper over `go/packages.Load` with
  `NeedName|NeedImports`, scoped to the module path; returns `[]pkg{path,
  imports}` or a load error.
- `import_graph_matrix.go` — pure logic: build package→layer map from globs,
  build allow-sets, validate config (unknown layer in `allow`, layer with no
  paths). No I/O — fully unit-testable.
- `import_graph_check.go` — pure logic: given packages + matrix, return the
  list of violating edges and unmapped packages. No I/O.
- Wire `checkImportGraph` into `RunWithFilter` behind
  `cfg.Gates.ImportGraph.Enabled`.

Each file ≤100 lines (G1). Pure logic split from I/O so the matrix/check
logic is tested without a real module load.

### Result semantics

- config error → `Fail` (message: `import_graph config: …`)
- ≥1 forbidden edge → `Fail` (details: one line per edge)
- only unmapped packages → `Warn`
- empty/disabled matrix → omit (disabled) or `Warn` (present but empty)
- all good → `Pass`

## Implementation slices

1. **Config schema** — add `ImportGraphConfig`/`Layer`, parse TOML, unit
   tests for parsing + defaulting `module` from `go.mod`.
2. **Matrix + check (pure logic)** — globs→layer map, allow-sets, edge
   check, unmapped detection, config validation; table-driven unit tests.
3. **Loader + wiring** — `go/packages` load, `checkImportGraph`, wire into
   `RunWithFilter`; integration test against a small fixture module.
4. **Dogfood** — add `[gates.import_graph]` to this repo's `centinela.toml`
   encoding the PROJECT.md G2 matrix; confirm `centinela validate` passes
   on the current tree, and fails on a deliberately-injected bad import.
5. **Acceptance** — executable acceptance artifact under `tests/acceptance/`
   driving validate against pass/fail fixtures, wired into `validate.commands`.

## Testing

- Unit: matrix building, allow-set checks, config validation, unmapped
  detection (`internal/gates/import_graph_*_test.go`, each ≤100 lines).
- Integration: real `go/packages` load against a fixture module with a known
  violation.
- Acceptance: `centinela validate` against a clean fixture (Pass) and a
  fixture with a forbidden edge (Fail), asserting exit status + message.

## Risks & mitigations

- `go/packages` load cost → load once, minimal `NeedName|NeedImports`.
- Diff-aware filtering unsound for graphs → v1 ignores the file filter and
  loads the whole module (documented).
- Uncompilable code → surface load errors as `Fail`, never false `Pass`.

## Gatekeeper / done

- All new files ≤100 lines.
- No G2 violations introduced by this feature (dogfood the gate on itself).
- `centinela validate` green (lint + type + full suite + coverage).
