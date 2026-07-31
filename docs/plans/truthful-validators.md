# Plan: truthful-validators

**Brief:** [docs/features/truthful-validators.md](../features/truthful-validators.md)
**Spec:** [specs/truthful-validators.feature](../../specs/truthful-validators.feature)
**Archetype:** canonical — plan → code → tests → validate → docs
**Plan contract:** `planner-v1` (one planner, both lenses)

## 0. What this plan is

Four independent truthfulness fixes to four validators, in four slices. No gate
default is lowered anywhere in this plan. Every slice has the same shape:
**make the validator say what it actually knows → pin it with a test that fails
in both directions.**

### 0.1 Verification of the source claims against the CURRENT tree

The retrospective that motivated this feature was written several merges ago
(`main` is now at the `0.52.x` line: single-run validate cycle, spec-conflict
hotfix, token-diet, unified-plan-specialist). Every cited defect was re-read at
worktree revision `1b0af2b`. **All four are still live; none was fixed en route.**
The line references below are the current ones and supersede the retrospective's.

| Item | Retrospective claim | Current tree | Verdict |
|---|---|---|---|
| WS1.2 | `evidence_validate.go:25` passes `nil` uiPaths | `cmd/centinela/evidence_validate.go:25` — `evidence.ValidateFeature(feature, nil)`; `internal/workflow/validate_orchestration.go:14` passes `config.UIPaths(cfg)` | **LIVE** |
| WS1.3 | validate judges acceptance by exit code only | `cmd/centinela/validate.go:92` — `passed, out := runCommand(cmd)`; `out` is only rendered (`RenderCmdResult`, and only on failure). `internal/verify/claim_tests.go:41` `classifyTestRun` switches on `TimedOut / StartErr / ExitCode` and ignores `RunOutcome.Output` | **LIVE (both paths)** |
| WS1.5 | `quality.go` reports a range error for a shape fault | `internal/roadmap/quality.go:79` `validateScoreRange` over a value-typed `QualityScores`; a missing/mistyped `scores` decodes to zeros → `"scores must be between 1 and 10"` | **LIVE** |
| WS5.2 | G1 and G11 report more confidence than earned | `internal/gates/file_size.go:19-26` returns `Pass` + `"No relevant changes — gate skipped."`; `internal/gates/i18n_keys.go:28` iterates `locales[1:]` and `:42` returns `Pass, "All locales have identical keys."`; `internal/gates/i18n_filter.go:18` returns `Pass` for a filtered-out run | **LIVE** |

**Nothing is dropped from scope.** One scope *narrowing* is recorded in §6: the
retrospective's "advisory G1 under non-strict profiles" is implemented as
**presentation only**, because `internal/config/enforcement_profile.go:8-11`
states the invariant that profiles scale *process*, not the verification axis,
and `.workflow/roadmap.json` makes `guided-by-default` depend on this feature —
that is where an enforcement-severity change belongs.

### 0.2 Facts that dominate the sequencing — read before touching anything

- **`cmd/centinela/validate.go` is 99 lines.** G1 (≤100, `_test.go` included)
  trips on the first line added. Slice D therefore *opens* with a split.
- **`internal/roadmap/quality.go` is 87 lines** and slice B roughly doubles its
  decode path. It also opens with a split.
- **`validateScoreRange` has a second caller**: `internal/roadmap/promote_scores.go:29`,
  which throws the returned message away and substitutes `"each score must be
  between 1 and 10"`. Changing the shared helper must not change `promote`'s
  contract (§3.3).
- **`hasAcceptanceExecutionCommand` is unexported and lives in
  `internal/workflow`** (`validate_tests_acceptance_commands.go`, 25 lines) and
  answers *any-of a list*, not *is this one command*. Slice D needs a per-command
  predicate in a package that `cmd/`, `internal/verify` **and** `internal/workflow`
  may all import — see §3.4 for the placement decision and why it is not
  `internal/workflow`.
- **Slice D changes the validate runner that this feature's own validate step
  will run.** This is the self-referential risk R1; §5 says exactly what must be
  true for that step to be trustworthy.
- **This repo's own command 1** is `go test ./... -coverprofile=coverage.out`.
  It classifies as acceptance and its **non-verbose output contains no skip
  information at all**. Under this plan it becomes a permanent *undetermined*
  warning, not a pass and not a failure. That is the intended honest outcome; the
  remedy (`-json`) is deferred, not adopted — see §6 and R4.

---

## 1. Slice A — `evidence validate` uses the configured UI paths

