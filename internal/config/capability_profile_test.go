package config

import "testing"

func cfgCapProfiles(m map[string]string) *Config {
	c := &Config{}
	c.Orchestration.CapabilityProfiles = m
	return c
}

// ProfileForCapability: the three built-in defaults, a capability_profiles
// override honored, and class normalization (trim + lower) on both lookup paths.
func TestProfileForCapability(t *testing.T) {
	cases := []struct {
		name, class string
		cfg         *Config
		want        string
	}{
		{"frontier default", CapabilityFrontier, nil, ProfileOutcome},
		{"capable default", CapabilityCapable, nil, ProfileGuided},
		{"limited default", CapabilityLimited, nil, ProfileStrict},
		{"override remaps frontier", CapabilityFrontier, cfgCapProfiles(map[string]string{"frontier": "guided"}), ProfileGuided},
		{"input class normalized", "  Frontier  ", nil, ProfileOutcome},
		{"override key normalized", CapabilityFrontier, cfgCapProfiles(map[string]string{"  FRONTIER  ": "strict"}), ProfileStrict},
		{"unknown class falls to strict", "genius", nil, ProfileStrict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ProfileForCapability(tc.class, tc.cfg); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// DefaultProfileForModel: ok=true with the resolved profile for a known model,
// ok=false for an unknown one (caller must not engage the capability tier).
func TestDefaultProfileForModel(t *testing.T) {
	if p, ok := DefaultProfileForModel("claude-opus-4-7", nil); !ok || p != ProfileOutcome {
		t.Fatalf("frontier builtin: got (%q,%v), want (outcome,true)", p, ok)
	}
	if p, ok := DefaultProfileForModel("anthropic/claude-haiku-4-5", nil); !ok || p != ProfileStrict {
		t.Fatalf("limited builtin: got (%q,%v), want (strict,true)", p, ok)
	}
	cfg := cfgCaps(map[string]string{"local/x": "capable"})
	if p, ok := DefaultProfileForModel("local/x", cfg); !ok || p != ProfileGuided {
		t.Fatalf("declared capable: got (%q,%v), want (guided,true)", p, ok)
	}
	if p, ok := DefaultProfileForModel("some/unknown", nil); ok || p != "" {
		t.Fatalf("unknown: got (%q,%v), want (\"\",false)", p, ok)
	}
}
