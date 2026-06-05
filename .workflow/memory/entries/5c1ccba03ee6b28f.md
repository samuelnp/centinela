---
id: 5c1ccba03ee6b28f
feature: g2-import-graph-gate
step: validate
type: verdict
title: ### Gatekeeper Report: g2-import-graph-gate
tags: gatekeeper, verdict
sourceArtifact: .workflow/g2-import-graph-gate-gatekeeper.md
createdAt: 2026-06-05T15:34:55Z
---

### Gatekeeper Report: g2-import-graph-gate
**Date:** 2026-06-05
**Status:** SAFE

#### Analyzed Specs

| Spec file | Relevance |
|-----------|-----------|
| `specs/g2-import-graph-gate.feature` | Primary spec for this feature — 14 scenarios covering pass, fail, warn, disabled, malformed config, intra-layer, diff-aware bypass, and load error |
| `specs/g1-justified-file-size-exceptions.feature` | Existing G1 gate spec — reviewed for conflict with gates.Result / RunWithFilter |
| `specs/cross-platform-build-gate.feature` | Existing G-Build gate spec — reviewed for conflict with RunWithFilter / AllPassed |
| `docs/plans/g2-import-graph-gate.md` | Implementation plan — verified architecture decisions align with spec |

#### Findings

**G1 (file size):** All new source files are within the 100-line limit. The largest file is `import_graph_check.go` at 97 lines, `import_graph_test.go` at 69 lines, `import_graph.go` at 83 lines. All pass G1 without exception.

**Layer boundaries (G2):** The new code is in `internal/gates/` and `internal/config/` — both are correctly-placed domain/leaf layers per PROJECT.md. `internal/gates/import_graph*.go` imports only `internal/config` and `internal/gitdiff` (both leaf-layer), consistent with the G2 rule: domain may import leaf only. No `cmd/` or `internal/ui` imports appear in the new files.

**`gates.Result` contract:** The `Result` struct (Name, Status, Message, Details) is unchanged. The new gate returns a standard `Result{Name: "import_graph"}`. No breaking change to the public gate surface.

**`RunWithFilter` signature/behavior:** The function signature `RunWithFilter(cfg *config.Config, filter *gitdiff.Set) []Result` is unchanged. The new gate is conditionally appended at the end of the function body (`if cfg.Gates.ImportGraph.Enabled`), leaving all existing gate branches untouched. A `nil` filter does not change existing gate behavior.

**`GatesConfig` struct:** `ImportGraph ImportGraphConfig` is a new additive field with TOML tag `import_graph`. No existing fields were renamed or removed. Backward compatibility: projects without a `[gates.import_graph]` block will have `Enabled = false` (zero value), so the gate is silently omitted — no regression for existing callers.

**Conflict review — existing gate specs:**
- `g1-justified-file-size-exceptions.feature`: tests `checkFileSize`. That codepath is unmodified (`file_size.go`, `file_size_exceptions.go` untouched in this feature's diff). No conflict.
- `cross-platform-build-gate.feature`: tests `checkBuild`. `build.go` and `build_runner.go` are unmodified. No conflict.
- Neither existing spec exercises `RunAll` in a way that would be broken by appending a new gate — `RunAll` delegates to `RunWithFilter(cfg, nil)` which is unchanged.

**Spec consistency — g2-import-graph-gate.feature vs implementation:**
- Scenario "All imports respect the layer matrix → Pass": `runImportGraph` returns Pass when `violations == 0` and `unmapped == 0`. ✓
- Scenario "forbidden edge → Fail": `checkEdges` formats `"<from> -> <to> (<fromLayer> may not import <toLayer>)"` and gate returns Fail. The spec expects the arrow `→` (U+2192) but the implementation uses `->` (ASCII). This is a spec-vs-implementation mismatch in the violation format string. The gate's Details lines use `->`, not `→`. **This is a cosmetic discrepancy already present in the test assertions** (`strings.Contains(..., "a -> b (...)")`) — the tests pass and spec language is descriptive, not a byte-exact assertion. No blocking issue, but the spec prose should be updated to use `->` for clarity.
- Scenario "No block present → omitted": `cfg.Gates.ImportGraph.Enabled` defaults to false → gate skipped. ✓
- Scenario "enabled = false → omitted": same as above. ✓
- Scenario "Empty matrix → Warn": `len(g.Layers) == 0` → `Warn`. ✓
- Scenario "Malformed config → Fail with import_graph config:": `buildMatrix` error prefixed with `"import_graph config: "`. ✓
- Scenario "Load error → Fail": `loadPackages()` error → `Fail`. ✓
- Scenario "Unmapped → Warn": `len(unmapped) > 0` path returns `Warn`. ✓
- Scenario "Intra-layer always allowed": `m.allowed(from, to)` returns true when `from == to`. ✓
- Scenario "Test files map to package-under-test layer": `scopePackages` folds `TestImports` and `XTestImports` into the base package's import set. ✓
- Scenario "Diff-aware bypass → gate still fails": `checkImportGraph` deliberately ignores the `filter` argument and always loads the whole module. ✓

**No business logic in outer layer:** `cmd/centinela/` was not modified by this feature.

**i18n:** English-only project; gate is disabled. Not applicable.

**G1 exceptions audit:** No G1 exceptions were added or modified.

#### Recommendation

SAFE — No conflicts with any existing spec. The feature adds a new optional gate behind an `Enabled` flag, adds a new `[gates.import_graph]` config block, and wires `checkImportGraph` into `RunWithFilter` without altering any existing gate's behavior, the `Result` contract, or any other spec's scenarios. The arrow-format cosmetic discrepancy (`->` vs `→`) is pre-existing between spec prose and test assertions; it does not affect correctness and does not block. Proceed to validation.
