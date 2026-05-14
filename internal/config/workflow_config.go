package config

// WorkflowConfig controls language-specific step validation behaviour.
type WorkflowConfig struct {
	TestSuffixes         []string `toml:"test_suffixes"`
	AcceptanceSuffix     string   `toml:"acceptance_suffix"`
	CodeDirs             []string `toml:"code_dirs"`
	DisableAutoCommit    bool     `toml:"disable_auto_commit"`
	StepConfirmationMode string   `toml:"step_confirmation_mode"`
	PlanAdvisorMode      string   `toml:"plan_advisor_mode"`
	PlanQuestionLimit    int      `toml:"plan_question_limit"`
	// UseWorktrees enables per-feature git worktrees at `.worktrees/<feature>/`.
	// Default false preserves the single-checkout flow.
	UseWorktrees bool `toml:"use_worktrees"`
}
