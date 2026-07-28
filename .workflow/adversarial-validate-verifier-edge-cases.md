# Edge Cases: adversarial-validate-verifier

## Covered

- Grounded SAFE verdict, matching revision/digest, advances `complete` —
  `tests/acceptance/adversarial_validate_verifier_happy_test.go:TestAVV_SafeVerdictAdvances`.
- Grounded WARNING verdict advances and the finding lands in the memory
  ledger — `tests/acceptance/adversarial_validate_verifier_happy_test.go:TestAVV_WarningVerdictAdvancesAndLandsInLedger`.
- CRITICAL verdict hard-blocks with the first Findings bullet echoed and a
  "FRESH verifier" instruction —
  `tests/acceptance/adversarial_validate_verifier_verdict_test.go:TestAVV_CriticalVerdictBlocksWithFindingEchoed`.
- Legacy severity aliases (`BLOCKING`, `UNSAFE`) normalize to CRITICAL and
  block — `tests/acceptance/adversarial_validate_verifier_verdict_test.go:TestAVV_LegacyAliasesBlockAsCritical` (table-driven, both aliases).
- Missing `**Status:**` line blocks with "missing or unparseable" —
  `tests/acceptance/adversarial_validate_verifier_verdict_test.go:TestAVV_MissingStatusLineBlocks`.
- Unparseable Status text (prose, not a token) blocks the same way, never
  fails open — `tests/acceptance/adversarial_validate_verifier_verdict_test.go:TestAVV_UnparseableStatusBlocks`.
- Empty `commands` array (dead-subagent stub shape) fails evidence
  validation and names `gatekeeper-prompt.md` as the remedy —
  `tests/acceptance/adversarial_validate_verifier_grounding_test.go:TestAVV_EmptyCommandsRecordBlocks`.
- `centinela artifact new <f> gatekeeper` followed immediately by
  `centinela complete <f>` FAILS (the dead-subagent regression, the second
  highest-value assertion in the plan) —
  `tests/acceptance/adversarial_validate_verifier_grounding_test.go:TestAVV_ArtifactNewStubThenCompleteFails`.
- Commands recorded but never a *passing* `centinela validate` entry
  (e.g. only `go test ./...` ran) is refused —
  `tests/acceptance/adversarial_validate_verifier_grounding_test.go:TestAVV_PartialCommandsWithoutPassingValidateRefused`.
- A verifier that cannot execute commands in its harness fails closed
  (CRITICAL + empty commands, never a narrated pass) —
  `tests/acceptance/adversarial_validate_verifier_grounding_test.go:TestAVV_NoBashHarnessFailsClosed`.
- Revision skew: a commit landed on top of the verified commit demands a
  fresh verifier, naming both the verified and current sha —
  `tests/acceptance/adversarial_validate_verifier_freshness_test.go:TestAVV_RevisionSkewDemandsFreshVerification`.
- Uncommitted in-place edits to a tracked file leave HEAD unchanged but
  stale the tree digest — the exact case a HEAD-only comparison would
  wrongly admit —
  `tests/acceptance/adversarial_validate_verifier_freshness_test.go:TestAVV_UncommittedPatchStalesTheDigest`.
- `.workflow/`-only churn after stamping does NOT stale the verification
  (D3a, both directions of the single highest-value assertion in the
  feature) —
  `tests/acceptance/adversarial_validate_verifier_freshness_test.go:TestAVV_WorkflowOnlyChurnSurvivesFreshness` and,
  through the REAL (unstubbed) git Runner rather than a stub,
  `tests/integration/adversarial_validate_verifier_integration_test.go:TestRealGitFreshnessRoundTrip`.
- `centinela artifact stamp` records the verified revision/treeDigest and
  leaves the `commands` array byte-identical —
  `tests/acceptance/adversarial_validate_verifier_freshness_test.go:TestAVV_ArtifactStampRecordsRevisionAndPreservesCommands`.
- Legacy in-flight workflow (empty `ValidateContract`) still completes
  against an old-format `validation-specialist` + checklist-style
  `gatekeeper` report — today's existence-only gate, verbatim —
  `tests/acceptance/adversarial_validate_verifier_legacy_test.go:TestAVV_LegacyInFlightWorkflowStillCompletes`.
