package workflow

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/config"
)

// ProfileProvenance returns the active enforcement profile and a short source
// annotation for read-only status, mirroring EffectiveProfile's precedence tiers:
//   - tier 1 (explicit per-feature pin): note "--profile"
//   - tier 2 (explicit global enforcement_profile): note "global"
//   - tier 3 hit (driver model maps to a class): note "driver: <id> → <class>"
//   - tier 3 miss (driver set, no class): note "driver: <id> → no capability, default strict"
//   - tier 4 hit (pinned contract + a real config): note "default (guided)"
//   - an unreadable centinela.toml (checked ahead of every tier below the
//     per-feature pin): note "unreadable centinela.toml, default strict"
//   - tier 4 miss (legacy/in-flight workflow): note "default (strict, legacy workflow)"
//
// The two tier-4 notes and the tier-3 driver note stay distinguishable on
// purpose: a guided resolved from a capable driver model is not the same fact
// as a guided inherited from the shipped default, and status must not conflate
// them. When cfg is nil (status has no config), it falls back to the pinned
// value or the tier-4 notes, so zero-config output stays sensible.
func ProfileProvenance(wf *Workflow, cfg *config.Config) (profile, note string) {
	if wf != nil && wf.EnforcementProfile != "" {
		return config.NormalizeEnforcementProfile(wf.EnforcementProfile), "--profile"
	}
	if cfg.LoadFailed() {
		return config.ProfileStrict, "unreadable " + config.Filename + ", default strict"
	}
	if cfg != nil && cfg.Workflow.RawEnforcementProfile != "" {
		return config.NormalizeEnforcementProfile(cfg.Workflow.EnforcementProfile), "global"
	}
	if driver := TierDriverModel(wf, cfg); driver != "" && cfg != nil {
		if class, ok := config.CapabilityClassFor(driver, cfg); ok {
			profile = config.NormalizeEnforcementProfile(config.ProfileForCapability(class, cfg))
			return profile, fmt.Sprintf("driver: %s → %s", driver, class)
		}
		if _, ok := config.LocalDefaultClass(driver, cfg); ok {
			return config.ProfileStrict, fmt.Sprintf("local default: %s → limited → strict", driver)
		}
		return config.ProfileStrict, fmt.Sprintf("driver: %s → no capability, default strict", driver)
	}
	if wf.UsesGuidedDefault() && cfg.ResolvedByLoad() {
		return config.ProfileGuided, "default (guided)"
	}
	if wf.UsesGuidedDefault() {
		// Post-flip workflow, but the config was never loaded — not a legacy
		// workflow, and saying so would misattribute the strictness.
		return config.ProfileStrict, "default (strict, unresolved config)"
	}
	return config.ProfileStrict, "default (strict, legacy workflow)"
}
