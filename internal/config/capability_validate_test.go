package config

import (
	"strings"
	"testing"
)

// validateCapabilities: valid passes; unknown class value; empty model-id key;
// unknown class key in capability_profiles; unknown profile value; absent/empty
// tables → nil; nil cfg → nil.
func TestValidateCapabilities(t *testing.T) {
	cases := []struct {
		name    string
		cfg     *Config
		wantErr string // substring; "" means expect nil
	}{
		{"nil cfg", nil, ""},
		{"absent tables", &Config{}, ""},
		{"empty tables", func() *Config {
			c := &Config{}
			c.Orchestration.Capabilities = map[string]string{}
			c.Orchestration.CapabilityProfiles = map[string]string{}
			return c
		}(), ""},
		{"valid caps", cfgCaps(map[string]string{"local/m": "  Frontier  "}), ""},
		{"unknown class value", cfgCaps(map[string]string{"local/m": "genius"}), "genius"},
		{"empty model-id key", cfgCaps(map[string]string{"  ": "frontier"}), "must not be empty"},
		{"unknown class key in profiles", cfgCapProfiles(map[string]string{"genius": "guided"}), "genius"},
		{"unknown profile value", cfgCapProfiles(map[string]string{"frontier": "turbo"}), "turbo"},
		{"valid profiles", cfgCapProfiles(map[string]string{"frontier": "outcome"}), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCapabilities(tc.cfg)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("want nil, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}
