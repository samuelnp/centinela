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