**Delivers:** AC1, E1–E3. **Blast radius:** one CLI command.
**Why first:** smallest correct slice, no new package, no new config, and it
makes the CLI agree with `complete` — which every later slice's evidence relies on.

### A1. Load config and pass real UI paths

`cmd/centinela/evidence_validate.go` (34 lines → ~44)

- In `runEvidenceValidate`, call `config.Load()` before `evidence.ValidateFeature`.
- On error, **return it** — do not fall back to `nil`. A malformed
  `centinela.toml` must surface as a config error, never as a silently
  degraded validator (E2). `config.Load` already returns `defaultConfig()` for
  `os.IsNotExist`, so a repo with no `centinela.toml` keeps working (E1).
- Pass `config.UIPaths(cfg)`. This mirrors `internal/workflow/validate_orchestration.go:14`
  exactly; the two entry points must not diverge again.
- `config.UIPaths(nil)` and an empty `orchestration.ui_paths` both fall back to
  `defaultUIPaths` (`internal/config/orchestration.go:73-75`), so E3 is already
  handled by the callee — assert it, do not re-implement it.

### A2. Tests (both directions)

`cmd/centinela/evidence_validate_test.go` (52 lines → ≤100; split to
`evidence_validate_uipaths_test.go` if it would exceed)

- **Red-before-green:** a `ux-ui-specialist` evidence file whose `outputs` touch
  no UI path must now produce at least one hint. Under `nil` this test passes
  vacuously — it must fail against the pre-change binary.
- **Green:** the same evidence with an output under a configured
  `orchestration.ui_paths` entry produces no hint.
- **Parity:** the hint set from `evidence.ValidateFeature(f, config.UIPaths(cfg))`
  equals the one the orchestration path produces for the same fixture.
- **No config file** → exit 0 on clean evidence (E1).
- **Unparseable `centinela.toml`** → non-zero exit naming the parse error (E2).

---

## 2. Slice B — roadmap-quality shape errors name the shape

**Delivers:** AC7, E11–E13. **Blast radius:** `internal/roadmap` + `centinela roadmap validate`.
**Why second:** fully self-contained, zero interaction with slices A, C, D.

### B1. Split first

`internal/roadmap/quality.go` is 87 lines. Move the decode + per-feature
validation into a new `internal/roadmap/quality_shape.go` (≤100), leaving
`quality.go` holding the types, `ValidateQuality`'s file/role/threshold
preamble, and the cross-set membership checks.

### B2. Two-stage decode

`internal/roadmap/quality_shape.go` (new, ≤100 lines)

Stage 1 — structural, on `json.RawMessage`:

```
type qualityFeatureRaw struct {
    Name    string           `json:"name"`
    Scores  *json.RawMessage `json:"scores"`
    Summary string           `json:"summary"`
}
```

- `features` absent / `null` / not an array → one error naming the field and
  the expected shape (E11). Do **not** use `DisallowUnknownFields` — it would
  reject every existing quality file.
- `Scores == nil` → `feature %q: required object field "scores" is missing —
  expected {acceptanceCriteria, userValue, definitionClarity, dependencies,
  effortEstimation, overall}, each an integer 1–10`.
- `Scores` present but not a JSON object (array, string, number, `null`) →
  the same schema sentence, prefixed with what was actually found (E12).

Stage 2 — per-field, into a pointer-typed mirror (`*int` per field):

- Any nil field after unmarshal → `feature %q: score field %q is missing`.
- An `encoding/json` type error (a float `9.0`, a numeric string `"9"`) →
  wrap it as a shape error naming the feature **and the offending field**,
  never as a range error (E13).

The words `must be between 1 and 10` may only ever be produced by a genuinely
out-of-range integer. That is the whole point of the slice.

### B3. Range errors name the field and the value

`internal/roadmap/quality.go` — `validateScoreRange`

- Replace the `[]int` loop with a `[]struct{name string; v int}` so the error
  reads `feature %q: score %q is %d, must be between 1 and 10`.
- **`internal/roadmap/promote_scores.go:29` is the second caller.** It currently
  discards the message. Let it pass the richer error through
  (`fmt.Errorf("%w", err)` or direct return) — its CSV parser makes a *structural*
  fault impossible, so the range message is the only one it can produce and
  naming the field is a strict improvement. Its existing tests assert the old
  string and must be updated deliberately, not by loosening the assertion.

### B4. Tests (both directions)

