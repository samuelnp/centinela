# Feature Brief — spec-reconstruction

> Phase 9: Brownfield Onboarding. The third consumer of the shipped
> `analyze.Inventory` (`.workflow/analysis.json`). Where `analyze` captures *what
> the repo is* and `synthesize` drafts *PROJECT.md*, `spec-reconstruction`
> reconstructs the **behavioral contract**: a deterministic `centinela
> reconstruct` derives one `specs/*.feature` Gherkin skeleton + one
> `docs/features/*.md` brief stub per significant module/surface in the
> Inventory, with honest `# TODO: confirm` gaps for behavior the scan cannot
> know. The operating agent (LLM) then fills the scenarios — the same
> "deterministic skeleton + swappable inference seam" philosophy `synthesize`
> established, **not** an in-process LLM call.

## Problem

Centinela is spec-first: the `plan` step demands a `.feature` spec, and the
`spec_traceability` gate fails `validate` unless every acceptance scenario maps
to an executable test. A mature codebase adopting Centinela has **zero specs** —
the behavior lives only in code and in the team's heads — so every spec-first
gate has nothing to anchor to, and the team faces authoring dozens of `.feature`
files by hand to describe behavior that already exists and works. That is
exactly the error-prone, drifts-from-reality busywork the framework exists to
eliminate. `brownfield-roadmap-generation` (next in Phase 9) `dependsOn` this:
it must distinguish already-built capability from net-new work, and "already
built" is precisely what a reconstructed spec corpus records. **Why now:**
`analyze` and `synthesize` shipped; the Inventory contract is frozen and proven;
this is the next slice that turns "what the repo is" into "what the repo does."

## User value

`centinela reconstruct` reads the Inventory and emits, into a **review
directory** (never clobbering hand-authored specs), a scaffolded behavioral
corpus: for each significant module/package (and detected CLI command / HTTP
endpoint surface where the Inventory exposes it) a `specs/<slug>.feature` with a
`Feature:` block, a derived narrative, and one or more `Scenario:` skeletons
seeded from the module's role, plus a matching `docs/features/<slug>.md` brief
stub. Every gap the deterministic scan cannot fill is an explicit `# TODO:
confirm` marker, not a fabricated assertion. The team reviews, the operating
agent fleshes out Given/When/Then from the real code, and a real spec corpus
exists in minutes instead of days — confirmable, editable, and ready to anchor
the brownfield roadmap and the spec-traceability gate.

## How it works (mechanism)

