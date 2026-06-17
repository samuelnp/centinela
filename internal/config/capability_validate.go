package config

import (
	"fmt"
	"strings"
)

// knownCapabilityClass reports whether a value (normalized trim+lower) is one of
// the three known capability classes.
func knownCapabilityClass(class string) bool {
	switch normClass(class) {
	case CapabilityFrontier, CapabilityCapable, CapabilityLimited:
		return true
	default:
		return false
	}
}

// validateCapabilities rejects malformed capability config: a value in
// [orchestration.capabilities] that is not a known class, an empty model-id key,
// a key in [orchestration.capability_profiles] that is not a known class, and a
// value there that is not a known enforcement profile. Absent/empty tables are
// valid (zero-config preserves the strict default).
func validateCapabilities(cfg *Config) error {
	if cfg == nil {
		return nil
	}
	for modelID, class := range cfg.Orchestration.Capabilities {
		if strings.TrimSpace(modelID) == "" {
			return fmt.Errorf("orchestration.capabilities: model id key must not be empty")
		}
		if !knownCapabilityClass(class) {
			return fmt.Errorf("orchestration.capabilities[%q]: unknown capability class %q (allowed: %s)",
				modelID, class, strings.Join(AllowedCapabilityClasses(), ", "))
		}
	}
	for class, profile := range cfg.Orchestration.CapabilityProfiles {
		if !knownCapabilityClass(class) {
			return fmt.Errorf("orchestration.capability_profiles: unknown capability class key %q (allowed: %s)",
				class, strings.Join(AllowedCapabilityClasses(), ", "))
		}
		if err := validateEnforcementProfile(profile); err != nil {
			return fmt.Errorf("orchestration.capability_profiles[%q]: %w", class, err)
		}
	}
	return nil
}