`internal/roadmap/quality_shape_test.go` (new, ≤100 lines); existing
`quality_test.go` / `quality_branches_test.go` / `quality_cover_test.go` keep
their current cases.

- Missing `scores` → error mentions `"scores"` and **does not** contain
  `between 1 and 10`.
- `scores` as an array / string / `null` → same assertion.
- one field missing / `9.0` / `"9"` → error names that field, not the range.
- a genuine `0` or `11` → **still** the range error, now naming field + value.
- `features` missing / `null` / not an array → named structural error.
- A valid file still validates byte-identically (regression pin).

---

## 3. Slice C — G1 and G11 report what they actually inspected

**Delivers:** AC8, AC9, AC10, E14–E17. **Blast radius:** two gates + their renderers.
**Non-negotiable:** `maxLines = 100` and the `Fail` severity are **unchanged**.
This slice changes *messages and the Pass/Skip/Warn axis only*.

### C1. G1: a gate that inspected nothing reports `Skip`

`internal/gates/file_size.go` (92 lines) — extract the message/status decision
into a new `internal/gates/file_size_message.go` (small) so `file_size.go` stays
under budget.

- `filter != nil && filter.Len() == 0` → `Status: Skip`, message
  `"No files in the diff scope — nothing inspected."` (was `Pass` + "gate skipped").
- Clean pass → name the effective cap and that exceptions exist:
  `"All in-scope files are within the 100-line cap (per-file exceptions are
  configurable under [[gates.file_size_exceptions]])."`
- Justified-exception pass keeps its distinct message (E15).
- The `Fail` branch is untouched.

`AllPassed` (`internal/gates/gates.go:74`) only fails on `Fail`, so `Skip` cannot
turn a green run red. `internal/ui/render_gates.go:23` and
`internal/ui/render_markdown.go:81` (`default → ➖ skip`) and
`internal/verdict/mappers.go:17,54` all already handle `Skip` — verify, do not add.

**Known assertion to update:** `internal/gates/diff_aware_test.go:25` asserts the
old Pass message. Update it to the new status **and** add its mirror (a non-empty
filter with a clean file still returns `Pass`).

### C2. G11: one locale is not a parity result

`internal/gates/i18n.go` / `i18n_keys.go`

- In the `json` path, when `len(i.Locales) == 1`: still read and parse the
  locale file (a missing or malformed file **still `Fail`s** — E17), then return
  `Status: Warn`, message `"1 locale configured — the key-parity check is
  trivially satisfied and verifies nothing."`
- Implement it where the locale count is known and the file has already been
  parsed, so the Fail paths above it are unreachable-by-construction rather than
  reordered.
- `compareKeysets`' `locales[1:]` loop is left as-is for ≥2 locales.
- **`gettext` is untouched** (`i18n_gettext.go` checks per-locale `msgstr`
  completeness, which is meaningful with a single locale — E16).

### C3. G11 filtered-out reports `Skip`

`internal/gates/i18n_filter.go:18` — `Pass` → `Skip`, message
`"No locale files in the diff scope — nothing inspected."`
Update `internal/gates/i18n_filter_test.go:19` accordingly (it asserts both the
status and the string).

### C4. Tests (both directions)

`internal/gates/file_size_more_test.go`, `i18n_more_test.go` (existing, room to
grow; new `_test.go` files if a cap is threatened — all ≤100 lines):

- empty filter → `Skip`, and **the message does not claim a pass**;
- non-empty filter, clean → `Pass`, message names the cap;
- a 101-line file → **still `Fail`** (the anti-weakening pin; assert the status
  *and* that `maxLines` is 100);
- one locale, valid file → `Warn`; one locale, missing file → `Fail`; one
  locale, malformed JSON → `Fail`;
- two locales in sync → `Pass` (unchanged); out of sync → `Fail` (unchanged);
- gettext + one locale → **not** the new warning;
- filtered-out i18n → `Skip`.

---

## 4. Slice D — acceptance skips stop reading as green

**Delivers:** AC2–AC6, E4–E10. **Blast radius:** `centinela validate`,
`centinela verify`, `centinela complete` at the validate step, `pr-gate` output.
**Land last:** largest diff, and it is the one that changes the machinery this
feature's own validate step runs (R1).

### D1. Split `cmd/centinela/validate.go` before adding a line

99 lines today. Move `runValidateCommands` (lines 84–98) into a new
`cmd/centinela/validate_commands.go`. `validate.go` drops to ~84 and gains no
behavior in this slice.

### D2. New leaf package `internal/acceptance`

