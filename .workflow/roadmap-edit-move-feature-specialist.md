### Feature-Specialist Report: roadmap-edit-move

**Date:** 2026-07-01

#### Behavior Summary

`roadmap edit`/`update <slug>` applies only the flags provided
(`--name`/`--description`/`--depends-on`/`--archetype`) to an existing feature,
leaving unspecified fields untouched; a `--depends-on` Changed sentinel
distinguishes "clear all deps" (`--depends-on ""`) from "leave deps alone"
(flag omitted). Renaming (`--name`) rewrites every other feature's `dependsOn`
that referenced the old name across ALL phases, refuses an invalid slug or a
name collision (naming the owning phase), and is a byte-identical no-op when
the new name equals the current one. Every edit re-runs `ValidateDependencies`
before writing, refusing unknown-feature deps or any introduced cycle
(including self-dependency) with `roadmap.json` left byte-identical.
`roadmap move <slug> --to-phase <P> [--before|--after <anchor>]` relocates a
feature to a different (non-Backlog/Baseline) phase, optionally anchored among
the target phase's existing features; untouched phases round-trip
byte-identical, and the feature's `draft`/quality state (keyed by name) is
preserved across the move since dependencies are keyed by name, not phase.
`roadmap reorder <slug> --before|--after <anchor>` repositions a feature within
or across a phase by anchor, and a no-op reorder (already in the requested
position) leaves the file byte-identical. All three commands are single
read-modify-write operations through the existing atomic raw layer reused from
`roadmap-crud-add-remove`; a rejected op never reaches the write and leaves
`roadmap.json` completely untouched.

#### Gherkin Scenarios

All scenarios live in `specs/roadmap-edit-move.feature` (34 scenarios/outlines,
several with Examples tables expanding to more concrete cases):

- edit changes only the flags provided, leaving unspecified fields intact
- edit --depends-on distinguishes "clear deps" from "unchanged" via the Changed sentinel
- edit --depends-on "" (sentinel present) clears dependencies
- edit --depends-on replaces the dependency list with the provided set
- edit --name renames the feature and rewrites dependents' dependsOn across ALL phases
- edit --name refuses an invalid slug
- edit --name refuses a collision with an existing feature, naming the owning phase
- edit --name to the SAME name is a no-op — no dependents rewritten, file byte-identical
- edit refuses a dependsOn that is unknown or introduces a cycle (Outline: unknown-dep, self-dep)
- edit --depends-on introduces a multi-hop cycle across two features is refused
- edit/update a slug that does not exist errors "not found"
- update is an alias for edit
- move relocates a feature to the target phase, appending by default
- move --before/--after anchors the feature at the first, last, or middle position (Outline)
- move preserves the feature's draft status and quality entries
- move preserves quality entries for an already-promoted feature
- move is allowed for a feature that another feature depends on
- move --before/--after an anchor that IS the feature itself is a no-op, byte-identical
- move refuses Backlog/Baseline as source or target, and unknown phase/anchor (Outline)
- move of a feature currently in the Backlog phase is refused, directing to promote
- move of a feature currently in the Baseline phase is refused
- move a slug that does not exist errors "not found"
- reorder repositions a feature within its own phase
- reorder repositions a feature relative to an anchor in a different phase, moving it across
- a no-op reorder (already adjacent) leaves the file byte-identical
- reorder into a Backlog/Baseline phase (via an anchor there) is refused
- reorder a slug that does not exist errors "not found"
- reorder against an unknown anchor errors clearly and leaves the file untouched
- edit/move/reorder against an empty roadmap errors cleanly with no silent mutation (Outline)
- edit/move/reorder against a missing or malformed roadmap.json surfaces an error, file untouched
- every mutation performs exactly one atomic write — a rejected edit/move/reorder writes nothing

#### UX States

| State | edit | move | reorder |
|-------|------|------|---------|
| not-found error | `stderr` contains "not found"; exit non-zero; file untouched | same | same |
| success | exit 0; only provided fields changed; untouched phases byte-identical | exit 0; feature relocated at anchor; untouched phases byte-identical; draft/quality preserved | exit 0; feature repositioned; untouched phases byte-identical |
| rejection (invalid slug / collision / cycle / unknown dep / unknown phase / unknown anchor) | exit non-zero; `roadmap.json` byte-identical to before | same | same |
| same-name / self-anchor / already-adjacent no-op | exit 0; `roadmap.json` byte-identical to before | exit 0; byte-identical to before | exit 0; byte-identical to before |
| n/a (no interactive prompts, no TUI) | n/a | n/a | n/a |

#### Out-of-Scope

`roadmap add`/`remove`/`rm`/`promote`/draft lifecycle (shipped in
`roadmap-crud-add-remove`); phase-level `add`/`rename`/`remove` operations
(successor `roadmap-phase-ops`). No persisted-schema change in this feature.

#### Deferred Findings

none (`--source roadmap-edit-move/feature-specialist`)

#### Handoff

senior-engineer — implement `internal/roadmap/{edit,edit_rename,move,reorder}.go`
+ `rewriteDependents`/`insertFeatureAt` raw helpers and the three thin cobra
commands per `docs/plans/roadmap-edit-move.md`, keeping each source +
`_test.go` ≤100 lines.
