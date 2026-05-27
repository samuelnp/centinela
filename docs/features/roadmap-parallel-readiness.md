# Feature: roadmap-parallel-readiness

> Promote feature dependencies into the roadmap and derive a "parallel frontier"
> so an operator running multiple Centinela/Claude instances can see — at a
> glance — which features are safe to start concurrently.

## Problem

Centinela already records inter-feature dependencies (`dependsOn` in
`.workflow/roadmap-analysis.json`) and already provisions one git worktree +
branch per feature on `centinela start`. But nothing turns that dependency
graph into an actionable signal:

- `centinela roadmap` and the SessionStart rehydration show per-feature status
  (planned/in-progress/done) only — never "ready to start now" vs. "blocked by X".
- `FirstIncomplete` returns a single next feature in declared order, so the
  operator cannot see the *set* of features that could be worked in parallel.
- `centinela start <f>` does not check dependencies, so a parallel instance can
  silently grab a feature whose prerequisites aren't done yet.

**Who is the user:** a developer driving Centinela who wants to fan out N Claude
instances (one per worktree) and advance as many independent features as
possible at once, without manually reasoning about the dependency graph.

## Decision Record (settled with the operator before planning)

1. **Dependencies are promoted to `roadmap.json`** as a first-class per-feature
   `dependsOn []string`, and shown in human-readable `ROADMAP.md`. `roadmap.json`
   becomes the single source of truth for dependencies.
2. **Option B for validation:** cycle + unknown-dependency validation moves onto
   `roadmap.json`. `roadmap-analysis.json` no longer carries `dependsOn`;
   `ValidateAnalysis` keeps only its non-dependency duties (role must be
   `senior-product-manager`, analysis must cover every roadmap feature). The
   senior-PM analysis becomes pure qualitative rationale.
3. **Surfaces:** annotate the existing roadmap render (🔓 ready / 🔒 blocked-by-X
   markers) AND make SessionStart rehydration list *all* ready features (plural).
4. **Capabilities:** readiness view + a `centinela roadmap ready` parallel-batch
   command + a `centinela start` dependency guard.

## Readiness Semantics

For each feature, given its `dependsOn` and its derived `FeatureStatus`:

| Derived state | Condition |
|---------------|-----------|
| `done` | `FeatureStatus == done` |
| `in-progress` | `FeatureStatus == in-progress` (a worktree already owns it) |
| `ready` | `FeatureStatus == planned` AND every dependency is `done` |
| `blocked` | `FeatureStatus == planned` AND ≥1 dependency is not `done` |

- The **parallel frontier** = the set of `ready` features. Two or more `ready`
  features ⇒ they can be started in separate worktrees concurrently.
- `blocked` features must report *which* unmet dependencies block them.

## User Stories

- As an operator, I want each roadmap feature annotated `ready` / `blocked-by:X`
  so I can see the parallel frontier without tracing the graph by hand.
- As an operator, I want `centinela roadmap ready` to print just the features
  safe to start right now, so I can spin up that many instances.
- As an operator, I want `centinela start <f>` to refuse a feature whose
  dependencies aren't `done`, so a parallel instance can't grab blocked work.
- As an operator, I want SessionStart rehydration to list every ready feature
  (not only the first incomplete), so a fresh session immediately shows me the
  frontier.
- As a roadmap author, I want to declare `dependsOn` directly in `roadmap.json`
  and see it in `ROADMAP.md`, so dependencies are first-class and editable.

## Acceptance Criteria

1. `roadmap.json` `Feature` accepts an optional `dependsOn: []string`; absent or
   `[]` means no dependencies (backward compatible with existing roadmaps).
2. Loading a `roadmap.json` whose `dependsOn` references an unknown feature, or
   forms a cycle, fails validation with a clear, specific error.
3. Readiness derivation classifies every feature as exactly one of
   done/in-progress/ready/blocked per the table above; `blocked` includes the
   list of unmet dependency names.
