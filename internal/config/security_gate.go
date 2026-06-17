package config

import (
	"fmt"
	"strings"
)

// SecretsTool is the only supported secret scanner in v1.
const SecretsTool = "gitleaks"

// VulnTools enumerates the supported dependency-vuln scanners in v1.
var VulnTools = []string{"govulncheck", "osv-scanner"}

// SecretsConfig configures the hard-fail secret scanner.
type SecretsConfig struct {
	Tool      string   `toml:"tool"`
	Allowlist []string `toml:"allowlist"`
}

// VulnConfig configures the warn-only dependency vulnerability audit.
type VulnConfig struct {
	Tools []string `toml:"tools"`
}

// SecurityGateConfig controls the opt-in [gates.security] gate. It is off by
// default (Enabled=false) so existing zero-config projects are unaffected.
type SecurityGateConfig struct {
	Enabled bool          `toml:"enabled"`
	Secrets SecretsConfig `toml:"secrets"`
	Vuln    VulnConfig    `toml:"vuln"`
}

// NormalizeSecurityGate trims whitespace and applies defaults. When enabled
// with no secrets tool, it falls back to the v1 default (gitleaks); when
// enabled with no vuln tools, it falls back to the full supported set.
func NormalizeSecurityGate(s SecurityGateConfig) SecurityGateConfig {
	s.Secrets.Tool = strings.TrimSpace(s.Secrets.Tool)
	s.Secrets.Allowlist = trimNonEmpty(s.Secrets.Allowlist)
	s.Vuln.Tools = trimNonEmpty(s.Vuln.Tools)
	if !s.Enabled {
		return s
	}
	if s.Secrets.Tool == "" {
		s.Secrets.Tool = SecretsTool
	}
	if len(s.Vuln.Tools) == 0 {
		s.Vuln.Tools = append([]string(nil), VulnTools...)
	}
	return s
}

// validateSecurityGate rejects unknown tool names so a typo cannot silently
// disable scanning. It is a no-op when the gate is disabled.
func validateSecurityGate(s SecurityGateConfig) error {
	if !s.Enabled {
		return nil
	}
	if s.Secrets.Tool != SecretsTool {
		return fmt.Errorf("gates.security.secrets.tool must be %q, got %q", SecretsTool, s.Secrets.Tool)
	}
	for i, t := range s.Vuln.Tools {
		if !knownVulnTool(t) {
			return fmt.Errorf("gates.security.vuln.tools[%d] is unknown: %q", i, t)
		}
	}
	return nil
}

func knownVulnTool(name string) bool {
	for _, t := range VulnTools {
		if t == name {
			return true
		}
	}
	return false
}
