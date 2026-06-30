# workflow-revise-loop — qa-senior

All three tiers exist; every new/changed source file has colocated tests.
Measured total coverage: **97.4%** (gate 95%, aim ≥97% — met). All suites green;
spec-traceability: **0 uncovered scenarios** for this feature (16/16 covered).

## Test Inventory

| Tier | File | Lines | Covers |
|------|------|------:|--------|
| Unit | `internal/workflow/rewind_test.go` | 90 | RewindTo re-open matrix, reopenedSteps (canonical+hotfix+last+absent), RevisionsSummary |
| Unit | `internal/workflow/rewind_reject_test.go` | 62 | all RewindTo rejections (empty/unknown/forward/equal/done), Revisions JSON round-trip + omitempty |
| Unit | `internal/evidence/invalidate_test.go` | 80 | Invalidate both-files + idempotency, safety (source survives), InvalidateArtifact, removeBoth real-error arm |
| Unit | `internal/evidence/invalidation_targets_test.go` | 68 | validate→gatekeeper+production-readiness, tests→edge-cases.md, internal vs user-facing code (ux-ui), plan |
| Unit | `cmd/centinela/revise_test.go` | 74 | runRevise happy path + empty-reason + missing-workflow + rewind-rejected |
| Unit | `cmd/centinela/revise_invalidate_test.go` | 38 | invalidateDownstream count/dedup (role+artifact) + error surfacing |
| Unit | `internal/telemetry/revised_test.go` | 35 | RecordRevised → TypeStepRevised event (from/to/model/feature) |
| Unit | `internal/ui/render_status_revisions_test.go` | 27 | Revisions row present-with-reason / omitted-when-none |
| Integration | `tests/integration/workflow_revise_loop_integration_test.go` | 93 | end-to-end RewindTo + invalidation on temp `.workflow`: downstream evidence gone, source/test survive, revision persisted |
| Acceptance | `tests/acceptance/workflow_revise_loop_test.go` | 80 | happy-path-validate-to-code, archetype-hotfix-order |
| Acceptance | `tests/acceptance/workflow_revise_loop_audit_test.go` | 52 | audit-visible-in-status, completed-at-cleared |
| Acceptance | `tests/acceptance/workflow_revise_loop_negative_test.go` | 57 | the 6 negative scenarios |
| Acceptance | `tests/acceptance/workflow_revise_loop_safety_test.go` | 92 | safety, idempotency, accumulate, internal-no-ux |
| Acceptance | `tests/acceptance/workflow_revise_loop_regating_test.go` | 54 | re-gating blocks / advances |
| Helper | `tests/acceptance/workflow_revise_loop_helper_test.go` | 83 | seed/load workflow JSON helpers |
| Helper | `tests/acceptance/workflow_revise_loop_assert_test.go` | 28 | mustGone / mustExist |

Every file is ≤100 lines (G1, incl. test files). Binary-driven acceptance for
the CLI-level scenarios (negatives, happy path, status); package-level for
re-gating (asserts the orchestration gate primitive that `complete` reuses).

## Coverage Gaps

- New-code per-function coverage: `RewindTo`, `RevisionsSummary`, `reopenedSteps`,
  `Invalidate`, `InvalidateArtifact`, `removeBoth`, `InvalidationTargets` = 100%;
  `invalidateDownstream` 95.8%; `runRevise` 86.4% (uncovered arms are the
  `config.Load`/`saveWorkflow` I/O-error paths, which are exercised in aggregate
  by the existing cmd suite and not cheaply forceable without seam injection).
- Package totals stay well above the 95% floor; overall total 97.4%.

## Acceptance Wiring

Each acceptance file repeats the header `// Acceptance: specs/workflow-revise-loop.feature`
and every test carries a `// Scenario: <exact name>` matching the spec. Example:

```go
// Scenario: Happy path — validate step is revised back to code
func TestRL_HappyPathValidateToCode(t *testing.T) {
	...
	out, code := runCent(t, buildCent(t), dir,
		"revise", "my-feature", "--to", "code", "--reason", "bug found in handler")
	if code != 0 { t.Fatalf("want exit 0, got %d: %s", code, out) }
	...
}
```

`validate.commands` already runs `go test ./tests/acceptance/...` and
`./scripts/check-coverage.sh` — no change needed.

## Deferred Findings

- **Spec wording vs. implementation (documented in edge-cases.md):** the
  *Internal feature code-step* scenario's literal line "senior-engineer.json is
  invalidated" cannot hold for a `--to code` rewind — `code` is the target and
  its evidence is deliberately preserved (the happy-path invariant). The
  acceptance test asserts the true behavior (current=code, ux-ui never
  referenced, downstream qa-senior shed); the code-step ux-ui exclusion is
  pinned at unit level. Recommend a spec copy-edit, non-blocking.

- **Stale cross-feature tripwire fixed (required to green the suite):**
  `tests/acceptance/coverage_hardening_test.go::TestNoBehaviourChange_OnlyTestFilesAdded`
  asserted `git diff main...HEAD` adds only `_test.go` files. That premise held
  only for the already-merged coverage-hardening feature, but the unscoped
  `main...HEAD` comparison made it fail for EVERY later feature that adds
  production code — here the senior-engineer's 5 new files (`rewind.go`,
  `invalidate.go`, etc.). Fixed by scoping it to its own branch: it now skips
  when the coverage-hardening sentinel test file is absent from the added set
  (i.e. once merged into main). File stays ≤100 lines; the guard is unchanged on
  the coverage-hardening branch itself. This was a pre-existing test-design bug
  surfaced by the full acceptance run, not a regression in this feature.

## Handoff → validation-specialist

Build + `go vet` clean; all unit/integration/acceptance suites pass; total
coverage 97.4% (≥95 gate, ≥97 aim met); spec-traceability 0 uncovered for this
feature; qa-senior evidence validates. Ready for the validate step: run the
gatekeeper + production-readiness subagents and the full `centinela validate`.
