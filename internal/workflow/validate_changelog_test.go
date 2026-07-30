package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// changelogFixture chdirs into a temp repo with a brief of the given body so
// validateChangelog runs under both surface classifications.
func changelogFixture(t *testing.T, feature, brief string) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/features/"+feature+".md", []byte(brief), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
}

// Changelog-for-all: the non-empty changelog contract applies identically to
// internal AND user-facing features (decision 2 — gate strengthened).
func TestValidateChangelogForBothSurfaces(t *testing.T) {
	surfaces := map[string]string{
		"internal":    "# f\n",
		"user-facing": "# f\nsurface: user-facing\n",
	}
	for name, brief := range surfaces {
		changelogFixture(t, "f", brief)
		path := filepath.Join(WorkflowDir, "f-changelog.md")

		err := validateChangelog("f")
		if err == nil || !strings.Contains(err.Error(), "changelog entry missing") {
			t.Fatalf("%s: missing changelog must fail naming it, got %v", name, err)
		}
		if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
		err = validateChangelog("f")
		if err == nil || !strings.Contains(err.Error(), "changelog entry is empty") {
			t.Fatalf("%s: empty changelog must fail as empty, got %v", name, err)
		}
		if err := os.WriteFile(path, []byte("   \n\t\n  \n"), 0o644); err != nil {
			t.Fatal(err)
		}
		err = validateChangelog("f")
		if err == nil || !strings.Contains(err.Error(), "changelog entry is empty") {
			t.Fatalf("%s: whitespace-only changelog must fail as empty, got %v", name, err)
		}
		if err := os.WriteFile(path, []byte("- feat: shipped\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := validateChangelog("f"); err != nil {
			t.Fatalf("%s: non-empty changelog must pass, got %v", name, err)
		}
	}
}
