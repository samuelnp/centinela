package config

import (
	"os"
	"strings"
)

// DriverModelFrom resolves the driver model id that keys a workflow's default
// enforcement profile. Precedence (highest → lowest): the --model flag value,
// then the CENTINELA_MODEL env var, then [orchestration] driver_model. Each
// candidate is trimmed; the first non-empty after trim wins. Empty when nothing
// is configured. The id is opaque — no existence check is performed.
func DriverModelFrom(flagModel string, cfg *Config) string {
	candidates := []string{flagModel, os.Getenv("CENTINELA_MODEL")}
	if cfg != nil {
		// [orchestration] driver_model outranks the declared local model: the
		// local block's model is the lowest-precedence candidate. Empty is
		// trimmed away, so the zero-config path is unaffected.
		candidates = append(candidates, cfg.Orchestration.DriverModel, cfg.Orchestration.Local.Model)
	}
	for _, c := range candidates {
		if v := strings.TrimSpace(c); v != "" {
			return v
		}
	}
	return ""
}
