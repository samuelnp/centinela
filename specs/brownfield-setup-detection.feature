Feature: Brownfield setup detection — route existing repos to analyze+synthesize instead of cold-questioning
  As a user onboarding an existing codebase with Centinela
  I want the setup hook to detect source signals and direct me through analyze/synthesize
  So that I am not cold-interrogated about a stack the tooling can already read from my manifests and design docs

  Background:
    Given centinela is initialized (PROJECT.md.template and centinela.toml are present)
    And PROJECT.md is absent

  Scenario: Brownfield repo with go.mod emits the brownfield directive
    Given the project root contains a go.mod file
    When centinela hook setup runs
    Then the exit code is zero
    And the output contains the BROWNFIELD directive
    And the output mentions "centinela analyze"
    And the output mentions "centinela synthesize"
    And the output mentions "Project Stage: existing"
    And the output does NOT contain the six-question greenfield setup prompt

  Scenario: Brownfield repo with package.json emits the brownfield directive
    Given the project root contains a package.json file
    When centinela hook setup runs
    Then the exit code is zero
    And the output contains the BROWNFIELD directive
    And the output mentions "centinela analyze"
    And the output mentions "centinela synthesize"
    And the output mentions "Project Stage: existing"

  Scenario: Brownfield repo with only a Makefile is detected as brownfield
    Given the project root contains only a Makefile (no other manifests, no populated source dirs)
    When centinela hook setup runs
    Then the exit code is zero
    And the output contains the BROWNFIELD directive
    And the output does NOT contain the six-question greenfield setup prompt

  Scenario: Brownfield repo with a populated src/ directory is detected as brownfield
    Given the project root contains a src/ directory with at least one non-hidden file
    When centinela hook setup runs
    Then the exit code is zero
    And the output contains the BROWNFIELD directive
    And the output mentions "centinela analyze"

  Scenario: Greenfield empty repo still emits the existing question-based setup directive
    Given the project root has no manifests (no go.mod, package.json, Cargo.toml, Gemfile, pyproject.toml, requirements.txt, Makefile)
    And there are no populated source directories (src, app, lib, cmd, pkg, internal)
    When centinela hook setup runs
    Then the exit code is zero
    And the output contains the GREENFIELD directive
    And the output contains the six-question greenfield setup prompt
    And the output does NOT contain the BROWNFIELD directive

  Scenario: Empty src/ directory is NOT a brownfield signal
    Given the project root contains an empty src/ directory (no files inside)
    And the project root has no manifest files
    When centinela hook setup runs
    Then the exit code is zero
    And the output contains the GREENFIELD directive
    And the output does NOT contain the BROWNFIELD directive

  Scenario: PROJECT.md already present bypasses both setup directives
    Given PROJECT.md is present in the project root
    When centinela hook setup runs
    Then the hook proceeds to roadmap checks without emitting a setup directive
    And neither the BROWNFIELD directive nor the GREENFIELD setup prompt is emitted

  Scenario: Brownfield directive instructs enrich-then-confirm workflow
    Given the project root contains a Cargo.toml file
    When centinela hook setup runs
    Then the output contains the BROWNFIELD directive
    And the output instructs the agent to enrich the draft by reading key source files
    And the output instructs the agent to confirm uncertain fields before finalizing PROJECT.md
    And the output does NOT instruct the agent to ignore the user's message

  Scenario: Brownfield repo with populated internal/ directory is detected as brownfield
    Given the project root contains an internal/ directory with at least one non-hidden file
    When centinela hook setup runs
    Then the exit code is zero
    And the output contains the BROWNFIELD directive

  Scenario: HasSource detector does not walk subdirectories (cheap root-only check)
    Given the project root has no manifests and no populated root-level source dirs
    And a deeply nested subdirectory contains source files (not at root level)
    When centinela hook setup runs
    Then the output contains the GREENFIELD directive
    And the output does NOT contain the BROWNFIELD directive
