Feature: Guided by default
  Centinela's zero-config enforcement profile becomes guided instead of strict.
  Profiles scale PROCESS — confirmation cadence, orchestration evidence, and the
  greenfield setup cascade. Profiles never scale PROOF — gates, the test suite,
  verification freshness, claim verification, the adversarial verifier's
  grounded verdict, and the production-readiness gate are identical under every
  profile. The self-graded roadmap quality threshold is deleted outright,
  because a number assigned by the party it constrains was never evidence.

  Background:
    Given a project governed by Centinela

  # ---------------------------------------------------------------
  # AC1 / AC2 / AC3 — the default flip and its blast radius
  # ---------------------------------------------------------------

  Scenario: A new workflow in a zero-config project defaults to guided
    Given the project has no centinela.toml
    And no driver model is configured
    When I start a workflow without passing a profile
    Then the workflow's effective enforcement profile is "guided"
    And its provenance is reported as the default

  Scenario: A workflow started before the flip keeps strict
    Given a workflow state file that pins no profile contract
    And the project has no centinela.toml
    When the effective enforcement profile is resolved for that workflow
    Then the effective enforcement profile is "strict"
    And no in-flight workflow changes profile as a result of the upgrade

  Scenario: An explicit per-feature profile still outranks the new default
    Given the project has no centinela.toml
    When I start a workflow with the profile flag set to "strict"
    Then the workflow's effective enforcement profile is "strict"
    And its provenance names the per-feature flag

  Scenario: An explicit global profile still outranks the new default
    Given centinela.toml sets the enforcement profile to "strict"
    When I start a workflow without passing a profile
    Then the workflow's effective enforcement profile is "strict"
    And its provenance names the global configuration

  Scenario: A driver model's capability class still outranks the new default
    Given centinela.toml declares a driver model whose capability class is limited
    When I start a workflow without passing a profile
    Then the workflow's effective enforcement profile is "strict"
    And its provenance names the driver model and its capability class

  Scenario: A capability-derived guided profile is distinguishable from the default
    Given centinela.toml declares a driver model whose capability class is capable
    When I start a workflow without passing a profile
    Then the workflow's effective enforcement profile is "guided"
    And its provenance names the driver model rather than the default

  # ---------------------------------------------------------------
  # AC4 — PROOF PARITY: the load-bearing guarantee
  # ---------------------------------------------------------------

  Scenario Outline: Gate failures block completion identically under strict and guided
    Given a project whose enforcement profile is "<profile>"
    And a validate command that fails
    When I run validate
    Then the run is refused
    And the refusal is identical to the refusal under every other profile

    Examples:
      | profile |
      | strict  |
      | guided  |

  Scenario Outline: A missing verifier report blocks the validate step under every profile
    Given a workflow on the validate step whose enforcement profile is "<profile>"
    And no gatekeeper report exists for the feature
    When I try to complete the validate step
    Then completion is refused because the gatekeeper report is missing

    Examples:
      | profile |
      | strict  |
      | guided  |

  Scenario Outline: An ungrounded verifier verdict blocks the validate step under every profile
    Given a workflow on the validate step whose enforcement profile is "<profile>"
    And a gatekeeper report with no commands-run record and no verified revision
    When I try to complete the validate step
    Then completion is refused because the verdict is not grounded

    Examples:
      | profile |
      | strict  |
      | guided  |

  Scenario Outline: A blocking production-readiness report blocks completion under every profile
    Given the production-readiness gate is enabled
    And a workflow on the validate step whose enforcement profile is "<profile>"
    And a production-readiness report whose status is BLOCKING
    When I try to complete the validate step
    Then completion is refused because production readiness is blocking

    Examples:
      | profile |
      | strict  |
      | guided  |

  Scenario Outline: A tree that satisfies every gate completes under every profile
    Given a workflow on the validate step whose enforcement profile is "<profile>"
    And every gate passes and the verifier verdict is grounded and SAFE
    When I try to complete the validate step
    Then the step completes

    Examples:
      | profile |
      | strict  |
      | guided  |

  Scenario: Guided relaxes process, and only process
    Given two identical projects, one strict and one guided
    When each reaches the validate step
    Then the guided project does not require the orchestration evidence bundle
    And the strict project does require the orchestration evidence bundle
    And both projects require the same gates and the same test suite
    And both projects require the same verification freshness checks and the same claim verification
    And both projects require the same grounded verifier verdict

  # ---------------------------------------------------------------
  # AC5 — the self-graded quality threshold is deleted
  # ---------------------------------------------------------------

  Scenario: A low self-assigned quality score no longer blocks anything
    Given a roadmap feature whose quality report scores its overall value 3
    When I validate the roadmap
    Then the roadmap is reported as valid
    And the low score is surfaced as advice rather than as a refusal

  Scenario: Promoting a feature with a low overall score succeeds
    Given a Backlog finding and a target phase
    When I promote it with an overall score of 3
    Then the promotion succeeds

  Scenario: The declared threshold field is no longer enforced
    Given a quality report that declares a threshold other than 9
    When I validate the roadmap
    Then the roadmap is reported as valid

  Scenario Outline: Out-of-range scores are still refused
    Given a roadmap feature whose quality report scores its overall value <score>
    When I validate the roadmap
    Then validation is refused naming the offending score field and its value

    Examples:
      | score |
      | 0     |
      | 11    |

  Scenario: A missing scores object is still reported as a shape fault
    Given a quality report whose feature entry has no scores object
    When I validate the roadmap
    Then validation is refused because the scores field is missing
    And the refusal is not reported as an out-of-range score

  Scenario: A scores field of the wrong JSON kind is still reported as a shape fault
    Given a quality report whose feature entry has scores encoded as an array
    When I validate the roadmap
    Then validation is refused naming the JSON kind that was found
    And the refusal is not reported as an out-of-range score

  Scenario: A non-integer score is still reported as a type fault
    Given a quality report whose overall score is the string "nine"
    When I validate the roadmap
    Then validation is refused naming the offending score field

  Scenario: Existing quality artifacts survive the deletion untouched
    Given an existing roadmap quality report written before this change
    When I validate the roadmap
    Then the report file is not rewritten
    And it is still parsed successfully

  # ---------------------------------------------------------------
  # AC6 / AC7 / AC8 — the greenfield cascade
  # ---------------------------------------------------------------

  Scenario: A guided greenfield project cold-starts from PROJECT.md and roadmap.json alone
    Given a guided project containing only PROJECT.md and a roadmap json with a bootstrap phase
    And no roadmap analysis, no roadmap quality report, and no ROADMAP.md
    When I start the first bootstrap feature
    Then the workflow starts on the plan step

  Scenario: A strict greenfield project still requires the full cascade
    Given a strict project containing only PROJECT.md and a roadmap json with a bootstrap phase
    And no roadmap analysis and no roadmap quality report
    When I start the first bootstrap feature
    Then the start is refused naming the missing roadmap analysis

  Scenario: The setup hook advises rather than halts on a guided project
    Given a guided project with no roadmap analysis and no roadmap quality report
    When the setup hook runs
    Then it emits an advisory naming the missing artifacts
    And it still reaches the roadmap checkpoint

  Scenario: The setup hook still halts on a strict project
    Given a strict project with no roadmap analysis
    When the setup hook runs
    Then it emits the roadmap analysis directive and stops

  Scenario Outline: Guided still refuses what the roadmap says it must
    Given a guided greenfield project
    When I try to start <case>
    Then the start is refused

    Examples:
      | case                                                     |
      | a Backlog finding                                        |
      | a draft feature                                          |
      | a feature with unmet dependencies                        |
      | any feature in a roadmap with no bootstrap phase          |
      | a non-bootstrap feature while bootstrap is incomplete     |

  Scenario: A missing roadmap json is still refused under guided
    Given a guided project with PROJECT.md but no roadmap json
    When I try to start a feature
    Then the start is refused naming the roadmap json

  Scenario: The production-readiness gate is untouched by cascade slimming
    Given a guided project with the production-readiness gate enabled
    And a production-readiness report whose status is BLOCKING
    When I try to complete the validate step
    Then completion is refused because production readiness is blocking

  # ---------------------------------------------------------------
  # AC9 — Centinela does not downgrade itself
  # ---------------------------------------------------------------

  Scenario: Centinela's own repository pins its profile explicitly
    Given Centinela's own centinela.toml
    Then it declares an explicit enforcement profile
    And that profile is "strict"

  Scenario: Projects inheriting the new default are told about it
    Given a project that has existing workflows and no explicit enforcement profile
    When I run the doctor
    Then it reports an advisory that the default is now guided
    And the advisory does not fail the doctor run

  Scenario: A project that already pins its profile is not advised
    Given a project whose centinela.toml sets an explicit enforcement profile
    When I run the doctor
    Then no profile-default advisory is reported
