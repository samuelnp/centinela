// Acceptance: specs/docs-step-markdown-first.feature
package acceptance_test

import (
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// dsmfUserFacing seeds a user-facing feature "uf" on a strict docs step with a
// non-empty changelog, then writes documentation-specialist evidence whose
// outputs are the given JSON list.
func dsmfUserFacing(t *testing.T, outputs string) {
	t.Helper()
	t.Chdir(t.TempDir())
	mustWrite(t, "docs/features/uf.md", "# uf\nsurface: user-facing\n")
	mustWrite(t, workflow.WorkflowDir+"/uf-changelog.md", "- feat: uf shipped\n")
	wf := workflow.New("uf")
	wf.CurrentStep = "docs"
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, orchestration.MarkdownPath("uf", orchestration.RoleDocsSpecialist), "# report")
	data := `{"feature":"uf","step":"docs","role":"documentation-specialist","status":"done","generatedAt":"` +
		// docs is the last step with any required evidence, so the derived
		// successor is the terminal literal, not a role name.
		time.Now().UTC().Format(time.RFC3339) + `","inputs":["i"],"outputs":[` + outputs + `],"edgeCases":[],"handoffTo":"complete"}`
	mustWrite(t, orchestration.JSONPath("uf", orchestration.RoleDocsSpecialist), data)
}

// Scenario: User-facing docs step passes with a real updated doc file
func TestDSMFUserFacingPassesWithRealDocFile(t *testing.T) {
	dsmfUserFacing(t, `"docs/guides/getting-started.md"`)
	mustWrite(t, "docs/guides/getting-started.md", "# Guide\nupdated\n")
	if err := workflow.ValidateArtifacts("uf", "docs", nil); err != nil {
		t.Fatalf("real updated doc file must pass validation, got %v", err)
	}
}

// Scenario: README.md alone satisfies the real-file rule in a project without docs/guides
func TestDSMFReadmeAloneSatisfiesRealFileRule(t *testing.T) {
	dsmfUserFacing(t, `"README.md"`)
	mustWrite(t, "README.md", "# Project\nupdated capability list\n")
	// No docs/guides directory exists in this project on purpose.
	if err := workflow.ValidateArtifacts("uf", "docs", nil); err != nil {
		t.Fatalf("README.md alone must satisfy the real-file rule, got %v", err)
	}
}

// Scenario: User-facing docs step fails when no output is a real doc file
func TestDSMFUserFacingFailsWithoutRealDocOutput(t *testing.T) {
	dsmfUserFacing(t, `".workflow/uf-documentation-specialist.md"`)
	err := workflow.ValidateArtifacts("uf", "docs", nil)
	if err == nil || !strings.Contains(err.Error(), "docs/ or README.md") {
		t.Fatalf("evidence-only outputs must fail naming docs/ or README.md, got %v", err)
	}
}

// Scenario: Documentation-specialist outputs must be real files
func TestDSMFSpecialistOutputsMustBeRealFiles(t *testing.T) {
	dsmfUserFacing(t, `"docs/guides/ghost.md"`)
	err := workflow.ValidateArtifacts("uf", "docs", nil)
	if err == nil || !strings.Contains(err.Error(), "actionable outputs must be real files") {
		t.Fatalf("a missing output path must fail the real-file check, got %v", err)
	}
	if !strings.Contains(err.Error(), "docs/guides/ghost.md") {
		t.Fatalf("error must name the ghost path, got %v", err)
	}
}
