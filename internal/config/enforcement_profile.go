package config

import (
	"fmt"
	"strings"
)

// Enforcement profiles scale how much *process* Centinela enforces while the
// verification axis (validate-step gates + claim verification) stays constant.
// strict is the back-compat default: an unconfigured project resolves to it and
// reproduces today's behavior byte-for-byte.
const (
	ProfileStrict  = "strict"
	ProfileGuided  = "guided"
	ProfileOutcome = "outcome"
)

// NormalizeEnforcementProfile maps any input to a known profile. Unknown or
// empty values fall through to ProfileStrict — the behavior-preserving default.
func NormalizeEnforcementProfile(profile string) string {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case ProfileGuided:
		return ProfileGuided
	case ProfileOutcome:
		return ProfileOutcome
	default:
		return ProfileStrict
	}
}

// validateEnforcementProfile rejects a NON-EMPTY unsupported value. Empty is
// valid (it resolves to the strict default); only an explicitly-set unknown
// string is a configuration error.
func validateEnforcementProfile(profile string) error {
	v := strings.ToLower(strings.TrimSpace(profile))
	switch v {
	case "", ProfileStrict, ProfileGuided, ProfileOutcome:
		return nil
	default:
		return fmt.Errorf("workflow.enforcement_profile %q is unsupported (use strict, guided, or outcome)", profile)
	}
}
