// Acceptance: specs/delivery-artifact-generation.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readChangelog returns the CHANGELOG.md body in dir.
func readChangelog(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	if err != nil {
		t.Fatalf("read CHANGELOG.md: %v", err)
	}
	return string(data)
}

// Scenario: Delivery inserts exactly one Keep-a-Changelog line, idempotently,
// and never falsely claims a PR was opened against a local bare origin.
func TestAccDeliveryArtifactChangelog(t *testing.T) {
	dir := cdpRepo(t, true) // local bare origin — push works offline
	cdpWorkflow(t, dir, "alpha", false, true)
	writeFile(t, dir, "CHANGELOG.md", "# Changelog\n\n## [Unreleased]\n\n### Added\n")
	writeFile(t, dir, ".workflow/alpha-changelog.md", "- feat: alpha\n")
	commitAll(t, dir)

	out, _ := runDeliverBin(t, dir, "alpha", "--via", "pr")
	// The changelog commit lands before push, so it persists regardless of gh.
	if strings.Contains(out, "Opened pull request") {
		t.Fatalf("must not claim a PR opened against a local bare origin:\n%s", out)
	}
	cl := readChangelog(t, dir)
	if n := strings.Count(cl, "- feat: alpha"); n != 1 {
		t.Fatalf("expected exactly one changelog line, got %d:\n%s", n, cl)
	}

	// Re-running deliver must leave exactly one copy (idempotent changelog step).
	out2, _ := runDeliverBin(t, dir, "alpha", "--via", "pr")
	if strings.Contains(out2, "Opened pull request") {
		t.Fatalf("re-run must not falsely claim a PR:\n%s", out2)
	}
	cl2 := readChangelog(t, dir)
	if n := strings.Count(cl2, "- feat: alpha"); n != 1 {
		t.Fatalf("re-run must keep exactly one changelog line, got %d:\n%s", n, cl2)
	}
}
