# Edge Cases: enforcement-profiles

## Covered

- **Explicit knob beats profile default.** Risk: a profile silently overriding a
  user's explicit `step_confirmation_mode` would break back-compat. Tested by
  `TestEffectiveConfirmationMode_Precedence` and
  `TestShouldRenderReviewPrompt_ExplicitEveryStepBeatsOutcome` (cmd) plus the
  acceptance `TestEP_ExplicitConfirmationOverridesProfile`: explicit `every_step`
  renders the prompt even under outcome.

- **applyDefaults raw-capture hazard.** Risk: `applyDefaults` overwrites an empty
  `step_confirmation_mode` with `every_step`, destroying the explicit-vs-defaulted
  signal. Tested by `TestLoad_RawStepConfirmationMode_DistinguishesExplicit`
  (config): unset leaves `RawStepConfirmationMode==""` while the normalized field
  is `every_step`; an explicit value is preserved raw.

- **Per-feature override > global.** Risk: the global profile leaking past a
  per-feature `--profile`. Tested by `TestEffectiveProfile_Precedence`,
  `TestEffectiveProfile_NilInputs` (workflow) and acceptance
  `TestEP_PerFeatureOverridesGlobal`; persistence proven by
  `TestRunStart_ProfileFlagPersisted` / the integration test.

- **Unknown profile rejected at load.** Risk: a typo silently normalizing to
  strict and hiding the misconfiguration. Tested by `TestLoad_UnknownProfileRejected`
  (config), `TestValidateEnforcementProfile_RejectsUnknownNamingField`, and
  acceptance `TestEP_UnknownProfileRejectedAtLoad` — validation runs on the RAW
  value before normalization and the error names `enforcement_profile`.

- **outcome bypasses ordering, NOT artifact/verification checks.** Risk: outcome
  perceived as "no governance". Tested by `TestEvaluatePrewrite_OutcomeBypassesStepGating`
  (allows code-in-plan) alongside the invariant tests below — ordering rails drop,
  the validate gate does not.

- **Verification constant under all profiles (the invariant).** Risk: a profile
  branch relaxing gates/claim-verify at the validate step. Tested by
  `TestExecuteValidation_BlockedUnderEveryProfile` (cmd, in-process) and
  `TestEP_GatesRunUnderEveryProfile` (acceptance, full binary): a failing
  `[validate] command` blocks completion identically under strict/guided/outcome;
  `TestExecuteValidation_PassesUnderEveryProfileWhenClean` is the control.

- **No-active-workflow blocks under any profile.** Risk: outcome being read as
  "writes allowed without a started feature". Tested by
  `TestEvaluatePrewrite_NoActiveWorkflowBlocksUnderAnyProfile` (hookpolicy) and
  acceptance `TestEP_NoActiveWorkflowAlwaysBlocked`: NeedInit fires for
  strict/guided/outcome with nil and done-only workflows.

## Residual Risks

- **`DisplayProfile` shows the pinned field, not re-resolved global.** A
  pre-existing workflow with an empty pinned profile renders `strict` in status
  even if the global config later changed — accepted per the senior-engineer
  trade-off (status has no `cfg`; both fallbacks default to strict). Covered by
  `TestRenderStatus_ProfileLine` / `TestDisplayProfile`.
- **Claim-verification (`runClaimVerification`) failure path** is proven constant
  structurally (it takes no profile param and complete.go adds no profile branch);
  the behavioral invariant tests drive the gate via the validate-command channel,
  which is the same hard block. Wiring `verify.Verify` to a synthetic failing
  claim per profile is left out of scope as redundant with the structural proof.
