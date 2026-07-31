# truthful-validators — senior-engineer

Four truthfulness fixes to four validators, landed as the plan's four slices
(A → B → C → D). **No gate default is lowered anywhere.** G1 stays a `Fail` at
100 lines; G11's `Fail` paths are untouched; the acceptance rule adds a new
failure mode and removes none.

## Files Touched

### Slice A — `evidence validate` uses the configured UI paths (AC1, E1–E3)
- `cmd/centinela/evidence_validate.go` (34 → 46) — loads `config.Load()` and
  passes `config.UIPaths(cfg)`, mirroring
  `internal/workflow/validate_orchestration.go:14`. A config load error is
  **returned**, never degraded to `nil`.
- `cmd/centinela/evidence_validate_uipaths_test.go` (new, 80)
- `cmd/centinela/evidence_validate_config_test.go` (new, 76)

The defect was a *permanently red* validator, not a silent one: with `nil`
uiPaths `hasAnyPrefix(files, nil)` is always false, so **every** ux-ui-specialist
role failed from the CLI while `complete` passed it. Four of the six new tests
fail against the pre-change code, and both binaries were run against the same
fixture by hand (see Trade-Offs → dogfood).

### Slice B — roadmap-quality shape errors name the shape (AC7, E11–E13)
- `internal/roadmap/quality.go` (87 → 67) — split; now types + envelope +
  `ValidateQuality` preamble only.
- `internal/roadmap/quality_features.go` (new, 44) — cross-set membership,
  threshold, summary, and `validateScoreRange` (now names field **and** value).
- `internal/roadmap/quality_shape.go` (new, 65) — two-stage decode.
- `internal/roadmap/quality_fields.go` (new, 98) — pointer-typed score mirror,
  index-aligned field tables, `scoreTypeError`, `jsonKind`.
- `internal/roadmap/promote_scores.go` — the second caller now passes the richer
  error through instead of substituting its own string.
- Tests: `quality_shape_test.go` (77), `quality_features_shape_test.go` (66),
  `quality_fields_test.go` (52).

`must be between 1 and 10` is now producible **only** by a genuinely
out-of-range integer; every structural fault is caught before it.

### Slice C — G1 and G11 report what they inspected (AC8–AC10, E14–E17)
- `internal/gates/file_size.go` (92 → 97) — empty diff scope is `Skip`.
- `internal/gates/file_size_message.go` (new, 29) — messages derive the cap from
  `maxLines`; the justified-exception pass keeps its own distinct message.
- `internal/gates/i18n.go` (64 → 75) — a single `json` locale returns `Warn`
  **after** the file has been read and parsed, so the `Fail` paths above it are
  unreachable-by-construction rather than reordered. `gettext` untouched.
- `internal/gates/i18n_filter.go` — filtered-out run is `Skip`, not `Pass`.
- Tests: `file_size_truthful_test.go` (78), `i18n_single_locale_test.go` (94),
  plus deliberate updates to `diff_aware_test.go` and `i18n_filter_test.go`
  (both previously pinned the old "gate skipped" **Pass** strings).

### Slice D — acceptance skips stop reading as green (AC2–AC6, E4–E10)
- `cmd/centinela/validate.go` (99 → 82) — split first, no behavior change.
- `cmd/centinela/validate_commands.go` (new, 60) — per-command three-valued
  verdict. Order is the contract: **exit code → classifier → parser → policy.**
- `internal/acceptance/` (new leaf, stdlib only): `classify.go` (49),
  `summary.go` (54), `parse_cucumber.go` (58), `parse_gotest.go` (87),
  `verdict.go` (64).
- `internal/workflow/validate_tests_acceptance_commands.go` — delegates to the
  leaf; **its existing test passes unchanged**, which is the proof the
  classification moved rather than broadened.
- `internal/config/acceptance_skip_policy.go` (new, 41) + `config.go` (88 → 95,
  raw-value rejection before `applyDefaults`) + `defaults.go`.
- `internal/ui/render_gates.go` (52 → 68) — `RenderCmdVerdict`; `RenderCmdResult`
  delegates, so existing callers and tests are untouched.
