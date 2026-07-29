package config

import "strings"

// Capability classes name a model's place on the strength spectrum. A driver
// model's declared class selects the DEFAULT enforcement profile (a new, lowest
// precedence tier). They are deliberately distinct from the profile names so a
// class can be remapped to any profile without a confusing identity mapping.
const (
	CapabilityFrontier = "frontier"
	CapabilityCapable  = "capable"
	CapabilityLimited  = "limited"
)

// AllowedCapabilityClasses returns the valid classes in stable order.
func AllowedCapabilityClasses() []string {
	return []string{CapabilityFrontier, CapabilityCapable, CapabilityLimited}
}

// normClass normalizes a capability class: trim + lowercase (fixed vocabulary).
func normClass(class string) string {
	return strings.ToLower(strings.TrimSpace(class))
}

// CapabilityClassFor returns the class for a model id: a user
// [orchestration.capabilities] override first (normalized), then the built-in
// map, else ("", false). Model-id keys are matched by trim only.
func CapabilityClassFor(modelID string, cfg *Config) (string, bool) {
	id := strings.TrimSpace(modelID)
	if id == "" {
		return "", false
	}
	if cfg != nil {
		for key, class := range cfg.Orchestration.Capabilities {
			if strings.TrimSpace(key) == id {
				return normClass(class), true
			}
		}
	}
	if class, ok := builtinModelCapability[id]; ok {
		return class, true
	}
	return "", false
}

// ProfileForCapability returns the default profile for a class, honoring an
// [orchestration.capability_profiles] override (normalized), else the built-in
// defaultProfileForClass.
func ProfileForCapability(class string, cfg *Config) string {
	c := normClass(class)
	if cfg != nil {
		for key, profile := range cfg.Orchestration.CapabilityProfiles {
			if normClass(key) == c {
				return NormalizeEnforcementProfile(profile)
			}
		}
	}
	return defaultProfileForClass(c)
}

// DefaultProfileForModel resolves a model id straight to its default profile:
// CapabilityClassFor → ProfileForCapability. ok=false when the id has no class —
// the caller must NOT engage the capability tier in that case.
func DefaultProfileForModel(modelID string, cfg *Config) (profile string, ok bool) {
	class, found := CapabilityClassFor(modelID, cfg)
	if !found {
		// Lowest tier: an unmapped declared local model defaults limited → strict.
		if _, ok := LocalDefaultClass(modelID, cfg); ok {
			return ProfileForCapability(CapabilityLimited, cfg), true
		}
		return "", false
	}
	return ProfileForCapability(class, cfg), true
}
