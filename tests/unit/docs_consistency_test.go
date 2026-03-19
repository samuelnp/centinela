package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocsConsistency_NoLegacyWorkflowScriptInCoreDocs(t *testing.T) {
	paths := []string{
		repoPath("CLAUDE.md"),
		repoPath("docs/architecture/new-project-guide.md"),
		repoPath("docs/architecture/workflow-enforcement.md"),
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		s := string(data)
		for _, bad := range []string{"scripts/centinela-workflow.sh", "workflow.sh complete"} {
			if strings.Contains(s, bad) {
				t.Fatalf("%s still contains %q", p, bad)
			}
		}
	}
}

func repoPath(p string) string {
	return filepath.Join("..", "..", p)
}
