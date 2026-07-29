# merge-truthful-delivery — qa-senior

## Test Inventory

Colocated (move the per-package gate):
- `internal/worktree/primary_test.go` + `primary_errors_test.go` — porcelain
  parsing: multi-worktree, bare-primary refusal, empty/garbage output,
  path with spaces.
- `internal/worktree/merge_verify_test.go` + `merge_verify_more_test.go` —
  every `verifyAdvance` branch (RefAdvanced / AlreadyMerged / neither→hard
  error), detached-HEAD and ancestor helpers via `gitRunner` stubs.
- `internal/worktree/merger_guards_test.go` — dirty, detached-HEAD and
  self-merge refusal ordering.
- `internal/worktree/merger_verify_fail_test.go` — removal-verification
  failure and unadvanced-ref hard-error paths.
- `internal/worktree/merge_realgit_test.go` + `merge_realgit_more_test.go` —
  REAL-git package-level regression: worktree-CWD merge advances main and
  removes the worktree; old-bug shape (repo pointed at the worktree)
  now refuses via the self-merge guard.
- `cmd/centinela/merge_report_test.go` — all three `reportMergeSuccess`
  branches, incl. EC-01 target-branch wording (asserted with "trunk").
- `cmd/centinela/merge_truthful_test.go` + `merge_truthful_more_test.go` —
  THE cmd-layer regression (runMerge from inside the worktree: main
  advances, worktree gone, notice printed) and unresolvable-primary refusal.

tests/ tier (spec-traceable):
- `tests/unit/merge_truthful_delivery_unit_test.go`,
  `tests/integration/merge_truthful_delivery_integration_test.go` (+helper),
  `tests/acceptance/merge_truthful_delivery_{regression,refusals,honest}_test.go`
  (+helper) — scenarios map 1:1 to specs/merge-truthful-delivery.feature
  via `// Acceptance:` + `// Scenario:` comments (8 traceability anchors).
- Updated `tests/acceptance/right_size_docs_step_merge_test.go`: its lexical
  bound was `RenderSuccess` in merge.go, which moved to merge_report.go —
  `strings.Index` returned -1 and the slice panicked. Re-anchored to the
  `reportMergeSuccess` hand-off with a -1 guard.

All test files ≤100 lines (G1); real-git fixtures are t.TempDir local repos
with explicit identity — no network remotes (suite-hang lesson).

## Coverage Gaps

Coverage gate: **97.1% ≥ 95%** (aim ≥97% met). Suite results: 631 passed
(internal/worktree + cmd/centinela), 987 passed (tests tier).
Known uncovered hard paths (accepted, documented in edge-cases report):
EC-15 external-commit race wording, EC-20/21 `Exists` IsDir gap and
mid-run path removal — low-probability races with no cheap deterministic
fixture; behavior on them is refusal-or-honest-error by construction.

## Acceptance Wiring

`[validate] commands` already executes the acceptance tier
(`go test ./tests/acceptance/...` and the full `go test ./...`). New
acceptance tests are plain Go tests — no cucumber runner, no skip risk.
EC-01 (target-branch wording) covered at unit level; EC-16 (removal refusal
after main advanced → honest non-zero exit, honest re-run) covered in
refusals acceptance test.

## Handoff

validation-specialist: run the full gate suite; the panic fixed in
right_size_docs_step_merge_test.go is the only sibling-fixture drift found.
Process note: the qa subagent stalled after authoring all test files; the
orchestrator ran the three mandated verification commands, fixed the stale
sibling fixture, added the EC-01 wording fix (code+test, allowed in tests
step), and re-ran everything green before recording this evidence.

---

## Post-Verification Round — regression tests per fix

Every fix from the verifier round carries a test at the tier that caught it.

**Finding 1 (CRITICAL) — acceptance, binary-driven, real induced conflict,
temp repos only (no network, no remote):**
`tests/acceptance/merge_truthful_delivery_continue_test.go` +
`_continue_helper_test.go`
- `TestAccMergeContinueFromWorktreeCwdNeverFakesSuccess` — stall from the
  worktree CWD on a real text conflict, valid steward evidence, conflict NOT
  resolved: asserts non-zero exit, no `and removed its worktree` anywhere in
  the output, `git rev-parse main` byte-identical, worktree still on disk.
  This is the exact command/CWD pair that printed the fabricated line.
- `TestAccMergeContinueResumesFromWorktreeCwd` — same stall, conflict really
  resolved and committed: exits 0, main advanced, `merge-base --is-ancestor`
  holds, the worktree is gone from disk AND from `git worktree list
  --porcelain`, and only then is the success line present.
