# Edge Cases: roadmap-crud-add-remove

## Covered

### add — rejections, each leaving roadmap.json byte-identical
- Invalid (non-kebab) slug — `validateSlug` rejects before any read.
- Slug collision with an existing feature — error names the owning phase (cross-phase).
- Unknown phase target — no silent phase creation.
- Backlog target — refused as "unknown phase" (drafts live only in schedulable phases).
- Baseline target — refused as "unknown phase" even when no Baseline phase exists.
- `--depends-on` an unknown feature — `ValidateDependencies` unknown-dep error.
- Self-dependency (`add x --depends-on x`) — reported as a dependency **cycle**, not
  unknown-dep (the appended draft IS a valid dependency target).
- Empty roadmap `{"phases":[]}` — "unknown phase", file remains exactly `{"phases":[]}`.
- Missing / unreadable roadmap.json — surfaced error, file stays absent.
- Malformed feature entry (`[123]`) / malformed phase (`"features":"x"`) — per-feature
  and per-phase decode errors propagate through findFeature/appendFeatureToPhase/
  toRoadmap/featureDependents/Add.

### remove — guards, each byte-identical on refusal
- Not-found slug → "not found".
- in-progress / done status refusal (status string surfaced verbatim).
- Dependent refusal — names the dependent; **a draft dependent still blocks remove**
  (dependency guards never special-case drafts).
- Removing the last feature of a phase leaves the phase present with `"features": []`.

### promote — generalized, branched by current location
- Draft-in-place finalize: clears `draft`, no phase move, appends analysis+quality
  artifacts, and `ValidateAnalysis`/`ValidateQuality` then PASS.
- Backlog move path unchanged (missing `--phase` and unknown-phase guards).
- Non-draft, non-Backlog slug → clear error, byte-identical.
- Overall score < 9 → refused by `ParseScores`; draft flag left intact, all five files
  unchanged.
- Missing/absent artifacts → preflight aborts the finalize, roadmap.json byte-identical.

### the four-reader draft invariant (both directions)
- `NonBacklogFeatureSet` (coverage set) EXCLUDES a draft but INCLUDES a non-draft.
- `dependencyTargetSet` INCLUDES a draft (dependable) — the coverage-vs-dependency split.
- `classifyFeature`/`ReadySet` gives a draft `State:"draft"` and excludes it from ready.
- `Summary()` does not count a draft as committed planned work.
- `BuildView` sets `draft:true`+`readiness:"draft"` and excludes drafts from `counts`.
- ` *(draft)*` deterministic ROADMAP.md marker; non-draft never carries it.

## Residual Risks

- **Filesystem write failures** (temp-file create/rename in `writeAtomic`, `Save`) are
  defensive error branches left uncovered — triggering them requires forcing OS-level
  I/O errors, which would be flaky and platform-specific. Mitigation: the write path is
  atomic (temp+rename) and shared with the already-exercised defer/promote commands.
- **Full `centinela start <draft>` refusal** is asserted at the guard level
  (`resolveArchetypeOrder`/`IsDraftFeature`, cmd/centinela) rather than by driving the
  whole `start` command through the acceptance binary, because `start` needs a valid
  PROJECT.md + centinela.toml the temp acceptance project does not carry. The acceptance
  tier instead asserts the draft's exemption through the `roadmap --json`/`ready`/
  `validate` readers. Mitigation: the guard is the single decision point `start` calls.
