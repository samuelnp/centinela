package gates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/gitdiff"
)

func writeFeature(t *testing.T, dir, name, body string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParseScenarios_OutlineAndPlainCountedWithSlug(t *testing.T) {
	d := t.TempDir()
	writeFeature(t, d, "watch.feature", "Feature: f\n  Scenario: Start it\n  Scenario Outline: Run rows\n    Examples:\n")
	got, err := parseScenarios(d, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 scenarios, got %d: %+v", len(got), got)
	}
	if got[0].Spec != "watch" || got[0].Name != "Start it" || got[1].Name != "Run rows" {
		t.Fatalf("slug/name wrong: %+v", got)
	}
}

func TestParseScenarios_DiffFilterIncludesAndExcludes(t *testing.T) {
	d := t.TempDir()
	in := writeFeature(t, d, "in.feature", "Feature: f\n  Scenario: Yes\n")
	writeFeature(t, d, "out.feature", "Feature: f\n  Scenario: No\n")
	got, err := parseScenarios(d, gitdiff.NewSet([]string{in}))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Yes" {
		t.Fatalf("filter must keep only changed file, got %+v", got)
	}
}

func TestParseScenarios_MissingDirIsNotInScope(t *testing.T) {
	got, err := parseScenarios(filepath.Join(t.TempDir(), "nope"), nil)
	if err != nil || got != nil {
		t.Fatalf("missing dir must be (nil,nil), got %v %v", got, err)
	}
}

func TestParseScenarios_SpecDirIsFileReturnsError(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "specs")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := parseScenarios(p, nil); err == nil {
		t.Fatal("a non-existent-dir error (path is a file) must propagate")
	}
}

func TestScanScenarios_UnreadablePathTolerated(t *testing.T) {
	if got := scanScenarios(filepath.Join(t.TempDir(), "ghost.feature"), "ghost"); got != nil {
		t.Fatalf("unreadable file must yield nil, got %+v", got)
	}
}

func TestParseScenarios_NonFeatureFilesAndMalformedTolerated(t *testing.T) {
	d := t.TempDir()
	writeFeature(t, d, "ok.feature", "Feature: f\n  Scenario: Real\n")
	writeFeature(t, d, "readme.txt", "not a feature with Scenario: trap\n")
	writeFeature(t, d, "empty.feature", "Feature: f\n  # no scenarios here\n")
	got, err := parseScenarios(d, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Real" {
		t.Fatalf("only real .feature scenario expected, got %+v", got)
	}
}

func TestNormalizeScenario_TrimCollapsePeriodLower(t *testing.T) {
	if got := normalizeScenario("  Start   the WATCHER . "); got != "start the watcher" {
		t.Fatalf("normalization wrong: %q", got)
	}
}
