# Edge Cases: team-dashboard

`centinela dashboard` is a read-only, pure aggregator: every degraded source
must yield an honest empty-state panel, never an error or panic. Compute is pure
and deterministic.

## Covered

### In-flight features panel
- **No active workflows** → empty `Features`; renderer prints "no active features".
  `compute_test.go::TestCompute_EmptyInputsHonestEmptyState`,
  `render_dashboard_test.go::TestRenderDashboard_AllEmptyStates`,
  `dashboard_test.go::TestRunDashboard_EmptyStates`, acceptance `TestDashboardEmptyStates`.
- **Nil entry in Active slice** skipped, not dereferenced.
  `features_test.go::TestFeatures_OwnerUnknownAndNilSkipAndOrder`,
  `dashboard_owner_test.go::TestDashboardOwners_MapsActiveAndSkipsNil`.
- **Input order preserved** (mtime-desc). `TestFeatures_OwnerUnknownAndNilSkipAndOrder`.
- **Zero / future StartedAt → age 0** (no negative); floor-days otherwise (7d23h→7).
  `features_age_test.go::TestAgeDays_ZeroFutureNormalFloor`.
- **Step index = done-count**; "done"→total, unknown step→0.
  `features_age_test.go::TestDoneCount_DonePositionUnknown`.
- **Owner fallback** missing/empty → "unknown" (advisory column).
  `features_age_test.go::TestOwnerOf_PresentEmptyMissing`,
  `TestFeatures_OwnerUnknownAndNilSkipAndOrder`.
- **Blank Profile/Archetype/Worktree** → default/canonical/— at render.
  `render_dashboard_defaults_test.go::TestRenderDashboard_BlankFieldDefaults`.
- **Git owner seam degrades** (bogus/non-repo branch → "unknown", no error).
  `dashboard_owner_test.go::TestGitOwner_DefaultUnknownOnBogusBranch`.

### Roadmap burn-down panel
- **Nil / absent roadmap** → `{Present:false}`; renderer prints "no roadmap".
  `burndown_test.go::TestBurndown_NilRoadmapEmptyState`, `TestCompute_EmptyInputsHonestEmptyState`.
- **Empty (present, zero schedulable)** → renders "0/0 done", not an error.
  `burndown_test.go::TestBurndown_EmptyRoadmapZeroTotals`,
  `render_dashboard_defaults_test.go::TestRenderDashboard_PresentEmptyRoadmap`.
- **Backlog/Baseline exclusion** from phase list and totals (mirrors `Summary()`).
  `burndown_test.go::TestBurndown_ExcludesBacklogBaselineCountsSchedulable`.
- **Done tally from workflow status** (`currentStep:"done"` → PhaseStatus.Done++).
  `burndown_done_test.go::TestBurndown_DoneCountsFromWorkflowStatus`.

### Gate health panel
- **Missing / empty telemetry** → empty `Gates`; "no gate failures recorded".
  `gatehealth_test.go::TestGatehealth_EmptyAndNoGateFailures`, acceptance `TestDashboardEmptyStates`.
- **Non-gate-failure events excluded** (block/step-advanced filtered out).
  `TestGatehealth_EmptyAndNoGateFailures`, `TestGatehealth_RanksAndMatchesInsights`.
- **Empty Gate field → "<none>" bucket** (inherited from `insights.Gates`).
  `gatehealth_test.go::TestGatehealth_EmptyGateBucketsNone`.
- **Ranking parity with insights** (count desc, key asc, verbatim).
  `TestGatehealth_RanksAndMatchesInsights`.

### Cross-cutting
- **Determinism** — same Inputs → identical Dashboard (no map ranged in output order).
  `compute_test.go::TestCompute_DeterministicSameInputs`, integration `TestTeamDashboard_RenderStability`.
- **--json stable fields** — top-level `Features`/`Roadmap`/`Gates` + sub-fields, ANSI-free.
  `dashboard_test.go::TestRunDashboard_JSONShape`, integration `TestTeamDashboard_ComputeJSONRoundTrip`,
  acceptance `TestDashboardJSONKeys`.
- **Read-only purity** — Compute does no I/O and no git; the command writes no files;
  the aggregator imports no `cmd/`/`os/exec`. Covered by the pure in-memory unit/integration
  tests (no disk writes) and the acceptance read-only contract.

## Residual Risks

- **Owner accuracy is best-effort, not authoritative.** `gitOwner` reports the last
  committer on the feature branch; detached HEAD, squash history, or no git all
  collapse to "unknown". Mitigation: advisory column, never fails the command; a real
  persisted owner model is documented Out-of-Scope (separate feature).
- **Per-phase Done depends on on-disk workflow state.** `roadmap.FeatureStatus` reads
  `.workflow/<name>.json`; a feature with no workflow file counts as "planned". This is
  intended (mirrors `roadmap.Summary()`), but means burn-down reflects local state only.
- **Acceptance owner column is "unknown" in the temp dir** (non-git tmp) — owner-name
  resolution against a real branch is exercised at the unit/seam level, not in acceptance,
  to keep the suite offline (no network, no committing into a throwaway repo).
