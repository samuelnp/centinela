Feature: Governed project memory
  As a developer using Centinela
  I want hard-won project knowledge harvested at step boundaries and recalled when I plan
  So that lessons, gate verdicts, and decisions are not lost between features

  Background:
    Given a Centinela-governed project with memory enabled

  # --- Capture: source 1 — edge-case lessons (tests step) ---

  Scenario: SC-01 Edge-case lessons are captured when the tests step completes
    Given a feature "alpha" with a valid edge-cases artifact at ".workflow/alpha-edge-cases.md"
    When I complete the tests step for "alpha"
    Then a ledger entry of type "lesson" exists for "alpha"
    And the entry frontmatter has sourceArtifact ".workflow/alpha-edge-cases.md"
    And the entry file is written under ".workflow/memory/entries/"

  # --- Capture: source 2 — gatekeeper verdict (validate step) ---

  Scenario: SC-02 Gatekeeper verdict is captured when the validate step completes
    Given a feature "alpha" with a gatekeeper report at ".workflow/alpha-gatekeeper.md"
    When I complete the validate step for "alpha"
    Then a ledger entry of type "verdict" exists for "alpha"
    And the entry frontmatter has sourceArtifact ".workflow/alpha-gatekeeper.md"

  # --- Capture: source 3 — decisions from plan step ---

  Scenario: SC-03 Each decision bullet is captured as a separate entry when plan step completes
    Given a feature "alpha" whose brief contains a "## Decisions" section with 3 bullet items
    When I complete the plan step for "alpha"
    Then 3 ledger entries of type "decision" exist for "alpha"
    And each entry links back to the feature brief or plan file

  Scenario: SC-04 No decision entries are created when the brief has no Decisions section
    Given a feature "alpha" whose brief has no "## Decisions" section
    When I complete the plan step for "alpha"
    Then no ledger entries of type "decision" exist for "alpha"
    And the step completes successfully without error

  # --- Idempotence ---

  Scenario: SC-05 Capture is idempotent — re-completing a step does not duplicate entries
    Given a feature "alpha" whose tests step has already been captured
    And the ledger contains 1 lesson entry for "alpha"
    When I complete the tests step for "alpha" again
    Then the ledger still contains exactly 1 lesson entry for "alpha"
    And no new files are written under ".workflow/memory/entries/"

  # --- Non-blocking capture failures ---

  Scenario: SC-06 A missing source artifact does not block step completion
    Given a feature "alpha" with no edge-cases artifact at ".workflow/alpha-edge-cases.md"
    When I complete the tests step for "alpha"
    Then the step completes successfully
    And no lesson entry exists for "alpha"
    And a warning is emitted referencing the missing artifact

  Scenario: SC-07 A malformed source artifact does not block step completion
    Given a feature "alpha" with a malformed edge-cases artifact at ".workflow/alpha-edge-cases.md"
    When I complete the tests step for "alpha"
    Then the step completes successfully
    And no lesson entry exists for "alpha"
    And a warning is emitted referencing the malformed artifact

  # --- Recall: plan step injection ---

  Scenario: SC-08 Relevant memory is recalled into the plan step context
    Given the ledger contains entries tagged "coverage" relevant to feature "beta"
    When I start the plan step for "beta"
    Then the plan advisor context includes a memory block with those entries
    And the injected entries are fewer than or equal to the recall_max_entries limit
    And the total injected bytes do not exceed the recall_max_bytes limit

  Scenario: SC-09 Recall uses deterministic ranking — dependency match beats shared tags beats recency
    Given the ledger contains:
      | feature   | type    | tags       | createdAt  |
      | dep-feat  | lesson  | coverage   | 2026-01-01 |
      | other-a   | lesson  | coverage   | 2026-01-05 |
      | other-b   | verdict | unrelated  | 2026-01-10 |
    And "beta" declares "dep-feat" as a dependency
    And "beta" shares the tag "coverage" with "other-a"
    When the plan advisor builds the memory context for "beta"
    Then the entry from "dep-feat" ranks first
    And the entry from "other-a" ranks second
    And the entry from "other-b" ranks last

  Scenario: SC-10 Recall injects nothing and raises no error on an empty ledger
    Given an empty ledger
    When I start the plan step for "beta"
    Then no memory block is present in the plan advisor context
    And no error is raised

  Scenario: SC-11 Recall caps the injected slice by count and byte budget
    Given the ledger contains 20 entries all tagged "perf" relevant to feature "gamma"
    And recall_max_entries is configured to 5
    And recall_max_bytes is configured to 1024
    When I start the plan step for "gamma"
    Then at most 5 entries are injected
    And the total size of injected entries does not exceed 1024 bytes

  # --- Config: memory disabled ---

  Scenario: SC-12 Memory disabled makes capture and recall no-ops
    Given a Centinela-governed project with memory disabled
    When I complete the tests step for "alpha"
    And I start the plan step for "beta"
    Then no ledger entries are created
    And no memory block is present in the plan advisor context

  # --- Concurrency safety ---

  Scenario: SC-13 Concurrent worktree completes do not clobber each other
    Given feature "alpha" completes its tests step in worktree-1
    And feature "bravo" completes its tests step in worktree-2 at the same time
    When both captures finish
    Then a lesson entry exists for "alpha"
    And a lesson entry exists for "bravo"
    And all entry files under ".workflow/memory/entries/" are intact
