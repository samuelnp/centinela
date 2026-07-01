# Plan — roadmap-crud-add-remove

> Feature 2 of 4. Brief: [docs/features/roadmap-crud-add-remove.md](../features/roadmap-crud-add-remove.md).
> Umbrella design: [docs/plans/roadmap-editing-suite-design.md](roadmap-editing-suite-design.md).

## Goal
Introduce the per-feature **Draft** lifecycle and the `add`/`remove` commands,
plus a generalized `promote` that finalizes an in-place draft — so a roadmap
feature can be authored directly without breaking `roadmap validate`. Reuse the
existing format-preserving raw I/O layer; no naive `Save()`.

## Deliverables

### Draft field + hooks
- `internal/roadmap/types.go` — add `Draft bool \`json:"draft,omitempty"\`` to `Feature`.
- `internal/roadmap/backlog.go` `NonBacklogFeatureSet` — add `if f.Draft { continue }`
  (THE single coverage-set exemption; nowhere else).
- `internal/roadmap/roadmap.go` `Summary()` — exclude drafts from committed counts.
- `internal/roadmap/readiness.go` `classifyFeature` — draft → `State:"draft"`,
  excluded from `ReadySet`.
- `internal/roadmap/mdgen_feature.go` — deterministic ` *(draft)*` marker.
- `cmd/centinela/start_guard.go` (or new `start_guard_draft.go` if it would exceed
  100 lines) — refuse `start` on a draft, mirroring the Backlog refusal.
- `internal/roadmap/draft.go` (~35) — `IsDraftFeature(r,name) bool`, `DraftFeatures(r) []Feature`.

### JSON view extension (from feature 1)
- `internal/roadmap/view_types.go` — add `Draft bool \`json:"draft,omitempty"\`` to `FeatureView`.
- `internal/roadmap/view.go` `BuildView` — set `Draft` and `Readiness:"draft"` for drafts.

### Generalized raw-feature helpers (layered on `rawDoc`)
- `internal/roadmap/rawfeature_find.go` (~70) — `findFeature(slug) (raw, phaseIdx, featIdx, err)`, `featurePhase(slug)`.
- `internal/roadmap/rawfeature_mutate.go` (~90 → split if needed) —
  `appendFeatureToPhase(target, entry)` (full-feature, non-Backlog, dup-guarded),
  `removeFeatureAt(phaseIdx, slug)` (generalizes `removeBacklogFeature`),
  `replaceFeatureAt(phaseIdx, featIdx, entry)`.
- `internal/roadmap/rawtyped.go` (~30) — `toRoadmap()` decodes phases → typed `Roadmap`
  for post-mutation `ValidateDependencies` reuse.
- `internal/roadmap/rawdeps.go` (~45) — `featureDependents(slug) []string` (remove guard).

### Mutation entry points (validate-then-mutate-then-`ValidateDependencies`-then-write-once)
- `internal/roadmap/add.go` (~45) — `type AddRequest{Slug,Phase,Description,Archetype string; DependsOn []string}`;
  `Add(path, req)`: validateSlug, no-collision, `appendFeatureToPhase` (errors on
  unknown/Backlog/Baseline phase), entry = `Feature{…, Draft:true}`, `toRoadmap`+`ValidateDependencies`, write.
- `internal/roadmap/remove.go` (~45) — `Remove(path, slug)`: `findFeature` (else not-found),
  reject if `FeatureStatus != planned`, reject if `featureDependents` non-empty (name them),
  `removeFeatureAt`, write.
- `internal/roadmap/promote.go` (generalize) — detect the slug's CURRENT location:
  Backlog finding → today's move-into-`--phase`; **draft already in a schedulable phase →
  in-place finalize** (clear `Draft` via `replaceFeatureAt`, append analysis+quality
  artifacts, NO move). Branch on location, not a flag. Non-draft non-Backlog → clear error.
- `internal/roadmap/artifacts_shared.go` (~40) — extract `appendScoreArtifacts(slug,summary,scores,bullet)`
  so both promote paths stay DRY and under the line budget.