- `internal/verify/claim_tests.go` (66 → 92) + `claim_tests_acceptance.go`
  (new, 39).
- `centinela.toml` + `internal/scaffold/assets/centinela.toml` + `PROJECT.md` G2.

### Post-verifier remediation — three findings fixed in-branch

The adversarial verifier returned WARNING with three findings that were
violations of this feature's OWN acceptance criteria. All three are fixed here,
not deferred; the verifier's three deferral records were removed from the
Backlog and `ROADMAP.md` regenerated.

**F1 (HIGH, over-block, AC5).** `go test -v ./...` is acceptance-*classified*,
so `parseGoVerbose` counted **every** `--- SKIP:` in the repo — a unit tier's
`t.Skip("requires docker")`, a `-short` skip or a build-constraint skip failed
validate. Worse, the feature's own R4 note ("add `-json` or `-v` to make skips
detectable") steered operators straight into it. Fix: a new
`internal/acceptance/scope.go` separates two questions that had been conflated —
"does this command RUN acceptance tests" (unchanged; what the tests-step gate
asks) from "is every result it reports acceptance work". `ScopeMixed` (a
whole-repo command) counts Go results **only when attributable to a package
under `tests/acceptance`**; `go test -json` attributes via the `Package` field
the event carries, and `go test -v` via the contiguous, terminator-delimited
per-package blocks Go emits. `IsExecutionCommand` is now *defined as*
`ScopeOf(cmd) != ScopeNone`, so classification and attribution cannot drift and
the predicate accepts exactly the same commands as before. A Gherkin summary is
counted in every scope: scenarios are acceptance work by definition, unlike a
bare Go `--- SKIP:`. The "executed no scenarios" rule now fires only for a
genuinely acceptance-scoped command — under `ScopeMixed`, zero attributed
scenarios means the tier was not identifiable, not that nothing ran.

**F2 (HIGH, false-green).** `Detect` returned the FIRST matching parser, so a
run printing a clean `3 scenarios (3 passed)` beside `--- SKIP: TestGoLevelHidden`
rendered a silent ✓. Fix: every skip-data parser runs and the results are
**unioned** (`summary.go: merge`); the skip-data-free non-verbose shape is now a
strict fallback, reached only when no skip-data shape matched (so R4 is
preserved). The verifier's exact mixed-output case is pinned at the unit, cmd
and acceptance tiers.

**F3 (MEDIUM, plan R8 inert).** `completedValidationOutcome()` synthesized a
fixed prose `Output` that matched no parser, so the reuse path resolved to
*undetermined* on 100% of production runs and printed "acceptance report could
not be parsed" seconds after the gate said the opposite. Fix takes the
coordinator's preferred option **and** its structural half: `validateRunRecord`
captures the REAL per-command output, and `verify.RunOutcome` gains a typed
`AcceptanceJudged *AcceptanceJudgement` carrying the analysis the validate gate
already performed **with each command's own scope**. `inheritJudgement` reuses
it rather than re-parsing a joined transcript that can no longer say which
command a line came from. `Analysed` distinguishes "ran and found nothing" from
"never ran", so the message can never claim an analysis that did not happen. The
fallback path (a caller that sets `PriorTestRun` without a judgement) judges
with `ScopeMixed`, the only reading of a joined transcript that cannot
over-block another tier.

### Verifier round 2 — five more fixed in-branch

Round 1's F1/F2/F3 were confirmed closed and mutation-proven; the round-1 union
fix opened one new hole, and four smaller honesty defects were found around it.

**R2-F1 (HIGH, zero-scenario union false-green).** `Detect` combined shapes and
`failable` read `Scenarios == 0` off the MERGED total, so a `0 scenarios` Gherkin
summary stopped being failable the moment any passing Go test shared the output —
which is the ordinary shape of godog driven from a Go wrapper. Fix:
`Summary.GherkinZero` records the fact per shape and survives combination;
`failable` fails on it directly. A wrapper's `--- PASS:` proves the wrapper ran,
never that a scenario did.

