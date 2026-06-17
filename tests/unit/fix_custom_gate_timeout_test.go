package unit_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// TestCustomGateTimeoutReturnsPromptly is the regression guard for the
// fix-custom-gate-timeout hotfix: a custom gate whose command hangs far longer
// than its timeout must fail (timed out) AND let validate return promptly,
// rather than blocking for the full command duration. Before the cmd.WaitDelay
// fix in runCustom, a grandchild `sleep` inheriting the output pipe kept the
// gate blocked for the whole sleep (CI saw 5.00s), defeating the timeout.
func TestCustomGateTimeoutReturnsPromptly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses the POSIX `sleep` shell builtin")
	}
	cfg := &config.Config{}
	cfg.Gates.CustomGates = []config.CustomGate{{
		Enabled:        true,
		Name:           "hang",
		Command:        "sleep 30",
		Severity:       "fail",
		Output:         "blob",
		TimeoutSeconds: 1,
	}}

	start := time.Now()
	results := gates.RunWithFilter(cfg, nil)
	elapsed := time.Since(start)

	if elapsed > 5*time.Second {
		t.Fatalf("custom gate timeout did not return promptly: took %s", elapsed)
	}
	var found bool
	for _, r := range results {
		if r.Name == "hang" {
			found = true
			if r.Status != gates.Fail {
				t.Fatalf("hung custom gate should Fail, got status %v", r.Status)
			}
		}
	}
	if !found {
		t.Fatalf("custom gate %q missing from results: %+v", "hang", results)
	}
}
