### Planner Report: truthful-validators
**Date:** 2026-07-30

#### Problem

Centinela's value proposition is that its verdict is load-bearing: when it says
PASS the operator stops looking. Four validators break that contract today, and
all four were re-verified as still live against the current worktree revision
(`1b0af2b`, after the 0.52.x line landed). `centinela evidence validate` passes
`nil` UI paths, so its UX-output rule is *structurally unable to fire* — the same
evidence validates differently depending on whether the CLI or `complete` ran it.
`centinela validate` judges every command on exit code alone and never inspects
the output it already captured, so a Go acceptance test that calls `t.Skip`, a
cucumber run reporting `0 scenarios`, and a godog run with only undefined steps
all print a green check; the claim verifier has the identical defect, so the
independent verifier cannot catch it either. `internal/roadmap/quality.go`
reports "scores must be between 1 and 10" for a *structural* fault, sending the
operator hunting for a number that does not exist. And two gates report more
confidence than they earned: G1 returns Pass when its diff filter was empty, and
G11 returns "All locales have identical keys" after comparing a single locale
against nothing. The user is the operator (and the agent) reading the verdict:
today they cannot distinguish "checked and clean" from "could not check".

#### Scope

- **In:** `evidence validate` loads config and passes `config.UIPaths(cfg)`;
  acceptance-classified validate commands are judged on their run summary with a
  default-on `[validate] acceptance_skip_policy = fail|warn|off`; unparseable
  acceptance reports render as *undetermined* warnings; the claim-verification
  path applies the same rule (including the reused single-run outcome);
  roadmap-quality shape faults are named as shape faults; G1 reports `Skip` when
  it inspected nothing and names its cap when it passes; G11 reports `Warn` for a
  single configured locale and `Skip` when filtered out of the diff scope.
- **Out:** lowering, relaxing or profile-scaling **any** gate default — G1 stays
  a Fail at 100 lines under every enforcement profile, and the advisory-severity
  half of WS5.2 belongs to `guided-by-default`, which the roadmap already declares
  `dependsOn: [truthful-validators]`; parsers for behave, pytest-bdd, RSpec, Jest,
  Playwright, Maven/Gradle; broadening the acceptance classifier's heuristics;
  re-running any suite to obtain a parseable report; changing this repo's own
  `validate.commands`; every other gate's reporting.

#### Dependencies & Assumptions

- No external services, no new dependencies; `internal/acceptance` is a new
  **stdlib-only leaf** package (`classify.go`, `summary.go`, `parse_cucumber.go`,
  `parse_gotest.go`), registered in the `leaf` layer of
  `[[gates.import_graph.layers]]` and mirrored into PROJECT.md G2.
- The classifier predicate is moved, not changed: today's
  `hasAcceptanceExecutionCommand` (`internal/workflow/validate_tests_acceptance_commands.go`)
  becomes a loop over `acceptance.IsExecutionCommand`, and its existing tests must
  pass **unchanged**. Leaf placement is what avoids an `internal/verify →
  internal/workflow` edge.
- The policy knob follows the `enforcement_profile` pattern exactly: validated
  against the **raw** decoded value in `config.Load` before `applyDefaults`
  normalizes it, otherwise an unknown value would be silently swallowed.
- Builds on `adversarial-validate-verifier` (claim verification) and
  `dedupe-validate-suite-runs` (the `Deps.PriorTestRun` single-run reuse wired at
  `cmd/centinela/complete_verify.go:23`) — that reused outcome is labelled
  `"validate.commands"`, not a real command, which is the one place AC6 can be
  silently disarmed.
- Assumes godog's summary line copies cucumber's `N scenarios (…)` shape, so one
  parser serves both.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| R1 self-reference: this feature's validate step runs the runner it changes | High | Certain | Land slice D last; build a `/tmp` binary from `./cmd/centinela` and dry-run `validate` before the gate — the installed binary lags the worktree. A `Warn` on command 1 is expected (R4); a `Fail` is a real defect. |
| R2 false-FAIL of legitimate skips in non-acceptance tiers | High | Medium | Structural, not heuristic: classification gates the parse, the parse gates the verdict. Pinned by an explicit "non-acceptance command reporting skips passes" test. |
| R3 parser breadth is the effort driver | Medium | High | v1 = `go test` (`-json`/`-v`) + the shared cucumber/godog summary line; everything else lands in the honest undetermined bucket, never a false pass. |
| R4 this repo's own `go test ./... -coverprofile` is unparseable | Medium | Certain | Non-verbose Go output carries no skip information, so validate will show one permanent `⚠ could not parse`. Honest, non-failing, and deferred rather than papered over. |
| R5 summary-shaped text inside a test's own stdout | Medium | Medium | Parsers anchor at line start on the runner's summary shape, never `strings.Contains`; explicit false-positive test. |
| R6 G11 `Warn` + `pr_gate.fail_on_warning` blocks single-locale repos | Medium | Low | `fail_on_warning` defaults false and the i18n gate is opt-in; documented in the message and changelog. |
| R7 G1 `Pass → Skip` shifts pr-gate tallies | Low | Medium | `AllPassed` ignores non-Fail and every renderer already handles `Skip`; new counts asserted in tests, not discovered in CI. |
| R8 the `PriorTestRun` path silently disarms AC6 | High | Medium | Classify over `cfg.Validate.Commands` for that path, with a dedicated test. |
| R9 `validateScoreRange` has a second caller (`promote_scores.go:29`) | Low | Medium | Identified; `promote`'s assertions updated deliberately, not loosened. |
| R10 four touched files are within 15 lines of the G1 cap | Medium | High | Each affected slice **opens with a split** (`validate.go` is 99, `quality.go` 87, `file_size.go` 92); `_test.go` counts too. |

