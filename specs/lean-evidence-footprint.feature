Feature: Lean evidence footprint
  As a Centinela maintainer
  I want machine-only workflow evidence kept out of git
  So that PRs and the repo stay small while readable narratives are preserved

  Background:
    Given a repository governed by Centinela
    And the .gitignore contains the lean-evidence-footprint patterns

  Scenario: Per-role evidence JSON is ignored
    Given a workflow writes ".workflow/demo-big-thinker.json"
    When I check whether git ignores it
    Then the file is ignored
    And it does not appear in "git status"

  Scenario: Advisory lock files are ignored
    Given a workflow writes ".workflow/demo-big-thinker.lock"
    When I check whether git ignores it
    Then the file is ignored

  Scenario: Per-feature root state JSON is ignored
    Given a workflow writes ".workflow/demo.json"
    When I check whether git ignores it
    Then the file is ignored

  Scenario: The roadmap is never ignored
    Given the file ".workflow/roadmap.json" exists
    When I check whether git ignores it
    Then the file is NOT ignored
    And it remains tracked

  Scenario: Readable role narratives stay tracked
    Given a workflow writes ".workflow/demo-big-thinker.md"
    When I check whether git ignores it
    Then the file is NOT ignored

  Scenario: Already-committed plumbing is untracked retroactively
    Given the index previously tracked ".workflow/old-feature-qa-senior.json"
    And the index previously tracked ".workflow/old-feature-qa-senior.lock"
    When the cleanup removes them from the index
    Then "git ls-files .workflow/*.json" excludes them
    And "git ls-files .workflow/*.lock" returns nothing
    And the local files still exist on disk

  Scenario: The workflow still validates with untracked local evidence
    Given evidence JSON exists locally but is gitignored
    When I run "centinela validate"
    Then it passes its gate checks
    And "centinela complete" still reads the local evidence
