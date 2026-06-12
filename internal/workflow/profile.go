package workflow

import "github.com/samuelnp/centinela/internal/config"

// EffectiveProfile resolves the enforcement profile in force for a workflow:
// the per-feature override (pinned at start) wins, else the global
// [workflow] enforcement_profile, else the strict back-compat default. The
// result is always a normalized, known profile.
func EffectiveProfile(wf *Workflow, cfg *config.Config) string {
	if wf != nil && wf.EnforcementProfile != "" {
		return config.NormalizeEnforcementProfile(wf.EnforcementProfile)
	}
	if cfg != nil {
		return config.NormalizeEnforcementProfile(cfg.Workflow.EnforcementProfile)
	}
	return config.ProfileStrict
}

// DisplayProfile returns the profile to surface in read-only views (status).
// It uses only the pinned per-feature value, defaulting to strict when unset,
// so it needs no config dependency.
func DisplayProfile(wf *Workflow) string {
	if wf == nil {
		return config.ProfileStrict
	}
	return config.NormalizeEnforcementProfile(wf.EnforcementProfile)
}
