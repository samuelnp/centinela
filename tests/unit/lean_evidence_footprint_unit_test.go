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

// TestGitignoreHasEvidencePatterns asserts the role-suffixed evidence + lock
// patterns are present so per-role machine plumbing stays untracked.
func TestGitignoreHasEvidencePatterns(t *testing.T) {
	gi := leanGitignore(t)
	for _, want := range []string{
		".workflow/*-big-thinker.json",
		".workflow/*-gatekeeper.json",
		".workflow/*.lock",
	} {
		if !strings.Contains(gi, want) {
			t.Errorf("missing .gitignore pattern %q", want)
		}
	}
}

// TestGitignoreKeepsDurableState guards the fix for the over-broad *.json rule:
// the gitignore must NOT contain a bare `.workflow/*.json` (which silently drops
// the per-feature root state ledger + roadmap bootstrap artifacts).
func TestGitignoreKeepsDurableState(t *testing.T) {
	gi := leanGitignore(t)
	for _, line := range strings.Split(gi, "\n") {
		if strings.TrimSpace(line) == ".workflow/*.json" {
			t.Error("a bare .workflow/*.json ignore drops durable state — ignore by role suffix instead")
		}
	}
}
