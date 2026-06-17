package config

import (
	"fmt"
	"strings"
)

// defaultBaselinePath is where `centinela audit baseline` writes the committed
// ratchet snapshot when no path is configured.
const defaultBaselinePath = ".workflow/audit-baseline.json"

// AuditBaselineConfig controls the optional audit-baseline ratchet gate. When
// Enabled, validate compares the current full-scan violations against the
// recorded baseline and maps new violations to Severity ("fail"|"warn").
// TargetGates optionally restricts which gates participate (empty = all
// detail-emitting gates).
type AuditBaselineConfig struct {
	Enabled      bool     `toml:"enabled"`
	Severity     string   `toml:"severity"`
	BaselinePath string   `toml:"baseline_path"`
	TargetGates  []string `toml:"target_gates"`
}

// NormalizeAuditBaseline trims whitespace and fills unset fields with safe
// defaults: severity warn (surface, don't block during adoption) and the
// standard baseline path.
func NormalizeAuditBaseline(c AuditBaselineConfig) AuditBaselineConfig {
	c.Severity = strings.TrimSpace(c.Severity)
	if c.Severity == "" {
		c.Severity = "warn"
	}
	c.BaselinePath = strings.TrimSpace(c.BaselinePath)
	if c.BaselinePath == "" {
		c.BaselinePath = defaultBaselinePath
	}
	return c
}

// validateAuditBaseline rejects an unknown severity so a typo cannot silently
// change the gate's strictness. It is a no-op when the gate is disabled.
func validateAuditBaseline(c AuditBaselineConfig) error {
	if !c.Enabled {
		return nil
	}
	if c.Severity != "fail" && c.Severity != "warn" {
		return fmt.Errorf("gates.audit_baseline.severity must be fail or warn, got %q", c.Severity)
	}
	return nil
}
