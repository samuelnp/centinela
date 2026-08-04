package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// The anti-bricking guarantee, through the public API: a state file from a
// newer Centinela loads and stays in the active set even when this binary
// cannot model its body — an empty active set is what blocks every governed
// write and makes the autostart hook fork a duplicate workflow.
func TestFutureVersionStaysInTheActiveSet(t *testing.T) {
	bodies := []string{
		`{"schemaVersion":"2.0","feature":"delta","currentStep":"code","steps":{}}`,
		`{"schemaVersion":1.5,"feature":"delta","currentStep":"code","steps":{}}`,
		`{"schemaVersion":true,"feature":"delta","currentStep":"code","steps":{}}`,
		`{"schemaVersion":99,"feature":"delta","currentStep":"code","steps":[{"s":["done"]}]}`,
	}
	for _, body := range bodies {
		t.Run(body[:28], func(t *testing.T) {
			dir := dwsRepo(t)
			dwsWrite(t, "delta", body)
			if _, err := workflow.Load("delta"); err != nil {
				t.Fatalf("a future-version file must never fail to load: %v", err)
			}
			got := workflow.ActiveWorkflows(filepath.Join(dir, workflow.WorkflowDir))
			if len(got) != 1 || got[0].Feature != "delta" {
				t.Fatalf("ActiveWorkflows = %v, want exactly the delta workflow", got)
			}
			if !got[0].Unmodellable() {
				t.Fatal("a file this binary cannot model must be marked unmodellable")
			}
		})
	}
}

// The refusal is the whole reason Load may be permissive: nothing this binary
// writes may truncate a file it does not understand.
func TestFutureVersionSaveIsRefusedWithAnActionableMessage(t *testing.T) {
	dwsRepo(t)
	body := `{"schemaVersion":99,"feature":"delta","currentStep":"code","steps":{}}`
	dwsWrite(t, "delta", body)
	wf, err := workflow.Load("delta")
	if err != nil {
		t.Fatal(err)
	}
	err = workflow.Save(wf)
	if err == nil {
		t.Fatal("saving over a newer schema must be refused")
	}
	for _, want := range []string{workflow.FilePath("delta"), "99", "schema version 1", "centinela update"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("refusal must name %q, got: %v", want, err)
		}
	}
	if string(mustRead(t, workflow.FilePath("delta"))) != body {
		t.Fatal("a refused save must leave the file byte-identical")
	}
}

// A workflow deleted underneath a long-running command must not be silently
// recreated from the stale copy that command is holding.
func TestDeletedWorkflowIsNotResurrected(t *testing.T) {
	dwsRepo(t)
	if err := workflow.Save(workflow.New("alpha")); err != nil {
		t.Fatal(err)
	}
	wf, err := workflow.Load("alpha")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(workflow.FilePath("alpha")); err != nil {
		t.Fatal(err)
	}
	if err := workflow.Save(wf); err == nil {
		t.Fatal("recreating a deleted state file must be reported as a conflict")
	}
	if _, err := os.Stat(workflow.FilePath("alpha")); !os.IsNotExist(err) {
		t.Fatalf("the deleted workflow was resurrected (%v)", err)
	}
}