**Placement decision.** The classifier and the parser are needed by `cmd/`,
`internal/verify` **and** `internal/workflow`. Exporting the predicate from
`internal/workflow` would force `internal/verify → internal/workflow`; a
stdlib-only leaf avoids that edge entirely and cannot create a cycle.

`internal/acceptance` imports **stdlib only**. Files, each ≤100 lines:

| File | Contents |
|---|---|
| `classify.go` | `func IsExecutionCommand(cmd string) bool` — the body of today's `hasAcceptanceExecutionCommand` loop, per command. **Byte-identical predicate; no broadened heuristics** (E9). |
| `summary.go` | `type Summary struct { Shape string; Scenarios, Passed, Skipped, Pending, Undefined int }`; `func Detect(output string) (Summary, bool)` — tries each parser, returns `false` when none matches. |
| `parse_cucumber.go` | The `N scenarios (…)` summary line emitted by **cucumber-js and godog alike**. Anchored: the line must *begin* with `^\d+ scenarios?\b` (or be exactly `0 scenarios`) at the start of a line, so a summary-shaped string inside a test's own stdout does not match (E8). |
| `parse_gotest.go` | `go test -json` (`{"Action":"skip","Test":"…"}` — package-level `skip` = "no test files", which is **not** a scenario skip) and `go test -v` text (`--- SKIP:` lines). |

`internal/workflow/validate_tests_acceptance_commands.go` is rewritten to
delegate: `hasAcceptanceExecutionCommand` loops calling
`acceptance.IsExecutionCommand`. Its existing tests must pass **unchanged** —
that is the proof the classification did not move.

Register `internal/acceptance/**` in the `leaf` layer of
`[[gates.import_graph.layers]]` in `centinela.toml`, with the one-line rationale
the neighbouring leaves carry, and mirror the note into PROJECT.md G2.

### D3. Config: `[validate] acceptance_skip_policy`

`internal/config/acceptance_skip_policy.go` (new, small)

- `PolicyFail = "fail"` (default), `PolicyWarn = "warn"`, `PolicyOff = "off"`.
- `NormalizeAcceptanceSkipPolicy("")` → `fail`.
- `validateAcceptanceSkipPolicy(raw string) error` rejects a **non-empty
  unknown** value.

Wire exactly like the enforcement profile, which has the same
default-on/raw-validation shape:

- `internal/config/config.go` (88 → ~91): add
  `AcceptanceSkipPolicy string \`toml:"acceptance_skip_policy"\`` to
  `ValidateConfig`, and call `validateAcceptanceSkipPolicy` **before**
  `applyDefaults` (mirroring `validateEnforcementProfile` at `config.go:69`) —
  otherwise normalization would swallow the unknown value (AC4).
- `internal/config/defaults.go`: `cfg.Validate.AcceptanceSkipPolicy =
  NormalizeAcceptanceSkipPolicy(cfg.Validate.AcceptanceSkipPolicy)`.
- Document the knob in `internal/scaffold/assets/centinela.toml` next to
  `[validate] commands` and in `centinela.toml`'s comment block.

### D4. Verdict, not a bool, in the validate runner

`cmd/centinela/validate_commands.go` (new, ≤100 lines)

Per configured command:

1. `passed, out := runCommand(cmd)` — unchanged; `runCommand` in
   `validate_runner.go` already captures combined output.
2. **Exit code first.** `!passed` → `Fail`, rendered as today. An acceptance
   command that both exits non-zero *and* reports skips is reported as the exit
   failure; no skip analysis is run (E6).
3. `!acceptance.IsExecutionCommand(cmd)` → `Pass`. **The classifier gates the
   parse and the parse gates the verdict** — this is the whole of AC5; a unit or
   integration command that legitimately skips (build tags, `-short`, platform
   gates) can never be failed by this feature.
4. `s, ok := acceptance.Detect(out)`:
   - `!ok` → **`Warn`**: `"acceptance report could not be parsed — skips
     undetected; run with `go test -json` / `-v` or a cucumber-compatible
     summary"`. Never a pass, never a failure (AC3). An undetected skip is not a
     proven skip.
   - `ok` and `s.Skipped+s.Pending+s.Undefined > 0`, or `s.Scenarios == 0` →
     apply the policy: `fail` → `Fail`, `warn` → `Warn`, `off` → `Pass`. The
     message names the counts, the shape, and the command.
   - otherwise → `Pass`.

`runValidateCommands` fails the run only on `Fail` (a `Warn` prints and does not
flip `allPassed`), preserving `executeValidationWithFlag`'s contract.

