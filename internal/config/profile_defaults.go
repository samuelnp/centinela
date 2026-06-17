package config

// ProfileKnobs holds the per-profile defaults for the governed *process* knobs.
// It is the single source of truth consumed by every relaxation site (prewrite
// gating, confirmation prompt, orchestration evidence, plan advisor). Explicit
// centinela.toml knobs still win over these defaults via the precedence model.
type ProfileKnobs struct {
	StepGating              bool   // prewrite ordering enforcement
	ConfirmationMode        string // step_confirmation_mode default
	RequireSubagentEvidence bool   // strict-subagents-v1 orchestration mode
	PlanAdvisorMode         string // plan_advisor_mode default
}

// ProfileDefaults returns the knob defaults for a profile. Unknown profiles map
// to strict via NormalizeEnforcementProfile, so callers may pass raw values.
func ProfileDefaults(profile string) ProfileKnobs {
	switch NormalizeEnforcementProfile(profile) {
	case ProfileGuided:
		return ProfileKnobs{
			StepGating:              true,
			ConfirmationMode:        ConfirmAfterPlan,
			RequireSubagentEvidence: false,
			PlanAdvisorMode:         PlanAdvisorMissingInfo,
		}
	case ProfileOutcome:
		return ProfileKnobs{
			StepGating:              false,
			ConfirmationMode:        ConfirmAuto,
			RequireSubagentEvidence: false,
			PlanAdvisorMode:         PlanAdvisorOff,
		}
	default: // strict — reproduces today's behavior exactly
		return ProfileKnobs{
			StepGating:              true,
			ConfirmationMode:        ConfirmEveryStep,
			RequireSubagentEvidence: true,
			PlanAdvisorMode:         PlanAdvisorAlways,
		}
	}
}
