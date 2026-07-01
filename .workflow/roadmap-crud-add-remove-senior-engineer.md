# roadmap-crud-add-remove — senior-engineer

Added the per-feature **Draft** lifecycle (four agreeing readers, single
exemption predicate) plus `roadmap add`/`remove`/`rm` and a location-branched
`promote` that finalizes an in-place draft — reusing the format-preserving raw
I/O layer (no naive `Save()`). Every mutation is validate-then-mutate-in-memory-
then-write-once, so a rejected op leaves `roadmap.json` byte-identical.

## Files Touched

### New — internal/roadmap
- `draft.go` (34) — `IsDraftFeature`, `DraftFeatures`.
- `rawfeature_find.go` (40) — `findFeature(slug)→(raw,phaseIdx,featIdx)`, `featurePhase`.
- `rawfeature_mutate.go` (66) — `appendFeatureToPhase` (full-feature, dup-guarded,
  non-schedulable refused), `removeFeatureAt` (generalizes `removeBacklogFeature`),
  `replaceFeatureAt` (in-place draft clear).
- `rawtyped.go` (18) — `toRoadmap()` decodes mutated phases for a post-mutation
  `ValidateDependencies` pass before any write.
- `rawdeps.go` (33) — `featureDependents(slug)` (remove guard; drafts included).
- `mutate_validate.go` (42) — `requirePlannedStatus`, `requireNoDependents`.
- `add.go` (55) — `AddRequest`, `Add`.
- `remove.go` (27) — `Remove`.
- `artifacts_shared.go` (27) — `appendScoreArtifacts` shared by both promote paths.
- `promote_inplace.go` (65) — `promoteDraftInPlace`, `draftSummary`, `draftFinalizeBullet`.

### New — cmd/centinela
- `roadmap_add.go` (48) — thin `add` command (`--phase --description --depends-on --archetype`).
- `roadmap_remove.go` (31) — thin `remove` command, alias `rm`.
- `start_guard_draft.go` (14) — `draftStartError`.

### Modified — internal/roadmap
- `types.go` (40) — `Feature.Draft bool json:"draft,omitempty"`.
- `backlog.go` (94) — `NonBacklogFeatureSet` delegates to new `schedulableFeatureSet`
  (THE single draft exemption lives here); added `dependencyTargetSet` (drafts INCLUDED).
- `roadmap.go` (82) — Reader 3: `Summary()` skips drafts.
- `readiness.go` (87) — Reader 2: `classifyFeature` → `State:"draft"` (excluded from `ReadySet`).
- `view.go` (81) — Reader 4: `buildFeatureView` sets `draft:true`+`readiness:"draft"`;
  drafts listed but not tallied into `counts`.
- `view_types.go` (38) — `FeatureView.Draft`.
- `mdgen_feature.go` (71) — deterministic trailing ` *(draft)*` marker.
- `dependencies.go` — dep target set now includes drafts (real, dependable features).
- `promote.go` (95) — `Promote` dispatches by LOCATION: Backlog→move (`promoteFromBacklog`,
  unchanged) vs draft-in-place (`promote_inplace.go`) vs clear error.
- `promote_artifacts.go` (55) — `appendPromotionArtifacts` now delegates to `appendScoreArtifacts`.
- `rawmove.go` (76) — `removeBacklogFeature` is now a thin alias over `removeFeatureAt`.

### Modified — cmd/centinela
- `roadmap_promote.go` (96) — `--phase` required only for the Backlog/evaluator path;
  scored path resolves per branch; branch-aware success message.
- `start_guard.go` (99) — `resolveArchetypeOrder` refuses a draft (mirrors Backlog refusal),
  independent of `--archetype`.

## Architecture Compliance

- **n-tier layering:** all business logic lives in `internal/roadmap`; the three
  cobra commands are thin (parse flags → call into the package → render). No logic
  in the outer layer.
