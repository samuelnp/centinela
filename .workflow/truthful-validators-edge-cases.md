# Edge Cases: truthful-validators

## Covered

- Malformed `centinela.toml` on `evidence validate` surfaces as a config
  parse error and never prints "evidence ok" —
  `tests/acceptance/truthful_validators_evidence_test.go:TestTV_EvidenceValidate_UnparseableConfig_IsReportedError`
- Explicitly empty `orchestration.ui_paths = []` falls back to built-in
  defaults instead of matching nothing —
  `tests/acceptance/truthful_validators_evidence_test.go:TestTV_EvidenceValidate_EmptyUIPaths_FallsBackToDefaults`
- No `centinela.toml` at all still applies the UX-output rule against
  built-in defaults (not a permanently-nil, always-failing list) —
  `tests/acceptance/truthful_validators_evidence_test.go:TestTV_EvidenceValidate_NoConfigFile_UsesBuiltinDefaults`
- A `go test -json` package-level `"skip"` event with no `Test` field
  ("[no test files]") is not miscounted as a scenario skip —
  `tests/acceptance/truthful_validators_skip_edge_test.go:TestTV_Skip_PackageLevelNoTestFilesIsNotAScenarioSkip`
- A cucumber-summary-shaped substring embedded mid-line in a test's own log
  output (not anchored at line start) is never matched as a real summary —
  `tests/acceptance/truthful_validators_skip_edge_test.go:TestTV_Skip_SummaryShapedSubstringInsideLogLineIsIgnored`
- A non-zero exit code always wins over any skip text in the same output —
  the failure is never re-labelled as a skip verdict —
  `tests/acceptance/truthful_validators_skip_edge_test.go:TestTV_Skip_FailingExitCodeWinsOverSkipText`
- A partial/truncated summary (command killed mid-write) never derives a
  skip verdict from the fragment; the exit failure is reported instead —
  `tests/acceptance/truthful_validators_skip_more_test.go:TestTV_Skip_TruncatedReportNeverDerivesASkipVerdict`
- A non-acceptance command (no acceptance-triggering substring) reporting
  skip-shaped text is never touched by skip analysis —
  `tests/acceptance/truthful_validators_skip_more_test.go:TestTV_Skip_NonAcceptanceCommandNeverFailedBySkipDetection`
- An empty `validate.commands` list runs nothing and performs no skip
  analysis (no crash on a zero-length command set) —
  `tests/acceptance/truthful_validators_skip_more_test.go:TestTV_Skip_NoConfiguredCommandsIsUnaffected`
- An unrecognized acceptance report shape is a WARN, never silently a PASS
  and never a FAIL (an undetected skip is not a proven skip) —
  `tests/acceptance/truthful_validators_skip_edge_test.go:TestTV_Skip_UnparseableReportIsAWarning`
- `acceptance_skip_policy = "maybe"` (an unknown value) is a hard config-load
  error naming the three supported values, never silently normalized to the
  `fail` default —
  `tests/acceptance/truthful_validators_policy_test.go:TestTV_Policy_UnknownValueIsAConfigError`
- The claim verifier does not invent a failure it cannot prove: an
  unparseable reused/fresh report stays a PASS naming the limitation —
  `tests/acceptance/truthful_validators_verify_test.go:TestTV_Verify_UnparseableStaysPass`
- A missing `scores` object, a wrongly-typed `scores` value (array), a
  missing individual score field, and a non-integer score are all reported
  as SHAPE faults and never fall through to the range-error message —
  `tests/acceptance/truthful_validators_quality_test.go` (all four cases)
- A missing/absent `features` list in roadmap-quality.json is a named
  structural error, not a panic or a silent empty-pass —
  `tests/acceptance/truthful_validators_quality_more_test.go:TestTV_Quality_MissingFeaturesListIsStructural`
- An empty diff-aware scope makes G1 report SKIP (not a false PASS) and G11
  report SKIP when no locale file is in scope —
  `tests/acceptance/truthful_validators_gates_filesize_test.go:TestTV_G1_EmptyDiffScopeReportsSkip`,
  `tests/acceptance/truthful_validators_gates_i18n_more_test.go:TestTV_G11_FilteredOutOfDiffScopeReportsSkip`
- A single configured locale still FAILS (never silently WARNs past a real
  fault) when the file is missing or malformed —
  `tests/acceptance/truthful_validators_gates_i18n_test.go:TestTV_G11_SingleLocaleMissingFileStillFails`,
  `TestTV_G11_SingleLocaleMalformedFileStillFails`
- A SKIP/WARN-only gate set never turns the overall run red and is not
  counted as a passed gate in the rendered line —
  `tests/acceptance/truthful_validators_gates_i18n_more_test.go:TestTV_G11_SkippedGateNeverTurnsGreenRunRed`

## Residual Risks

- **Parser breadth (R3/R4 in the plan).** v1 only parses `go test` (`-json`
  and `-v`) and the cucumber/godog `N scenarios (…)` summary line. behave,
  pytest-bdd, RSpec, Jest and Playwright land in the honest
  undetermined/WARN bucket rather than being detected. Tracked as the
  planner's deferred `acceptance-skip-parser-breadth`; not re-deferred here.
- **This repo's own command 1** (`go test ./... -coverprofile=coverage.out`,
  non-verbose) carries no skip data by construction. Per the senior-engineer's
  documented divergence from plan §D6 (R4), this is rendered as a quiet
  Pass-with-note rather than a permanent WARN — pinned indirectly by
  `internal/acceptance/parse_gotest_test.go`'s non-verbose case (code step);
  not re-tested here since it is unit-level parser behavior, not an
  assembled CLI/claim-verifier path this step owns.
- **`PriorTestRun` reuse end-to-end.** The reuse path only fires from
  `centinela complete` at the validate step and requires a genuine git tree,
  `workflow.VerificationFresh`, and a real `executeValidation()` run. Driving
  it fully end-to-end was judged disproportionate (the same tradeoff this
  repo already made in
  `tests/acceptance/dedupe_validate_suite_runs_complete_test.go`). Covered
  behaviorally at the unit tier
  (`internal/verify/claim_tests_acceptance_test.go:TestCheckTestsPass_PriorRunAppliesTheAcceptanceRule`)
  and pinned structurally at the acceptance tier
  (`tests/acceptance/truthful_validators_verify_priorrun_test.go:TestTV_Verify_PriorRunWiringReachesTheAcceptanceRule`).
  If that wiring is ever refactored, both must be re-verified together.
- **`centinela verdict` / `pr-gate` machine output.** The new per-command
  verdict (Pass/Warn/Fail) is only rendered in the terminal; CI consumers of
  machine-readable output (verdict JSON, pr-gate) cannot currently see an
  undetermined acceptance report. This is the senior-engineer's already-
  recorded `validate-command-verdicts-in-machine-output` deferral — not
  re-deferred here.
