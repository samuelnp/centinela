Feature: MCP governance server
  As a host harness with no Centinela-specific code
  I want to obtain a governance verdict through MCP tool calls
  So that enforcement works without a bespoke per-harness adapter

  Background:
    Given Centinela exposes a versioned MCP server over stdio
    And the server registers the tools read_rules, run_gates, verify_claims, workflow_state

  Scenario: A zero-integration harness obtains a verdict via tool calls
    Given an MCP client connected to "centinela mcp serve"
    When the client lists tools
    Then it sees read_rules, run_gates, verify_claims, and workflow_state
    And calling run_gates returns gate results and an allow/warn/block decision

  Scenario: The verdict is versioned
    When the client inspects the verdict payload
    Then it carries the schema identifier "centinela.mcp/v1"

  Scenario: The shim denies a write on a block verdict
    Given a repository whose gates fail
    When I run "centinela mcp shim" for the active feature
    Then it exits with code 2 (the harness pre-write deny)

  Scenario: The shim allows a write on an allow verdict
    Given a repository whose gates pass and claims verify
    When I run "centinela mcp shim" for the active feature
    Then it exits with code 0

  Scenario: MCP verdict matches the native-hook verdict (parity)
    Given the same feature, step, and repository state
    When I obtain the verdict through MCP tool calls
    And I obtain the verdict through the native "centinela verdict" path
    Then the two decisions are identical

  Scenario: The server is advisory and never mutates state
    When any tool is called
    Then the server only reads gates, claims, and workflow state
    And it returns a verdict without performing or blocking any write