Rendering: `internal/ui/render_gates.go` (52 lines) gains
`RenderCmdVerdict(cmd string, status gates.Status, detail, output string) string`
(the file already imports `internal/gates`). `RenderCmdResult` stays and
delegates, so existing callers and tests are untouched.

**Zero configured commands** short-circuits before any of this (E10, unchanged).

### D5. The same rule on the claim-verification path

`internal/verify/claim_tests.go` (66 lines) + new
`internal/verify/claim_tests_acceptance.go`

- `classifyTestRun`'s `default:` branch (exit 0) additionally consults
  `acceptance.Detect(out.Output)` when the command is acceptance-classified,
  and returns `StatusFail` with a detail naming the counts — **distinct from
  `StatusPass`**, so `centinela verify` cannot certify a `tests-pass` claim the
  validate path would now reject (AC6).
- `classifyTestRun` needs the policy: thread it as an explicit parameter from
  `checkTestsPass` (which already has `cfg`). Do not read config inside the
  classifier.
- **`deps.PriorTestRun` (the single-run reuse wired at
  `cmd/centinela/complete_verify.go:23`) is labelled `"validate.commands"`, not a
  real command string.** For that path, classify on
  `cfg.Validate.Commands` as a whole (any command acceptance-classified → parse
  the shared output). Getting this wrong silently disarms AC6 on the exact path
  `complete` uses.
- An unparseable report on this path is `StatusPass` with the detail naming the
  limitation — the verifier must not invent a failure it cannot prove.

### D6. Tests (both directions)

`internal/acceptance/*_test.go` (table-driven, each file ≤100 lines):

- cucumber-js `3 scenarios (1 skipped, 2 passed)` → `Skipped=1`;
- godog `2 scenarios (1 undefined, 1 passed)` → `Undefined=1`;
- `0 scenarios` / `0 scenarios (0 passed)` → `Scenarios=0`;
- all-passed summary → detected, zero skips;
- `go test -json` with a test-level `"Action":"skip"` → `Skipped≥1`;
- `go test -json` with only a **package-level** skip (`[no test files]`) → **not**
  a scenario skip (the false-positive pin);
