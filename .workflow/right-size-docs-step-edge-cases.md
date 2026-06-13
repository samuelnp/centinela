# Edge Cases: right-size-docs-step

Surface-aware docs step. The hard paths cluster around surface resolution
(default vs declared), the changelog contract (missing / blank / valid), the
best-effort merge-time portal regen, and the two surface-aware hook nags.

## Covered

- **Default surface = internal (forward-only light path).**
  Risk: a brief with no `surface:` line could be mis-routed to the heavy
  user-facing path (or vice-versa), changing what evidence the docs step demands.
  Tested: `IsUserFacingFeature` returns false for a brief with no surface line
  AND for a missing brief (`internal/orchestration/feature_surface_default_test.go`);
  acceptance scenario 10 (`TestRDSDefaultSurfaceIsInternal`).

- **Internal docs step requires only the one-line changelog (no KB / portal).**
  Risk: the light path might still demand the knowledge-base bundle, defeating
  the feature. Tested: `validateDocsInternal` passes with a one-liner and NO KB
  or index.html (`validate_docs_internal_test.go`); acceptance scenario 5;
  integration `TestRDSIntegration_InternalLightPath`.

- **Missing changelog is rejected and names the file + remediation.**
  Risk: a silent pass would let an internal feature ship with no record.
  Tested: error contains `changelog entry missing` and the `<feature>-changelog.md`
  path (`validate_docs_internal_test.go`); acceptance scenario 6.

- **Blank / whitespace-only changelog is rejected DISTINCTLY from missing.**
  Risk: a created-but-empty file (placeholder never filled) could pass an
  exists-only check. Tested: a `"   \n\t\n"` file yields the distinct
  `changelog entry is empty` error, separate from the missing-file error
  (`validate_docs_internal_test.go`); acceptance scenario 7. The scanner walks
  every line, so a leading blank line followed by content still passes.

- **User-facing path is byte-identical (KB + portal still required).**
  Risk: regression that weakens the heavy path. Tested: user-facing brief passes
  only with portal + KB md + KB html, and fails naming `knowledge base markdown
  missing` when the KB is absent (acceptance scenarios 3/4, existing
  `validate_docs_test.go` preserved); integration `TestRDSIntegration_UserFacingNeedsFullBundle`;
  `RequiredRolesForFeature("docs")` still includes `documentation-specialist`
  for user-facing (`policy_docs_surface_test.go`, scenario 1).

- **Merge-time portal regen is best-effort and NEVER fails a clean merge.**
  Risk (High): a regen failure (missing roadmap.json/PROJECT.md inputs) aborting
  an otherwise-clean merge. Tested via the `docsPortalRegen` seam
  (`merge_regen_test.go`): on a clean merge the seam is invoked exactly once
  (scenario 8); when it returns an error the merge STILL succeeds and a
  `notice: portal regen skipped` is printed to stdout (scenario 9). Original seam
  restored in defer.

- **Two hook nags are surface-aware.**
  Risk: an internal feature nagged for the portal (wrong remediation) or a
  user-facing feature nagged for a changelog. Tested:
  - statusline: internal docs step blocks `MISSING_DOCS_OUTPUT`/`write-changelog`,
    and clears once the changelog exists
    (`hook_statusline_docs_internal_test.go`); user-facing path unchanged
    (existing `TestStatuslineRulesDocs`).
  - context banner: internal renders `RenderChangelogNeeded` (and NOT the
    documentation banner), and goes silent once the changelog is present
    (`hook_context_docs_internal_test.go`); user-facing still renders
    `RenderDocumentationNeeded` (existing `hook_context_docs_test.go`).

- **Changelog artifact is mechanically creatable and non-blank by construction.**
  Risk: `artifact new <f> changelog` emitting a blank first line would create a
  file that immediately fails the gate. Tested: `KindChangelog` is in
  `KindsAllowed`/`ParseKind`, and `RenderTemplate` emits exactly one file at
  `<feature>-changelog.md` whose first line is non-blank
  (`artifact_changelog_test.go`).

- **Prompt-doc + scaffold-mirror parity.**
  Risk: editing the docs-generator prompt in one copy only, breaking the parity
  acceptance gate. Verified: `diff docs/architecture/documentation-generator-prompt.md
  internal/scaffold/assets/...` is empty; the existing `scaffold_arch_parity`
  acceptance test guards it and stays green.

- **This feature dogfoods its own light path.**
  right-size-docs-step is itself internal (no `surface:` line), so it advances
  through the docs step on the new changelog one-liner — the same path under
  test. The spec-traceability gate reports all 10 scenarios COVERED.

## Residual Risks

- **A user-facing feature that forgets the `surface:` declaration silently takes
  the light path (lost KB).** Out of scope to auto-detect (no oracle for intended
  surface). Mitigation: user-facing features already declare the surface for the
  code-step ux-ui gate; the surface-aware status/banner output makes the chosen
  path visible. Big-thinker confirmed 0 existing briefs are affected.

- **Portal staleness between merges.** Internal features no longer regen the
  portal per docs step; merge-time regen refreshes it once per delivery and the
  release workflow can also regen. Acceptable: the portal is a derived artifact,
  not a gate input at the internal docs step.

- **Acceptance scenarios 8/9 assert the merge-regen wiring at the source-contract
  level (merge.go) because the `docsPortalRegen` seam is unexported package
  `main`.** The deep behavioral coverage (seam invoked once; failure tolerated +
  notice emitted) lives in the colocated `cmd/centinela/merge_regen_test.go`,
  which drives a real git merge. No behavior is left unverified.
