// Acceptance: specs/docs-step-markdown-first.feature
// The KB/portal scenarios of specs/right-size-docs-step.feature are superseded
// by the markdown-first contract: the docs ARTIFACT gate requires a changelog
// for every feature; the user-facing real-doc-file rule moved to the
// documentation-specialist evidence outputs (covered in
// docs_step_markdown_first_evidence_test.go).
package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// rdsUserFacing chdirs into a temp repo with a user-facing brief.
func rdsUserFacing(t *testing.T, feature string, withChangelog bool) {
	t.Helper()
	t.Chdir(t.TempDir())
	mustWrite(t, "docs/features/"+feature+".md", "# "+feature+"\nsurface: user-facing\n")
	if withChangelog {
		mustWrite(t, workflow.WorkflowDir+"/"+feature+"-changelog.md", "- feat: uf shipped\n")
	}
}

// Scenario: User-facing docs step passes with a real updated doc file
// (artifact-gate half: the changelog side of the contract; no KB bundle or
// portal is ever demanded any more).
func TestRDSUserFacingArtifactGatePassesWithChangelog(t *testing.T) {
	rdsUserFacing(t, "uf", true)
	if err := workflow.ValidateArtifacts("uf", "docs", nil); err != nil {
		t.Fatalf("user-facing docs artifact gate must pass on the changelog, got %v", err)
	}
}

// Scenario: User-facing docs step fails without a changelog entry
func TestRDSUserFacingFailsWithoutChangelog(t *testing.T) {
	rdsUserFacing(t, "uf", false)
	err := workflow.ValidateArtifacts("uf", "docs", nil)
	if err == nil || !strings.Contains(err.Error(), "changelog entry missing") {
		t.Fatalf("missing changelog must fail naming it, got %v", err)
	}
	if !strings.Contains(err.Error(), "uf-changelog.md") {
		t.Fatalf("error must name the changelog path, got %v", err)
	}
}
