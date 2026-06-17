package config

import (
	"fmt"
	"strings"
)

// CustomGate is a command-backed gate declared via [[gates.custom]]. It produces
// the same gates.Result contract as a built-in: exit 0 passes, non-zero fails
// (or warns), and the combined output flows into the report and audit ratchet.
type CustomGate struct {
	Enabled        bool   `toml:"enabled"`
	Name           string `toml:"name"`
	Command        string `toml:"command"`
	Severity       string `toml:"severity"`        // fail | warn (default fail)
	Output         string `toml:"output"`          // blob | lines (default blob)
	TimeoutSeconds int    `toml:"timeout_seconds"` // default 60 when <= 0
	DiffAware      bool   `toml:"diff_aware"`
}

// builtinGateNames are the reserved built-in gate Names a custom gate must not
// shadow (case-sensitive exact match), so its Details never collide with a
// built-in's in the report or the audit fingerprinter.
var builtinGateNames = map[string]struct{}{
	"G1: File Size":            {},
	"G11: i18n":                {},
	"G-Build: Cross-Compile":   {},
	"import_graph":             {},
	"G-Secrets: Secret Scan":   {},
	"G-Vuln: Dependency Audit": {},
	"spec-traceability-gate":   {},
	"roadmap_drift":            {},
	"audit_baseline":           {},
}

// NormalizeCustomGates trims string fields and fills defaults (severity fail,
// output blob, timeout 60) so downstream consumers never see a zero value.
func NormalizeCustomGates(gs []CustomGate) []CustomGate {
	for i := range gs {
		gs[i].Name = strings.TrimSpace(gs[i].Name)
		gs[i].Command = strings.TrimSpace(gs[i].Command)
		gs[i].Severity = strings.TrimSpace(gs[i].Severity)
		gs[i].Output = strings.TrimSpace(gs[i].Output)
		if gs[i].Severity == "" {
			gs[i].Severity = "fail"
		}
		if gs[i].Output == "" {
			gs[i].Output = "blob"
		}
		if gs[i].TimeoutSeconds <= 0 {
			gs[i].TimeoutSeconds = 60
		}
	}
	return gs
}

// validateCustomGates rejects malformed custom gates with indexed errors. It
// validates every configured entry (so a typo surfaces even before the gate is
// enabled), mirroring the array-of-tables validation in file_size_exceptions.go.
func validateCustomGates(gs []CustomGate) error {
	seen := make(map[string]int, len(gs))
	for i, g := range gs {
		if g.Name == "" {
			return fmt.Errorf("gates.custom[%d].name is required", i)
		}
		if j, ok := seen[g.Name]; ok {
			return fmt.Errorf("gates.custom[%d].name %q duplicates gates.custom[%d]", i, g.Name, j)
		}
		seen[g.Name] = i
		if _, ok := builtinGateNames[g.Name]; ok {
			return fmt.Errorf("gates.custom[%d].name %q collides with built-in gate", i, g.Name)
		}
		if g.Command == "" {
			return fmt.Errorf("gates.custom[%d].command is required", i)
		}
		if g.Severity != "fail" && g.Severity != "warn" {
			return fmt.Errorf("gates.custom[%d].severity must be fail or warn, got %q", i, g.Severity)
		}
		if g.Output != "blob" && g.Output != "lines" {
			return fmt.Errorf("gates.custom[%d].output must be blob or lines, got %q", i, g.Output)
		}
	}
	return nil
}
