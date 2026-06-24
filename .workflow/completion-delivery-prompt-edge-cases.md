# Edge Cases: completion-delivery-prompt

> The delegated qa-senior subagent was not used for this feature; the test suite was authored
> directly by the orchestrator and every gate re-verified independently.

## Covered

- **Completion emits guidance only — never delivers.** The done-branch prints the panel + directive
  and performs no push/merge. Tests: `tests/acceptance/completion_delivery_complete_test.go:TestAccCompleteEmitsTextOnly`,
  `cmd/centinela/complete_delivery_test.go:TestCompleteDoneEmitsDeliveryDirective`.
- **Option matrix — all four rows.** origin&&worktree→[pr,merge]; origin→[pr]; worktree→[merge];
  neither→none ("no delivery target"). Tests: `internal/gitutil/options_test.go:TestDeliveryOptionsMatrix`
  and the four `tests/acceptance/completion_delivery_complete_test.go` scenarios.
- **`HasOriginRemote` distinguishes absent-remote from a real failure.** A non-zero `git` exit
  (ExitError) ⇒ (false, nil); a real exec failure ⇒ error; empty URL ⇒ false. Tests:
  `internal/gitutil/remote_test.go`, and a real-git check in
  `tests/integration/completion_delivery_integration_test.go:TestHasOriginRemoteRealRepos`.
- **`deliver` requires an explicit, supported `--via`.** Missing `--via` (cobra required flag) and an
  unsupported value both refuse with no side effects. Tests:
  `cmd/centinela/deliver_test.go:TestRunDeliverRejectsBadVia`,
  `tests/acceptance/completion_delivery_deliver_test.go:TestAccDeliverNoVia`.
- **`--via pr` without an origin remote refuses and never pushes.** Tests:
  `cmd/centinela/deliver_test.go:TestRunDeliverPRWithoutOrigin`,
  `cmd/centinela/deliver_pr_test.go:TestRunDeliverPRNoOrigin`,
  `tests/acceptance/completion_delivery_deliver_test.go:TestAccDeliverPRNoOrigin`.
- **`--via merge` without worktree mode refuses; with worktree mode it delegates to the merge flow**
  (not guard-rejected). Tests: `cmd/centinela/deliver_test.go:TestRunDeliverMergeWithoutWorktree`,
  `tests/acceptance/completion_delivery_deliver_test.go:TestAccDeliverMergeDelegates`.
- **`--via pr` refuses to push a dirty tree** (no auto-commit of unreviewed work). Test:
  `cmd/centinela/deliver_pr_test.go:TestRunDeliverPRDirtyTree`.
- **Honest PR degradation** — the pr path is reached past the guards and never falsely claims a PR was
  opened. Test: `tests/acceptance/completion_delivery_deliver_test.go:TestAccDeliverPRPathReachedHonestly`.

## Residual Risks

- **Real push / `gh pr create` / clean-merge success are not reproduced in tests** — they require a
  live remote and `gh` auth. Acceptance asserts the deterministic, reachable guarantees (guards
  passed, no false PR claim, delegation reached); the real PR-open and merge-finalize behaviors are
  exercised by `gh`/`git` and the existing `centinela merge` command's own acceptance tests. The
  `gh`-absent push+manual-instructions branch is environment-dependent (depends on `gh` being on
  PATH) and is therefore not unit-asserted.
- **Non-GitHub remotes** are out of scope: `--via pr` uses `gh` (GitHub-only); other remotes degrade
  to push + manual instructions. Rich PR body/changelog is deferred to `delivery-artifact-generation`.