1. **Load** — read `.workflow/analysis.json` via `analyze.Load` (the same seam
   `synthesize` uses). Surface `analyze.ErrNoInventory` with a "run `centinela
   analyze` first" message; never crash on a missing/old inventory.
2. **Select reconstruction targets** — deterministically pick the modules/
   surfaces worth a spec from `inv.Packages` + `inv.Graph` + `inv.Manifests`
   (e.g. top-level packages that own behavior; CLI subcommands or HTTP routes
   when the manifest/graph exposes them). Noise (vendored, generated, test-only,
   leaf-config) is excluded by a data-driven rule table, mirroring
   `synthesize`'s rules approach.
3. **Reconstruct skeletons** — for each target, build a byte-stable
   `specs/<slug>.feature` (Gherkin `Feature:` + narrative + `Scenario:`
   skeletons with `# TODO: confirm` Given/When/Then) and a `docs/features/
   <slug>.md` brief stub. Generation is pure string assembly (no I/O), behind a
   swappable `Reconstructor` interface whose default is the deterministic rule
   backend — the drop-in seam for a future LLM backend.
4. **Write under review, never clobber** — write to a dedicated output dir
   (proposed: `.workflow/reconstructed/specs/` + `.../features/`), or skip any
   target whose real `specs/<slug>.feature` already exists, so a partially
   spec'd repo is augmented, never overwritten. Print a concise summary
   (targets found, files written, TODO count) to stdout.
5. **Re-runnable & deterministic** — a second run on an unchanged Inventory
   produces byte-identical output (sorted targets, stable section order, no map
   iteration order), so it diffs cleanly and is safe to commit/review.

## Key decisions to resolve in the plan

- **Layer placement.** `internal/reconstruct/` is an **aggregator** (reads the
  `internal/analyze` domain read-only, like `internal/synthesize`). Confirm the
  `synthesize → analyze`-style single allowed edge, add `internal/reconstruct/**`
  to the `aggregator` layer `paths` in `centinela.toml`, register it in
  PROJECT.md (G2 prose + folder/layer/gatekeeper tables), and mirror the toml
  change into `internal/scaffold/assets` if scaffolded. `internal/ui/
  render_reconstruct.go` keeps `reconstruct` free of any `ui` import.
- **Target-selection rule table.** Which Inventory signals promote a package to
  a spec target, and which exclude it (test-only, generated, config leaf,
  vendored). Data-driven table like `synthesize/rules.go`, so adding heuristics
  is a table edit.
- **Output location & clobber policy.** Review dir under `.workflow/` vs writing
  `specs/` directly with skip-if-exists. The brief should fix this and the
  `--out`/`--in` flags (mirror `synthesize`).
- **Skeleton shape.** How many `Scenario:` stubs per target and what narrative
  the `Feature:` block carries, given the scan only knows structure, not
  behavior. Honest `# TODO: confirm` over fabricated assertions.
- **Brief stub vs spec.** Whether every target gets both a `.feature` and a
  `docs/features/*.md` stub, or briefs only for top-level surfaces.

## Acceptance Criteria

1. `centinela reconstruct`, run against a repo with a valid
   `.workflow/analysis.json`, writes at least one `specs/<slug>.feature`
   skeleton and one `docs/features/<slug>.md` brief stub into the review output
   location for the selected targets.
2. Every generated `.feature` is valid Gherkin parseable by the existing
   `spec_traceability` scenario parser (a `Feature:` line + ≥1 `Scenario:`).
3. Behavior the scan cannot determine appears as explicit `# TODO: confirm`
   markers — the tool never fabricates a concrete Given/When/Then assertion.
4. Output is **deterministic**: a second run on an unchanged Inventory produces
   byte-identical files (sorted targets, stable section order).
5. The command **never clobbers** a hand-authored `specs/<slug>.feature`: it
   writes to a review dir or skips existing targets, and says which it did.
6. Missing/old inventory surfaces `analyze.ErrNoInventory` with a "run
   `centinela analyze` first" message and a non-crashing non-zero exit.
7. A concise stdout summary reports targets selected, files written, and total
   TODO markers.
8. The `Reconstructor` is an interface with a deterministic default backend
   (`NewReconstructor()`), so an LLM backend can drop in without touching `cmd/`.
9. All new source files ≤100 lines; `internal/reconstruct/**` is mapped as an
   aggregator and introduces no cross-layer import violation.

## Edge Cases

- **No inventory** → `ErrNoInventory`, "run `centinela analyze` first", exit ≠0,
  no files written.
- **Empty/doc-only inventory** (no behavioral packages) → zero targets selected;
  exit 0; summary reports "0 targets" rather than emitting an empty `.feature`.
- **Target whose `specs/<slug>.feature` already exists** → skipped (or routed to
  review dir); never overwritten; reported as skipped.
- **Re-run after output exists** → byte-identical overwrite of the review
  artifacts; no spurious diff when the Inventory is unchanged.
- **Polyglot / non-Go inventory** (empty Go graph, declared-deps only) → still
  selects targets from `inv.Packages`/manifests; degrades gracefully.
- **Slug collisions** (two packages → same slug) → disambiguated deterministically
  so two targets never fight over one file.
- **Huge package list** → target count is bounded/sorted so output stays small
  and the review set is reviewable.
- **Target with no inferable behavior** → emits a `Feature:` + a single
  `# TODO: confirm` scenario stub, not an empty or assertion-bearing file.

## Data Model

No new persisted schema is strictly required; the inputs are the frozen
`analyze.Inventory` and the outputs are `.feature` / `.md` text files. The plan
may introduce small in-package types:

```go
// internal/reconstruct
type Target struct {
    Slug      string   // file stem, deterministic, collision-disambiguated
    Package   string   // source package/module this documents
    Role      string   // inferred role hint (command | endpoint | module)
    TodoCount int      // # of "# TODO: confirm" markers in the skeleton
}

type Reconstruction struct {
    Targets []Target   // sorted, what was selected
    Written []string   // repo-relative paths written
    Skipped []string   // targets skipped because a real spec already exists
}
```

Target-selection rules and skeleton templates live as data tables (mirroring
`synthesize/rules.go` + `synthesize/sections_*.go`) so adding heuristics or
adjusting skeleton shape is a table edit, keeping files ≤100 lines.

## Integration Points

- **Reads `analyze.Load(.workflow/analysis.json)`** — the same consumer seam
  `synthesize` uses; no new edge into `analyze` internals.
- **Logic in `internal/reconstruct/`** — select, reconstruct, marshal text;
  aggregator layer; imported only by `cmd/` (and its result type by
  `internal/ui` for rendering).
- **Command**: new `cmd/centinela/reconstruct.go` — thin wiring (flags → load →
  reconstruct → render → write files), mirroring `cmd/centinela/synthesize.go`.
- **Render**: `internal/ui/render_reconstruct.go` for the stdout summary
  (presentation only; `reconstruct` returns the typed `Reconstruction`).
- **Gherkin compatibility**: generated `.feature` files must parse with
  `internal/gates/spec_traceability_parse.go` so reconstructed specs are
  first-class citizens of the existing gate.
- **Downstream consumer**: `brownfield-roadmap-generation` (next, blocked on
  this) reads the reconstructed corpus to mark already-built capability.

## Risks

- **Thin skeletons add little value** — a `Feature:` + a lone TODO per module
  could feel like noise. Mitigated by the rule table choosing *meaningful*
  targets and seeding role-aware scenario stubs; the LLM seam carries the real
  behavioral lift.
- **Clobbering hand-authored specs** — overwriting a real `.feature` is data
  loss. Mitigated by the review-dir / skip-if-exists policy frozen in the plan
  and an explicit acceptance test.
- **Layering / import-graph regression** — `internal/reconstruct/**` is not in
  the layer matrix yet; an unmapped package warns and any stray edge fails the
  gate. Mitigated by mapping it as an aggregator in the plan and mirroring the
  toml into `internal/scaffold/assets`.
- **Scope creep on "endpoints and flows"** — extracting real HTTP routes / call
  flows is unbounded across frameworks. v1 ships package/manifest-derived
  targets and **defers** framework-specific route extraction to the roadmap.
- **Gherkin shape drift** — if the skeleton's Gherkin doesn't match what the
  parser expects, generated specs are invisible to the gate. Mitigated by an
  acceptance test that parses every generated file with the real parser.
