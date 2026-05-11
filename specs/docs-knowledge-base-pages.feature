Feature: Per-feature knowledge base in the docs step
  As a Centinela end-user
  I want each feature to ship a plain-language guide
  So I can understand what Centinela does without reading Gherkin or code

  Scenario: docs generate produces a knowledge base index
    Given at least one knowledge base markdown exists under docs/project-docs/kb
    When centinela docs generate is run
    Then docs/project-docs/kb/index.html should exist
    And the main docs index should include a Knowledge Base navigation link

  Scenario: each KB markdown produces a per-feature HTML page
    Given a KB markdown file docs/project-docs/kb/<feature>.md exists with all required sections
    When centinela docs generate is run
    Then docs/project-docs/kb/<feature>.html should exist
    And the page should render the "What it does", "When you'd use it", "How it behaves", and "Examples" sections

  Scenario: docs validation fails when the current feature has no KB markdown
    Given the docs step is being completed for feature <feature>
    And docs/project-docs/kb/<feature>.md is missing
    When centinela complete <feature> is attempted
    Then the command should fail with an actionable error naming the missing KB markdown path

  Scenario: features without a KB markdown appear as placeholder cards
    Given a feature has a spec but no KB markdown
    When centinela docs generate is run
    Then the KB index should include a "guide not yet written" card for that feature
    And no per-feature HTML page should be generated for it

  Scenario: KB markdown rejects missing required sections
    Given a KB markdown is missing the "What it does" section
    When centinela docs generate is run
    Then it should fail with an error naming the feature and the missing section
