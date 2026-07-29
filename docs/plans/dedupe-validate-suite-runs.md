# Plan: dedupe-validate-suite-runs

Remove redundant full-suite executions from the validate-step cycle without
weakening any gate. The design is DECIDED in
`docs/features/dedupe-validate-suite-runs.md` (5 levers, trust invariants,
non-goals); this plan sequences it into commits that each keep
`centinela validate` green. The full Go suite costs ~200s (uncacheable
acceptance tier); one cycle currently runs it ~7 times.

## Run-count budget (the contract this plan ships)

| Surface | Before | After | Mechanism |
|---|---|---|---|
| `centinela validate` | 2 | 1 | Lever A: profiled run + script reuse |
| validate-step `centinela complete` | 4 | 1 | A (2→1 per validation) + B (claim verification reuses the gate's run) |
| gatekeeper session | 3 | 1 | A + C (conditional bare-suite mandate) |
| full validate-step cycle | ~7 | 2 | gatekeeper's own `centinela validate` + `complete`'s machine-side run |
| CI per push | 3 | 1 | A + D (drop bare suite step) |

## Lever A — one profiled run serves tests-pass AND coverage

`centinela.toml`:

```toml
[validate]
commands = [
  "go test ./... -coverprofile=coverage.out",
  "COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh",
  "./scripts/check-fmt.sh"
]
```

`scripts/check-coverage.sh` gains a reuse branch: run the internal
`go test ./... -coverprofile` ONLY when `COVERAGE_PROFILE` is not explicitly
set OR the named profile file does not exist:

```sh
THRESHOLD="${MIN_COVERAGE:-95.0}"
PROFILE="${COVERAGE_PROFILE:-coverage.out}"

if [ -n "${COVERAGE_VALUE:-}" ]; then
  TOTAL_PCT="$COVERAGE_VALUE"
else
  if [ -z "${COVERAGE_PROFILE:-}" ] || [ ! -f "$PROFILE" ]; then
    go test ./... -coverprofile="$PROFILE" >/tmp/centinela-coverage.log
  fi
  TOTAL_LINE="$(go tool cover -func="$PROFILE" | awk '/^total:/ {print $3}')"
  TOTAL_PCT="${TOTAL_LINE%%%}"
fi
```

- **Fail-safe:** `COVERAGE_PROFILE` set but file missing → the script still
  runs the suite itself (never "skip coverage").
- **Bare invocation unchanged:** no `COVERAGE_PROFILE` in the environment →
  self-contained run, exactly today's behavior, even if a stale `coverage.out`
  sits on disk. Scaffolded projects are untouched (non-goal; the scaffold
  ships no `check-coverage.sh` — only this repo's copy changes).
- **Untouched literals (guarded by existing tests, must keep passing
  verbatim):** `MIN_COVERAGE:-95.0`
  (`tests/integration/coverage_hardening_integration_test.go`,
  `tests/acceptance/coverage_hardening_test.go`) and the `COVERAGE_VALUE`
  fast path (`tests/unit/coverage_gate_script_test.go`,
  `tests/acceptance/enforce_coverage_in_validate_test.go`).
- `coverage.out` is already gitignored (`git check-ignore coverage.out`
  passes) and the freshness digest is porcelain/diff-based
  (`internal/treestate/digest.go`), so the profile write cannot stale a
  gatekeeper stamp.

### Masking analysis (why a stale profile cannot hide a failure)

`runValidateCommands` runs the commands in order and aggregates `allPassed`.
The profile the script reads is written by command 1 of the SAME validate
run. The only path where the script could read a profile NOT written by a
just-passed suite run is when command 1 already failed — and then validate
fails regardless of what the coverage script prints. An agent-doctored
`coverage.out` is overwritten by command 1 before the script reads it
(`go test -coverprofile` always rewrites the file). The reuse branch
requires `COVERAGE_PROFILE` to be explicitly set, so no ambient stale file
ever activates it.

## Lever B — wire `PriorTestRun` in the complete gate

`internal/verify/verify.go` already defines `Deps.PriorTestRun` and
`checkTestsPass` consumes it (unit-proven in
`internal/verify/claim_tests_test.go` — runner suppressed when set; that
test keeps passing verbatim). Production wiring:

1. `cmd/centinela/complete_verify.go`: `runClaimVerification` gains a
   `prior *verify.RunOutcome` parameter; when non-nil it is set as
   `deps.PriorTestRun` on the deps built by `verifyDepsFor` (which stays
   UNTOUCHED — `centinela verify`, `verdict`, and MCP keep re-running for
   real, always passing nil/omitting).
2. New helper `completedValidationOutcome() *verify.RunOutcome` returning
   `&verify.RunOutcome{ExitCode: 0}` — the synthesized machine record of the
   `executeValidation()` run that just succeeded in the same process.
3. `cmd/centinela/complete_validate_gates.go` — `runValidateGates` becomes:
   `VerificationFresh` → `executeValidation()` → **re-check
   `VerificationFresh`** (closes the race where the tree changes while the
   ~200s suite runs — a mutated tree blocks completion even though the suite
   passed on the pre-mutation tree) → `runClaimVerification(..., completedValidationOutcome())`.

Trust argument: the reused outcome is synthesized ONLY after
`executeValidation()` exited 0 in the SAME process, seconds earlier, at a
tree re-verified fresh. No on-disk, agent-writable state is involved
(cross-process reuse is the deferred non-goal).

Breaking-test update: `cmd/centinela/verify_gate_test.go`'s two
`runClaimVerification("feat", "validate", "", cfg)` calls gain a `nil` final
arg — semantics unchanged (nil = real re-run), so both tests still exercise
the hard-block and honest-pass paths.

## Lever C — gatekeeper prompt: conditional bare-suite mandate

`docs/architecture/gatekeeper-prompt.md` Mandatory Execution item 2 becomes
conditional: run the bare project suite ONLY when `[validate] commands` does
not already execute the full suite; when it does (as in this repo), the
single `centinela validate` run IS the suite run — do not run it again.
Everything else (fresh context, run-it-yourself, verbatim commands record,
`artifact stamp` last, fail-closed clause) is untouched. Scaffolded projects
with empty `validate.commands` keep the bare-suite obligation (non-goal).

BOTH copies byte-identical:
`docs/architecture/gatekeeper-prompt.md` and
`internal/scaffold/assets/docs/architecture/gatekeeper-prompt.md`. The
existing parity acceptance test
(`tests/acceptance/scaffold_arch_parity_acceptance_test.go`) covers this
file (it is NOT allowlisted) and enforces byte-parity.

## Lever F — qa-senior prompt: one full run to close the tests step

`docs/architecture/qa-senior-prompt.md` (+ scaffold mirror, lockstep, same
parity test) gains working-loop guidance: while iterating, run only affected
packages; before handing off the tests step, run exactly ONE full profiled
run + coverage check
(`go test ./... -coverprofile=coverage.out` then
`COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh`). The tests step
mandates no execution — this is guidance, not a gate change.

## Lever D — CI: drop the bare suite step

`.github/workflows/validate.yml`: delete the `Run test suite` step;
`go run ./cmd/centinela validate` is the single suite run (after Lever A it
runs the profiled suite once + reuses the profile).

Breaking-test update:
`tests/integration/ci_validate_workflow_integration_test.go` — drop
`"go test ./..."` from the required-substring list, keep
`"go run ./cmd/centinela validate"`, and ADD a negative assertion that no
`Run test suite` step remains (the single-run property).

## Breaking-test updates (summary)

| Test | Break | Fix |
|---|---|---|
| `tests/integration/coverage_validate_config_test.go` | exact match `c == "./scripts/check-coverage.sh"` fails once the entry carries the env prefix | `strings.Contains` + a **profile-pairing assertion**: the commands include `go test ./... -coverprofile=<p>` AND the coverage entry sets `COVERAGE_PROFILE=<p>` with the SAME `<p>` |
| `tests/integration/ci_validate_workflow_integration_test.go` | requires `go test ./...` in validate.yml | see Lever D |
| `cmd/centinela/verify_gate_test.go` | `runClaimVerification` arity | add `nil` prior arg |

## New tests

- **Script reuse branch (temp Go module)** — acceptance,
  `tests/acceptance/dedupe_validate_suite_runs_*_test.go` (≤100 lines each,
  `// Acceptance: specs/dedupe-validate-suite-runs.feature` +
  `// Scenario:` comments): scaffold a tiny throwaway Go module, generate a
  real profile with one passing test, then add a FAILING test file; with
  `COVERAGE_PROFILE` pointing at the existing profile the script must pass
  (proves the internal `go test` was skipped — a re-run would fail under
  `set -e`). Companion scenarios: profile missing → script runs the suite
  itself (fail-safe); bare invocation → self-contained even with a stale
  profile on disk.
- **`completedValidationOutcome()`** — colocated
  `cmd/centinela/*_test.go` (≤100 lines, G1 applies): non-nil, `ExitCode`
  0, no `StartErr`/`TimedOut` — i.e. `classifyTestRun` maps it to PASS.
- **Prompt wording** — acceptance assertions that BOTH gatekeeper-prompt
  copies carry the conditional mandate (and byte-parity via the existing
  parity test), and both qa-senior copies carry the one-full-run guidance.
- **toml pairing** — the coverage_validate_config profile-pairing assertion
  above (integration) plus an acceptance assertion that `validate.commands`
  contains exactly one full-suite entry.

Guard rails that must keep passing VERBATIM (no edits):
`MIN_COVERAGE:-95.0` literal tests, `COVERAGE_VALUE` fast-path tests,
`internal/verify/claim_tests_test.go` PriorTestRun suppression test.

## Commit sequencing (each commit leaves `centinela validate` green)

1. **A** — `feat:` script reuse branch + `centinela.toml` profiled pair +
   `coverage_validate_config_test.go` update + new script/pairing tests.
2. **B** — `feat:` `runClaimVerification` prior param +
   `completedValidationOutcome()` + post-run `VerificationFresh` re-check in
   `runValidateGates` + `verify_gate_test.go` nil arg + colocated test.
3. **C** — `docs:` gatekeeper prompt conditional mandate, both copies +
   wording acceptance test.
4. **F** — `docs:` qa-senior prompt one-full-run guidance, both copies +
   wording acceptance test.
5. **D** — `chore:` drop CI bare suite step +
   `ci_validate_workflow_integration_test.go` update.

Order matters: A first (C's conditional text and D's single-run CI both
rely on `validate.commands` running the full suite once); B is independent
of C/F/D but lands before the prompts advertise reduced run counts; D last
so CI never loses its bare run before validate itself covers it.

## Trust invariants (verbatim from the brief — every commit preserves them)

- Gatekeeper still executes `centinela validate` itself, fresh context,
  stamps revision+treeDigest; `VerificationFresh` still refuses stale stamps.
- `complete` still executes the full suite once, machine-side, at the tree
  being completed; claim verification reuses the machine's OWN
  just-completed run, never agent-written text.
- Coverage measured on every validate from a profile written by that same
  run; 95.0 floor and `MIN_COVERAGE:-95.0` literals untouched.
- No gate lowered, skipped, or made conditional on agent claims.

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| Stale profile masks a test failure | High | Impossible by ordering: reuse only reads a profile after command 1 of the same run; command-1 failure fails validate regardless (see Masking analysis) |
| Tree mutates during the ~200s gate suite run, reused outcome vouches for the wrong tree | High | Post-run `VerificationFresh` re-check in `runValidateGates` before any reuse |
| `verify`/`verdict`/MCP accidentally inherit the reuse | Medium | `verifyDepsFor` untouched; only `runValidateGates` passes a non-nil prior; colocated test pins nil-path behavior |
| Prompt copies drift | Medium | Existing scaffold parity test covers both prompts (not allowlisted); wording tests pin the new clauses |
| POSIX-sh portability of the reuse branch (`[ -n ]`/`[ -f ]`) | Low | Plain `sh` test constructs; acceptance tests execute the script with `sh -c` |
| `verify_timeout=360` becomes misleading | Low | It still bounds real `centinela verify`/`verdict` re-runs; leave as-is, note in docs step |

## Deferred (pre-agreed)

`cross-process-suite-result-reuse` — complete trusting a digest-keyed record
of the gatekeeper's run; any on-disk record is writable by the agent it
would excuse. Recorded via `centinela roadmap defer`.