- **Line budget:** every touched source file ≤100 lines (largest: `start_guard.go`
  99, `roadmap_promote.go` 96, `promote.go` 95, `backlog.go` 94). Verified by `wc -l`.
- **Reuse, not fork:** the raw I/O layer (`rawio`/`rawmutate`/`rawmove`/`rawrender`)
  is reused verbatim; `removeBacklogFeature`→`removeFeatureAt` and
  `appendPromotionArtifacts`→`appendScoreArtifacts` are generalized, not duplicated.
- **THE four-reader invariant:** source of truth is persisted `Feature.Draft`. The
  exemption predicate exists ONLY in `NonBacklogFeatureSet` (via `schedulableFeatureSet(_,false)`);
  the other three readers (`classifyFeature`, `Summary`, `BuildView`) read `f.Draft`
  directly. Dogfooded: a freshly-added draft simultaneously (1) keeps `validate` PASS,
  (2) is excluded from `roadmap ready`, (3) is not counted in `counts.planned`,
  (4) serializes `draft:true`+`readiness:"draft"`, and (5) is refused by `start`.
- **Determinism/byte-stability:** untouched phases round-trip byte-identically;
  `roadmap --json` is byte-identical across consecutive runs; a rejected
  add/remove/promote leaves `roadmap.json` byte-identical (sha-verified).
- **i18n:** matches sibling roadmap commands (same `ui.RenderSuccess` convention;
  no new locale surface introduced — consistent with the existing roadmap CLI).

## Type-Safety Notes

- No `interface{}`/`any`. New request types are concrete structs (`AddRequest`).
- `DependsOn` is coalesced to a non-nil `[]string` on add so it serializes as `[]`.
- Two distinct feature-name sets are now explicit: the analysis/quality **coverage
  set** (`NonBacklogFeatureSet`, drafts excluded) vs the **dependency-target set**
  (`dependencyTargetSet`, drafts included). Previously conflated; splitting them is
  what lets a self-dependent draft be reported as a cycle rather than an unknown dep.

## Trade-Offs

- **Draft summary fallback:** an in-place finalize with no `--summary` and no
  `description` falls back to the slug for the (required, non-empty) quality entry
  summary. Chosen over failing the finalize; operators can pass `--summary`.
- **`start` guard placement:** put at the top of `resolveArchetypeOrder` (one extra
  `roadmap.Load`) rather than only inside `workflowOrderForFeature`, so a draft that
  carries an `--archetype`/roadmap archetype cannot bypass the refusal.
- **Concurrency:** unchanged last-writer-wins atomic temp+rename (not upgraded to
  locking) — same guarantee as `defer`/`promote`.

## Verification

- `go build ./...` and `go vet ./...` clean.
- Test-compile of all packages (`go test ./... -run xxxNONE`) clean — no sibling
  test-compile break from the `Feature`/`FeatureView` signature additions.
- Full `go test ./...` green (exit 0). No tests authored (next step).
- Dogfooded against a temp project (dev binary): add happy-path + all seven reject
  rows (byte-identical, sha-verified), four-reader agreement, `start` refusal,
  in-place finalize (no move, draft cleared, artifacts written, validate PASS),
  Backlog move path unchanged, below-9 refusal (draft intact, artifacts untouched),
  remove success + not-found + draft-dependent refusal, `--json` determinism,
  ` *(draft)*` marker + double-render determinism.

## Deferred Findings

None.

## Handoff

**→ qa-senior.** Colocate `internal/roadmap/*_test.go` (≤100 lines each) to hit the
≥97% target and the tests/ tier trio with `// Acceptance:`/`// Scenario:` tags per
the plan's test list. Priority coverage: the four-reader invariant both ways
(exempt vs leak), byte-identical-on-reject for every rejected mutation, both
promote branches selected by location, the coverage-set-vs-dependency-set split
(self-dep cycle, depend-on-a-draft allowed), and the raw-layer byte-stable render
assertions. Acceptance must drive a temp-built binary (no network).
