# Feature Brief — roadmap-edit-move

> Feature 3 of 4 in the Roadmap Editing Suite. Umbrella design:
> [docs/plans/roadmap-editing-suite-design.md](../plans/roadmap-editing-suite-design.md).
> This brief isolates the acceptance criteria + edge cases for THIS feature only
> (not roadmap-crud-add-remove / roadmap-phase-ops).

## Problem — what pain, who
`roadmap-crud-add-remove` (shipped) can create and delete features but cannot
**change** one in place, **move** it between phases, or **reorder** it. An
operator (and the Magallanes Plan page) needs to rename/retarget a feature,
fix its dependencies, relocate it to the right phase, and order features within
a phase — without hand-editing `roadmap.json`. This feature adds the mutate/
relocate half, reusing the raw-feature helpers feature 2 introduced.

## Scope (this feature ONLY)
- **In:** `roadmap edit`/`update` (rename + change description/dependsOn/archetype
  in place, rewriting dependents' `dependsOn` on rename); `roadmap move` (general
  phase→phase relocation with optional `--before`/`--after` anchor); `roadmap
  reorder` (reposition within/across a phase by anchor). Cycle/unknown-dep
  re-validation on every edit/move.
- **Out:** add/remove/draft/promote (roadmap-crud-add-remove, shipped); phase
  add/rename/remove (roadmap-phase-ops).

## User Stories
- As an operator, `roadmap edit <slug> --name <new>` renames the feature AND
  rewrites every other feature's `dependsOn` that referenced the old name, so the
  graph stays intact.
- As an operator, `roadmap edit <slug> --depends-on a,b` replaces its deps, and
  the command refuses if that introduces a cycle or an unknown dependency.
- As an operator, `roadmap move <slug> --to-phase <P> [--before <anchor>]`
  relocates the feature and preserves formatting of untouched phases.
- As an operator, `roadmap reorder <slug> --after <anchor>` repositions a feature
  among its siblings.

## Acceptance Criteria (THIS feature → Gherkin)
1. `roadmap edit <slug>` changes only the flags provided (`--name`,
   `--description`, `--depends-on`, `--archetype`); unspecified fields are left
   intact (a `--depends-on` sentinel distinguishes "clear deps" from "unchanged").
2. `edit --name <new>` renames the feature and rewrites `dependsOn` references to
   the old name across ALL phases; refuses if `<new>` is an invalid slug or
   collides with an existing feature (naming the owning phase).
3. `edit` refuses when the resulting `dependsOn` references an unknown feature or
   introduces a dependency cycle (`ValidateDependencies` re-run before write);
   `roadmap.json` left byte-identical on any rejection.
4. `roadmap move <slug> --to-phase <P>` relocates the feature to phase P (optionally
   at `--before`/`--after <anchor>`); untouched phases round-trip byte-identical;
   the feature's draft/finalized status and quality entries are preserved.
5. `move` refuses Backlog/Baseline as source or target (directing to `defer`/
   `promote`), an unknown target phase, and an unknown anchor.
6. `roadmap reorder <slug> --before|--after <anchor>` repositions the feature;
   a no-op reorder (already adjacent) leaves the file byte-identical.
7. All three commands are read-modify-write via the existing atomic raw layer; a
   rejected op writes nothing.

## Edge Cases (THIS feature ONLY)
- `edit`/`move`/`reorder` a slug that does not exist → clear "not found".
- `edit --name` to the SAME name → no-op rename (no dependents rewritten), file
  byte-identical.
- `edit --name` colliding with a feature in a DIFFERENT phase → refused, naming
  that phase.
- `edit --depends-on` set to empty (`--depends-on ""` with the sentinel) → clears
  deps; distinguished from omitting the flag (unchanged).
- Rename a feature that others depend on → every dependent's `dependsOn` updated;
  verify a downstream `roadmap ready`/`ValidateDependencies` still resolves.
- `move` a feature depended-on by another → allowed (dependency is by name, not
  phase); graph still valid.
- `move --before`/`--after` an anchor that is the feature itself → refused or no-op
  (define: treated as no-op, byte-identical).
- `reorder` across into a Backlog/Baseline phase → refused (same guard as move).
- Cycle introduced via `edit --depends-on` (including self-dependency) → refused.
- Empty/missing/malformed `roadmap.json` → error, file untouched.
- Concurrent writers → inherited atomic temp+rename, last-writer-wins (unchanged).

## Data Model
No persisted-schema change. Reuses `Feature`/`Phase`/`Roadmap` and the raw layer.
New raw helper `rawdeps.go` `rewriteDependents(old,new)` (rename); reuses feature
2's `findFeature`/`removeFeatureAt`/`replaceFeatureAt`/`insertFeatureAt`/`toRoadmap`.

## Integration Points
- Reuses `rawfeature_find.go`/`rawfeature_mutate.go`/`rawtyped.go` (feature 2),
  `ValidateDependencies`, and the atomic `writeRawRoadmap`. Consumer: Magallanes
  drives edit/move/reorder via the CLI; `roadmap --json` reflects the result.

## Risks
- **Rename dependent-rewrite miss** (High): a dependent whose `dependsOn` isn't
  rewritten leaves a dangling ref → caught by post-write `ValidateDependencies`,
  but must be tested across multiple dependents/phases.
- **Anchor resolution off-by-one** (Med): `--before`/`--after` insertion index;
  test first/last/middle and cross-phase.
- **Non-byte-preserving move** (Med): untouched phases must round-trip exactly;
  assert exact bytes.
- **Per-package coverage** (Low): colocated `internal/roadmap/*_test.go`; ≥97%.

## Decomposition
Coherent slice; depends on roadmap-crud-add-remove (shipped). No further split.
Successor: roadmap-phase-ops (phase add/rename/remove).
