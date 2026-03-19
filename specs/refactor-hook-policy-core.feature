Feature: Shared hook policy core
  As a Centinela maintainer
  I want hook policy logic in a shared package
  So multiple agent integrations stay behaviorally consistent

  Scenario: Write is blocked when no workflow exists
    Given no active workflows
    When a code file write is evaluated
    Then policy should block with reason "start workflow"

  Scenario: Write is allowed when any active workflow permits file type
    Given active workflows in mixed steps
    When a file type is allowed by one active workflow
    Then policy should allow the write

  Scenario: Roadmap and other files are always allowed
    Given any workflow state
    When roadmap or uncategorized files are evaluated
    Then policy should allow the write

  Scenario: Block decision includes feature and step context
    Given active workflows that all disallow a file type
    When the write is evaluated
    Then policy should return feature and step used for block messaging
