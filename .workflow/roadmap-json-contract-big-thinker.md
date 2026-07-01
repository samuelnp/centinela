### Big-Thinker Report: roadmap-json-contract
**Date:** 2026-07-01

#### Problem

The `roadmap` command family prints human-oriented text only (the sole existing
`--json` is `brownfield`). Magallanes — the multi-tenant SaaS control plane —
needs to render a "Plan project" page from a project's roadmap: phases, features,
per-feature derived status and readiness. Today its only options are to scrape
`ROADMAP.md` prose or read `.workflow/roadmap.json` raw, and the raw file lacks
the *derived* status/readiness that Centinela computes from workflow state at
runtime. Magallanes already shells out to `centinela validate`/`complete`; it
needs the same shell-out path to obtain a stable, machine-readable roadmap view.
This feature delivers that read-only contract and unblocks Magallanes ahead of
the mutation features in the suite.

#### Scope

- **In:**
  - `--json` on `roadmap` → a deterministic `RoadmapView` (ordered `phases[]`,
    each `PhaseView` with ordered `features[]`; each `FeatureView` carries
    `name, phase, status (planned|in-progress|done), readiness (ready|blocked),
    dependsOn[], blockedBy[]`) plus a top-level `counts` (`StatusCounts`).
  - `--json` on `roadmap ready` → JSON array of ready feature names in declared
    order (`ReadySet(r)`), identical set to the text view.
  - New `roadmap show` (alias `list`) with `--json` → the **persisted** typed
    `Roadmap` verbatim (storage contract); text mode reuses `ui.RenderRoadmap`.
  - New derived view types in `internal/roadmap/` (`view_types.go`) and the
    `BuildView` builder (`view.go`), reusing `Load`, `FeatureStatus`,
    `DeriveReadiness`/`classifyFeature`/`ReadySet`, and `Summary()` scoping.
  - Optional `json:"…"` tags on `FeatureReadiness` fields so readiness data can
    serialize — behavior unchanged.
- **Out (do NOT build here):**
  - Any mutation (`add`/`remove`/`edit`/`move`/`reorder`/`phase ops`).
  - The `draft` status/readiness dimension and the `Feature.Draft` field — those
    arrive in the successor `roadmap-crud-add-remove`.
  - MCP roadmap tools (MCP stays governance-read-only).
  - Any Magallanes-side code (separate `../magallanes` repo).
  - Any change to the persisted `roadmap.json` schema.

#### Dependencies & Assumptions

- Builds entirely on existing `internal/roadmap` read APIs: `Load()`
  (`roadmap.go`), `FeatureStatus`, `Summary()`, and the readiness machinery
  (`DeriveReadiness`/`classifyFeature`/`ReadySet`/`collectUnmet` in
  `readiness.go`). No new derivation logic — the view is a projection.
- Mirrors the established `--json` flag precedent in `cmd/centinela/dashboard.go`
  and `verdict.go`: a `BoolVar(&…,"json",…)` plus a one-line `json.MarshalIndent`
  in the command; all view construction stays in `internal/roadmap` so `cmd/`
  holds no view logic (N-Tier layering: outer layer stays thin).
- `counts` scope matches `Summary()` exactly — Backlog/Baseline phases are
  non-schedulable and excluded from counts and from `ready`, per
  `isNonSchedulablePhase`. `show --json` includes them raw (verbatim persisted).
- Consumers (Magallanes) pin to the emitted JSON field names; the shape is a
  stable contract with additive-only evolution downstream.
- `status` and `readiness` are two distinct dimensions in the view, but the
  existing `FeatureReadiness.State` conflates them into one value
  (`done|in-progress|ready|blocked`). `BuildView` must set `status =
  FeatureStatus(name)` and derive `readiness` separately from the classified
  state (see Handoff — one design decision to nail down).

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Determinism / byte-stability regression (any map iteration in the view builder) | High | Medium | Iterate ordered slices only (`r.Phases` → `phase.Features`); no map ranging. Add an exact-bytes marshal test comparing two independent builds. |
| Downstream contract drift (Magallanes pins the shape) | Medium | Medium | Treat JSON field names as a frozen contract; document them in the plan/spec; only additive changes in successor features. |
| Text-output regression (altering existing human output) | Medium | Low | `--json` gated strictly behind the flag; default path calls the unchanged `ui.Render*`. Existing text tests must stay green byte-for-byte. |
| Per-package coverage gate (95%, aim ≥97%) not moved by `tests/`-tier files | Low | Medium | Colocate `internal/roadmap/view_test.go` and `cmd/centinela/roadmap_json_test.go`; no `-coverpkg`, so coverage must be local. |
| Status/readiness conflation mis-mapped in the view | Medium | Low | Explicit `BuildView` mapping table + unit tests asserting every state combination (done/in-progress carry status; only planned splits ready vs blocked). |

#### Rollout

- **Step 1 — types + builder (pure, internal):** add
  `internal/roadmap/view_types.go` (`FeatureView`/`PhaseView`/`StatusCounts`/
  `RoadmapView`, JSON-tagged, ordered slices) and `internal/roadmap/view.go`
  (`BuildView(r *Roadmap) RoadmapView`) reusing `FeatureStatus` +
  `classifyFeature`/`DeriveReadiness` and `Summary()`-scoped counts. Fully
  unit-testable with an in-memory `Roadmap` (no CLI, no I/O). Add `json` tags to
  `FeatureReadiness` if needed for serialization. This is the smallest correct
  slice and de-risks determinism first.
- **Step 2 — wire the commands (thin):** add `--json` to
  `cmd/centinela/roadmap.go` (`json.MarshalIndent(BuildView(r))`) and
  `roadmap_ready.go` (`json.MarshalIndent(ReadySet(r))`); add
  `cmd/centinela/roadmap_show.go` (`show`/alias `list`, text →
  `ui.RenderRoadmap`, `--json` → `json.MarshalIndent(r)`), registered via
  `init()` + `roadmapCmd.AddCommand`. Preserve exit codes and stderr error paths
  for missing/malformed `roadmap.json` in both text and `--json` modes.
- **Step 3 — verify end-to-end:** `go test ./...` green; coverage ≥97%; build a
  local binary and diff repeated `--json` runs for byte-stability;
  `centinela validate` passes in the worktree.

#### Deferred Findings

- none. Scope is tight and fully covered by the umbrella design; the `draft`
  dimension and mutations are already-agreed exclusions owned by successor
  features, not new gaps.

#### Handoff

- **Next role:** feature-specialist
- **Outstanding questions:**
  1. `readiness` field value for non-planned features: the brief lists
     `readiness (ready|blocked)`, but `done`/`in-progress` features carry their
     signal in `status`. Feature-specialist must decide the `readiness` value
     for those rows (leave empty string, use `omitempty`, or a sentinel) and
     encode it in the Gherkin spec so the contract is unambiguous for Magallanes.
     The plan's `FeatureView.Readiness` comment ("done/in-progress carry status")
     points to leaving `readiness` empty for those rows — confirm and lock it.
  2. `blockedBy` serialization: plan uses `omitempty`; confirm that an empty
     `blockedBy` is omitted rather than emitted as `[]`/`null`, and that this is
     acceptable to the consumer contract (asymmetric with `dependsOn`).
  3. Confirm the intended asymmetry between `roadmap --json` (`RoadmapView`,
     scoped to schedulable phases via `DeriveReadiness`, so Backlog/Baseline
     features do NOT appear) and `roadmap show --json` (persisted `Roadmap`
     verbatim, which DOES include Backlog/Baseline). Both are in-scope; the spec
     should state this explicitly.
