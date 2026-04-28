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
}
