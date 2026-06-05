### Big-Thinker Report: g2-import-graph-gate
**Date:** 2026-06-03

#### Problem

Centinela's G2 layer-boundary rule exists only as prose in `PROJECT.md`
(`internal/config` is a leaf, `internal/workflow`/`internal/gates` may import
`internal/config` only, `internal/ui` is presentation, etc.). Nothing
mechanically enforces it. An agent — or a hurried human — can silently invert
the dependency stack (e.g. `internal/config` importing `internal/ui`) and no
gate catches it; the gatekeeper subagent reviews prose and intent, not the
actual compiled import graph, so violations ship.

**Who is hurting:** (1) the maintainer who defined the architecture and now
relies on human review to defend it, and (2) the agent doing the work, who gets
no fast, specific signal and only learns of the violation at PR review (or
never). **Why now:** Roadmap Phase 2 is explicitly "convert the remaining
'requested' gates into mechanically enforced ones." This gate is the anchor
feature of that phase — `custom-gate-sdk` and the layer-erosion follow-ups both
depend on it. The gate interface (`internal/gates`) and the diff-aware validate
flow are now stable enough to extend cleanly.

#### Scope (In / Out)

**In (v1):**
- Go-only import-graph analysis via `go/packages` (`NeedName|NeedImports`).
- Layer allow/deny matrix as **structured config** in a `[gates.import_graph]`
  block in `centinela.toml` (layers = path globs + per-layer allow-list).
- A single new `import_graph` `gates.Result`, wired into `RunWithFilter` behind
  `cfg.Gates.ImportGraph.Enabled`, following the existing enable/skip
  conventions (mirrors how `checkBuild` is gated by `Build.Enabled`).
- Result semantics: forbidden edge → Fail (one Details line per edge);
  unmapped package → Warn; config error → Fail (distinct message); empty
  matrix → Warn; disabled/absent block → omit; load error → Fail.
- Dogfood: encode this repo's PROJECT.md G2 matrix and confirm validate stays
  green on the current tree.

**Out (explicit, fixed by user decisions):**
- TypeScript (ts-morph/madge) and Python (AST) parsers — config schema stays
  language-generic but only the Go parser ships.
- Matrix parsed from PROJECT.md prose — the matrix is config, full stop.
- Pre-write hook enforcement — the `centinela validate` gate is the only
  enforcement point in v1.
- PROJECT.md ↔ config drift checker (future `custom-gate-sdk` territory).

#### Dependencies & Assumptions

- **`golang.org/x/tools` (go/packages) — NEW external dependency.** Confirmed
  NOT present in `go.mod`/`go.sum`; there is no `vendor/` dir. This adds a real
  supply-chain surface and requires `go mod tidy`; CI must stay tidy-clean.
- `internal/gates` — extend `RunWithFilter`; reuse `Result`/`Status`. The gate
  takes the `filter *gitdiff.Set` argument for signature parity but **must
  ignore it** (whole-module load).
- `internal/config` — extend `GatesConfig` with `ImportGraph ImportGraphConfig`
  (`toml:"import_graph"`), following the `BuildGateConfig` pattern incl. a
  `NormalizeImportGraph` for defaulting `module` from `go.mod` and dropping
  malformed layers.
- `cmd/centinela/validate.go` — no change expected; it already calls
  `gates.RunWithFilter(cfg, filter)` and renders all results uniformly.
- `go.mod` — read the module path (`github.com/samuelnp/centinela`) to scope
  "internal" imports vs std-lib/third-party.
- Prior features: diff-aware gatekeeper (the filter this gate deliberately
  bypasses), G1 file-size gate (forces the load/matrix/check/orchestrate
  split), build gate (the structural template for an optional config-driven
  gate).
- Assumption: `go/packages.Load` works against the module root from the
  validate CWD; acceptance fixtures need a self-contained mini Go module.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Diff-aware filter applied to graph → false Pass (edge fixed/broken by a file outside the diff) | High | Medium | v1 ignores the filter and loads the whole module; document loudly; integration test that a violation outside the diff set still Fails |
| New `golang.org/x/tools` dep bloats build / drifts go.sum / fails CI tidy check | Medium | Medium | Pin version, run `go mod tidy`, verify cross-platform build gate still green; minimal `Need*` flags |
| Dogfooding a stricter-than-reality matrix breaks current `centinela validate` (regresses every other feature's validate) | High | Medium | Encode PROJECT.md G2 exactly; run validate on the clean tree before committing the centinela.toml block; gate disabled-by-default so absence is safe |
| Uncompilable code returns false Pass instead of Fail | High | Low | Treat any `packages.Load` error / `pkg.Errors` as Fail with the load message |
| go/packages load latency regresses validate runtime | Medium | Medium | Single load, `NeedName|NeedImports` only, reuse the package set |
| Files exceed G1 100-line cap | Low | Medium | Pre-planned split: load (I/O) vs matrix/check (pure) vs orchestrate |
| Unmapped package silently passes, hiding new untracked layers | Medium | Low | Criterion 5: unmapped → Warn, surfaced explicitly |
| Test-package layer mis-classification (external `_test` pkgs) couples layers undetected | Medium | Low | Map `_test` external packages to the package under test's layer; unit-test this case |

#### Rollout

- **Step 1 — Config schema (smallest correct slice).** Add
  `ImportGraphConfig{Enabled, Module, Layers}` + `Layer{Name, Paths, Allow}` to
  `internal/config`, TOML decode + `NormalizeImportGraph` (default module from
  go.mod, drop layers with no paths). Unit tests for parse + default. No gate
  behavior yet — fully safe, nothing wired in.
- **Step 2 — Matrix + check (pure logic, no I/O).** Build package→layer map
  from globs, allow-sets, edge check, unmapped detection, config validation
  (unknown layer in allow, empty matrix). Table-driven unit tests — this is the
  correctness core and is fully testable without a real module load.
- **Step 3 — Loader + wiring.** Thin `go/packages.Load` wrapper; add the
  `golang.org/x/tools` dep; `checkImportGraph(cfg, filter)` (ignoring filter)
  wired into `RunWithFilter` behind `Enabled`. Integration test against a small
  fixture module with a known violation + a load-error fixture.
- **Step 4 — Dogfood.** Add `[gates.import_graph]` to this repo's
  `centinela.toml` encoding the PROJECT.md G2 matrix; confirm validate stays
  green on the clean tree and Fails on a deliberately injected bad import.
- **Step 5 — Acceptance.** Executable acceptance artifact under
  `tests/acceptance/` driving validate against pass/fail fixtures, asserting
  exit status + the `<importer> → <forbidden> (...)` message; wire into
  `validate.commands`.

**Can wait:** TS/Python parsers, pre-write hook, drift checker — all out of v1.

#### Handoff (Next role: feature-specialist; Outstanding questions)

- **Module-path scoping:** confirm the rule for distinguishing the module's own
  packages from std-lib/third-party is a simple `strings.HasPrefix(importPath,
  module)` — adequate for v1 single-module repos.
- **Gate Name string:** existing gates use human labels (e.g. "G-Build:
  Cross-Compile") but criterion 1 says the Result is named `import_graph`.
  Confirm the exact `Result.Name` the validate renderer should display.
- **Disabled-by-default:** confirm the gate is opt-in (absent block → omitted,
  like Build) so adding the dep can't regress projects that haven't configured
  it.
- **`go/packages` dep approval:** flag that this is the first `golang.org/x/`
  dependency; confirm the maintainer accepts it (vs. a lighter `go list -json`
  subprocess alternative the specialist may want to weigh).
