package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocsConsistency_NoLegacyValidateScriptReferences(t *testing.T) {
	paths := []string{
		repoPath("docs/architecture/gatekeepers.md"),
		repoPath("docs/architecture/architecture-overview.md"),
		repoPath("internal/scaffold/assets/docs/architecture/gatekeepers.md"),
		repoPath("internal/scaffold/assets/docs/architecture/architecture-overview.md"),
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		if strings.Contains(string(data), "scripts/validate.sh") {
			t.Fatalf("%s still contains scripts/validate.sh", p)
		}
	}
}

func repoPath(p string) string {
	return filepath.Join("..", "..", p)
}
