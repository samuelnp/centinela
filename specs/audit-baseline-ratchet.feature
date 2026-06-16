Feature: audit baseline ratchet — tolerate existing violations, block only new ones
  As a maintainer adopting Centinela on a legacy codebase, a governance owner, or a
  CI gate author
  I want `centinela audit baseline` to snapshot today's mechanical-gate violations as
  tolerated debt, and `centinela audit` to re-scan and fail only on violations that are
  new relative to that baseline
  So that a team can adopt enforcement without a big-bang cleanup, pay debt down over
  time, and never let a new violation slip in

  # `centinela audit baseline` runs every participating mechanical gate in a FULL-REPO
  # scan (bypassing [validate] diff_mode), fingerprints each gate's Result.Details
  # violation entries on a stable identity (gate, path, rule — volatile numerics like
  # line counts are normalized out), and writes a committed, versioned baseline at
  # .workflow/audit-baseline.json.
  #   `centinela audit` re-scans full-repo and partitions current violations into:
  #     new       — not in the baseline                  → BLOCKING, non-zero exit
  #     baselined — in the baseline and still present     → tolerated, exit 0
  #     resolved  — in the baseline but now gone          → pruned on next baseline record
  # The ratchet only tightens: a resolved (pruned) violation, if reintroduced, is then
  # new and blocking. Behaviour is configured under [gates.audit_baseline]
  # (enabled, severity, baseline path, target gates). Defaults are safe-adoption: with
  # no baseline recorded, audit does not block. The baseline file is deterministic
  # (stable ordering) so it diffs cleanly in git. Scenario titles map 1:1 to Go
  # acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with a valid centinela.toml
    And "[gates.audit_baseline] enabled" is true unless a scenario states otherwise
    And the baseline file path is ".workflow/audit-baseline.json"
    And the participating gates are the mechanical gates that emit per-violation Details

  # ---------------------------------------------------------------------------
  # Recording a baseline
  # ---------------------------------------------------------------------------

  Scenario: Recording a baseline on a repo with existing violations captures them and exits 0
    Given the repo currently has gate violations across participating gates
    When the operator runs:
      centinela audit baseline
    Then the command exits with code 0
    And the baseline file ".workflow/audit-baseline.json" is written
    And the baseline records a fingerprint for every current violation across participating gates
    And the output reports the number of violations baselined

  Scenario: Recording a baseline on an empty repo with zero violations writes an empty baseline and exits 0
    Given the repo currently has no gate violations
    When the operator runs:
      centinela audit baseline
    Then the command exits with code 0
    And the baseline file ".workflow/audit-baseline.json" is written
    And the baseline contains zero violation fingerprints

  # ---------------------------------------------------------------------------
  # Ratchet check — no change tolerates the baselined debt
  # ---------------------------------------------------------------------------

  Scenario: Audit with no change reports all violations baselined and exits 0
    Given a baseline has been recorded capturing the repo's current violations
    And no code has changed since the baseline was recorded
    When the operator runs:
      centinela audit
    Then the command exits with code 0
    And every current violation is reported as baselined and tolerated
    And the output reports "0 new" violations
    And the output does not contain an error message or stack trace

  Scenario: Empty-baseline repo with zero violations audits clean and exits 0
    Given a baseline has been recorded on a repo with zero violations
    And the repo still has zero violations
    When the operator runs:
      centinela audit
    Then the command exits with code 0
    And the output reports "0 new" violations
    And the output reports "0 baselined" violations

  # ---------------------------------------------------------------------------
  # Ratchet check — a NEW violation blocks
  # ---------------------------------------------------------------------------

  Scenario: Introducing a new violation fails the audit and names it while baselined ones stay tolerated
    Given a baseline has been recorded capturing the repo's current violations
    And a new violation is introduced that is not present in the baseline
    When the operator runs:
      centinela audit
    Then the command exits with a non-zero code
    And the output names the new violation
    And the output reports the new violation under the "new" partition
    And the pre-existing baselined violations are still reported as tolerated
    And the baselined violations do not contribute to the non-zero exit

  Scenario: Multiple new violations are all named and the exit code is non-zero
    Given a baseline has been recorded capturing the repo's current violations
    And two distinct new violations are introduced that are not in the baseline
    When the operator runs:
      centinela audit
    Then the command exits with a non-zero code
    And the output names both new violations

  # ---------------------------------------------------------------------------
  # Ratchet only tightens — fixing prunes, reintroducing blocks
  # ---------------------------------------------------------------------------

  Scenario: Fixing a baselined violation never fails the audit
    Given a baseline has been recorded capturing the repo's current violations
    And one baselined violation is fixed in the working tree
    When the operator runs:
      centinela audit
    Then the command exits with code 0
    And the fixed violation is reported as resolved
    And no new violation is reported

  Scenario: Re-recording the baseline prunes a resolved violation so the ratchet tightens
    Given a baseline has been recorded capturing the repo's current violations
    And one baselined violation has been fixed
    When the operator runs:
      centinela audit baseline
    Then the command exits with code 0
    And the resolved violation is no longer present in the baseline file

  Scenario: A pruned violation reintroduced after re-recording is treated as new and blocks
    Given a baseline was re-recorded after a violation was fixed and pruned
    And that same violation is later reintroduced
    When the operator runs:
      centinela audit
    Then the command exits with a non-zero code
    And the reintroduced violation is reported under the "new" partition

  # ---------------------------------------------------------------------------
  # Fingerprint stability — cosmetic churn is not a new violation
  # ---------------------------------------------------------------------------

  Scenario: A baselined oversized file that grows by more lines stays the same tolerated violation
    Given a baseline captures an oversized-file violation for "src/big.go"
    And "src/big.go" is edited so it grows by additional lines but remains oversized
    When the operator runs:
      centinela audit
    Then the command exits with code 0
    And the violation for "src/big.go" is reported as baselined and tolerated
    And no new violation is reported for "src/big.go"

  Scenario: Deleting a baselined oversized file resolves its violation
    Given a baseline captures an oversized-file violation for "src/big.go"
    And "src/big.go" is deleted from the working tree
    When the operator runs:
      centinela audit
    Then the command exits with code 0
    And the violation for "src/big.go" is reported as resolved

  # ---------------------------------------------------------------------------
  # Missing baseline — safe-adoption default
  # ---------------------------------------------------------------------------

  Scenario: Audit with no baseline file reports a hint and does not block
    Given no baseline file exists at ".workflow/audit-baseline.json"
    When the operator runs:
      centinela audit
    Then the command exits with code 0
    And the output contains "no baseline; run centinela audit baseline"
    And the output does not contain an error message or stack trace

  # ---------------------------------------------------------------------------
  # Newly-enabled gate after the baseline
  # ---------------------------------------------------------------------------

  Scenario: A gate enabled after the baseline has its violations treated as new until re-recorded
    Given a baseline was recorded while one participating gate was disabled
    And that gate is subsequently enabled and reports violations
    When the operator runs:
      centinela audit
    Then the command exits with a non-zero code
    And the newly-enabled gate's violations are reported under the "new" partition

  Scenario: Re-recording the baseline after enabling a gate absorbs its violations as baselined
    Given a baseline was recorded while one participating gate was disabled
    And that gate is subsequently enabled and reports violations
    When the operator runs:
      centinela audit baseline
    And then the operator runs:
      centinela audit
    Then the second command exits with code 0
    And the once-new gate's violations are now reported as baselined

  # ---------------------------------------------------------------------------
  # Full-scan enforcement — diff-aware mode must not narrow the audit
  # ---------------------------------------------------------------------------

  Scenario: Audit scans the full repo even when diff-aware mode is enabled
    Given "[validate] diff_mode" is enabled so validate would scan only changed files
    And the repo has violations in files outside the current diff
    When the operator runs:
      centinela audit baseline
    Then the command exits with code 0
    And the baseline includes violations in files outside the current diff
    And the audit scan is not narrowed by the diff filter

  # ---------------------------------------------------------------------------
  # Determinism — the baseline file diffs cleanly in git
  # ---------------------------------------------------------------------------

  Scenario: Re-recording the baseline with no change produces a byte-identical file
    Given a baseline has been recorded capturing the repo's current violations
    And no code has changed since the baseline was recorded
    When the operator runs centinela audit baseline twice in succession
    Then the two baseline files are byte-identical
    And the violation entries appear in a stable deterministic order

  Scenario: Two audit runs on the same repo and baseline produce byte-identical output
    Given a baseline has been recorded capturing the repo's current violations
    When the operator runs centinela audit twice in succession
    Then both outputs are byte-identical

  Scenario: The baseline file records a fingerprint scheme version
    Given a baseline has been recorded capturing the repo's current violations
    When the baseline file ".workflow/audit-baseline.json" is read
    Then it contains a fingerprint-scheme version field

  # ---------------------------------------------------------------------------
  # Configuration — disabled or warn severity must not block
  # ---------------------------------------------------------------------------

  Scenario: Audit does not block when the gate is disabled in config
    Given "[gates.audit_baseline] enabled" is false
    And a baseline has been recorded and a new violation has since been introduced
    When the operator runs:
      centinela audit
    Then the command exits with code 0
    And the new violation does not cause a non-zero exit

  Scenario: Audit does not block when severity is configured to warn
    Given "[gates.audit_baseline] severity" is "warn"
    And a baseline has been recorded and a new violation has since been introduced
    When the operator runs:
      centinela audit
    Then the command exits with code 0
    And the new violation is reported as a warning rather than a blocking failure

  Scenario: A custom baseline path is honored for both record and ratchet
    Given "[gates.audit_baseline] baseline" points to a custom path
    When the operator runs:
      centinela audit baseline
    Then the baseline is written to the configured custom path
    And a subsequent "centinela audit" reads the baseline from the configured custom path