4. `centinela roadmap` (and the rehydration phase overview) renders 🔓 on ready
   features and 🔒 with the blocking dependency names on blocked features.
5. `centinela roadmap ready` prints the ready set (one per line); when none are
   ready it prints a clear empty-state line; exit code 0.
6. `centinela start <f>` blocks with a specific error naming the unmet
   dependencies when `f` has a dependency that is not `done`; it proceeds
   normally when all dependencies are `done` (or `f` has none).
7. SessionStart rehydration lists all ready features; when the frontier is empty
   but work remains, it explains why (everything ready is done/in-progress or
   blocked); when the roadmap is complete it keeps the existing message.

## Edge Cases

- Empty `dependsOn` / field omitted ⇒ feature is ready as soon as it is planned.
- Self-dependency (`A dependsOn A`) ⇒ rejected as a cycle.
- Dependency on an unknown feature slug ⇒ rejected with the offending names.
- Dependency on a feature in a *later* phase ⇒ allowed structurally but flagged;
  declared phase order and dependency order may disagree (graph wins).
- A dependency that is `in-progress` (not `done`) ⇒ dependent stays `blocked`.
- Diamond dependencies (A→B, A→C, B→D, C→D) ⇒ D ready only when B and C done.
- All features done ⇒ ready set empty; rehydration shows roadmap-complete.
- No features ready and none in progress but some blocked ⇒ rehydration must not
  look "complete"; it must explain the block.
- `roadmap.json` missing or invalid ⇒ existing behavior (silent on rehydration,
  guarded errors on start) is preserved.

## Data Model

- `roadmap.Feature` gains `DependsOn []string \`json:"dependsOn,omitempty"\``.
- New derived type (e.g. `roadmap.FeatureReadiness{ Name, State, BlockedBy }`)
  computed from the graph + `FeatureStatus`; never persisted.
- `roadmap-analysis.json` `AnalysisFeature` drops `DependsOn` (option B).

## Integration Points

- `internal/roadmap/`: schema (`roadmap.go`), readiness derivation (new file),
  cycle/dep validation (move/extend `analysis_cycle.go` usage onto roadmap),
  `ValidateAnalysis` (drop dependency checks), `FirstIncomplete`/new "ready set".
- `internal/ui/render_roadmap.go` + `render_session.go`: ready/blocked markers,
  plural rehydration list.
- `cmd/centinela/roadmap.go` (+ new `roadmap_ready.go`): the batch command.
- `cmd/centinela/start.go` / `start_guard.go`: dependency guard before provision.
- `cmd/centinela/hook_session.go`: pass the ready set to rehydration render.
- `internal/planadvisor/roadmap_context.go`: `dependencyNames` must read deps
  from `roadmap.json` instead of `roadmap-analysis.json`.
- `ROADMAP.md` generation/doc: show `dependsOn` per feature.

## Risks

- **Two sources of truth during migration** (Medium): existing
  `roadmap-analysis.json` files still carry `dependsOn`. Mitigation: option B
  makes roadmap.json authoritative and ignores analysis `dependsOn`; provide a
  clear validation error and document the move.
- **Start guard over-blocking** (Medium): blocking may frustrate intentional
  out-of-order starts. Mitigation: precise error message; the plan should decide
  whether an explicit override flag is in scope (lean: out of scope for v1).
- **File-size (G1) pressure** (Low): readiness + rendering must stay ≤100 lines
  per file; split derivation, formatting, and command wiring.
- **Backward compatibility** (Low): roadmaps without `dependsOn` must behave
  exactly as today.

## Decomposition

Single cohesive feature. If the plan judges it too large for one branch, split:
- `roadmap-dependsOn-schema` — schema + validation move (option B) + planadvisor
  read-path switch.
- `roadmap-readiness-surface` — readiness derivation, render markers, plural
  rehydration, `roadmap ready` command, start guard.
</content>
</invoke>