**R2-F2 (HIGH, Gherkin ignored tier scope + a false provenance claim).**
`parseCucumber` scanned the whole output and ignored `Scope`, so a whole-repo
command failed on a UNIT package's scenario summary while the acceptance tier was
clean — and `Describe` appended ", attributed to tests/acceptance" to counts that
never passed attribution. Fix: Gherkin summaries are now bound to the enclosing
`go test` package block exactly like Go results (`parse_text.go` replaces the two
independent whole-output scans with one block-attributing pass), so the tier-level
"never" the four documents assert is finally true. The attribution clause is gated
on `Summary.Attributed`, set only by a parser that actually filtered, so no
message can assert an attribution that did not happen.

**R2-F4 (LOW, inflated totals).** `merge` summed counts across shapes, reporting a
3-scenario godog run as "of 4 scenarios" and one real skip three times when
`-json` and `-v` markers co-occurred. Fix: `atLeast` combines by per-field maximum
across SHAPES (which re-describe one run), while `add` still sums across package
BLOCKS (which are genuinely different work).

**R2-F5 (LOW, `off` claimed an analysis ran).** Under `acceptance_skip_policy =
off` nothing is parsed, yet the record set `Analysed: true` — the one field whose
job is distinguishing "ran and found nothing" from "never ran". Fix: the record
carries `disabled`, and the reuse message reads "was not performed — skip
detection disabled by [validate] acceptance_skip_policy = off".

**R2-F6 (LOW, false test rationale).** The `-json` fixture's inline `printf`
payload named `tests/acceptance`, which also put that literal in the COMMAND
string and flipped the command's own scope, so the comment's claim about
whole-repo attribution was false. Fix: the fixture uses `tvValidateWithScript`, so
the payload cannot leak into the command string, and the comment now explains
exactly why that matters.

**R2-F3 is deferred, not fixed**, per instruction: an unindented go-test
terminator inside a package's own stdout can steal a block's attribution —
recorded as `acceptance-verbose-terminator-attribution-steal`. All four round-2
deferral records were left in place.

The reuse fallback also gained `reusedScope`: a joined transcript is `ScopeMixed`
only when some configured acceptance command is whole-repo, so a bare Gherkin
report from an all-acceptance-scoped set is still counted rather than silently
discarded for lacking go package blocks.

## Architecture Compliance

- **G1:** every source **and** test file ≤ 100 lines. Three files were split
  *before* a line was added (`validate.go` 99, `quality.go` 87, `file_size.go`
  92), and `quality.go` was split a second time when the first split still
  landed at 106.
- **G2:** `internal/acceptance` is a **stdlib-only leaf**, registered in
  `[[gates.import_graph.layers]]` and documented in PROJECT.md G2. That is the
  whole point of the placement: `cmd/`, `internal/verify` **and**
  `internal/workflow` all need the predicate, and exporting it from
  `internal/workflow` would have forced an `internal/verify → internal/workflow`
  edge. Its only imports are `encoding/json`, `fmt`, `regexp`, `strconv`,
  `strings`.
- **G7:** `cmd/centinela/validate_commands.go` holds no rule. The exit-code
  precedence, the classification, the parsing and the policy all live in
  `internal/acceptance`; `cmd/` only translates the leaf's `Verdict` into the
  `gates.Status` vocabulary for rendering.

## Type-Safety Notes

- The bare `bool` per validate command is replaced by a three-valued
  `acceptance.Verdict`; `Warn` is now representable instead of being collapsed
  into `passed == true`.
- `QualityScores` decoding moved to a **pointer-typed mirror**: `*int` per
  field, so "absent" is a distinct state from "zero". That distinction *is* the
  bug fix — a value-typed decode is what produced "scores must be between 1 and
  10" for a missing object.
- `verify.runVerdict` carries the advisory note as a **field**, so the aggregate
  PASS detail composes notes instead of any caller re-parsing them out of a
  formatted string.
- `Summary.SkipData` makes "recognized shape, no skip information" a typed state
  rather than an absence.
- No `interface{}`/`any` added anywhere.

## Trade-Offs

