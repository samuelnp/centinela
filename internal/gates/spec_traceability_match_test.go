package gates

import (
	"os"
	"path/filepath"
	"testing"
)

func writeGo(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCoveredScenarios_HeaderAnnotationAndTypoAndNormalize(t *testing.T) {
	d := t.TempDir()
	writeGo(t, d, "a_test.go", "// Acceptance: spec/foo.feature (AC1)\n// Scenario:  Do  IT .\n")
	covered, err := coveredScenarios(d)
	if err != nil {
		t.Fatal(err)
	}
	if !covered["foo"]["do it"] {
		t.Fatalf("annotation+spec/-typo+normalize must record coverage: %+v", covered)
	}
}

func TestCoveredScenarios_CommentWithoutHeaderIgnored(t *testing.T) {
	d := t.TempDir()
	writeGo(t, d, "a_test.go", "// Scenario: Orphan\n")
	covered, err := coveredScenarios(d)
	if err != nil {
		t.Fatal(err)
	}
	if len(covered) != 0 {
		t.Fatalf("comment with no header above it must be ignored: %+v", covered)
	}
}

func TestCoveredScenarios_MissingDirIsEmpty(t *testing.T) {
	covered, err := coveredScenarios(filepath.Join(t.TempDir(), "nope"))
	if err != nil || len(covered) != 0 {
		t.Fatalf("missing test dir must be empty map, got %v %v", covered, err)
	}
}

func TestCoveredScenarios_NonGoFilesSkipped(t *testing.T) {
	d := t.TempDir()
	writeGo(t, d, "notes.md", "// Acceptance: specs/foo.feature\n// Scenario: Trap\n")
	covered, err := coveredScenarios(d)
	if err != nil {
		t.Fatal(err)
	}
	if len(covered) != 0 {
		t.Fatalf("non-.go files must not contribute coverage: %+v", covered)
	}
}

func TestCoveredScenarios_TestDirIsFileReturnsError(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "acc")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := coveredScenarios(p); err == nil {
		t.Fatal("a non-existent-dir error (path is a file) must propagate")
	}
}

func TestUncovered_PartitionsBySlugAndName(t *testing.T) {
	scenarios := []Scenario{{Spec: "foo", Name: "Covered"}, {Spec: "foo", Name: "Gap"}, {Spec: "bar", Name: "Covered"}}
	covered := map[string]map[string]bool{"foo": {"covered": true}}
	got := uncovered(scenarios, covered)
	if len(got) != 2 {
		t.Fatalf("want 2 uncovered (foo Gap, bar Covered), got %+v", got)
	}
	if got[0].Name != "Gap" || got[1].Spec != "bar" {
		t.Fatalf("uncovered partition wrong: %+v", got)
	}
}
