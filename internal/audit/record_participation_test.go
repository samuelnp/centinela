package audit

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// TestParticipatingGatesEmptyMeansAll: an empty allowlist participates in every
// default detail-emitting gate.
func TestParticipatingGatesEmptyMeansAll(t *testing.T) {
	cfg := &config.Config{}
	p := participatingGates(cfg)
	for _, name := range defaultParticipants {
		if !p[name] {
			t.Fatalf("%q should participate by default", name)
		}
	}
	if !isParticipating("G1: File Size", cfg) {
		t.Fatal("isParticipating disagrees with the set")
	}
}

// TestParticipatingGatesAllowlist restricts participation to the configured
// target_gates intersected with the defaults.
func TestParticipatingGatesAllowlist(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gates.AuditBaseline.TargetGates = []string{"G1: File Size", "not-a-default"}
	p := participatingGates(cfg)
	if !p["G1: File Size"] {
		t.Fatal("allowlisted default missing")
	}
	if p["G11: i18n"] {
		t.Fatal("non-allowlisted default should be excluded")
	}
	if isParticipating("not-a-default", cfg) {
		t.Fatal("non-default name should not participate even if allowlisted")
	}
}

// TestRecordCapturesOversizedFile records a real full-scan and names the
// oversized file as a baselined violation.
func TestRecordCapturesOversizedFile(t *testing.T) {
	cfg := tempRepo(t, "fail", map[string]string{"internal/big.go": oversizedGo(0)})
	b := Record(cfg)
	if b.Scheme != fingerprintScheme || b.Version != 1 {
		t.Fatalf("bad header: %+v", b)
	}
	if !recordHasKey(b, "internal/big.go") {
		t.Fatalf("big.go not recorded: %+v", b.Gates)
	}
}

func recordHasKey(b Baseline, key string) bool {
	for _, e := range b.Gates {
		if containsKey(e.Fingerprints, key) {
			return true
		}
	}
	return false
}
