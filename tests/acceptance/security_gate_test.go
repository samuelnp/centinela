package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// Acceptance for specs/security-gate.feature (AC1, AC2, AC4, AC5). Drives the
// REAL security gate through gates.RunWithFilter with fake scanner binaries on
// PATH, asserting G-Secrets hard-Fail / Pass / Skip semantics and the
// disabled-gate zero-config-safe behavior. Shared helpers live in
// security_gate_helpers_test.go; AC3/AC6/AC7 in security_gate_more_test.go.

// AC1: enabled + a detectable secret -> G-Secrets Fail and AllPassed false.
func TestAccept_Security_SecretDetected_Fails(t *testing.T) {
	dir := secPath(t)
	secBin(t, dir, "gitleaks", `printf '[{"RuleID":"generic-api-key","File":"app.go"}]' > "$6"
exit 1`)
	results := gates.RunWithFilter(secCfg(), nil)
	r, ok := secResult(t, results, "G-Secrets")
	if !ok || r.Status != gates.Fail {
		t.Fatalf("AC1: G-Secrets must Fail, got ok=%v %v: %q", ok, r.Status, r.Message)
	}
	if gates.AllPassed(results) {
		t.Fatal("AC1: AllPassed must be false when a secret is detected")
	}
}

// AC2: secrets scanner finds nothing -> G-Secrets Pass.
func TestAccept_Security_NoSecret_Passes(t *testing.T) {
	dir := secPath(t)
	secBin(t, dir, "gitleaks", `printf '[]' > "$6"
exit 0`)
	r, ok := secResult(t, gates.RunWithFilter(secCfg(), nil), "G-Secrets")
	if !ok || r.Status != gates.Pass {
		t.Fatalf("AC2: G-Secrets must Pass on clean scan, got %v: %q", r.Status, r.Message)
	}
}

// AC4: gitleaks absent (empty PATH) -> G-Secrets Skip naming the tool, not Fail.
func TestAccept_Security_GitleaksAbsent_Skips(t *testing.T) {
	secPath(t) // empty PATH dir; no fake gitleaks dropped
	r, ok := secResult(t, gates.RunWithFilter(secCfg(), nil), "G-Secrets")
	if !ok {
		t.Fatal("AC4: G-Secrets result must be present")
	}
	if r.Status != gates.Skip || !strings.Contains(r.Message, "gitleaks") {
		t.Fatalf("AC4: must Skip naming gitleaks, got %v: %q", r.Status, r.Message)
	}
}

// AC5: gate disabled -> no G-Secrets / G-Vuln results emitted.
func TestAccept_Security_Disabled_EmitsNothing(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gates.Security.Enabled = false
	results := gates.RunWithFilter(cfg, nil)
	if _, ok := secResult(t, results, "G-Secrets"); ok {
		t.Fatal("AC5: disabled gate must emit no G-Secrets result")
	}
	if _, ok := secResult(t, results, "G-Vuln"); ok {
		t.Fatal("AC5: disabled gate must emit no G-Vuln result")
	}
}
