// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// rshClean reports whether the roadmap-state pathspec is clean (no staged or
// unstaged changes) in dir.
func rshClean(t *testing.T, dir string) bool {
	t.Helper()
	return rshGitOut(t, dir, "status", "--porcelain", "--",
		".workflow/roadmap.json", "ROADMAP.md") == ""
}

// containsAll fails unless every want string appears somewhere in hay.
func containsAll(t *testing.T, hay string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(hay, w) {
			t.Fatalf("expected %q in:\n%s", w, hay)
		}
	}
}
