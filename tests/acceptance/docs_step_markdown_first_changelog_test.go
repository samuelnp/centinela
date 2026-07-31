// Acceptance: specs/docs-step-markdown-first.feature
package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// Scenario: User-facing docs step fails without a changelog entry
func TestDSMFUserFacingFailsWithoutChangelog(t *testing.T) {
	t.Chdir(t.TempDir())
	mustWrite(t, "docs/features/uf.md", "# uf\nsurface: user-facing\n")
	mustWrite(t, workflow.WorkflowDir+"/.gitkeep", "")
	err := workflow.ValidateArtifacts("uf", "docs", nil)
	if err == nil || !strings.Contains(err.Error(), ".workflow/uf-changelog.md") {
		t.Fatalf("missing changelog must fail naming .workflow/uf-changelog.md, got %v", err)
	}
}

// Scenario: Internal docs step keeps the one-line changelog contract
func TestDSMFInternalKeepsChangelogOnlyContract(t *testing.T) {
	t.Chdir(t.TempDir())
	mustWrite(t, "docs/features/int.md", "# int\n")
	mustWrite(t, workflow.WorkflowDir+"/int-changelog.md", "- refactor: tidy\n")
	// Strict workflow saved, and no documentation-specialist evidence exists —
	// the internal contract requires none.
	wf := workflow.New("int")
	wf.CurrentStep = "docs"
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
	if err := workflow.ValidateArtifacts("int", "docs", nil); err != nil {
		t.Fatalf("internal feature with a one-line changelog must pass, got %v", err)
	}
}
