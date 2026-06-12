### Gatekeeper Report: enforcement-profiles
**Date:** 2026-06-12
**Status:** SAFE

#### Analyzed Specs
All 76 `specs/*.feature` scanned. Materially relevant (describe behavior this feature
touches): `configurable-step-confirmation-mode.feature`, `enforce-step-subagent-orchestration.feature`,
`enforce-actionable-orchestration-evidence.feature`, `auto-start-feature-intent.feature`,
`fix-roadmap-write-blocked.feature`, and this feature's own `enforcement-profiles.feature`
(12 scenarios, all with acceptance coverage). Every relevant spec describes today's
DEFAULT (strict) behavior, which this feature reproduces exactly; non-default profiles
are opt-in and assert no existing spec's behavior.

#### Findings

**No conflicts detected.** Behavior-preservation verified across all six core-machinery
checks:

1. **NewWithOrder signature change — SAFE.** Five call sites grepped; all updated:
   `start.go`, `hook_autostart.go` (production), `state.go::New` (passes `ProfileStrict`),
   `order_test.go`, `start_guard_test.go`, plus the new `order_profile_test.go` and
   acceptance evidence tests. No stale 2-arg caller remains. Every existing assertion that
   a default-started feature carries `OrchestrationMode == "strict-subagents-v1"` still
   holds: default → strict → `ProfileDefaults(strict).RequireSubagentEvidence == true` →
   `StrictOrchestrationMode` set. Confirmed in `order.go:31-34` and `state.go:77`.

2. **Prewrite gating change — SAFE.** `prewrite.go:41-42` only adds an `EffectiveProfile == outcome`
   short-circuit OR'd before `IsAllowedInStep`. Under strict/guided the OR's left side is
   false, so flow falls through to the identical `IsAllowedInStep` block. The no-active-workflow
   (`len(wfs)==0`), done-skip, and TypeRoadmap/TypeOther early-returns are untouched and run
   before profile logic. Existing `TestEvaluatePrewriteBranches` (plan-step code write with
   empty cfg → blocked with context) and `hook_prewrite_block_test.go` still pass — verified
   in the 1313/1313 green suite. No existing spec asserts step-gating that a DEFAULT run regresses.

3. **Confirmation resolver change — SAFE.** `RawStepConfirmationMode` captures the decoded
   value before `applyDefaults` normalizes it (`config.go:69`). Resolver precedence
   (`hook_context_review_mode.go:26-31`): explicit raw knob > profile default > hardcoded.
   A project that DID set `step_confirmation_mode` has a non-empty raw value, so its explicit
   mode still wins unchanged — no behavior change. Existing `TestRunHookContextReviewPromptModes`
   passes: unset → strict default `every_step` prompts; `after_plan`/`auto` explicit → preserved.

4. **Config struct change — SAFE.** `RawStepConfirmationMode` carries `toml:"-"`
   (`workflow_config.go:18`); it never serializes. No TOML round-trip / Marshal / snapshot
   test exists for `WorkflowConfig` (config is decode-only; the snapshot hits in the tree
   are unrelated — import_graph, roadmap, agent-evidence). `EnforcementProfile` is a plain
   additive `toml:"enforcement_profile"` field. No round-trip or snapshot test breaks.

5. **Default-behavior invariant — SAFE.** A freshly started feature with no `--profile` and
   no config resolves to strict at every layer: `New → NewWithOrder(..., ProfileStrict)` sets
   `strict-subagents-v1`; `EffectiveProfile`/`DisplayProfile` default to strict; prewrite keeps
   `IsAllowedInStep` gating; confirmation resolves to `every_step`. Verified by reading the
   resolved code paths (order.go, profile.go, prewrite.go, hook_context_review_mode.go,
   profile_defaults.go strict branch) and corroborated by the unconfigured-keeps-strict
   acceptance scenario passing.

6. **verify/gates untouched — SAFE (verification constant).** `git diff --stat main...HEAD --
   internal/verify internal/gates` is EMPTY (ran it; exit 0, no output). `complete.go`'s
   validate-step block runs `executeValidation()` + `runClaimVerification()` with no profile
   branch. The invariant test (`TestExecuteValidation_BlockedUnderEveryProfile`, in-process +
   full-binary acceptance) proves a failing gate blocks completion identically under all three
   profiles. No profile relaxes the verification axis.

**Validation:** fresh binary `/tmp/cent-gk2 validate` from the worktree → all gates pass,
including `✓ spec-traceability-gate All 12 scenarios have acceptance coverage` and
`✓ G1 All files under 100 lines`. Full suite: 1313 passed across 24 packages; coverage 95.2%.
(`⚠ import_graph` is a pre-existing "no configured layer" advisory, not introduced here.)

#### Recommendation
**SAFE** — default = strict reproduces today's behavior byte-for-byte across all six core
paths; non-default profiles are opt-in and conflict with no existing spec; verify/gates
diff is empty; full suite green.
