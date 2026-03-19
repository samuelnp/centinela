package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/docs-consistency-pass.feature
func TestDocsConsistency_ScaffoldDocsUseCentinelaCommands(t *testing.T) {
	paths := []string{
		repoPath("internal/scaffold/assets/docs/architecture/new-project-guide.md"),
		repoPath("internal/scaffold/assets/docs/architecture/workflow-enforcement.md"),
		repoPath("internal/scaffold/assets/docs/architecture/gatekeeper-prompt.md"),
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		s := string(data)
		if strings.Contains(s, "scripts/centinela-workflow.sh") {
			t.Fatalf("%s still contains legacy workflow script", p)
		}
		if strings.Contains(s, "workflow.sh complete") {
			t.Fatalf("%s still contains workflow.sh complete", p)
		}
	}
}

func repoPath(p string) string {
	return filepath.Join("..", "..", p)
}
