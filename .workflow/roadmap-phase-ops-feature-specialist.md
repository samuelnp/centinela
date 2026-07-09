# Feature-Specialist Report: roadmap-phase-ops

**Date:** 2026-07-09

## Behavior Summary

This feature completes the Roadmap Editing Suite by adding phase-level structural
operations — `roadmap phase add <name> [--note --after <phase>]`, `roadmap phase
rename <old> <new>`, and `roadmap phase remove <name> [--force]` — on top of the
format-preserving raw layer already used by the shipped feature-level commands
(`roadmap-crud-add-remove`, `roadmap-edit-move`). `add` inserts a new empty phase
either right after a named anchor or just before the reserved `Backlog` phase (else
last); `rename` renames a phase in place without touching its features or any other
phase; `remove` deletes an empty phase outright, refuses a non-empty phase unless
`--force` is given (naming the feature count), and with `--force` prunes the phase,
its features, and those features' analysis/quality entries in one atomic write,
re-validating dependencies/analysis/quality before committing and refusing
byte-identical if the result would break (e.g. a surviving dependent). All three
commands reuse the existing validate-then-mutate-then-write-once contract: a
rejected op never reaches `writeRawRoadmap`. The core new complexity is that
inserting/removing a phase shifts every later phase's index, so the raw layer's
`dirty` map (keyed by phase index) must be reindexed in `insertPhaseAt`/
`removePhaseAt` so a phase op and a feature mutation in the same run both render
correctly instead of corrupting an unrelated, already-dirty later phase.

## Acceptance Criteria (Gherkin)

All scenarios live in `specs/roadmap-phase-ops.feature`:
- `phase add with --after inserts immediately after the named phase`
- `phase add without --after lands before the Backlog phase`
- `phase add without --after and without a Backlog phase lands last`
- `phase add --note sets the phase note`
- `phase add --after the Backlog phase inserts as a normal schedulable phase, not inside Backlog`
- `phase add on an empty roadmap succeeds as the first phase`
- `phase add refuses a duplicate name, reserved name, empty name, or unknown --after anchor` (outline, 5 examples)
- `phase rename renames in place, leaving its features and other phases untouched`
- `phase rename to the SAME name is a no-op, byte-identical`
- `phase rename refuses an unknown old name, a collision, an empty new name, or either side reserved` (outline, 5 examples)
- `phase remove deletes an empty phase`
- `phase remove of a non-empty phase without --force is refused, naming the feature count`
- `phase remove --force removes the phase, its features, and their analysis/quality entries, then validate PASSes`
- `phase remove --force is REFUSED byte-identical when a surviving feature depends on a removed one`
- `phase remove of an unknown phase errors "not found"`
- `phase remove refuses the reserved Backlog/Baseline phase, with or without --force` (outline, 3 examples)
- `phase remove of the only phase leaves an empty roadmap`
- `inserting an earlier phase while mutating a later feature reindexes the dirty map so both renders are correct`
- `removing a middle phase while mutating a later feature reindexes the dirty map so both renders are correct`
- `phase rename/remove against an empty roadmap errors "not found"` (outline, 2 examples)
- `phase add/rename/remove against a missing roadmap.json surfaces an error and leaves the file absent`
- `phase add/rename/remove against a malformed roadmap.json surfaces an error and leaves the file untouched`
- `every phase add/rename/remove performs exactly one atomic write — a rejected op writes nothing`

## UX States

| State | CLI trigger | Result |
|---|---|---|
| Unknown phase (rename/remove) | `phase rename "Phase 9: X" "Y"` / `phase remove "Phase 9: X"` | non-zero exit, stderr `"not found"`, file byte-identical |
| Unknown --after anchor (add) | `phase add "X" --after "Phase 9: Y"` | non-zero exit, stderr `"unknown phase"`, file byte-identical |
| Duplicate name (add) | `phase add "Phase 1: Foundations"` | non-zero exit, stderr `"already exists"`, file byte-identical |
| Reserved name (add/rename/remove) | `phase add "Backlog"` / `phase rename x "Baseline"` / `phase remove "Backlog"` | non-zero exit, stderr `"reserved phase name"`, file byte-identical |
| Empty name | `phase add ""` / `phase rename "Phase 1: Foundations" ""` | non-zero exit, stderr `"phase name is required"`, file byte-identical |
| Add success | `phase add "Phase 3: Scale" --after "Phase 2: Growth"` | exit 0, new phase inserted with `features: []`, untouched phases byte-identical |
| Rename success | `phase rename "Phase 1: Foundations" "Phase 1: Core"` | exit 0, name updated in place, features/other phases untouched |
| Rename no-op (same name) | `phase rename "Phase 1: Foundations" "Phase 1: Foundations"` | exit 0, file byte-identical |
| Non-empty remove refusal (no --force) | `phase remove "Phase 2: Growth"` | non-zero exit, stderr names feature count (e.g. `"2 features"`), file byte-identical |
| --force success | `phase remove "Phase 2: Growth" --force` | exit 0, phase + features + analysis/quality entries removed, `roadmap validate` PASS |
| --force refused (dependent survives) | `phase remove "Phase 2: Growth" --force` (a surviving feature dependsOn a removed one) | non-zero exit, stderr contains `"depends on"`, all three files (`roadmap.json`, `roadmap-analysis.json`, `roadmap-quality.json`) byte-identical |
| Remove-only-phase | `phase remove "Phase 1: Foundations"` (only phase, empty) | exit 0, `.workflow/roadmap.json` becomes exactly `{"phases":[]}` |
| Malformed/missing roadmap.json | any `phase` subcommand | non-zero exit, stderr error message, file untouched/absent |
| i18n | n/a — CLI stderr/stdout strings for this feature, no UI component | n/a |

## Out-of-Scope

Feature-level `add`/`remove`/`edit`/`move`/`reorder` — shipped in
`roadmap-crud-add-remove` and `roadmap-edit-move`. No new persisted-schema fields;
no Magallanes UI work (CLI + `roadmap --json`/ROADMAP.md consumption only).

## Deferred Findings

none (`--source roadmap-phase-ops/feature-specialist`)

## Handoff

senior-engineer
