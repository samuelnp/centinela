// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// Scenario: Capability classification covers aliases and legacy pins
func TestTD_CapabilityClassificationCoversAliasesAndLegacyPins(t *testing.T) {
	cases := []struct{ model, class string }{
		{"opus", config.CapabilityFrontier},
		{"sonnet", config.CapabilityCapable},
		{"haiku", config.CapabilityLimited},
		{"claude-opus-4-7", config.CapabilityFrontier},
		{"anthropic/claude-opus-4-7", config.CapabilityFrontier},
		{"claude-sonnet-4-6", config.CapabilityCapable},
		{"anthropic/claude-sonnet-4-6", config.CapabilityCapable},
		{"claude-haiku-4-5-20251001", config.CapabilityLimited},
		{"anthropic/claude-haiku-4-5", config.CapabilityLimited},
	}
	for _, tc := range cases {
		class, ok := config.CapabilityClassFor(tc.model, nil)
		if !ok || class != tc.class {
			t.Fatalf("model %q: class = (%q, %v), want (%q, true)", tc.model, class, ok, tc.class)
		}
		// DefaultProfileForModel must engage the capability tier (ok=true) for
		// both the new aliases and every retired dated pin — the superset
		// guarantee this whole slice exists to protect.
		if _, ok := config.DefaultProfileForModel(tc.model, nil); !ok {
			t.Fatalf("model %q: DefaultProfileForModel must report ok=true", tc.model)
		}
	}
}

// Scenario: An operator's retired pin keeps its default enforcement profile
func TestTD_OperatorRetiredPinKeepsDefaultEnforcementProfile(t *testing.T) {
	retired := "claude-haiku-4-5-20251001"
	alias := "haiku"

	retiredProfile, retiredOK := config.DefaultProfileForModel(retired, nil)
	aliasProfile, aliasOK := config.DefaultProfileForModel(alias, nil)
	if !retiredOK {
		t.Fatalf("retired pin %q must still engage the capability tier", retired)
	}
	if retiredProfile != aliasProfile {
		t.Fatalf("retired pin profile %q must match its alias's profile %q", retiredProfile, aliasProfile)
	}
	if !aliasOK {
		t.Fatal("sanity: alias must resolve too")
	}
	if retiredProfile != config.ProfileStrict {
		t.Fatalf("limited-class default profile must be strict, got %q", retiredProfile)
	}
}
