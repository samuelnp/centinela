# Edge Cases: brownfield-roadmap-generation

## Covered

| Hard path | Behavior asserted | Test(s) |
|-----------|-------------------|---------|
| Missing inventory | Non-zero exit, error guides to `centinela analyze`, **no draft written** | `cmd/centinela/roadmap_brownfield_errors_test.go::TestRoadmapBrownfield_MissingInventoryGuidesToAnalyze`; `tests/acceptance/brownfield_edge_test.go::TestAccBrown_MissingInventoryGuidesAndWritesNothing` |
| Malformed inventory | Distinct error (not the analyze-guidance path) | `cmd/centinela/roadmap_brownfield_errors_test.go::TestRoadmapBrownfield_MalformedInventoryErrors` |
| Canonical roadmap.json never clobbered | Pre-existing curated `roadmap.json` is **byte-equal** after the run; draft lands at the draft path only | `internal/brownmap/write_test.go::TestWriteDraft_RefusesCanonicalRoadmap`; `tests/integration/brownfield_pipeline_test.go::TestBrownfieldPipeline`; `tests/acceptance/brownfield_happy_test.go::TestAccBrown_NeverClobbersCanonical` |
| Refusing `--out` = canonical path | WriteDraft refuses; wrapped as a write error; no file created at the canonical path | `cmd/centinela/roadmap_brownfield_errors_test.go::TestRoadmapBrownfield_RefusesCanonicalOut` |
| Empty / doc-only inventory | Empty (non-nil) Baseline, **0 gaps**, no malformed draft, summary reports 0/0 | `internal/brownmap/generate_test.go::TestGenerate_EmptyInventoryEmptyBaselineNoGaps`; `tests/acceptance/brownfield_edge_test.go::TestAccBrown_DocOnlyEmptyBaselineZeroGaps` |
| No-TODO / no-goal → Baseline-only + hint | No Gaps phase in the draft; summary hints `supply --goal` | `internal/brownmap/gaps_test.go::TestGapPhases_NoWorkReturnsNil`; `internal/ui/render_brownfield_test.go::TestRenderBrownfieldSummary_NoGapsHint`; `tests/acceptance/brownfield_edge_test.go::TestAccBrown_BaselineOnlyDraftHasHint` |
| Determinism (byte-identical re-run) | Two runs on an unchanged inventory produce byte-identical draft JSON | `internal/brownmap/write_test.go::TestWriteDraft_Deterministic`; `internal/brownmap/generate_test.go::TestGenerate_Deterministic`; `tests/acceptance/brownfield_edge_test.go::TestAccBrown_Deterministic` |
| Baseline exempt from status/coverage/readiness | Baseline features excluded from `Summary`, `NonBacklogFeatureSet`, `DeriveReadiness`; same predicate that exempts Backlog; Backlog/Bootstrap regression unchanged | `internal/roadmap/baseline_test.go` (all); `tests/unit/brownfield_baseline_unit_test.go::TestBrownfield_BaselineExcludedFromStatusAndCoverage`; `tests/acceptance/brownfield_edge_test.go::TestAccBrown_BaselineExcludedFromStatusAndCoverage` |
| Empty role on a target | Description renders the module label rather than a blank role | `internal/brownmap/baseline_test.go::TestRoleOrModule_NormalizesEmpty` |
| TodoTargets accessor | Returns TODO-bearing targets in order; zero-able (nil) when no marker present | `internal/reconstruct/todotargets_test.go` (both) |
| Mkdir / rename write failure | Parent-dir creation failure and rename-over-directory failure are surfaced as errors; temp file cleaned up | `internal/brownmap/write_test.go::TestWriteDraft_MkdirFailureWrapped`; `internal/brownmap/write_more_test.go::TestWriteDraft_RenameFailureWrapped` |

## Residual Risks

- **`atomicWrite` mid-write fault branches** (`tmp.Write`/`tmp.Close` errors) and the `json.MarshalIndent` error in `WriteDraft` are not unit-covered: they require OS-level fault injection and a Roadmap of plain structs cannot fail to marshal. Overall statement coverage stays at 95.1% (≥ 95.0%); these branches are defensive returns with no logic. Mitigation: the rename and mkdir failure paths (the realistic crash-safety edges) *are* covered.
- **Today every reconstructed target is TODO-bearing** (the reconstruct skeleton emits a uniform set of `# TODO: confirm` markers), so the binary cannot produce a `baseline > 0, gaps == 0` draft. The "Baseline-only + hint" acceptance scenario is therefore driven by the doc-only inventory (0 baseline, 0 gaps, hint shown) and the unit layer (`gapPhases(nil,nil) == nil`), which together prove the no-gaps draft/hint behavior. The accessor is genuinely zero-able, so a future TODO-free reconstruction would still produce a Baseline-only-with-content draft.
