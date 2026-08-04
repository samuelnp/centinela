package workflow

import "github.com/samuelnp/centinela/internal/config"

// EffectiveProfile resolves the enforcement profile in force for a workflow.
// Precedence (highest → lowest): the explicit per-feature pin (--profile), then
// an unreadable-config refusal, then an explicit global enforcement_profile,
// then the capability default of the driver model, then the tail default. The
// result is always a normalized, known profile.
//
// Tier 3 is TERMINAL once a driver model is in play: a model with no capability
// class returns strict rather than falling through to the tail, matching
// ResolveStart, ProjectDefaultProfile and ProfileProvenance. The driver comes
// from TierDriverModel, not from wf.DriverModel alone — see there.
//
// The tail additionally requires cfg.ResolvedByLoad(). A nil, zero-valued or
// error-path-fabricated config means the profile is UNKNOWABLE, and unknowable
// resolves strict. That check lives here, at the decision, rather than at each
// caller: a surface that swallows a config load error and substitutes an empty
// struct — the idiom four hooks used — cannot reach guided by accident. Fixing
// this one call site at a time did not converge; this is the seam that does.
func EffectiveProfile(wf *Workflow, cfg *config.Config) string {
	if wf != nil && wf.EnforcementProfile != "" {
		return config.NormalizeEnforcementProfile(wf.EnforcementProfile)
	}
	if cfg.LoadFailed() {
		return config.ProfileStrict
	}
	if cfg != nil && cfg.Workflow.RawEnforcementProfile != "" {
		return config.NormalizeEnforcementProfile(cfg.Workflow.EnforcementProfile)
	}
	if driver := TierDriverModel(wf, cfg); driver != "" && cfg != nil {
		if profile, ok := config.DefaultProfileForModel(driver, cfg); ok {
			return config.NormalizeEnforcementProfile(profile)
		}
		return config.ProfileStrict
	}
	if wf.UsesGuidedDefault() && cfg.ResolvedByLoad() {
		return config.ProfileGuided
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

// TierDriverModel resolves the driver model that keys the capability tier for a
// workflow: the model PINNED on the workflow at start when it has one, otherwise
// the project's declared model ($CENTINELA_MODEL, then [orchestration]
// driver_model, then the local block) — the same source DriverModelFrom gives
// every ACTING path through ProjectDefaultProfile and ResolveStart.
//
// Consulting only wf.DriverModel made every reporting surface disagree with
// every acting one for a workflow that predates the declaration: status, verdict
// and MCP read_rules said guided while start, hook setup, roadmap promote and
// doctor behaved strict on the identical tree. The pin still wins when present —
// it is the state-dated decision for that workflow — but its absence must fall
// through to the project declaration rather than skipping the tier entirely.
func TierDriverModel(wf *Workflow, cfg *config.Config) string {
	if wf != nil && wf.DriverModel != "" {
		return wf.DriverModel
	}
	return config.DriverModelFrom("", cfg)
}
