package workflow

import "github.com/samuelnp/centinela/internal/config"

// EffectiveProfile resolves the enforcement profile in force for a workflow.
// Precedence (highest → lowest): the explicit per-feature pin (--profile or an
// explicit global captured at start), then an explicit global enforcement_profile
// (RawEnforcementProfile non-empty), then the capability default of the pinned
// driver model, then the strict back-compat default. The result is always a
// normalized, known profile. Zero-config still resolves to strict byte-identically.
func EffectiveProfile(wf *Workflow, cfg *config.Config) string {
	if wf != nil && wf.EnforcementProfile != "" {
		return config.NormalizeEnforcementProfile(wf.EnforcementProfile)
	}
	if cfg != nil && cfg.Workflow.RawEnforcementProfile != "" {
		return config.NormalizeEnforcementProfile(cfg.Workflow.EnforcementProfile)
	}
	if wf != nil && wf.DriverModel != "" && cfg != nil {
		if profile, ok := config.DefaultProfileForModel(wf.DriverModel, cfg); ok {
			return config.NormalizeEnforcementProfile(profile)
		}
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
