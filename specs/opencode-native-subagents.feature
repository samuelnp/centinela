Feature: OpenCode native subagents
  As an OpenCode user running Centinela
  I want Centinela specialist roles configured as native OpenCode subagents
  So orchestration can use real child agents instead of simulated role evidence

  Scenario: Generated OpenCode config includes Centinela subagents
    Given a project initialized with OpenCode support
    When Centinela injects opencode.json
    Then opencode.json should include big-thinker as a subagent
    And opencode.json should include feature-specialist as a subagent
    And opencode.json should include senior-engineer as a subagent
    And opencode.json should include qa-senior as a subagent
    And opencode.json should include documentation-specialist as a subagent
    And opencode.json should include ux-ui-specialist as a subagent

  Scenario: Existing OpenCode agent config is preserved
    Given opencode.json already has a custom agent
    When Centinela injects OpenCode defaults
    Then the custom agent should remain unchanged
    And missing Centinela subagents should be added

  Scenario: Build agent can invoke Centinela subagents
    Given a project initialized with OpenCode support
    When Centinela injects opencode.json
    Then the build agent task permission should allow Centinela subagents
