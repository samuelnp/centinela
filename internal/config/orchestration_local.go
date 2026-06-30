package config

import "strings"

// LocalConfig declares a local-model backend under [orchestration.local]. It is
// a pure leaf shape: the runner consumes Endpoint/Model/APIKeyEnv as opaque
// strings (availability is the runner's job, never verified at config load).
// Provider is a fixed vocabulary (normalized trim+lower); the other fields are
// opaque (trimmed only).
type LocalConfig struct {
	Provider  string `toml:"provider"`
	Endpoint  string `toml:"endpoint"`
	Model     string `toml:"model"`
	APIKeyEnv string `toml:"api_key_env"`
}

// allowedLocalProviders enumerates the supported provider kinds. Both kinds emit
// an OpenCode provider block via npm @ai-sdk/openai-compatible; they differ only
// in whether an api_key reference is written.
var allowedLocalProviders = []string{"ollama", "openai-compatible"}

// normProvider normalizes a provider value: trim + lowercase (fixed vocabulary,
// unlike the opaque endpoint/model strings which are trimmed only).
func normProvider(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// isAllowedLocalProvider reports whether a normalized provider is supported.
func isAllowedLocalProvider(provider string) bool {
	for _, p := range allowedLocalProviders {
		if p == provider {
			return true
		}
	}
	return false
}

// LocalProviderConfig returns the trimmed/normalized local block and whether it
// is set (provider non-empty after normalization). Opaque fields are trimmed;
// provider is normalized. nil-safe.
func LocalProviderConfig(cfg *Config) (LocalConfig, bool) {
	if cfg == nil {
		return LocalConfig{}, false
	}
	lc := cfg.Orchestration.Local
	out := LocalConfig{
		Provider:  normProvider(lc.Provider),
		Endpoint:  strings.TrimSpace(lc.Endpoint),
		Model:     strings.TrimSpace(lc.Model),
		APIKeyEnv: strings.TrimSpace(lc.APIKeyEnv),
	}
	return out, out.Provider != ""
}
