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
