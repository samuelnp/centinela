Feature: Auto-start workflow from prompt intent
  As a maintainer
  I want Centinela to auto-start new workflows from user intent
  So process enforcement continues after completed features

  Scenario: Done workflows do not unlock writes
    Given all existing workflows are done
    When a non-roadmap write is evaluated
    Then prewrite should require starting a new workflow

  Scenario: Prompt intent auto-starts workflow
    Given no active workflow exists
    When the user asks to add or extend a feature
    Then Centinela should auto-start a new feature workflow

  Scenario: Active workflow prevents auto-start
    Given an active workflow exists
    When the prompt hook evaluates new-feature intent
    Then no new workflow should be created
