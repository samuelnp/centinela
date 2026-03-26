Feature: README documents full-sync migration workflow

  Scenario: User reads migration command guidance
    Given full-sync migration is implemented
    When the user reads README migration docs
    Then they see `centinela migrate` preview and `--apply` full-sync usage

  Scenario: User reads setup-scoped migration guidance
    Given setup migration supports agent scope
    When the user reads README migration docs
    Then they see `centinela migrate setup --agent claude|opencode|both`
