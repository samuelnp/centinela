Feature: G1 file size with justified exceptions
  As a maintainer
  I want strict file size defaults with narrow justified exceptions
  So architecture quality stays high without forcing harmful splits

  Scenario: Oversized source file fails without exception
    Given a source file with 110 lines
    And no matching file size exception
    When centinela validate runs G1
    Then G1 should fail with split guidance

  Scenario: Configuration file passes with justified exception
    Given a source file with 120 lines
    And a file size exception with kind configuration and max_lines 130
    When centinela validate runs G1
    Then G1 should pass and report the justified exception

  Scenario: Exception above cap is invalid
    Given centinela.toml defines a file size exception with max_lines 140
    When configuration is loaded
    Then centinela should fail config validation
