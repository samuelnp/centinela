package workflow

import (
	"strings"

	"github.com/samuelnp/centinela/internal/config"
)

// StartDecision is what `start` pins on the workflow and what governs the
// orchestration evidence mode at creation time.
type StartDecision struct {
	// PinnedProfile is set ONLY when --profile was passed. An explicit global
	// enforcement_profile is deliberately NOT pinned — it resolves live at runtime
	// (tier 2 of EffectiveProfile) so its status provenance reads "global", distinct
	// from a per-feature "--profile" pin. Empty lets runtime EffectiveProfile
	// re-derive through the global/capability/strict tiers.
	PinnedProfile string
	// DriverModel is the resolved driver model id (flag → env → config).
	DriverModel string
	// EffectiveProfile governs the orchestration evidence mode at start; it mirrors
	// runtime EffectiveProfile's precedence over the start-time inputs.
	EffectiveProfile string
}

// ResolveStart computes the start-time decision. flagProfile is --profile (may be
// ""); flagModel is --model (may be ""). Precedence mirrors runtime
// EffectiveProfile: explicit --profile > explicit global > capability default >
// guided. Only an explicit --profile is pinned; an explicit global is left to
// resolve live so it stays distinguishable in status provenance.
//
// The nothing-configured tail is guided — every workflow born here carries
// ProfileContractGuidedDefault, so runtime EffectiveProfile agrees. A DECLARED
// but unmapped driver model still falls back to strict: naming a model
// Centinela has no capability class for is a signal to keep maximum
// scaffolding, not to inherit the zero-config default.
func ResolveStart(flagProfile, flagModel string, cfg *config.Config) StartDecision {
	d := StartDecision{}
	d.DriverModel = config.DriverModelFrom(flagModel, cfg)
	if strings.TrimSpace(flagProfile) != "" {
		d.PinnedProfile = config.NormalizeEnforcementProfile(flagProfile)
	}
	switch {
	case d.PinnedProfile != "":
		d.EffectiveProfile = d.PinnedProfile
	case cfg != nil && cfg.Workflow.RawEnforcementProfile != "":
		d.EffectiveProfile = config.NormalizeEnforcementProfile(cfg.Workflow.EnforcementProfile)
	case d.DriverModel != "":
		if p, ok := config.DefaultProfileForModel(d.DriverModel, cfg); ok {
			d.EffectiveProfile = config.NormalizeEnforcementProfile(p)
		} else {
			d.EffectiveProfile = config.ProfileStrict
		}
	default:
		d.EffectiveProfile = config.ProfileGuided
	}
	return d
}
