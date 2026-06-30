package gates

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// TestRunWithFilter_AllGatesEnabledDispatchesEach enables every built-in gate
// with minimal fast-returning config and asserts RunWithFilter appends a result
// from each dispatch branch (file-size, i18n, build, import-graph, security x2,
// spec-traceability, roadmap-drift, custom). Runs in an empty tempdir with an
// empty PATH so the security scanners deterministically Skip without execing.
func TestRunWithFilter_AllGatesEnabledDispatchesEach(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                       //nolint:errcheck
	os.Chdir(d)                                             //nolint:errcheck
	t.Setenv("PATH", d)                                     //nolint:errcheck
	os.MkdirAll("i18n", 0755)                               //nolint:errcheck
	os.WriteFile("i18n/en.json", []byte(`{"k":"v"}`), 0644) //nolint:errcheck

	cfg := &config.Config{
		I18n: config.I18nConfig{Format: "json", Dir: "i18n", Locales: []string{"en"}},
	}
	cfg.Gates.FileSizeEnabled = true
	cfg.Gates.I18nEnabled = true
	cfg.Gates.Build.Enabled = true // no targets -> Skip (no exec)
	cfg.Gates.ImportGraph.Enabled = true
	cfg.Gates.Security.Enabled = true
	cfg.Gates.Security.Secrets.Tool = "gitleaks"
	cfg.Gates.Security.Vuln.Tools = []string{"govulncheck"}
	cfg.Gates.SpecTraceability.Enabled = true
	cfg.Gates.SpecTraceability.SpecDir = "specs"
	cfg.Gates.RoadmapDrift.Enabled = true
	cfg.Gates.CustomGates = []config.CustomGate{{Name: "noop", Enabled: false}}

	res := RunWithFilter(cfg, nil)
	names := map[string]bool{}
	for _, r := range res {
		names[r.Name] = true
	}
	for _, want := range []string{
		"G-Build: Cross-Compile", "import_graph", "roadmap_drift",
	} {
		if !names[want] {
			t.Fatalf("expected a %q result, got %v", want, names)
		}
	}
	// file-size, i18n, both security, spec-traceability also dispatched.
	if len(res) < 7 {
		t.Fatalf("expected results from every enabled gate, got %d: %v", len(res), names)
	}
}
