// Acceptance: specs/right-size-docs-step.feature
package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// rdsUserFacing chdirs into a temp repo with a user-facing brief and optionally
// provisions the portal + knowledge-base bundle.
func rdsUserFacing(t *testing.T, feature string, withKB bool) {
	t.Helper()
	t.Chdir(t.TempDir())
	mustWrite(t, "docs/features/"+feature+".md", "# "+feature+"\nsurface: user-facing\n")
	mustWrite(t, "docs/project-docs/index.html", "<html></html>")
	if withKB {
		mustWrite(t, "docs/project-docs/kb/"+feature+".md", "guide")
		mustWrite(t, "docs/project-docs/kb/"+feature+".html", "<html></html>")
	}
}

// Scenario: A user-facing docs step still requires the knowledge-base guide
func TestRDSUserFacingPassesWithKnowledgeBase(t *testing.T) {
	rdsUserFacing(t, "uf", true)
	if err := workflow.ValidateArtifacts("uf", "docs", nil); err != nil {
		t.Fatalf("user-facing docs with portal+KB must pass, got %v", err)
	}
}

// Scenario: A user-facing docs step fails without the knowledge-base guide
func TestRDSUserFacingFailsWithoutKnowledgeBase(t *testing.T) {
	rdsUserFacing(t, "uf", false)
	err := workflow.ValidateArtifacts("uf", "docs", nil)
	if err == nil || !strings.Contains(err.Error(), "knowledge base markdown missing") {
		t.Fatalf("missing KB guide must fail naming it, got %v", err)
	}
}
