# Feature Brief — truthful-validators

**Slug:** `truthful-validators`
**Roadmap entry:** Fix nil-uiPaths evidence validate, detect skipped acceptance
scenarios (skips fail validate by default), honest quality-score shape errors,
honest gate reporting for file-size cap and single-locale i18n.
**Archetype:** canonical · 5 steps · plan contract `planner-v1`
**Source of record:** the post-mortem retrospective, work-streams WS1.2, WS1.3,
WS1.5 and WS5.2 (retrospective failures #4 and #6).

## Problem — what pain does this solve? Who is the user?

Centinela's whole value proposition is that its verdict is *load-bearing*: when
it says PASS, the operator stops looking. Four validators currently break that
contract. Each was re-verified against the current tree (revisions have moved
since the retrospective was written); **all four are still live**.

1. **A permanently-red validator.** `cmd/centinela/evidence_validate.go:25`
   calls `evidence.ValidateFeature(feature, nil)`. The `uiPaths` argument feeds
   `orchestration.validateUXOutputs`, which decides whether a role's outputs
   touch a UI surface. With `nil`, `hasAnyPrefix(files, nil)` can never match, so
   the UX-output rule is structurally unable to fire from the CLI. The
   orchestration path does it correctly —
   `internal/workflow/validate_orchestration.go:14` passes
   `config.UIPaths(cfg)` — so the *same* evidence validates differently
   depending on which entry point ran. Retrospective failure #4.

2. **Acceptance skips read as green.** `cmd/centinela/validate.go:92` judges
   every validate command purely on exit code: `passed, out := runCommand(cmd)`,
   and `out` is only ever *rendered*, never *inspected*. A Go acceptance test
   that calls `t.Skip`, a Cucumber run reporting `0 scenarios`, and a godog run
   with only undefined steps all exit 0. `centinela validate` prints a green
   check, `complete` advances, and the ship gate certifies a suite that asserted
   nothing. `internal/verify/claim_tests.go:classifyTestRun` has the identical
   defect on the claim-verification path, so the independent verifier cannot
   catch it either. Retrospective failure #6.

3. **A misleading roadmap-quality error.** `internal/roadmap/quality.go`
   unmarshals into a value-typed `QualityScores`. A feature entry whose `scores`
   object is *missing*, spelled differently, or nested wrongly decodes to all
   zeros, and `validateScoreRange` reports `scores must be between 1 and 10`.
   The operator then hunts for an out-of-range number that does not exist. The
   real fault — a structural one — is never named.

4. **Two gates that report more confidence than they earned.**
   - **G1 (file size):** `checkFileSize` returns `Status: Pass` with
     `"No relevant changes — gate skipped."` when the diff filter is empty. A
     green ✓ for a gate that inspected zero files is a false assurance. The
     100-line cap is also presented as an absolute with no statement of where
     it comes from or that per-file exceptions exist.
   - **G11 (i18n):** `compareKeysets` iterates `locales[1:]`. With a single
     locale that loop body never runs, and the gate returns `Status: Pass,
     "All locales have identical keys."` — a confident claim derived from
     comparing nothing. `checkI18nFiltered` has the same Pass-means-skipped
     problem as G1.

The user is the operator (and the agent) reading a Centinela verdict. Today
they cannot distinguish "checked and clean" from "structurally could not
check". This feature makes that distinction visible everywhere it is currently
collapsed.

## User Stories

- As an operator, I want `centinela evidence validate <feature>` to apply the
  same UX-output rule the workflow applies, so the CLI and `complete` never
  disagree about the same evidence file.
- As an operator, I want a validate run whose acceptance suite skipped
  everything to **fail**, so a green board means the scenarios actually ran.
- As an operator, when Centinela *cannot* determine whether acceptance
  scenarios were skipped, I want it to say so out loud rather than pass
  silently, so an unparseable runner is a visible gap and not an invisible one.
- As an operator with a malformed `roadmap-quality.json`, I want the error to
  name the schema it expected and the field that is wrong, so I fix the shape
  instead of hunting for a bad number.
- As an operator, I want a gate that inspected nothing to render as *skipped*,
  not as *passed*.
- As a single-locale project owner, I want G11 to tell me the parity check is
  trivially satisfied, so I do not read it as evidence my translations are in
  sync.

## Acceptance Criteria

**AC1 — evidence validate uses the configured UI paths.**
`runEvidenceValidate` loads config and passes `config.UIPaths(cfg)` to
`evidence.ValidateFeature`. A repo with no `centinela.toml` still works
(`config.Load` returns defaults on `IsNotExist`, and `UIPaths(nil)` falls back
to `defaultUIPaths`). Evidence that fails the UX-output rule under `complete`
also fails under `centinela evidence validate`, and vice versa.

**AC2 — acceptance-classified commands are judged on their report, not only
their exit code.** For each `validate.commands` entry classified as an
acceptance execution, the captured output is parsed for a run summary. A
detected skipped / pending / undefined scenario (or a `0 scenarios` run) makes
the command **fail** by default, with a message naming the counts and the
command.

**AC3 — an unparseable acceptance report is reported as undetermined, never as
a pass.** When a command is acceptance-classified but its output matches no
supported summary shape, the command renders as a warning that names the
limitation and the remedy (e.g. `go test -json` / `-v`). It does not fail the
run — an undetected skip is not a proven skip — but it is never silently green.

**AC4 — the skip policy is configurable and defaults to on.**
`[validate] acceptance_skip_policy` accepts `fail` (default), `warn`, `off`.
An absent key behaves as `fail`. An unknown value is a config error.

**AC5 — non-acceptance tiers are never failed by skip detection.** A unit or
integration command that legitimately skips (build-tag-gated, platform-gated,
short-mode) is unaffected: classification gates the parse, and the parse gates
the verdict.

**AC6 — the claim verifier applies the same rule.**
`internal/verify` classifies a `RunOutcome` with detected acceptance skips as a
claim FAIL, distinct from `StatusPass`, so `centinela verify` cannot certify a
`tests-pass` claim that the validate path would now reject.

**AC7 — roadmap-quality shape errors name the shape.** A quality entry with a
missing, non-object, or wrongly-typed `scores` produces an error naming the
feature, the offending field, and the expected schema — not `scores must be
between 1 and 10`. A genuinely out-of-range integer still produces the range
error, and it names the field and the value.

**AC8 — G1 reports honestly.** An empty diff-filter run returns `Skip`, not
`Pass`. The pass message names the effective cap and that per-file exceptions
are configurable. **The 100-line default and its Fail severity are unchanged.**

**AC9 — G11 single-locale is not a confident pass.** With exactly one locale
configured under the `json` format, the gate returns `Warn` with a message
stating that one locale makes the parity check trivial and that it verifies
nothing. The locale file is still parsed, and a malformed file still fails. The
`gettext` path is unaffected — it checks per-locale `msgstr` completeness, which
is meaningful with one locale.

**AC10 — the filtered-out i18n run reports as skipped.**
`checkI18nFiltered` returns `Skip`, not `Pass`, when no locale file is in the
diff filter.

## Edge Cases

- `centinela evidence validate` in a directory with no `centinela.toml`.
- `centinela.toml` present but unparseable — the command must surface the config
  error rather than silently degrade to `nil` paths.
- `orchestration.ui_paths` explicitly configured to an empty list.
- Acceptance command that both exits non-zero **and** reports skips — the exit
  failure wins and is reported first (the run already failed).
- Acceptance command that times out mid-run with a partial summary in the
  buffer — treated as undetermined, not as a skip verdict.
- A summary-looking line appearing inside a *failure* message or a test's own
  stdout (e.g. a test that prints `2 scenarios (1 skipped)`) — parser must anchor
  on the runner's summary position/shape, not a bare substring match.
- A command matching the acceptance classifier only incidentally (e.g.
  `go test ./... ` covers `tests/acceptance` transitively) — the classifier is
  the existing `hasAcceptanceExecutionCommand` predicate; broadening it is out
  of scope.
- Mixed output: a Go run where a *non*-acceptance package skips and an
  acceptance package does not.
- `go test -json` interleaved output from parallel packages.
- Zero configured `validate.commands`.
- `roadmap-quality.json` where `features` is absent, `null`, or not an array.
- A quality entry where `scores` is present but an array, a string, or `null`.
- A quality entry where one score field is a float (`9.0`) or a numeric string.
- G1 run with `filter != nil && filter.Len() == 0` (nothing changed).
- G1 run where every oversized file is exception-justified.
- G11 with `locales = []` (already `Skip`), with one locale, and with one locale
  whose file is missing or malformed.
- G11 gettext with one locale — must stay a real pass, not the new warning.

## Data Model

No persisted entities. Three in-memory value types:

| Type | Fields | Purpose |
|---|---|---|
| `AcceptanceSummary` | `Total, Passed, Skipped, Pending, Undefined int`, `Shape string` | Parsed runner summary; `Shape` names which parser matched. |
| `CommandVerdict` | `Status (Pass\|Warn\|Fail)`, `Reason string` | Per-command outcome replacing the bare `bool`. |
| `SkipPolicy` | `fail \| warn \| off` | New `[validate] acceptance_skip_policy` knob. |

Config surface added: `ValidateConfig.AcceptanceSkipPolicy string`
(`toml:"acceptance_skip_policy"`), defaulted in `applyDefaults`, validated in
`validateConfig`.

## Out of Scope

- Lowering, relaxing, or profile-scaling **any** gate default. G1 stays a Fail
  at 100 lines under every enforcement profile. `internal/config/enforcement_profile.go`
  documents the invariant that profiles scale *process*, not the verification
  axis; making G1 advisory under `guided`/`outcome` would break it. The
  retrospective's "configurable/advisory presentation" is implemented here as
  **presentation** only — naming the cap and its source, and reporting `Skip`
  when nothing was inspected. Any enforcement-severity change belongs to the
  successor roadmap entry `guided-by-default`.
- Parsers for behave, pytest, RSpec, Jest, Playwright, Maven/Gradle. v1 covers
  the runners this repo and Centinela-scaffolded projects actually run.
- Broadening `hasAcceptanceExecutionCommand`'s classification heuristics.
- Re-running any suite to obtain a parseable report. Validate wall-clock is
  already at the job-cap limit; a probe run is explicitly rejected.
- Changing this repo's own `validate.commands` to emit `-json`.
- Spec-traceability, security, build, import-graph, or roadmap-drift gate
  reporting.
