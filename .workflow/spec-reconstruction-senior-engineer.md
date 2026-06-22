# spec-reconstruction — senior-engineer

## Files Touched

| File | Lines | Purpose |
|------|------:|---------|
| internal/reconstruct/reconstruct.go | 52 | Package doc (aggregator contract) + result types Target, Reconstruction, Artifact, Role. |
| internal/reconstruct/signals.go | 61 | Lowercased/flattened signals view over analyze.Inventory (deps, frameworks, kinds, graph in-edges). |
| internal/reconstruct/rules.go | 55 | Exclude + promote rule tables (data, not control flow) + Role hints; maxTargets bound. |
| internal/reconstruct/slug.go | 58 | Deterministic slugify + collision disambiguation (stdlib-only, no strconv). |
| internal/reconstruct/select.go | 62 | Select(inv) []Target: exclude → promote → slug → disambiguate → sort → bound. |
| internal/reconstruct/templates.go | 60 | Role-keyed scenario/narrative templates as data tables; todoMarker constant. |
| internal/reconstruct/feature.go | 35 | featureSkeleton(t): role-aware Feature:/Scenario:/# TODO: confirm Gherkin; counts TODOs. |
| internal/reconstruct/brief.go | 21 | briefStub(t): docs/features/<slug>.md stub with honest TODO gaps. |
| internal/reconstruct/reconstructor.go | 32 | Reconstructor interface + ruleReconstructor + NewReconstructor() + Reconstruct orchestrator. |
| internal/reconstruct/write.go | 63 | WriteCorpus: skip-if-exists vs canonical specs/, MkdirAll, single-call writes; DefaultOutRoot. |
| cmd/centinela/reconstruct.go | 66 | Thin Cobra command (--in/--out/--json), ErrNoInventory guidance, auto-registered via init(). |
| internal/ui/render_reconstruct.go | 24 | RenderReconstructionSummary: targets/written/skipped/TODO totals — presentation only. |
| centinela.toml | — | Added internal/reconstruct/** to aggregator layer paths + rationale comment. |
| PROJECT.md | — | G2 prose (reconstruct aggregator + ui read-only allowance) + Folder Structure entry. |

All 12 source files are ≤100 lines (largest: write.go at 63).

## Architecture Compliance

- **Aggregator layer.** `internal/reconstruct` imports only `internal/analyze` (domain, read-only) and stdlib. Its sole edge `reconstruct → analyze` is allowed by the aggregator layer's `allow = ["domain", "leaf"]`; `analyze` never imports `reconstruct`, so no cycle. Registered in `centinela.toml` aggregator `paths` and PROJECT.md G2 before the import_graph gate runs.
- **No cmd/ or internal/ui import** from the package. `internal/ui/render_reconstruct.go` depends on `reconstruct` (read-only, for the Reconstruction render type) — the same direction as render_synthesize.go; PROJECT.md's ui allowance was extended accordingly.
- **G7 (thin outer layer).** `cmd/centinela/reconstruct.go` only wires flags, calls `analyze.Load` → `NewReconstructor().Reconstruct` → `WriteCorpus` → `ui.RenderReconstructionSummary`. No selection logic, classification, or string assembly in cmd/.
- **Structural mirror of synthesize.** Inferer→Reconstructor seam, signals/rules data tables, never-clobber WriteCorpus, ErrNoInventory wrap, init()+rootCmd.AddCommand registration.

## Type-Safety Notes

- Strict Go; `go build ./...` and `go vet ./...` both clean. No `interface{}`/`any`, no reflection.
- `Role` is a string newtype with three named constants; unknown/empty roles fall back to RoleModule via `templateFor`/`roleOrModule` so no target is ever skeleton-less.
- The rule tables are typed structs with `func` predicates (not stringly-typed dispatch). Map iteration is never used in output paths — selection sorts by slug, artifacts follow target order — guaranteeing byte-stability.

## Trade-Offs

- **One scenario per target (v1).** Each feature emits a single role-aware Scenario with three TODO-marked steps. Multi-scenario / HTTP-route / call-flow extraction is deferred to the roadmap (per plan) — the LLM seam (Reconstructor interface) carries that future lift.
- **Skip-if-exists checks canonical `specs/` even though the default --out is the review dir** (belt-and-suspenders against `--out specs`); a skipped target also suppresses its brief, so an augmented repo never gets an orphan brief for a hand-authored spec.
- **slug.go ships a tiny local `itoa`** to keep the package strconv-free and the disambiguation deterministic; trivial and fully covered by the disambiguation path.

## Dry-Run Evidence

Built `/tmp/centinela-srecon`, ran `analyze` then `reconstruct` against the worktree's own inventory (56 packages, 143 graph edges):
- 31 targets selected, 62 files written, **93 TODO markers** (3 per feature).
- Two consecutive runs produced **byte-identical** output (shasum match).
- All 31 generated `.feature` files parse with the **real** `spec_traceability` parser (`parseScenarios` returned 31 scenarios); zero Given/When/Then steps lack a `# TODO: confirm` marker (no fabrication).
- Skip-if-exists: a hand-authored `specs/internal-analyze.feature` → 1 skipped, 60 written, canonical file byte-unchanged.
- No inventory: exit 1 with "run `centinela analyze` first". Test output removed; `.workflow/analysis.json` is gitignored.

## Handoff

→ **qa-senior**. Implement tests/ tier + colocated `_test.go` (each ≤100, per-package 95% coverage) per the plan's test plan: select_test (fixture Inventories → sorted Targets, exclusion precedence, slug collision, maxTargets bound, empty/polyglot), feature_test (real spec_traceability parse + no-fabrication + golden fragment), brief_test, reconstructor_test (determinism + TodoCount), write_test (skip-if-exists, re-run byte-identical), cmd reconstruct happy/--json/errors, integration reconstruct_pipeline, and acceptance reconstruct_{helper,happy,edge} carrying `// Scenario:` traceability comments for all 9 spec scenarios. Add the acceptance run to `validate.commands`. Public seams: `reconstruct.NewReconstructor()`, `Select`, `WriteCorpus`, `DefaultOutRoot`, `ui.RenderReconstructionSummary`.
