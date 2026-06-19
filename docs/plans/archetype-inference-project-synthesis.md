# Plan: archetype-inference-project-synthesis

> Consume the shipped `analyze.Inventory` (`.workflow/analysis.json`) to infer
> the best-fit architecture archetype and synthesize a `PROJECT.md` draft,
> deterministically (no LLM). Reconciles big-thinker + feature-specialist.

## Layer decision (resolved)

`internal/synthesize` is an **aggregator** package (NOT domain): it imports the
domain package `internal/analyze` read-only, mirroring `internal/doctor`,
`internal/insights`, `internal/calibration`, `internal/audit`. A
`synthesize → analyze` edge is domain-from-aggregator (allowed); `analyze` never
imports `synthesize`, so no cycle. **centinela.toml**: add
`internal/synthesize/**` to the existing `aggregator` layer `paths`. **PROJECT.md**:
register `internal/synthesize` in the G2 prose + folder/layer/gatekeeper tables.
`internal/ui/render_synthesize.go` keeps `synthesize` free of any `ui` import.

## Inference model (deterministic)

`Infer(inv) Inference` sums per-archetype integer weights from a swappable rule
table over the Inventory, sorts deterministically (total desc, archetype asc),
and sets `Best` + `Confidence(high|medium|low)` + `Ambiguous`:
- **Framework signals (strongest):** `rails` gem / Ruby + `app/models|controllers|views`
  → rails-native; `django`/Flask, Express/Fastify, Go `handler|service|repository`
  → n-tier; `systems|components|entities`, game engines/`bevy` → ecs;
  `domain|application|infrastructure|ports|adapters` → hexagonal;
  `modules/*/public`+`modules/*/internal` → modular.
- **Graph shape** is a tiebreaker boost (may be empty for non-Go).
- Confidence from winning score + margin; tie within margin → `Ambiguous`, forced
  `low`, runners-up listed. Nothing scores → `Custom`, `low`.
- `Inferer` is an interface (`NewInferer()` → deterministic default) — the swap
  seam for a future LLM backend.

## PROJECT.md synthesis

`Draft(inv, inf) string` assembles sections (no I/O), prefixed with the
doc-version marker + a DRAFT/confidence banner:
- **Archetype-specific** (`sections_arch.go`, `archetypeProfile` table): Pattern,
  G2/G7 rule text (from architecture-overview.md), reference doc, Layer Mapping +
  Gatekeeper Paths rows. Layer rows are **derived** by bucketing `inv.Packages`
  into the archetype's abstract slots; unmatched slots → `<!-- TODO: confirm -->`.
- **Inventory-derived** (`sections_meta.go`): Tech Stack (language/framework/
  build/test), Folder Structure (tree from `Packages`), Locales, Naming
  Conventions (`namingByLang` table keyed on `PrimaryLanguage`). Human-only
  sections → guided stubs.

## Source files (each ≤100 lines, aggregator layer)

1. `internal/synthesize/archetype.go` — result types (`Archetype`, `Signal`,
   `Score`, `Inference`).
2. `internal/synthesize/rules.go` (+ `predicates.go` if needed) — the scoring
   rule table + predicates over `analyze.Inventory`.
3. `internal/synthesize/infer.go` — `Inferer` interface + `ruleInferer` +
   `NewInferer`; scoring/confidence/ambiguity.
4. `internal/synthesize/draft.go` — `Draft(inv, inf) string` orchestrator + banner.
5. `internal/synthesize/sections_arch.go` (+ `_table.go` if needed) — archetype
   profile table + layer-mapping/gatekeeper rendering.
6. `internal/synthesize/sections_meta.go` (+ `sections_naming.go` if needed) —
   tech-stack/folder/locales/naming sections.
7. `internal/synthesize/load.go` — `Load`? No — add `analyze.Load` in the
   contract owner; synthesize calls it. (`internal/analyze/load.go`.)
8. `internal/synthesize/write.go` — `WriteDraft(target, content)`: never
   overwrite `PROJECT.md`; write `PROJECT.draft.md` instead; report path+clobbered.
9. `cmd/centinela/synthesize.go` — thin Cobra command (mirror `analyze.go`);
   `--in`/`--out`/`--json`; actionable errors; no business logic (G7).
10. `internal/ui/render_synthesize.go` — `RenderInferenceSummary(inf)`.
11. `internal/analyze/load.go` — `Load(path) (Inventory, error)` with
    missing-file / malformed / schema-drift errors.

Config: `centinela.toml` aggregator paths += `internal/synthesize/**`; this
repo's `PROJECT.md` G2 tables register the package. Check the scaffold
`centinela.toml` mirror only if the parity test covers it.

## Test plan (per-package 95% → colocated `_test.go`, each ≤100)

- **Unit (`internal/synthesize`):** infer_test (fixture Inventory → Best/
  Confidence/Ambiguous/deterministic order: Go n-tier, Rails, ECS, hexagonal,
  modular, empty→Custom, tie→Ambiguous), rules/predicates_test, draft +
  sections_arch + sections_meta_test (section-content assertions + one small
  golden fragment), write_test (clobber-safety). `internal/analyze` load_test
  (missing/malformed/drift/round-trip via `analyze.Save`).
- **Command:** `cmd/centinela/synthesize_test.go` + `synthesize_errors_test.go`.
- **Integration:** `tests/integration/synthesize_pipeline_test.go` — real
  analyze→Save→Load→Infer→Draft→WriteDraft on a Go fixture.
- **Acceptance:** `tests/acceptance/synthesize_{helper,happy,edge}_test.go` —
  build binary once, run `analyze` then `synthesize` over fixture repos; carry
  `// Acceptance:`/`// Scenario:` traceability for all 7 spec scenarios.

All fixtures are committed/in-test `Inventory` values — no LLM, no network.

## Spec

`specs/archetype-inference-project-synthesis.feature` — 7 scenarios: Rails→
rails-native draft; Go→n-tier; game→ecs; ambiguous→low-confidence rationale;
missing analysis.json→actionable error; existing PROJECT.md→draft (no clobber);
deterministic re-run byte-identical.

## Risks

- Heuristic mis-classification → mitigated by confidence + rationale + draft-first
  confirm/correct model; inferer is swappable.
- Per-package 95% coverage → pure functions + fixtures.
- ≤100-line rule → ~11 source + ~12 test files, splits pre-planned.
