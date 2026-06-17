package gates

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func stgCfg(t *testing.T, severity string) *config.Config {
	t.Helper()
	cfg := &config.Config{}
	cfg.Gates.SpecTraceability = config.SpecTraceabilityConfig{
		Enabled: true, SpecDir: "specs", TestDir: "tests/acceptance", Severity: severity,
	}
	return cfg
}

func TestCheckSpecTraceability_CoveredPasses(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFeature(t, "specs", "x.feature", "Feature: f\n  Scenario: Go\n")
	writeGo(t, "tests/acceptance", "x_test.go", "// Acceptance: specs/x.feature\n// Scenario: Go\n")
	r := checkSpecTraceability(stgCfg(t, "fail"), nil)
	if r.Status != Pass || !strings.Contains(r.Message, "1 scenarios") {
		t.Fatalf("want Pass with count, got %v %q", r.Status, r.Message)
	}
}

func TestCheckSpecTraceability_UncoveredFailsWithDetails(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFeature(t, "specs", "x.feature", "Feature: f\n  Scenario: Go\n")
	r := checkSpecTraceability(stgCfg(t, "fail"), nil)
	if r.Status != Fail || !strings.Contains(strings.Join(r.Details, "\n"), `specs/x.feature: "Go"`) {
		t.Fatalf("want Fail naming gap, got %v %v", r.Status, r.Details)
	}
}

func TestCheckSpecTraceability_WarnSeverityWarns(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFeature(t, "specs", "x.feature", "Feature: f\n  Scenario: Go\n")
	if r := checkSpecTraceability(stgCfg(t, "warn"), nil); r.Status != Warn {
		t.Fatalf("warn severity must Warn not Fail, got %v", r.Status)
	}
}

func TestCheckSpecTraceability_NoSpecsSkips(t *testing.T) {
	t.Chdir(t.TempDir())
	if r := checkSpecTraceability(stgCfg(t, "fail"), nil); r.Status != Skip {
		t.Fatalf("want Skip when no specs, got %v %q", r.Status, r.Message)
	}
}

func TestCheckSpecTraceability_ParseErrorFails(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("specs", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if r := checkSpecTraceability(stgCfg(t, "fail"), nil); r.Status != Fail {
		t.Fatalf("spec-dir read error must Fail, got %v", r.Status)
	}
}

func TestCheckSpecTraceability_CoverageErrorFails(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFeature(t, "specs", "x.feature", "Feature: f\n  Scenario: Go\n")
	if err := os.WriteFile("tests", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := stgCfg(t, "fail")
	cfg.Gates.SpecTraceability.TestDir = "tests/acceptance"
	if r := checkSpecTraceability(cfg, nil); r.Status != Fail {
		t.Fatalf("test-dir read error must Fail, got %v", r.Status)
	}
}

func TestReportTraceability_EmptyGapsPasses(t *testing.T) {
	r := reportTraceability([]Scenario{{Spec: "a", Name: "x"}}, nil, "fail")
	if r.Status != Pass || !strings.Contains(r.Message, "1 scenarios") {
		t.Fatalf("no gaps must Pass with count, got %v %q", r.Status, r.Message)
	}
}
