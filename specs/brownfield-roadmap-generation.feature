Feature: brownfield roadmap generation — a deterministic draft that records already-built capability as Baseline and surfaces gaps
  As a team lead adopting Centinela on a brownfield repo that already ran centinela analyze
  I want centinela roadmap brownfield to emit a draft roadmap partitioning capability into a Baseline phase plus net-new gap phases
  So that already-built work is never re-planned, the real gaps are schedulable, and no curated roadmap.json is ever clobbered — with no LLM and byte-stable output

  Scenario: A built repo produces a draft with a Baseline phase listing already-built surfaces
    Given an inventory with behavioral packages at .workflow/analysis.json
    When the operator runs centinela roadmap brownfield
    Then the exit code is zero
    And a draft roadmap file is written to the draft path
    And the draft contains a Baseline phase identified by the baseline phase-name convention
    And the Baseline phase lists at least one already-built surface as a feature

  Scenario: The command never clobbers an existing canonical roadmap.json
    Given an inventory with behavioral packages at .workflow/analysis.json
    And an existing canonical .workflow/roadmap.json with hand-authored content
    When the operator runs centinela roadmap brownfield
    Then the exit code is zero
    And the draft roadmap is written to the draft path and not to .workflow/roadmap.json
    And the existing .workflow/roadmap.json is left byte-for-byte unchanged
    And the summary reports the draft path it wrote

  Scenario: Gap phase — reconstruct TODO confirm markers become net-new gap features
    Given an inventory whose reconstructed targets carry TODO confirm markers
    When the operator runs centinela roadmap brownfield
    Then the exit code is zero
    And the draft contains at least one net-new gap phase distinct from the Baseline phase
    And each TODO-bearing target appears as a schedulable gap feature in a gap phase

  Scenario: A user-stated goal adds a net-new gap feature
    Given an inventory with behavioral packages at .workflow/analysis.json
    When the operator runs centinela roadmap brownfield with goal "Add OAuth login"
    Then the exit code is zero
    And the draft contains a gap feature derived from the goal text "Add OAuth login"
    And the goal-derived feature lives in a gap phase and not in the Baseline phase

  Scenario: Baseline features are excluded from status counts and validate coverage
    Given a draft roadmap whose Baseline phase contains already-built features
    When the roadmap summary and the non-schedulable coverage set are computed
    Then Baseline-phase features are excluded from the status counts
    And Baseline-phase features are excluded from the validate coverage set
    And the exclusion uses the same predicate mechanism that already exempts the Backlog phase

  Scenario: Running twice on an unchanged inventory yields byte-identical draft output
    Given a fixed inventory at .workflow/analysis.json
    When the operator runs centinela roadmap brownfield twice into the same draft path
    Then both runs exit zero
    And the draft roadmap file is byte-identical between the two runs

  Scenario: Missing inventory fails with guidance and writes nothing
    Given the project directory has no analysis inventory
    When the operator runs centinela roadmap brownfield
    Then the exit code is non-zero
    And the error message tells the operator to run centinela analyze first
    And no draft roadmap file is written

  Scenario: An empty doc-only inventory yields an empty Baseline and zero gaps
    Given an inventory with no behavioral packages
    When the operator runs centinela roadmap brownfield
    Then the exit code is zero
    And the summary reports zero baseline entries and zero gaps
    And no malformed empty roadmap is produced

  Scenario: A built repo with no TODOs and no goals produces a Baseline-only draft with a hint
    Given an inventory with behavioral packages and no TODO confirm markers
    When the operator runs centinela roadmap brownfield with no goal flag
    Then the exit code is zero
    And the draft contains the Baseline phase and no gap phase
    And the summary hints that the operator may supply a goal to add net-new work

  Scenario: The summary reports baseline count gap count and draft path
    Given an inventory with behavioral packages at .workflow/analysis.json
    When the operator runs centinela roadmap brownfield
    Then the exit code is zero
    And the stdout summary reports the number of baseline entries
    And the stdout summary reports the number of gaps
    And the stdout summary reports the draft path written
