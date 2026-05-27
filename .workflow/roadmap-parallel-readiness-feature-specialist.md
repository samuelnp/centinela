### Feature-Specialist Report: roadmap-parallel-readiness
**Date:** 2026-05-26

#### Behavior Summary

This feature promotes inter-feature dependencies from the buried `roadmap-analysis.json` sidecar into a first-class `dependsOn []string` field on every `roadmap.json` Feature entry, then derives a never-persisted `FeatureReadiness` (done / in-progress / ready / blocked) for each feature using that graph plus live workflow status. Cycle and unknown-dependency validation now runs at `roadmap.Load()` time (Option B), so every consumer — render, start guard, rehydration, planadvisor, docgen — receives a validated graph and a cycle can never produce a misleading "ready". The parallel frontier (the set of `ready` features) is surfaced three ways: annotated 🔓/🔒 markers on `centinela roadmap` output, a `centinela roadmap ready` command that prints the frontier one name per line (exit 0 always, clear empty-state when none), and plural rehydration in SessionStart that lists every ready feature instead of just the first incomplete. `centinela start <f>` gains an unconditional dependency guard that refuses the start and names each unmet dependency; no `--force` override exists in v1.

#### Gherkin Scenarios

All scenarios live in `specs/roadmap-parallel-readiness.feature` (36 scenarios across 7 groups):

**Schema & Backward Compatibility (3 scenarios)**
- A roadmap.json with dependsOn fields loads successfully
- A roadmap.json without any dependsOn field loads exactly as before
- An empty dependsOn array is treated the same as an absent dependsOn field

**Validation — Negative Paths (4 scenarios)**
- A dependency on an unknown feature slug is rejected at load time
- A dependency cycle between two features is rejected at load time
- A self-dependency is rejected at load time as a cycle
- A longer cycle (A→B→C→A) is rejected at load time

**Readiness Derivation (8 scenarios)**
- A done feature is classified as done
- An in-progress feature is classified as in-progress
- A planned feature with all dependencies done is classified as ready
- A planned feature with an unmet dependency is classified as blocked
- A dependency that is in-progress keeps the dependent blocked
- Multiple unmet dependencies are all listed in BlockedBy
- Diamond dependency — D is ready only when both B and C are done
- Diamond dependency — D remains blocked when B is done but C is not

**centinela roadmap ready Command (3 scenarios)**
- roadmap ready prints each ready feature on its own line
- roadmap ready prints a clear empty-state line when no features are ready
- roadmap ready exits 0 when all features are done

**Start Guard (4 scenarios)**
- centinela start is refused when a dependency is not done
- centinela start is refused when a dependency is only in-progress
- centinela start proceeds when all dependencies are done
- centinela start proceeds when the feature has no dependencies

**Render Markers (3 scenarios)**
- centinela roadmap renders the ready marker on ready features
- centinela roadmap renders the blocked marker and dep names on blocked features
- centinela roadmap does not render ready or blocked markers on done features

**Plural Rehydration (3 scenarios)**
- SessionStart rehydration lists all ready features when multiple are ready
- SessionStart rehydration explains the block when frontier is empty but work remains
- SessionStart rehydration shows the roadmap-complete message when all features are done

#### UX States

| State | Trigger | Surface |
|-------|---------|---------|
| loading | n/a (CLI; no loading state) | n/a |
| empty — ready set empty, all done | All features have status "done" | `centinela roadmap ready` prints empty-state line; rehydration shows roadmap-complete message |
| empty — ready set empty, blocked work remains | All planned features have ≥1 unmet dep | `centinela roadmap ready` prints empty-state line; rehydration explains block (does NOT look complete) |
| error — invalid roadmap | roadmap.json references unknown dep or contains cycle | `centinela roadmap`, `centinela start`, `centinela roadmap ready` all surface the load error; rehydration is silent |
| error — start blocked | Feature has ≥1 dep not done when `centinela start <f>` runs | `centinela start` exits non-zero with error naming the unmet dependencies |
| success — ready list | ≥1 feature is ready | `centinela roadmap ready` prints one name per line; rehydration lists all ready features; `centinela roadmap` shows 🔓 per ready feature |
| success — annotated roadmap | Roadmap loaded with dep graph | `centinela roadmap` shows 🔓 ready, 🔒 blocked-by-<names> per feature |

#### Out-of-Scope

- Start-guard override flag (`--force` / `--ignore-deps`): unconditional guard is the v1 design; override is a fast follow if real friction appears.
- Readiness caching or persistence: readiness is always derived fresh from the loaded roadmap.
- Topological re-ordering of roadmap declaration: the graph informs readiness, not the order features are listed.
- Auto-migration that rewrites old `roadmap-analysis.json` files to strip `dependsOn`: Option B simply ignores analysis `dependsOn`.

#### Handoff

- Next role: senior-engineer
- Open clarifications:
  - `ValidateDependencies` is confirmed to run inside `Load()` so every consumer sees a validated graph; sentinel nil-roadmap case must remain a no-op.
  - The `hasCycle` DFS in `analysis_cycle.go` is reused unchanged; `dependencies.go` calls it with the roadmap-sourced dep map.
  - `internal/docgen/load.go` must be updated alongside `internal/planadvisor/roadmap_context.go` — both are confirmed read-path switches sourcing deps from roadmap.json.
  - Render markers (🔓/🔒) are Unicode literals; they must route through the existing icon constant pattern (`IconReady`, `IconBlocked`) rather than being hardcoded inline.
  - `render_readiness.go` (and `start_dep_guard.go`) should be created proactively if the parent files are close to the 100-line limit, rather than waiting for them to overflow.
