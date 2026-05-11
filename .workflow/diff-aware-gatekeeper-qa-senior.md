# Orchestration Evidence: qa-senior

- Feature: `diff-aware-gatekeeper`
- Step: `tests`
- Outcome: Built the three test layers (unit, integration via
  package-internal tests, acceptance) plus the edge-cases catalog.
  Verified the feature against every Gherkin scenario in
  `specs/diff-aware-gatekeeper.feature` plus the edge cases
  enumerated in the feature brief and surfaced during code.

  - Unit: `internal/gitdiff/set_test.go`,
    `internal/gitdiff/resolver_test.go` (eight cases including all
    degrade paths via a stubbed runner; one case exercising the real
    `runGit` error wrapping for a missing binary).
  - Unit: `internal/config/validate_mode_test.go` covering the
    normalization helpers and every cell of the resolution truth
    table (auto/always/off × CI/local × FlagNone/Changed/Full).
  - Unit: `internal/gates/diff_aware_test.go` covering G1 with a
    nil-vs-empty-vs-restrictive filter and the G11 short-circuit
    both ways.
  - Acceptance: `tests/acceptance/diff_aware_gatekeeper_acceptance_test.go`
    builds the centinela binary, spins up a tmp git repo per
    scenario, and asserts 10 observable behaviors end-to-end:
    local-default diff-aware, CI-default full, branch violation
    flagged, untracked file flagged, historical-violation full
    coverage, non-git degrade, mutex flag rejection, missing-base
    degrade, user-commands always run, configurable diff base.
  - Edge cases: `.workflow/diff-aware-gatekeeper-edge-cases.md`
    cross-references each edge case to the test that proves it.

  Full suite green (`go test ./...`). Coverage gate
  (`./scripts/check-coverage.sh`) reports 95.1% ≥ 95.0% threshold.
  `centinela.toml [validate] commands` already runs `go test ./...`,
  so the new acceptance test is part of `centinela validate` without
  any toml edits.

- Inputs: `specs/diff-aware-gatekeeper.feature`,
  `docs/features/diff-aware-gatekeeper.md`,
  `docs/plans/diff-aware-gatekeeper.md`,
  `.workflow/diff-aware-gatekeeper-senior-engineer.md`,
  `internal/gitdiff/set.go`, `internal/gitdiff/resolver.go`,
  `internal/config/validate_mode.go`,
  `internal/gates/i18n_filter.go`,
  `internal/gates/file_size.go`, `internal/gates/gates.go`,
  `cmd/centinela/validate.go`, `cmd/centinela/validate_mode.go`.
- Outputs:
  `internal/gitdiff/set_test.go`,
  `internal/gitdiff/resolver_test.go`,
  `internal/config/validate_mode_test.go`,
  `internal/gates/diff_aware_test.go`,
  `tests/acceptance/diff_aware_gatekeeper_acceptance_test.go`,
  `.workflow/diff-aware-gatekeeper-edge-cases.md`.
- Handoff: `validation-specialist`
