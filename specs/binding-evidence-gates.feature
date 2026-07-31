Feature: Binding evidence gates
  Three evidence/artifact gates must reject exactly what they already claim to
  reject: an out-of-chain handoffTo, a malformed commands array at stamp time,
  and an unfilled changelog template. Each maps to an executable assertion
  added in the tests step (`// Acceptance:` + `// Scenario:` comments name the
  scenario below).

  # --- Defect 1: handoffTo chain -------------------------------------------

  Scenario: A rejected banana handoff
    Given a feature "demo" at the "tests" step with qa-senior evidence
    And the evidence's handoffTo is "banana"
    When "centinela complete demo" runs
    Then completion is rejected
    And the error names the expected successor derived from the workflow's
      own contract, not a hardcoded literal

  Scenario: A valid terminal handoff on an internal feature
    Given an internal (non-user-facing) feature "demo-internal" at the
      "validate" step with gatekeeper evidence
    And the docs step requires no role evidence for this feature (internal
      features ship only a changelog)
    And the evidence's handoffTo is "complete"
    When "centinela complete demo-internal" runs
    Then completion succeeds
    And no hardcoded assumption of "documentation-specialist" as the next
      role is required for this to pass

  Scenario: A valid mid-chain handoff on a legacy workflow
    Given a feature "demo-legacy" whose workflow predates the adversarial-v1
      validate contract
    And its "tests" step qa-senior evidence has handoffTo "validation-specialist"
    When "centinela complete demo-legacy" runs
    Then completion succeeds
    And the expected successor was derived from the legacy contract pin, not
      from the literal "gatekeeper" a fresh workflow would require

  Scenario: A valid same-step handoff when UX is required
    Given a user-facing feature "demo-ux" at the "code" step
    And senior-engineer evidence has handoffTo "ux-ui-specialist"
    When "centinela complete demo-ux" runs
    Then completion succeeds because ux-ui-specialist is the next required
      role WITHIN the same step, not the next step's role

  Scenario: A rejected merge-steward handoff outside {complete, user}
    Given a merge-steward evidence file for feature "demo" with handoffTo
      "orchestrator"
    When the orchestration evidence validator runs
    Then validation is rejected
    And the error states handoffTo must be "complete" or "user"

  Scenario: A valid merge-steward escalation handoff
    Given a merge-steward evidence file for feature "demo" with handoffTo
      "user"
    When the orchestration evidence validator runs
    Then validation succeeds

  # --- Defect 2: artifact stamp commands schema -----------------------------

  Scenario: A malformed commands array is rejected at stamp time
    Given a gatekeeper report whose centinela:verification block has a
      commands entry missing "exitCode"
    When "centinela artifact stamp demo" runs
    Then the stamp is rejected
    And the error names the malformed entry, not a downstream reader failure

  Scenario: A commands entry with an empty argv is rejected at stamp time
    Given a gatekeeper report whose centinela:verification block has a
      commands entry with argv: []
    When "centinela artifact stamp demo" runs
    Then the stamp is rejected

  Scenario: A well-formed commands array is accepted at stamp time
    Given a gatekeeper report whose centinela:verification block has a
      commands entry {"argv": ["centinela", "validate"], "exitCode": 0}
    When "centinela artifact stamp demo" runs
    Then the stamp succeeds
    And the revision and treeDigest are recorded
    And the commands array is preserved verbatim

  # --- Defect 3: changelog template placeholder -----------------------------

  Scenario: An unfilled changelog template fails the docs gate
    Given ".workflow/demo-changelog.md" contains the literal scaffolded stub
      "- <FILL: type>: <FILL: one-line summary of the change>"
    When the docs step artifact gate runs for "demo"
    Then validation is rejected
    And the error says to replace the <FILL: ...> placeholder with a real
      one-line summary

  Scenario: A filled-in changelog passes the docs gate
    Given ".workflow/demo-changelog.md" contains "- feat: bind the evidence
      gates so a stub can no longer pass"
    When the docs step artifact gate runs for "demo"
    Then validation succeeds
