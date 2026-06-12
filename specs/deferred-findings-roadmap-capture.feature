Feature: Deferred findings roadmap capture
  Out-of-scope discoveries and deferred fixes recorded by workflow agents
  must land in a machine-readable ledger and be promotable into the roadmap
  without ever breaking `centinela roadmap validate`.

  # ---------------------------------------------------------------------------
  # Slice 1 — ledger core + roadmap defer
  # ---------------------------------------------------------------------------

  Scenario: Happy-path defer creates a ledger entry with required fields
    Given a project with an active workflow for feature "hook-timeout-config"
    And the roadmap has no feature named "hook-timeout-config"
    And no ledger entry exists for slug "hook-timeout-config"
    When the agent runs:
      centinela roadmap defer hook-timeout-config --summary "Prewrite hook timeout is hardcoded; should be configurable" --source hook-timeout-config/senior-engineer
    Then the file .workflow/deferred/hook-timeout-config.json is created
    And it contains slug "hook-timeout-config"
    And it contains summary "Prewrite hook timeout is hardcoded; should be configurable"
    And it contains source.feature "hook-timeout-config" and source.role "senior-engineer"
    And it contains status "open"
    And it contains a non-empty createdAt timestamp
    And the command exits with code 0

  Scenario: Defer rejects a slug that already exists in the ledger
    Given a ledger entry exists at .workflow/deferred/hook-timeout-config.json with status "open"
    When the agent runs:
      centinela roadmap defer hook-timeout-config --summary "Duplicate attempt"
    Then the command exits with a non-zero code
    And the output contains "already exists" or "collision"
    And the existing ledger file is unchanged

  Scenario: Defer rejects a slug that matches an existing roadmap feature name
    Given roadmap.json contains a feature named "enforce-coverage-in-validate"
    When the agent runs:
      centinela roadmap defer enforce-coverage-in-validate --summary "Raise the bar further"
    Then the command exits with a non-zero code
    And the output contains "already a roadmap feature"
    And no ledger file is created

  Scenario: Defer with an empty summary is rejected before any file is written
    Given no ledger entry exists for slug "empty-summary-test"
    When the agent runs:
      centinela roadmap defer empty-summary-test --summary ""
    Then the command exits with a non-zero code
    And the output contains "summary" and "required" or "empty"
    And no file is created at .workflow/deferred/empty-summary-test.json

  Scenario: Defer with an invalid slug is rejected
    Given no ledger entry exists for slug "bad slug!"
    When the agent runs:
      centinela roadmap defer "bad slug!" --summary "Something"
    Then the command exits with a non-zero code
    And the output names the invalid slug and the required format

  Scenario: Source flag is optional; when inside a worktree the feature is auto-detected
    Given the shell CWD is inside .worktrees/auto-source-feat
    And the workflow file .workflow/auto-source-feat.json has currentStep "code"
    And no ledger entry exists for slug "needs-source-detection"
    When the agent runs without --source:
      centinela roadmap defer needs-source-detection --summary "Auto-sourced finding"
    Then the command exits with code 0
    And .workflow/deferred/needs-source-detection.json contains source.feature "auto-source-feat"

  Scenario: Source flag is required when run outside any worktree and CWD detection yields nothing
    Given the shell CWD is the repo root (not inside any .worktrees/ directory)
    And no active workflow is detectable from the CWD
    And no ledger entry exists for slug "no-source-slug"
    When the agent runs without --source:
      centinela roadmap defer no-source-slug --summary "Root-level finding"
    Then the command exits with code 0
    And .workflow/deferred/no-source-slug.json is created with source omitted or null

  # ---------------------------------------------------------------------------
  # Slice 2 — visibility
  # ---------------------------------------------------------------------------

  Scenario: Deferred findings count and list are shown in centinela roadmap output
    Given the ledger contains two open entries: "hook-timeout-config" and "doc-sync-reminder"
    When the agent runs:
      centinela roadmap
    Then the output contains "Deferred findings" or "deferred"
    And the output mentions both slugs "hook-timeout-config" and "doc-sync-reminder"
    And the output shows the count "2"

  Scenario: Deferred findings section is hidden when there are no open entries
    Given the ledger contains no open entries
    When the agent runs:
      centinela roadmap
    Then the output does not contain a deferred-findings section or count

  Scenario: defer --list prints all open ledger entries as machine-readable JSON
    Given the ledger contains open entries for slugs "alpha-finding" and "beta-finding"
    And "gamma-finding" exists in the ledger with status "promoted"
    When the agent runs:
      centinela roadmap defer --list
    Then the command exits with code 0
    And the output is valid JSON or a structured listing
    And the output includes "alpha-finding" and "beta-finding"
    And the output does not include "gamma-finding"

  # ---------------------------------------------------------------------------
  # Slice 3 — roadmap promote
  # ---------------------------------------------------------------------------

  Scenario: Happy-path promote appends to all three roadmap artifacts
    Given the ledger has an open entry for slug "hook-timeout-config"
    And roadmap.json has a phase named "Phase 5 — Operability & DX"
    And roadmap-analysis.json and roadmap-quality.json are valid and consistent
    When the agent runs:
      centinela roadmap promote hook-timeout-config --phase "Phase 5 — Operability & DX" --summary "Make hook timeout configurable via centinela.toml" --scores 9,9,8,7,9,9
    Then the command exits with code 0
    And roadmap.json contains "hook-timeout-config" as a feature in "Phase 5 — Operability & DX"
    And roadmap-analysis.json contains an entry with name "hook-timeout-config"
    And roadmap-quality.json contains a scored entry for "hook-timeout-config" with overall >= 9
    And ROADMAP.md sync reminder is printed
    And centinela roadmap validate passes

  Scenario: Promote preserves unknown JSON fields in analysis and quality artifacts
    Given roadmap-analysis.json contains entries with a legacy "dependsOn" field not in the Go struct
    And the ledger has an open entry for slug "preserve-fields-test"
    When the agent promotes "preserve-fields-test" into an existing phase with valid scores
    Then the existing entries in roadmap-analysis.json still contain their "dependsOn" fields
    And no previously-present field is dropped from roadmap-analysis.json or roadmap-quality.json

  Scenario: Promote marks the ledger entry status as promoted
    Given the ledger has an open entry for slug "hook-timeout-config"
    When the agent promotes "hook-timeout-config" into an existing phase with valid scores
    Then .workflow/deferred/hook-timeout-config.json has status "promoted"
    And it contains a non-empty promotedAt timestamp

  Scenario: Promote with overall score below 9 is rejected before any write
    Given the ledger has an open entry for slug "low-score-test"
    And roadmap.json has a phase named "Phase 5 — Operability & DX"
    When the agent runs:
      centinela roadmap promote low-score-test --phase "Phase 5 — Operability & DX" --summary "Something" --scores 9,9,8,7,9,7
    Then the command exits with a non-zero code
    And the output mentions overall score requirement (>= 9)
    And roadmap.json is unchanged
    And roadmap-analysis.json is unchanged
    And roadmap-quality.json is unchanged

  Scenario: Promote into a non-existent phase is rejected with known phases listed
    Given roadmap.json has phases "Phase 0: Bootstrap" and "Phase 5 — Operability & DX"
    And the ledger has an open entry for slug "phase-test"
    When the agent runs:
      centinela roadmap promote phase-test --phase "Phase 99 — Does Not Exist" --summary "Something" --scores 9,9,9,9,9,9
    Then the command exits with a non-zero code
    And the output lists the known phases including "Phase 0: Bootstrap" and "Phase 5 — Operability & DX"
    And roadmap.json is unchanged

  Scenario: Promote a slug already present as a roadmap feature is rejected cleanly
    Given roadmap.json already contains a feature named "enforce-coverage-in-validate"
    And a stale ledger entry exists for slug "enforce-coverage-in-validate" with status "open"
    When the agent runs:
      centinela roadmap promote enforce-coverage-in-validate --phase "Phase 5 — Operability & DX" --summary "Stale" --scores 9,9,9,9,9,9
    Then the command exits with a non-zero code
    And the output explains the slug is already a roadmap feature
    And the three roadmap artifacts are unchanged

  Scenario: centinela roadmap validate passes after a successful promotion
    Given a promotion of "hook-timeout-config" into "Phase 5 — Operability & DX" has just succeeded
    When the agent runs:
      centinela roadmap validate
    Then the command exits with code 0

  # ---------------------------------------------------------------------------
  # Slice 4 — prompt contract
  # ---------------------------------------------------------------------------

  Scenario: Four role prompts and their scaffold mirrors contain the Deferred Findings obligation byte-identically
    Given the files docs/architecture/big-thinker-prompt.md
    And the files docs/architecture/feature-specialist-prompt.md
    And the files docs/architecture/senior-engineer-prompt.md
    And the files docs/architecture/qa-senior-prompt.md
    And their mirrors under internal/scaffold/assets/docs/architecture/
    When each source file is compared byte-for-byte against its mirror
    Then all four pairs are byte-identical
    And each of the eight files contains the text "centinela roadmap defer"
    And each of the eight files contains the section heading "Deferred Findings"
    And each of the eight files references recording slugs or stating "none"
