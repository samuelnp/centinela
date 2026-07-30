# spec-conflict-false-positives — senior-engineer

## Files Touched

Implementation (`internal/worktree/`):

| File | Lines | Change |
|------|-------|--------|
| `spec_conflicts.go` | 63 | Rewritten: `DetectSpecConflicts` now compares the merging worktree against other worktrees only; `FormatSpecConflicts` names owners and caps output. |
| `spec_collect.go` | 67 | New: `worktreeSpecsDir`, `mainSpecIndex` (baseline), `otherWorktreeSpecs`, `readSpecsFrom` (moved out of `spec_conflicts.go`). |
| `spec_pairing.go` | 61 | New: `specKey`, `indexByKey`, `diverges`, `bothEdited`, `appendDivergences` (dedup). |
| `spec_parser.go` | 46 | Removed the dead `scenariosConflicts` Given-bucketing pairer; `parseScenarios` unchanged. |
| `cmd/centinela/merge.go` | — | Untouched; the pre-check call site is unchanged. |
| `cmd/centinela/merge_test.go` | 98 | `TestRunMerge_SpecConflict_Blocks` set up a main-spec-vs-worktree-spec false positive; rewritten to two worktrees diverging on one file. Caught only by the full suite — it compiled fine. |

Specs and tests:

- `specs/parallel-feature-worktrees.feature` — the spec-conflict scenario was rewritten to state the real contract (same file, same scenario, both worktrees edited away from main, reported at most once) and a companion `Superseding and identical specs never block a merge` scenario was added.
- `tests/acceptance/parallel_feature_worktrees_test.go` (186) — the spec-conflict acceptance test previously encoded a false positive (main spec vs. a differently-named worktree spec); it now drives two worktrees, and a second test covers the non-blocking classes. Traceability headers use the `// Acceptance: specs/…` + `// Scenario: …` form the gate parses.
- Colocated unit tests: `spec_conflicts_test.go` (74), `spec_conflicts_more_test.go` (60), `spec_conflicts_falsepos_test.go` (96), `spec_conflicts_baseline_test.go` (63), `spec_conflicts_format_test.go` (74), `spec_conflicts_collect_test.go` (63).
- `coverage_merge_helpers_test.go` (62) — `TestScenariosConflicts_SkipsEmptyFields` targeted the deleted pairer; replaced by `TestIndexByKey_FirstRecordWins` with a comment naming this hotfix.

## Architecture Compliance

All work is confined to `internal/worktree` (the detector's blast radius) plus its
spec and acceptance tests. No new dependencies, no exported-signature changes to
anything outside the package: `DetectSpecConflicts` and `FormatSpecConflicts` keep
their signatures, so `cmd/centinela/merge.go` is unchanged. `SpecConflict` gained
two additive fields (`OwnerA`, `OwnerB`); no existing field was removed or retyped.
Every source and test file is at or below the 100-line limit.

## Type-Safety Notes

No `interface{}`, no reflection, no type assertions. The dedup key is a
`strings.Join` over an explicit five-field tuple with a NUL separator rather than
a formatted string, so a scenario name containing the separator cannot collide
with a file name. `ownedSpecs` is a named struct rather than a bare map so the
owner slug travels with its records and cannot be lost at the call site — the
original bug was exactly that: `collectScenarios` accepted `mergingFeature` and
discarded it with `_ = mergingFeature`.

## Trade-Offs

**The behaviour contract genuinely narrowed, and the spec says so.** The old
detector paired any two scenarios sharing a Given clause across the whole
comparison set. That is not a conflict signal; it flagged two byte-identical
copies of one file against each other, two companion scenarios inside a single
file, and two unrelated files on main. The new rule is: the merging worktree and
a *different* worktree carry the same `(file, scenario)`, with different content,
and **both** differ from main's baseline.

**The baseline check was not in the original design contract and is load-bearing.**
Without it the fix is still unusable in practice. Every worktree is a full
checkout, so a bystander worktree carries main's copy of every spec the merging
feature edits — comparing worktree-to-worktree alone would flag every edit against
every idle worktree. `bothEdited` requires each side to have moved away from
main's copy; main is a reference point, never a comparison peer. When main has no
such scenario there is no baseline, and two worktrees introducing it differently
still conflict.

**Loss of coverage, accepted.** Cross-file semantic contradictions (two different
spec files asserting opposite outcomes for the same Given) are no longer detected.
They were the noisiest false-positive source and this detector's Gherkin reader is
too shallow to judge them; the merged-tree `centinela validate` run remains the
real semantic gate. A weaker precise signal beats a strong one nobody can act on.

**Output cap.** `FormatSpecConflicts` prints at most 10 entries plus an
`and N more` remainder because the string is embedded in a CLI error; the observed
failure produced 720KB. Dedup happens at detection time, so the cap only ever
elides genuinely distinct conflicts.

## Verification

- `go build ./...` — clean.
- `go vet ./...` — no issues.
- `go test ./... -run xxxNONE` — every package's tests compile (no stale callers of
  the removed `scenariosConflicts`).