#### Rollout

- Step 1: slice A — `evidence validate` loads config and passes `config.UIPaths(cfg)`.
- Step 2: slice B — roadmap-quality two-stage decode; shape faults named as shape faults.
- Step 3: slice C — G1 `Skip`/cap-naming, G11 single-locale `Warn`, filtered-out `Skip`.
- Step 4: slice D — `internal/acceptance` leaf, the policy knob, the per-command
  verdict in the validate runner, and the same rule on the claim-verification path.
- Deliberately later: the enforcement-severity half of WS5.2 (`guided-by-default`).

#### Behavior Summary

After this feature, every one of the four validators reports only what it
actually established. `centinela evidence validate` applies the same UX-output
rule `complete` applies, and a broken `centinela.toml` surfaces as a config error
instead of a silent downgrade. Each validate command classified as an acceptance
execution is judged on its run summary as well as its exit code: detected skipped,
pending or undefined scenarios (or a `0 scenarios` run) fail the run by default,
`warn` and `off` are available, and a report matching no supported shape renders
as an undetermined warning that names the remedy — never green, never a fabricated
failure. Non-acceptance tiers are untouched. `centinela verify` applies the same
rule, including on the reused single-run outcome. A malformed roadmap-quality file
is described by its shape, and the range message is reserved for a genuinely
out-of-range integer. G1 reports `Skip` when its diff scope was empty and names
the 100-line cap and the exception mechanism when it passes — the cap and its
`Fail` severity are unchanged. G11 tells a single-locale project that its parity
check is trivially satisfied and verifies nothing, and reports `Skip` rather than
`Pass` when no locale file was in scope.

#### Gherkin Scenarios

Full spec at [specs/truthful-validators.feature](../specs/truthful-validators.feature)
— 43 scenarios in six sections (A evidence-validate UI paths, B acceptance skip
detection, C the policy knob, D the claim verifier, E quality shape errors, F gate
reporting), each acceptance criterion carrying at least one happy and one negative
path. Representative pair:

- **Happy:** *Given* a validate command classified as an acceptance execution
  *And* the command exits zero and reports `5 scenarios (5 passed)` *When*
  `centinela validate` runs *Then* the command should be reported as passed *And*
  no skip warning should be emitted.
- **Negative:** *Given* a validate command classified as an acceptance execution
  *And* the command exits zero and reports `3 scenarios (1 skipped, 2 passed)`
  *When* `centinela validate` runs *Then* the command should be reported as failed
  *And* the message should name the skipped count and the command *And* the run
  should exit non-zero.
- **Anti-weakening pin:** *Given* a source file of 101 lines with no justified
  exception *When* the file-size gate runs *Then* the gate status should be FAIL
  *And* the effective cap should still be 100 lines.

#### UX States

| State | Trigger | Surface |
|-------|---------|---------|
| loading | n/a — every path is synchronous CLI output | n/a |
| empty | no configured `validate.commands`; empty diff scope; no locale files in scope | no commands executed; `— G1: File Size  No files in the diff scope — nothing inspected.`; `— G11: i18n  No locale files in the diff scope — nothing inspected.` |
| error | acceptance skips detected under the default policy; unparseable `centinela.toml`; malformed roadmap-quality; oversized source file | `✗ <command>` naming counts, shape and command; config parse error on stderr with non-zero exit; a shape error naming feature, field and expected schema; unchanged G1 `Fail` |
| warning | unparseable acceptance report; `acceptance_skip_policy = "warn"`; single configured locale | `⚠ <command>  acceptance report could not be parsed — skips undetected; run with go test -json / -v …`; `⚠ G11: i18n  1 locale configured — the key-parity check is trivially satisfied and verifies nothing.` |
| success | everything executed and nothing skipped | `✓ <command>`; `✓ G1: File Size  All in-scope files are within the 100-line cap (per-file exceptions are configurable …)` |

#### Out-of-Scope

- Lowering, relaxing or profile-scaling any gate default; G1's advisory-severity
  question is `guided-by-default`'s.
- Acceptance-summary parsers for behave, pytest-bdd, RSpec, Jest, Playwright,
  Maven/Gradle.
- Broadening the acceptance classifier's heuristics.
- Re-running any suite to obtain a parseable report (validate wall clock is
  already at the job cap).
- Changing this repo's own `validate.commands` to emit `-json`.
- Spec-traceability, security, build, import-graph and roadmap-drift reporting.

#### Deferred Findings

Checked against `.workflow/roadmap.json` Backlog first — neither overlaps an
existing entry (nearest neighbours `validate-flake-diagnosability`,
`cross-process-suite-result-reuse`, `render-warn-gate-details` are all distinct):

- `acceptance-skip-parser-breadth` — behave is *classified* as an acceptance
  runner but gets no parser in v1 (nor pytest-bdd/RSpec/Jest/Playwright), so those
  projects sit permanently in the undetermined-warning bucket.
- `self-validate-acceptance-report-shape` — Centinela's own
  `go test ./... -coverprofile` emits no skip information, so its own acceptance
  gate is permanently undetermined; choosing `-json`, `-v`, or a separate
  acceptance command interacts with the single-run coverage-profile reuse.

#### Handoff

Next role: **senior-engineer**. No blocking questions. Two decisions are the
planner's and should not be relitigated silently: the `internal/acceptance` leaf
placement (it is what keeps `internal/verify` from importing `internal/workflow`),
and G11 single-locale rendering as `Warn` rather than `Skip`. Read
`docs/plans/truthful-validators.md` §0.2 before touching any file — three of the
four slices must open with a G1 split.
