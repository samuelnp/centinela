# Implementation Plan: roadmap-parallel-readiness

> Promote feature dependencies into `roadmap.json` and derive a parallel
> frontier (ready / blocked) so an operator running N Centinela instances can
> see at a glance which features are safe to start concurrently.
>
> Honors the brief's **Decision Record**: deps are first-class `dependsOn` on
> `roadmap.json`; **Option B** for validation (cycle + unknown-dep checks move
> onto roadmap.json, `roadmap-analysis.json` drops `dependsOn`, `ValidateAnalysis`
> keeps only role + coverage checks); surfaces = annotate render (🔓/🔒) + plural
> rehydration; capabilities = readiness view + `centinela roadmap ready` + a
> `centinela start` dependency guard. See
> `docs/features/roadmap-parallel-readiness.md`.

## Constraints

- n-tier (PROJECT.md G2/G7): readiness derivation + dep validation in
  `internal/roadmap`; `internal/ui` renders only; `cmd/centinela` is thin.
- Max 100 lines per source file (`_test.go` included). Split aggressively.
- Backward compatible: roadmaps without `dependsOn` behave exactly as today.
- Ship as ONE feature/branch (see big-thinker rollout rationale).

## Build order (smallest correct slice first)

### Step 1 — Schema: `Feature.DependsOn` (inert, backward compatible)
**Modify `internal/roadmap/roadmap.go`:**
- Add `DependsOn []string \`json:"dependsOn,omitempty"\`` to `Feature`.
- No behavior change yet. `omitempty` keeps existing roadmap.json byte-identical
  when no deps are set.
- File stays ≤100 lines (currently 82; +1 line).

### Step 2 — Option B validation move (cycle + unknown-dep onto roadmap.json)
**New file `internal/roadmap/dependencies.go`:**
- `func ValidateDependencies(r *Roadmap) error` — nil/empty-safe; builds
  `names := roadmapFeatureSet(r)` and `deps map[string][]string` from
  `r.Phases[].Features[].DependsOn`; returns:
  - `feature %s depends on unknown feature %s` for any dep not in `names`
    (self-dep on an unknown name also caught here),
  - `roadmap dependency cycle detected` when `hasCycle(deps)` is true
    (self-dep `A→A` is a 1-node cycle, naturally caught).
- Keep `analysis_cycle.go` `hasCycle` as the shared, dep-source-agnostic DFS
  (no change needed; reused here).
- File well under 100 lines.

**Modify `internal/roadmap/roadmap.go` `Load()`:**
- After unmarshal, call `ValidateDependencies(&r)`; return its error so every
  consumer (render, ready set, start, rehydration, planadvisor, docgen) sees a
  validated graph. Keeps a cycle from ever surfacing as a misleading "ready".
- Guard: `Load` already returns `(nil, err)` on read/parse failure; the new
  call only adds a dependency-validation error path. Absent/empty roadmap is
  unaffected (nil-safe).

**Modify `internal/roadmap/analysis.go` (Option B — drop deps):**
- Remove `DependsOn` field from `AnalysisFeature` (now just `Name`).
- In `ValidateAnalysis`: delete the `for _, dep := range f.DependsOn` unknown-dep
  loop, the `deps` map, and the `hasCycle(deps)` call. Keep: role must be
  `senior-product-manager`; every roadmap feature appears (coverage);
  analysis-references-unknown-feature check. File shrinks well under 100.

### Step 3 — Read-path switches (consumers source deps from roadmap.json)
**Modify `internal/planadvisor/roadmap_context.go` `dependencyNames`:**
- Replace the read of `roadmap.RoadmapAnalysisFile` + `roadmap.Analysis` with
  `roadmap.Load()`; walk phases, find `feature`, return a copy of its
  `DependsOn`. Same `[]string` return shape; callers unchanged.
- Drop the now-unused `encoding/json` / `os` imports if they fall away.

**Modify `internal/docgen/load.go` `loadRoadmapNodes`:**
- Source the graph from `roadmap.json`: decode `phases[].features[]` into a
  struct that includes `DependsOn []string \`json:"dependsOn"\`` and build
  `RoadmapNode{Name, DependsOn}`. Drop the now-empty analysis-deps branch (the
  analysis no longer carries `dependsOn`). This preserves the dependency-edge
  rendering in the generated docs graph.
- `internal/docgen/types.go` `RoadmapNode` is unchanged (already has
  `DependsOn`).