**R4 — the permanent ⚠ does not ship (orchestrator override to the plan).**
The plan accepted a standing `⚠ could not parse` on this repo's command 1
(`go test ./... -coverprofile=coverage.out`, whose non-verbose output carries no
skip information). That violates this feature's own north star: a validator that
is permanently warning trains everyone to ignore warnings. **Option (a) was
chosen**: `parse_gotest.go` *recognizes* plain non-verbose `go test` output
(`ok`/`FAIL`/`?` package lines) and returns `SkipData: false` — a quiet
detail-level note on a **Pass** ("go test (non-verbose) carries no skip data —
add -json or -v to make skips detectable"). `⚠` stays reserved for output
matching no known runner at all, which is the honest "should have carried skip
data and did not" case. Option (b) (`-json`) was rejected as *not* trivially
safe: although `check-coverage.sh` reads the coverage **file**
(`COVERAGE_PROFILE`), not stdout, `-json` would replace the rendered failure
output with an unreadable JSON stream on every failing run. Divergence from plan
§D6 recorded: the plan listed "plain non-verbose `go test` → `Detect` returns
**false**"; it now returns `true` with `SkipData: false`.

**R8 — the `PriorTestRun` reuse path is armed.** The reused outcome is labelled
`"validate.commands"`, not a real command, so per-command classification cannot
apply. `checkTestsPass` classifies it over `cfg.Validate.Commands` as a whole
(`acceptance.AnyExecutionCommand`), with a dedicated test that fails if the path
is disarmed. This is the exact path `complete` uses.

**R2 — non-acceptance tiers are structurally safe.** AC5 is not a heuristic: the
classifier gates the parse and the parse gates the verdict. Pinned in both
directions at the unit level and end-to-end through `runValidateCommands`.

**R9 — `promote`'s message changed deliberately.** Contrary to the plan's note,
no existing test asserted the old `"each score must be between 1 and 10"`
string, so nothing had to be loosened; a new assertion pins the richer message.

**Dogfood (scratch binary, `/tmp/cv-truthful`):** acceptance command reporting
skips → validate **fails** (exit 1, message names counts + command); same with
`acceptance_skip_policy = "warn"` → `⚠`, exit 0; non-acceptance command
reporting skips → green; unparseable acceptance output → `⚠`, exit 0; the
nil-uiPaths ux fixture → `evidence ok` (exit 0) on the new binary, and
`ux-ui-specialist outputs must include at least one real UI file` (exit 1) on a
binary built with the pre-change `nil`.

**R6 — the single-locale `Warn` names its own blast radius.** The G11 message
states that it does not fail validate and that `pr_gate.fail_on_warning` (off by
default) would surface it, so a single-locale project reading the warning knows
exactly what it costs them.

**Deferrals removed.** The verifier recorded
`acceptance-skip-whole-repo-command-overblocks`,
`acceptance-parser-precedence-hides-go-skips` and
`prior-run-skip-analysis-structurally-undetermined`. All three are fixed in this
branch, so the records were removed with `centinela roadmap remove` and
`ROADMAP.md` regenerated — the roadmap must not carry false open items, and the
`roadmap_drift` ⚠ this branch introduced is now gone.

**Deferred (new, `--source truthful-validators/senior-engineer`):**
`validate-command-verdicts-in-machine-output` — the new per-command verdict is
rendered only in the terminal; `centinela verdict` and `pr-gate` carry gate
results and claim checks but no validate-command line, so an undetermined
acceptance report is invisible to a CI consumer of the machine output. Checked
against the Backlog first: `render-warn-gate-details` is about WARN *gate*
details in *deliver* output, and the planner's two entries
(`acceptance-skip-parser-breadth`, `self-validate-acceptance-report-shape`)
remain accurate and were not re-recorded.

**Deliberately not done:** parsers for behave/pytest-bdd/RSpec/Jest/Playwright
(they land in the honest undetermined bucket), any change to this repo's
`validate.commands`, and the enforcement-severity half of WS5.2 — which belongs
to `guided-by-default`.

## Handoff

Next role: **qa-senior**. Acceptance tests under `tests/acceptance/` are the
tests step and were deliberately not written here; the spec's 43 scenarios in
sections A–F map onto the unit coverage listed above. Two things worth
adversarial attention: (1) the anchored cucumber regex and its false-positive
pin (E8), and (2) that `acceptance_skip_policy = "off"` really does restore
exit-code-only behavior end to end.
