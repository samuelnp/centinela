Feature: Truthful validators — a green verdict means something was actually checked
  As an operator (and as an orchestrating agent) reading a Centinela verdict
  I want a validator that structurally could not check to say so, an acceptance
    suite that skipped everything to fail, and a malformed input to be named by
    its shape
  So that "PASS" and "checked and clean" mean the same thing, without any gate
    default being lowered anywhere

  # Design decisions this spec encodes (docs/plans/truthful-validators.md):
  # D1 `centinela evidence validate` loads config and passes config.UIPaths(cfg),
  # mirroring internal/workflow/validate_orchestration.go; a config LOAD ERROR
  # surfaces, it never degrades to nil. D2 acceptance skip detection is gated by
  # the EXISTING classifier: classification gates the parse and the parse gates
  # the verdict, so non-acceptance tiers can never be failed by it. D3 an
  # unparseable acceptance report is UNDETERMINED (warn), never a pass and never
  # a failure — an undetected skip is not a proven skip. D4 exit code wins: a
  # command that fails AND reports skips is reported as the exit failure. D5 the
  # skip policy is [validate] acceptance_skip_policy = fail|warn|off, default
  # fail, unknown value is a config error. D6 shape faults in roadmap-quality
  # are reported as shape faults; "must be between 1 and 10" is reserved for a
  # genuinely out-of-range integer. D7 a gate that inspected nothing reports
  # SKIP, not PASS. D8 one locale reports WARN because parity verified nothing.
  # D9 NO gate default is lowered: G1 stays a Fail at 100 lines under every
  # enforcement profile; the advisory-severity question belongs to the successor
  # feature guided-by-default. D10 v1 parses go test (-json/-v) and the
  # cucumber/godog "N scenarios (…)" summary only; other runners land in the
  # honest undetermined bucket.

  # ── A. evidence validate applies the UX-output rule ───────────────────────

  Scenario: Evidence that fails the UX-output rule under complete also fails from the CLI
    Given a feature with "ux-ui-specialist" evidence whose outputs touch no UI path
    When "centinela evidence validate" runs for that feature
    Then it should report a UX-output issue for the ux-ui-specialist role
    And it should exit non-zero
    And the issues should match the ones the workflow completion path reports
      # under the previous nil-uiPaths behavior the rule could never fire

  Scenario: Evidence whose outputs include a configured UI path validates
    Given a project configuring "orchestration.ui_paths" as ["web/app"]
    And a feature with "ux-ui-specialist" evidence listing an output under "web/app/"
    When "centinela evidence validate" runs for that feature
    Then it should report no issues
    And it should exit zero

  Scenario: A repository with no centinela.toml still validates evidence
    Given a repository containing no "centinela.toml"
    And a feature with clean evidence for every required role
    When "centinela evidence validate" runs for that feature
    Then it should exit zero
    And the UX-output rule should be applied against the built-in default UI paths

  Scenario: An unparseable centinela.toml is a reported error, not a silent downgrade
    Given a repository whose "centinela.toml" is not valid TOML
    When "centinela evidence validate" runs for a feature
    Then it should exit non-zero
    And the output should name the configuration parse failure
    And it should not report "evidence ok"

  Scenario: An explicitly empty ui_paths list falls back to the built-in defaults
    Given a project configuring "orchestration.ui_paths" as an empty list
    When "centinela evidence validate" runs for a feature
    Then the UX-output rule should be evaluated against the built-in default UI paths

  # ── B. Acceptance skips fail; unparseable reports are undetermined ────────

  Scenario: A cucumber run reporting a skipped scenario fails validate
    Given a validate command classified as an acceptance execution
    And the command exits zero and reports "3 scenarios (1 skipped, 2 passed)"
    When "centinela validate" runs
    Then the command should be reported as failed
    And the message should name the skipped count and the command
    And the run should exit non-zero

  Scenario: A godog run whose steps are all undefined fails validate
    Given a validate command classified as an acceptance execution
    And the command exits zero and reports "2 scenarios (2 undefined)"
    When "centinela validate" runs
    Then the command should be reported as failed
    And the message should name the undefined count

  Scenario: A run that executed no scenarios at all fails validate
    Given a validate command classified as an acceptance execution
    And the command exits zero and reports "0 scenarios"
    When "centinela validate" runs
    Then the command should be reported as failed
    And the message should state that no scenarios were executed

  Scenario: A Go acceptance test that calls t.Skip fails validate
    Given a validate command classified as an acceptance execution
    And the command emits "go test -json" output containing a test-level skip action
    When "centinela validate" runs
    Then the command should be reported as failed
    And the message should name the skipped test count

  Scenario: A package with no test files is not a skipped scenario
    Given a validate command classified as an acceptance execution
    And the command emits "go test -json" output whose only skip is package-level
      with "no test files"
    When "centinela validate" runs
    Then the command should be reported as passed

  Scenario: An acceptance run with everything executed and nothing skipped passes
    Given a validate command classified as an acceptance execution
    And the command exits zero and reports "5 scenarios (5 passed)"
    When "centinela validate" runs
    Then the command should be reported as passed
    And no skip warning should be emitted

  Scenario: An unparseable acceptance report is a warning, never a pass and never a failure
    Given a validate command classified as an acceptance execution
    And the command exits zero with output matching no supported summary shape
    When "centinela validate" runs
    Then the command should be reported as a warning
    And the warning should state that skips could not be detected
    And the warning should name a remedy that produces a parseable report
    And the run should exit zero

  Scenario: A summary-shaped line inside a test's own output is not a skip verdict
    Given a validate command classified as an acceptance execution
    And the command exits zero and its output contains the text
      "2 scenarios (1 skipped)" inside a log line rather than as a run summary
    When "centinela validate" runs
    Then the command should not be reported as failed on the strength of that text

  Scenario: A failing exit code is reported as the failure, not as a skip verdict
    Given a validate command classified as an acceptance execution
    And the command exits non-zero and also reports "3 scenarios (1 skipped)"
    When "centinela validate" runs
    Then the command should be reported as failed for its non-zero exit
    And the failure message should not claim the failure was caused by skips

  Scenario: A truncated acceptance report is undetermined, not a skip verdict
    Given a validate command classified as an acceptance execution
    And the command times out leaving a partial summary in its output
    When "centinela validate" runs
    Then the command should be reported as failed for the timeout
    And no skip verdict should be derived from the partial summary

  Scenario: A skipping unit or integration command is never failed by skip detection
    Given a validate command that is NOT classified as an acceptance execution
    And the command exits zero and reports skipped tests
    When "centinela validate" runs
    Then the command should be reported as passed
    And no skip analysis should be performed for it

  Scenario: A project with no configured validate commands is unaffected
    Given a project with an empty "validate.commands" list
    When "centinela validate" runs
    Then no validate commands should be executed
    And no acceptance skip analysis should be performed

  Scenario: The acceptance classifier is unchanged by this feature
    Given the set of command strings the previous classifier accepted and rejected
    When each is classified by the per-command acceptance predicate
    Then every verdict should be identical to the previous classifier's verdict

  # ── C. The skip policy is configurable and defaults to on ─────────────────

  Scenario: An absent policy key behaves as fail
    Given a project whose "[validate]" table sets no "acceptance_skip_policy"
    And an acceptance command reporting a skipped scenario
    When "centinela validate" runs
    Then the command should be reported as failed

  Scenario: The warn policy surfaces the skip without failing the run
    Given a project configuring "acceptance_skip_policy" as "warn"
    And an acceptance command reporting a skipped scenario
    When "centinela validate" runs
    Then the command should be reported as a warning naming the skipped count
    And the run should exit zero

  Scenario: The off policy restores the previous exit-code-only behavior
    Given a project configuring "acceptance_skip_policy" as "off"
    And an acceptance command reporting a skipped scenario
    When "centinela validate" runs
    Then the command should be reported as passed

  Scenario: An unknown policy value is a configuration error
    Given a project configuring "acceptance_skip_policy" as "maybe"
    When the configuration is loaded
    Then loading should fail with an error naming the supported values
    And the value should not be silently normalized to the default

  # ── D. The claim verifier applies the same rule ───────────────────────────

  Scenario: The claim verifier rejects a tests-pass claim whose acceptance run skipped
    Given a "tests-pass" claim to verify
    And the acceptance-classified validate command exits zero while reporting a
      skipped scenario
    When the claim is verified
    Then the check status should be FAIL
    And the detail should name the skipped count
    And the status should be distinct from PASS

  Scenario: The reused single-run suite outcome is analysed for acceptance skips
    Given a prior suite run outcome reused by the claim verifier instead of a
      fresh execution
    And at least one configured validate command is classified as an acceptance
      execution
    And the reused output reports a skipped scenario
    When the "tests-pass" claim is verified
    Then the check status should be FAIL

  Scenario: The claim verifier does not invent a failure it cannot prove
    Given a prior suite run outcome whose output matches no supported summary shape
    When the "tests-pass" claim is verified
    Then the check status should be PASS
    And the detail should state that skip detection was not possible

  # ── E. roadmap-quality shape errors name the shape ────────────────────────

  Scenario: A missing scores object is reported as a missing object, not a bad number
    Given a roadmap quality entry for a known feature with no "scores" field
    When the roadmap quality report is validated
    Then the error should name the feature and the "scores" field
    And the error should list the expected score field names
    And the error should not say scores must be between 1 and 10

  Scenario: A wrongly typed scores value is reported as a shape error
    Given a roadmap quality entry whose "scores" value is an array
    When the roadmap quality report is validated
    Then the error should name the feature and the "scores" field
    And the error should state what shape was expected
    And the error should not say scores must be between 1 and 10

  Scenario: A missing individual score field names that field
    Given a roadmap quality entry whose "scores" object omits "effortEstimation"
    When the roadmap quality report is validated
    Then the error should name "effortEstimation"
    And the error should not say scores must be between 1 and 10

  Scenario: A non-integer score is reported as a type error naming the field
    Given a roadmap quality entry whose "overall" score is the number 9.0
    When the roadmap quality report is validated
    Then the error should name "overall"
    And the error should describe the expected type
    And the error should not say scores must be between 1 and 10

  Scenario: A genuinely out-of-range score still produces the range error
    Given a roadmap quality entry whose "userValue" score is 11
    When the roadmap quality report is validated
    Then the error should say the score must be between 1 and 10
    And the error should name the field "userValue" and the value 11

  Scenario: A missing or non-array features list is reported structurally
    Given a roadmap quality report whose "features" value is absent
    When the roadmap quality report is validated
    Then the error should name the "features" field and the expected shape

  Scenario: A well-formed quality report still validates unchanged
    Given a roadmap quality report where every roadmap feature has valid scores
    When the roadmap quality report is validated
    Then validation should succeed

  # ── F. Gates report what they actually inspected ──────────────────────────

  Scenario: A file-size gate that inspected nothing reports skipped
    Given a diff-aware validate run whose changed-file set is empty
    When the file-size gate runs
    Then the gate status should be SKIP
    And the message should state that nothing was inspected
    And the message should not claim files were within the cap

  Scenario: A clean file-size pass names the cap and the exception mechanism
    Given a diff-aware run with changed source files all under the cap
    When the file-size gate runs
    Then the gate status should be PASS
    And the message should name the 100-line cap
    And the message should state that per-file exceptions are configurable

  Scenario: The file-size cap and severity are unchanged
    Given a source file of 101 lines with no justified exception
    When the file-size gate runs
    Then the gate status should be FAIL
    And the effective cap should still be 100 lines

  Scenario: A justified oversized file still passes with its own message
    Given a source file of 120 lines with a justified exception allowing 130
    When the file-size gate runs
    Then the gate status should be PASS
    And the details should name the file, its line count and the justification

  Scenario: A single configured locale is reported as a trivial parity check
    Given a project with the "json" i18n format and exactly one configured locale
    And that locale's file exists and parses
    When the i18n gate runs
    Then the gate status should be WARN
    And the message should state that one locale makes the parity check trivial
    And the message should not claim all locales have identical keys

  Scenario: A single locale with a missing file still fails
    Given a project with the "json" i18n format and exactly one configured locale
    And that locale's file is missing
    When the i18n gate runs
    Then the gate status should be FAIL

  Scenario: A single locale with a malformed file still fails
    Given a project with the "json" i18n format and exactly one configured locale
    And that locale's file is not valid JSON
    When the i18n gate runs
    Then the gate status should be FAIL

  Scenario: Two locales in sync still pass and out of sync still fail
    Given a project with the "json" i18n format and two configured locales
    When their key sets are identical
    Then the gate status should be PASS
    And when a key is present in only one of them
    Then the gate status should be FAIL

  Scenario: The gettext path is unaffected by the single-locale warning
    Given a project with the "gettext" i18n format and exactly one configured locale
    And every message in that locale has a translation
    When the i18n gate runs
    Then the gate status should be PASS
    And the single-locale parity warning should not be emitted

  Scenario: An i18n gate filtered out of the diff scope reports skipped
    Given a diff-aware validate run whose changed-file set contains no locale file
    When the i18n gate runs
    Then the gate status should be SKIP
    And the message should state that nothing was inspected

  Scenario: A skipped gate never turns a green run red
    Given a validate run whose only non-passing gate results are SKIP or WARN
    When the overall verdict is computed
    Then the run should be reported as passing
    And the skipped gates should not be counted as passed
