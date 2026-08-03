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
// workflow", and it is consulted for process decisions only. A surface that
// reports a profile for a specific workflow must use EffectiveProfile, or it
// will disagree with status and verdict on every legacy workflow.
func ProjectDefaultProfile(cfg *Config) string {
	if cfg == nil {
		return ProfileStrict
	}
	if cfg.LoadFailed() {
		return ProfileStrict
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
	// The TAIL — and only the tail — requires a config a successful Load
	// produced. Every explicit tier above is honoured as written, including on a
	// fabricated config (that is how the strict fallback pins itself). But a nil,
	// zero-valued or error-path config states nothing, and nothing means the
	// profile is unknowable, and unknowable is strict. Same rule as
	// workflow.EffectiveProfile, so all three resolvers agree on any input.
	if !cfg.ResolvedByLoad() {
		return ProfileStrict
	}
	return ProfileGuided
}