- `internal/roadmap/mutate_validate.go` (~50) — shared `requirePlannedStatus`, `requireSchedulablePhase`
  (reuse existing `validateSlug`, `validateNoCollision`).

### Thin cobra commands
- `cmd/centinela/roadmap_add.go` — `add <slug> --phase --description --depends-on(StringSlice) --archetype` → `roadmap.Add`.
- `cmd/centinela/roadmap_remove.go` — `remove|rm <slug>` (Aliases `rm`) → `roadmap.Remove`.
- `cmd/centinela/roadmap_promote.go` — extend to pass through for the in-place-draft branch
  (use cobra `Changed("scores")` sentinel like today).

## Reuse (do not reimplement)
- Raw layer: `rawio.go` (`readRawRoadmap`/`writeRawRoadmap`/`writeAtomic`), `rawmutate.go`,
  `rawmove.go` (`appendToPhase`/`removeBacklogFeature` — generalize, don't fork),
  `rawrender.go` (dirty-index render contract).
- `ValidateDependencies` (`dependencies.go`), `FeatureStatus` (`roadmap.go`),
  `ParseScores` (`promote_scores.go`), `appendPromotionArtifacts` (`promote_artifacts.go` → refactor).

## Constraints
- Every source + `_test.go` ≤ 100 lines (split per the file list above).
- Commands thin; all logic in `internal/roadmap`.
- Deterministic/byte-stable; untouched phases round-trip byte-identical.
- One mutation = one atomic write; rejected mutation leaves `roadmap.json` byte-identical.
- The draft exemption lives ONLY in `NonBacklogFeatureSet`.

## Tests (colocated for coverage — aim ≥97%)
- `add_test.go` / `add_edge_test.go` — draft flag set, unknown/Backlog phase, collision,
  invalid slug, unknown-dep, cycle rejection, file byte-identical on reject.
- `remove_test.go` / `remove_guard_test.go` — success, not-found, in-progress/done refusal,
  dependents refusal (incl. a draft dependent).
- `promote_inplace_test.go` — in-place draft finalize clears draft + writes artifacts;
  Backlog move path unchanged; non-draft error; then `ValidateAnalysis`/`ValidateQuality` pass.
- `draft_test.go` — `IsDraftFeature`, `NonBacklogFeatureSet` exemption, `Summary`/`ReadySet` exclusion.
- `rawfeature_find_test.go`, `rawfeature_mutate_test.go`, `rawdeps_test.go`, `rawtyped_test.go` —
  raw-layer units incl. exact byte-stable render assertions.
- `view_draft_test.go` — `BuildView` sets `draft`/`readiness:"draft"`.
- `mdgen_feature_draft_test.go` — the ` *(draft)*` marker.
- `cmd/centinela/roadmap_add_test.go`, `roadmap_remove_test.go`, `start_guard_draft_test.go` —
  flag parsing + draft-start refusal.
- tests/ tier trio (`tests/{unit,integration,acceptance}/roadmap_crud_add_remove_*_test.go`) with
  `// Acceptance:`/`// Scenario:` traceability tags; acceptance drives a temp-built binary (no network).

## Verification (end-to-end)
1. `go test ./...` green; `./scripts/check-coverage.sh` ≥95% (target ≥97%); `./scripts/check-fmt.sh`.
2. Dev binary: `go build -o /tmp/cen-dev ./cmd/centinela`; in a temp project:
   `roadmap add x --phase "Phase 1: …"` → `roadmap validate` still PASS; `roadmap --json` shows
   `x` with `draft:true`,`readiness:"draft"`; `roadmap ready` excludes `x`; `start x` refused;
   `roadmap promote x --scores 9,9,9,9,9,9` finalizes in place (validate still PASS);
   `roadmap remove x` guarded/works. Assert `roadmap.json` untouched on each rejected op.
3. `centinela validate` passes in the worktree.
