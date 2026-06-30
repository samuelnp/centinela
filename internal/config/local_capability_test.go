package config

import "testing"

func localPlusCaps(caps map[string]string) *Config {
	c := cfgLocal("ollama", "http://x/v1", "qwen2.5-coder", "")
	c.Orchestration.Capabilities = caps
	return c
}

// LocalDefaultClass: hit only when id==local.Model and the id is unmapped; miss
// for empty/whitespace id, nil cfg, no local block, id != local model, or an id
// already classed by an explicit mapping or the builtin map.
func TestLocalDefaultClass(t *testing.T) {
	local := cfgLocal("ollama", "http://x/v1", "qwen2.5-coder", "")
	cases := []struct {
		name, id string
		cfg      *Config
		want     string
		wantOK   bool
	}{
		{"hit unmapped local model", "qwen2.5-coder", local, CapabilityLimited, true},
		{"miss empty id", "", local, "", false},
		{"miss whitespace id", "   ", local, "", false},
		{"miss nil cfg", "qwen2.5-coder", nil, "", false},
		{"miss no local block", "qwen2.5-coder", &Config{}, "", false},
		{"miss id != local model", "other", local, "", false},
		{"miss explicitly mapped", "qwen2.5-coder", localPlusCaps(map[string]string{"qwen2.5-coder": "capable"}), "", false},
		{"miss builtin mapped", "claude-opus-4-7", cfgLocal("ollama", "http://x/v1", "claude-opus-4-7", ""), "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := LocalDefaultClass(tc.id, tc.cfg)
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("got (%q,%v) want (%q,%v)", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

// DefaultProfileForModel local fallback: an unmapped declared local model defaults
// limited → strict; honors a [capability_profiles] limited override; a non-local
// unmapped id still returns ok=false (caller must not engage the capability tier).
func TestDefaultProfileForModelLocalFallback(t *testing.T) {
	local := cfgLocal("ollama", "http://x/v1", "qwen2.5-coder", "")
	if p, ok := DefaultProfileForModel("qwen2.5-coder", local); !ok || p != ProfileStrict {
		t.Fatalf("local fallback: got (%q,%v) want (strict,true)", p, ok)
	}
	override := cfgLocal("ollama", "http://x/v1", "qwen2.5-coder", "")
	override.Orchestration.CapabilityProfiles = map[string]string{"limited": "guided"}
	if p, ok := DefaultProfileForModel("qwen2.5-coder", override); !ok || p != ProfileGuided {
		t.Fatalf("limited override: got (%q,%v) want (guided,true)", p, ok)
	}
	if p, ok := DefaultProfileForModel("some/unmapped", &Config{}); ok || p != "" {
		t.Fatalf("non-local unmapped: got (%q,%v) want (\"\",false)", p, ok)
	}
}
