package config

import "testing"

func cfgCaps(caps map[string]string) *Config {
	c := &Config{}
	c.Orchestration.Capabilities = caps
	return c
}

// CapabilityClassFor: built-in hit (both claude and anthropic/ forms), user
// override, user-beats-built-in, unknown → ("",false), empty/whitespace, trim,
// case-sensitivity of the model id (NOT lowercased).
func TestCapabilityClassFor(t *testing.T) {
	cases := []struct {
		name, id  string
		cfg       *Config
		wantClass string
		wantOK    bool
	}{
		{"builtin claude form", "claude-opus-4-7", nil, CapabilityFrontier, true},
		{"builtin anthropic form", "anthropic/claude-opus-4-7", nil, CapabilityFrontier, true},
		{"builtin capable", "claude-sonnet-4-6", nil, CapabilityCapable, true},
		{"builtin limited", "claude-haiku-4-5-20251001", nil, CapabilityLimited, true},
		{"user override", "local/m", cfgCaps(map[string]string{"local/m": "capable"}), CapabilityCapable, true},
		{"user beats builtin", "claude-opus-4-7", cfgCaps(map[string]string{"claude-opus-4-7": "limited"}), CapabilityLimited, true},
		{"unknown id", "some/unknown", nil, "", false},
		{"empty id", "", nil, "", false},
		{"whitespace id", "   ", nil, "", false},
		{"trimmed match builtin", "  claude-opus-4-7  ", nil, CapabilityFrontier, true},
		{"trimmed match user key", "local/m", cfgCaps(map[string]string{"  local/m  ": "frontier"}), CapabilityFrontier, true},
		{"case-sensitive miss", "Claude-Opus-4-7", nil, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			class, ok := CapabilityClassFor(tc.id, tc.cfg)
			if class != tc.wantClass || ok != tc.wantOK {
				t.Fatalf("got (%q,%v), want (%q,%v)", class, ok, tc.wantClass, tc.wantOK)
			}
		})
	}
}

// AllowedCapabilityClasses returns the three classes in stable order.
func TestAllowedCapabilityClasses(t *testing.T) {
	got := AllowedCapabilityClasses()
	want := []string{CapabilityFrontier, CapabilityCapable, CapabilityLimited}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d = %q, want %q", i, got[i], want[i])
		}
	}
}
