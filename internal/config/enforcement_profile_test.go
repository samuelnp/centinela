package config

import (
	"strings"
	"testing"
)

func TestNormalizeEnforcementProfile(t *testing.T) {
	cases := map[string]string{
		"strict":     ProfileStrict,
		"guided":     ProfileGuided,
		"outcome":    ProfileOutcome,
		"GUIDED":     ProfileGuided,  // case-insensitive
		"  outcome ": ProfileOutcome, // trims whitespace
		"":           ProfileStrict,  // empty falls through to strict
		"wat":        ProfileStrict,  // unknown falls through to strict
	}
	for in, want := range cases {
		if got := NormalizeEnforcementProfile(in); got != want {
			t.Fatalf("NormalizeEnforcementProfile(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidateEnforcementProfile_AllowsEmptyAndKnown(t *testing.T) {
	for _, ok := range []string{"", ProfileStrict, ProfileGuided, ProfileOutcome, "OUTCOME", " guided "} {
		if err := validateEnforcementProfile(ok); err != nil {
			t.Fatalf("validateEnforcementProfile(%q) = %v, want nil", ok, err)
		}
	}
}

func TestValidateEnforcementProfile_RejectsUnknownNamingField(t *testing.T) {
	err := validateEnforcementProfile("turbo")
	if err == nil {
		t.Fatal("expected error for unsupported profile")
	}
	if !strings.Contains(err.Error(), "enforcement_profile") || !strings.Contains(err.Error(), "turbo") {
		t.Fatalf("error must name the field and the bad value, got: %v", err)
	}
}
