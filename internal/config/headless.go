package config

import (
	"os"
	"strings"
)

// HeadlessConfig is the [headless] section: non-interactive parity for CI/fleet.
// Default zero value (all false) is byte-identical to interactive behavior.
type HeadlessConfig struct {
	Enabled  bool `toml:"enabled"`
	DetectCI bool `toml:"detect_ci"`
}

// IsHeadless resolves headless mode by precedence:
//  1. CENTINELA_HEADLESS env, when set, is AUTHORITATIVE ("1"/"true" → on, else off)
//  2. else [headless] enabled
//  3. else ([headless] detect_ci AND CI env is "1"/"true")
//
// Headless WINS over explicit step_confirmation_mode / plan_advisor_mode — the
// whole point is unattended runs. Default (nothing set) → false.
func IsHeadless(cfg *Config) bool {
	if v := strings.TrimSpace(os.Getenv("CENTINELA_HEADLESS")); v != "" {
		return v == "1" || strings.EqualFold(v, "true")
	}
	if cfg == nil {
		return false
	}
	if cfg.Headless.Enabled {
		return true
	}
	return cfg.Headless.DetectCI && envTrue("CI")
}

func envTrue(key string) bool {
	v := os.Getenv(key)
	return v == "1" || strings.EqualFold(v, "true")
}
