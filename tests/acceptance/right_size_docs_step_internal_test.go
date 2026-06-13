// Acceptance: specs/right-size-docs-step.feature
package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// rdsInternal chdirs into a temp repo with an internal (no surface) brief.
func rdsInternal(t *testing.T, feature string) {
	t.Helper()
	t.Chdir(t.TempDir())
	mustWrite(t, "docs/features/"+feature+".md", "# "+feature+"\n")
	mustWrite(t, workflow.WorkflowDir+"/.gitkeep", "")
}

// Scenario: An internal docs step passes with only a one-line changelog
func TestRDSInternalPassesWithChangelog(t *testing.T) {
	rdsInternal(t, "in")
	mustWrite(t, workflow.WorkflowDir+"/in-changelog.md", "- refactor: right-size the docs step\n")
	// No knowledge-base guide present — the internal path must still pass.
	if err := workflow.ValidateArtifacts("in", "docs", nil); err != nil {
		t.Fatalf("internal docs with a one-line changelog must pass, got %v", err)
	}
}

// Scenario: An internal docs step fails without a changelog entry
func TestRDSInternalFailsWithoutChangelog(t *testing.T) {
	rdsInternal(t, "in")
	err := workflow.ValidateArtifacts("in", "docs", nil)
	if err == nil || !strings.Contains(err.Error(), "changelog entry missing") {
		t.Fatalf("missing changelog must fail naming it, got %v", err)
	}
}

// Scenario: An internal docs step fails when the changelog entry is blank
func TestRDSInternalFailsWhenChangelogBlank(t *testing.T) {
	rdsInternal(t, "in")
	mustWrite(t, workflow.WorkflowDir+"/in-changelog.md", "   \n\t\n")
	err := workflow.ValidateArtifacts("in", "docs", nil)
	if err == nil || !strings.Contains(err.Error(), "changelog entry is empty") {
		t.Fatalf("blank changelog must fail as empty, got %v", err)
	}
}
