// Acceptance: specs/docs-step-markdown-first.feature
// Supersedes the portal-regeneration scenarios of
// specs/right-size-docs-step.feature: merge must no longer regenerate any
// documentation portal.
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// rdsMergeSource reads cmd/centinela/merge.go from the repo root.
func rdsMergeSource(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "cmd", "centinela", "merge.go"))
	if err != nil {
		t.Fatalf("read merge.go: %v", err)
	}
	return string(data)
}

// Scenario: The HTML pipeline is gone
// merge.go must not reference the portal-regen seam, the docgen package, or
// the generated portal path — no code path regenerates
// docs/project-docs/index.html at merge time.
func TestRDSMergeDoesNotReferencePortalRegen(t *testing.T) {
	src := rdsMergeSource(t)
	for _, banned := range []string{"docsPortalRegen", "docgen", "project-docs", "portal regen"} {
		if strings.Contains(src, banned) {
			t.Fatalf("merge.go must not reference %q — the portal pipeline is deleted", banned)
		}
	}
}

// Scenario: The HTML pipeline is gone
// The docgen package itself is deleted from the tree.
func TestRDSDocgenPackageDeleted(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "..", "internal", "docgen")); !os.IsNotExist(err) {
		t.Fatalf("internal/docgen must be deleted (stat err=%v)", err)
	}
	if _, err := os.Stat(filepath.Join("..", "..", "docs", "project-docs")); !os.IsNotExist(err) {
		t.Fatalf("docs/project-docs must be deleted (stat err=%v)", err)
	}
}
