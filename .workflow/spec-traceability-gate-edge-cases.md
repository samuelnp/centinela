# Edge Cases: spec-traceability-gate

Each edge case lists the **risk** it guards against and **how it is tested**.

## Covered

- **Scenario-name normalization (trim / collapse-spaces / strip one trailing
  period / lowercase).**
  Risk: a comment that differs only in spacing, casing, or a trailing period
  would be mis-reported as an uncovered scenario, producing false gate failures.
  Tested: `normalizeScenario` directly (`TestNormalizeScenario_TrimCollapsePeriodLower`);
  end-to-end via `TestSTG_NormalizationMatches` (spec `Start the watcher` vs
  comment `// Scenario:  start the WATCHER .` -> Pass).

- **Scenario Outline counted once.**
  Risk: an outline with an N-row Examples table could be counted N times (or zero)
  and never reconcile against a single `// Scenario:` comment.
  Tested: `TestParseScenarios_OutlineAndPlainCountedWithSlug` (outline yields one
  Scenario); `TestSTG_ScenarioOutlineCountsOnce` (Pass with "1 scenarios").

- **Acceptance header with a trailing annotation.**
  Risk: real headers carry annotations like `(AC4, AC5)` after the filename; a
  strict matcher would drop the slug and mark everything uncovered.
  Tested: `TestSTG_HeaderAnnotationMatches` and
  `TestCoveredScenarios_HeaderAnnotationAndTypoAndNormalize`.

- **`spec/` (singular) header typo tolerated.**
  Risk: the known legacy typo `// Acceptance: spec/foo.feature` would silently
  contribute no coverage.
  Tested: `TestCoveredScenarios_HeaderAnnotationAndTypoAndNormalize` (slug `foo`
  recorded from a `spec/` header).

- **`// Scenario:` comment with no header above it is ignored.**
  Risk: stray scenario comments in non-acceptance files would create phantom
  coverage keyed under an empty slug.
  Tested: `TestCoveredScenarios_CommentWithoutHeaderIgnored`.

- **No spec files in scope -> Skip (not Pass/Fail).**
  Risk: an empty/absent specs dir, or a diff with no `.feature` changes, must not
  block; it must skip with an explanatory message.
  Tested: `TestCheckSpecTraceability_NoSpecsSkips`, `TestSTG_NoSpecsSkips`,
  `TestParseScenarios_MissingDirIsNotInScope`.

- **Warn vs fail severity.**
  Risk: the adoption knob must downgrade a gap to a non-blocking Warn while still
  listing it; a bug could either hide the gap or block CI.
  Tested: `TestCheckSpecTraceability_WarnSeverityWarns`,
  `TestSTG_WarnSeverityDoesNotFail` (Warn status + gap still in Details).

- **Diff-aware scope excludes unchanged specs.**
  Risk: an unchanged spec with a legacy uncovered scenario would fail the gate on
  an unrelated change, violating the documented diff-aware contract.
  Tested: `TestParseScenarios_DiffFilterIncludesAndExcludes`,
  `TestSTG_DiffAwareScopesChangedSpecs` (unchanged uncovered spec out of scope ->
  Pass).

- **Unknown severity rejected at config load.**
  Risk: a severity typo (e.g. `loud`) would silently change strictness.
  Tested: `TestValidateSpecTraceability_RejectsUnknownSeverity`,
  `TestSTG_UnknownSeverityRejected` (`config.Load` returns an error naming
  `severity`).

- **Malformed / scenario-less `.feature` and non-feature files tolerated.**
  Risk: a `.txt` containing the word `Scenario:`, or a `.feature` with no
  scenarios, must not create or break coverage; an unreadable file must not panic.
  Tested: `TestParseScenarios_NonFeatureFilesAndMalformedTolerated`,
  `TestScanScenarios_UnreadablePathTolerated`,
  `TestCoveredScenarios_NonGoFilesSkipped`.

- **Directory-read error paths propagate as Fail.**
  Risk: a spec/test path that exists but is a regular file (ENOTDIR) must surface
  as a gate Fail, not a silent Skip or panic.
  Tested: `TestParseScenarios_SpecDirIsFileReturnsError`,
  `TestCoveredScenarios_TestDirIsFileReturnsError`,
  `TestCheckSpecTraceability_ParseErrorFails`,
  `TestCheckSpecTraceability_CoverageErrorFails`.

## Residual Risks

- **Scenario Outline rows are not individually verified.** v1 treats an outline as
  one logical scenario (matches docgen counting). If per-row acceptance becomes a
  requirement, the matcher and counting must change. Mitigation: documented here
  and in the plan; out of scope for v1.

- **Coverage is keyed by (slug, normalized-name) only.** Two specs sharing a slug
  base filename in different dirs would collide. Mitigation: the repo keeps one
  flat `specs/` dir; not a concern for the current layout.

- **Centinela's own legacy backlog (397 uncovered scenarios) ships as a single
  non-blocking WARN in full-scan CI.** Intentional per the resolved plan; ratchets
  to `fail` after a backfill. Mitigation: the warn-severity path is genuinely
  tested, so the surfacing is real and actionable.
