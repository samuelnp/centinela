package gates

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// emptyPathCfg returns a config with security gate enabled and both tool
// families configured. Tests must call t.Setenv("PATH", emptyDir) before
// using this so exec.LookPath fails deterministically.
func emptyPathCfg() *config.Config {
	return &config.Config{
		Gates: config.GatesConfig{
			Security: config.SecurityGateConfig{
				Enabled: true,
				Secrets: config.SecretsConfig{Tool: "gitleaks"},
				Vuln:    config.VulnConfig{Tools: []string{"govulncheck", "osv-scanner"}},
			},
		},
	}
}

// TestCheckSecrets_AbsentToolReturnsSkip exercises AC4: gitleaks absent ->
// Skip naming the missing tool, never Fail.
func TestCheckSecrets_AbsentToolReturnsSkip(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	cfg := emptyPathCfg()
	r := checkSecrets(cfg, nil)
	if r.Status != Skip {
		t.Fatalf("absent gitleaks must Skip, got %v", r.Status)
	}
	if !strings.Contains(r.Message, "gitleaks") {
		t.Fatalf("Skip message must name the tool, got %q", r.Message)
	}
}

// TestCheckVuln_AllAbsentReturnsSkip exercises AC4 for vuln side: all tools
// absent -> Skip.
func TestCheckVuln_AllAbsentReturnsSkip(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	cfg := emptyPathCfg()
	r := checkVuln(cfg)
	if r.Status != Skip {
		t.Fatalf("all vuln tools absent must Skip, got %v", r.Status)
	}
}

// TestCheckSecurity_BothAbsentYieldsTwoSkipsWithNoScannersNote exercises the
// both-absent edge case: two Skips + "no scanners available" signal.
func TestCheckSecurity_BothAbsentYieldsTwoSkipsWithNoScannersNote(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	cfg := emptyPathCfg()
	results := checkSecurity(cfg, nil)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Status != Skip {
			t.Fatalf("both must be Skip, got %v (%s)", r.Status, r.Name)
		}
		if !strings.Contains(r.Message, "No security scanners available") {
			t.Fatalf("no-scanners note missing in %q", r.Message)
		}
	}
}

// TestRunWithFilter_SecurityDisabledEmitsNoResults exercises AC5: disabled gate
// must produce zero security results.
func TestRunWithFilter_SecurityDisabledEmitsNoResults(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gates.Security.Enabled = false
	results := RunWithFilter(cfg, nil)
	for _, r := range results {
		if strings.HasPrefix(r.Name, "G-Secrets") || strings.HasPrefix(r.Name, "G-Vuln") {
			t.Fatalf("disabled security gate must emit no results, got %q", r.Name)
		}
	}
}

// TestRunWithFilter_SecurityEnabledBothAbsentAppendsTwo verifies that enabled
// gate with both tools absent still appends exactly two results.
func TestRunWithFilter_SecurityEnabledBothAbsentAppendsTwo(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	cfg := emptyPathCfg()
	results := RunWithFilter(cfg, nil)
	var secCount int
	for _, r := range results {
		if strings.HasPrefix(r.Name, "G-") {
			secCount++
		}
	}
	if secCount != 2 {
		t.Fatalf("expected 2 security results, got %d", secCount)
	}
}
