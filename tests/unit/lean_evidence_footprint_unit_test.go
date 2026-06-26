package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// leanGitignore reads the repo-root .gitignore (tests/unit -> ../..).
func leanGitignore(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	b, err := os.ReadFile(filepath.Join(wd, "..", "..", ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	return string(b)
}

// TestGitignoreHasEvidencePatterns asserts the lean-evidence-footprint block
// is present so machine plumbing stays untracked and roadmap.json is kept.
func TestGitignoreHasEvidencePatterns(t *testing.T) {
	gi := leanGitignore(t)
	for _, want := range []string{
		".workflow/*.json",
		"!.workflow/roadmap.json",
		".workflow/*.lock",
	} {
		if !strings.Contains(gi, want) {
			t.Errorf("missing .gitignore pattern %q", want)
		}
	}
}

// TestRoadmapNegationAfterJSONIgnore guards ordering: the negation must come
// after the broad *.json ignore or git would not re-include roadmap.json.
func TestRoadmapNegationAfterJSONIgnore(t *testing.T) {
	gi := leanGitignore(t)
	ignore := strings.Index(gi, ".workflow/*.json")
	negate := strings.Index(gi, "!.workflow/roadmap.json")
	if ignore < 0 || negate < 0 {
		t.Fatal("expected both the json ignore and the roadmap negation")
	}
	if negate < ignore {
		t.Errorf("roadmap negation (%d) must follow the json ignore (%d)", negate, ignore)
	}
}