- `TestAccMergeContinueResumesFromPrimaryCwd` — the same worktree-initiated
  stall resumed from the primary CWD, proving the marker is discoverable
  from both.
- The helper asserts, on every stall, that the pending marker lands in the
  PRIMARY tree — the desync at the root of the finding.
- cmd tier: `cmd/centinela/merge_continue_worktree_test.go`.
- unit tier: `internal/worktree/finalize_verify_test.go`
  (`ApplyWithoutLandedBranch_Refuses` and the landed/half-success cases),
  `merge_verify_registry_test.go` (`TestVerifyLanded_*`).

**Finding 2 — registry-verified removal:**
`internal/worktree/registry_test.go` (porcelain parsing incl. paths with
spaces, prunable entries, CRLF, runner failure),
`merge_verify_registry_test.go` (`StillRegisteredElsewhere_Refuses`,
`RegistryUnreadable_Refuses`), `merge_removal_test.go`
(`WorktreeOutsideConvention_ReallyRemoved`, real `git worktree move`), and
acceptance `TestAccMergeWorktreeOutsideConventionIsReallyRemoved`.

**Finding 3 — busy worktree half-success:**
`internal/worktree/merge_removal_test.go`
(`BusyWorktree_HalfSuccessThenForceRemoveRecovers`),
`internal/worktree/finalize_verify_test.go`
(`BusyWorktree_ReportsLandedMerge`), `cmd/centinela/merge_recovery_test.go`
(message shape, pass-through cases, `--force-remove` flag plumbing), and
acceptance `TestAccMergeBusyWorktreeReportsBothHalvesAndRecovers`, which
drives the full loop: half-success -> idempotent re-run -> forced recovery.

**Finding 4 — validate/portal scope:**
acceptance `TestAccMergeValidatesMergedPrimaryTreeNotInvokingCwd` (asserts
`Built-in Gates (full scan)` and the absence of `0 files changed`) and
`TestAccMergePortalRegenTargetsPrimaryTree` (docgen inputs reachable only
from the primary tree; the portal must appear there).

Tests changed rather than added: `TestResolveMerge_ApplyCleanFinalizes`,
`TestResolveMerge_WorktreeGoneStillFinalizes` and
`TestAcceptance_ContinueApplyFinalizes` previously finalized a merge that had
never landed (one of them aborted the merge outright). They now apply a real
steward resolution first — the old assertions were asserting the defect.

---

## Traceability Closure Round (re-verifier Finding 1)

All 21 spec scenarios now carry an executing acceptance marker; the
`spec-traceability-gate` reports `Pass — All 21 scenarios have acceptance
coverage.` (previously `Warn`, 15/21).

| Scenario | Test |
|---|---|
| a text conflict still dispatches the Merge Steward with no success claim | `TestAccMergeTextConflictKeepsWorktreeAndClaimsNothing` (`_steward_test.go`) |
| a validate failure after a clean text merge still dispatches the steward | `TestAccMergeValidateFailureDispatchesStewardWithoutClaim` (`_steward_test.go`) |
| merge refuses when the primary tree is in detached HEAD state | `TestAccMergeDetachedPrimaryHeadRefused` (`_primary_refusals_test.go`) |
| merge refuses when the primary working tree is bare | `TestAccMergeBarePrimaryRefused` (`_primary_refusals_test.go`) |
| removal is only claimed when the worktree directory is actually gone | `TestAccMergeSurvivingWorktreeIsNeverClaimedRemoved` (`_noclaim_test.go`) |
| the success message is never printed when the ref did not advance | `TestAccMergeNoRefAdvanceNeverPrintsSuccess` (`_noclaim_test.go`) |

Fixture notes (`merge_truthful_delivery_fixtures_test.go`):
- `mtdfBarePrimaryRepo` clones a fixture into a LOCAL bare repo (`git clone
  --bare <local dir>` — no network) and registers a linked worktree, so the
  first porcelain entry carries `bare`.
- `mtdfValidateFailRepo` commits `validate.commands = ["exit 1"]` on main and a
  non-conflicting change on the branch, so the merge is text-clean and the
  post-merge validate genuinely fails.
- `mtdfNoOpMergeGit` writes a `git` shim that no-ops ONLY `merge --no-ff` and
  `exec`s the real binary for everything else. Real git cannot produce "merge
  exits 0, HEAD unmoved, branch NOT an ancestor" — the two conditions are
  coupled — so the shim manufactures that one shape while the shipped guard is
  still exercised end-to-end through the real `centinela` binary. No scenario
  was demoted to a lower tier.

Each test asserts the behaviour, not merely the exit code: the ref before/after,
worktree survival, the pending-marker reason (`post-merge-validate-failed`),
git's own worktree registry, and the absence of any success wording.
