package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktree"
)

func TestLoadActiveWorkflows(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	workflow.Save(workflow.New("a"))        //nolint:errcheck
	done := workflow.New("b")
	done.CurrentStep = "done"
	workflow.Save(done) //nolint:errcheck
	if wfs := loadActiveWorkflows(); len(wfs) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(wfs))
	}
}

// When cwd is inside a worktree, loadActiveWorkflows scopes to that worktree's
// feature even though other active workflows exist in .workflow/.
func TestLoadActiveWorkflows_WorktreeScoped(t *testing.T) {
	repo := t.TempDir()
	wtRoot := filepath.Join(repo, worktree.Dir, "scoped-feat")
	wfDir := filepath.Join(wtRoot, workflow.WorkflowDir)
	if err := os.MkdirAll(wfDir, 0o755); err != nil {
		t.Fatal(err)
	}
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(wtRoot)  //nolint:errcheck
	saveActive := func(f string) {
		wf := workflow.New(f)
		wf.CurrentStep = "code"
		workflow.Save(wf) //nolint:errcheck
	}
	saveActive("scoped-feat")
	saveActive("other-feat")
	wfs := loadActiveWorkflows()
	if len(wfs) != 1 || wfs[0].Feature != "scoped-feat" {
		t.Fatalf("expected only scoped-feat, got %+v", wfs)
	}
}
