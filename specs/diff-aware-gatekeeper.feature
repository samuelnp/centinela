Feature: Diff-aware gatekeeper
  As a developer on a large repository
  I want centinela validate to report only gate violations
  introduced by the current branch, while keeping CI strict
  So that the inner dev loop is fast and actionable without
  weakening the ship gate.

  Background:
    Given a git repository with a "main" branch as the merge base
    And built-in gate G1 (file size) is enabled
    And the project uses the default diff_mode "auto"

  Scenario: Local default is diff-aware
    Given the CI environment variable is unset
    And the current branch has not modified any oversized file
    When the user runs "centinela validate"
    Then the output header should indicate diff-aware mode
    And the output header should reference the resolved diff base
    And G1 should report Pass with a "no relevant changes" message

  Scenario: CI default is full scan
    Given the CI environment variable is set to "true"
    When the user runs "centinela validate"
    Then the output header should indicate full scan mode
    And G1 should walk every source file as before

  Scenario: Branch introduces a new oversized file
    Given the current branch adds a file exceeding 100 lines
    When the user runs "centinela validate"
    Then G1 should report Fail
    And the failing details should include the newly added file
    And the failing details should not include any pre-existing
    oversized file that was not modified on this branch

  Scenario: Full scan reports historical violations
    Given the current branch adds a file exceeding 100 lines
    And there is a pre-existing oversized file on main that was
    not modified on this branch
    When the user runs "centinela validate --full"
    Then G1 should report Fail
    And the failing details should include both the newly added
    file and the pre-existing oversized file

  Scenario: Untracked file is part of the diff set
    Given the current branch contains an untracked source file
    exceeding 100 lines
    When the user runs "centinela validate"
    Then G1 should report Fail
    And the failing details should include the untracked file

  Scenario: Configurable diff base
    Given centinela.toml sets "[validate] diff_base = \"master\""
    And the repository has a "master" branch but no "main" branch
    When the user runs "centinela validate"
    Then the diff resolution should succeed against "master"
    And the header should reference "master" as the base

  Scenario: i18n gate is skipped when no locale file changed
    Given gate G11 (i18n) is enabled with a locales directory
    And the current branch has not modified any file under that
    directory
    When the user runs "centinela validate"
    Then G11 should report Pass with a "no locale changes" message

  Scenario: i18n gate runs when a locale file changed
    Given gate G11 (i18n) is enabled with a locales directory
    And the current branch modified one locale file
    When the user runs "centinela validate"
    Then G11 should run the full key-completeness comparison

  Scenario: Non-git directory degrades to full scan
    Given the current directory is not a git repository
    When the user runs "centinela validate"
    Then the output should include a notice that diff-aware
    degraded to full
    And the header should indicate full scan mode

  Scenario: Flag overrides take precedence over mode
    Given centinela.toml sets "[validate] diff_mode = \"off\""
    When the user runs "centinela validate --changed"
    Then the header should indicate diff-aware mode

  Scenario: Mutually exclusive flags are rejected
    When the user runs "centinela validate --changed --full"
    Then the command should exit with a non-zero status
    And the error message should explain that --changed and --full
    are mutually exclusive

  Scenario: User validate commands are not scoped by diff
    Given centinela.toml lists a validate command "echo hello"
    When the user runs "centinela validate --changed"
    Then the validate command should still execute exactly once
    in full
