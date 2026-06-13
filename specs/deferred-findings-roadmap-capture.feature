Feature: Deferred findings roadmap capture
  Out-of-scope discoveries and deferred fixes produced by workflow agents
  must land in a validate-exempt Backlog phase in roadmap.json, be visible
  in centinela roadmap output, and be promotable into a real phase via an
  honest quality-evaluator scoring pass — without ever breaking
  `centinela roadmap validate` for real features.

  # ---------------------------------------------------------------------------
  # Slice 1 — roadmap defer: happy path
  # ---------------------------------------------------------------------------

  Scenario: Happy-path defer appends a Backlog entry with all required fields
    Given roadmap.json contains phases "Phase 0: Bootstrap" and "Phase 5 — Operability & DX"
    And no feature named "hook-timeout-config" exists in any phase
    When the agent runs:
      centinela roadmap defer hook-timeout-config --summary "Prewrite hook timeout is hardcoded; should be configurable" --source deferred-findings-roadmap-capture/senior-engineer
    Then the command exits with code 0
    And roadmap.json contains a phase named "Backlog" as the last phase
    And the Backlog phase contains an entry with name "hook-timeout-config"
    And that entry contains summary "Prewrite hook timeout is hardcoded; should be configurable"
    And that entry contains source.feature "deferred-findings-roadmap-capture" and source.role "senior-engineer"
    And that entry contains a non-empty deferredAt timestamp in RFC3339 format
    And all previously existing entries in other phases are byte-identical to before the command ran

  Scenario: Defer appends to an existing Backlog phase without disturbing prior entries
    Given roadmap.json already contains a Backlog phase with one entry named "prior-finding"
    And no feature named "new-finding" exists in any phase
    When the agent runs:
      centinela roadmap defer new-finding --summary "Another deferred finding"
    Then the command exits with code 0
    And the Backlog phase contains both "prior-finding" and "new-finding"
    And the "prior-finding" entry is byte-identical to its pre-command state

  # ---------------------------------------------------------------------------
  # Slice 1 — roadmap defer: validation rejections
  # ---------------------------------------------------------------------------

  Scenario: Defer with an empty summary is rejected before any write
    Given no feature named "empty-summary-test" exists in any phase
    When the agent runs:
      centinela roadmap defer empty-summary-test --summary ""
    Then the command exits with a non-zero code
    And the output contains the word "summary" and one of "required" or "empty"
    And roadmap.json is unchanged

  Scenario: Defer rejects a slug that already exists in the Backlog phase
    Given roadmap.json contains a Backlog phase with an entry named "hook-timeout-config"
    When the agent runs:
      centinela roadmap defer hook-timeout-config --summary "Duplicate attempt"
    Then the command exits with a non-zero code
    And the output indicates a slug collision
    And roadmap.json is unchanged

  Scenario: Defer rejects a slug that already exists in a non-Backlog phase
    Given roadmap.json contains a feature named "enforce-coverage-in-validate" in a non-Backlog phase
    When the agent runs:
      centinela roadmap defer enforce-coverage-in-validate --summary "Raise the bar further"
    Then the command exits with a non-zero code
    And the output indicates a slug collision
    And no Backlog entry is created

  Scenario: Defer with an invalid slug is rejected before any write
    When the agent runs:
      centinela roadmap defer "bad slug!" --summary "Something"
    Then the command exits with a non-zero code
    And the output names the invalid slug and the required format
    And roadmap.json is unchanged

  # ---------------------------------------------------------------------------
  # Slice 1 — roadmap defer: source resolution
  # ---------------------------------------------------------------------------

  Scenario: Defer auto-resolves --source from worktree CWD when flag is omitted
    Given the shell CWD is inside .worktrees/auto-source-feat
    And no feature named "needs-source-detection" exists in any phase
    When the agent runs without --source:
      centinela roadmap defer needs-source-detection --summary "Auto-sourced finding"
    Then the command exits with code 0
    And the Backlog entry for "needs-source-detection" contains source.feature "auto-source-feat"

  Scenario: Defer outside a worktree with no --source creates entry without source field
    Given the shell CWD is the repo root and not inside any .worktrees/ directory
    And no feature named "no-source-slug" exists in any phase
    When the agent runs without --source:
      centinela roadmap defer no-source-slug --summary "Root-level finding"
    Then the command exits with code 0
    And the Backlog entry for "no-source-slug" contains no source field

  # ---------------------------------------------------------------------------
  # Slice 2 — centinela roadmap rendering
  # ---------------------------------------------------------------------------

  Scenario: Backlog findings are shown in centinela roadmap output when present
    Given roadmap.json contains a Backlog phase with entries "hook-timeout-config" and "doc-sync-reminder"
    And each entry has a summary
    When the agent runs:
      centinela roadmap
    Then the output contains a Backlog section
    And the Backlog section mentions "hook-timeout-config" and "doc-sync-reminder"
    And each entry is shown with its slug and summary

  Scenario: Backlog section is absent from centinela roadmap output when Backlog phase is missing
    Given roadmap.json contains no Backlog phase
    When the agent runs:
      centinela roadmap
    Then the output does not contain any Backlog section

  Scenario: Backlog section is absent when Backlog phase exists but contains no entries
    Given roadmap.json contains a Backlog phase with no features
    When the agent runs:
      centinela roadmap
    Then the output does not contain any Backlog section

  Scenario: Backlog features do not appear in centinela roadmap ready output
    Given roadmap.json contains a Backlog phase with entry "backlog-finding"
    And "backlog-finding" has no dependsOn and would otherwise qualify as ready
    When the agent runs:
      centinela roadmap ready
    Then the output does not list "backlog-finding"

  # ---------------------------------------------------------------------------
  # Slice 2 — validate exemption
  # ---------------------------------------------------------------------------

  Scenario: roadmap validate passes when Backlog entries have no analysis or quality coverage
    Given roadmap.json contains a non-Backlog phase with feature "real-feature"
    And roadmap.json contains a Backlog phase with entry "backlog-finding"
    And roadmap-analysis.json covers "real-feature" but not "backlog-finding"
    And roadmap-quality.json covers "real-feature" but not "backlog-finding"
    When the agent runs:
      centinela roadmap validate
    Then the command exits with code 0

  Scenario: roadmap validate still fails when a non-Backlog feature is missing analysis coverage
    Given roadmap.json contains a non-Backlog phase with feature "uncovered-feature"
    And roadmap.json contains a Backlog phase with entry "backlog-finding"
    And roadmap-analysis.json does NOT cover "uncovered-feature"
    When the agent runs:
      centinela roadmap validate
    Then the command exits with a non-zero code
    And the error message names "uncovered-feature" as missing from analysis

  Scenario: A phase named similarly to Backlog but not matching is NOT exempt from validate
    Given roadmap.json contains a phase named "Pre-Backlog Work" with feature "borderline-feature"
    And roadmap-analysis.json does NOT cover "borderline-feature"
    When the agent runs:
      centinela roadmap validate
    Then the command exits with a non-zero code
    And the error message names "borderline-feature" as missing from analysis

  # ---------------------------------------------------------------------------
  # Slice 2 — start guard
  # ---------------------------------------------------------------------------

  Scenario: centinela start refuses a Backlog feature with a promote-first error
    Given roadmap.json contains a Backlog phase with entry "backlog-finding"
    When the agent runs:
      centinela start backlog-finding
    Then the command exits with a non-zero code
    And the output tells the operator to promote the finding first before starting it

  # ---------------------------------------------------------------------------
  # Slice 3 — roadmap promote: no --scores path (evaluator context)
  # ---------------------------------------------------------------------------

  Scenario: Promote without --scores prints evaluator context and writes nothing
    Given roadmap.json contains a Backlog phase with entry "hook-timeout-config"
    And the "hook-timeout-config" entry has summary "Prewrite hook timeout is hardcoded" and source "deferred-findings-roadmap-capture/senior-engineer"
    And roadmap.json contains a phase "Phase 5 — Operability & DX"
    When the agent runs:
      centinela roadmap promote hook-timeout-config --phase "Phase 5 — Operability & DX"
    Then the command exits with code 0
    And the output contains the finding's name "hook-timeout-config"
    And the output contains the finding's summary "Prewrite hook timeout is hardcoded"
    And the output contains the finding's source
    And the output contains the target phase "Phase 5 — Operability & DX"
    And the output states the threshold is 9
    And the output describes all six scoring dimensions: acceptanceCriteria, userValue, definitionClarity, dependencies, effortEstimation, overall
    And the output contains a re-invocation line of the form:
      centinela roadmap promote hook-timeout-config --phase "Phase 5 — Operability & DX" --scores ac,uv,dc,dep,ee,overall
    And the output instructs the operator to run a quality-evaluator pass and re-invoke with --scores
    And roadmap.json is unchanged
    And roadmap-analysis.json is unchanged
    And roadmap-quality.json is unchanged

  # ---------------------------------------------------------------------------
  # Slice 3 — roadmap promote: scored path (happy path)
  # ---------------------------------------------------------------------------

  Scenario: Promote with valid --scores moves entry from Backlog to target phase and appends artifacts
    Given roadmap.json contains a Backlog phase with entry "hook-timeout-config" with summary "Prewrite hook timeout is hardcoded; should be configurable"
    And the "hook-timeout-config" entry has source "deferred-findings-roadmap-capture/senior-engineer" and a deferredAt timestamp
    And roadmap.json contains a phase "Phase 5 — Operability & DX"
    And roadmap-analysis.json and roadmap-quality.json are valid and consistent
    When the agent runs:
      centinela roadmap promote hook-timeout-config --phase "Phase 5 — Operability & DX" --scores 9,9,8,7,9,9
    Then the command exits with code 0
    And roadmap.json contains "hook-timeout-config" as a feature in "Phase 5 — Operability & DX"
    And the "hook-timeout-config" entry in "Phase 5 — Operability & DX" has no summary, source, or deferredAt fields
    And roadmap.json no longer contains "hook-timeout-config" in the Backlog phase
    And roadmap-analysis.json contains an entry with name "hook-timeout-config"
    And roadmap-quality.json contains a scored entry for "hook-timeout-config" with overall 9
    And the quality entry's summary matches the original deferred finding's summary
    And roadmap-analysis.md contains a provenance bullet for "hook-timeout-config" referencing the original source and deferredAt
    And roadmap-quality.md contains a provenance bullet for "hook-timeout-config" referencing the original source and deferredAt
    And the command output includes a ROADMAP.md sync reminder
    And centinela roadmap validate exits with code 0 afterwards

  Scenario: Promote preserves unknown JSON fields on untouched entries (raw-preserving I/O)
    Given roadmap-analysis.json contains existing entries with a custom field not in the Go struct
    And roadmap.json contains a Backlog phase with entry "preserve-fields-test"
    And roadmap.json contains a phase "Phase 5 — Operability & DX"
    When the agent promotes "preserve-fields-test" into "Phase 5 — Operability & DX" with valid scores
    Then the existing entries in roadmap-analysis.json still contain their custom fields
    And no previously-present field is dropped from roadmap-analysis.json or roadmap-quality.json
    And no previously-present field is dropped from roadmap.json entries in other phases

  # ---------------------------------------------------------------------------
  # Slice 3 — roadmap promote: validation rejections (zero writes)
  # ---------------------------------------------------------------------------

  Scenario: Promote with overall score below 9 is rejected before any write
    Given roadmap.json contains a Backlog phase with entry "low-score-test"
    And roadmap.json contains a phase "Phase 5 — Operability & DX"
    When the agent runs:
      centinela roadmap promote low-score-test --phase "Phase 5 — Operability & DX" --scores 9,9,8,7,9,7
    Then the command exits with a non-zero code
    And the output states the overall score must be at least 9
    And roadmap.json is unchanged
    And roadmap-analysis.json is unchanged
    And roadmap-quality.json is unchanged

  Scenario: Promote with any dimension score outside 1-10 is rejected before any write
    Given roadmap.json contains a Backlog phase with entry "bad-score-test"
    And roadmap.json contains a phase "Phase 5 — Operability & DX"
    When the agent runs:
      centinela roadmap promote bad-score-test --phase "Phase 5 — Operability & DX" --scores 11,9,9,9,9,9
    Then the command exits with a non-zero code
    And the output states each score must be between 1 and 10
    And roadmap.json is unchanged

  Scenario: Promote into a non-existent phase is rejected with known phases listed
    Given roadmap.json has phases "Phase 0: Bootstrap" and "Phase 5 — Operability & DX" and a Backlog phase
    And the Backlog phase contains entry "phase-test"
    When the agent runs:
      centinela roadmap promote phase-test --phase "Phase 99 — Does Not Exist" --scores 9,9,9,9,9,9
    Then the command exits with a non-zero code
    And the output lists the known non-Backlog phases including "Phase 0: Bootstrap" and "Phase 5 — Operability & DX"
    And roadmap.json is unchanged

  Scenario: Promote a slug not in the Backlog phase is rejected cleanly
    Given roadmap.json contains no Backlog entry named "not-in-backlog"
    When the agent runs:
      centinela roadmap promote not-in-backlog --phase "Phase 5 — Operability & DX" --scores 9,9,9,9,9,9
    Then the command exits with a non-zero code
    And the output states the slug is not a Backlog finding
    And roadmap.json is unchanged

  Scenario: Promote with a malformed --scores CSV is rejected before any write
    Given roadmap.json contains a Backlog phase with entry "malformed-scores-test"
    And roadmap.json contains a phase "Phase 5 — Operability & DX"
    When the agent runs:
      centinela roadmap promote malformed-scores-test --phase "Phase 5 — Operability & DX" --scores 9,9,9
    Then the command exits with a non-zero code
    And the output states that --scores requires exactly six comma-separated integers
    And roadmap.json is unchanged

  # ---------------------------------------------------------------------------
  # Slice 4 — prompt contract
  # ---------------------------------------------------------------------------

  Scenario: All eight role prompts and their scaffold mirrors contain the Deferred Findings section byte-identically
    Given the following source prompt files:
      docs/architecture/big-thinker-prompt.md
      docs/architecture/feature-specialist-prompt.md
      docs/architecture/senior-engineer-prompt.md
      docs/architecture/qa-senior-prompt.md
      docs/architecture/edge-case-tester-prompt.md
      docs/architecture/ux-ui-specialist-prompt.md
      docs/architecture/validation-specialist-prompt.md
      docs/architecture/gatekeeper-prompt.md
    And each has a corresponding mirror under internal/scaffold/assets/docs/architecture/
    When each source file is compared byte-for-byte against its mirror
    Then all eight pairs are byte-identical
    And each of the sixteen files contains the section heading "Deferred Findings"
    And each of the sixteen files contains the text "centinela roadmap defer"
    And each of the sixteen files references recording slugs or stating "none"
