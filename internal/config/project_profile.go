package config

// ProjectDefaultProfile resolves the enforcement profile for surfaces that have
// NO workflow to consult — `centinela hook setup` and the pre-`start` greenfield
// guards both run before (or without) any feature state existing.
//
// It mirrors workflow.EffectiveProfile's tiers minus the per-feature pin: an
// explicit global enforcement_profile wins, then the declared driver model's
// capability class, then guided. A DECLARED but unmapped driver model falls back
// to strict, matching workflow.ResolveStart: naming a model with no known class
// asks for maximum scaffolding, not for the zero-config default.
//
// Workflow-scoped callers must NOT use this: only workflow.EffectiveProfile
// honours the ProfileContract pin that keeps pre-flip workflows strict. This
// function answers "what does this PROJECT default to", not "what governs this
// workflow", and it is consulted for process decisions only.
func ProjectDefaultProfile(cfg *Config) string {
	if cfg == nil {
		return ProfileGuided
	}
	if cfg.Workflow.RawEnforcementProfile != "" {
		return NormalizeEnforcementProfile(cfg.Workflow.EnforcementProfile)
	}
	if model := DriverModelFrom("", cfg); model != "" {
		if profile, ok := DefaultProfileForModel(model, cfg); ok {
			return NormalizeEnforcementProfile(profile)
		}
		return ProfileStrict
	}
	return ProfileGuided
}