### Step 4 — Readiness derivation (the core domain addition)
**New file `internal/roadmap/readiness.go`:**
- Type `FeatureReadiness struct { Name string; State string; BlockedBy []string }`
  (never persisted). States: `"done"`, `"in-progress"`, `"ready"`, `"blocked"`.
- `func DeriveReadiness(r *Roadmap) []FeatureReadiness` — single pass:
  1. Build `status := map[string]string{}` via `FeatureStatus(name)` for every
     feature (one lookup each).
  2. For each feature in declared order, classify per the brief's table:
     - `done`/`in-progress` straight from status;
     - else (planned): collect deps whose `status != "done"` into `BlockedBy`;
       empty ⇒ `ready`, non-empty ⇒ `blocked` (BlockedBy = unmet dep names).
- `func ReadySet(r *Roadmap) []string` — names where `State == "ready"`, declared
  order. Used by the command and rehydration.
- Keep this file ≤100 lines; if classification + helpers crowd it, move the
  per-feature classify into `readiness_classify.go`.

### Step 5 — Surfaces: render markers + plural rehydration
**Modify `internal/ui/render_roadmap.go` `RenderRoadmap`:**
- Switch from per-feature `FeatureStatus` to `roadmap.DeriveReadiness(r)` so each
  line knows ready/blocked. For `ready` append 🔓 marker; for `blocked` append 🔒
  plus `blocked-by: <names>` (from `BlockedBy`); done/in-progress keep their
  existing icons. Keep rendering pure (no logic — derivation is in roadmap).
- If this pushes `render_roadmap.go` over 100 lines, extract the line/marker
  formatting into a new `internal/ui/render_readiness.go` helper (icons +
  blocked-by formatting), called by `RenderRoadmap`.
- Add marker constants (e.g. `IconReady = "🔓"`, `IconBlocked = "🔒"`) next to the
  existing `IconDone/IconActive/IconPending`.

**Modify `internal/ui/render_session.go` `RenderSessionRehydration`:**
- Change signature to take the ready set, e.g.
  `RenderSessionRehydration(r *roadmap.Roadmap, ready []string, hasIncomplete bool)`.
- Body: keep banner + `RenderRoadmap(r)` (now annotated) + pointers; replace the
  single "Next feature to plan" line with a plural "Ready to start now" block
  listing every ready feature (one per line) with a pointer per feature.
