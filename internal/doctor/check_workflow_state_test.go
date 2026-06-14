package doctor

import (
	"strings"
	"testing"
)

// seedWorkflow writes a genuine active workflow-state file for feature.
func seedWorkflow(t *testing.T, feature string) {
	t.Helper()
	writeFile(t, ".workflow/"+feature+".json",
		`{"feature":"`+feature+`","currentStep":"code","steps":{}}`)
}

func TestWorkflowStateNoneOK(t *testing.T) {
	dir := repoFixture(t)
	stubGit(t, okGit(""))
	d := workflowStateCheck{}.Run(Context{Root: dir})
	if d.Status != OK {
		t.Fatalf("no orphans must be OK, got %v", d.Status)
	}
}

func TestWorkflowStateOrphanReported(t *testing.T) {
	dir := repoFixture(t)
	seedWorkflow(t, "ghost")
	// branch does not exist + no worktree => orphan.
	stubGit(t, func(repo string, args ...string) ([]byte, error) {
		if args[0] == "rev-parse" && len(args) > 1 && args[1] == "--is-inside-work-tree" {
			return []byte("true\n"), nil
		}
		return nil, errStub
	})
	d := workflowStateCheck{}.Run(Context{Root: dir})
	if d.Status != Error {
		t.Fatalf("orphan must Error, got %v", d.Status)
	}
	if d.Repair == nil || d.Repair.Apply != nil {
		t.Fatal("workflow-state repair must be report-only")
	}
	if !strings.Contains(d.Repair.Command, "rm ") {
		t.Fatalf("command must surface rm: %q", d.Repair.Command)
	}
}

func TestWorkflowStateLiveBranchNotOrphan(t *testing.T) {
	dir := repoFixture(t)
	seedWorkflow(t, "live")
	stubGit(t, func(repo string, args ...string) ([]byte, error) {
		if args[0] == "rev-parse" && len(args) > 1 && args[1] == "--is-inside-work-tree" {
			return []byte("true\n"), nil
		}
		return []byte("ok"), nil // branch exists
	})
	d := workflowStateCheck{}.Run(Context{Root: dir})
	if d.Status != OK {
		t.Fatalf("workflow with live branch is not orphaned, got %v", d.Status)
	}
}
