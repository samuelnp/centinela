Feature: Project test coverage above 90%
  As a maintainer
  I want comprehensive automated tests across core packages
  So the project stays safe to evolve

  Scenario: Coverage baseline is identified
    Given current project tests
    When full-package coverage is measured
    Then low-coverage packages are listed for test expansion

  Scenario: Core internal packages are covered
    Given packages with missing statement coverage
    When new tests are added for their branches
    Then those packages contribute measurable coverage

  Scenario: Coverage target is validated
    Given the full test suite passes
    When coverage is measured with the standard command
    Then total statement coverage is greater than 90%
