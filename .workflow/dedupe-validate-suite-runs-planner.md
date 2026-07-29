# dedupe-validate-suite-runs — planner

### Planner Report: dedupe-validate-suite-runs
**Date:** 2026-07-29

## Problem

Every validate-step cycle on this repo executes the full Go suite ~7 times at
~200s per run (the acceptance tier is uncacheable), with zero added assurance
after the first run per surface: `centinela validate` runs it twice (bare
`go test ./...` + `check-coverage.sh`'s internal `-coverprofile` re-run, which
misses the build cache), a validate-step `centinela complete` runs it four
times (executeValidation's two + claim verification re-running every
validate.commands entry), the gatekeeper prompt mandates `centinela validate`
PLUS a bare suite (three in that session), and CI runs a bare step plus
validate (three per push). The dedup seam already exists —
`internal/verify/verify.go` defines `Deps.PriorTestRun`, `checkTestsPass`
consumes it, and a unit test proves the runner is suppressed — but no
production code populates it. The people hurting are the operator (wall-clock:
validate-step completes routinely brush the ~10-minute job cap; see the
validate-complete wall-clock memory) and every agent session burning its
budget on redundant 200s waits. Why now: the acceptance tier grows with each
feature, so the waste compounds monotonically.

## Scope

- **In:** (A) `centinela.toml` profiled-run pairing + `check-coverage.sh`
  reuse branch with fail-safe fallback; (B) wiring `PriorTestRun` in
  `runValidateGates` via a new `runClaimVerification` prior param +
  `completedValidationOutcome()` helper + post-run `VerificationFresh`
  re-check; (C) gatekeeper prompt conditional bare-suite mandate, both copies
  byte-identical; (F) qa-senior prompt one-full-run guidance, both copies;
  (D) CI drops the bare "Run test suite" step. Plus the three breaking-test
  updates and the new tests listed in the plan.
