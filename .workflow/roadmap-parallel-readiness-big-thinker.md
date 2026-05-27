### Big-Thinker Report: roadmap-parallel-readiness
**Date:** 2026-05-26

#### Problem

A developer driving Centinela wants to fan out N Claude instances — one per
worktree (built on `parallel-feature-worktrees`) — and advance as many
*independent* features at once as the dependency graph allows. Today nothing
turns the dependency data into an actionable signal:

- Centinela records inter-feature dependencies only in
  `.workflow/roadmap-analysis.json` (`AnalysisFeature.DependsOn`), produced as a
  side effect of the senior-PM analysis (`roadmap-senior-pm-analysis`). They are
  buried in an analysis artifact, not in the roadmap the operator edits.
- `centinela roadmap` and the SessionStart rehydration render only
  planned / in-progress / done per feature — never "ready to start now" vs.
  "blocked by X".
- `FirstIncomplete` returns a single next feature in declared phase order, so
  the operator cannot see the *set* (frontier) of features workable in parallel.
- `centinela start <f>` (via `start_guard.go` → `workflowOrderForFeature`) never
  checks dependencies, so a parallel instance can silently grab a feature whose
  prerequisites are not `done`.

Who hurts: the operator orchestrating parallel work. What they do today: trace
the graph by hand from the analysis JSON, or just guess. Why now: worktrees made
parallel execution real, so the missing piece is the readiness signal.

#### Scope

- **In:**
  - `roadmap.Feature.DependsOn []string` as the first-class, authoritative
    dependency field in `roadmap.json` (omitempty; absent/`[]` ⇒ no deps, fully
    backward compatible).
  - **Option B** validation move: unknown-dependency + cycle validation runs
    against `roadmap.json` at load/validate time; `ValidateAnalysis` keeps only
    its non-dependency duties (role must be `senior-product-manager`; analysis
    must cover every roadmap feature). `AnalysisFeature.DependsOn` is dropped.
  - A derived, never-persisted `FeatureReadiness{ Name, State, BlockedBy }` type
    + derivation classifying every feature as exactly one of
    done / in-progress / ready / blocked (semantics table in the brief).
  - Surfaces: 🔓 ready / 🔒 blocked-by-X markers on the existing `centinela
    roadmap` render and the rehydration phase overview; plural rehydration that
    lists *all* ready features.
  - `centinela roadmap ready` command — prints the ready set (one per line),
    clear empty-state line when none, exit 0.
  - `centinela start <f>` dependency guard — refuses when a dependency is not
    `done`, naming the unmet deps; proceeds when all done or none.
  - `internal/planadvisor` read-path switch: `dependencyNames` reads deps from
    `roadmap.json` instead of `roadmap-analysis.json`.
  - `internal/docgen` read-path switch: `loadRoadmapNodes` sources the
    dependency graph from `roadmap.json` (analysis no longer carries deps).
  - `ROADMAP.md` doc guidance updated to show `dependsOn` per feature
    (handled in the docs step; the render scaffold instructions in
    `render_roadmap.go` get the new schema example).

- **Out (v1 — recommended):**
  - **Start-guard override flag** (e.g. `--force` / `--ignore-deps`): OUT for
    v1. Rationale: keeping the guard unconditional matches the brief's lean
    ("lean: out of scope for v1"), avoids a footgun that defeats the guard's
    purpose for parallel safety, and keeps `start.go`/`start_guard.go` within the
    100-line budget. The precise error message (naming unmet deps) is the
    pressure-relief valve; an override can be a fast follow if real friction
    appears. RECOMMENDATION: out.
  - Auto-migration that rewrites old `roadmap-analysis.json` files to strip
    `dependsOn` — unnecessary; Option B simply ignores analysis deps.
  - Any change to phase ordering or to how features are *declared*; topological
    re-ordering of the roadmap is out (graph informs readiness, not declaration).
  - A persisted readiness cache — readiness is always derived fresh.

#### Dependencies & Assumptions

- Builds on `parallel-feature-worktrees` (per-feature worktrees + branches make
  the frontier meaningful) and on `roadmap-senior-pm-analysis` (which today owns
  the `dependsOn` data we are promoting). Touches `roadmap-checkpoint-prompt`
  (bootstrap/first-feature) and `session-context-rehydration` (rehydration
  render) only at their seams.
- n-tier layering (PROJECT.md G2/G7): readiness derivation + dep validation live
  in `internal/roadmap`; `internal/ui` only renders (may read roadmap types);
  `cmd/centinela` is a thin orchestrator with no logic. `internal/docgen` and
  `internal/planadvisor` are read-only consumers below cmd.