- `go test ./internal/worktree/` — 145 pass.
- `go test ./...` — 3742 pass, 45 packages, exit 0.
- `centinela validate` — **All gates passed**, exit 0. G1 file size ✓, cross-compile ✓;
  three pre-existing warnings (import_graph, spec-traceability-gate, roadmap_drift —
  the last cleared by regenerating ROADMAP.md). Validate commands: `go test ./...
  -coverprofile` ✓, `check-coverage.sh` ✓, `check-fmt.sh` ✓.
- `./scripts/check-fmt.sh` — exit 0.
- `go tool cover` on `internal/worktree` — 98.6% of statements; all twelve
  `spec_*.go` functions at 100%.
- Dogfood, real binaries against real git worktrees: the old binary blocked all
  three cases; the new binary clears the two false-positive cases and still blocks
  the true positive, now reported once (574 bytes, down from 920 with duplicates).

## Deferred Findings

Two genuinely new gaps were captured to the Backlog via
`centinela roadmap defer … --source spec-conflict-false-positives/senior-engineer`:

- `spec-conflict-scenario-deletion-detection` — a worktree that *deletes* a
  scenario another worktree edited is invisible to the detector.
- `spec-conflict-deep-gherkin-diff` — only the first `Given` and first `Then`
  per scenario are compared; `When`/`And`/second-`Then` divergence is not seen.

**Side effect worth a reviewer's eye:** `centinela roadmap defer` re-serialised
`.workflow/roadmap.json` from its compact one-line-per-feature form into expanded
JSON (+99/−424 lines). That is the CLI's own serialiser, not a hand edit, and it
is semantically identical apart from the two new Backlog entries — but it is a
large diff in shared state while the `token-diet` worktree is also in flight, so
it is a likely merge conflict. `ROADMAP.md` was regenerated (`centinela roadmap
generate`) to clear the resulting `roadmap_drift` warning.

## Handoff

To **qa-senior**. Worth adversarial attention:

1. `bothEdited` is the load-bearing addition and was not in the original contract —
   probe the case where main's spec is absent, and where a worktree deletes a
   scenario rather than editing it (a missing scenario is currently invisible, not
   a divergence).
2. The detector only runs against worktrees on disk. A feature merged via PR
   whose worktree was already removed contributes nothing — confirm that is
   acceptable rather than a silent hole.
3. `parseScenarios` reads only the first `Given` and first `Then` per scenario.
   Two worktrees that differ only in a `When` step, an `And` step, or a second
   `Then` will not be flagged. Deliberate, but worth an explicit edge-case entry.
4. The spec-traceability gate parses `// Acceptance: specs/<slug>.feature` followed
   by `// Scenario: <name>`; the two renamed scenarios must stay in sync with
   `specs/parallel-feature-worktrees.feature`.
