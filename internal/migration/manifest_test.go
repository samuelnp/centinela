package migration

import "testing"

func TestManagedPathsIncludesCoreFiles(t *testing.T) {
	paths, err := managedPaths()
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) < 3 {
		t.Fatal("expected multiple managed paths")
	}
	foundClaude := false
	foundProject := false
	for _, p := range paths {
		if p == "CLAUDE.md" {
			foundClaude = true
		}
		if p == "PROJECT.md.template" {
			foundProject = true
		}
	}
	if !foundClaude || !foundProject {
		t.Fatal("expected core managed files in manifest")
	}
}