- Assumption: ROADMAP.md is human/LLM-authored (no code generator writes it), so
  "show dependsOn in ROADMAP.md" is doc-step + scaffold-instruction guidance, not
  a code-gen change. Verified: no programmatic ROADMAP.md writer exists.
- Assumption: feature status comes only from `workflow.Load` via
  `FeatureStatus` (planned/in-progress/done); readiness composes that with deps.
- Assumption: existing roadmaps in the wild have no `dependsOn` and must behave
  exactly as today (every planned feature ⇒ ready once planned).

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|------------|
| Validation move regresses cycle/unknown-dep detection (it currently runs on analysis) | High | Med | Keep `hasCycle` (it is dep-source-agnostic), add `ValidateDependencies(r)` on roadmap; reuse the same DFS; keep + extend existing cycle/unknown tests against roadmap.json. Drop the analysis-side dep checks only after the roadmap-side ones pass. |
| docgen graph silently loses dependency edges (it reads deps from analysis JSON today) | Med | High | Flip `loadRoadmapNodes` to read deps from `roadmap.json`; the analysis branch no longer supplies `dependsOn`. Add/adjust docgen load test. Flagged explicitly in plan step 8. |
| Start guard over-blocks intentional out-of-order starts | Med | Med | Precise error naming unmet deps; no override flag in v1 (documented decision); guard only fires for greenfield non-bootstrap path, leaving existing-project flow untouched. |
| Rehydration "looks complete" when frontier empty but blocked work remains | Med | Med | Distinct empty-frontier states: all-done ⇒ existing complete message; none-ready-but-blocked/in-progress ⇒ explain the block (edge case AC #7). Cover with table tests. |
| File-size (G1) creep on touched files | Low | High | Split: readiness derivation (`readiness.go`), the ready-set helper, render markers in a small helper, `roadmap_ready.go` as a thin command, guard in `start_guard.go`. Keep each ≤100 lines; `_test.go` also ≤100 (project memory). |
| Backward incompat for roadmaps without `dependsOn` | Low | Low | `omitempty` + nil-deps ⇒ ready-when-planned; explicit backward-compat table test with a no-deps roadmap. |
| planadvisor read-path switch changes advisor context unexpectedly | Low | Med | `dependencyNames` returns the same shape (`[]string`), just sourced from roadmap.json; keep its existing test, add a roadmap-sourced case. |
| Self-dependency / dep on unknown slug slips through | Low | Low | Explicit validation: self-dep treated as cycle; unknown slug rejected by name. Edge-case tests from the brief. |

#### Rollout

Ship as **ONE feature** (single branch). Recommendation: do NOT split into the
optional `roadmap-dependsOn-schema` + `roadmap-readiness-surface` pair. Reasons:
the schema change and Option B validation move are inert until a surface reads
them, and shipping the schema alone leaves the analysis still carrying deps that
the surface step must then untangle — more migration churn than value. The whole
thing is a cohesive, modestly sized change (one new derivation file, one new
command file, small edits to ~7 existing files) that fits one branch under the
file-size budget. If the branch grows past comfortable review size, the brief's
split (schema/validation/planadvisor first, then surfaces) is the fallback seam.

Smallest-correct-slice ordering within the one branch (detailed in the plan):
1. Schema: `Feature.DependsOn` + `omitempty` (inert, backward compatible).
2. Validation move (Option B): `ValidateDependencies(r)` on roadmap reusing
   `hasCycle`; wire into Load/roadmap-validate/start path; strip
   `AnalysisFeature.DependsOn` and the dep checks from `ValidateAnalysis`.
3. Read-path switches: planadvisor + docgen now read deps from roadmap.json.
4. Readiness derivation: `FeatureReadiness` + `DeriveReadiness(r)` + `ReadySet`.
5. Surfaces: render markers (roadmap + rehydration), plural rehydration list.
6. Commands/guard: `roadmap ready` command, `centinela start` dependency guard.
7. Docs-step guidance for ROADMAP.md `dependsOn` (and scaffold instruction text).

Can wait: override flag (out), readiness caching (out), any topo re-ordering.

#### Handoff
- Next role: feature-specialist
- Outstanding questions:
  - Confirm the start guard stays unconditional (no `--force`) for v1 — this
    report recommends out; feature-specialist should encode the decision in the
    `.feature` spec so QA tests the block path, not an override.
  - Confirm `ValidateDependencies` runs at `Load()` time vs. only at explicit
    validate points. RECOMMENDATION: run it inside `Load()` so every consumer
    (render, ready, start, rehydration) gets a validated graph and a cycle never
    produces a misleading "ready"; keep it cheap and nil-safe so an absent/empty
    roadmap path is unaffected.
  - Confirm docgen graph edges should come from roadmap.json (yes — analysis no
    longer carries deps); feature-specialist to note the docgen load test update.