- `go test -v` `--- SKIP:` → `Skipped≥1`;
- plain non-verbose `go test ./...` `ok` lines → `Detect` returns **`false`**
  (undetermined — this is this repo's own case, R4);
- a test's stdout containing the literal `2 scenarios (1 skipped)` mid-line →
  **not** matched (E8);
- interleaved `-json` output from parallel packages → counts still correct (E9);
- truncated / partial summary → `false`, never a skip verdict (E7);
- `IsExecutionCommand` parity table against every branch of the old predicate.

`cmd/centinela/validate_commands_test.go` (new, ≤100):

- acceptance command, exit 0, skips reported, default policy → run **fails**,
  message names counts + command;
- same with `warn` → warns, run passes; with `off` → passes silently;
- acceptance command, exit **non-zero**, skips reported → the *exit* failure is
  what is reported (E6);
- **non-acceptance** command reporting skips → passes untouched (AC5);
- unparseable acceptance output → warn, run still passes (AC3).

`internal/config`: unknown policy value → `config.Load` errors; absent key →
`fail`; explicit `fail|warn|off` round-trip.

`internal/verify`: acceptance skip on a `PriorTestRun` → `StatusFail`;
non-acceptance commands → `StatusPass`; unparseable → `StatusPass` with the
limitation named.

`tests/acceptance/truthful_validators_*_test.go` — one file per spec section,
each `// Acceptance: specs/truthful-validators.feature` + `// Scenario: <name>`
(the spec-traceability gate reads these headers).

---

## 5. Risks

| Risk | Impact | Likelihood | Mitigation |
|---|---|---|---|
| **R1 — self-reference:** this feature's own validate step runs the runner slice D changes | High | **Certain** | Land D last with A/B/C green. Before the validate step, run `go build ./cmd/centinela -o /tmp/cv` and dry-run `/tmp/cv validate` from the worktree — the installed binary lags. Read the per-command verdicts by hand; a `Warn` on command 1 is expected (R4), a `Fail` is a real defect. |
| **R2 — false-FAIL of legitimate skips in non-acceptance tiers** | High | Medium | AC5 is structural, not a heuristic: `IsExecutionCommand` gates the parse and the parse gates the verdict. Pinned by an explicit "non-acceptance command reporting skips passes" test. The classifier predicate is copied byte-for-byte and its existing workflow tests must pass unchanged. |
| **R3 — parser breadth is the effort driver** | Medium | High | v1 covers exactly what this repo and Centinela-scaffolded projects run: `go test` (`-json` and `-v`) and the cucumber/godog `N scenarios (…)` line — one parser serves both, since godog copies cucumber's summary. behave / pytest-bdd / RSpec / Jest / Playwright are **deferred**, and land in the honest `Warn` bucket meanwhile, never in a false pass. |
| **R4 — this repo's own command 1 is unparseable** | Medium | Certain | `go test ./... -coverprofile=coverage.out` (non-verbose) emits no skip information, so every local and CI validate will show one `⚠ acceptance report could not be parsed`. This is the honest reading, it does not fail the run, and switching to `-json` is out of scope (it would reshape the single-run coverage-profile reuse and the rendered failure output). Deferred as `self-validate-acceptance-report-shape`. |
| **R5 — summary-shaped text inside test output** | Medium | Medium | Parsers anchor at line start on the runner's summary shape, never a bare `strings.Contains`. Explicit false-positive test (E8). |
| **R6 — G11 `Warn` + `pr_gate.fail_on_warning = true` blocks single-locale repos** | Medium | Low | `fail_on_warning` defaults `false` (`internal/config/pr_gate.go:12-16`) and the i18n gate is opt-in (`gates.i18n_enabled` is false by default and unset in this repo). Document the interaction in the gate message and the changelog. |
| **R7 — G1 `Pass → Skip` shifts downstream tallies** | Low | Medium | `AllPassed` ignores non-`Fail`; `Skip` is already handled in `render_gates.go`, `render_markdown.go` (`default → ➖ skip`) and `verdict/mappers.go`. `markdownHeader` counts only pass/fail/warn, so a skipped G1 simply stops inflating the "passed" count — correct, but it changes `pr-gate` header numbers. Assert the new counts in a test rather than discovering them in CI. |
| **R8 — `PriorTestRun` path silently disarms AC6** | High | Medium | The reused outcome is labelled `"validate.commands"`, so per-command classification does not apply. §D5 classifies over `cfg.Validate.Commands` for that path, with a dedicated test. |
| **R9 — changing `validateScoreRange` breaks `roadmap promote`** | Low | Medium | Second caller identified (`promote_scores.go:29`); its tests assert the old string and are updated deliberately. |
| **R10 — G1 budget:** four of the touched files are within 15 lines of the cap | Medium | High | Every slice that grows a near-cap file **opens with a split** (§D1, §B1, §C1). `_test.go` files count too. |

## 6. Rollout

- **Step 1 — Slice A.** Independent, no new package. Proves the loop.
- **Step 2 — Slice B.** Independent; `internal/roadmap` only.
- **Step 3 — Slice C.** Independent; two gates + their renderers.
- **Step 4 — Slice D.** Last, largest, self-referential.
- **Not now:** the enforcement-severity half of WS5.2 (G1 advisory under
  `guided`/`outcome`) belongs to `guided-by-default`, which `.workflow/roadmap.json`
  already declares `dependsOn: [truthful-validators]`. Parser breadth beyond
  go/cucumber/godog, and re-shaping this repo's own `validate.commands`, are
  deferred (§7).

## 7. Deferred findings

Recorded via `centinela roadmap defer … --source truthful-validators/planner`:

- `acceptance-skip-parser-breadth` — `behave` is *classified* as an acceptance
  runner today but gets no parser in v1, so behave-based projects sit
  permanently in the `Warn`/undetermined bucket; same for pytest-bdd, RSpec,
  Jest and Playwright.
- `self-validate-acceptance-report-shape` — Centinela's own
  `go test ./... -coverprofile=coverage.out` emits no skip information, so its
  own acceptance gate is permanently undetermined; deciding between `-json`,
  `-v`, and a separate acceptance command interacts with the single-run
  coverage-profile reuse and is its own piece of work.

Checked against `.workflow/roadmap.json` Backlog before recording: neither
overlaps an existing entry (the nearest neighbours are
`validate-flake-diagnosability`, `cross-process-suite-result-reuse` and
`render-warn-gate-details`, all distinct).

## 8. Handoff

Next role: **senior-engineer**. Open questions: none blocking. Two decisions are
the planner's and should not be relitigated without saying so explicitly: the
`internal/acceptance` leaf placement (§D2) and G11 single-locale rendering as
`Warn` rather than `Skip` (§C2, R6).
