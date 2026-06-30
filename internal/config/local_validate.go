package config

import (
	"fmt"
	"strings"
)

// validateLocalConfig shape-validates [orchestration.local]. The block is
// all-or-nothing: with no provider, every other field must also be empty;
// otherwise the provider must be in the allow-list and endpoint + model must be
// non-empty after trim. api_key_env is optional and is never resolved here
// (availability is the runner's job). Each error names the offending key.
func validateLocalConfig(cfg *Config) error {
	if cfg == nil {
		return nil
	}
	lc := cfg.Orchestration.Local
	provider := normProvider(lc.Provider)
	endpoint := strings.TrimSpace(lc.Endpoint)
	model := strings.TrimSpace(lc.Model)
	if provider == "" {
		if endpoint != "" || model != "" || strings.TrimSpace(lc.APIKeyEnv) != "" {
			return fmt.Errorf("orchestration.local.provider must not be empty when an endpoint, model, or api_key_env is set")
		}
		return nil
	}
	if !isAllowedLocalProvider(provider) {
		return fmt.Errorf("orchestration.local.provider %q unsupported (allowed: %s)",
			lc.Provider, strings.Join(allowedLocalProviders, ", "))
	}
	if endpoint == "" {
		return fmt.Errorf("orchestration.local.endpoint must not be empty")
	}
	if model == "" {
		return fmt.Errorf("orchestration.local.model must not be empty")
	}
	return nil
}
