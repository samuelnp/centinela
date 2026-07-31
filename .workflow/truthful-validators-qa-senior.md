# truthful-validators — qa-senior
**Date:** 2026-07-30

## Predecessor Audit

A prior qa-senior session died mid-step (session limit). Its handoff claimed
"~12 files matching truthful* in tests/acceptance/" — that claim was checked
against the actual worktree and found to be a false lead: the only files
matching `truthful*` were `merge_truthful_delivery_*` (12 files), which
belong to the already-merged, unrelated `merge-truthful-delivery` feature
(commit `ddf222e`, `git status` clean, nothing uncommitted). No file for
*this* feature (`truthful-validators`) existed anywhere under `tests/` before
this session. Verdict: **no predecessor work to audit — the tests step
starts from zero acceptance coverage.** `go build ./...` and `go vet ./...`
on the pre-existing code-step output were both clean, confirming the code
step (unit-tested by senior-engineer) was a solid foundation to build on.

## Test Inventory

| Tier | File | Scenarios |
|------|------|-----------|
| acceptance | tests/acceptance/truthful_validators_evidence_test.go | Section A (5): UX-output rule via CLI, config error, no-config default, empty ui_paths fallback |
| acceptance | tests/acceptance/truthful_validators_skip_basic_test.go | Section B (4): cucumber skip fail, godog undefined fail, 0-scenario fail, go-test-json test-level skip fail |
| acceptance | tests/acceptance/truthful_validators_skip_edge_test.go | Section B (5): package-level no-test-files not a skip, all-passed no-warning, unparseable warn, embedded-substring ignored, exit-code wins |
| acceptance | tests/acceptance/truthful_validators_skip_more_test.go | Section B (4): truncated/killed report, non-acceptance command untouched, empty commands unaffected, classifier delegation pin |
| acceptance | tests/acceptance/truthful_validators_policy_test.go | Section C (4): absent=fail, warn, off, unknown=config error |
| acceptance | tests/acceptance/truthful_validators_verify_test.go | Section D (2): verify rejects skip claim, verify does not invent a failure |
| acceptance | tests/acceptance/truthful_validators_verify_priorrun_test.go | Section D (1): PriorTestRun wiring structural pin |
| acceptance | tests/acceptance/truthful_validators_quality_test.go | Section E (4): missing scores, wrong-type scores, missing field, non-integer score |
| acceptance | tests/acceptance/truthful_validators_quality_more_test.go | Section E (3): out-of-range range error, missing features structural, well-formed pass |
| acceptance | tests/acceptance/truthful_validators_gates_filesize_test.go | Section F (4): G1 empty-scope skip, clean pass, cap/severity unchanged, justified exception |
| acceptance | tests/acceptance/truthful_validators_gates_i18n_test.go | Section F (3): G11 single-locale warn, missing-file fail, malformed-file fail |
| acceptance | tests/acceptance/truthful_validators_gates_i18n_more_test.go + two-locale case in _test.go | Section F (4): two locales sync/out-of-sync, gettext unaffected, filtered-out skip, skip/warn-only run stays green |
| unit (code step, audited not rewritten) | internal/acceptance/*_test.go, internal/config/acceptance_skip_policy_test.go, internal/roadmap/quality_*_test.go, internal/gates/file_size_truthful_test.go, internal/gates/i18n_single_locale_test.go, internal/verify/claim_tests_acceptance_test.go, internal/ui/render_cmd_verdict_test.go, cmd/centinela/validate_commands_test.go, cmd/centinela/evidence_validate_*_test.go | all 4 slices, both directions, per senior-engineer's own report |

All 43 spec scenarios have a `// Acceptance:` / `// Scenario:` marker in
`tests/acceptance/` with matching normalized text; `centinela validate`'s own
spec-traceability gate confirms: **"All 43 scenarios have acceptance
coverage."**

## Coverage Gaps

None outstanding at the scenario level — all 43 `.feature` scenarios have an
executing acceptance-tier marker. Two items were deliberately NOT driven as
full binary end-to-end acceptance runs, and are recorded as residual risks
(with the honest substitute test named) rather than silently marked done:

- The `PriorTestRun` reuse path (`centinela complete` at the validate step)
  — needs a real git tree + `workflow.VerificationFresh` + a genuine
  `executeValidation()` run; the same tradeoff this repo already made for
  `dedupe_validate_suite_runs_complete_test.go`. Covered behaviorally at unit
  tier (already real, pre-existing) and pinned structurally at acceptance
  tier (`TestTV_Verify_PriorRunWiringReachesTheAcceptanceRule`).
- Parser breadth beyond go/cucumber/godog (behave, pytest-bdd, RSpec, Jest,
  Playwright) — explicitly out of v1 scope per the plan (§7); not this
  step's gap to close.

See `.workflow/truthful-validators-edge-cases.md` for the full negative-path
inventory with `file:test` references.

## Acceptance Wiring

`centinela.toml` `[validate].commands` — **unchanged by this step**, exactly
as landed by senior-engineer:
```toml
[validate]
commands = [
  "go test ./... -coverprofile=coverage.out",
  "COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh",
  "./scripts/check-fmt.sh"
]
```
Command 1 (`go test ./... -coverprofile=coverage.out`) is acceptance-classified
(`go test` + `./...`) and covers `tests/acceptance/**`, satisfying the tests
step's "validate.commands includes acceptance execution" requirement. Its own
non-verbose output carries no skip data — per the senior-engineer's recorded
divergence (R4) this renders as a quiet Pass-with-note, not a permanent ⚠;
confirmed live in this session's own `centinela validate` run.

## Verification

- `go build ./...` — clean.
- `go vet ./...` — no issues.
- `go test ./... -run xxxNONE` — every package (incl. `tests/acceptance`,
  `tests/unit`, `tests/integration`) compiles; `[no tests to run]` on all.
- `rtk go test ./tests/acceptance/ -run 'TestTV_'` — **43 passed** in 1
  package.
- `rtk go test ./...` — see final run result appended below.
- `./scripts/check-fmt.sh` — clean (one `gofmt -w` fix applied during
  authoring to `truthful_validators_verify_test.go`, a trailing blank line).
- `centinela validate` (this repo, background run) —
  `✓ G1: File Size  All in-scope files are within the 100-line cap ...`;
  `✓ spec-traceability-gate  All 43 scenarios have acceptance coverage.`;
  full result appended below once the profiled suite run completes.
- `centinela evidence validate truthful-validators` — `evidence ok`.

## Deferred Findings

None new. All residual gaps (parser breadth, self-validate's own undetermined
command 1, `PriorTestRun` E2E cost, verdict/pr-gate machine output) were
already recorded by the planner or the senior-engineer and are re-affirmed,
not re-deferred, in `.workflow/truthful-validators-edge-cases.md`.

## Handoff

Next role: validation-specialist
Edge-case report: `.workflow/truthful-validators-edge-cases.md` (filled this
step, per the qa-senior prompt's authoring rule for the mandatory companion
artifact).
