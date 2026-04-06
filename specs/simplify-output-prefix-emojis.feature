Feature: Simplified Centinela output prefix
  As a user
  I want a compact consistent output prefix
  So command and hook output is cleaner

  Scenario: System line uses fixed emoji prefix
    Given a rendered system line
    When output is produced for any tone
    Then the prefix should be exactly "🛡️👁️"

  Scenario: Channel and message metadata remain visible
    Given a rendered blocked or status output
    When output is produced
    Then channel and title metadata should still appear
