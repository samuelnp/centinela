Feature: brownfield adoption — record the legacy debt as the accepted baseline, once, deliberately
  As a brownfield adopter (team lead / staff engineer) turning Centinela on for a mature repo,
  or the onboarding agent driving the brownfield flow
  I want a first-class `centinela adopt` command that snapshots today's full-repo gate violations
  as the accepted audit baseline, refuses to silently overwrite an established baseline, and shows
  me the bill of accepted debt
  So that day-one `validate` is not drowned by thousands of pre-existing violations, I consciously
  accept the legacy debt, and I never widen an established baseline by accident

  # `centinela adopt` composes the already-shipped baseline machinery: it Loads the configured
  # baseline path (existence check), and if none exists Records the current full-repo Fail
  # violations across the participating gates and Saves the deterministic
  # .workflow/audit-baseline.json — the SAME file the ratchet (`centinela audit`) reads. It adds
  # three deliberate deltas over `centinela audit baseline`:
  #   1. skip-if-exists DEFAULT — refuses to overwrite an existing baseline unless --force
  #      (the opposite of `audit baseline`, which always overwrites for the ongoing ratchet).
  #   2. an adoption report — per-gate accepted-violation counts, the total, and a "you are
  #      starting with N accepted findings; ratchet to zero over time" framing.
  #   3. named flow placement — the step between `centinela roadmap brownfield` and the first
  #      `centinela start`.
  # The written baseline is byte-identical to what `audit.Record` + `audit.Save` already produce;
  # adopt adds semantics (skip rule, report), not different data. Scenario titles map 1:1 to Go
  # acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with a valid centinela.toml
    And "[gates.audit_baseline] enabled" is true unless a scenario states otherwise
    And the configured baseline path is ".workflow/audit-baseline.json"
    And the participating gates are the mechanical gates that emit per-violation Details

  # ---------------------------------------------------------------------------
  # First adoption — record the legacy debt and show the bill
  # ---------------------------------------------------------------------------

  Scenario: First adoption on a repo with pre-existing violations records the baseline and reports the accepted debt
    Given no baseline file exists at ".workflow/audit-baseline.json"
    And the repo currently has gate violations across participating gates
    When the operator runs:
      centinela adopt
    Then the command exits with code 0
    And the baseline file ".workflow/audit-baseline.json" is written
    And the baseline records a fingerprint for every current violation across participating gates
    And the report lists each participating gate with its accepted-violation count
    And the report states the total number of accepted findings
    And the report contains the ratchet-to-zero framing telling the adopter to drive the debt to zero over time

  Scenario: After adoption a fresh audit reports zero new violations so day-one validate is not drowned
    Given the operator has just run "centinela adopt" capturing the repo's current violations
    And no code has changed since adoption
    When the operator runs:
      centinela audit
    Then the command exits with code 0
    And the output reports "0 new" violations
    And every pre-existing violation is reported as baselined and tolerated
    And the output does not contain an error message or stack trace

  # ---------------------------------------------------------------------------
  # Skip-if-exists — the one-time-adoption safety default
  # ---------------------------------------------------------------------------

  Scenario: Re-running adopt when a baseline already exists is refused and leaves the file byte-unchanged
    Given a baseline file already exists at ".workflow/audit-baseline.json"
    When the operator runs:
      centinela adopt
    Then the command exits with a non-zero code
    And the output contains "baseline already exists" and instructs the operator to use --force to overwrite
    And the existing baseline file ".workflow/audit-baseline.json" is left byte-identical to before the command ran

  Scenario: Re-running adopt with --force overwrites the existing baseline and exits 0
    Given a baseline file already exists at ".workflow/audit-baseline.json"
    And the repo's current violations differ from the recorded baseline
    When the operator runs:
      centinela adopt --force
    Then the command exits with code 0
    And the baseline file ".workflow/audit-baseline.json" is rewritten to capture the repo's current violations
    And the report states the total number of accepted findings

  # ---------------------------------------------------------------------------
  # Clean repo — nothing to ratchet
  # ---------------------------------------------------------------------------

  Scenario: Adopting on a clean repo with no violations writes a zero-finding baseline and reports zero accepted findings
    Given no baseline file exists at ".workflow/audit-baseline.json"
    And the repo currently has no gate violations
    When the operator runs:
      centinela adopt
    Then the command exits with code 0
    And the baseline file ".workflow/audit-baseline.json" is written
    And the baseline contains zero violation fingerprints
    And the report states "0 accepted findings" so there is nothing to ratchet

  # ---------------------------------------------------------------------------
  # --json — machine-readable adoption verdict for the onboarding agent
  # ---------------------------------------------------------------------------

  Scenario: Adopt with --json emits a machine-readable adoption summary instead of the human report
    Given no baseline file exists at ".workflow/audit-baseline.json"
    And the repo currently has gate violations across participating gates
    When the operator runs:
      centinela adopt --json
    Then the command exits with code 0
    And the output is valid JSON
    And the JSON reports adopted true and skipped false
    And the JSON includes the total accepted-finding count and a per-gate count map
    And the JSON includes the baseline path ".workflow/audit-baseline.json"
    And the human adoption report prose is not printed

  Scenario: Adopt --json when a baseline already exists emits a skipped verdict and exits non-zero
    Given a baseline file already exists at ".workflow/audit-baseline.json"
    When the operator runs:
      centinela adopt --json
    Then the command exits with a non-zero code
    And the output is valid JSON
    And the JSON reports adopted false and skipped true
    And the existing baseline file ".workflow/audit-baseline.json" is left byte-identical to before the command ran

  # ---------------------------------------------------------------------------
  # Determinism — adopt adds semantics, not different data
  # ---------------------------------------------------------------------------

  Scenario: The baseline written by adopt is byte-identical to audit.Record plus audit.Save output
    Given no baseline file exists at ".workflow/audit-baseline.json"
    And the repo currently has gate violations across participating gates
    When the operator runs "centinela adopt"
    And a reference baseline is produced by audit.Record followed by audit.Save on the same unchanged repo
    Then the baseline file written by adopt is byte-identical to the reference baseline
    And the violation entries appear in a stable deterministic order
