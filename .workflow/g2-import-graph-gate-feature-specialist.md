### Feature-Specialist Report: g2-import-graph-gate
**Date:** 2026-06-03

#### Behavior Summary

The `g2-import-graph-gate` feature adds a mechanical enforcement gate to `centinela validate` that parses the Go module's import graph using `go/packages` and checks every intra-module import edge against a per-layer allow matrix configured under `[gates.import_graph]` in `centinela.toml`. When enabled, the gate loads all module packages (with `NeedName|NeedImports` flags), maps each package to a layer via path-glob matching, and reports `Fail` for any edge where the importing package's layer does not permit importing the imported package's layer — listing each violating edge as `<importer-pkg> → <forbidden-pkg> (<importer-layer> may not import <imported-layer>)`. Unmapped packages produce a `Warn`, config errors produce a `Fail` with a `"import_graph config:"` prefix distinct from an import violation, and an absent or `enabled = false` block omits the gate entirely from the result set, following the same opt-in convention as the existing Build gate. Standard-library and third-party imports are ignored; test packages map to the layer of their package under test; intra-layer imports are unconditionally allowed. The gate deliberately ignores the diff-aware file filter and performs a whole-module load to avoid false passes from violations outside the changed set.

#### Gherkin Scenarios (see `specs/g2-import-graph-gate.feature`)

1. **All imports respect the layer matrix — gate passes** — happy path; all packages within their allow bounds → `Pass`.
2. **A package imports a layer it is not allowed to import — gate fails** (Scenario Outline, 3 examples) — each forbidden edge is listed; exit code 1.
3. **Multiple forbidden edges are all listed in the failure output** — ensures every violation, not just the first, appears in Details.
4. **No `[gates.import_graph]` block present — gate is omitted** — absent block produces no `import_graph` result.
5. **Gate explicitly disabled with `enabled = false` — gate is omitted** — explicit disable also produces no result.
6. **A package matches no configured layer — gate warns** — unmapped package → `Warn`, not silent Pass.
7. **Malformed `[gates.import_graph]` config — gate fails with config error** (Scenario Outline, 3 examples) — config errors carry `"import_graph config:"` prefix, never the arrow format.
8. **Block present with no layers defined — gate warns rather than silently passing** — empty matrix → `Warn`.
9. **The module contains uncompilable code — gate fails with load error** — `go/packages` load failure → `Fail`, not false Pass.
10. **A package imports standard-library and third-party packages — not flagged** — non-module imports are excluded from matrix evaluation.
11. **An external test package (`_test` suffix) imports across a forbidden layer boundary** — `_test` packages inherit the base package's layer; violations are caught.
12. **A package imports another package in the same layer — always allowed** — intra-layer imports are unconditionally permitted.
13. **A violation exists outside the current diff set — gate still fails** — whole-module load; diff filter is ignored.

#### UX States

| State   | Trigger                                                                          | Surface (CLI output)                                                                                           |
|---------|----------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------|
| Pass    | All intra-module imports respect the configured layer allow matrix                | `import_graph   PASS` (one line in validate summary); exit 0                                                  |
| Fail    | One or more forbidden cross-layer edges detected                                 | `import_graph   FAIL` + one Details line per edge: `<importer> → <forbidden> (<layer> may not import <layer>)`; exit 1 |
| Fail    | Config is malformed (empty paths, unknown allow-list layer, empty module path)   | `import_graph   FAIL` + Details starting with `import_graph config: <reason>`; exit 1                        |
| Fail    | `go/packages` load error (uncompilable code, missing module)                    | `import_graph   FAIL` + Details containing the loader error message; exit 1                                   |
| Warn    | One or more packages match no configured layer path glob                         | `import_graph   WARN` + Details listing unmapped package import paths; exit 0                                  |
| Warn    | Block present but zero layers defined (empty matrix)                             | `import_graph   WARN` + Details indicating the layer matrix is empty; exit 0                                   |
| Omitted | No `[gates.import_graph]` block in `centinela.toml`, or `enabled = false`       | No `import_graph` line in validate output; gate is absent from results entirely                               |

#### Out-of-Scope

- **TypeScript and Python import parsers** — config schema is language-generic but only the Go (`go/packages`) parser ships in v1; ts-morph/madge and Python AST parsers are explicitly deferred.
- **Pre-write hook enforcement** — the gate runs only at `centinela validate` time; there is no pre-file-write guard in v1.
- **PROJECT.md vs config drift checker** — detecting when the prose layer rules in `PROJECT.md` diverge from the `[gates.import_graph]` config is out of scope; deferred to the `custom-gate-sdk` roadmap item.
- **Multi-module repos** — v1 assumes a single-module repo; the `strings.HasPrefix(importPath, module)` scoping rule is not validated against workspace (`go.work`) setups.
- **Cycle detection** — import cycles are a `go build` error and are caught by the existing Build gate; this gate only checks the allow matrix, not circularity.

#### Handoff

**Next role:** senior-engineer

**Open clarifications (carry forward to implementation):**

1. **`go/packages` vs `go list -json` subprocess** — The big-thinker flagged this as an unresolved dependency decision. `go/packages` is the canonical, version-aware API but adds `golang.org/x/tools` as the first `golang.org/x/` dependency in `go.mod`. A `go list -json ./...` subprocess avoids the new dep but introduces shell-out fragility and loses structured error types. The senior engineer must confirm the approved approach with the maintainer before adding the dependency.
2. **Result.Name display label** — Existing gates use human labels (e.g. `"G-Build: Cross-Compile"`). The spec uses `import_graph` as the `Result.Name`. Confirm whether the validate renderer should display a human-friendly label like `"G2: Import Graph"` or the raw config key `import_graph`.
3. **Module-path scoping rule** — Confirm `strings.HasPrefix(importPath, cfg.Module)` is sufficient for v1 single-module repos (i.e. no edge cases with module path being a prefix of a third-party import path sharing the same prefix).
4. **`golang.org/x/tools` version pin** — Once approved, pin the minimum version that includes `go/packages` with the `NeedName|NeedImports` flags used by this feature, and verify `go mod tidy` keeps the CI tidy-check green.
