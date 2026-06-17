package config

// WorkflowConfig controls language-specific step validation behaviour.
type WorkflowConfig struct {
	TestSuffixes           []string `toml:"test_suffixes"`
	AcceptanceSuffix       string   `toml:"acceptance_suffix"`
	CodeDirs               []string `toml:"code_dirs"`
	DisableAutoCommit      bool     `toml:"disable_auto_commit"`
	StepConfirmationMode   string   `toml:"step_confirmation_mode"`
	PlanAdvisorMode        string   `toml:"plan_advisor_mode"`
	PlanQuestionLimit      int      `toml:"plan_question_limit"`
	PlanAdvisorFailureTopN int      `toml:"plan_advisor_failure_top_n"`
	EnforcementProfile     string   `toml:"enforcement_profile"`
	// RawStepConfirmationMode preserves the value decoded from centinela.toml
	// BEFORE applyDefaults normalizes StepConfirmationMode. An empty raw value
	// means the knob was unset, which the precedence resolver needs to tell an
	// explicit "every_step" apart from a defaulted one. Not a toml field — it is
	// captured in Load and never serialized.
	RawStepConfirmationMode string `toml:"-"`
	// RawEnforcementProfile preserves the value decoded from centinela.toml BEFORE
	// applyDefaults normalizes EnforcementProfile to strict. An empty raw value
	// means the knob was unset, which the precedence resolver needs to tell an
	// explicit "strict" apart from a defaulted one — so the capability tier engages
	// only when no explicit global profile was set. Not a toml field — captured in
	// Load and never serialized.
	RawEnforcementProfile string `toml:"-"`
	// UseWorktrees enables per-feature git worktrees at `.worktrees/<feature>/`.
	// Default false preserves the single-checkout flow.
	UseWorktrees bool `toml:"use_worktrees"`
}