- **Out:** cross-process suite-result reuse (deferred, pre-agreed:
  `cross-process-suite-result-reuse`); making the acceptance tier
  cacheable/faster; changing scaffold defaults for new projects (the
  scaffold ships no check-coverage.sh; scaffolded projects with empty
  `validate.commands` keep the gatekeeper's bare-suite obligation).

## Dependencies & Assumptions

- `Deps.PriorTestRun` + `checkTestsPass` consumption already shipped
  (claim-verification feature); `internal/verify/claim_tests_test.go` proves
  runner suppression — that test must keep passing verbatim.
- `runClaimVerification` has exactly one production caller
  (`runValidateGates` in `cmd/centinela/complete_validate_gates.go`);
  `verifyDepsFor` is shared by `verify`/`verdict`/MCP/complete and stays
  untouched.
- `workflow.VerificationFresh` is cheap and re-runnable (porcelain/diff
  digest via `internal/treestate`); `coverage.out` is already gitignored, so
  profile writes cannot stale a gatekeeper stamp (verified with
  `git check-ignore`).
- `tests/acceptance/scaffold_arch_parity_acceptance_test.go` already
  byte-compares gatekeeper-prompt.md and qa-senior-prompt.md against their
  scaffold mirrors (neither is allowlisted) — it is the parity enforcement
  for levers C and F.
- `go test -coverprofile` always rewrites the profile file, so a doctored
  profile cannot survive command 1 of a validate run.
- CI sets `CI=true` → full-scan validate; dropping the bare step leaves
  `go run ./cmd/centinela validate` as the single suite execution.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Stale profile masks a test failure | High | Low | Ordering makes it impossible: reuse reads only a profile written by command 1 of the same run; if command 1 failed, validate fails regardless (masking analysis in plan) |
| Tree mutates during the ~200s gate run; reused outcome vouches for the wrong tree | High | Medium | Post-run `VerificationFresh` re-check in `runValidateGates` before synthesizing the prior outcome |
| `verify`/`verdict`/MCP silently inherit reuse (false green) | Medium | Low | `verifyDepsFor` untouched; only `runValidateGates` passes non-nil; scenario + colocated test pin the nil path |
| Prompt copies drift after edit | Medium | Medium | Existing scaffold parity test fails on any byte drift; wording acceptance tests pin the new clauses |
| Breaking-test updates weaken guards (exact→contains) | Medium | Low | The contains-relaxation is paired with a STRONGER profile-pairing assertion; MIN_COVERAGE/COVERAGE_VALUE guard tests untouched |
| POSIX-sh portability of the reuse branch | Low | Low | Plain `[ -n ]`/`[ -f ]` constructs; acceptance tests execute the real script under `sh` |
| Regression of earlier features (coverage-hardening, claim-verification, adversarial-verifier) | High | Low | Their guard tests are enumerated as must-pass-verbatim; commit sequencing keeps validate green at every step |

## Rollout

- Step 1 (commit A): script reuse branch + toml profiled pair + updated
  `coverage_validate_config_test.go` (contains + pairing) + new script tests
  — validate immediately drops 2→1.
- Step 2 (commit B): `runClaimVerification` prior param +
  `completedValidationOutcome()` + freshness re-check + `verify_gate_test.go`
  nil arg + colocated helper test — complete drops 4→1.
- Step 3 (commit C): gatekeeper prompt conditional mandate, both copies +
  wording test — gatekeeper session drops 3→1.
- Step 4 (commit F): qa-senior prompt one-full-run guidance, both copies +
  wording test.
- Step 5 (commit D): CI drops the bare suite step + updated CI integration
  test — CI drops 3→1. D lands last so CI never loses its bare run before
  validate itself covers it; A lands first because C and D depend on
  `validate.commands` running the full suite once.

## Behavior Summary

After this feature, one `centinela validate` executes the full suite exactly
once (`go test ./... -coverprofile=coverage.out`) and the coverage gate reads
that same run's profile; a validate-step `centinela complete` runs validation
once, re-checks tree freshness, and lets claim verification consume the
synthesized outcome of its own just-completed run instead of re-running the
suite; the gatekeeper runs `centinela validate` once and (because
validate.commands runs the full suite) owes no bare suite run; CI runs the
suite once via `centinela validate`. Standalone `centinela verify`,
`verdict`, and the MCP surface still re-run everything for real, the
`COVERAGE_VALUE`/`MIN_COVERAGE` contracts are byte-identical, bare
`check-coverage.sh` stays self-contained, and every trust invariant from the
brief holds verbatim. Net: cycle ~7→2 suite runs, CI 3→1.

## Acceptance Criteria (Gherkin)

Full scenarios in `specs/dedupe-validate-suite-runs.feature` (12 scenarios;
acceptance tests reference them via `// Scenario:` comments):

- **validate runs the suite exactly once** — Given the toml pairs the
  profiled run with `COVERAGE_PROFILE=coverage.out`, Then exactly one entry
  executes the suite (integration pairing assertion + acceptance).
- **check-coverage.sh reuses an existing profile when COVERAGE_PROFILE is
  set** — Given a temp Go module whose profile predates a newly added
  FAILING test, When the script runs with the profile set, Then it exits 0 —
  proof the internal `go test` was skipped (a re-run would fail).
- **check-coverage.sh falls back to running the suite when the profile is
  missing** — fail-safe: profile named but absent → the script runs the
  suite itself and writes it.
- **bare check-coverage.sh invocation stays self-contained** — no env var →
  full self-run even with a stale profile on disk; 95.0 default floor.
- **COVERAGE_VALUE fast path bypasses the suite unchanged** — existing guard
  behavior, re-pinned.
- **complete's claim verification reuses the gate's own run** — synthesized
  `ExitCode: 0` outcome consumed; no second suite execution in the same
  complete.
- **a failing validate command blocks completion before any reuse** —
  negative path: executeValidation failure blocks; nothing synthesized.
- **a tree change during the gate's suite run blocks completion** — negative
  path: post-run `VerificationFresh` refuses before claim verification.
- **standalone verify surfaces still re-run the suite for real** —
  verify/verdict/MCP get nil `PriorTestRun`.
- **gatekeeper prompt copies stay identical and carry the conditional suite
  mandate** — byte-parity + wording.
- **qa-senior prompt copies carry the one-full-run guidance** — byte-parity
  + wording.
- **CI runs the suite once via centinela validate** — no bare "Run test
  suite" step; `go run ./cmd/centinela validate` remains.

## UX States

| State   | Trigger | Surface |
|---------|---------|---------|
| success | validate/complete pass with one suite run | existing CLI output (unchanged rendering; faster wall-clock) |
| error   | failing validate command | existing red gate output; completion blocked before reuse |
| error   | tree mutated during gate run | `VerificationFresh` stale-verification message at complete |
| loading / empty | n/a | no new UI surface (CLI internals + prompts + CI only) |

## Edge Cases

- `COVERAGE_PROFILE` set but the file missing → fail-safe self-run, never a
  skipped coverage measurement.
- Bare script invocation with a stale `coverage.out` on disk → still
  self-contained (reuse requires the env var explicitly).
- Command 1 fails but writes a partial/stale-readable profile → validate
  already fails; the coverage read cannot mask it.
- Tree mutated during the ~200s suite run → post-run freshness re-check
  blocks before any reuse.
- Nil prior on every non-complete surface (`verify`/`verdict`/MCP) → real
  re-run preserved.
- Doctored `coverage.out` pre-planted by an agent → overwritten by command 1
  before the script reads it.
- Scaffolded project with empty `validate.commands` → gatekeeper's
  conditional keeps the bare-suite obligation.
- `verify_timeout=360` still bounds single real verification commands on the
  standalone surfaces; unchanged.

## Out-of-Scope

- Cross-process suite-result reuse (complete trusting a record of the
  gatekeeper's run) — deferred as `cross-process-suite-result-reuse`; any
  on-disk record is writable by the agent it would excuse.
- Making the acceptance tier cacheable or faster.
- Changing scaffold defaults for new projects (scaffold check-coverage
  behavior, empty-commands bare-suite obligation).
- Lowering, skipping, or conditioning any gate on agent claims.

## Deferred Findings

- `cross-process-suite-result-reuse` — pre-agreed deferral from the brief,
  recorded via `centinela roadmap defer cross-process-suite-result-reuse
  --summary "..." --source dedupe-validate-suite-runs/planner`.
- No NEW out-of-scope gaps discovered during planning: none.

## Handoff

- Next role: senior-engineer. Implement in the commit order A→B→C→F→D from
  `docs/plans/dedupe-validate-suite-runs.md`; each commit must leave
  `centinela validate` green.
- Clarifications resolved during planning: `coverage.out` is already
  gitignored (no digest-staleness risk); `runClaimVerification` has exactly
  one production caller, so the prior param is additive; both prompt files
  are already under scaffold byte-parity enforcement.
- Outstanding for implementation: keep every new `_test.go` under `cmd/` and
  `internal/` ≤100 lines (G1 applies to tests); do not touch
  `verifyDepsFor`, the `MIN_COVERAGE:-95.0` literal, the `COVERAGE_VALUE`
  fast path, or `claim_tests_test.go`.
