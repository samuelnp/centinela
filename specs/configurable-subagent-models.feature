Feature: Configurable subagent model tiers
  As a developer running Centinela
  I want to declare a model tier per subagent role in centinela.toml
  So that I can right-size cost and latency without editing prompts

  # AC1 — configured tier is annotated on the emitted directive
  Scenario: Configured tier is reflected in the orchestration directive
    Given a centinela.toml with '[orchestration.models]' containing 'big-thinker = "reasoning"'
    When the orchestration hook emits the plan-step directive
    Then the directive contains 'big-thinker (model: reasoning)'

  # AC2 — unconfigured role uses its built-in default tier
  Scenario: Unconfigured role falls back to its default tier
    Given a centinela.toml with '[orchestration.models]' that does not mention 'documentation-specialist'
    When the orchestration hook emits the plan-step directive
    Then the directive contains 'documentation-specialist (model: fast)'

  # AC3 — absent [orchestration.models] table: zero-config-safe
  Scenario: Absent orchestration.models table — all defaults apply
    Given a centinela.toml with no '[orchestration.models]' section
    When the orchestration hook emits the plan-step directive
    Then every step role is annotated with its default tier
    And 'big-thinker (model: reasoning)' appears in the directive
    And 'feature-specialist (model: balanced)' appears in the directive
    And 'documentation-specialist (model: fast)' appears in the directive

  # AC4 — invalid tier value fails config load with a precise error
  Scenario: Invalid tier value is rejected at config load time
    Given a centinela.toml with '[orchestration.models]' containing 'qa-senior = "genius"'
    When Centinela loads the config
    Then config loading fails
    And the error message names the offending key 'qa-senior'
    And the error message lists the allowed tiers 'reasoning', 'balanced', 'fast'

  # AC5 — unknown role key fails config load with a precise error
  Scenario: Unknown role key is rejected at config load time
    Given a centinela.toml with '[orchestration.models]' containing 'backend-wizard = "fast"'
    When Centinela loads the config
    Then config loading fails
    And the error message names the offending key 'backend-wizard'

  # AC6 — runner-agnostic emission: tier name + both-runner reference line
  Scenario: Directive is runner-agnostic — emits tier name and both-runner reference
    Given a centinela.toml with '[orchestration.models]' containing 'big-thinker = "reasoning"'
    When the orchestration hook emits the plan-step directive
    Then the directive contains 'big-thinker (model: reasoning)'
    And the directive contains a model reference line mapping 'reasoning' to both 'claude-opus-4-7' and 'anthropic/claude-opus-4-7'
    And the directive does not contain a bare Claude-only model ID for any role annotation

  # Edge case — empty [orchestration.models] table behaves like absent
  Scenario: Empty orchestration.models table — all defaults apply
    Given a centinela.toml with an empty '[orchestration.models]' section
    When the orchestration hook emits the plan-step directive
    Then every step role is annotated with its default tier

  # Edge case — tier value is normalized (case + whitespace) before validation
  Scenario: Tier value with uppercase is normalized and accepted
    Given a centinela.toml with '[orchestration.models]' containing 'feature-specialist = "Reasoning"'
    When Centinela loads the config
    Then config loading succeeds
    And the resolved tier for 'feature-specialist' is 'reasoning'

  Scenario: Tier value with surrounding whitespace is normalized and accepted
    Given a centinela.toml with '[orchestration.models]' containing 'validation-specialist = " fast "'
    When Centinela loads the config
    Then config loading succeeds
    And the resolved tier for 'validation-specialist' is 'fast'

  # Edge case — normalized value that is still invalid is rejected
  Scenario: Tier value that is still invalid after normalization is rejected
    Given a centinela.toml with '[orchestration.models]' containing 'senior-engineer = " Genius "'
    When Centinela loads the config
    Then config loading fails
    And the error message names the offending key 'senior-engineer'
    And the error message lists the allowed tiers 'reasoning', 'balanced', 'fast'

  # Edge case — resolver never crashes on a missing tier→model mapping
  Scenario: Missing internal tier-to-model mapping falls back to tier name without crashing
    Given a tier value is configured that has no entry in the internal tier-to-model table
    When the orchestration hook emits the plan-step directive
    Then the hook completes without aborting
    And the role annotation falls back to the tier name
    And a warning is surfaced

  # Edge case — out-of-band roles are not emitted in the directive
  Scenario: Out-of-band roles are not annotated in the orchestration directive
    Given any centinela.toml configuration
    When the orchestration hook emits the plan-step directive
    Then the directive does not contain an annotation for 'gatekeeper'
    And the directive does not contain an annotation for 'production-readiness'
    And the directive does not contain an annotation for 'edge-case-tester'
    And the directive does not contain an annotation for 'merge-steward'
