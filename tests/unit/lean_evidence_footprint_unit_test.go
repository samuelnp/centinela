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

// TestGitignoreHasEvidencePatterns asserts role evidence is ignored by explicit
// suffix (the f138f90 fail-safe policy) plus the advisory .lock files.
func TestGitignoreHasEvidencePatterns(t *testing.T) {
	gi := leanGitignore(t)
	for _, want := range []string{
		".workflow/*-big-thinker.json",
		".workflow/*-senior-engineer.json",
		".workflow/*-gatekeeper.json",
		".workflow/*.lock",
	} {
		if !strings.Contains(gi, want) {
			t.Errorf("missing .gitignore pattern %q", want)
		}
	}
}

// TestNoBroadWorkflowJSONIgnore guards the fail-safe policy (f138f90): evidence
// is ignored by explicit role suffix, never a bare ".workflow/*.json" glob that
// would silently drop durable state (roadmap.json, per-feature ledgers).
func TestNoBroadWorkflowJSONIgnore(t *testing.T) {
	gi := "\n" + leanGitignore(t) + "\n"
	if strings.Contains(gi, "\n.workflow/*.json\n") {
		t.Errorf("broad %q ignore would drop durable .workflow state", ".workflow/*.json")
	}
}
