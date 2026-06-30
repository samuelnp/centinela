package config

import (
	"strings"
	"testing"
)

func cfgLocal(provider, endpoint, model, apiKeyEnv string) *Config {
	c := &Config{}
	c.Orchestration.Local = LocalConfig{Provider: provider, Endpoint: endpoint, Model: model, APIKeyEnv: apiKeyEnv}
	return c
}

// validateLocalConfig: all-empty valid; valid ollama/openai-compatible; unknown
// provider lists allowed; endpoint/model empty when provider set; provider empty
// while other fields set; nil cfg; provider normalized before validation.
func TestValidateLocalConfig(t *testing.T) {
	cases := []struct {
		name      string
		cfg       *Config
		wantErr   bool
		errSubstr string
	}{
		{"nil cfg", nil, false, ""},
		{"all empty", cfgLocal("", "", "", ""), false, ""},
		{"valid ollama", cfgLocal("ollama", "http://x/v1", "m", ""), false, ""},
		{"valid openai-compatible", cfgLocal("openai-compatible", "http://x/v1", "m", "K"), false, ""},
		{"unknown provider lists allowed", cfgLocal("groq", "http://x/v1", "m", ""), true, "ollama, openai-compatible"},
		{"endpoint empty", cfgLocal("ollama", "  ", "m", ""), true, "endpoint"},
		{"model empty", cfgLocal("ollama", "http://x/v1", "", ""), true, "model"},
		{"provider empty but endpoint set", cfgLocal("", "http://x/v1", "", ""), true, "provider"},
		{"provider empty but model set", cfgLocal("", "", "m", ""), true, "provider"},
		{"provider empty but apikey set", cfgLocal("", "", "", "K"), true, "provider"},
		{"provider normalized then valid", cfgLocal("  Ollama  ", "http://x/v1", "m", ""), false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateLocalConfig(tc.cfg)
			if tc.wantErr != (err != nil) {
				t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
			}
			if err != nil && tc.errSubstr != "" && !strings.Contains(err.Error(), tc.errSubstr) {
				t.Fatalf("err %q missing %q", err.Error(), tc.errSubstr)
			}
		})
	}
}

// The unknown-provider error names the offending raw value and the provider key.
func TestValidateLocalConfigUnknownNamesKey(t *testing.T) {
	err := validateLocalConfig(cfgLocal("groq", "http://x/v1", "m", ""))
	if err == nil || !strings.Contains(err.Error(), "groq") || !strings.Contains(err.Error(), "provider") {
		t.Fatalf("expected provider/groq error, got %v", err)
	}
}