- Empty-frontier states (AC #7):
  - `len(ready)==0 && !hasIncomplete` ⇒ keep existing green "Roadmap complete".
  - `len(ready)==0 && hasIncomplete` ⇒ explain: everything not done is
    in-progress or blocked (so the session does not look complete).
- Keep ≤100 lines; if the plural block + empty-state branches overflow, extract a
  `renderReadyBlock(ready, hasIncomplete)` helper into a small new file.

**Modify `cmd/centinela/hook_session.go` `runHookSession`:**
- Replace `next, hasNext := roadmap.FirstIncomplete(r)` with
  `ready := roadmap.ReadySet(r)` and compute `hasIncomplete` from
  `r.Summary()` (planned+inProgress > 0) — thin wiring only, no logic.
- Pass `(r, ready, hasIncomplete)` to `RenderSessionRehydration`. Absent/invalid
  roadmap still exits silently (unchanged).

### Step 6 — Capabilities: `roadmap ready` command + start guard
**New file `cmd/centinela/roadmap_ready.go`:**
- Subcommand `ready` under `roadmapCmd`: `centinela roadmap ready`.
- Thin: `roadmap.Load()` → on error `roadmapCommandError(err)`;
  `ready := roadmap.ReadySet(r)`; print one name per line via a ui renderer; when
  empty print a clear empty-state line (e.g. via `ui.RenderReadyEmpty()` or
  `StyleMuted`). Exit 0 always (success). No business logic in cmd.
- Add a tiny `internal/ui` renderer for the ready list + empty-state line if a
  styled output is wanted (keeps formatting out of cmd).

**Modify `cmd/centinela/start_guard.go` `workflowOrderForFeature`:**
- After roadmap is loaded and `ValidateAnalysis`/`ValidateQuality` pass (and
  before returning the step order), add a dependency guard:
  derive readiness / look up `feature`'s `DependsOn`; if any dep's
  `FeatureStatus != "done"`, return
  `fmt.Errorf("cannot start %q — blocked by unmet dependencies: %s", feature, names)`.
  Proceed when all deps done or none.
- Put the actual "which deps are unmet" computation in `internal/roadmap`
  (e.g. `func UnmetDependencies(r *Roadmap, feature string) []string`) so the
  guard in cmd stays thin (just calls it + formats the error). This keeps logic
  out of the outer layer (G7).
- The guard only applies on the greenfield non-bootstrap path already in this
  function; existing-project flow (returns early) is untouched. No override flag
  in v1 (decision: out — see big-thinker scope).
- If `start_guard.go` approaches 100 lines, extract the guard into
  `start_dep_guard.go`.

### Step 7 — Docs-step guidance for ROADMAP.md `dependsOn`
- ROADMAP.md has no code generator; update the scaffold instruction text in
  `internal/ui/render_roadmap.go` `RenderRoadmapNeeded` so the roadmap.json
  example shows the optional `dependsOn` field, and note that ROADMAP.md should
  list dependencies per feature. The narrative docs update (showing dependsOn in
  ROADMAP.md and the readiness model) is finished in the docs step.

## New / changed symbols (summary)

- `roadmap.Feature.DependsOn []string` (new field) — `roadmap.go`.
- `roadmap.ValidateDependencies(r) error` — new, `dependencies.go`.
- `roadmap.FeatureReadiness{ Name, State, BlockedBy }` + `DeriveReadiness(r)` +
  `ReadySet(r)` — new, `readiness.go` (+ optional `readiness_classify.go`).
- `roadmap.UnmetDependencies(r, feature) []string` — new, for the start guard.
- `roadmap.AnalysisFeature` loses `DependsOn`; `ValidateAnalysis` loses dep/cycle
  checks (Option B) — `analysis.go`.
- `roadmap.Load()` now also validates dependencies — `roadmap.go`.
- `ui.RenderRoadmap` annotated; `ui.RenderSessionRehydration` plural signature;
  `IconReady`/`IconBlocked`; optional `render_readiness.go` — `internal/ui`.
- `roadmapReadyCmd` — new, `cmd/centinela/roadmap_ready.go`.
- Start dependency guard in `workflowOrderForFeature` — `cmd/centinela/start_guard.go`.
- `runHookSession` passes ready set — `cmd/centinela/hook_session.go`.
- `planadvisor.dependencyNames` reads roadmap.json — `roadmap_context.go`.
- `docgen.loadRoadmapNodes` sources deps from roadmap.json — `load.go`.

## Files created vs. modified

**Created:**
- `internal/roadmap/dependencies.go`
- `internal/roadmap/readiness.go` (+ maybe `readiness_classify.go`)
- `cmd/centinela/roadmap_ready.go`
- `internal/ui/render_readiness.go` (only if `render_roadmap.go` would exceed 100)
- `cmd/centinela/start_dep_guard.go` (only if `start_guard.go` would exceed 100)

**Modified:**
- `internal/roadmap/roadmap.go` (schema + Load validation)
- `internal/roadmap/analysis.go` (Option B: drop deps + cycle)
- `internal/ui/render_roadmap.go` (markers + scaffold example)
- `internal/ui/render_session.go` (plural rehydration)
- `cmd/centinela/hook_session.go` (pass ready set)
- `cmd/centinela/start_guard.go` (dependency guard)
- `internal/planadvisor/roadmap_context.go` (read deps from roadmap.json)
- `internal/docgen/load.go` (graph edges from roadmap.json)

`internal/roadmap/analysis_cycle.go` (`hasCycle`) is reused unchanged.

## Backward compatibility & edge cases (drive the tests step)

- Omitted/empty `dependsOn` ⇒ feature `ready` once planned (no-deps roadmap test).
- Self-dep `A→A` ⇒ cycle rejected at `Load`. Unknown dep slug ⇒ rejected by name.
- Diamond A→B, A→C, B→D, C→D ⇒ D ready only when B and C done.
- Dep `in-progress` (not done) ⇒ dependent stays `blocked`.
- Cross-phase dep allowed structurally but flagged by readiness (graph wins).
- All done ⇒ ready set empty ⇒ rehydration shows existing complete message.
- None ready, none in-progress, some blocked ⇒ rehydration explains the block
  (must NOT look complete) — AC #7.
- roadmap.json missing/invalid ⇒ existing behavior preserved (rehydration silent;
  start errors guarded).

## Out of scope (v1)

- Start-guard override flag (`--force`/`--ignore-deps`) — out (precise error is
  the relief valve; revisit only if real friction appears).
- Auto-migration stripping `dependsOn` from old `roadmap-analysis.json` — Option
  B simply ignores it.
- Readiness caching / persistence — always derived fresh.
- Topological re-ordering of roadmap declaration.