- A workflow pinned to `adversarial-v1` refuses a hand-authored legacy pair
  (a fresh feature cannot dodge the new format by hand-authoring legacy
  files) —
  `tests/acceptance/adversarial_validate_verifier_legacy_test.go:TestAVV_FreshWorkflowRefusesHandAuthoredLegacyFormat`.
- The validate step no longer requires `validation-specialist` evidence;
  `RequiredRolesForFeature(f, "validate")` is exactly `[gatekeeper]` —
  `tests/acceptance/adversarial_validate_verifier_legacy_test.go:TestAVV_NoValidationSpecialistRequired`.
- The validate-step hook directive names the gatekeeper role, its
  reasoning tier, and the paths-only/no-summarize delegation contract
  (AC6) —
  `tests/acceptance/adversarial_validate_verifier_directive_test.go:TestAVV_HookDirectiveNamesVerifierAndReasoningTier`.
- Prompt content: refutation stance, paths-only input contract, the
  no-orchestrator-summaries prohibition, the CRITICAL token, the
  `centinela:verification` fence, the stamp instruction, and the
  fail-closed clause —
  `tests/acceptance/adversarial_verifier_prompt_test.go:TestAVVPrompt_RefutationStanceAndGroundingContract`.

## Residual Risks

- **Spec/implementation gap — WARNING verdict is not surfaced in
  `centinela complete`'s stdout.** The Gherkin scenario "A WARNING verdict
  advances validate but the finding is recorded" has two `Then` clauses:
  the finding must land in the memory ledger (verified, covered above) AND
  it "should be surfaced in the complete output". Dogfooding the real
  binary (`cmd/centinela complete`) shows only "Step ... completed" /
  "Next step: docs" — no gatekeeper-specific text prints, unlike
  `ProductionReadinessWarning`/`RenderProductionReadinessWarning`, which
  has an analogous banner for the production-readiness role but no
  equivalent exists for the gatekeeper role. This predates and is
  independent of this feature (the pre-feature `complete.go` had no such
  print either). Not asserted in the acceptance test to avoid shipping a
  test that documents a false claim. Recommend a follow-up mirroring
  `workflow.ProductionReadinessWarning` for the gatekeeper WARNING case.
- **A `.workflow/<feature>-<role>.json`-shaped filename is never inert.**
  While writing the D3a "`.workflow/`-only churn" fixture, naming a
  scratch file `<feature>-gatekeeper.json` caused `centinela complete` to
  fail differently than expected: `verify.Verify` treats any file at that
  exact path as real gatekeeper evidence and runs live claim checks
  (`tests-pass`, `coverage-moved`, etc.) against it, which then fail as
  malformed evidence rather than the test's intended assertion. The
  acceptance test now uses a non-evidence-shaped filename
  (`*-scratch.txt`) to avoid the collision. Anyone hand-writing junk under
  `.workflow/` for a repro should avoid the `<feature>-<role>.json`
  naming convention, or `verify.Verify` will treat it as claims to check.
- **Wall-clock**: every `complete` invocation in the acceptance suite runs
  a real `executeValidation()` (built-in gates) against a two/three-file
  fixture repo, which is fast (<0.3s/test). `centinela validate`'s own
  `[validate].commands` (the project's real `go test ./...` etc.) never
  runs inside these fixtures — only the fixture's own (absent) config, so
  this suite does not compound with the project's own multi-minute
  validate-step cost. Confirmed by `centinela.toml`'s `[validate]` section
  comment: `go test ./...` already includes `tests/acceptance`, so no
  separate acceptance-execution command was added or is needed.
- **Not exercised**: mechanical cross-checking of the verifier's recorded
  argv/exit codes against ground truth (deferred as
  `verify-crosscheck-verifier-commands` in the plan step, still
  out of scope for this feature) and `centinela revise`'s invalidation of
  gatekeeper evidence on a validate-step rewind (already covered by
  `tests/acceptance/workflow_revise_loop_*_test.go`, pre-existing and
  outside this feature's own scope).
