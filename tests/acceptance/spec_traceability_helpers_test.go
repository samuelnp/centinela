package acceptance_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// stgWrite writes a file under the test's temp CWD, creating parent dirs.
func stgWrite(t *testing.T, rel, body string) {
	t.Helper()
	p := filepath.Join(rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// stgConfig builds a Config with the spec-traceability gate enabled at the
// given severity, pointing at the default specs / tests/acceptance dirs.
func stgConfig(severity string) *config.Config {
	cfg := &config.Config{}
	cfg.Gates.SpecTraceability = config.SpecTraceabilityConfig{
		Enabled:  true,
		SpecDir:  "specs",
		TestDir:  "tests/acceptance",
		Severity: severity,
	}
	return cfg
}

// stgRun runs the gate through the public RunWithFilter surface and returns the
// spec-traceability-gate Result.
func stgRun(t *testing.T, cfg *config.Config, filter *gitdiff.Set) gates.Result {
	t.Helper()
	for _, r := range gates.RunWithFilter(cfg, filter) {
		if r.Name == "spec-traceability-gate" {
			return r
		}
	}
	t.Fatal("spec-traceability-gate result missing")
	return gates.Result{}
}

// detailsJoined concatenates a result's detail lines for substring assertions.
func detailsJoined(r gates.Result) string {
	out := ""
	for _, d := range r.Details {
		out += d + "\n"
	}
	return out
}
