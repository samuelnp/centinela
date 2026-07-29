# Feature: dedupe-validate-suite-runs

- surface: internal
- status: planned
- fixes: a validate-step cycle executes the full test suite ~7 times; velocity loss with zero added assurance

## Problem

The full Go suite costs ~200s per run (the acceptance tier is uncacheable), and
one validate-step cycle runs it ~7 times without any run adding assurance the
previous one did not already provide:

- `centinela validate` runs it twice — `go test ./...` plus
  `./scripts/check-coverage.sh`, which internally re-runs the whole suite with
  `-coverprofile` (different flags ⇒ different Go build/test cache key).
- Validate-step `centinela complete` runs `executeValidation()` (all
  `validate.commands`) and then claim verification, whose `checkTestsPass`
  re-runs every `validate.commands` entry again — 4 suite executions in one
  `complete`.
- The gatekeeper prompt mandates the agent run `centinela validate` PLUS a bare
  `go test ./...`, even though `validate.commands` already contains the suite —
  3 more executions.
- CI runs a bare suite step and then `centinela validate` — 3 per push.

`internal/verify/verify.go` already defines `Deps.PriorTestRun` ("reused by the
tests-pass check instead of re-running the suite — the complete gate already
ran it once"), consumed by `checkTestsPass` with a unit test proving the runner
is suppressed — but no production code ever populates it. The dedup seam was
built and never wired.

## Goal

Cut the cycle to **2 mandated suite runs** (gatekeeper's own `centinela
validate` + `complete`'s machine-side run) and CI to 1, by removing only
redundant executions:

1. **One profiled run serves tests-pass AND coverage.** `validate.commands`
   becomes `go test ./... -coverprofile=coverage.out` followed by
   `COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh`; the script
   gains a reuse branch (profile explicitly set + file exists → skip its
   internal `go test`). Fail-safe: profile set but missing → the script still
   runs the suite itself. Bare invocation unchanged.
2. **Wire `PriorTestRun` in the complete gate.** After `executeValidation()`
   succeeds, re-check `workflow.VerificationFresh` (closes the tree-change
   race), then pass a synthesized `RunOutcome{ExitCode: 0}` into
   `runClaimVerification` — a true machine record from the same process,
   seconds earlier, at the verified tree. `verifyDepsFor` stays untouched so
   `centinela verify`/`verdict`/MCP keep re-running for real.
3. **Gatekeeper prompt (both copies, byte-identical):** the bare-suite mandate
   becomes conditional — when `[validate] commands` already runs the full
   suite, the single `centinela validate` run IS the suite run.
4. **qa-senior prompt (both copies):** iterate on affected packages, then
   exactly one full profiled run + coverage check before closing the tests
   step (the tests step mandates no execution; this is working-loop guidance).
5. **CI:** drop the bare "Run test suite" step; `go run ./cmd/centinela
   validate` is the single suite run.

## Trust invariants (must hold, verbatim)

- The gatekeeper still executes `centinela validate` itself, in a fresh
  context, and stamps revision+treeDigest; `VerificationFresh` still refuses
  stale stamps.
- `centinela complete` still executes the full suite once, machine-side, at
  the tree being completed — claim verification reuses the machine's OWN
  just-completed run, never agent-written text.
- Coverage is measured on every validate from a profile written by that same
  run; the 95.0 floor and its `MIN_COVERAGE:-95.0` literal guards are
  untouched.
- No gate is lowered, skipped, or made conditional on agent claims.

## Non-goals (v1)

- **Cross-process suite-result reuse** (complete trusting a digest-keyed
  record of the gatekeeper's run). Any on-disk record is writable by the agent
  it would excuse; deferred as `cross-process-suite-result-reuse` until a
  tamper-evident record design exists.
- **Making the acceptance tier cacheable/faster.** Orthogonal velocity work.
- **Changing scaffold defaults for new projects.** The scaffold-shipped
  check-coverage.sh keeps its self-contained behavior; scaffolded projects
  with empty `validate.commands` keep the gatekeeper's bare-suite obligation.
