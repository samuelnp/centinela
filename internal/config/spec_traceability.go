package config

import (
	"fmt"
	"strings"
)

// SpecTraceabilityConfig controls the spec-traceability gate. When Enabled, the
// gate maps every in-scope Gherkin scenario in SpecDir to a covering acceptance
// test in TestDir. Severity ("fail"|"warn") is the adoption knob: warn surfaces
// uncovered scenarios without blocking, fail rejects them.
type SpecTraceabilityConfig struct {
	Enabled  bool   `toml:"enabled"`
	SpecDir  string `toml:"spec_dir"`
	TestDir  string `toml:"test_dir"`
	Severity string `toml:"severity"`
}

// NormalizeSpecTraceability trims whitespace and fills unset directory and
// severity fields with their defaults (specs, tests/acceptance, fail).
func NormalizeSpecTraceability(s SpecTraceabilityConfig) SpecTraceabilityConfig {
	s.SpecDir = strings.TrimSpace(s.SpecDir)
	s.TestDir = strings.TrimSpace(s.TestDir)
	s.Severity = strings.TrimSpace(s.Severity)
	if s.SpecDir == "" {
		s.SpecDir = "specs"
	}
	if s.TestDir == "" {
		s.TestDir = "tests/acceptance"
	}
	if s.Severity == "" {
		s.Severity = "fail"
	}
	return s
}

// validateSpecTraceability rejects an unknown severity so a typo cannot silently
// change the gate's strictness. It is a no-op when the gate is disabled.
func validateSpecTraceability(s SpecTraceabilityConfig) error {
	if !s.Enabled {
		return nil
	}
	if s.Severity != "fail" && s.Severity != "warn" {
		return fmt.Errorf("gates.spec_traceability.severity must be fail or warn, got %q", s.Severity)
	}
	return nil
}
