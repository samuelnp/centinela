package audit

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func cfgWithCustom(names ...string) *config.Config {
	cfg := &config.Config{}
	for _, n := range names {
		cfg.Gates.CustomGates = append(cfg.Gates.CustomGates,
			config.CustomGate{Name: n, Command: "true", Severity: "fail", Output: "lines"})
	}
	return cfg
}

// TestCustomGateParticipatesByDefault is the bug-fix regression guard: a
// configured custom-gate Name participates in the ratchet with no target_gates.
func TestCustomGateParticipatesByDefault(t *testing.T) {
	cfg := cfgWithCustom("no-console-log")
	if !participatingGates(cfg)["no-console-log"] {
		t.Fatal("custom gate should participate by default")
	}
	if !isParticipating("no-console-log", cfg) {
		t.Fatal("isParticipating disagrees for the custom gate")
	}
	// Defaults still participate alongside the custom gate.
	if !isParticipating("G1: File Size", cfg) {
		t.Fatal("defaults should still participate when a custom gate is added")
	}
}

// TestCustomGateHonorsTargetGates: target_gates restricts participation; a custom
// gate outside the allowlist is excluded, one inside it is kept.
func TestCustomGateHonorsTargetGates(t *testing.T) {
	cfg := cfgWithCustom("kept", "dropped")
	cfg.Gates.AuditBaseline.TargetGates = []string{"kept"}
	p := participatingGates(cfg)
	if !p["kept"] {
		t.Fatal("allowlisted custom gate should participate")
	}
	if p["dropped"] {
		t.Fatal("non-allowlisted custom gate should be excluded")
	}
	if p["G1: File Size"] {
		t.Fatal("non-allowlisted default should be excluded too")
	}
}
