# Big-Thinker Report — spec-reconstruction

## Problem

Centinela is spec-first: the `plan` step demands a `.feature` spec and the
`spec_traceability` gate fails `validate` unless every acceptance scenario maps
to an executable test. A mature codebase adopting Centinela has **zero specs** —
behavior lives only in code and in heads — so every spec-first gate has nothing
to anchor to, and the team faces hand-authoring dozens of `.feature` files for
behavior that already exists. `spec-reconstruction` is the third consumer of the
frozen `analyze.Inventory` (after `analyze` = *what the repo is* and `synthesize`
= *draft PROJECT.md*): a deterministic `centinela reconstruct` derives a behavioral
spec corpus skeleton. `brownfield-roadmap-generation` `dependsOn` it (it must
distinguish already-built capability from net-new work). **Why now:** the
Inventory contract is frozen and proven; this turns "what the repo is" into "what
the repo does."

## Scope

**In:**
- New aggregator package `internal/reconstruct/` reading `internal/analyze`
  read-only via the existing `analyze.Load` seam.
- Deterministic `Select(inv) []Target` over a data-driven promote/exclude rule
  table (mirrors `synthesize/rules.go`).
- `Reconstructor` interface + deterministic `NewReconstructor()` default —
  the swappable LLM seam, **no in-process LLM call**.
- Per target: a `specs/<slug>.feature` Gherkin skeleton (`Feature:` + role-aware
  `Scenario:` stubs with `# TODO: confirm`) + a `docs/features/<slug>.md` brief
  stub. Pure string assembly, byte-stable.
- Write to a **review dir** (`.workflow/reconstructed/`), skip-if-exists against
  canonical `specs/`; `--in`/`--out`/`--json` flags mirroring `synthesize`.
- `cmd/centinela/reconstruct.go` (thin) + `internal/ui/render_reconstruct.go`.
- `centinela.toml` + `PROJECT.md` aggregator registration.

**Out:**
- Framework-specific HTTP route / call-flow extraction (deferred —
  `brownfield-route-flow-extraction`).
- Any in-process LLM inference (explicitly excluded by the decided approach).
- New persisted schema (inputs are the frozen Inventory; outputs are text files).
- Changes to `internal/analyze` (the `Load`/`ErrNoInventory` seam is reused).
- Scaffold-asset toml mirror edit (the generic template carries no aggregator
  paths — verified).
- The downstream `brownfield-roadmap-generation` consumer itself.

## Dependencies & Assumptions

- **Depends on:** the shipped, frozen `analyze.Inventory` contract + `analyze.Load`
  / `analyze.ErrNoInventory`; the existing `spec_traceability` Gherkin parser
  (`Feature:` line + `^\s+Scenario:`); the aggregator layer already present in
  `centinela.toml` and PROJECT.md G2.
- **Assumes:** `synthesize` is the structural template (rules/signals/write/cmd/
  render split) and may be copied in shape. Per-package 95% coverage is met by
  pure functions + fixture Inventories.
- **Assumes:** writing to a review dir (not `specs/` directly) is the accepted
  clobber-safety stance, consistent with `synthesize`'s `PROJECT.draft.md`.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Thin skeletons feel like noise | Medium | Medium | Rule table selects *meaningful* targets; role-aware scenario stubs; LLM seam carries behavioral lift |
| Clobbering hand-authored `.feature` (data loss) | High | Low | Review-dir default + skip-if-exists against canonical `specs/`; explicit acceptance test |
| Layering / import-graph regression | Medium | Low | Map `internal/reconstruct/**` as aggregator in `centinela.toml` + PROJECT.md *before* code; sole edge `reconstruct → analyze` |
| Gherkin shape drift (specs invisible to gate) | High | Low | Test parses every generated file with the *real* `spec_traceability` parser |
| Non-determinism (map order / unsorted) breaks re-run diff | Medium | Medium | Sorted targets, stable section order, no map iteration; byte-identical re-run test |
| Slug collisions overwrite one another | Medium | Low | Deterministic path-derived disambiguation; collision fixture test |
| Scope creep into route/flow extraction | Medium | Medium | v1 = package/manifest targets only; route extraction deferred to roadmap |

## Rollout (smallest correct slice first)

1. **Layer + contract first:** register `internal/reconstruct/**` in
   `centinela.toml` + PROJECT.md; create `reconstruct.go` (types) + `signals.go`.
2. **Selection:** `rules.go` + `select.go` (+ slug) with `select_test.go` — the
   deterministic target set is the load-bearing decision.
3. **Skeletons:** `feature.go` + `brief.go` + `templates.go` producing
   parser-valid Gherkin + brief stubs (tested against the real parser).
4. **Reconstructor + write:** `reconstructor.go` (interface/seam) + `write.go`
   (review dir, skip-if-exists, byte-stable).
5. **Wiring last:** `cmd/centinela/reconstruct.go` + `render_reconstruct.go`.
6. Integration + acceptance over committed fixture repos; full traceability.

Each slice is independently testable; the seam (`Reconstructor` interface) lands
before `cmd/` so an LLM backend can drop in without touching `cmd/`.

## Deferred Findings

- **`brownfield-route-flow-extraction`** — framework-specific HTTP route / call-
  flow extraction across web frameworks is unbounded; v1 derives targets from
  packages + manifests only. Deferred to the roadmap via
  `centinela roadmap defer`.

## Handoff

→ **feature-specialist.** Plan resolves all five open decisions from the brief:
layer (aggregator), output location (`.workflow/reconstructed/` review dir +
skip-if-exists), skeleton shape (role-aware `Feature:` + `# TODO: confirm`
scenarios), brief-vs-spec (both per target), and the rule-table selection model.
Feature-specialist should author `specs/spec-reconstruction.feature` (9 scenarios
per the acceptance criteria) and tighten the per-file source split if any file
risks exceeding 100 lines.
